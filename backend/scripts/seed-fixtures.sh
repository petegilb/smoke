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

-- 30 days of snapshots per game, follower counts following a per-game baseline + trend + noise.
-- Some games are made to grow fast (high trend) so the Trending tab has visible movers.
INSERT INTO daily_snapshots (app_id, snapshot_date, follower_count, wishlist_rank)
SELECT
    9000000 + i,
    (CURRENT_DATE - d * INTERVAL '1 day')::date,
    GREATEST(0,
        -- baseline scales with i so top games have ~500k followers
        (1000 + (i * 47) % 500000)
        -- 30-day linear trend; every 7th game is a fast riser
        + ((30 - d) * (CASE WHEN i % 7 = 0 THEN 800 ELSE 50 + (i % 200) END))
        -- pseudo-random noise per (game, day)
        + (((i * 31 + d * 17) % 600) - 300)
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
