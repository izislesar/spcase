-- +goose Up
-- Development-only fixtures. Every account uses the password "password".
-- Production deployments must stop at migration 00002.
INSERT INTO users (
    id, full_name, university, email, telegram, password_hash, role
) VALUES
    (
        '10000000-0000-4000-8000-000000000001',
        'Development Admin',
        NULL,
        'admin@dev.spcase.ru',
        NULL,
        '$2a$10$4Hln1mjL0p6gjY0lTWu7Q.LmkmTPMW1KjLRKevBqE.Q5dNvYOumbu',
        'ADMIN'
    ),
    (
        '20000000-0000-4000-8000-000000000001',
        'Jury One',
        NULL,
        'jury1@dev.spcase.ru',
        NULL,
        '$2a$10$4Hln1mjL0p6gjY0lTWu7Q.LmkmTPMW1KjLRKevBqE.Q5dNvYOumbu',
        'JURY'
    ),
    (
        '20000000-0000-4000-8000-000000000002',
        'Jury Two',
        NULL,
        'jury2@dev.spcase.ru',
        NULL,
        '$2a$10$4Hln1mjL0p6gjY0lTWu7Q.LmkmTPMW1KjLRKevBqE.Q5dNvYOumbu',
        'JURY'
    ),
    (
        '20000000-0000-4000-8000-000000000003',
        'Jury Three',
        NULL,
        'jury3@dev.spcase.ru',
        NULL,
        '$2a$10$4Hln1mjL0p6gjY0lTWu7Q.LmkmTPMW1KjLRKevBqE.Q5dNvYOumbu',
        'JURY'
    ),
    (
        '30000000-0000-4000-8000-000000000001',
        'Participant One',
        'SPbU',
        'user1@dev.spcase.ru',
        '@spcase_user1',
        '$2a$10$4Hln1mjL0p6gjY0lTWu7Q.LmkmTPMW1KjLRKevBqE.Q5dNvYOumbu',
        'USER'
    ),
    (
        '30000000-0000-4000-8000-000000000002',
        'Participant Two',
        'SPbU',
        'user2@dev.spcase.ru',
        '@spcase_user2',
        '$2a$10$4Hln1mjL0p6gjY0lTWu7Q.LmkmTPMW1KjLRKevBqE.Q5dNvYOumbu',
        'USER'
    ),
    (
        '30000000-0000-4000-8000-000000000003',
        'Participant Three',
        'ITMO',
        'user3@dev.spcase.ru',
        '@spcase_user3',
        '$2a$10$4Hln1mjL0p6gjY0lTWu7Q.LmkmTPMW1KjLRKevBqE.Q5dNvYOumbu',
        'USER'
    ),
    (
        '30000000-0000-4000-8000-000000000004',
        'Participant Four',
        'ITMO',
        'user4@dev.spcase.ru',
        '@spcase_user4',
        '$2a$10$4Hln1mjL0p6gjY0lTWu7Q.LmkmTPMW1KjLRKevBqE.Q5dNvYOumbu',
        'USER'
    ),
    (
        '30000000-0000-4000-8000-000000000005',
        'Participant Five',
        'HSE',
        'user5@dev.spcase.ru',
        '@spcase_user5',
        '$2a$10$4Hln1mjL0p6gjY0lTWu7Q.LmkmTPMW1KjLRKevBqE.Q5dNvYOumbu',
        'USER'
    ),
    (
        '30000000-0000-4000-8000-000000000006',
        'Participant Six',
        'HSE',
        'user6@dev.spcase.ru',
        '@spcase_user6',
        '$2a$10$4Hln1mjL0p6gjY0lTWu7Q.LmkmTPMW1KjLRKevBqE.Q5dNvYOumbu',
        'USER'
    ),
    (
        '30000000-0000-4000-8000-000000000007',
        'Participant Seven',
        'MIPT',
        'user7@dev.spcase.ru',
        '@spcase_user7',
        '$2a$10$4Hln1mjL0p6gjY0lTWu7Q.LmkmTPMW1KjLRKevBqE.Q5dNvYOumbu',
        'USER'
    ),
    (
        '30000000-0000-4000-8000-000000000008',
        'Participant Eight',
        'MIPT',
        'user8@dev.spcase.ru',
        '@spcase_user8',
        '$2a$10$4Hln1mjL0p6gjY0lTWu7Q.LmkmTPMW1KjLRKevBqE.Q5dNvYOumbu',
        'USER'
    );

