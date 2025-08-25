package storage

import (
	"context"
	"fmt"
	"os"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq"

	"github.com/zeshi09/go_web_parser/ent"
)

// DatabaseConfig содержит настройки подключения к БД
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// SocialLinkService предоставляет методы для работы с социальными ссылками
type SocialLinkService struct {
	client *ent.Client
}

// LoadConfigFromEnv загружает конфигурацию из переменных окружения
func LoadConfigFromEnv() *DatabaseConfig {
	return &DatabaseConfig{
		Host:     getEnv("DB_HOST", "172.17.0.2"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "social_links_db"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// DSN возвращает строку подключения
func (cfg *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)
}

// NewSocialLinkService создает новый сервис для работы с социальными ссылками
func NewSocialLinkService(cfg *DatabaseConfig) (*SocialLinkService, error) {
	drv, err := sql.Open(dialect.Postgres, cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed opening connection to postgres: %w", err)
	}

	// Настройка пула подключений
	db := drv.DB()
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Hour)

	// Проверяем подключение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	client := ent.NewClient(ent.Driver(drv))

	// Создаем таблицы если их нет
	if err := client.Schema.Create(context.Background()); err != nil {
		return nil, fmt.Errorf("failed creating schema: %w", err)
	}

	return &SocialLinkService{client: client}, nil
}

// Close закрывает соединение с базой данных
func (s *SocialLinkService) Close() error {
	return s.client.Close()
}

// SaveSocialLinks сохраняет массив социальных ссылок в базу данных
func (s *SocialLinkService) SaveSocialLinks(ctx context.Context, socialLinks []string, sourceDomain string) error {
	if len(socialLinks) == 0 {
		return nil
	}

	// Создаем bulk операцию для массовой вставки
	bulk := make([]*ent.SocialLinkCreate, 0, len(socialLinks))

	for _, link := range socialLinks {
		bulk = append(bulk, s.client.SocialLink.Create().
			SetLink(link),
		)
	}

	// Если есть что сохранять, выполняем bulk insert
	if len(bulk) > 0 {
		_, err := s.client.SocialLink.CreateBulk(bulk...).Save(ctx)
		if err != nil {
			return fmt.Errorf("error saving social links: %w", err)
		}
	}

	return nil
}
