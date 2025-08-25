package crawler

import (
	"time"
	"github.com/gocolly/colly"
	"github.com/zeshi09/go_web_parser/internal/config"

)

// Создаем сущность скрапера
func CreateCollector(config Config, domains []string) *colly.Collector {
	c := colly.NewCollector(
		colly.UserAgent(
			"Mozilla/5.0 (X11; Linux x86_64; rv:141.0) Gecko/20100101 Firefox/141.0",
		),
		colly.AllowedDomains(
			domains...,
		),
	)

	if config.ProxyURL != "" {
		c.SetProxy(config.ProxyURL)
	}

	c.SetRequestTimeout(config.RequestTimeout)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: config.Parallelism,
		Delay:       1 * time.Second,
	})

	if config.MaxDepth > 0 {
		c.MaxDepth = config.MaxDepth
	}

	return c
}
