package main

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/zeshi09/go_web_parser/ent"
	"github.com/zeshi09/go_web_parser/internal/config"
	"github.com/zeshi09/go_web_parser/internal/crawler"
	"github.com/zeshi09/go_web_parser/internal/filter"
	"github.com/zeshi09/go_web_parser/internal/input"
	"github.com/zeshi09/go_web_parser/internal/storage"
)

func main() {
	var (
		cookie = flag.String("cookie", "", "Your cookie for tools.kontur.ru")
		config = config.Config{
			ProxyURL:       "http://192.168.2.200:8080",
			RequestTimeout: 10 * time.Second,
			MaxDepth:       2,
			Parallelism:    15,
			OutputFile:     "social_media_links.json",
		}
		socialLinks = make(map[string]string)
		domains     = []string{}
		c           = crawler.CreateCollector(config, domains)
	)

	// подгружаем .env файл, в котором хранятся все переменные для базы
	err := godotenv.Load()
	if err != nil {
		log.Error().Err(err).Msg("Failed to load .env")
	}

	// обозначаем время в формате unix для логов
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// считываем куки из аргумента -cookie при запуске команды
	flag.Parse()

	// описываем алгоритм действия для коллектора для случая, в котором он получает html
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := strings.TrimSpace(e.Attr("href"))
		if !filter.IsValidURL(link) {
			return
		}
		abs_link := e.
			Request.
			AbsoluteURL(link)
		if abs_link == "" || !filter.IsValidURL(abs_link) {
			return
		}
		clean_abs_link := filter.CleanPath(abs_link)
		if filter.ContainsAny(clean_abs_link, filter.SocialMediaDomains) {
			socialLinks[clean_abs_link] = e.Request.URL.String()
		}
		_ = e.Request.Visit(abs_link)
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// записываем домены и ссылки в базу данных
	cfg := storage.LoadConfigFromEnv()

	client, err := storage.OpenClient(cfg) // единый клиент для сервиса ссылок и доменов
	if err != nil {
		log.
			Err(err).
			Msg("failed to connect db")
	}
	defer client.Close()
	log.
		Info().
		Msg("Database connected successfully")

	if err = loop(ctx, domains, cookie, socialLinks, c, client); err != nil {
		log.
			Fatal().
			Msg("loop error")
	}
}

func loop(ctx context.Context, domains []string, cookie *string, socialLinks map[string]string, c *colly.Collector, client *ent.Client) error {

	// первоначальный скан, чтобы не ждать 24 часа
	if err := looploop(ctx, domains, cookie, socialLinks, c, client); err != nil {
		return err
	}

	// установка таймера скана
	t := time.NewTicker(24 * time.Hour)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
			if err := looploop(ctx, domains, cookie, socialLinks, c, client); err != nil {
				log.
					Error().
					Err(err).
					Msg("Scan cycle failed")
				continue
			}
		}
	}

}

func looploop(ctx context.Context, domains []string, cookie *string, socialLinks map[string]string, c *colly.Collector, client *ent.Client) error {

	var err error
	// Получаем домены из веб студии
	domains, err = input.GetLandingsUrls(*cookie)
	if err != nil {
		log.
			Error().
			Err(err).
			Msg("Failed to read domains list from tools.kontur.ru")
		time.Sleep(10 * time.Second)
	}
	if len(domains) == 0 {
		log.
			Warn().
			Msg("No domains found")
		time.Sleep(10 * time.Second)
	}
	log.Printf("%d domains was loaded to scan\n", len(domains))

	// основной цикл прохождения краулера по каждому из списка доменов, которые мы получили в функции GetLandingsUrls
	// for i := range domains {
	// 	err := c.Visit("https://" + domains[i])
	// 	if err != nil {
	// 		fmt.Printf("Request error: %v\n", err)
	// 	}
	// }
	for i := 0; i < 7; i++ {
		err := c.Visit("https://" + domains[i])
		if err != nil {
			fmt.Printf("Request error: %v\n", err)
		}
	}
	c.Wait()

	// сохраняем уникальные ссылки в мап
	if len(socialLinks) == 0 {
		log.
			Warn().
			Msg("No social links was found")
		return fmt.Errorf("social links not found")
	}
	// out := make([]string, 0, len(socialLinks))
	// for u := range socialLinks {
	// 	out = append(out, u)
	// }

	// log.
	// 	Debug().
	// 	Strs("Links", out).
	// 	Msg("Links was found")

	if err := storage.SaveAll(ctx, client, socialLinks, domains); err != nil {
		log.
			Err(err).
			Msg("error saving to db")
	} else {
		log.Printf("successfully saved links and domains to db")
	}

	return nil
}