INSERT INTO teams (id, name, invite_code, captain_id) VALUES
    (
        '40000000-0000-4000-8000-000000000001',
        'Dev Alpha',
        'DEVALPHA',
        '30000000-0000-4000-8000-000000000001'
    ),
    (
        '40000000-0000-4000-8000-000000000002',
        'Dev Beta',
        'DEVBETA2',
        '30000000-0000-4000-8000-000000000004'
    ),
    (
        '40000000-0000-4000-8000-000000000003',
        'Dev Gamma',
        'DEVGAMM3',
        '30000000-0000-4000-8000-000000000006'
    );

INSERT INTO team_members (team_id, user_id) VALUES
    ('40000000-0000-4000-8000-000000000001', '30000000-0000-4000-8000-000000000001'),
    ('40000000-0000-4000-8000-000000000001', '30000000-0000-4000-8000-000000000002'),
    ('40000000-0000-4000-8000-000000000001', '30000000-0000-4000-8000-000000000003'),
    ('40000000-0000-4000-8000-000000000002', '30000000-0000-4000-8000-000000000004'),
    ('40000000-0000-4000-8000-000000000002', '30000000-0000-4000-8000-000000000005'),
    ('40000000-0000-4000-8000-000000000003', '30000000-0000-4000-8000-000000000006');

INSERT INTO submissions (id, team_id, solution_url) VALUES
    (
        '50000000-0000-4000-8000-000000000001',
        '40000000-0000-4000-8000-000000000001',
        'https://example.com/dev-alpha'
    ),
    (
        '50000000-0000-4000-8000-000000000002',
        '40000000-0000-4000-8000-000000000002',
        'https://example.com/dev-beta'
    );

-- +goose Down
UPDATE evaluation_state
SET is_closed = FALSE,
    closed_at = NULL,
    closed_by = NULL,
    updated_at = clock_timestamp()
WHERE closed_by = '10000000-0000-4000-8000-000000000001';

ALTER TABLE evaluation_state_events
    DISABLE TRIGGER trg_evaluation_events_append_only;

DELETE FROM evaluation_state_events
WHERE admin_id = '10000000-0000-4000-8000-000000000001';

ALTER TABLE evaluation_state_events
    ENABLE TRIGGER trg_evaluation_events_append_only;

DELETE FROM evaluations
WHERE jury_id IN (
    '20000000-0000-4000-8000-000000000001',
    '20000000-0000-4000-8000-000000000002',
    '20000000-0000-4000-8000-000000000003'
);

DELETE FROM teams
WHERE captain_id IN (
    '30000000-0000-4000-8000-000000000001',
    '30000000-0000-4000-8000-000000000002',
    '30000000-0000-4000-8000-000000000003',
    '30000000-0000-4000-8000-000000000004',
    '30000000-0000-4000-8000-000000000005',
    '30000000-0000-4000-8000-000000000006',
    '30000000-0000-4000-8000-000000000007',
    '30000000-0000-4000-8000-000000000008'
);

DELETE FROM team_members
WHERE user_id IN (
    '30000000-0000-4000-8000-000000000001',
    '30000000-0000-4000-8000-000000000002',
    '30000000-0000-4000-8000-000000000003',
    '30000000-0000-4000-8000-000000000004',
    '30000000-0000-4000-8000-000000000005',
    '30000000-0000-4000-8000-000000000006',
    '30000000-0000-4000-8000-000000000007',
    '30000000-0000-4000-8000-000000000008'
);

DELETE FROM submissions
WHERE (SELECT COUNT(*) FROM team_members WHERE team_id = submissions.team_id) < 2;

DELETE FROM users
WHERE id IN (
    '10000000-0000-4000-8000-000000000001',
    '20000000-0000-4000-8000-000000000001',
    '20000000-0000-4000-8000-000000000002',
    '20000000-0000-4000-8000-000000000003',
    '30000000-0000-4000-8000-000000000001',
    '30000000-0000-4000-8000-000000000002',
    '30000000-0000-4000-8000-000000000003',
    '30000000-0000-4000-8000-000000000004',
    '30000000-0000-4000-8000-000000000005',
    '30000000-0000-4000-8000-000000000006',
    '30000000-0000-4000-8000-000000000007',
    '30000000-0000-4000-8000-000000000008'
);
