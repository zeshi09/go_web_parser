package config

import (
	"time"
)

// Основной конфиг для приложения
type Config struct {
	ProxyURL       string
	RequestTimeout time.Duration
	MaxDepth       int
	Parallelism    int
	OutputFile     string
}
