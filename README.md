# smoke

go backend in /backend

## postgresql db

access db: psql -h localhost -U smoke -d smoke

### migrations

# 1. generate empty migration files
migrate create -ext sql -dir migrations/ -seq create_games

# this creates:
# migrations/000001_create_games.up.sql   (empty)
# migrations/000001_create_games.down.sql (empty)

# 2. write your SQL in those files
# (edit 000001_create_games.up.sql with your CREATE TABLE)
# (edit 000001_create_games.down.sql with your DROP TABLE)

# 3. run the migration
migrate -path migrations/ -database "postgres://smoke:smoke@localhost:5432/smoke?sslmode=disable" up