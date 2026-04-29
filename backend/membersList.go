package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"time"
)

type MemberListXML struct {
	XMLName      xml.Name     `xml:"memberList"`
	GroupID64    string       `xml:"groupID64"`
	GroupDetails GroupDetails `xml:"groupDetails"`
}

type GroupDetails struct {
	GroupName    string `xml:"groupName"`
	GroupURL     string `xml:"groupURL"`
	Headline     string `xml:"headline"`
	Summary      string `xml:"summary"`
	AvatarIcon   string `xml:"avatarIcon"`
	AvatarMedium string `xml:"avatarMedium"`
	AvatarFull   string `xml:"avatarFull"`
	MemberCount  int    `xml:"memberCount"`
}

var membersListUrl = "https://steamcommunity.com/games/%d/memberslistxml/?xml=1"

// longBackoffMinutes are the per-attempt wait times (in minutes) after the
// quick-retry budget is exhausted. Each entry is a fresh attempt that waits
// the listed duration first, giving Steam's rate limit progressively more
// time to reset before giving up.
var longBackoffMinutes = []int{5, 10, 15, 20, 30}

// fetchMembersList does a single HTTP GET + XML decode of the members list feed.
func fetchMembersList(requestURL string) (*GroupDetails, error) {
	res, err := http.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("fetching members list: %w", err)
	}
	defer res.Body.Close()

	var data MemberListXML
	if err := xml.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}
	return &data.GroupDetails, nil
}

// getMembersList fetches a game's follower/member count from the Steam Community XML feed.
// First does 3 quick retries with exponential backoff, then waits progressively longer
// (5, 10, 15, 20, 30 minutes) for the rate limit to reset.
func getMembersList(appid int) (*GroupDetails, error) {
	requestURL := fmt.Sprintf(membersListUrl, appid)

	for attempt := range 3 {
		if attempt > 0 {
			backoff := time.Duration(attempt*attempt) * 5 * time.Second
			log.Printf("  Retrying members list for %d in %s (attempt %d/3)", appid, backoff, attempt+1)
			time.Sleep(backoff)
		}

		details, err := fetchMembersList(requestURL)
		if err == nil {
			return details, nil
		}
		log.Printf("  fetch failed for %d (attempt %d/3): %v", appid, attempt+1, err)
	}

	// Quick retries exhausted — wait increasingly longer for the rate limit to reset.
	for i, minutes := range longBackoffMinutes {
		wait := time.Duration(minutes) * time.Minute
		log.Printf("  Rate limited on app %d, waiting %d minutes (attempt %d/%d)...", appid, minutes, i+1, len(longBackoffMinutes))
		time.Sleep(wait)

		details, err := fetchMembersList(requestURL)
		if err == nil {
			return details, nil
		}
		log.Printf("  fetch failed for %d (long-backoff attempt %d/%d): %v", appid, i+1, len(longBackoffMinutes), err)
	}

	return nil, fmt.Errorf("members list for %d: still failing after long backoff waits %v", appid, longBackoffMinutes)
}
