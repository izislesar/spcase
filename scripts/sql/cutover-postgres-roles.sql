\set ON_ERROR_STOP on
\getenv cutover_confirmation SPCASE_CONFIRM_EXISTING_DB_CUTOVER
\getenv cutover_database SPCASE_CUTOVER_DATABASE
\getenv migrator_password DB_MIGRATOR_PASSWORD
\getenv app_password DB_APP_PASSWORD
\getenv legacy_owner SPCASE_LEGACY_DB_ROLE
\getenv admin_user POSTGRES_ADMIN_USER

SELECT set_config('spcase.cutover_confirmation', :'cutover_confirmation', false) AS confirmation_setting \gset
SELECT set_config('spcase.cutover_database', :'cutover_database', false) AS database_setting \gset
SELECT set_config('spcase.legacy_owner', :'legacy_owner', false) AS legacy_owner_setting \gset
SELECT set_config('spcase.admin_user', :'admin_user', false) AS admin_user_setting \gset

\echo 'Cutover preflight: validating administrative connection and legacy ownership'
DO $preflight$
DECLARE
    database_owner name;
    schema_owner name;
    unexpected text;
    required_table text;
    goose_table_exists boolean;
    goose_sequence_exists boolean;
    migration_version bigint;
    already_converted boolean;
    expected record;
    privilege text;
    actual boolean;
    should_have boolean;
