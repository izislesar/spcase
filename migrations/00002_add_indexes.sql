-- +goose Up
CREATE INDEX idx_evaluations_team_id ON evaluations(team_id);
CREATE INDEX idx_evaluations_jury_id ON evaluations(jury_id);
CREATE INDEX idx_teams_captain_id ON teams(captain_id);

-- +goose Down
DROP INDEX IF EXISTS idx_teams_captain_id;
DROP INDEX IF EXISTS idx_evaluations_jury_id;
DROP INDEX IF EXISTS idx_evaluations_team_id;
