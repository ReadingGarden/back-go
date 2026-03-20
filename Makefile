SWAG_CMD := go run github.com/swaggo/swag/cmd/swag@latest
SQLC_CMD := go run github.com/sqlc-dev/sqlc/cmd/sqlc@latest
GOOSE_CMD := go run github.com/pressly/goose/v3/cmd/goose@latest
GOOSE_DRIVER ?= mysql
GOOSE_DBSTRING ?= root:@tcp(127.0.0.1:3306)/reading_garden?parseTime=true

.PHONY: run test cover fmt tidy swag sqlc goose-status goose-up goose-down

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