BEGIN
    IF current_setting('spcase.cutover_confirmation') <> 'YES' THEN
        RAISE EXCEPTION 'cutover confirmation guard is missing';
    END IF;
    IF current_database() <> current_setting('spcase.cutover_database') THEN
        RAISE EXCEPTION 'connected database % does not match explicit target', current_database();
    END IF;
    IF current_user <> current_setting('spcase.admin_user') THEN
        RAISE EXCEPTION 'administrative connection role does not match POSTGRES_ADMIN_USER';
    END IF;
    IF NOT (SELECT rolsuper FROM pg_roles WHERE rolname = current_user) THEN
        RAISE EXCEPTION 'administrative connection must be a superuser';
    END IF;
    IF (SELECT datistemplate FROM pg_database WHERE datname = current_database()) THEN
        RAISE EXCEPTION 'target database must not be a template database';
    END IF;
    IF current_database() IN ('postgres', 'template0', 'template1') THEN
        RAISE EXCEPTION 'target must be an application database, got %', current_database();
    END IF;
    IF NOT (SELECT datallowconn FROM pg_database WHERE datname = current_database()) THEN
        RAISE EXCEPTION 'target database does not allow connections';
    END IF;

    SELECT pg_get_userbyid(datdba)
    INTO database_owner
    FROM pg_database
    WHERE datname = current_database();

    SELECT pg_get_userbyid(nspowner)
    INTO schema_owner
    FROM pg_namespace
    WHERE nspname = 'public';

    IF database_owner IS NULL OR schema_owner IS NULL THEN
        RAISE EXCEPTION 'target database or public schema owner cannot be determined';
    END IF;
    already_converted := database_owner = 'spcase_migrator' AND schema_owner = 'spcase_migrator';
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = current_setting('spcase.legacy_owner')) THEN
        RAISE EXCEPTION 'explicit legacy owner role % does not exist', current_setting('spcase.legacy_owner');
    END IF;
    IF database_owner NOT IN (current_setting('spcase.legacy_owner'), 'spcase_migrator') THEN
        RAISE EXCEPTION 'database owner % is neither explicit legacy owner nor spcase_migrator', database_owner;
    END IF;
    IF schema_owner NOT IN (current_setting('spcase.legacy_owner'), 'pg_database_owner', 'spcase_migrator') THEN
        RAISE EXCEPTION 'public schema owner % is unexpected', schema_owner;
    END IF;

    FOREACH required_table IN ARRAY ARRAY[
        'users', 'teams', 'team_members', 'submissions', 'evaluations',
        'evaluation_state', 'evaluation_state_events'
    ] LOOP
        IF to_regclass(format('public.%I', required_table)) IS NULL THEN
            RAISE EXCEPTION 'required legacy application table public.% is missing', required_table;
        END IF;
    END LOOP;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_type t
        JOIN pg_namespace n ON n.oid = t.typnamespace
        WHERE n.nspname = 'public' AND t.typname = 'user_role' AND t.typtype = 'e'
    ) THEN
        RAISE EXCEPTION 'required enum public.user_role is missing';
    END IF;

    goose_table_exists := to_regclass('public.goose_db_version') IS NOT NULL;
    goose_sequence_exists := to_regclass('public.goose_db_version_id_seq') IS NOT NULL;
    IF NOT goose_table_exists OR NOT goose_sequence_exists THEN
        RAISE EXCEPTION 'Goose metadata table and sequence are required before cutover';
    END IF;
    EXECUTE 'SELECT COALESCE(MAX(version_id) FILTER (WHERE is_applied), 0) FROM public.goose_db_version'
    INTO migration_version;
    IF migration_version < 2 OR migration_version > 5 THEN
        RAISE EXCEPTION 'database migration version % is outside the supported range 2..5', migration_version;
    END IF;
    IF EXISTS (
        SELECT 1
        FROM public.goose_db_version
        WHERE version_id = 3 AND is_applied
    ) THEN
        RAISE EXCEPTION 'development seed migration 00003 must not be applied in production';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM pg_roles
        WHERE rolname IN ('spcase_migrator', 'spcase_app')
          AND (rolsuper OR rolcreatedb OR rolcreaterole OR rolreplication OR rolbypassrls)
    ) THEN
        RAISE EXCEPTION 'existing target roles have unsafe cluster attributes';
    END IF;
    IF EXISTS (
        SELECT 1
        FROM pg_auth_members membership
        JOIN pg_roles member_role ON member_role.oid = membership.member
        JOIN pg_roles granted_role ON granted_role.oid = membership.roleid
        WHERE member_role.rolname IN ('spcase_migrator', 'spcase_app')
           OR granted_role.rolname IN ('spcase_migrator', 'spcase_app')
    ) THEN
        RAISE EXCEPTION 'existing target roles must not participate in role memberships';
    END IF;

    SELECT string_agg(n.nspname, ', ' ORDER BY n.nspname)
    INTO unexpected
    FROM pg_namespace n
    WHERE n.nspname <> 'public'
      AND n.nspname <> 'information_schema'
      AND n.nspname !~ '^pg_'
      AND (
          EXISTS (SELECT 1 FROM pg_class c WHERE c.relnamespace = n.oid)
          OR EXISTS (SELECT 1 FROM pg_proc p WHERE p.pronamespace = n.oid)
          OR EXISTS (
              SELECT 1 FROM pg_type t
              WHERE t.typnamespace = n.oid AND t.typtype IN ('c', 'd', 'e')
          )
      );
    IF unexpected IS NOT NULL THEN
        RAISE EXCEPTION 'unexpected user-defined schemas require manual review: %', unexpected;
    END IF;

    SELECT string_agg(format('%I.%I owned by %I', n.nspname, c.relname, pg_get_userbyid(c.relowner)), ', ')
    INTO unexpected
    FROM pg_class c
    JOIN pg_namespace n ON n.oid = c.relnamespace
    WHERE n.nspname = 'public'
      AND c.relkind IN ('r', 'p', 'S', 'v', 'm', 'f', 'i', 'I')
      AND NOT EXISTS (
          SELECT 1
          FROM pg_depend d
          JOIN pg_extension e ON e.oid = d.refobjid
          WHERE d.classid = 'pg_class'::regclass AND d.objid = c.oid AND d.deptype = 'e'
      )
      AND (
          (already_converted AND pg_get_userbyid(c.relowner) <> 'spcase_migrator')
          OR (
              NOT already_converted
              AND pg_get_userbyid(c.relowner) NOT IN (current_setting('spcase.legacy_owner'), 'spcase_migrator')
          )
      );
    IF unexpected IS NOT NULL THEN
        RAISE EXCEPTION 'unexpected relation owners require manual review: %', unexpected;
    END IF;

    SELECT string_agg(format('%I.%I owned by %I', n.nspname, p.proname, pg_get_userbyid(p.proowner)), ', ')
    INTO unexpected
    FROM pg_proc p
    JOIN pg_namespace n ON n.oid = p.pronamespace
    WHERE n.nspname = 'public'
      AND NOT EXISTS (
          SELECT 1
          FROM pg_depend d
          JOIN pg_extension e ON e.oid = d.refobjid
          WHERE d.classid = 'pg_proc'::regclass AND d.objid = p.oid AND d.deptype = 'e'
      )
      AND (
          (already_converted AND pg_get_userbyid(p.proowner) <> 'spcase_migrator')
          OR (
              NOT already_converted
              AND pg_get_userbyid(p.proowner) NOT IN (current_setting('spcase.legacy_owner'), 'spcase_migrator')
          )
      );
    IF unexpected IS NOT NULL THEN
        RAISE EXCEPTION 'unexpected routine owners require manual review: %', unexpected;
    END IF;

    SELECT string_agg(format('%I.%I owned by %I', n.nspname, t.typname, pg_get_userbyid(t.typowner)), ', ')
    INTO unexpected
    FROM pg_type t
    JOIN pg_namespace n ON n.oid = t.typnamespace
    WHERE n.nspname = 'public'
      AND t.typrelid = 0
      AND t.typelem = 0
      AND t.typtype IN ('b', 'c', 'd', 'e', 'm', 'r')
      AND NOT EXISTS (
          SELECT 1
          FROM pg_depend d
          JOIN pg_extension e ON e.oid = d.refobjid
          WHERE d.classid = 'pg_type'::regclass AND d.objid = t.oid AND d.deptype = 'e'
      )
      AND (
          (already_converted AND pg_get_userbyid(t.typowner) <> 'spcase_migrator')
          OR (
              NOT already_converted
              AND pg_get_userbyid(t.typowner) NOT IN (current_setting('spcase.legacy_owner'), 'spcase_migrator')
          )
      );
    IF unexpected IS NOT NULL THEN
        RAISE EXCEPTION 'unexpected type owners require manual review: %', unexpected;
    END IF;

    IF already_converted THEN
        IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'spcase_app') THEN
            RAISE EXCEPTION 'converted database is missing spcase_app';
        END IF;
        IF has_schema_privilege('spcase_app', 'public', 'CREATE') OR
           NOT has_schema_privilege('spcase_app', 'public', 'USAGE') THEN
            RAISE EXCEPTION 'converted database has runtime schema ACL drift';
        END IF;
        IF NOT has_database_privilege('spcase_app', current_database(), 'CONNECT') OR
           has_database_privilege('spcase_app', current_database(), 'CREATE') OR
           has_database_privilege('spcase_app', current_database(), 'TEMP') THEN
            RAISE EXCEPTION 'converted database has runtime database ACL drift';
        END IF;
        IF NOT has_database_privilege('spcase_migrator', current_database(), 'CONNECT') OR
           NOT has_database_privilege('spcase_migrator', current_database(), 'CREATE') OR
           NOT has_database_privilege('spcase_migrator', current_database(), 'TEMP') THEN
            RAISE EXCEPTION 'converted database has migrator database ACL drift';
        END IF;
        IF NOT has_type_privilege('spcase_app', 'public.user_role', 'USAGE') THEN
            RAISE EXCEPTION 'converted database has runtime type ACL drift';
        END IF;
        IF current_setting('spcase.legacy_owner') <> 'postgres' AND (
            NOT has_database_privilege(current_setting('spcase.legacy_owner'), current_database(), 'CONNECT')
            OR has_database_privilege(current_setting('spcase.legacy_owner'), current_database(), 'CREATE')
            OR has_database_privilege(current_setting('spcase.legacy_owner'), current_database(), 'TEMP')
            OR NOT has_schema_privilege(current_setting('spcase.legacy_owner'), 'public', 'USAGE')
            OR has_schema_privilege(current_setting('spcase.legacy_owner'), 'public', 'CREATE')
            OR NOT has_type_privilege(current_setting('spcase.legacy_owner'), 'public.user_role', 'USAGE')
        ) THEN
            RAISE EXCEPTION 'converted database has transitional legacy ACL drift';
        END IF;
        FOR expected IN
            SELECT * FROM (VALUES
                ('users', ARRAY['SELECT', 'INSERT', 'UPDATE']),
                ('teams', ARRAY['SELECT', 'INSERT', 'UPDATE', 'DELETE']),
                ('team_members', ARRAY['SELECT', 'INSERT', 'DELETE']),
                ('submissions', ARRAY['SELECT', 'INSERT', 'UPDATE', 'DELETE']),
                ('evaluations', ARRAY['SELECT', 'INSERT', 'UPDATE']),
                ('evaluation_state', ARRAY['SELECT', 'UPDATE']),
                ('evaluation_state_events', ARRAY['INSERT'])
            ) AS permissions(table_name, allowed)
        LOOP
            FOREACH privilege IN ARRAY ARRAY['SELECT', 'INSERT', 'UPDATE', 'DELETE', 'TRUNCATE', 'REFERENCES', 'TRIGGER']
            LOOP
                should_have := privilege = ANY(expected.allowed);
                actual := has_table_privilege('spcase_app', format('public.%I', expected.table_name), privilege);
                IF actual <> should_have THEN
                    RAISE EXCEPTION 'converted database has runtime ACL drift on public.% privilege %',
                        expected.table_name, privilege;
                END IF;
                IF current_setting('spcase.legacy_owner') <> 'postgres' THEN
                    actual := has_table_privilege(
                        current_setting('spcase.legacy_owner'),
                        format('public.%I', expected.table_name),
                        privilege
                    );
                    IF actual <> should_have THEN
                        RAISE EXCEPTION 'converted database has transitional legacy ACL drift on public.% privilege %',
                            expected.table_name, privilege;
                    END IF;
                END IF;
            END LOOP;
        END LOOP;
        FOREACH privilege IN ARRAY ARRAY['SELECT', 'INSERT', 'UPDATE', 'DELETE', 'TRUNCATE'] LOOP
            IF has_table_privilege('spcase_app', 'public.goose_db_version', privilege) THEN
                RAISE EXCEPTION 'converted database has Goose table ACL drift for %', privilege;
            END IF;
            IF current_setting('spcase.legacy_owner') <> 'postgres' AND
               has_table_privilege(
                   current_setting('spcase.legacy_owner'), 'public.goose_db_version', privilege
               ) THEN
                RAISE EXCEPTION 'converted database has transitional legacy Goose table ACL drift for %', privilege;
            END IF;
        END LOOP;
        FOREACH privilege IN ARRAY ARRAY['USAGE', 'SELECT', 'UPDATE'] LOOP
            IF has_sequence_privilege('spcase_app', 'public.goose_db_version_id_seq', privilege) THEN
                RAISE EXCEPTION 'converted database has Goose sequence ACL drift for %', privilege;
            END IF;
            IF current_setting('spcase.legacy_owner') <> 'postgres' AND
               has_sequence_privilege(
                   current_setting('spcase.legacy_owner'),
                   'public.goose_db_version_id_seq',
                   privilege
               ) THEN
                RAISE EXCEPTION 'converted database has transitional legacy Goose sequence ACL drift for %', privilege;
            END IF;
        END LOOP;
        IF EXISTS (
            WITH expected(object_type, privilege_type) AS (
                VALUES
                    ('r'::"char", 'SELECT'), ('r'::"char", 'INSERT'),
                    ('r'::"char", 'UPDATE'), ('r'::"char", 'DELETE'),
                    ('S'::"char", 'USAGE'), ('S'::"char", 'SELECT'), ('S'::"char", 'UPDATE'),
                    ('T'::"char", 'USAGE')
            ), actual AS (
                SELECT defaults.defaclobjtype, privileges.privilege_type
                FROM pg_default_acl defaults
                JOIN pg_roles owner ON owner.oid = defaults.defaclrole
                JOIN pg_namespace namespace ON namespace.oid = defaults.defaclnamespace
                CROSS JOIN LATERAL aclexplode(defaults.defaclacl) privileges
                JOIN pg_roles grantee ON grantee.oid = privileges.grantee
                WHERE owner.rolname = 'spcase_migrator'
                  AND namespace.nspname = 'public'
                  AND grantee.rolname = 'spcase_app'
            )
            (SELECT * FROM expected EXCEPT SELECT * FROM actual)
            UNION ALL
            (SELECT * FROM actual EXCEPT SELECT * FROM expected)
        ) THEN
            RAISE EXCEPTION 'converted database has default privilege drift';
        END IF;
        RAISE NOTICE 'database is already converted and passed ownership/ACL preflight';
    ELSE
        RAISE NOTICE 'legacy ownership accepted: database=%, public schema=%', database_owner, schema_owner;
    END IF;
