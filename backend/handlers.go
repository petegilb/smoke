package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

// handleListGames returns all tracked games ordered by name.
// GET /api/games
func handleListGames(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		games, err := ListGames(db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if games == nil {
			games = []Game{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(games)
	}
}

// handleGetGame returns a single game by its Steam app ID.
// GET /api/games/{appID}
func handleGetGame(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appIDStr := r.PathValue("appID")
		appID, err := strconv.Atoi(appIDStr)
		if err != nil {
			http.Error(w, "invalid app ID", http.StatusBadRequest)
			return
		}

		game, err := GetGame(db, appID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if game == nil {
			http.Error(w, "game not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(game)
	}
}

// handleGameList returns the dashboard list with deltas and sparklines.
// GET /api/games/list?sort=followers|trending&type=game&indie=true&limit=200
func handleGameList(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		sort := q.Get("sort")
		if sort != "trending" {
			sort = "followers"
		}
		gameType := q.Get("type")
		indie := q.Get("indie") == "true"

		limit := 200
		if v := q.Get("limit"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 500 {
				limit = n
			}
		}

		items, err := ListGameItems(db, sort, gameType, indie, limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(items)
	}
}

// handleMeta returns scrape metadata: when the last scrape ran and when the next is scheduled.
// GET /api/meta
func handleMeta(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		last, err := GetLastScrapedAt(db)
		if err != nil && err != sql.ErrNoRows {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		now := time.Now().UTC()
		next := time.Date(now.Year(), now.Month(), now.Day(), 6, 0, 0, 0, time.UTC)
		if !next.After(now) {
			next = next.Add(24 * time.Hour)
		}

		resp := map[string]any{
			"last_scraped_at": last,
			"next_scrape_at":  next,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// handleGetSnapshots returns all daily snapshots for a game, ordered by date.
// GET /api/games/{appID}/snapshots
func handleGetSnapshots(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appIDStr := r.PathValue("appID")
		appID, err := strconv.Atoi(appIDStr)
		if err != nil {
			http.Error(w, "invalid app ID", http.StatusBadRequest)
			return
		}

		snapshots, err := GetSnapshots(db, appID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if snapshots == nil {
			snapshots = []DailySnapshot{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(snapshots)
	}
}
