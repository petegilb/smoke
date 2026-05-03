package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
)

// statusRecorder wraps http.ResponseWriter to capture the response status code for logging.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// requestLogger logs every HTTP request with method, path, status, and duration.
func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		log.Printf("%d %s %s (%s)", rec.status, r.Method, r.URL.RequestURI(), time.Since(start).Round(time.Microsecond))
	})
}

// serverError logs the underlying error context and returns a generic 500 to the client.
// The error detail stays in the log; the response intentionally hides driver/schema details
// so internal mistakes don't leak to callers.
func serverError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("error %s %s: %v", r.Method, r.URL.Path, err)
	http.Error(w, "internal server error", http.StatusInternalServerError)
}

// handleListGames returns all tracked games ordered by name.
// GET /api/games
func handleListGames(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		games, err := ListGames(db)
		if err != nil {
			serverError(w, r, err)
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
			serverError(w, r, err)
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
// GET /api/games/list?sort=followers|trending&type=game&indie=true&include_released=true&limit=500
func handleGameList(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		sort := q.Get("sort")
		if sort != "trending" && sort != "pct" && sort != "gain_24h" {
			sort = "followers"
		}
		gameType := q.Get("type")
		indie := q.Get("indie") == "true"
		includeReleased := q.Get("include_released") == "true"

		limit := 500
		if v := q.Get("limit"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 500 {
				limit = n
			}
		}

		items, err := ListGameItems(db, sort, gameType, indie, includeReleased, limit)
		if err != nil {
			serverError(w, r, err)
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
			serverError(w, r, err)
			return
		}

		now := time.Now().UTC()
		next := time.Date(now.Year(), now.Month(), now.Day(), 6, 0, 0, 0, time.UTC)
		if !next.After(now) {
			next = next.Add(24 * time.Hour)
		}

		resp := map[string]any{
			"last_scraped_at":     last,
			"next_scrape_at":      next,
			"scrape_in_progress":  IsScrapeInProgress(),
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
			serverError(w, r, err)
			return
		}
		if snapshots == nil {
			snapshots = []DailySnapshot{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(snapshots)
	}
}