END
$preflight$;

SELECT md5(concat_ws('|',
    (SELECT string_agg(id::text, ',' ORDER BY id) FROM public.users),
    (SELECT string_agg(id::text, ',' ORDER BY id) FROM public.teams),
    (SELECT string_agg(team_id::text || ':' || user_id::text, ',' ORDER BY team_id, user_id) FROM public.team_members),
    (SELECT string_agg(id::text, ',' ORDER BY id) FROM public.submissions),
    (SELECT string_agg(id::text, ',' ORDER BY id) FROM public.evaluations),
    CASE WHEN to_regclass('public.goose_db_version') IS NULL THEN 'absent'
         ELSE (SELECT string_agg(version_id::text || ':' || is_applied::text, ',' ORDER BY id) FROM public.goose_db_version)
    END
)) AS pre_cutover_data_fingerprint
\gset

\echo 'Cutover mutation: creating safe roles and transferring public application objects'
BEGIN;

SELECT format(
    'CREATE ROLE spcase_migrator LOGIN PASSWORD %L NOSUPERUSER NOCREATEDB NOCREATEROLE INHERIT NOREPLICATION NOBYPASSRLS',
    :'migrator_password'
)
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'spcase_migrator') \gexec

SELECT format(
    'CREATE ROLE spcase_app LOGIN PASSWORD %L NOSUPERUSER NOCREATEDB NOCREATEROLE INHERIT NOREPLICATION NOBYPASSRLS',
    :'app_password'
)
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'spcase_app') \gexec

