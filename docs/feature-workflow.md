# Feature Workflow

This document is the default workflow for adding features in this project.
It keeps implementation simple while preserving CQRS boundaries.

## 1) Define the feature first

- Write scope and acceptance criteria.
- Confirm API behavior, validation rules, and error cases.
- List out read paths and write paths separately.

## 2) Choose CQRS path

- Write use case: add or update command flow in `internal/application/command`.
- Read use case: add or update query flow in `internal/application/query`.
- Keep business logic in `internal/application/services`.

## 3) Implement application logic

- Add or update service methods in `internal/application/services`.
- Keep service code focused on domain rules and orchestration.
- Do not place HTTP or database-specific details in services.

## 4) Update persistence if needed

- Add migration in `migrations/` when schema changes.
- Add or update SQL in `sql/queries/*.sql`.
- Generate SQLC code.

```bash
sqlc generate
```

- Adapt repository implementation in `internal/infrastructure/db/postgres`.

### Add a new SQL query (SQLC)

- Open the right file in `sql/queries/` (for example `products.sql`, `sellers.sql`).
- Add query name using SQLC comment format.
- Reference docs:
- Query annotations (`:one`, `:many`, `:exec`, `:execrows`): https://docs.sqlc.dev/en/stable/reference/query-annotations.html
- Select queries: https://docs.sqlc.dev/en/stable/howto/select.html
- Delete/exec queries: https://docs.sqlc.dev/en/stable/howto/delete.html

```sql
-- name: FindProductsBySellerID :many
SELECT id, name, price, seller_id, created_at, updated_at
FROM products
WHERE seller_id = $1
ORDER BY created_at DESC;
```

- Run SQLC generation.

```bash
sqlc generate
```

- Use the generated method from `internal/infrastructure/db/sqlc/*.sql.go`.
- Map generated DB rows to domain entities in `internal/infrastructure/db/postgres/*_repository.go`.
- Add tests for repository behavior and service usage.

## 5) Update transport layer

- Add or update request/response DTOs in `internal/handler/api/rest/dto`.
- Add or update mapper functions in `internal/handler/api/rest/dto/mapper`.
- Add or update handlers in `internal/handler/api/rest`.

## 6) Add or update business errors

- Register new error code and message constants in `internal/infrastructure/pkg/errorx/constants.go`.
- Reuse existing error kinds (`InvalidArgument`, `NotFound`, `Conflict`, etc.).
- Ensure new codes match the team convention.

## 7) Wire dependencies

- Wire new repositories/services/controllers in `internal/builders/server_builder.go`.
- Keep the builder explicit and easy to follow.

## 8) Add tests

- Service tests for business rules.
- Controller tests for request/response behavior.
- Repository tests when query or mapping changes.
- Cover both success and failure paths.

## 9) Run local checks

```bash
gofmt -w .
go test ./...
```

## 10) Done checklist

- Feature behavior matches acceptance criteria.
- Error responses are stable and documented by code constants.
- No CQRS boundary leaks between command and query concerns.
- Tests pass locally.
