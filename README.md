# gocars-api

REST API for [gocars.mn](https://gocars.mn) — Mongolian automotive parts marketplace.

| | |
|---|---|
| **Language** | Go 1.25 |
| **Framework** | Gin |
| **Database** | PostgreSQL (Supabase) |
| **Port** | 9000 |

## Related repos

| Repo | Purpose |
|---|---|
| [gocars-local](https://github.com/Shinekhuu/gocars-local) | Local dev environment (Docker Compose, env files) |
| [auth-service](https://github.com/Shinekhuu/auth-service) | Authentication microservice (Supabase wrapper, port 9001) |

---

## Prerequisites

- Go 1.25+
- PostgreSQL (Supabase project or local instance)
- Redis
- [air](https://github.com/air-verse/air) — hot reload in dev (`go install github.com/air-verse/air@latest`)

---

## Local development

### 1. Clone and set up env

```bash
git clone https://github.com/Shinekhuu/gocars-local
cd gocars-local
cp env/api.env.example env/dev/api.env
cp env/auth.env.example env/dev/auth.env
cp env/infra.env.example env/dev/infra.env
# Fill in credentials in env/dev/*.env
```

### 2. Start infrastructure + services via Docker

```bash
cd gocars-local
make up              # starts meilisearch, gocars-auth, gocars-api (hot reload via air)
make logs            # tail all logs
make down            # stop everything
```

### 3. Run gocars-api directly (without Docker)

```bash
export MODE=DEVELOPMENT
export PG_HOST=...   # set remaining vars
go run ./cmd/api
```

> **Meilisearch**: when running outside Docker set `MEILI_URL=http://localhost:7700`.  
> Inside Docker the hostname resolves as `http://meilisearch:7700`.

---

## Build

```bash
go build -o gocars-api ./cmd/api
```

### Docker

```bash
docker build -t gocars-api .
docker run -p 9000:9000 --env-file ./env/dev/api.env gocars-api
```

---

## Environment variables

| Variable | Purpose |
|---|---|
| `MODE` | `PRODUCTION` or `DEVELOPMENT` — controls Gin mode and AutoMigrate |
| `PG_HOST` | PostgreSQL host (Supabase connection pooler) |
| `PG_PORT` | PostgreSQL port (`6543` for Supabase pooler) |
| `PG_USER` | PostgreSQL user |
| `PG_PASSWORD` | PostgreSQL password |
| `PG_NAME` | Database name |
| `PG_SSL_MODE` | `require` for Supabase |
| `REDIS_ADDR` | Redis address (`host:port`) |
| `REDIS_PASSWORD` | Redis password |
| `AUTH_API_KEY` | Outgoing service-to-service key (gocars-api → gocars-auth) |
| `MEILI_URL` | Meilisearch URL |
| `MEILI_MASTER_KEY` | Meilisearch master key |
| `X_RAPIDAPI_KEY` | RapidAPI key (TecDoc parts catalog) |
| `X_RAPIDAPI_HOST` | RapidAPI host |
| `OPENAI_API_KEY` | OpenAI key (GPT-4.1-mini, parts AI + TecDoc chassis mapping) |
| `GARAGE_HOST` | Web scraper API base URL |
| `GARAGE_KEY` | Web API key |
| `SENTRY_DSN` | Sentry error tracking DSN |

---

## API

| Method | Path | Description |
|---|---|---|
| GET | `/health` | Liveness check |
| GET | `/vehicle?plate_number=` | Plate lookup → vehicle info + TecDoc match |
| GET | `/search?query=&vehicle_id=&page=&limit=` | Part search (Meili → DB → RapidAPI fallback) |
| GET | `/shop?vehicle_id=&category_id=&page=&limit=` | Browse parts by vehicle + category |
| GET | `/article?id=&article_id=` | Single part detail |
| GET | `/products?search=&category_id=&page=&limit=` | Product catalog |
| GET | `/oems?id[]=&article_id[]=` | OEM number lookup |
| GET | `/manufacturers` | All TecDoc manufacturers |
| GET | `/model?manufacturer_id=` | Models for a manufacturer |
| GET | `/engine?manufacturer_id=&model_id=` | Engines for a model |
| POST | `/order` | Create order + invoice |
| GET | `/order/:id` | Order summary |
| GET | `/orders/:id/pdf` | Order PDF download |
| GET | `/profile` | Current user profile *(auth required)* |
| PUT | `/profile` | Update profile *(auth required)* |

### Auth

Protected routes require nginx ingress annotations:

```yaml
nginx.ingress.kubernetes.io/auth-url: "http://gocars-auth-svc:9001/auth/validate"
nginx.ingress.kubernetes.io/auth-response-headers: "X-User-ID,X-User-Email"
```

gocars-api does no JWT parsing — it reads `X-User-ID` and `X-User-Email` headers forwarded by nginx after gocars-auth validates the token.

---

## Architecture

```
cmd/api/main.go          ← entry point
internal/
  config/                ← env → Config struct (no os.Getenv outside here)
  database/postgres/     ← pgx + GORM (simple protocol for Supabase PgBouncer)
  app/app.go             ← DI wiring: repos → services → handlers
  server/server.go       ← routes, CORS, middleware
  middleware/auth.go     ← reads X-User-ID / X-User-Email from nginx headers
  articles/              ← parts catalog domain
  vehicle/               ← plate lookup, TecDoc, scrapers, VIN decoder
  roder/                 ← orders and invoices
  profile/               ← user profiles
  search/meili/          ← Meilisearch client
  cache/redis/           ← Redis client
scripts/
  language_worker.go     ← CLI: sync part name translations from XLS
  xls_worker.go          ← CLI: import parts catalog from XLS
```

**Layer rules:** `model/` structs only → `repository/` DB queries → `service/` business logic → `handler/` HTTP only.

---

## AutoMigrate

Runs on every startup regardless of `MODE`.

---

## CLI workers

```bash
go run ./cmd/api --commands-sync   # runs the language sync worker, not the HTTP server
```
