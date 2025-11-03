package telegram

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ValidateLoginWidget validates Telegram Login Widget data
// See: https://core.telegram.org/widgets/login#checking-authorization
// ttl - time to live for auth_date in seconds (e.g., 604800 for 7 days)
func ValidateLoginWidget(initData string, botToken string, ttl int64) error {
	if initData == "" {
		return fmt.Errorf("initData is empty")
	}

	// Parse URL query string
	values, err := url.ParseQuery(initData)
	if err != nil {
		return fmt.Errorf("failed to parse initData: %w", err)
	}

	// Extract hash
	hash := values.Get("hash")
	if hash == "" {
		return ErrMissingHash
	}
	values.Del("hash")

	// Проверяем auth_date (срок действия 1 день для Login Widget)
	authDateStr := values.Get("auth_date")
	if authDateStr == "" {
		return ErrInvalidAuthDate
	}

	authDate, err := strconv.ParseInt(authDateStr, 10, 64)
	if err != nil {
		return ErrInvalidAuthDate
	}

	// Проверяем, что auth_date не старше установленного TTL
	now := time.Now().Unix()
	if now-authDate > ttl {
		return ErrAuthDateExpired
	}

	// Create data-check-string для Login Widget
	// Формат: ключ=значение, отсортированные по ключу, соединённые \n
	var keys []string
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var dataCheckArr []string
	for _, key := range keys {
		dataCheckArr = append(dataCheckArr, fmt.Sprintf("%s=%s", key, values.Get(key)))
	}
	dataCheckString := strings.Join(dataCheckArr, "\n")

	// Create secret key для Login Widget
	// Для Login Widget используется SHA256(botToken) как ключ
	secretKey := sha256.Sum256([]byte(botToken))

	// Calculate hash
	calculatedHash := hmac.New(sha256.New, secretKey[:])
	calculatedHash.Write([]byte(dataCheckString))
	calculatedHashStr := hex.EncodeToString(calculatedHash.Sum(nil))

	// Compare hashes
	if calculatedHashStr != hash {
		return ErrInvalidHash
	}

	return nil
}

// ParseUserFromLoginWidget parses user data from Login Widget initData
func ParseUserFromLoginWidget(initData string) (*TelegramUser, error) {
	values, err := url.ParseQuery(initData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse initData: %w", err)
	}

	// Для Login Widget данные приходят напрямую в query параметрах
	idStr := values.Get("id")
	if idStr == "" {
		return nil, ErrMissingUserData
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	user := &TelegramUser{
		ID:           id,
		FirstName:    values.Get("first_name"),
		LastName:     values.Get("last_name"),
		Username:     values.Get("username"),
		PhotoURL:     values.Get("photo_url"),
		LanguageCode: values.Get("language_code"),
	}

	return user, nil
}

// ValidateAndParseLoginWidget validates Login Widget data and returns user data
// ttl - time to live for auth_date in seconds (e.g., 604800 for 7 days)
func ValidateAndParseLoginWidget(initData string, botToken string, ttl int64) (*TelegramUser, error) {
	// Сначала валидируем
	if err := ValidateLoginWidget(initData, botToken, ttl); err != nil {
		return nil, err
	}

	// Затем парсим данные пользователя
	user, err := ParseUserFromLoginWidget(initData)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// DetectAuthType определяет тип авторизации по формату initData
func DetectAuthType(initData string) string {
	values, err := url.ParseQuery(initData)
	if err != nil {
		return "unknown"
	}

	// Если есть параметр "user" с JSON - это Mini App
	if values.Get("user") != "" {
		return "miniapp"
	}

	// Если есть параметр "id" напрямую - это Login Widget
	if values.Get("id") != "" {
		return "loginwidget"
	}

	return "unknown"
}
