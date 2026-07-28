GOOSE ?= goose
GOOSE_DRIVER ?= postgres
MIGRATIONS_DIR ?= ./migrations

-include .env

DATABASE_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
GOOSE_COMMAND = $(GOOSE) -dir $(MIGRATIONS_DIR) $(GOOSE_DRIVER) "$(DATABASE_URL)"

.PHONY: migrate-up migrate-down migrate-status migrate-reset test-database

migrate-up:
	$(GOOSE_COMMAND) up

migrate-down:
	$(GOOSE_COMMAND) down

migrate-status:
	$(GOOSE_COMMAND) status

migrate-reset:
	$(GOOSE_COMMAND) reset
	$(GOOSE_COMMAND) up

test-database:
	SPCASE_TEST_DATABASE_URL="$(DATABASE_URL)" go test -tags=integration ./internal/repository
