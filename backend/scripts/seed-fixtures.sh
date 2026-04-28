#!/bin/bash
# Seeds the smoke database with 1000 fake games and 30 days of snapshots each.
# Useful for exercising the dashboard UI without waiting on a real scrape.
# Safe to re-run; uses ON CONFLICT to upsert.

set -euo pipefail

DB_URL="${DATABASE_URL:-postgres://smoke:smoke@localhost:5432/smoke?sslmode=disable}"

echo "Seeding fixtures into $DB_URL..."

psql "$DB_URL" <<'SQL'
-- 1000 fake games, app_ids 9000000..9000999 to avoid colliding with real Steam app IDs.
INSERT INTO games (app_id, name, type, is_free, coming_soon, release_date,
                   developers, publishers, header_image, short_description,
                   genres, categories)
SELECT
    9000000 + i,
    'Test Game #' || i,
    (ARRAY['game','game','game','game','game','dlc','demo','music','software'])[1 + (i % 9)],
    (i % 7 = 0),
    false,
    (CURRENT_DATE - ((i * 13) % 1500) * INTERVAL '1 day')::text,
    ARRAY['Studio ' || ((i * 7) % 200)],
    ARRAY[(ARRAY['Megapub','IndieHouse','SmallCo','OneDev','Studio ' || ((i * 11) % 150)])[1 + (i % 5)]],
    'https://placehold.co/460x215/1f2937/9ca3af?text=Game+' || i,
    'A fake game for testing the dashboard. Number ' || i || '.',
    CASE WHEN i % 3 = 0
         THEN ARRAY['Indie', (ARRAY['Action','Adventure','RPG','Strategy','Puzzle','Simulation'])[1 + (i % 6)]]
         ELSE ARRAY[(ARRAY['Action','Adventure','RPG','Strategy','Puzzle','Simulation','Sports','Racing'])[1 + (i % 8)]]
    END,
    ARRAY['Single-player', 'Steam Achievements']
FROM generate_series(0, 999) AS i
ON CONFLICT (app_id) DO UPDATE SET
    name = EXCLUDED.name,
    type = EXCLUDED.type,
    genres = EXCLUDED.genres,
    categories = EXCLUDED.categories,
    publishers = EXCLUDED.publishers,
    updated_at = NOW();

-- 30 days of snapshots per game. Combines a per-game baseline with a mild linear
-- trend plus two sinusoidal waves and an occasional drop event so that sparklines
-- show real dips instead of straight lines.
INSERT INTO daily_snapshots (app_id, snapshot_date, follower_count, wishlist_rank)
SELECT
    9000000 + i,
    (CURRENT_DATE - d * INTERVAL '1 day')::date,
    GREATEST(100,
        -- baseline scales with i so top games have ~500k followers
        (1000 + (i * 47) % 500000)
        -- linear trend; every 7th game is a fast riser, every 11th is a faller
        + ((30 - d) * (CASE
            WHEN i % 7 = 0 THEN 600
            WHEN i % 11 = 0 THEN -300
            ELSE 30 + (i % 150)
          END))
        -- primary swing: amplitude is 8% of baseline, period 4–14 days per game
        + ((1000 + (i * 47) % 500000) * 0.08
            * sin(2 * pi() * d / (4 + (i % 11)) + (i % 17)))::int
        -- secondary higher-frequency wobble
        + ((1000 + (i * 47) % 500000) * 0.03
            * sin(2 * pi() * d / 3 + (i * 0.7)))::int
        -- occasional drop events (controversy, delay, etc.)
        + CASE WHEN ((i * 5 + d * 13) % 23) = 0
               THEN -((1000 + (i * 47) % 500000) * 0.06)::int
               ELSE 0
          END
        -- broader pseudo-random noise per (game, day)
        + (((i * 31 + d * 17) % 1400) - 700)
    )::int,
    (i + 1)
FROM generate_series(0, 999) AS i
CROSS JOIN generate_series(0, 29) AS d
ON CONFLICT (app_id, snapshot_date) DO UPDATE SET
    follower_count = EXCLUDED.follower_count,
    wishlist_rank = EXCLUDED.wishlist_rank;

-- Mark scrape as just-completed so the cron doesn't kick off on next server start.
INSERT INTO scrape_metadata (key, value) VALUES ('last_scraped_at', to_char(NOW(), 'YYYY-MM-DD"T"HH24:MI:SS"Z"'))
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value;
SQL

echo "Done. Inserted 1000 games × 30 snapshots."