ALTER ROLE spcase_migrator LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE INHERIT NOREPLICATION NOBYPASSRLS;
ALTER ROLE spcase_app LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE INHERIT NOREPLICATION NOBYPASSRLS;

ALTER SCHEMA public OWNER TO spcase_migrator;

DO $relations$
DECLARE
    object record;
    statement text;
BEGIN
    FOR object IN
        SELECT c.relkind, n.nspname, c.relname
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE n.nspname = 'public'
          AND c.relkind IN ('r', 'p', 'S', 'v', 'm', 'f')
          AND NOT EXISTS (
              SELECT 1
              FROM pg_depend d
              JOIN pg_extension e ON e.oid = d.refobjid
              WHERE d.classid = 'pg_class'::regclass AND d.objid = c.oid AND d.deptype = 'e'
          )
          AND NOT (
              c.relkind = 'S'
              AND EXISTS (
                  SELECT 1
                  FROM pg_depend d
                  WHERE d.classid = 'pg_class'::regclass
                    AND d.objid = c.oid
                    AND d.refclassid = 'pg_class'::regclass
                    AND d.deptype IN ('a', 'i')
              )
          )
        ORDER BY CASE WHEN c.relkind IN ('r', 'p') THEN 0 ELSE 1 END, c.relname
    LOOP
        statement := CASE object.relkind
            WHEN 'S' THEN format('ALTER SEQUENCE %I.%I OWNER TO spcase_migrator', object.nspname, object.relname)
            WHEN 'v' THEN format('ALTER VIEW %I.%I OWNER TO spcase_migrator', object.nspname, object.relname)
            WHEN 'm' THEN format('ALTER MATERIALIZED VIEW %I.%I OWNER TO spcase_migrator', object.nspname, object.relname)
            WHEN 'f' THEN format('ALTER FOREIGN TABLE %I.%I OWNER TO spcase_migrator', object.nspname, object.relname)
            ELSE format('ALTER TABLE %I.%I OWNER TO spcase_migrator', object.nspname, object.relname)
        END;
        EXECUTE statement;
    END LOOP;
END
$relations$;

DO $routines$
DECLARE
    object record;
BEGIN
    FOR object IN
        SELECT p.prokind, n.nspname, p.proname, pg_get_function_identity_arguments(p.oid) AS arguments
        FROM pg_proc p
        JOIN pg_namespace n ON n.oid = p.pronamespace
        WHERE n.nspname = 'public'
          AND NOT EXISTS (
              SELECT 1
              FROM pg_depend d
              JOIN pg_extension e ON e.oid = d.refobjid
              WHERE d.classid = 'pg_proc'::regclass AND d.objid = p.oid AND d.deptype = 'e'
          )
        ORDER BY p.proname, p.oid
    LOOP
        CASE object.prokind
            WHEN 'p' THEN
                EXECUTE format('ALTER PROCEDURE %I.%I(%s) OWNER TO spcase_migrator',
                    object.nspname, object.proname, object.arguments);
            WHEN 'a' THEN
                EXECUTE format('ALTER AGGREGATE %I.%I(%s) OWNER TO spcase_migrator',
                    object.nspname, object.proname, object.arguments);
            ELSE
                EXECUTE format('ALTER FUNCTION %I.%I(%s) OWNER TO spcase_migrator',
                    object.nspname, object.proname, object.arguments);
        END CASE;
    END LOOP;
