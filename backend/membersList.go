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

const longBackoffAttempts = 5
const longBackoffDuration = 5 * time.Minute

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
// First does 3 quick retries with exponential backoff, then up to 5 longer 5-minute waits
// for the rate limit to reset.
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
	}

	// Quick retries exhausted — wait 5 min for rate limit to reset, up to 5 times
	for attempt := range longBackoffAttempts {
		log.Printf("  Rate limited on app %d, waiting 5 minutes (attempt %d/%d)...", appid, attempt+1, longBackoffAttempts)
		time.Sleep(longBackoffDuration)

		details, err := fetchMembersList(requestURL)
		if err == nil {
			return details, nil
		}
	}

	return nil, fmt.Errorf("members list for %d: still failing after %d×5min waits", appid, longBackoffAttempts)
}
