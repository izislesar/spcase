-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TYPE user_role AS ENUM ('USER', 'JURY', 'ADMIN');

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    full_name VARCHAR(255) NOT NULL,
    university VARCHAR(255),
    email VARCHAR(255) NOT NULL,
    telegram VARCHAR(100),
    password_hash VARCHAR(255) NOT NULL,
    role user_role NOT NULL,
    auth_version INTEGER NOT NULL DEFAULT 1 CHECK (auth_version > 0),
    disabled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_user_profile CHECK (
        role <> 'USER' OR (university IS NOT NULL AND telegram IS NOT NULL)
    )
);

CREATE TABLE teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    invite_code VARCHAR(8) NOT NULL,
    captain_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_invite_code_format CHECK (invite_code ~ '^[A-Z0-9]{8}$'),
    CONSTRAINT uq_teams_invite_code UNIQUE (invite_code)
);

CREATE TABLE team_members (
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (team_id, user_id),
    CONSTRAINT uq_team_members_user UNIQUE (user_id)
);

ALTER TABLE teams
    ADD CONSTRAINT fk_teams_captain_membership
    FOREIGN KEY (id, captain_id)
    REFERENCES team_members(team_id, user_id)
    DEFERRABLE INITIALLY DEFERRED;

CREATE TABLE submissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL UNIQUE REFERENCES teams(id) ON DELETE CASCADE,
    solution_url TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_solution_url_format CHECK (solution_url ~* '^https?://[^[:space:]]+$')
);

CREATE TABLE evaluations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    jury_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    criterion_id SMALLINT NOT NULL CHECK (criterion_id BETWEEN 1 AND 6),
    score SMALLINT NOT NULL CHECK (score BETWEEN 0 AND 10),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_jury_team_criterion UNIQUE (jury_id, team_id, criterion_id)
);

CREATE TABLE evaluation_state (
    singleton_id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (singleton_id = 1),
    is_closed BOOLEAN NOT NULL DEFAULT FALSE,
    closed_at TIMESTAMPTZ,
    closed_by UUID REFERENCES users(id) ON DELETE RESTRICT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_evaluation_closed_fields CHECK (
        (is_closed AND closed_at IS NOT NULL AND closed_by IS NOT NULL)
        OR (NOT is_closed AND closed_at IS NULL AND closed_by IS NULL)
    )
);

INSERT INTO evaluation_state (singleton_id) VALUES (1);

CREATE TABLE evaluation_state_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    action VARCHAR(10) NOT NULL CHECK (action IN ('OPEN', 'CLOSE')),
    admin_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose StatementBegin
CREATE FUNCTION enforce_team_member_role() RETURNS TRIGGER
LANGUAGE plpgsql AS $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM users WHERE id = NEW.user_id AND role = 'USER') THEN
        RAISE EXCEPTION 'team member must have USER role' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE CONSTRAINT TRIGGER trg_team_member_role
AFTER INSERT OR UPDATE ON team_members
DEFERRABLE INITIALLY IMMEDIATE
FOR EACH ROW EXECUTE FUNCTION enforce_team_member_role();

-- +goose StatementBegin
CREATE FUNCTION enforce_jury_role() RETURNS TRIGGER
LANGUAGE plpgsql AS $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM users WHERE id = NEW.jury_id AND role = 'JURY') THEN
        RAISE EXCEPTION 'evaluation author must have JURY role' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE CONSTRAINT TRIGGER trg_evaluation_jury_role
AFTER INSERT OR UPDATE ON evaluations
DEFERRABLE INITIALLY IMMEDIATE
FOR EACH ROW EXECUTE FUNCTION enforce_jury_role();

-- +goose StatementBegin
CREATE FUNCTION enforce_evaluation_state_admin_role() RETURNS TRIGGER
LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.closed_by IS NOT NULL
       AND NOT EXISTS (SELECT 1 FROM users WHERE id = NEW.closed_by AND role = 'ADMIN') THEN
        RAISE EXCEPTION 'evaluation state actor must have ADMIN role' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION enforce_evaluation_event_admin_role() RETURNS TRIGGER
LANGUAGE plpgsql AS $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM users WHERE id = NEW.admin_id AND role = 'ADMIN') THEN
        RAISE EXCEPTION 'evaluation event actor must have ADMIN role' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE CONSTRAINT TRIGGER trg_evaluation_state_admin_role
AFTER INSERT OR UPDATE ON evaluation_state
DEFERRABLE INITIALLY IMMEDIATE
FOR EACH ROW EXECUTE FUNCTION enforce_evaluation_state_admin_role();

