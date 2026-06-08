# gocars-api

REST API for the GoCars automotive parts marketplace.

|               |                       |
| ------------- | --------------------- |
| **Language**  | Go                    |
| **Framework** | Gin                   |
| **Database**  | PostgreSQL (Supabase) |
| **Port**      | 9000                  |

## Related Repositories

| Repository                                | Purpose                                                           |
| ----------------------------------------- | ----------------------------------------------------------------- |
| https://github.com/Shinekhuu/gocars-local | Local development environment (Docker Compose, environment files) |
| https://github.com/Shinekhuu/auth-service | Authentication and authorization service                          |

---

## Prerequisites

* Go 1.25+
* PostgreSQL (Supabase project or local instance)
* Redis
* Air (hot reload for development)

```bash
go install github.com/air-verse/air@latest
```

---

## Local Development

### 1. Clone and Configure Environment

```bash
git clone https://github.com/Shinekhuu/gocars-local
cd gocars-local

cp env/api.env.example env/dev/api.env
cp env/auth.env.example env/dev/auth.env
cp env/infra.env.example env/dev/infra.env
```

Update the values in `env/dev/*.env`.

### 2. Start Infrastructure and Services

```bash
make up
```

Useful commands:

```bash
make logs
make ps
make down
```

### 3. Run gocars-api Directly

```bash
export MODE=DEVELOPMENT

export PG_HOST=...
export PG_PORT=...
export PG_USER=...
export PG_PASSWORD=...
export PG_NAME=...

go run ./cmd/api
```

> When running outside Docker, set `MEILI_URL=http://localhost:7700`.
>
> When running inside Docker, use `http://meilisearch:7700`.

---

## Build

### Binary

```bash
go build -o gocars-api ./cmd/api
```

### Docker

```bash
docker build -t gocars-api .

docker run \
  -p 9000:9000 \
  --env-file ./env/dev/api.env \
  gocars-api
```

---

## Environment Variables

| Variable           | Purpose                                                             |
| ------------------ | ------------------------------------------------------------------- |
| `MODE`             | Application mode (`PRODUCTION` or `DEVELOPMENT`)                    |
| `PG_HOST`          | PostgreSQL host (Supabase connection pooler)                        |
| `PG_PORT`          | PostgreSQL port (`6543` when using Supabase PgBouncer)              |
| `PG_USER`          | PostgreSQL user                                                     |
| `PG_PASSWORD`      | PostgreSQL password                                                 |
| `PG_NAME`          | Database name                                                       |
| `PG_SSL_MODE`      | SSL mode (`require` for Supabase)                                   |
| `REDIS_ADDR`       | Redis address (`host:port`)                                         |
| `REDIS_PASSWORD`   | Redis password                                                      |
| `AUTH_API_KEY`     | Service-to-service authentication key (`gocars-api → auth-service`) |
| `MEILI_URL`        | Meilisearch URL                                                     |
| `MEILI_MASTER_KEY` | Meilisearch master key                                              |
| `X_RAPIDAPI_KEY`   | RapidAPI key (TecDoc integration)                                   |
| `X_RAPIDAPI_HOST`  | RapidAPI host                                                       |
| `OPENAI_API_KEY`   | OpenAI API key (parts AI and TecDoc mapping)                        |
| `GARAGE_HOST`      | Vehicle Identity Services scraper API base URL                      |
| `GARAGE_KEY`       | Vehicle Identity Services API key                                   |
| `SENTRY_DSN`       | Sentry error tracking                                               |

---

## API Endpoints

| Method | Path                                          | Description                            |
| ------ | --------------------------------------------- | -------------------------------------- |
| GET    | `/health`                                     | Health check                           |
| GET    | `/vehicle?plate_number=`                      | Vehicle lookup and TecDoc matching     |
| GET    | `/search?query=&vehicle_id=&page=&limit=`     | Part search                            |
| GET    | `/shop?vehicle_id=&category_id=&page=&limit=` | Browse parts by vehicle and category   |
| GET    | `/article?id=&article_id=`                    | Part details                           |
| GET    | `/products?search=&category_id=&page=&limit=` | Product catalog                        |
| GET    | `/oems?id[]=&article_id[]=`                   | OEM lookup                             |
| GET    | `/manufacturers`                              | TecDoc manufacturers                   |
| GET    | `/model?manufacturer_id=`                     | Manufacturer models                    |
| GET    | `/engine?manufacturer_id=&model_id=`          | Engine variants                        |
| POST   | `/order`                                      | Create order                           |
| GET    | `/order/:id`                                  | Order details                          |
| GET    | `/orders/:id/pdf`                             | Order PDF                              |
| GET    | `/profile`                                    | Current user profile *(authenticated)* |
| PUT    | `/profile`                                    | Update profile *(authenticated)*       |

---

## Authentication

Protected routes use external authentication through NGINX Ingress.

```yaml
nginx.ingress.kubernetes.io/auth-url: "http://gocars-auth-svc:9001/auth/validate"
nginx.ingress.kubernetes.io/auth-response-headers: "X-User-ID,X-User-Email"
```

Authentication is delegated to `auth-service`.

`gocars-api` does not parse JWT tokens directly. It reads the `X-User-ID` and `X-User-Email` headers forwarded by NGINX after token validation.

---

## Architecture

```text
cmd/api/main.go          ← application entry point

internal/
  config/                ← environment configuration
  database/postgres/     ← pgx + GORM (Supabase PgBouncer compatible)
  app/app.go             ← dependency injection wiring
  server/server.go       ← routes, middleware, CORS
  middleware/auth.go     ← reads X-User-* headers
  articles/              ← parts catalog domain
  vehicle/               ← plate lookup, TecDoc integration, VIN decoding
  order/                 ← orders and invoices
  profile/               ← user profile management
  search/meili/          ← Meilisearch integration
  cache/redis/           ← Redis integration

scripts/
  language_worker.go     ← translation synchronization worker
  xls_worker.go          ← catalog import worker
```

### Layer Rules

```text
Model → Repository → Service → Handler
```

* Models contain data structures.
* Repositories handle database access.
* Services implement business logic.
* Handlers expose HTTP endpoints.

---

## Database Migration

AutoMigrate runs on every application startup regardless of `MODE`.

---

## PostgreSQL (Supabase) Notes

The application uses the Supabase connection pooler (PgBouncer).

`pgx` is configured with `QueryExecModeSimpleProtocol` to disable prepared statements, preventing:

```text
SQLSTATE 42P05
```

errors when using transaction pooling.

---

## CLI Workers

Run background workers directly:

```bash
go run ./cmd/api --commands-sync
```

This executes the language synchronization worker instead of starting the HTTP server.

---

## License

Licensed under the MIT License.
