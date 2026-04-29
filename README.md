# smoke

A Steam analytics dashboard that tracks daily follower counts as a wishlist proxy and surfaces trending and fast-growing games.

Live at **[steamwishlists.com](https://steamwishlists.com)**.

## What it does

- Scrapes Steam Store metadata (genres, developers, release date, etc.) and Steam Community follower counts daily
- Stores snapshots in Postgres for trend analysis
- Serves a virtualized React dashboard with three views: Top Wishlisted, Trending (24h/7d deltas), and % Movers

## Stack

- **Backend**: Go (stdlib `net/http`, `robfig/cron`, `golang-migrate`, `lib/pq`)
- **Frontend**: React 19 + Vite + TypeScript, Tailwind v4 + daisyUI v5, `@tanstack/react-virtual`
- **Database**: PostgreSQL 17
- **Reverse proxy**: nginx (serves the SPA, proxies `/api` to the backend)

## Local development

Two modes: hot-reload (run each piece on the host) or prod-like (everything in Docker).

### Hot-reload

```bash
# 1. Start Postgres only
docker compose up -d postgres

# 2. Configure env
cp example.env .env
# fill in POSTGRES_PASSWORD and any Steam creds you have

# 3. Backend
cd backend && go run .

# 4. Frontend (separate terminal)
cd frontend && pnpm install && pnpm dev
```

Vite proxies `/api` to `127.0.0.1:8080`. Visit http://localhost:5173.

### Prod-like (full Docker stack)

```bash
cp example.env .env
# set POSTGRES_PASSWORD to a real value
docker compose up -d --build
```

Visit http://localhost. nginx serves the built SPA on :80 and proxies `/api` to the backend container.

## Useful commands

```bash
# Backend tests / build
cd backend && go test ./... && go build .

# Frontend lint / build
cd frontend && pnpm lint && pnpm build

# Reset the database
./backend/scripts/reset-db.sh

# Seed 1000 fake games with 30 days of snapshots (great for UI testing)
./backend/scripts/seed-fixtures.sh

# Connect to the DB
psql -h localhost -U smoke -d smoke

# Backup / restore
docker exec smoke-db pg_dump -U smoke -d smoke -F c > backups/smoke_$(date +%Y%m%d_%H%M%S).dump
docker exec -i smoke-db pg_restore -U smoke -d smoke --clean < backups/<filename>.dump
```

Migrations run automatically on backend startup.

## Project layout

```
backend/        Go HTTP API + scraper
  migrations/   golang-migrate SQL files (auto-applied on startup)
  scripts/      reset-db.sh, seed-fixtures.sh
  Dockerfile
frontend/       React/Vite SPA
  src/
  nginx.conf    serves SPA, proxies /api → backend
  Dockerfile
docker-compose.yml
```

## API endpoints

- `GET /api/games/list?sort={followers|trending|pct}&type=&indie=true&limit=200` — dashboard list with deltas + 30-day sparkline
- `GET /api/meta` — `{ last_scraped_at, next_scrape_at }`
- `GET /api/games` — all tracked games
- `GET /api/games/{appID}` — single game
- `GET /api/games/{appID}/snapshots` — time series for one game
