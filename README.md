# Snippy

A URL shortener REST API written in Go. Accepts a long URL, generates a cryptographically random base62 short code, stores it in PostgreSQL, and redirects clients on lookup. Includes a click-tracking analytics endpoint.

## Features

- Shorten any URL to a 6-character code
- 302 redirect on code lookup with async click tracking
- Per-link click analytics
- Multi-stage Docker build
- Integration tests using testcontainers (real Postgres, no mocks)

## Tech stack

- **Go 1.25** — `net/http` with 1.22 method+path routing
- **PostgreSQL 16** — pgx/v5 connection pool
- **Docker / docker-compose** — local development and containerized deployment
- **testcontainers-go** — spins up a real Postgres instance for store integration tests

## API

| Method | Path | Body / Response |
|--------|------|-----------------|
| `POST` | `/links` | `{"url": "https://..."}` → `201` with link object |
| `GET` | `/{code}` | `302` redirect to original URL |
| `GET` | `/stats/{code}` | `{"short_code": "...", "clicks": N}` |

## Running locally

**With Docker (recommended)**

```bash
cp .env.example .env
docker compose up --build
```

Postgres is initialized automatically from `migrations/001_create_links.sql`. The API is available at `http://localhost:8080`.

**Without Docker**

```bash
cp .env.example .env
# start Postgres separately, apply the migration, then:
go run ./cmd/api
```

**Environment variables**

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | — | Postgres DSN (required) |
| `PORT` | `8080` | Port the server listens on |

## Example usage

```bash
# Shorten a URL
curl -X POST http://localhost:8080/links \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}'
# {"id":1,"short_code":"aB3xYz","long_url":"https://example.com","created_at":"..."}

# Follow the short link (redirects to original URL)
curl -L http://localhost:8080/aB3xYz

# Check click stats
curl http://localhost:8080/stats/aB3xYz
# {"short_code":"aB3xYz","clicks":1}
```

## Project structure

```
cmd/api/            server entry point and router
internal/
  model/            Link struct
  db/               pgxpool connection + Store interface and implementation
  handler/          HTTP handlers (CreateLink, Resolve, Stats)
  shortcode/        base62 short-code generator
migrations/         SQL schema
```

## Short code generation

Codes are generated with `crypto/rand` rather than `math/rand` so they are unpredictable — a predictable RNG would let an attacker enumerate all shortened links. The base62 alphabet (`a-z A-Z 0-9`) at length 6 produces ~56 billion possible codes.

## Tests

```bash
go test ./...
```

- `internal/shortcode` — unit tests for the generator (uniqueness, length, alphabet)
- `internal/handler` — unit tests with a mock store (no database required)
- `internal/db` — integration test using testcontainers (requires Docker)