END
$routines$;

DO $types$
DECLARE
    object record;
BEGIN
    FOR object IN
        SELECT t.typtype, n.nspname, t.typname
        FROM pg_type t
        JOIN pg_namespace n ON n.oid = t.typnamespace
        WHERE n.nspname = 'public'
          AND t.typrelid = 0
          AND t.typelem = 0
          AND t.typtype IN ('b', 'c', 'd', 'e', 'm', 'r')
          AND NOT EXISTS (
              SELECT 1
              FROM pg_depend d
              JOIN pg_extension e ON e.oid = d.refobjid
              WHERE d.classid = 'pg_type'::regclass AND d.objid = t.oid AND d.deptype = 'e'
          )
        ORDER BY t.typname
    LOOP
        IF object.typtype = 'd' THEN
            EXECUTE format('ALTER DOMAIN %I.%I OWNER TO spcase_migrator', object.nspname, object.typname);
        ELSE
            EXECUTE format('ALTER TYPE %I.%I OWNER TO spcase_migrator', object.nspname, object.typname);
        END IF;
    END LOOP;
END
$types$;

SELECT format('REVOKE ALL PRIVILEGES ON DATABASE %I FROM PUBLIC', current_database()) \gexec
SELECT format('REVOKE ALL PRIVILEGES ON DATABASE %I FROM spcase_app', current_database()) \gexec
SELECT format('REVOKE ALL PRIVILEGES ON DATABASE %I FROM %I', current_database(), current_setting('spcase.legacy_owner'))
WHERE current_setting('spcase.legacy_owner') <> 'postgres' \gexec
SELECT format('GRANT CONNECT, CREATE, TEMPORARY ON DATABASE %I TO spcase_migrator', current_database()) \gexec
SELECT format('GRANT CONNECT ON DATABASE %I TO spcase_app', current_database()) \gexec

REVOKE ALL PRIVILEGES ON SCHEMA public FROM PUBLIC;
GRANT USAGE, CREATE ON SCHEMA public TO spcase_migrator;
REVOKE ALL PRIVILEGES ON SCHEMA public FROM spcase_app;
GRANT USAGE ON SCHEMA public TO spcase_app;
SELECT format('REVOKE ALL PRIVILEGES ON SCHEMA public FROM %I', current_setting('spcase.legacy_owner'))
WHERE current_setting('spcase.legacy_owner') <> 'postgres' \gexec
SELECT format('GRANT USAGE ON SCHEMA public TO %I', current_setting('spcase.legacy_owner'))
WHERE current_setting('spcase.legacy_owner') <> 'postgres' \gexec

REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA public FROM PUBLIC;
REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA public FROM spcase_app;
SELECT format('REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA public FROM %I', current_setting('spcase.legacy_owner'))
WHERE current_setting('spcase.legacy_owner') <> 'postgres' \gexec
GRANT SELECT, INSERT, UPDATE ON TABLE public.users TO spcase_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.teams TO spcase_app;
GRANT SELECT, INSERT, DELETE ON TABLE public.team_members TO spcase_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.submissions TO spcase_app;
GRANT SELECT, INSERT, UPDATE ON TABLE public.evaluations TO spcase_app;
GRANT SELECT, UPDATE ON TABLE public.evaluation_state TO spcase_app;
GRANT INSERT ON TABLE public.evaluation_state_events TO spcase_app;

SELECT format('GRANT %s ON TABLE public.%I TO %I', privileges, table_name, current_setting('spcase.legacy_owner'))
FROM (VALUES
    ('users', 'SELECT, INSERT, UPDATE'),
    ('teams', 'SELECT, INSERT, UPDATE, DELETE'),
    ('team_members', 'SELECT, INSERT, DELETE'),
    ('submissions', 'SELECT, INSERT, UPDATE, DELETE'),
    ('evaluations', 'SELECT, INSERT, UPDATE'),
    ('evaluation_state', 'SELECT, UPDATE'),
    ('evaluation_state_events', 'INSERT')
) AS legacy_grant(table_name, privileges)
WHERE current_setting('spcase.legacy_owner') <> 'postgres' \gexec

REVOKE ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public FROM PUBLIC;
REVOKE ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public FROM spcase_app;
SELECT format('REVOKE ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public FROM %I', current_setting('spcase.legacy_owner'))
WHERE current_setting('spcase.legacy_owner') <> 'postgres' \gexec
DO $sequence_grants$
DECLARE
    object record;
