package filter

import (
	"net/url"
	"strings"

	"github.com/rs/zerolog/log"
)

// Домены, по которым мы ищем ссылки в социальных сетях
var SocialMediaDomains []string = []string{
	"t.me",
	"tg://",
	"vk.com",
	"rutube.ru",
	"/ok.ru",
	"youtube.com",
	"youtu.be",
	"vkvideo.ru",
	"oneme.ru",
	"max.ru",
	"dzen.",
}

func CleanPath(abs_link string) string {
	parsed, err := url.Parse(abs_link)
	if err != nil {
		log.Err(err).Msg("error to parse url")
	}
	parsed.RawQuery = ""
	parsed.Fragment = ""

	return strings.TrimSuffix(parsed.String(), "/")

}

// Функция для поиска вхождения в списке доменов
func ContainsAny(str string, list []string) bool {
	for _, item := range list {
		if strings.Contains(str, item) {
			return true
		}
	}
	return false
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
