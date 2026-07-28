-- +goose Up
-- Intentionally empty: credentials are never committed. Development accounts
-- may be inserted locally with a separate, untracked migration.

-- +goose Down
SELECT 1;
