package main

import (
    "fmt"
    "log"
    "net/http"
    "os"
    "io"
    "encoding/json"
	"strings"
)

// ------ Start appid types ------
type AppDetailsResponse map[string]AppDetailsEntry

type AppDetailsEntry struct {
	Success bool       `json:"success"`
	Data    AppDetails `json:"data"`
}

type AppDetails struct {
	Type                string            `json:"type"`
	Name                string            `json:"name"`
	Steamappid          int               `json:"steam_appid"`
	RequiredAge         int               `json:"required_age"`
	IsFree              bool              `json:"is_free"`
	ControllerSupport   string            `json:"controller_support"`
	ShortDescription    string            `json:"short_description"`
	SupportedLanguages  string            `json:"supported_languages"`
	HeaderImage         string            `json:"header_image"`
	CapsuleImage        string            `json:"capsule_image"`
	Website             string            `json:"website"`
	Developers          []string          `json:"developers"`
	Publishers          []string          `json:"publishers"`
	Platforms           Platforms         `json:"platforms"`
	Categories          []Category        `json:"categories"`
	Genres              []Genre           `json:"genres"`
	Screenshots         []Screenshot      `json:"screenshots"`
	Movies              []Movie           `json:"movies"`
	ReleaseDate         ReleaseDate       `json:"release_date"`
	SupportInfo         SupportInfo       `json:"support_info"`
	Background          string            `json:"background"`
	BackgroundRaw       string            `json:"background_raw"`
	ContentDescriptors  ContentDescriptors `json:"content_descriptors"`
	PriceOverview       *PriceOverview    `json:"price_overview"`
	PackageGroups       []json.RawMessage `json:"package_groups"`
	PCRequirements      json.RawMessage   `json:"pc_requirements"`
	MacRequirements     json.RawMessage   `json:"mac_requirements"`
	LinuxRequirements   json.RawMessage   `json:"linux_requirements"`
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

func getAppDetails(appid int){
    requestURL := fmt.Sprintf(appDetailsUrl, appid)
    res, err := http.Get(requestURL)
    if err != nil {
        log.Printf("error making http request: %s\n", err)
        os.Exit(1)
    }

    resBody, err := io.ReadAll(res.Body)
    if err != nil {
        log.Printf("client: could not read response body: %s\n", err)
        os.Exit(1)
    }

    log.Printf("client: got response!\n")
    log.Printf("client: status code: %d\n", res.StatusCode)
    // log.Printf("Body: %s\n", resBody)

	var payload AppDetailsResponse
    err = json.Unmarshal(resBody, &payload)
    if err != nil  {
        log.Fatal("Error during Unmarshal(): ", err)
        os.Exit(1)
    }

	entry := payload[fmt.Sprint(appid)]
	if entry.Success {
		log.Printf("Game: %s, Type %s, Developers: %s\n", 
			entry.Data.Name, 
			entry.Data.Type,
			strings.Join(entry.Data.Developers, ", "),
		)
		if entry.Data.ReleaseDate.ComingSoon {
			log.Printf("%s - not yet released (date: %s)", entry.Data.Name, entry.Data.ReleaseDate.Date)
		} else {
			log.Printf("%s - released on %s", entry.Data.Name, entry.Data.ReleaseDate.Date)
		}
	}
	
}