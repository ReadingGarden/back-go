# ReadingGarden Back-go

Go `Gin` migration backend for ReadingGarden. The legacy Django app in `back/` remains the API contract source of truth.

## Stack

- HTTP: `Gin`
- Logging: `log/slog`
- DB access: `database/sql` + `sqlc`
- Migrations: `goose`
- API docs: `swag` + `gin-swagger`

## Commands

- `go run ./cmd/server`
- `go test ./...`
- `go test ./... -cover`
- `make swag`
- `make sqlc`
- `make goose-status`

## Directories

- `cmd/server`: application bootstrap
- `internal/http`: router and handlers
- `internal/db`: MySQL connection and generated sqlc access code
- `sql/schema`: schema files for sqlc analysis
- `sql/query`: query files for sqlc generation
- `migrations`: goose SQL migrations
