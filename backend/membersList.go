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

// getMembersList fetches a game's follower/member count from the Steam Community XML feed.
// Retries up to 3 times with exponential backoff on failure (Steam Community rate limits aggressively).
func getMembersList(appid int) (*GroupDetails, error) {
	requestURL := fmt.Sprintf(membersListUrl, appid)

	for attempt := range 3 {
		if attempt > 0 {
			backoff := time.Duration(attempt*attempt) * 5 * time.Second
			log.Printf("  Retrying members list for %d in %s (attempt %d/3)", appid, backoff, attempt+1)
			time.Sleep(backoff)
		}

		res, err := http.Get(requestURL)
		if err != nil {
			return nil, fmt.Errorf("fetching members list for %d: %w", appid, err)
		}

		var data MemberListXML
		err = xml.NewDecoder(res.Body).Decode(&data)
		res.Body.Close()
		if err != nil {
			// EOF or decode errors are likely rate limiting — retry
			continue
		}

		return &data.GroupDetails, nil
	}

	return nil, fmt.Errorf("members list for %d: failed after 3 attempts (likely rate limited)", appid)
}
