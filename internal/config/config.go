package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	ServerPort           string
	DatabaseURL          string
	TelegramBotToken     string
	AllowedChatID        int64    // Telegram group chat ID for membership check
	JWTSecret            string
	StoragePath          string
	Environment          string
	SupabaseURL          string
	SupabaseKey          string
	AllowedOrigins       []string // CORS allowed origins
	AuthDateTTLMiniApp   int64    // TTL for Mini App auth_date in seconds (default: 3600 = 1 hour)
	AuthDateTTLLoginWidget int64  // TTL for Login Widget auth_date in seconds (default: 604800 = 7 days)
	BotAPIToken          string   // Secret token for bot API authentication
	BotWebhookURL        string   // URL of the bot webhook for sending notifications
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Загружаем .env файл (игнорируем ошибку если файла нет)
	_ = godotenv.Load()

	// JWT Secret validation - критически важно для безопасности
	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required and must not be empty")
	}
	if len(jwtSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters long for security")
	}

	// Parse AllowedChatID
	allowedChatID := int64(0)
	if chatIDStr := getEnv("ALLOWED_CHAT_ID", ""); chatIDStr != "" {
		if parsed, err := strconv.ParseInt(chatIDStr, 10, 64); err == nil {
			allowedChatID = parsed
		}
	}

	// Parse Auth Date TTL values
	authDateTTLMiniApp := parseInt64WithDefault(getEnv("AUTH_DATE_TTL_MINIAPP", ""), 3600) // 1 hour default
	authDateTTLLoginWidget := parseInt64WithDefault(getEnv("AUTH_DATE_TTL_LOGIN_WIDGET", ""), 604800) // 7 days default

	config := &Config{
		ServerPort:           getEnv("SERVER_PORT", "8080"),
		DatabaseURL:          getEnv("DATABASE_URL", ""),
		TelegramBotToken:     getEnv("TELEGRAM_BOT_TOKEN", ""),
		AllowedChatID:        allowedChatID,
		JWTSecret:            jwtSecret,
		StoragePath:          getEnv("STORAGE_PATH", "./storage"),
		Environment:          getEnv("ENVIRONMENT", "development"),
		SupabaseURL:          getEnv("SUPABASE_URL", ""),
		SupabaseKey:          getEnv("SUPABASE_SECRET_KEY", ""),
		AllowedOrigins:       parseAllowedOrigins(getEnv("ALLOWED_ORIGINS", "")),
		AuthDateTTLMiniApp:   authDateTTLMiniApp,
		AuthDateTTLLoginWidget: authDateTTLLoginWidget,
		BotAPIToken:          getEnv("BOT_API_TOKEN", ""),
		BotWebhookURL:        getEnv("BOT_WEBHOOK_URL", "http://localhost:8081"),
	}

	// Если DATABASE_URL не задан, но есть SUPABASE_URL - строим DATABASE_URL из Supabase
	if config.DatabaseURL == "" && config.SupabaseURL != "" {
		config.DatabaseURL = buildSupabaseDatabaseURL(config.SupabaseURL)
		// Не выводим DATABASE_URL даже в development режиме из соображений безопасности
	}

	// Валидация обязательных параметров
	if config.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL or SUPABASE_URL is required")
	}

	if config.TelegramBotToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}

	if config.BotAPIToken == "" {
		return nil, fmt.Errorf("BOT_API_TOKEN is required for bot authentication")
	}

	if len(config.BotAPIToken) < 32 {
		return nil, fmt.Errorf("BOT_API_TOKEN must be at least 32 characters long for security")
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

	// Строим полный connection string
	// TimeZone=UTC важно для корректной работы с часовыми поясами
	return fmt.Sprintf("postgresql://postgres:%s@db.%s.supabase.co:5432/postgres?sslmode=require&TimeZone=UTC", password, projectRef)
}

// getEnv gets environment variable or returns default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseInt64WithDefault parses string to int64 or returns default value
func parseInt64WithDefault(value string, defaultValue int64) int64 {
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return defaultValue
	}
	return parsed
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

// parseAllowedOrigins парсит строку разрешённых origins, разделённых запятыми
func parseAllowedOrigins(originsStr string) []string {
	if originsStr == "" {
		// Дефолтные origins для разработки
		return []string{
			"http://localhost:5173",
			"http://localhost:5174",
			"http://localhost:3000",
			"https://telegram.org",
		}
	}

	// Разбиваем по запятой и убираем пробелы
	origins := strings.Split(originsStr, ",")
	result := make([]string, 0, len(origins))
	for _, origin := range origins {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
