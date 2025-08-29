package main

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/zeshi09/go_web_parser/internal/config"
	"github.com/zeshi09/go_web_parser/internal/crawler"
	"github.com/zeshi09/go_web_parser/internal/filter"
	"github.com/zeshi09/go_web_parser/internal/input"
	"github.com/zeshi09/go_web_parser/internal/storage"
)

func main() {

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	cookie := flag.String("cookie", "", "Your cookie for tools.kontur.ru")
	flag.Parse()

	config := config.Config{
		ProxyURL:       "http://192.168.2.200:8080",
		RequestTimeout: 10 * time.Second,
		MaxDepth:       2,
		Parallelism:    15,
		OutputFile:     "social_media_links.json",
	}

	socialLinks := make(map[string]struct{})

	domains := []string{}
	err := error(nil)

	for {
		domains, err = input.GetLandingsUrls(*cookie)
		if err != nil {
			log.Error().Err(err).Msg("Failed to read domains list from tools.kontur.ru")
			time.Sleep(10 * time.Second)
			continue
		}
		if len(domains) == 0 {
			log.Warn().Msg("No domains found")
			time.Sleep(10 * time.Second)
			continue
		}
		log.Printf("%d domains was loaded to scan\n", len(domains))
		break
	}

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

		clean_abs_link := filter.CleanPath(abs_link)

		if filter.ContainsAny(clean_abs_link, filter.SocialMediaDomains) {
			socialLinks[clean_abs_link] = struct{}{}
		}

		_ = e.Request.Visit(abs_link)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	// for i := range domains {
	// 	err := c.Visit("https://" + domains[i])
	// 	if err != nil {
	// 		fmt.Printf("Request error: %v\n", err)
	// 	}
	// }

	for i := 0; i < 4; i++ {
		err := c.Visit("https://" + domains[i])
		if err != nil {
			log.Err(err).Msg("Request error")
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

	cfg := storage.LoadConfigFromEnv()

	client, err := storage.OpenClient(cfg) // единый клиент для сервиса ссылок и доменов
	if err != nil {
		log.Err(err).Msg("failed to connect db")
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := storage.SaveAll(ctx, client, out, domains); err != nil {
		log.Err(err).Msg("error saving to db")
	} else {
		log.Printf("successfully saved links and domains to db")
	}

	// // запись в базу данных
	// dbConfig := storage.LoadConfigFromEnv()
	// dbService, err := storage.NewSocialLinkService(dbConfig)
	// if err != nil {
	// 	log.Printf("failes to connect to db: %v", err)
	// }
	// defer dbService.Close()
	// fmt.Println("Connected to db successfully")

	// ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	// err = dbService.SaveSocialLinks(ctx, out, "batch_"+fmt.Sprintf("%d", len(domains)))
	// cancel()

	// dbConfig2 := storage.LoadConfigFromEnv()
	// dbService2, err := storage.NewDomainService(dbConfig2)

	// if err != nil {
	// 	log.Printf("failes to connect to db: %v", err)
	// }
	// defer dbService2.Close()
	// fmt.Println("Connected to db successfully")

	// ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	// err = dbService2.SaveDomain(ctx, domains)
	// cancel()

	// if err != nil {
	// 	log.Printf("error saving to db: %v", err)
	// } else {
	// 	log.Printf("successfully saved links to db\n")
	// }
}
