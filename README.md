# DJBook Backend

Go + GraphQL backend for the DJBook iOS app — a booking and event management tool for DJs.

## Stack

- **Go 1.22** — HTTP server and business logic
- **gqlgen** — Code-first GraphQL with type-safe resolvers  
- **MySQL 8.0** — Primary database (via `go-sql-driver/mysql`)
- **Redis 7** — JWT token deny-list and caching
- **Docker Compose** — Local development environment

## Quick Start

### 1. Prerequisites

- Go 1.22+
- Docker & Docker Compose

### 2. Setup

```bash
# Clone and enter the project
cd djbook

# Create .env from template and fill in your values
cp .env.example .env

# Download Go dependencies
go mod tidy

# Start MySQL + Redis containers
docker compose up mysql redis -d

# Run the server locally
go run ./cmd/server
```

### 3. Docker (full stack)

```bash
docker compose up --build
```

The GraphQL playground will be available at **http://localhost:8080/**.

## Project Layout

```
djbook-backend/
├── cmd/server/main.go          # Entry point
├── internal/
│   ├── auth/                   # Google/Apple Sign In + JWT middleware
│   ├── graph/                  # GraphQL schema, generated executor, resolvers
│   │   ├── schema.graphqls
│   │   ├── generated.go        # gqlgen-generated executor (or re-generate below)
│   │   ├── model/              # GraphQL types and enums
│   │   └── resolvers/          # Resolver implementations
│   ├── model/                  # Database/domain models
│   ├── repository/             # SQL repositories
│   └── service/                # Business logic
└── migrations/                 # SQL migrations (run on startup)
```

## Regenerating GraphQL Code

If you update `internal/graph/schema.graphqls`, regenerate the executor:

```bash
make generate
# or
go run github.com/99designs/gqlgen generate
```

## Auth Flow

```
iOS App → POST /auth/google  { id_token: "..." }
       ← { token: "JWT...", userId: "uuid" }

iOS App → POST /auth/apple   { id_token: "..." }
       ← { token: "JWT...", userId: "uuid" }

# All GraphQL requests:
Authorization: Bearer <JWT>
POST /query  { "query": "..." }
```

## Key Design Decisions

- **Repository pattern** — SQL stays in `repository/`, services stay clean
- **Ownership checks** — Every mutation verifies the JWT user owns the resource
- **Event status history** — Automatically tracked on every status change
- **Finance queries** — Computed with single MySQL aggregation queries (no Go loops)
- **Upcoming events** — `date >= TODAY AND status IN (PENDING, CONFIRMED)` at DB level
- **JWT deny-list** — Redis stores revoked tokens until their natural expiry
