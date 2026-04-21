package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
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
