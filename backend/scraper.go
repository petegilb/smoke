package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

var appIDRegex = regexp.MustCompile(`/apps/(\d+)/`)
var popularWishlistUrl = "https://store.steampowered.com/search/results/?filter=popularwishlist&json=1&count=500&page=%d"

const scrapePages = 10
const pageDelay = 2 * time.Second
const requestDelay = 2 * time.Second

// storeResponse is the JSON structure returned by the Steam store search/results endpoint.
type storeResponse struct {
	Desc  string `json:"desc"`
	Items []struct {
		Name string `json:"name"`
		Logo string `json:"logo"`
	} `json:"items"`
}

// getPopularWishlists fetches a single page of popular wishlists and returns app IDs.
func getPopularWishlists(page int) ([]int, error) {
	requestURL := fmt.Sprintf(popularWishlistUrl, page)
	res, err := http.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("fetching popular wishlists page %d: %w", page, err)
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading popular wishlists response: %w", err)
	}

	var payload storeResponse
	if err := json.Unmarshal(resBody, &payload); err != nil {
		return nil, fmt.Errorf("unmarshalling popular wishlists: %w", err)
	}

	var appIDs []int
	for _, game := range payload.Items {
		m := appIDRegex.FindStringSubmatch(game.Logo)
		if len(m) >= 2 {
			appID, _ := strconv.Atoi(m[1])
			appIDs = append(appIDs, appID)
		}
	}

	return appIDs, nil
}

// collectAppIDs fetches the first 10 pages of popular wishlists with a 2s delay between pages.
func collectAppIDs() ([]int, error) {
	var allIDs []int
	for page := 1; page <= scrapePages; page++ {
		if page > 1 {
			time.Sleep(pageDelay)
		}
		log.Printf("Fetching wishlist page %d/%d", page, scrapePages)
		ids, err := getPopularWishlists(page)
		if err != nil {
			log.Printf("  Error on page %d: %v (skipping)", page, err)
			continue
		}
		log.Printf("  Got %d app IDs from page %d", len(ids), page)
		allIDs = append(allIDs, ids...)
	}
	return allIDs, nil
}

// RunScrape fetches the popular wishlists (10 pages), then scrapes details and follower
// counts for each game with a 1s delay between requests to avoid Steam rate limits.
func RunScrape(db *sql.DB) error {
	log.Println("Starting scrape...")

	appIDs, err := collectAppIDs()
	if err != nil {
		return fmt.Errorf("collecting app IDs: %w", err)
	}
	log.Printf("Found %d total games to scrape", len(appIDs))

	today := time.Now()
	var errCount int

	for i, appID := range appIDs {
		log.Printf("[%d/%d] Scraping app %d", i+1, len(appIDs), appID)

		// Fetch and upsert game details
		time.Sleep(requestDelay)
		details, err := getAppDetails(appID)
		if err != nil {
			log.Printf("  Error fetching app details for %d: %v", appID, err)
			errCount++
			continue
		}

		game := Game{
			AppID:            details.Steamappid,
			Name:             details.Name,
			Type:             details.Type,
			IsFree:           details.IsFree,
			ComingSoon:       details.ReleaseDate.ComingSoon,
			ReleaseDate:      details.ReleaseDate.Date,
			Developers:       details.Developers,
			Publishers:       details.Publishers,
			HeaderImage:      details.HeaderImage,
			ShortDescription: details.ShortDescription,
		}
		if err := UpsertGame(db, game); err != nil {
			log.Printf("  Error upserting game %d: %v", appID, err)
			errCount++
			continue
		}

		// Fetch follower count
		time.Sleep(requestDelay)
		members, err := getMembersList(appID)
		followerCount := 0
		if err != nil {
			log.Printf("  Warning: could not get members list for %d: %v", appID, err)
		} else {
			followerCount = members.MemberCount
		}

		// Upsert daily snapshot
		wishlistRank := i + 1
		snapshot := DailySnapshot{
			AppID:         appID,
			SnapshotDate:  today,
			FollowerCount: &followerCount,
			WishlistRank:  &wishlistRank,
		}
		if err := UpsertSnapshot(db, snapshot); err != nil {
			log.Printf("  Error upserting snapshot for %d: %v", appID, err)
			errCount++
			continue
		}

		log.Printf("  Done: %s (followers: %d, rank: %d)", details.Name, followerCount, wishlistRank)
	}

	// Record scrape completion time
	if err := SetLastScrapedAt(db, time.Now()); err != nil {
		log.Printf("Warning: failed to update last_scraped_at: %v", err)
	}

	log.Printf("Scrape complete. %d games processed, %d errors", len(appIDs), errCount)
	return nil
}
