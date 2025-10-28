package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	ServerPort       string
	DatabaseURL      string
	TelegramBotToken string
	JWTSecret        string
	StoragePath      string
	Environment      string
	SupabaseURL      string
	SupabaseKey      string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Загружаем .env файл (игнорируем ошибку если файла нет)
	_ = godotenv.Load()

	config := &Config{
		ServerPort:       getEnv("SERVER_PORT", "8080"),
		DatabaseURL:      getEnv("DATABASE_URL", ""),
		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		JWTSecret:        getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		StoragePath:      getEnv("STORAGE_PATH", "./storage"),
		Environment:      getEnv("ENVIRONMENT", "development"),
		SupabaseURL:      getEnv("SUPABASE_URL", ""),
		SupabaseKey:      getEnv("SUPABASE_SECRET_KEY", ""),
	}

	// Если DATABASE_URL не задан, но есть SUPABASE_URL - строим DATABASE_URL из Supabase
	if config.DatabaseURL == "" && config.SupabaseURL != "" {
		config.DatabaseURL = buildSupabaseDatabaseURL(config.SupabaseURL)
		// Вывод для отладки (скрываем пароль)
		if config.Environment == "development" {
			maskedURL := maskPassword(config.DatabaseURL)
			fmt.Printf("📊 Built DATABASE_URL from SUPABASE_URL: %s\n", maskedURL)
		}
	}

	// Валидация обязательных параметров
	if config.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL or SUPABASE_URL is required")
	}

	if config.TelegramBotToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}

	return config, nil
}

// buildSupabaseDatabaseURL строит PostgreSQL connection string из Supabase URL
// Пример: https://xxx.supabase.co → postgres://postgres:[password]@db.xxx.supabase.co:5432/postgres
func buildSupabaseDatabaseURL(supabaseURL string) string {
	// Извлекаем project reference из URL
	// https://abcdefghijklmn.supabase.co → abcdefghijklmn
	supabaseURL = strings.TrimPrefix(supabaseURL, "https://")
	supabaseURL = strings.TrimPrefix(supabaseURL, "http://")
	supabaseURL = strings.TrimSuffix(supabaseURL, "/")

	parts := strings.Split(supabaseURL, ".")
	if len(parts) < 1 {
		return ""
	}

	projectRef := parts[0]

	// Получаем пароль из переменной окружения
	password := getEnv("SUPABASE_DB_PASSWORD", getEnv("DB_PASSWORD", ""))

	if password == "" {
		// Если пароль не задан, возвращаем URL с плейсхолдером
		// Пользователь должен будет заменить [YOUR-PASSWORD]
		return fmt.Sprintf("postgresql://postgres:[YOUR-PASSWORD]@db.%s.supabase.co:5432/postgres?sslmode=require", projectRef)
	}

	// Строим полный connection string с принудительным IPv4
	// prefer_simple=true помогает избежать проблем с IPv6
	return fmt.Sprintf("postgresql://postgres:%s@db.%s.supabase.co:5432/postgres?sslmode=require&prefer_simple=true", password, projectRef)
}

// getEnv gets environment variable or returns default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// maskPassword скрывает пароль в connection string для безопасного вывода
func maskPassword(connStr string) string {
	// Находим пароль между postgres: и @
	parts := strings.Split(connStr, "postgres:")
	if len(parts) < 2 {
		return connStr
	}

	afterUser := parts[1]
	atIndex := strings.Index(afterUser, "@")
	if atIndex == -1 {
		return connStr
	}

	return "postgresql://postgres:***@" + afterUser[atIndex+1:]
}
