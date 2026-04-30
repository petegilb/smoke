package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/lib/pq"
)

// Game represents a Steam game's static metadata, stored in the games table.
type Game struct {
	AppID            int       `db:"app_id" json:"app_id"`
	Name             string    `db:"name" json:"name"`
	Type             string    `db:"type" json:"type"`
	IsFree           bool      `db:"is_free" json:"is_free"`
	ComingSoon       bool      `db:"coming_soon" json:"coming_soon"`
	ReleaseDate      string    `db:"release_date" json:"release_date"`
	Developers       []string  `db:"developers" json:"developers"`
	Publishers       []string  `db:"publishers" json:"publishers"`
	HeaderImage      string    `db:"header_image" json:"header_image"`
	ShortDescription string    `db:"short_description" json:"short_description"`
	Genres           []string  `db:"genres" json:"genres"`
	Categories       []string  `db:"categories" json:"categories"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}

// DailySnapshot represents a point-in-time record of a game's follower count,
// wishlist ranking, and player count. One snapshot per game per day.
type DailySnapshot struct {
	ID             int       `db:"id" json:"id"`
	AppID          int       `db:"app_id" json:"app_id"`
	SnapshotDate   time.Time `db:"snapshot_date" json:"snapshot_date"`
	FollowerCount  *int      `db:"follower_count" json:"follower_count"`
	WishlistRank   *int      `db:"wishlist_rank" json:"wishlist_rank"`
	CurrentPlayers *int      `db:"current_players" json:"current_players"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}

// RunMigrations applies all pending database migrations from the migrations/ directory.
func RunMigrations(dbURL string) error {
	m, err := migrate.New("file://migrations", dbURL)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	log.Println("migrations complete")
	return nil
}

// OpenDB opens a PostgreSQL connection pool and verifies connectivity with a ping.
func OpenDB(dbURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}
	return db, nil
}

// UpsertGame inserts a game or updates it if a game with the same app_id already exists.
func UpsertGame(db *sql.DB, g Game) error {
	_, err := db.Exec(`
		INSERT INTO games (app_id, name, type, is_free, coming_soon, release_date, developers, publishers, header_image, short_description, genres, categories, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())
		ON CONFLICT (app_id) DO UPDATE SET
			name = EXCLUDED.name,
			type = EXCLUDED.type,
			is_free = EXCLUDED.is_free,
			coming_soon = EXCLUDED.coming_soon,
			release_date = EXCLUDED.release_date,
			developers = EXCLUDED.developers,
			publishers = EXCLUDED.publishers,
			header_image = EXCLUDED.header_image,
			short_description = EXCLUDED.short_description,
			genres = EXCLUDED.genres,
			categories = EXCLUDED.categories,
			updated_at = NOW()`,
		g.AppID, g.Name, g.Type, g.IsFree, g.ComingSoon, g.ReleaseDate,
		pq.Array(g.Developers), pq.Array(g.Publishers), g.HeaderImage, g.ShortDescription,
		pq.Array(g.Genres), pq.Array(g.Categories),
	)
	return err
}

// UpsertSnapshot inserts a daily snapshot or merges it with an existing one for the same
// (app_id, snapshot_date). Uses COALESCE so that non-nil fields from the new snapshot
// overwrite existing values while preserving previously collected fields.
func UpsertSnapshot(db *sql.DB, s DailySnapshot) error {
	_, err := db.Exec(`
		INSERT INTO daily_snapshots (app_id, snapshot_date, follower_count, wishlist_rank, current_players)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (app_id, snapshot_date) DO UPDATE SET
			follower_count = COALESCE(EXCLUDED.follower_count, daily_snapshots.follower_count),
			wishlist_rank = COALESCE(EXCLUDED.wishlist_rank, daily_snapshots.wishlist_rank),
			current_players = COALESCE(EXCLUDED.current_players, daily_snapshots.current_players)`,
		s.AppID, s.SnapshotDate, s.FollowerCount, s.WishlistRank, s.CurrentPlayers,
	)
	return err
}

