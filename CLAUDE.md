# GoCars API — CLAUDE.md

## Related Repos

| Repo | Purpose |
|---|---|
| [`Shinekhuu/gocars-local`](https://github.com/Shinekhuu/gocars-local) | Local development environment (Docker, env files, tooling) |
| [`Shinekhuu/auth-service`](https://github.com/Shinekhuu/auth-service) | gocars-auth — authentication microservice (Supabase wrapper) |

---

## Project Overview

REST API for **gocars.mn**, a Mongolian automotive parts marketplace. Core flow:
1. User provides a plate number → XYP API (Mongolia vehicle registry) returns make/model/year
2. Vehicle is matched to a TecDoc `vehicle_id` via scraping or DB lookup
3. Parts are searched by `vehicle_id` + query against local DB, falling back to RapidAPI
4. Orders are created and invoiced; PDFs generated from HTML templates

Server listens on **port 9000**.

---

## Commands

```bash
go run ./cmd/api          # start (loads .env if present)
go build -o gocars-api ./cmd/api

# CLI sync mode (language worker, not the HTTP server)
go run ./cmd/api --commands-sync

# Infra (from project-go/infra/)
make up ENV=dev           # docker compose up
make down ENV=dev
```

---

## Environment Variables

Config is loaded from env vars (or `.env` in dev). All vars live in `/dev|stagenving|production/api.env`.

| Variable | Purpose |
|---|---|
| `MODE` | `PRODUCTION` or `DEVELOPMENT` — controls gin mode and AutoMigrate |
| `PG_HOST`, `PG_PORT`, `PG_USER`, `PG_PASSWORD`, `PG_NAME`, `PG_SSL_MODE` | PostgreSQL (Supabase) |
| `REDIS_ADDR`, `REDIS_PASSWORD` | Redis |
| `AUTH_API_KEY` | Outgoing service-to-service key (gocars-api → gocars-auth) |
| `MEILI_URL`, `MEILI_MASTER_KEY` | Meilisearch |
| `X_RAPIDAPI_KEY`, `X_RAPIDAPI_HOST` | RapidAPI TecDoc catalog |
| `OPENAI_API_KEY` | OpenAI GPT-4.1-mini (parts AI + TecDoc mapping) |
| `GARAGE_HOST`, `GARAGE_KEY` | Garage.mn scraper API |
| `SENTRY_DSN` | Sentry error tracking |

**AutoMigrate** runs only when `cfg.MODE == "PRODUCTION"` (`server.go`). Skipped in `DEVELOPMENT` and `STAGING`.

**PostgreSQL note:** Uses Supabase connection pooler (PgBouncer). pgx is configured with `QueryExecModeSimpleProtocol` to disable prepared statements — required to avoid `SQLSTATE 42P05` errors through the pooler.

---

## Architecture

```
cmd/api/main.go                     ← entry: config, logger, sentry, pgdb, redis, meili, app.NewApp
internal/
  config/config.go                  ← env vars → Config struct (no os.Getenv outside here)
  database/postgres/db.go           ← pgx + GORM init, global pgdb.DB
  app/app.go                        ← DI wiring: repos → services → handlers
  server/server.go                  ← Gin router, CORS, middleware, routes
  server/migrate.go                 ← AutoMigrate (PostgreSQL, idempotent)
  middleware/auth.go                ← reads X-User-ID / X-User-Email from nginx headers
  articles/
    repository/postgresql/model/    ← GORM structs only (no DB/HTTP logic)
    repository/postgresql/          ← DB queries via injected *gorm.DB
    service/                        ← business logic, RapidAPI, OpenAI calls
    handler/http/                   ← thin HTTP handlers
    jobs/                           ← background worker (ArticleQueue)
  vehicle/
    repository/postgresql/model/    ← GORM structs only
    repository/postgresql/          ← DB queries
    repository/                     ← scraper repos (garage, partsouq, toyodiy)
    service/                        ← engine/model/vehicle/vin/crawler services
    handler/                        ← thin HTTP handlers
  roder/                            ← orders domain (repo / service / handler / dto)
  profile/                          ← profile domain (repo / service / handler)
  shared/utils/                     ← pure helpers (string, pointer, OEM normalization)
  search/meili/                     ← Meilisearch client
  cache/redis/                      ← Redis init
  logger/                           ← zap logger
scripts/
  language_worker.go                ← CLI: sync translations from XLS (uses pgdb.DB)
  xls_worker.go                     ← CLI: import parts from XLS (uses pgdb.DB)
```

---

## Layer Rules

| Layer | Rule |
|---|---|
| `model/` | Structs only — no DB calls, no HTTP calls, no business logic |
| `repository/` | DB queries only — injected `*gorm.DB`, no HTTP |
| `service/` | Business logic + external API calls — no `c *gin.Context` |
| `handler/` | HTTP only — parse request, call service, return JSON |
| `jobs/` | Background goroutines — use package-level `gdb *gorm.DB` set by `StartWorker(db)` |

Never call `os.Getenv` outside `config/config.go`.

---

## Auth

User authentication is handled by **nginx ingress** + **gocars-auth** (port 9001):

```
Client request
  → nginx → GET gocars-auth/auth/validate (Authorization: Bearer <JWT>)
           → 200 + X-User-ID / X-User-Email  →  nginx forwards to gocars-api
           → 401  →  nginx rejects
```

`middleware/auth.go` reads `X-User-ID` and `X-User-Email` from request headers set by nginx. gocars-api does **no** JWT parsing.

**K8s ingress annotations** (on protected routes):
```yaml
nginx.ingress.kubernetes.io/auth-url: "http://gocars-auth-svc:9001/auth/validate"
nginx.ingress.kubernetes.io/auth-response-headers: "X-User-ID,X-User-Email"
```

`AUTH_API_KEY` is used by gocars-api when it needs to call gocars-auth endpoints directly (outgoing, service-to-service). Not used for validating incoming requests.

---

## Routes

| Method | Path | Handler | Auth |
|---|---|---|---|
| GET | `/health` | inline | — |
| GET | `/manufacturers` | `ManufacturerHdl.GetManufacturers` | — |
| GET | `/model` | `ModelHdl.GetModels` | — |
| GET | `/engine` | `EngineHdl.GetEngines` | — |
| GET | `/vehicle` | `VehicleHdl.FetchData` | — |
| GET | `/search` | `SearchHdl.Search` | — |
| GET | `/shop` | `ShopHdl.Shop` | — |
| GET | `/oems` | `OemHdl.GetOEMs` | — |
| GET | `/article` | `ArticleHdl.Article` | — |
| GET | `/products` | `ProductHdl.GetProducts` | — |
| POST | `/order` | `OrderHdl.CreateOrder` | — |
| GET | `/order/:id` | `OrderHdl.GetOrder` | — |
| GET | `/orders/:id/pdf` | `OrderHdl.GetOrderPDF` | — |
| GET | `/profile` | `ProfileHdl.Profile` | nginx |
| PUT | `/profile` | `ProfileHdl.UpdateProfile` | nginx |

---

## Background Worker

`jobs.StartWorker(pgdb.DB)` (called from `server.go`) starts a goroutine consuming `jobs.ArticleQueue` (buffered, size 100).

Flow: `processArticle` → `saveMain` → `saveEngines` (PostgreSQL `ON CONFLICT DO NOTHING`).

To enqueue:
```go
select {
case jobs.ArticleQueue <- article:
default:
    log.Println("queue full, skip")
}
```

---

## Search Flow

`GET /search?query=...&vehicle_id=...&page=1&limit=40`

1. Try **Meilisearch** — if hits, return immediately
2. Fall back to **`ArticleRepository.SearchProducts`** — PostgreSQL full-text query with OEM priority ranking
3. Fall back to **`ArticleService.GetByOemFromRapidAPI`** — HTTP call to RapidAPI, async-saves results to DB

---

## External Integrations

| Service | Package | Purpose |
|---|---|---|
| XYP API (`developer.xyp.gov.mn`) | `vehicle/service/vehicle_service.go` | Plate → vehicle info |
| RapidAPI (`rapidapi.com`) | `articles/service/rapidapi_service.go` | TecDoc parts catalog |
| OpenAI GPT-4.1-mini | `articles/service/openai_service.go` | Parts AI + TecDoc chassis mapping |
| Meilisearch | `search/meili/` | Full-text search |
| Garage.mn scraper | `vehicle/service/crawler_service.go` | Vehicle data scraping |
| Sentry | `cmd/api/main.go` | Error tracking |

---

## Coding Conventions

- Model files contain **structs only** — no functions that touch DB or HTTP
- Repositories take `*gorm.DB` via constructor — never use a global DB var (except `jobs/`)
- Services take repositories via constructor — orchestrate logic, call external APIs
- Handlers are thin — parse → call service → respond
- GORM not-found: `errors.Is(err, gorm.ErrRecordNotFound)` not `err != nil`
- Pointer helpers: `utils.SafeString()`, `utils.StringToUintPtr()`, `utils.UintToIntPtr()`
- OEM normalization: `utils.NormalizeOEM()`, `utils.IsOEM()`
- New routes → `server/server.go`; protected routes under the `authorized` group
- New env vars → add to `Config` struct **and** `Load()` in `config/config.go`
