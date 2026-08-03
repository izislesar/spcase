-- +goose Up
REVOKE ALL PRIVILEGES ON TABLE public.goose_db_version FROM spcase_app;
REVOKE ALL PRIVILEGES ON SEQUENCE public.goose_db_version_id_seq FROM spcase_app;

-- +goose Down
GRANT USAGE, SELECT, UPDATE ON SEQUENCE public.goose_db_version_id_seq TO spcase_app;
