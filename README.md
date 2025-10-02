# Go Gin Order Settlement

A production-ready sample service for managing products, orders, and transaction settlement jobs. The project is built with [Gin](https://gin-gonic.com/), uses PostgreSQL for persistence, and demonstrates domain-driven modular architecture with background job orchestration.

## Features

- Product CRUD APIs with stock management.
- Order workflows with validation and pagination.
- Asynchronous settlement job processing with cancellable jobs and CSV exports.
- Makefile tasks for dependency management, running, testing, and Docker orchestration.

## Project Structure

```
cmd/              # Application entrypoint
config/           # Logger, database, email configuration
modules/          # Domain modules (product, order, settlement, etc.)
database/         # Entities, migrations, seeders
pkg/              # Shared utilities and DTOs
script/           # CLI command wiring
```

## Prerequisites

- Go 1.22+
- PostgreSQL 14+
- Docker & Docker Compose (optional for containerized setup)

Create a `.env` file (copy from `.env.example`) and adjust values to your environment:

```bash
cp .env.example .env
```

## Latest Test Results

- **Command**

  ```bash
  make test-order
  ```

- **Output**

  ```text
  🔧 Loading .env file...
  go test -v ./modules/order/tests/...
  === RUN   TestConcurrentBuyers500
  --- PASS: TestConcurrentBuyers500 (0.28s)
  PASS
  ok  	github.com/xkillx/go-gin-order-settlement/modules/order/tests	(cached)
  ```

- **Command**

  ```bash
  make test-settlement
  ```

- **Output**

  ```text
  🔧 Loading .env file...
  go test -v ./modules/settlement/tests/...
  === RUN   TestSettlementCreateJob
  2025/10/01 21:31:23 transaction seeder: rows=3000 merchants=10 days=3 batch=1000
  2025/10/01 21:31:23 transaction seeder done: inserted=3000
  --- PASS: TestSettlementCreateJob (0.38s)
  === RUN   TestSettlementGetJob
  2025/10/01 21:31:24 transaction seeder: rows=3000 merchants=10 days=3 batch=1000
  2025/10/01 21:31:24 transaction seeder done: inserted=3000
  --- PASS: TestSettlementGetJob (0.45s)
  === RUN   TestSettlementCancelJob
  2025/10/01 21:31:24 transaction seeder: rows=3000 merchants=10 days=3 batch=1000
  2025/10/01 21:31:24 transaction seeder done: inserted=3000
  --- PASS: TestSettlementCancelJob (0.46s)
  PASS
  ok  	github.com/xkillx/go-gin-order-settlement/modules/settlement/tests	(cached)
  ```

## Running the Application

Common workflows are captured in the `Makefile`. All commands run from the repository root.

- `make dep` – install and tidy Go module dependencies.
- `make run` – run the API locally using `cmd/main.go`.
- `make build` – produce a binary at `./main`.
- `make run-build` – build and execute the compiled binary.
- `make migrate-local` / `make seed-local` / `make migrate-seed-local` – run migrations and seed data against your local database.

### Dockerized Environment

- `make init-docker` – build images and start app and PostgreSQL containers in the background.
- `make up` / `make down` – start or stop the Compose stack without rebuilding.
- `make logs` – tail logs from all services.
- `make container-go` / `make container-postgres` – open a shell inside the respective containers.

### Database Utilities (Docker)

- `make create-db` – create the database inside the PostgreSQL container.
- `make init-uuid` – enable the `uuid-ossp` extension.
- `make migrate` / `make seed` / `make migrate-seed` – run migrations and seeders from inside the Go container.

## API Overview

### Authentication

All endpoints shown below are unauthenticated out of the box. Add middleware under `middlewares/` to integrate auth as needed.

### Base URL

`http://localhost:8888` (default when `APP_ENV=localhost` and `GOLANG_PORT=8888`).

### Product APIs

| Method | Path | Description |
| --- | --- | --- |
| GET | `/api/products` | Paginated list of products. Supports `page` and `size` query params. |
| GET | `/api/products/:id` | Retrieve product details by ID. |
| POST | `/api/products` | Create a product. Expects payload matching `modules/product/dto/product_request.go`. |
| PUT | `/api/products/:id` | Update an existing product. |
| DELETE | `/api/products/:id` | Remove a product.

### Order APIs

| Method | Path | Description |
| --- | --- | --- |
| GET | `/api/orders` | Paginated list of orders. Accepts pagination query params. |
| GET | `/api/orders/:id` | Retrieve order details by ID. |
| POST | `/api/orders` | Create an order. Validates stock availability. |
| DELETE | `/api/orders/:id` | Delete an order.

### Settlement Job APIs

| Method | Path | Description |
| --- | --- | --- |
| POST | `/jobs/settlement` | Start a settlement job for a date range `{ "from": "YYYY-MM-DD", "to": "YYYY-MM-DD" }`. Returns `job_id`. |
| GET | `/jobs/:id` | Check job status and progress. When completed, includes `download_url`. |
| POST | `/jobs/:id/cancel` | Request cancellation for a running job. |
| GET | `/downloads/:job_id.csv` | Download the generated settlement CSV.

## Testing

The project includes Go test suites under `modules/*/tests/`.

- `make test` – run integration tests located in `./tests` (if present).
- `make test-order` – execute order module tests.
- `make test-settlement` – execute settlement module tests (uses a real PostgreSQL instance; set env vars accordingly).
- `make test-all` – run all module test suites.
- `make test-coverage` – generate coverage profile (`coverage.out`) and open the report in a browser.

When running settlement tests locally, ensure PostgreSQL is available and environment variables (`DB_HOST`, `DB_USER`, `DB_PASS`, `DB_NAME`, `DB_PORT`, `DB_SSLMODE`) are configured. The tests default to `localhost` values if unset.

## Scripts & Automation

- `make module name=<module_name>` calls `./create_module.sh` to scaffold a new module.
- `make run -- --migrate` can be used to run the server with migration flags via `cmd/main.go --migrate`.
- Additional CLI commands are defined under `script/` and activated when running `cmd/main.go` with arguments.

## Contributing

1. Fork the repository and create a feature branch.
2. Run `make test-all` to ensure tests pass.
3. Submit a pull request following the guidelines in `.github/pull_request_template.md`.

## License

This repository is licensed under the MIT License. See `LICENSE` for details.
