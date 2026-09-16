# Project Summary (tk-bank)

## 1. Executive Summary

**What the project is:**
tk-bank is a RESTful banking API built with Go, following the classic "simple bank" tutorial architecture. It provides core banking operations including account management with money transfers between accounts.

**What is already working/implemented:**
- Database layer with PostgreSQL using SQLC for type-safe SQL
- Core domain: Accounts, Entries (ledger), and Transfers
- Full CRUD operations for accounts via REST API (Gin framework)
- Transactional money transfers with deadlock prevention
- Comprehensive integration tests for database layer and API handlers
- CI pipeline with GitHub Actions (postgres service, migrations, tests)
- Mock-based unit testing for API layer using gomock

**What is missing / next milestones:**
- Authentication (JWT/PASETO) – planned but not implemented
- User management endpoints
- gRPC API layer
- Kubernetes deployment configuration
- Background workers (Asynq + Redis)
- Structured logging and observability
- Entry and Transfer API endpoints (only Account endpoints exist)

---

## 2. Tech Stack

| Category | Technology | Version/Details |
|----------|------------|-----------------|
| **Language** | Go | 1.25 (per `go.mod`) |
| **Web Framework** | Gin | v1.10.0 |
| **Database** | PostgreSQL | 16-alpine (Docker) |
| **Database Driver** | pgx/v5 | v5.6.0 |
| **SQL Code Generation** | SQLC | v1.30.0 (per generated files) |
| **Migrations** | golang-migrate | v4.18.2 (installed in CI) |
| **Testing** | testify | v1.11.1 |
| **Mocking** | gomock | v1.6.0 |
| **Configuration** | Viper | v1.21.0 |
| **Containerization** | Docker | postgres:16-alpine image |

---

## 3. Repository Structure

```
tk-bank/
├── main.go                    # Application entry point
├── app.env                    # Environment configuration (DB_SOURCE, SERVER_ADDRESS)
├── go.mod / go.sum            # Go module dependencies
├── Makefile                   # Build/run/test commands
├── sqlc.yaml                  # SQLC configuration
├── README.md                  # Project overview
├── backlog.md                 # Development backlog (Portuguese)
├── api/
│   ├── server.go              # HTTP server setup and routing
│   ├── account.go             # Account handlers (CRUD)
│   └── account_test.go        # API handler unit tests (gomock)
├── db/
│   ├── migration/
│   │   ├── 000001_init_schema.up.sql    # Initial schema (accounts, entries, transfers)
│   │   └── 000001_init_schema.down.sql  # Rollback migration
│   ├── query/
│   │   ├── account.sql        # Account SQL queries
│   │   ├── entry.sql          # Entry SQL queries
│   │   └── transfers.sql      # Transfer SQL queries
│   ├── sqlc/
│   │   ├── account.sql.go     # Generated account operations
│   │   ├── entry.sql.go       # Generated entry operations
│   │   ├── transfers.sql.go   # Generated transfer operations
│   │   ├── models.go          # Generated Go structs
│   │   ├── db.go              # DBTX interface
│   │   ├── querier.go         # Generated Querier interface
│   │   ├── store.go           # Store with TransferTx transaction
│   │   ├── *_test.go          # Integration tests
│   │   └── main_test.go       # Test setup/teardown
│   └── mock/
│       └── store.go           # MockStore (gomock generated)
├── util/
│   ├── config.go              # Viper configuration loader
│   └── random.go              # Random data generators for tests
└── .github/workflows/
    └── ci.yml                 # GitHub Actions CI pipeline
```

---

## 4. Implemented Features

### Domains/Use-Cases

| Domain | Status | Description |
|--------|--------|-------------|
| **Accounts** | ✅ Implemented | Full CRUD with REST API |
| **Entries** | ✅ DB Layer Only | Ledger entries created during transfers |
| **Transfers** | ✅ DB Layer Only | Transfer transaction logic implemented |

