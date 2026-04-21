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
		INSERT INTO games (app_id, name, type, is_free, coming_soon, release_date, developers, publishers, header_image, short_description, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
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
			updated_at = NOW()`,
		g.AppID, g.Name, g.Type, g.IsFree, g.ComingSoon, g.ReleaseDate,
		pq.Array(g.Developers), pq.Array(g.Publishers), g.HeaderImage, g.ShortDescription,
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
			developers, publishers, header_image, short_description, created_at, updated_at
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
			developers, publishers, header_image, short_description, created_at, updated_at
		FROM games WHERE app_id = $1`, appID).Scan(
		&g.AppID, &g.Name, &g.Type, &g.IsFree, &g.ComingSoon, &g.ReleaseDate,
		pq.Array(&g.Developers), pq.Array(&g.Publishers), &g.HeaderImage, &g.ShortDescription,
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
