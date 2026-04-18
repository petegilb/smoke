CREATE TABLE games (
    app_id      INTEGER PRIMARY KEY,
    name        TEXT NOT NULL,
    type        TEXT,
    is_free     BOOLEAN DEFAULT FALSE,
    coming_soon BOOLEAN DEFAULT FALSE,
    release_date TEXT,
    developers  TEXT[],
    publishers  TEXT[],
    header_image TEXT,
    short_description TEXT,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE daily_snapshots (
    id              SERIAL PRIMARY KEY,
    app_id          INTEGER NOT NULL REFERENCES games(app_id),
    snapshot_date   DATE NOT NULL DEFAULT CURRENT_DATE,
    follower_count  INTEGER,
    wishlist_rank   INTEGER,
    current_players INTEGER,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(app_id, snapshot_date)
);

CREATE INDEX idx_snapshots_app_date ON daily_snapshots(app_id, snapshot_date);