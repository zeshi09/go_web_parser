package storage

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq"

	"github.com/zeshi09/go_web_parser/ent"
	"github.com/zeshi09/go_web_parser/ent/domain"
	"github.com/zeshi09/go_web_parser/ent/sociallink"
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

// DomainsService предоставляет методы для работы со списком доменов лендинга
type DomainsService struct {
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

func OpenClient(cfg *DatabaseConfig) (*ent.Client, error) {
	drv, err := sql.Open(dialect.Postgres, cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed opening connection to postgres: %w", err)
	}

	db := drv.DB()
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Hour)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	client := ent.NewClient(ent.Driver(drv))

	if err := client.Schema.Create(context.Background()); err != nil {
		return nil, fmt.Errorf("failed creating schema: %w", err)
	}

	return client, nil
}

func NewDomainServiceWithClient(client *ent.Client) *DomainsService {
	return &DomainsService{client: client}
}

func NewSocialLinkServiceWithClient(client *ent.Client) *SocialLinkService {
	return &SocialLinkService{client: client}
}

func SaveAll(ctx context.Context, client *ent.Client, socialLinks []string, landingDomains []string) error {
	tx, err := client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	if err := NewSocialLinkServiceWithClient(tx.Client()).SaveSocialLinks(ctx, socialLinks, "batch"); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("save social links: %w", err)
	}

	if err := NewDomainServiceWithClient(tx.Client()).SaveDomain(ctx, landingDomains); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("save domains: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

// extractDomain извлекает домен из URL
func extractDomain(rawURL string) string {
	// Для telegram ссылок вида tg://
	if strings.HasPrefix(rawURL, "tg://") {
		return "tg://"
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	domain := parsed.Host
	// Убираем www. если есть
	if strings.HasPrefix(domain, "www.") {
		domain = strings.TrimPrefix(domain, "www.")
	}

	return domain
}

// func NewDomainService(cfg *DatabaseConfig) (*DomainsService, error) {

// 	drv, err := sql.Open(dialect.Postgres, cfg.DSN())
// 	if err != nil {
// 		return nil, fmt.Errorf("failed opening connection to postgres: %w", err)
// 	}

// 	// Настройка пула подключений
// 	db := drv.DB()
// 	db.SetMaxIdleConns(10)
// 	db.SetMaxOpenConns(100)
// 	db.SetConnMaxLifetime(time.Hour)

// 	// Проверяем подключение
// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()

// 	if err := db.PingContext(ctx); err != nil {
// 		return nil, fmt.Errorf("failed to ping database: %w", err)
// 	}

// 	client := ent.NewClient(ent.Driver(drv))

// 	// Создаем таблицы если их нет
// 	if err := client.Schema.Create(context.Background()); err != nil {
// 		return nil, fmt.Errorf("failed creating schema: %w", err)
// 	}

// 	return &DomainsService{client: client}, nil
// }

func (s *DomainsService) Close() error {
	return s.client.Close()
}

func (s *DomainsService) SaveDomain(ctx context.Context, landingDomain []string) error {
	bulk := make([]*ent.DomainCreate, 0, len(landingDomain))
	for i := range landingDomain {
		exists, err := s.client.Domain.Query().
			Where(domain.LandingDomain(landingDomain[i])).
			Exist(ctx)
		if err != nil {
			return err
		}
		if exists {
			return nil
		}

		bulk = append(bulk, s.client.Domain.Create().
			SetLandingDomain(landingDomain[i]),
		)
	}

	if len(bulk) > 0 {
		_, err := s.client.Domain.CreateBulk(bulk...).Save(ctx)
		if err != nil {
			return fmt.Errorf("error saving social links: %w", err)
		}
	}

	return nil
}

// // NewSocialLinkService создает новый сервис для работы с социальными ссылками
// func NewSocialLinkService(cfg *DatabaseConfig) (*SocialLinkService, error) {
// 	drv, err := sql.Open(dialect.Postgres, cfg.DSN())
// 	if err != nil {
// 		return nil, fmt.Errorf("failed opening connection to postgres: %w", err)
// 	}

// 	// Настройка пула подключений
// 	db := drv.DB()
// 	db.SetMaxIdleConns(10)
// 	db.SetMaxOpenConns(100)
// 	db.SetConnMaxLifetime(time.Hour)

// 	// Проверяем подключение
// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()

// 	if err := db.PingContext(ctx); err != nil {
// 		return nil, fmt.Errorf("failed to ping database: %w", err)
// 	}

// 	client := ent.NewClient(ent.Driver(drv))

// 	// Создаем таблицы если их нет
// 	if err := client.Schema.Create(context.Background()); err != nil {
// 		return nil, fmt.Errorf("failed creating schema: %w", err)
// 	}

// 	return &SocialLinkService{client: client}, nil
// }

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
		domain := extractDomain(link)

		// Проверяем, существует ли уже такая ссылка
		exists, err := s.client.SocialLink.Query().
			Where(sociallink.URL(link)).
			Exist(ctx)
		if err != nil {
			return fmt.Errorf("error checking if link exists: %w", err)
		}

		// Если ссылка уже существует, пропускаем
		if exists {
			continue
		}

		bulk = append(bulk, s.client.SocialLink.Create().
			SetURL(link).      // Устанавливаем и URL
			SetDomain(domain), // Устанавливаем домен
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
