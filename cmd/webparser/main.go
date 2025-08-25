package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/zeshi09/go_web_parser/internal/config"
	"github.com/zeshi09/go_web_parser/internal/crawler"
	"github.com/zeshi09/go_web_parser/internal/filter"
	"github.com/zeshi09/go_web_parser/internal/input"
)

type Findings struct {
	Url      string   `json:"url"`
	Findings []string `json:"findings"`
}

func main() {

	cookie := flag.String("cookie", "", "Your cookie for tools.kontur.ru")
	flag.Parse()

	config := config.Config{
		ProxyURL:       "http://192.168.2.200:8080",
		RequestTimeout: 20 * time.Second,
		MaxDepth:       2,
		Parallelism:    2,
		OutputFile:     "social_media_links.json",
	}

	socialLinks := make(map[string]struct{})

	domains, err := input.GetLandingsUrls(*cookie)
	if err != nil {
		log.Fatalf("Read domains list in tools.kontur.ru error: %v", err)
	}
	if len(domains) == 0 {
		log.Fatal("No domains found")
	}

	fmt.Printf("%d domains was loaded to scan\n", len(domains))
	// for i := range domains {
	// 	fmt.Println(domains[i])
	// }

	c := crawler.CreateCollector(config, domains)

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := strings.TrimSpace(e.Attr("href"))
		if !filter.IsValidURL(link) {
			return
		}
		abs_link := e.Request.AbsoluteURL(link)

		if abs_link == "" || !filter.IsValidURL(abs_link) {
			return
		}
		if filter.ContainsAny(abs_link, filter.SocialMediaDomains) {
			socialLinks[abs_link] = struct{}{}
		}

		_ = e.Request.Visit(abs_link)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	for i := range domains {
		err = c.Visit("https://" + domains[i])
		if err != nil {
			fmt.Printf("Request error: %v\n", err)
		}
	}

	// сохраняем уникальные ссылки в массив
	if len(socialLinks) == 0 {
		fmt.Println("Social links not found")
		return
	}
	out := make([]string, 0, len(socialLinks))
	for u := range socialLinks {
		out = append(out, u)
	}
	fmt.Println(out)
	// запись в базу данных
	// другой сервис который ходит в базу
	// + вебхук в мм
}