BEGIN
    FOR object IN
        SELECT n.nspname, c.relname
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE n.nspname = 'public'
          AND c.relkind = 'S'
          AND c.relname <> 'goose_db_version_id_seq'
    LOOP
        EXECUTE format('GRANT USAGE, SELECT, UPDATE ON SEQUENCE %I.%I TO spcase_app',
            object.nspname, object.relname);
        IF current_setting('spcase.legacy_owner') <> 'postgres' THEN
            EXECUTE format('GRANT USAGE, SELECT, UPDATE ON SEQUENCE %I.%I TO %I',
                object.nspname, object.relname, current_setting('spcase.legacy_owner'));
        END IF;
    END LOOP;
END
$sequence_grants$;

REVOKE ALL PRIVILEGES ON TYPE public.user_role FROM PUBLIC;
GRANT USAGE ON TYPE public.user_role TO spcase_app;
SELECT format('REVOKE ALL PRIVILEGES ON TYPE public.user_role FROM %I', current_setting('spcase.legacy_owner'))
WHERE current_setting('spcase.legacy_owner') <> 'postgres' \gexec
SELECT format('GRANT USAGE ON TYPE public.user_role TO %I', current_setting('spcase.legacy_owner'))
WHERE current_setting('spcase.legacy_owner') <> 'postgres' \gexec

ALTER DEFAULT PRIVILEGES FOR ROLE spcase_migrator IN SCHEMA public
    REVOKE ALL PRIVILEGES ON TABLES FROM PUBLIC;
ALTER DEFAULT PRIVILEGES FOR ROLE spcase_migrator IN SCHEMA public
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO spcase_app;
ALTER DEFAULT PRIVILEGES FOR ROLE spcase_migrator IN SCHEMA public
    REVOKE ALL PRIVILEGES ON SEQUENCES FROM PUBLIC;
ALTER DEFAULT PRIVILEGES FOR ROLE spcase_migrator IN SCHEMA public
    GRANT USAGE, SELECT, UPDATE ON SEQUENCES TO spcase_app;
ALTER DEFAULT PRIVILEGES FOR ROLE spcase_migrator IN SCHEMA public
    REVOKE ALL PRIVILEGES ON TYPES FROM PUBLIC;
ALTER DEFAULT PRIVILEGES FOR ROLE spcase_migrator IN SCHEMA public
    GRANT USAGE ON TYPES TO spcase_app;

REVOKE ALL PRIVILEGES ON TABLE public.goose_db_version FROM spcase_app;
REVOKE ALL PRIVILEGES ON SEQUENCE public.goose_db_version_id_seq FROM spcase_app;

COMMIT;

\echo 'Cutover mutation: transferring database ownership outside the transactional object stage'
SELECT format('ALTER DATABASE %I OWNER TO spcase_migrator', current_database()) \gexec
SELECT format('GRANT CONNECT ON DATABASE %I TO %I', current_database(), current_setting('spcase.legacy_owner'))
WHERE current_setting('spcase.legacy_owner') <> 'postgres' \gexec

\echo 'Cutover post-validation: checking ownership and least-privilege ACLs'
DO $post_validation$
DECLARE
    unexpected text;
    expected record;
    privilege text;
    actual boolean;
    should_have boolean;
    goose_table_exists boolean;