// ListGames returns all tracked games ordered alphabetically by name.
func ListGames(db *sql.DB) ([]Game, error) {
	rows, err := db.Query(`
		SELECT app_id, name, type, is_free, coming_soon, release_date,
			developers, publishers, header_image, short_description, genres, categories, created_at, updated_at
		FROM games ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []Game
	for rows.Next() {
		var g Game
		err := rows.Scan(
			&g.AppID, &g.Name, &g.Type, &g.IsFree, &g.ComingSoon, &g.ReleaseDate,
			pq.Array(&g.Developers), pq.Array(&g.Publishers), &g.HeaderImage, &g.ShortDescription,
			pq.Array(&g.Genres), pq.Array(&g.Categories),
			&g.CreatedAt, &g.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		games = append(games, g)
	}
	return games, rows.Err()
}

// GetGame returns a single game by app ID, or nil if not found.
func GetGame(db *sql.DB, appID int) (*Game, error) {
	var g Game
	err := db.QueryRow(`
		SELECT app_id, name, type, is_free, coming_soon, release_date,
			developers, publishers, header_image, short_description, genres, categories, created_at, updated_at
		FROM games WHERE app_id = $1`, appID).Scan(
		&g.AppID, &g.Name, &g.Type, &g.IsFree, &g.ComingSoon, &g.ReleaseDate,
		pq.Array(&g.Developers), pq.Array(&g.Publishers), &g.HeaderImage, &g.ShortDescription,
		pq.Array(&g.Genres), pq.Array(&g.Categories),
		&g.CreatedAt, &g.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &g, nil
}

// GetSnapshots returns all daily snapshots for a game ordered chronologically.
func GetSnapshots(db *sql.DB, appID int) ([]DailySnapshot, error) {
	rows, err := db.Query(`
		SELECT id, app_id, snapshot_date, follower_count, wishlist_rank, current_players, created_at
		FROM daily_snapshots WHERE app_id = $1 ORDER BY snapshot_date`, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []DailySnapshot
	for rows.Next() {
		var s DailySnapshot
		var snapshotDate time.Time
		err := rows.Scan(&s.ID, &s.AppID, &snapshotDate, &s.FollowerCount, &s.WishlistRank, &s.CurrentPlayers, &s.CreatedAt)
		if err != nil {
			return nil, err
		}
		s.SnapshotDate = snapshotDate
		snapshots = append(snapshots, s)
	}
	return snapshots, rows.Err()
}

// GameListItem is a denormalized row for the dashboard list endpoint.
type GameListItem struct {
	AppID            int      `json:"app_id"`
	Name             string   `json:"name"`
	HeaderImage      string   `json:"header_image"`
	Type             string   `json:"type"`
	Genres           []string `json:"genres"`
	Developers       []string `json:"developers"`
	Publishers       []string `json:"publishers"`
	ReleaseDate      string   `json:"release_date"`
	ShortDescription string   `json:"short_description"`
	CurrentFollowers int      `json:"current_followers"`
	WishlistRank     *int     `json:"wishlist_rank"`
	Delta24h         *int     `json:"delta_24h"`
	Delta7d          *int     `json:"delta_7d"`
	Sparkline        []int64  `json:"sparkline"`
}

// ListGameItems returns the dashboard list with current followers, 24h/7d deltas,
// and a 30-day follower-count sparkline per game. Filters by type and indie genre,
// sorted by either current followers or 7-day delta.
func ListGameItems(db *sql.DB, sort, gameType string, indieOnly bool, limit int) ([]GameListItem, error) {
	// Default = Top Wishlisted, sorted by Steam's own wishlist rank so the
	// displayed #N matches the row order (lower rank = more wishlisted).
	orderBy := "l.wishlist_rank ASC NULLS LAST"
	switch sort {
	case "gain_24h":
		orderBy = "delta_24h DESC NULLS LAST"
	case "trending":
		orderBy = "delta_7d DESC NULLS LAST"
	case "pct":
		// Percentage growth over the last 7 days. NULLIF avoids divide-by-zero;
		// references raw CTE columns because Postgres only resolves SELECT aliases
		// as standalone names in ORDER BY, not inside expressions.
		orderBy = "((l.follower_count - p7.follower_count)::float / NULLIF(p7.follower_count, 0)) DESC NULLS LAST"
	}

	query := `
		WITH latest AS (
			SELECT DISTINCT ON (app_id) app_id, follower_count, wishlist_rank, snapshot_date
			FROM daily_snapshots
			WHERE follower_count IS NOT NULL
			ORDER BY app_id, snapshot_date DESC
		),
		-- Anchor delta windows on each game's own latest snapshot rather than
		-- CURRENT_DATE. Otherwise, when today's scrape hasn't run yet, "latest"
		-- falls inside the calendar window and the CTE matches the same row,
		-- producing a misleading delta of 0.
		prev_24h AS (
			SELECT DISTINCT ON (s.app_id) s.app_id, s.follower_count
			FROM daily_snapshots s
			JOIN latest l ON s.app_id = l.app_id
			WHERE s.follower_count IS NOT NULL
			  AND s.snapshot_date BETWEEN (l.snapshot_date - INTERVAL '2 days') AND (l.snapshot_date - INTERVAL '1 day')
			ORDER BY s.app_id, s.snapshot_date DESC
		),
		prev_7d AS (
			SELECT DISTINCT ON (s.app_id) s.app_id, s.follower_count
			FROM daily_snapshots s
			JOIN latest l ON s.app_id = l.app_id
			WHERE s.follower_count IS NOT NULL
			  AND s.snapshot_date BETWEEN (l.snapshot_date - INTERVAL '8 days') AND (l.snapshot_date - INTERVAL '6 days')
			ORDER BY s.app_id, s.snapshot_date DESC
		),
		spark AS (
			SELECT app_id, array_agg(follower_count ORDER BY snapshot_date) AS pts
			FROM daily_snapshots
			WHERE follower_count IS NOT NULL
			  AND snapshot_date >= (CURRENT_DATE - INTERVAL '30 days')
			GROUP BY app_id
		)
		SELECT g.app_id, g.name, g.header_image, g.type, g.genres,
			g.developers, g.publishers, g.release_date, g.short_description,
			COALESCE(l.follower_count, 0) AS current_followers,
			l.wishlist_rank,
			(l.follower_count - p1.follower_count) AS delta_24h,
			(l.follower_count - p7.follower_count) AS delta_7d,
			COALESCE(s.pts, '{}') AS sparkline
		FROM games g
		JOIN latest l ON g.app_id = l.app_id
		LEFT JOIN prev_24h p1 ON g.app_id = p1.app_id
		LEFT JOIN prev_7d p7 ON g.app_id = p7.app_id
		LEFT JOIN spark s ON g.app_id = s.app_id
		WHERE ($1 = '' OR g.type = $1)
		  AND (NOT $2 OR 'Indie' = ANY(g.genres))
		ORDER BY ` + orderBy + `
		LIMIT $3`

	rows, err := db.Query(query, gameType, indieOnly, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []GameListItem{}
	for rows.Next() {
		var item GameListItem
		err := rows.Scan(
			&item.AppID, &item.Name, &item.HeaderImage, &item.Type,
			pq.Array(&item.Genres),
			pq.Array(&item.Developers), pq.Array(&item.Publishers),
			&item.ReleaseDate, &item.ShortDescription,
			&item.CurrentFollowers, &item.WishlistRank, &item.Delta24h, &item.Delta7d,
			pq.Array(&item.Sparkline),
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// ListAllAppIDs returns every app_id currently tracked in the games table.
// Used by the scraper to refresh games that have fallen out of the top wishlist pages.
func ListAllAppIDs(db *sql.DB) ([]int, error) {
	rows, err := db.Query(`SELECT app_id FROM games`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// GetLastScrapedAt returns the time of the last completed scrape.
func GetLastScrapedAt(db *sql.DB) (time.Time, error) {
	var val string
	err := db.QueryRow(`SELECT value FROM scrape_metadata WHERE key = 'last_scraped_at'`).Scan(&val)
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339, val)
}

// SetLastScrapedAt records the time of the most recent completed scrape.
func SetLastScrapedAt(db *sql.DB, t time.Time) error {
	_, err := db.Exec(`
		UPDATE scrape_metadata SET value = $1 WHERE key = 'last_scraped_at'`,
		t.UTC().Format(time.RFC3339),
	)
	return err
}
