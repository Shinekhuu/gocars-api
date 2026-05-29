# GoCars API — CLAUDE.md

## Project Overview

REST API for **gocars.mn**, a Mongolian automotive parts marketplace. Core flow:
1. User provides a plate number → system looks up vehicle via XYP API (Mongolia vehicle registry) and/or web scraping
2. Vehicle maps to a TecDoc `vehicle_id`
3. Parts are searched by `vehicle_id` + query string against local DB, falling back to RapidAPI
4. Orders are created and invoiced; PDFs generated from HTML templates

Server listens on **port 9000**.

---

## Commands

```bash
make run      # go run . (production-like, no hot reload)
make dev      # air (hot reload, requires `air` installed)
make build    # go build -o gocars-api .

# CLI sync mode (run workers, not the HTTP server)
go run . --commands-sync

# Start Meilisearch (required for search features)
docker compose up -d
```

---

## Environment Variables

Copy `.env-example` to `.env`. Required vars:

| Variable | Purpose |
|---|---|
| `MODE` | `PRODUCTION` or `DEVELOPMENT` — controls gin mode, DB debug, and AutoMigrate |
| `DB_USER`, `DB_PASSWORD`, `DB_HOST`, `DB_PORT`, `DB_NAME` | MySQL connection |
| `JWT_SECRET` | HMAC secret for JWT signing — **must not be empty** |
| `X_RAPIDAPI_KEY`, `X_RAPIDAPI_HOST` | RapidAPI auto-parts catalog |
| `OPENAI_API_KEY` | OpenAI GPT-4.1-mini for parts AI and TecDoc mapping |

> Note: `.env-example` uses `DATABASE_HOST` style but `config/config.go` reads `DB_HOST` style. Use the `DB_*` style in your actual `.env`.

**AutoMigrate runs only when `MODE != "DEVELOPMENT"`** (checked via `os.Getenv`, not the config struct — see known issues).

---

## Layer Responsibilities

| Layer | Package | Responsibility |
|---|---|---|
| Router | `main.go` | Route registration, middleware wiring |
| Config | `config/` | Load env vars into `Config` struct |
| Database | `database/` | GORM init, global `database.DB` var |
| Models | `models/` | GORM structs + some complex SQL (see known issues) |
| Repositories | `repositories/` | Data access — prefer adding queries here |
| Services | `services/` | Business logic, external API calls |
| Controllers | `controllers/` | HTTP request/response only — no DB access directly |
| Workers | `workers/` | Background goroutine queue for async article processing |
| Scrapers | `scraper/` | Chromedp-based web scrapers (garage.go, partsouq.go, toyodiy.go) |
| Commands | `commands/` | CLI-only workers (language sync, XLS import) |
| DTOs | `dto/` | Request/response shapes for service boundaries |
| Mappers | `mappers/` | Convert models → DTOs |
| Middleware | `middleware/` | JWT auth (`AuthRequired()`) |
| Utils | `utils/` | Pure helper functions (string, pointer, OEM normalization) |

---

## Key Domain Models

- **`ArticleItem`** — a parts catalog entry (maps to `article_items` table)
- **`Oem`** — OEM part number with brand (maps to `oems` table)
- **`ArticleOem`** — join between articles and OEMs
- **`ArticleVehicles`** — join between articles and TecDoc vehicle IDs
- **`ArticleCategory`** — join between articles and categories
- **`Xyr`** — Mongolia vehicle registry record (plate number → make/model/year)
- **`Engine`** / **`EngineFamily`** — TecDoc engine data
- **`Order`** / **`OrderItem`** / **`Invoice`** — order management

---

## External Integrations

| Service | Used In | Purpose |
|---|---|---|
| **XYP API** (`xyp-api.smartcar.mn`) | `controllers/vehicle.go` | Mongolia plate number → vehicle info |
| **RapidAPI** (`auto-parts-catalog.p.rapidapi.com`) | `services/rapidapi_service.go`, `models/article.go` | TecDoc parts catalog |
| **OpenAI GPT-4.1-mini** | `services/open_ai.go` | Parts identification + TecDoc chassis mapping |
| **Meilisearch** | `services/meilisearch_service.go`, `docker-compose.yml` | Full-text search (partially implemented) |
| **Sentry** | `main.go` | Error tracking |
| **Mailgun** | `services/otp_service.go` | OTP email delivery |
| **Chromedp** | `scraper/` | Headless browser scraping for vehicle data |

---

## Background Worker

`workers/StartWorker()` starts a goroutine that consumes `workers.ArticleQueue` (buffered channel, size 100).

To enqueue an article for async processing:
```go
workers.ArticleQueue <- articleItem
```

The worker calls `processArticle()` → `saveMain()` → `saveEngines()`. Queue overflow (>100 items) will block the sending goroutine.

---

## Search Flow

`GET /search?query=...&vehicle_id=...&page=1&limit=40`

1. Try `models.GetProducts(filter)` — queries local DB with priority ranking (OEM match > category > product name)
2. If no results → fall back to `models.GetArticleItemsByOemFromRapidAPI(query)` and async-save results to DB

---

## Known Issues / Technical Debt

These are documented so you don't re-introduce them or work around them incorrectly:

1. **Sentry DSN is hardcoded** in `main.go:27` — should be an env var
2. **`bcrypt` error ignored** in `controllers/auth.go:44` — `hashedPassword, _ := bcrypt.GenerateFromPassword(...)` must handle the error
3. **`SignIn` doesn't check `IsVerified`** — unverified users receive a valid JWT
4. **User created before OTP sends** — if OTP fails, a dangling unverified user remains in DB
5. **`localhost` in production CORS** — `http://localhost:5173` in `main.go:70` must be removed for production
6. **`JWT_SECRET` not validated at startup** — empty secret will sign tokens silently
7. **`GetOrderForPDF` missing early return on error** — `repositories/order_repository.go:47` returns `&order, err` even when `err != nil`
8. **`PersistFetchedArticles` goroutine errors are invisible** — only prints to stdout; consider Sentry capture
9. **MODE check inconsistency** — `main.go:52` uses `cfg.MODE` but `main.go:80` uses `os.Getenv("MODE")` directly
10. **`app/app.go` DI is unused** — `NewApp()` wires `ProductHandler` correctly but is never called in `main.go`; routes for it are never registered
11. **Complex SQL in model layer** — `models/product.go:GetProducts()` and `models/article.go` have raw SQL that belongs in repositories
12. **Duplicate `ProductFilter`** — defined in both `models/product.go` and `repositories/product_repository.go` with different fields
13. **Controllers bypass repository layer** — `controllers/auth.go` and `controllers/vehicle.go` call `database.DB` directly
14. **`order_service.go` creates repository inside functions** — should be injected via constructor like `ProductHandler` does

---

## Coding Conventions

- Use `dto/` types at service boundaries — don't return raw GORM models from services when a DTO exists
- Use `utils.SafeString()`, `utils.StringToUintPtr()`, etc. for pointer/nil handling — don't write inline nil checks
- OEM normalization: use `utils.NormalizeOEM()` and `utils.IsOEM()` — don't write ad-hoc regex
- Errors from GORM: check with `errors.Is(err, gorm.ErrRecordNotFound)` for not-found, not just `err != nil`
- New routes go in `main.go`; protected routes go under the `authorized` group using `middleware.AuthRequired()`
- New queries belong in `repositories/` not in `models/` or `controllers/`
