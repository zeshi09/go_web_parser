package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gocolly/colly"
)

type Findings struct {
	Url      string   `json:"url"`
	Findings []string `json:"findings"`
}

var list []string = []string{
	"t.me",
	"vk.com",
	"rutube.ru",
	"ok.ru",
}

func containsAny(str string, list []string) bool {
	for _, item := range list {
		if strings.Contains(str, item) {
			return true
		}
	}
	return false
}

func ReadUrlFile(path string) ([]string, error) {
	file, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	var domainlist []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		domainlist = append(domainlist, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return domainlist, nil
}

func main() {
	links := make(map[string]struct{})
	domains, err := ReadUrlFile("landings_urls.txt")

	if err != nil {
		log.Fatalf("Read file error: %v", err)
	}

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (X11; Linux x86_64; rv:141.0) Gecko/20100101 Firefox/141.0"),
		colly.AllowedDomains(domains...),
	)

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")

		if containsAny(link, list) {
			links[link] = struct{}{}
		}

		c.Visit(e.Request.AbsoluteURL(link))
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	for i := range domains {
		c.Visit("https://" + domains[i])
	}

	fmt.Println(links)
}