### REST API Endpoints

| Method | Endpoint | Handler | Description |
|--------|----------|---------|-------------|
| `POST` | `/accounts` | `createAccount` | Create new account (owner, currency) |
| `GET` | `/accounts/:id` | `getAccount` | Get single account by ID |
| `GET` | `/accounts` | `listAccount` | List accounts with pagination |
| `PUT` | `/accounts/:id` | `updateAccount` | Update account balance |
| `DELETE` | `/accounts/:id` | `deleteAccount` | Delete account |

**See:** [api/server.go](file:///home/tkosudi/projects/golang/tk-bank/api/server.go#L20-L25)

### Implementation Notes

- **Validation:** Uses Gin's built-in validator (currency must be `EUR`, `USD`, or `BRL`)
- **Error Handling:** Checks for `pgx.ErrNoRows` to return 404; other errors return 500
- **Balance initialization:** New accounts start with balance = 0
- **Pagination:** `page_id` and `page_size` parameters for listing accounts

---

## 5. Database Design

### Tables

```
┌──────────────────┐       ┌──────────────────┐       ┌──────────────────┐
│     accounts     │       │     entries      │       │    transfers     │
├──────────────────┤       ├──────────────────┤       ├──────────────────┤
│ id (PK)          │◄──────│ account_id (FK)  │       │ id (PK)          │
│ owner            │       │ id (PK)          │       │ from_account_id  │──┐
│ balance          │       │ amount           │       │ to_account_id    │──┤
│ currency         │       │ created_at       │       │ amount           │  │
│ created_at       │       └──────────────────┘       │ created_at       │  │
└──────────────────┘                                  └──────────────────┘  │
        ▲                                                                   │
        └───────────────────────────────────────────────────────────────────┘
```

**See:** [db/migration/000001_init_schema.up.sql](file:///home/tkosudi/projects/golang/tk-bank/db/migration/000001_init_schema.up.sql)

### Indexes

| Table | Index |
|-------|-------|
| `accounts` | `owner` |
| `entries` | `account_id` |
| `transfers` | `from_account_id`, `to_account_id`, `(from_account_id, to_account_id)` |

### Transactions & Locking

The `TransferTx` function in [db/sqlc/store.go](file:///home/tkosudi/projects/golang/tk-bank/db/sqlc/store.go#L66-L109) implements:

1. Creates transfer record
2. Creates two entries (debit and credit)
3. Updates both account balances atomically
4. **Deadlock prevention:** Accounts are updated in consistent order (lower ID first)
5. Uses `FOR NO KEY UPDATE` locking for account reads during updates

### Migration Commands

```bash
# Apply all migrations
make migrateup

# Rollback all migrations
make migratedown

# Full database reset
make resetdb
```

---

## 6. Tests

### Test Types

| Type | Location | Description |
|------|----------|-------------|
| **Integration (DB)** | `db/sqlc/*_test.go` | Tests against real PostgreSQL |
| **Unit (API)** | `api/account_test.go` | Mock-based handler tests |

### Test Files Summary

| File | Tests |
|------|-------|
| `db/sqlc/account_test.go` | `TestCreateAccount`, `TestGetAccount`, `TestUpdateAccount`, `TestDeleteAccount`, `TestListAccounts` |
| `db/sqlc/entry_test.go` | Entry CRUD tests |
| `db/sqlc/transfers_test.go` | Transfer CRUD tests |
| `db/sqlc/store_test.go` | `TestTransferTx`, `TestTransferTxDeadlock` (concurrent transfer testing) |
| `api/account_test.go` | `TestGetAccountApi`, `TestCreateAccountAPI` (table-driven, gomock) |

### Test Execution

```bash
# Run all tests with coverage
make test

# Equivalent to:
go test -v -cover ./...
```

**Prerequisites:** Running PostgreSQL with migrations applied.

### Mocking Strategy

- **gomock** generates mocks from the `Store` interface
- Mock location: `db/mock/store.go`
- API tests use `httptest.NewRecorder()` with mock store injection

### Coverage Observations

- ✅ Good coverage of account CRUD operations
- ✅ Concurrent transfer tests verify deadlock prevention
- ⚠️ No tests for entry/transfer API endpoints (not implemented yet)
- ⚠️ API tests don't cover `updateAccount` or `deleteAccount` handlers
- ⚠️ No tests for error paths in configuration loading

### Recommended Tests to Add

1. Tests for `listAccount`, `updateAccount`, `deleteAccount` API handlers
2. Edge cases: negative amounts, same from/to account
3. Configuration loading tests
4. End-to-end integration tests

---

## 7. CI/CD Pipeline

**File:** [.github/workflows/ci.yml](file:///home/tkosudi/projects/golang/tk-bank/.github/workflows/ci.yml)

### Triggers

- Push to `dev` or `main` branches
- Pull requests to `main` branch

### Pipeline Steps

| Step | Description |
|------|-------------|
| 1. Checkout | `actions/checkout@v4` |
| 2. Setup Go | `actions/setup-go@v5` with `go.mod` version, caching enabled |
| 3. Install golang-migrate | Downloads v4.18.2 from GitHub releases |
| 4. Create app.env | Writes secrets to config file |
| 5. Run Migrations | `make migrateup` |
| 6. Test | `make test` |

### Database Service

PostgreSQL 16-alpine runs as a GitHub Actions service with health checks:
```yaml
services:
  postgres:
    image: postgres:16-alpine
    env:
      POSTGRES_DB: tk_bank
      POSTGRES_USER: root
      POSTGRES_PASSWORD: secret
```

### Concurrency

```yaml
concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true
```

### CI Improvements Recommended

| Priority | Improvement |
|----------|-------------|
| **P0** | Add secrets `DB_SOURCE` and `SERVER_ADDRESS` (required for tests) |
| **P1** | Add linting step (`golangci-lint`) |
| **P1** | Add `go vet` and `go fmt` checks |
| **P2** | Cache golang-migrate installation |
| **P2** | Add security scanning (gosec, govulncheck) |
| **P3** | Generate coverage reports and upload to Codecov |

---

## 8. Local Development Guide

### Prerequisites

- Go 1.25+
- Docker
- make
- golang-migrate CLI

### Step-by-Step Setup

```bash
# 1. Start PostgreSQL container
make postgres

# 2. Create database
make createdb

# 3. Run migrations
make migrateup

# 4. Start the API server
make server
# Server runs at http://0.0.0.0:8080

# 5. Run tests (in another terminal)
make test
```

### Makefile Commands

| Command | Description |
|---------|-------------|
| `make postgres` | Start PostgreSQL 16 container |
| `make createdb` | Create `tk_bank` database |
| `make dropdb` | Drop `tk_bank` database |
| `make migrateup` | Apply all migrations |
| `make migratedown` | Rollback all migrations |
| `make resetdb` | Drop + create + migrate |
| `make sqlc` | Regenerate SQLC code |
| `make test` | Run all tests with coverage |
| `make server` | Start the API server |

### Environment Variables

Defined in `app.env`:

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_SOURCE` | `postgresql://root:secret@127.0.0.1:5432/tk_bank?sslmode=disable` | PostgreSQL connection string |
| `SERVER_ADDRESS` | `0.0.0.0:8080` | HTTP server bind address |

---

## 9. Notable Engineering Decisions

### 1. SQLC for Type-Safe SQL

The project uses SQLC instead of an ORM. This provides:
- Compile-time type checking
- No runtime reflection
- Explicit SQL control

**Evidence:** [sqlc.yaml](file:///home/tkosudi/projects/golang/tk-bank/sqlc.yaml) configured with `pgx/v5` driver, JSON tags, and interface generation.

### 2. Store Interface for Testability

The `Store` interface combines `Querier` (generated) with custom transaction methods:

```go
type Store interface {
    Querier
    TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error)
}
```

This enables mock injection for API handler tests.

**Evidence:** [db/sqlc/store.go](file:///home/tkosudi/projects/golang/tk-bank/db/sqlc/store.go#L11-L14)

### 3. Deadlock Prevention via Ordering

The `TransferTx` ensures accounts are always updated in ID order:

```go
if arg.FromAccountID < arg.ToAccountID {
    result.FromAccount, result.ToAccount, err = addMoney(...)
} else {
    result.ToAccount, result.FromAccount, err = addMoney(...)
}
```

This prevents database deadlocks during concurrent bi-directional transfers.

**Evidence:** [db/sqlc/store.go#L99-L103](file:///home/tkosudi/projects/golang/tk-bank/db/sqlc/store.go#L99-L103) and verified in `TestTransferTxDeadlock`.

### 4. Viper for Configuration

Configuration is loaded from `app.env` with automatic environment variable override via Viper.

**Evidence:** [util/config.go](file:///home/tkosudi/projects/golang/tk-bank/util/config.go)

### 5. pgxpool Connection Pooling

Uses `pgxpool.Pool` for connection management, suitable for concurrent API requests.

**Evidence:** [main.go#L20](file:///home/tkosudi/projects/golang/tk-bank/main.go#L20)

---

## 10. Open Questions / Gaps

### Documentation Gaps

- [ ] README installation instructions incomplete (no test section)
- [ ] No API documentation (OpenAPI/Swagger)
- [ ] Backlog is in Portuguese, mixed language

### Incomplete Features

- [ ] Entry API endpoints not exposed
- [ ] Transfer API endpoints not exposed
- [ ] User management not implemented
- [ ] Authentication not implemented

### Code Quality Issues

- [ ] Typo in `createAccountRequest`: `biding` should be `binding` (line 13 in `api/account.go`)
- [ ] Typo in migration comment: "must bem positive" should be "must be positive"
- [ ] No structured logging (uses `log.Fatal`)
- [ ] `updateAccount` allows arbitrary balance changes (security concern)
- [ ] No rate limiting or request validation middleware

### Risks & Tech Debt

- CI requires secrets (`DB_SOURCE`, `SERVER_ADDRESS`) that may not be configured
- No linting in CI pipeline
- `math/rand` used instead of `crypto/rand` (though acceptable for test utilities)
- Tests leave data in database (no cleanup)

---

## 11. Next Steps (Prioritized)

### P0 – Critical / Immediate

1. **Configure CI secrets** – Add `DB_SOURCE` and `SERVER_ADDRESS` secrets to GitHub repository
2. **Fix typo** in `createAccountRequest` (`biding` → `binding`)
3. **Add Entry/Transfer API endpoints** – Expose existing database functionality

### P1 – High Priority

4. **Implement User entity and authentication** – Add JWT/PASETO as per README roadmap
5. **Add linting to CI** – Integrate `golangci-lint` workflow step
6. **Complete API handler tests** – Add tests for `updateAccount`, `deleteAccount`, `listAccount`
7. **Add OpenAPI/Swagger documentation**

### P2 – Medium Priority

8. **Implement gRPC API layer** – As per README roadmap
9. **Add structured logging** – Replace `log.Fatal` with zerolog or similar
10. **Add test cleanup** – Ensure tests clean up created data
11. **Security hardening** – Rate limiting, input sanitization

### P3 – Nice to Have

12. **Kubernetes deployment** – Create Helm charts or Kustomize manifests
13. **Background workers** – Integrate Asynq + Redis for async processing
14. **Observability** – Add Prometheus metrics, distributed tracing
15. **CI coverage reports** – Upload to Codecov/Coveralls

---

*Generated: 2026-01-20 | Based on repository analysis*