CREATE CONSTRAINT TRIGGER trg_evaluation_event_admin_role
AFTER INSERT OR UPDATE ON evaluation_state_events
DEFERRABLE INITIALLY IMMEDIATE
FOR EACH ROW EXECUTE FUNCTION enforce_evaluation_event_admin_role();

-- +goose StatementBegin
CREATE FUNCTION prevent_evaluation_state_removal() RETURNS TRIGGER
LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'evaluation state singleton cannot be removed' USING ERRCODE = '55000';
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER trg_evaluation_state_no_delete
BEFORE DELETE ON evaluation_state
FOR EACH ROW EXECUTE FUNCTION prevent_evaluation_state_removal();

CREATE TRIGGER trg_evaluation_state_no_truncate
BEFORE TRUNCATE ON evaluation_state
FOR EACH STATEMENT EXECUTE FUNCTION prevent_evaluation_state_removal();

-- +goose StatementBegin
CREATE FUNCTION prevent_evaluation_event_mutation() RETURNS TRIGGER
LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'evaluation state events are append-only' USING ERRCODE = '55000';
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER trg_evaluation_events_append_only
BEFORE UPDATE OR DELETE ON evaluation_state_events
FOR EACH ROW EXECUTE FUNCTION prevent_evaluation_event_mutation();

CREATE TRIGGER trg_evaluation_events_no_truncate
BEFORE TRUNCATE ON evaluation_state_events
FOR EACH STATEMENT EXECUTE FUNCTION prevent_evaluation_event_mutation();

-- +goose StatementBegin
CREATE FUNCTION enforce_user_role_relations() RETURNS TRIGGER
LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.role <> 'USER'
       AND EXISTS (SELECT 1 FROM team_members WHERE user_id = NEW.id) THEN
        RAISE EXCEPTION 'team member must retain USER role' USING ERRCODE = '23514';
    END IF;
    IF NEW.role <> 'JURY'
       AND EXISTS (SELECT 1 FROM evaluations WHERE jury_id = NEW.id) THEN
        RAISE EXCEPTION 'evaluation author must retain JURY role' USING ERRCODE = '23514';
    END IF;
    IF NEW.role <> 'ADMIN'
       AND (
           EXISTS (SELECT 1 FROM evaluation_state WHERE closed_by = NEW.id)
           OR EXISTS (SELECT 1 FROM evaluation_state_events WHERE admin_id = NEW.id)
       ) THEN
        RAISE EXCEPTION 'evaluation state actor must retain ADMIN role' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER trg_user_role_relations
BEFORE UPDATE OF role ON users
FOR EACH ROW
WHEN (OLD.role IS DISTINCT FROM NEW.role)
EXECUTE FUNCTION enforce_user_role_relations();

-- +goose Down
DROP TRIGGER IF EXISTS trg_user_role_relations ON users;
DROP TRIGGER IF EXISTS trg_evaluation_events_no_truncate ON evaluation_state_events;
DROP TRIGGER IF EXISTS trg_evaluation_events_append_only ON evaluation_state_events;
DROP TRIGGER IF EXISTS trg_evaluation_state_no_truncate ON evaluation_state;
DROP TRIGGER IF EXISTS trg_evaluation_state_no_delete ON evaluation_state;
DROP TRIGGER IF EXISTS trg_evaluation_event_admin_role ON evaluation_state_events;
DROP TRIGGER IF EXISTS trg_evaluation_state_admin_role ON evaluation_state;
DROP TRIGGER IF EXISTS trg_evaluation_jury_role ON evaluations;
DROP TRIGGER IF EXISTS trg_team_member_role ON team_members;
DROP FUNCTION IF EXISTS prevent_evaluation_event_mutation();
DROP FUNCTION IF EXISTS prevent_evaluation_state_removal();
DROP FUNCTION IF EXISTS enforce_evaluation_event_admin_role();
DROP FUNCTION IF EXISTS enforce_evaluation_state_admin_role();
DROP FUNCTION IF EXISTS enforce_user_role_relations();
DROP FUNCTION IF EXISTS enforce_jury_role();
DROP FUNCTION IF EXISTS enforce_team_member_role();
DROP TABLE IF EXISTS evaluation_state_events;
DROP TABLE IF EXISTS evaluation_state;
DROP TABLE IF EXISTS evaluations;
DROP TABLE IF EXISTS submissions;
ALTER TABLE IF EXISTS teams DROP CONSTRAINT IF EXISTS fk_teams_captain_membership;
DROP TABLE IF EXISTS team_members;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS user_role;
