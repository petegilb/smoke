package main

import (
    "fmt"
    "log"
    "net/http"
    "os"
    "io"
    "encoding/json"
    "regexp"
    "strconv"
)

var appIDRegex = regexp.MustCompile(`/apps/(\d+)/`)

type storeResponse struct {
    Desc string `json:"desc"`
	Items []struct {
		Name string `json:"name"`
		Logo string `json:"logo"`
	} `json:"items"`
}


func main() {
    log.Println("Starting Steam store search...")

    requestURL := fmt.Sprintf("https://store.steampowered.com/search/results/?filter=popularwishlist&json=1&count=500&page=1")
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

    var payload storeResponse
    err = json.Unmarshal(resBody, &payload)
    if err != nil {
        log.Fatal("Error during Unmarshal(): ", err)
        os.Exit(1)
    }

    for i, game := range payload.Items {
        // Check if we can grab the appid from the url using regex
        m := appIDRegex.FindStringSubmatch(game.Logo)
        if len(m) >= 2 {
            appID, _ := strconv.Atoi(m[1])
            log.Printf("#%d: %s (appid: %d)\n", i+1, game.Name, appID)
        }
    }
}
