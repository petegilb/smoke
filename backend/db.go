package main

import (
	"time"
	"log"
    "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

type Game struct {
    AppID            int       `db:"app_id"`
    Name             string    `db:"name"`
    Type             string    `db:"type"`
    IsFree           bool      `db:"is_free"`
    ComingSoon       bool      `db:"coming_soon"`
    ReleaseDate      string    `db:"release_date"`
    Developers       []string  `db:"developers"`
    Publishers       []string  `db:"publishers"`
    HeaderImage      string    `db:"header_image"`
    ShortDescription string    `db:"short_description"`
    CreatedAt        time.Time `db:"created_at"`
    UpdatedAt        time.Time `db:"updated_at"`
}

type DailySnapshot struct {
    ID             int       `db:"id"`
    AppID          int       `db:"app_id"`
    SnapshotDate   time.Time `db:"snapshot_date"`
    FollowerCount  *int      `db:"follower_count"`
    WishlistRank   *int      `db:"wishlist_rank"`
    CurrentPlayers *int      `db:"current_players"`
    CreatedAt      time.Time `db:"created_at"`
}

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