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

	// обозначаем время в формате unix для логов
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// считываем куки из аргумента -cookie при запуске команды
	cookie := flag.String("cookie", "", "Your cookie for tools.kontur.ru")
	flag.Parse()

	// конфигурируем какую-то хуйню, которую мы потом, скорее всего, просто снесем
	config := config.Config{
		// ProxyURL:       "http://192.168.2.200:8080",
		RequestTimeout: 10 * time.Second,
		MaxDepth:       2,
		Parallelism:    15,
		OutputFile:     "social_media_links.json",
	}

	// объявляем мап для хранения ссылок
	socialLinks := make(map[string]struct{})

	// объявляем слайс для доменов и ошибку
	domains := []string{}
	err := error(nil)

	// for {
	// 	domains, err = input.GetLandingsUrls(*cookie)
	// 	if err != nil {
	// 		log.Error().Err(err).Msg("Failed to read domains list from tools.kontur.ru")
	// 		time.Sleep(10 * time.Second)
	// 		continue
	// 	}
	// 	if len(domains) == 0 {
	// 		log.Warn().Msg("No domains found")
	// 		time.Sleep(10 * time.Second)
	// 		continue
	// 	}
	// 	log.Printf("%d domains was loaded to scan\n", len(domains))
	// 	break
	// }

	// получаем сущность краулера при помощи написанной функции
	c := crawler.CreateCollector(config, domains)

	// описываем алгоритм действия для коллектора для случая, в котором он получает html
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

	// c.OnRequest(func(r *colly.Request) {
	// 	// fmt.Println("Visiting", r.URL.String())
	// })

	// основной цикл прохождения краулера по каждому из списка доменов, которые мы получили в функции GetLandingsUrls
	for i := range domains {
		err := c.Visit("https://" + domains[i])
		if err != nil {
			fmt.Printf("Request error: %v\n", err)
		}
	}

	// for i := 0; i < 10; i++ {
	// 	err := c.Visit("https://" + domains[i])
	// 	if err != nil {
	// 		log.Err(err).Msg("Request error")
	// 	}
	// }

	// сохраняем уникальные ссылки в мап
	if len(socialLinks) == 0 {
		fmt.Println("Social links not found")
		return
	}
	out := make([]string, 0, len(socialLinks))
	for u := range socialLinks {
		out = append(out, u)
	}
	log.Debug().Strs("Links", out).Msg("Найденные ссылки")

	// записываем домены и ссылки в базу данных
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

	if err = loop(ctx, domains, cookie); err != nil {
		log.Fatal().Msg("loop error")
	}
}

func loop(ctx context.Context, domains []string, cookie *string) error {
	t := time.NewTicker(24 * time.Hour)
	defer t.Stop()

	var err error = error(nil)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
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

		}
	}

}