BEGIN
    IF (SELECT pg_get_userbyid(datdba) FROM pg_database WHERE datname = current_database()) <> 'spcase_migrator' THEN
        RAISE EXCEPTION 'database owner is not spcase_migrator';
    END IF;
    IF (SELECT pg_get_userbyid(nspowner) FROM pg_namespace WHERE nspname = 'public') <> 'spcase_migrator' THEN
        RAISE EXCEPTION 'public schema owner is not spcase_migrator';
    END IF;
    IF EXISTS (
        SELECT 1 FROM pg_roles
        WHERE rolname IN ('spcase_migrator', 'spcase_app')
          AND (NOT rolcanlogin OR rolsuper OR rolcreatedb OR rolcreaterole OR rolreplication OR rolbypassrls)
    ) THEN
        RAISE EXCEPTION 'target role attributes are not least privilege';
    END IF;
    IF EXISTS (
        SELECT 1
        FROM pg_auth_members membership
        JOIN pg_roles member_role ON member_role.oid = membership.member
        JOIN pg_roles granted_role ON granted_role.oid = membership.roleid
        WHERE member_role.rolname IN ('spcase_migrator', 'spcase_app')
           OR granted_role.rolname IN ('spcase_migrator', 'spcase_app')
    ) THEN
        RAISE EXCEPTION 'target role membership drift detected';
    END IF;

    SELECT string_agg(format('%I.%I', n.nspname, c.relname), ', ')
    INTO unexpected
    FROM pg_class c
    JOIN pg_namespace n ON n.oid = c.relnamespace
    WHERE n.nspname = 'public'
      AND c.relkind IN ('r', 'p', 'S', 'v', 'm', 'f', 'i', 'I')
      AND NOT EXISTS (
          SELECT 1 FROM pg_depend d JOIN pg_extension e ON e.oid = d.refobjid
          WHERE d.classid = 'pg_class'::regclass AND d.objid = c.oid AND d.deptype = 'e'
      )
      AND pg_get_userbyid(c.relowner) <> 'spcase_migrator';
    IF unexpected IS NOT NULL THEN
        RAISE EXCEPTION 'relations not owned by spcase_migrator: %', unexpected;
    END IF;

    SELECT string_agg(format('%I.%I', n.nspname, p.proname), ', ')
    INTO unexpected
    FROM pg_proc p
    JOIN pg_namespace n ON n.oid = p.pronamespace
    WHERE n.nspname = 'public'
      AND NOT EXISTS (
          SELECT 1 FROM pg_depend d JOIN pg_extension e ON e.oid = d.refobjid
          WHERE d.classid = 'pg_proc'::regclass AND d.objid = p.oid AND d.deptype = 'e'
      )
      AND pg_get_userbyid(p.proowner) <> 'spcase_migrator';
    IF unexpected IS NOT NULL THEN
        RAISE EXCEPTION 'routines not owned by spcase_migrator: %', unexpected;
    END IF;

    SELECT string_agg(format('%I.%I', n.nspname, t.typname), ', ')
    INTO unexpected
    FROM pg_type t
    JOIN pg_namespace n ON n.oid = t.typnamespace
    WHERE n.nspname = 'public'
      AND t.typrelid = 0
      AND t.typelem = 0
      AND t.typtype IN ('b', 'c', 'd', 'e', 'm', 'r')
      AND NOT EXISTS (
          SELECT 1 FROM pg_depend d JOIN pg_extension e ON e.oid = d.refobjid
          WHERE d.classid = 'pg_type'::regclass AND d.objid = t.oid AND d.deptype = 'e'
      )
      AND pg_get_userbyid(t.typowner) <> 'spcase_migrator';
    IF unexpected IS NOT NULL THEN
        RAISE EXCEPTION 'types not owned by spcase_migrator: %', unexpected;
    END IF;

    IF has_schema_privilege('spcase_app', 'public', 'CREATE') OR
       NOT has_schema_privilege('spcase_app', 'public', 'USAGE') THEN
        RAISE EXCEPTION 'runtime schema privileges are invalid';
    END IF;
    IF NOT has_database_privilege('spcase_app', current_database(), 'CONNECT') OR
       has_database_privilege('spcase_app', current_database(), 'CREATE') OR
       has_database_privilege('spcase_app', current_database(), 'TEMP') THEN
        RAISE EXCEPTION 'runtime database privileges are invalid';
    END IF;
    IF NOT has_database_privilege('spcase_migrator', current_database(), 'CONNECT') OR
       NOT has_database_privilege('spcase_migrator', current_database(), 'CREATE') OR
       NOT has_database_privilege('spcase_migrator', current_database(), 'TEMP') THEN
        RAISE EXCEPTION 'migrator database-owner privileges are invalid';
    END IF;
    IF current_setting('spcase.legacy_owner') <> 'postgres' AND (
        NOT has_database_privilege(current_setting('spcase.legacy_owner'), current_database(), 'CONNECT')
        OR has_database_privilege(current_setting('spcase.legacy_owner'), current_database(), 'CREATE')
        OR has_database_privilege(current_setting('spcase.legacy_owner'), current_database(), 'TEMP')
        OR NOT has_schema_privilege(current_setting('spcase.legacy_owner'), 'public', 'USAGE')
        OR has_schema_privilege(current_setting('spcase.legacy_owner'), 'public', 'CREATE')
    ) THEN
        RAISE EXCEPTION 'transitional legacy runtime database/schema privileges are invalid';
    END IF;

    FOR expected IN
        SELECT * FROM (VALUES
            ('users', ARRAY['SELECT', 'INSERT', 'UPDATE']),
            ('teams', ARRAY['SELECT', 'INSERT', 'UPDATE', 'DELETE']),
            ('team_members', ARRAY['SELECT', 'INSERT', 'DELETE']),
            ('submissions', ARRAY['SELECT', 'INSERT', 'UPDATE', 'DELETE']),
            ('evaluations', ARRAY['SELECT', 'INSERT', 'UPDATE']),
            ('evaluation_state', ARRAY['SELECT', 'UPDATE']),
            ('evaluation_state_events', ARRAY['INSERT'])
        ) AS permissions(table_name, allowed)
    LOOP
        FOREACH privilege IN ARRAY ARRAY['SELECT', 'INSERT', 'UPDATE', 'DELETE', 'TRUNCATE', 'REFERENCES', 'TRIGGER']
        LOOP
            should_have := privilege = ANY(expected.allowed);
            actual := has_table_privilege('spcase_app', format('public.%I', expected.table_name), privilege);
            IF actual <> should_have THEN
                RAISE EXCEPTION 'unexpected spcase_app privilege on %.%: % expected %, got %',
                    'public', expected.table_name, privilege, should_have, actual;
            END IF;
            IF current_setting('spcase.legacy_owner') <> 'postgres' THEN
                actual := has_table_privilege(
                    current_setting('spcase.legacy_owner'),
                    format('public.%I', expected.table_name),
                    privilege
                );
                IF actual <> should_have THEN
                    RAISE EXCEPTION 'unexpected transitional legacy privilege on %.%: % expected %, got %',
                        'public', expected.table_name, privilege, should_have, actual;
                END IF;
            END IF;
        END LOOP;
    END LOOP;

    goose_table_exists := to_regclass('public.goose_db_version') IS NOT NULL;
    IF goose_table_exists THEN
        FOREACH privilege IN ARRAY ARRAY['SELECT', 'INSERT', 'UPDATE', 'DELETE', 'TRUNCATE'] LOOP
            IF has_table_privilege('spcase_app', 'public.goose_db_version', privilege) THEN
                RAISE EXCEPTION 'spcase_app retains Goose table privilege %', privilege;
            END IF;
            IF current_setting('spcase.legacy_owner') <> 'postgres' AND
               has_table_privilege(
                   current_setting('spcase.legacy_owner'),
                   'public.goose_db_version',
                   privilege
               ) THEN
                RAISE EXCEPTION 'transitional legacy role retains Goose table privilege %', privilege;
            END IF;
        END LOOP;
        FOREACH privilege IN ARRAY ARRAY['USAGE', 'SELECT', 'UPDATE'] LOOP
            IF has_sequence_privilege('spcase_app', 'public.goose_db_version_id_seq', privilege) THEN
                RAISE EXCEPTION 'spcase_app retains Goose sequence privilege %', privilege;
            END IF;
            IF current_setting('spcase.legacy_owner') <> 'postgres' AND
               has_sequence_privilege(
                   current_setting('spcase.legacy_owner'),
                   'public.goose_db_version_id_seq',
                   privilege
               ) THEN
                RAISE EXCEPTION 'transitional legacy role retains Goose sequence privilege %', privilege;
            END IF;
        END LOOP;
    END IF;

    IF NOT has_type_privilege('spcase_app', 'public.user_role', 'USAGE') THEN
        RAISE EXCEPTION 'spcase_app lacks user_role USAGE';
    END IF;
    IF EXISTS (
        SELECT 1 FROM pg_class WHERE relowner = (SELECT oid FROM pg_roles WHERE rolname = 'spcase_app')
    ) OR EXISTS (
        SELECT 1 FROM pg_proc WHERE proowner = (SELECT oid FROM pg_roles WHERE rolname = 'spcase_app')
    ) OR EXISTS (
        SELECT 1 FROM pg_type WHERE typowner = (SELECT oid FROM pg_roles WHERE rolname = 'spcase_app')
    ) THEN
        RAISE EXCEPTION 'spcase_app unexpectedly owns database objects';
    END IF;
