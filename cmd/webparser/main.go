package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/zeshi09/go_web_parser/internal/config"
	"github.com/zeshi09/go_web_parser/internal/input"
	"github.com/zeshi09/go_web_parser/internal/crawler"
	"github.com/zeshi09/go_web_parser/internal/storage"
)

type Findings struct {
	Url      string   `json:"url"`
	Findings []string `json:"findings"`
}

// // Основной конфиг для приложения
// type Config struct {
// 	ProxyURL       string
// 	RequestTimeout time.Duration
// 	MaxDepth       int
// 	Parallelism    int
// 	OutputFile     string
// }

// Домены, по которым мы ищем ссылки в социальных сетях
var socialMediaDomains []string = []string{
	"t.me",
	"tg://",
	"vk.com",
	"rutube.ru",
	"ok.ru",
	"youtube.com",
	"youtu.be",
	"vkvideo.ru",
	"oneme.ru",
	"max.ru",
	"dzen",
}

// Функция для поиска вхождения в списке доменов
func containsAny(str string, list []string) bool {
	for _, item := range list {
		if strings.Contains(str, item) {
			return true
		}
	}
	return false
}

// Функция чтения файла построчно
func ReadUrlFile(path string) ([]string, error) {
	file, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	var domainlist []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			domainlist = append(domainlist, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return domainlist, nil
}

// Проверяем, что это валидная ссылка, чтобы избежать ошибок
func IsValidURL(rawURL string) bool {
	if rawURL == "" {
		return false
	}

	// Skip javascript:, mailto:, tel: links
	if strings.HasPrefix(strings.ToLower(rawURL), "javascript:") ||
		strings.HasPrefix(strings.ToLower(rawURL), "mailto:") ||
		strings.HasPrefix(strings.ToLower(rawURL), "tel:") {
		return false
	}

	_, err := url.Parse(rawURL)
	return err == nil
}

// // Создаем сущность скрапера
// func CreateCollector(config Config, domains []string) *colly.Collector {
// 	c := colly.NewCollector(
// 		colly.UserAgent(
// 			"Mozilla/5.0 (X11; Linux x86_64; rv:141.0) Gecko/20100101 Firefox/141.0",
// 		),
// 		colly.AllowedDomains(
// 			domains...,
// 		),
// 	)

// 	if config.ProxyURL != "" {
// 		c.SetProxy(config.ProxyURL)
// 	}

// 	c.SetRequestTimeout(config.RequestTimeout)

// 	c.Limit(&colly.LimitRule{
// 		DomainGlob:  "*",
// 		Parallelism: config.Parallelism,
// 		Delay:       1 * time.Second,
// 	})

// 	if config.MaxDepth > 0 {
// 		c.MaxDepth = config.MaxDepth
// 	}

// 	return c
// }

func main() {

	cookie := flag.String("cookie", "", "Your cookie for tools.kontur.ru")
	flag.Parse()

	config := Config{
		ProxyURL:       "http://192.168.2.200:8080",
		RequestTimeout: 20 * time.Second,
		MaxDepth:       2,
		Parallelism:    2,
		OutputFile:     "social_media_links.json",
	}

	socialLinks := make(map[string]struct{})

	// domains, err := ReadUrlFile("landings_urls.txt")
	domains, err := internal.GetLandingsUrls(*cookie)
	if err != nil {
		log.Fatalf("Read domains list in tools.kontur.ru error: %v", err)
	}
	if len(domains) == 0 {
		log.Fatal("No domains found")
	}

	fmt.Printf("%d domains was loaded to scan\n", len(domains))

	c := internal.CreateCollector(config, domains)

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := strings.TrimSpace(e.Attr("href"))
		if !IsValidURL(link) {
			return
		}
		abs_link := e.Request.AbsoluteURL(link)

		if abs_link == "" || !IsValidURL(abs_link) {
			return
		}
		if containsAny(abs_link, socialMediaDomains) {
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
