-- +goose Up
CREATE UNIQUE INDEX uq_users_email_ci ON users (LOWER(email));
CREATE UNIQUE INDEX uq_teams_name_ci ON teams (LOWER(name));
CREATE INDEX idx_evaluations_team_id ON evaluations(team_id);
CREATE INDEX idx_teams_captain_id ON teams(captain_id);

-- +goose Down
DROP INDEX IF EXISTS idx_teams_captain_id;
DROP INDEX IF EXISTS idx_evaluations_team_id;
DROP INDEX IF EXISTS uq_teams_name_ci;
DROP INDEX IF EXISTS uq_users_email_ci;
