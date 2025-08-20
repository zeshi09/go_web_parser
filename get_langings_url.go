package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

var urlForGetLandings string = "https://tools.kontur.ru/module/LandingPage/list"

type Payload struct {
	SelectItems []struct {
		Value string `json:"value"`
		Text  string `json:"text"`
	} `json:"selectItems"`
	Type string `json:"type"`
}

func GetLandingsUrls() {
	resp, err := http.Get(urlForGetLandings)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	raw := doc.Find(`script#filter-template-LandingDomainId-app-model`).First().Text()
	var p Payload
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		log.Fatal(err)
	}

	for _, it := range p.SelectItems {
		fmt.Println(it.Text)
	}
}
