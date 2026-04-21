package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/rs/cors"
)

var defaultDB = "postgres://smoke:smoke@localhost:5432/smoke?sslmode=disable"

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = defaultDB
	}

	// Run migrations
	log.Println("Running migrations...")
	if err := RunMigrations(dbURL); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Migrations complete")

	// Open DB connection pool
	db, err := OpenDB(dbURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Check if we need to scrape immediately (>24 hours since last scrape)
	lastScraped, err := GetLastScrapedAt(db)
	if err != nil {
		log.Printf("Warning: could not get last scraped time: %v", err)
		lastScraped = time.Time{} // zero time triggers immediate scrape
	}
	if time.Since(lastScraped) > 24*time.Hour {
		log.Printf("Last scrape was %s ago, starting immediate scrape...", time.Since(lastScraped).Round(time.Minute))
		go func() {
			if err := RunScrape(db); err != nil {
				log.Printf("Immediate scrape error: %v", err)
			}
		}()
	} else {
		log.Printf("Last scrape was %s ago, skipping immediate scrape", time.Since(lastScraped).Round(time.Minute))
	}

	// Start cron scheduler — scrape daily at 6am UTC
	c := cron.New()
	c.AddFunc("0 6 * * *", func() {
		log.Println("Cron: starting daily scrape")
		if err := RunScrape(db); err != nil {
			log.Printf("Cron: scrape error: %v", err)
		}
	})
	c.Start()
	defer c.Stop()
	log.Println("Cron scheduler started (daily scrape at 6am UTC)")

	// Set up routes
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/games", handleListGames(db))
	mux.HandleFunc("GET /api/games/{appID}", handleGetGame(db))
	mux.HandleFunc("GET /api/games/{appID}/snapshots", handleGetSnapshots(db))

	// CORS middleware
	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:*", "http://127.0.0.1:*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: false,
	}).Handler(mux)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