END
$post_validation$;

WITH expected(object_type, privilege_type) AS (
    VALUES
        ('r'::"char", 'SELECT'), ('r'::"char", 'INSERT'),
        ('r'::"char", 'UPDATE'), ('r'::"char", 'DELETE'),
        ('S'::"char", 'USAGE'), ('S'::"char", 'SELECT'), ('S'::"char", 'UPDATE'),
        ('T'::"char", 'USAGE')
), actual AS (
    SELECT defaults.defaclobjtype, privileges.privilege_type
    FROM pg_default_acl defaults
    JOIN pg_roles owner ON owner.oid = defaults.defaclrole
    JOIN pg_namespace namespace ON namespace.oid = defaults.defaclnamespace
    CROSS JOIN LATERAL aclexplode(defaults.defaclacl) privileges
    JOIN pg_roles grantee ON grantee.oid = privileges.grantee
    WHERE owner.rolname = 'spcase_migrator'
      AND namespace.nspname = 'public'
      AND grantee.rolname = 'spcase_app'
), unsafe_public AS (
    SELECT 1
    FROM pg_default_acl defaults
    JOIN pg_roles owner ON owner.oid = defaults.defaclrole
    JOIN pg_namespace namespace ON namespace.oid = defaults.defaclnamespace
    CROSS JOIN LATERAL aclexplode(defaults.defaclacl) privileges
    WHERE owner.rolname = 'spcase_migrator'
      AND namespace.nspname = 'public'
      AND privileges.grantee = 0
      AND defaults.defaclobjtype IN ('r', 'S', 'T')
)
SELECT NOT EXISTS (
    (SELECT * FROM expected EXCEPT SELECT * FROM actual)
    UNION ALL
    (SELECT * FROM actual EXCEPT SELECT * FROM expected)
) AND NOT EXISTS (SELECT 1 FROM unsafe_public) AS default_privileges_valid
\gset
\if :default_privileges_valid
\else
    \warn 'cutover post-validation failed: default privileges are unsafe or incomplete'
    \quit 1
\endif

SELECT :'pre_cutover_data_fingerprint' = md5(concat_ws('|',
    (SELECT string_agg(id::text, ',' ORDER BY id) FROM public.users),
    (SELECT string_agg(id::text, ',' ORDER BY id) FROM public.teams),
    (SELECT string_agg(team_id::text || ':' || user_id::text, ',' ORDER BY team_id, user_id) FROM public.team_members),
    (SELECT string_agg(id::text, ',' ORDER BY id) FROM public.submissions),
    (SELECT string_agg(id::text, ',' ORDER BY id) FROM public.evaluations),
    CASE WHEN to_regclass('public.goose_db_version') IS NULL THEN 'absent'
         ELSE (SELECT string_agg(version_id::text || ':' || is_applied::text, ',' ORDER BY id) FROM public.goose_db_version)
    END
)) AS data_preserved
\gset
\if :data_preserved
\else
    \warn 'cutover post-validation failed: application data or migration history changed'
    \quit 1
\endif

SELECT
    current_database() AS database,
    pg_get_userbyid(d.datdba) AS database_owner,
    pg_get_userbyid(n.nspowner) AS public_schema_owner,
    CASE WHEN to_regclass('public.goose_db_version') IS NULL THEN 'absent' ELSE 'present' END AS goose_metadata
FROM pg_database d
JOIN pg_namespace n ON n.nspname = 'public'
WHERE d.datname = current_database();
