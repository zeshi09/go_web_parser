package input

import (
	"encoding/json"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

// Структура для получения урлов из веб-админки лендингов
type Payload struct {
	SelectItems []struct {
		Value string `json:"value"`
		Text  string `json:"text"`
	} `json:"selectItems"`
	Type string `json:"type"`
}

// URL админки лендингов
var urlForGetLandings string = "https://tools.kontur.ru/module/LandingPage/list"

// Получаем список лендинговых урлов
func GetLandingsUrls(cookie string) ([]string, error) {

	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}
	client := &http.Client{
		Transport: tr,
	}

	req, err := http.NewRequest("GET", urlForGetLandings, nil)

	req.Header.Set("Cookie", cookie)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	raw := doc.Find(`script#filter-template-LandingDomainId-app-model`).First().Text()
	var p Payload
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		return nil, err
	}

	var urls []string

	for _, it := range p.SelectItems {
		urls = append(urls, it.Text)
	}

	return urls, nil
}
