package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ------ Start appid types ------
type AppDetailsResponse map[string]AppDetailsEntry

type AppDetailsEntry struct {
	Success bool       `json:"success"`
	Data    AppDetails `json:"data"`
}

type AppDetails struct {
	Type               string             `json:"type"`
	Name               string             `json:"name"`
	Steamappid         int                `json:"steam_appid"`
	RequiredAge        int                `json:"required_age"`
	IsFree             bool               `json:"is_free"`
	ControllerSupport  string             `json:"controller_support"`
	ShortDescription   string             `json:"short_description"`
	SupportedLanguages string             `json:"supported_languages"`
	HeaderImage        string             `json:"header_image"`
	CapsuleImage       string             `json:"capsule_image"`
	Website            string             `json:"website"`
	Developers         []string           `json:"developers"`
	Publishers         []string           `json:"publishers"`
	Platforms          Platforms          `json:"platforms"`
	Categories         []Category         `json:"categories"`
	Genres             []Genre            `json:"genres"`
	Screenshots        []Screenshot       `json:"screenshots"`
	Movies             []Movie            `json:"movies"`
	ReleaseDate        ReleaseDate        `json:"release_date"`
	SupportInfo        SupportInfo        `json:"support_info"`
	Background         string             `json:"background"`
	BackgroundRaw      string             `json:"background_raw"`
	ContentDescriptors ContentDescriptors `json:"content_descriptors"`
	PriceOverview      *PriceOverview     `json:"price_overview"`
	PackageGroups      []json.RawMessage  `json:"package_groups"`
	PCRequirements     json.RawMessage    `json:"pc_requirements"`
	MacRequirements    json.RawMessage    `json:"mac_requirements"`
	LinuxRequirements  json.RawMessage    `json:"linux_requirements"`
}

type Platforms struct {
	Windows bool `json:"windows"`
	Mac     bool `json:"mac"`
	Linux   bool `json:"linux"`
}

type Category struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
}

type Genre struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

type Screenshot struct {
	ID            int    `json:"id"`
	PathThumbnail string `json:"path_thumbnail"`
	PathFull      string `json:"path_full"`
}

type Movie struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Thumbnail string `json:"thumbnail"`
	DashAV1   string `json:"dash_av1"`
	DashH264  string `json:"dash_h264"`
	HLSH264   string `json:"hls_h264"`
	Highlight bool   `json:"highlight"`
}

type ReleaseDate struct {
	ComingSoon bool   `json:"coming_soon"`
	Date       string `json:"date"`
}

type SupportInfo struct {
	URL   string `json:"url"`
	Email string `json:"email"`
}

type PriceOverview struct {
	Currency       string `json:"currency"`
	Initial        int    `json:"initial"`
	Final          int    `json:"final"`
	FinalFormatted string `json:"final_formatted"`
}

type ContentDescriptors struct {
	IDs   []int   `json:"ids"`
	Notes *string `json:"notes"`
}

// ------ End appid types ------

var appDetailsUrl = "https://store.steampowered.com/api/appdetails?appids=%d"

// getAppDetails fetches a game's metadata from the Steam Store API (store.steampowered.com/api/appdetails).
func getAppDetails(appid int) (*AppDetails, error) {
	requestURL := fmt.Sprintf(appDetailsUrl, appid)
	res, err := http.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("fetching app details for %d: %w", appid, err)
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body for %d: %w", appid, err)
	}

	var payload AppDetailsResponse
	err = json.Unmarshal(resBody, &payload)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling app details for %d: %w", appid, err)
	}

	entry, ok := payload[fmt.Sprint(appid)]
	if !ok || !entry.Success {
		return nil, fmt.Errorf("app details not found or failed for %d", appid)
	}

	return &entry.Data, nil
}
