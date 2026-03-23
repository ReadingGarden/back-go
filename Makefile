ifneq (,$(wildcard ./.env))
include .env
export
endif

SWAG_CMD := go run github.com/swaggo/swag/cmd/swag@latest
SQLC_CMD := go run github.com/sqlc-dev/sqlc/cmd/sqlc@latest
GOOSE_CMD := go run github.com/pressly/goose/v3/cmd/goose@latest
GOOSE_DRIVER ?= mysql
GOOSE_DBSTRING ?= $(DB_USER):$(DB_PASSWORD)@tcp($(DB_HOST):$(DB_PORT))/$(DB_NAME)?parseTime=true
MYSQL_SERVICE ?= mysql

.PHONY: run test cover fmt tidy swag sqlc goose-status goose-up goose-down db-up db-down db-logs db-wait test-db

run:
	go run ./cmd/server

test:
	go test ./...

cover:
	go test ./... -cover

fmt:
	gofmt -w $$(find cmd internal pkg -name '*.go' -type f)

tidy:
	go mod tidy

swag:
	$(SWAG_CMD) init -g cmd/server/main.go -o docs --parseInternal

sqlc:
	$(SQLC_CMD) generate

goose-status:
	$(GOOSE_CMD) -dir migrations $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" status

goose-up:
	$(GOOSE_CMD) -dir migrations $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" up

goose-down:
	$(GOOSE_CMD) -dir migrations $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" down

db-up:
	docker compose up -d $(MYSQL_SERVICE)

db-down:
	docker compose down

db-logs:
	docker compose logs -f $(MYSQL_SERVICE)

db-wait:
	until [ "$$(docker inspect -f '{{.State.Health.Status}}' reading-garden-mysql 2>/dev/null)" = "healthy" ]; do sleep 2; done

test-db:
	INTEGRATION_MYSQL=true go test ./internal/db -run TestOpenMySQLWithDocker -count=1 -v
