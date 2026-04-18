package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"encoding/xml"
)

type MemberListXML struct {
	XMLName      xml.Name     `xml:"memberList"`
	GroupID64    string       `xml:"groupID64"`
	GroupDetails GroupDetails `xml:"groupDetails"`
}

type GroupDetails struct {
	GroupName   string `xml:"groupName"`
	GroupURL    string `xml:"groupURL"`
	Headline    string `xml:"headline"`
	Summary     string `xml:"summary"`
	AvatarIcon  string `xml:"avatarIcon"`
	AvatarMedium string `xml:"avatarMedium"`
	AvatarFull  string `xml:"avatarFull"`
	MemberCount int    `xml:"memberCount"`
}

var membersListUrl = "https://steamcommunity.com/games/%d/memberslistxml/?xml=1"

func getMembersList(appid int){
	requestURL := fmt.Sprintf(membersListUrl, appid)
    res, err := http.Get(requestURL)
    if err != nil {
        log.Printf("error making http request: %s\n", err)
        os.Exit(1)
    }

    log.Printf("client: got response!\n")
    log.Printf("client: status code: %d\n", res.StatusCode)
    // log.Printf("Body: %s\n", resBody)

	var data MemberListXML
	if err := xml.NewDecoder(res.Body).Decode(&data); err != nil {
		log.Fatal("failed to retrieve members list for %d", appid)
        os.Exit(1)
	}

	log.Printf("Game: %d, Followers: %d", data.GroupDetails.GroupName, data.GroupDetails.MemberCount)
	
}