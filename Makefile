GOOSE ?= goose
GOOSE_DRIVER ?= postgres
MIGRATIONS_DIR ?= ./migrations
PRODUCTION_MIGRATIONS_MANIFEST ?= $(MIGRATIONS_DIR)/production.txt
GOVULNCHECK ?= go run golang.org/x/vuln/cmd/govulncheck@v1.6.0

-include .env

DATABASE_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
SPCASE_TEST_MIGRATOR_DATABASE_URL ?= postgres://$(DB_MIGRATOR_USER):$(DB_MIGRATOR_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
SPCASE_TEST_APP_DATABASE_URL ?= postgres://$(DB_APP_USER):$(DB_APP_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
GOOSE_COMMAND = $(GOOSE) -dir $(MIGRATIONS_DIR) $(GOOSE_DRIVER) "$(DATABASE_URL)"

.PHONY: migrate-up migrate-production migrate-down migrate-status migrate-reset test-database frontend-build security-check

migrate-up:
	$(GOOSE_COMMAND) up

migrate-production:
	@sh ./scripts/migrate-production.sh "$(GOOSE)" "$(GOOSE_DRIVER)" "$(DATABASE_URL)" "$(MIGRATIONS_DIR)" "$(PRODUCTION_MIGRATIONS_MANIFEST)"

migrate-down:
	$(GOOSE_COMMAND) down

migrate-status:
	$(GOOSE_COMMAND) status

migrate-reset:
	$(GOOSE_COMMAND) reset
	$(GOOSE_COMMAND) up

test-database:
	SPCASE_TEST_MIGRATOR_DATABASE_URL="$(SPCASE_TEST_MIGRATOR_DATABASE_URL)" \
	SPCASE_TEST_APP_DATABASE_URL="$(SPCASE_TEST_APP_DATABASE_URL)" \
	go test -race -count=1 -tags=integration ./internal/...

frontend-build:
	npm ci
	npm run build

security-check:
	go test -race -count=1 ./...
	go vet ./...
	$(GOVULNCHECK) ./...
	npm audit
