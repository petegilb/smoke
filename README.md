# smoke

A Steam wishlist analytics tool that tracks daily follower counts to identify trending games. Built as a learning project for Go and React.

## What it does

- Fetches game metadata from the Steam Store API
- Tracks daily follower/member counts via the Steam Community feed
- Stores snapshots in PostgreSQL to analyze trends over time
- Plans to surface insights via a React frontend

## Stack

- **Backend**: Go
- **Database**: PostgreSQL (via Docker)
- **Frontend**: React (planned)

## Setup

**1. Start the database**
```bash
docker compose up -d
```

**2. Configure environment**
```bash
cp example.env .env
# edit .env with your Steam API credentials
```

**3. Run the backend**
```bash
cd backend
go run .
```

Migrations run automatically on startup.

## Database

Access the DB directly:
```bash
psql -h localhost -U smoke -d smoke
```

Add a migration:
```bash
migrate create -ext sql -dir migrations/ -seq <name>
migrate -path migrations/ -database "postgres://smoke:smoke@localhost:5432/smoke?sslmode=disable" up
```
