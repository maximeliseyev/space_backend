package telegram

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidHash     = errors.New("invalid hash")
	ErrMissingHash     = errors.New("missing hash parameter")
	ErrMissingUserData = errors.New("missing user data")
	ErrAuthDateExpired = errors.New("auth_date expired (older than 1 hour)")
	ErrInvalidAuthDate = errors.New("invalid auth_date")
)

// TelegramUser represents Telegram user data from initData
type TelegramUser struct {
	ID           int64  `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name,omitempty"`
	Username     string `json:"username,omitempty"`
	PhotoURL     string `json:"photo_url,omitempty"`
	LanguageCode string `json:"language_code,omitempty"`
}

// ValidateInitData validates Telegram Mini App initData
// See: https://core.telegram.org/bots/webapps#validating-data-received-via-the-mini-app
func ValidateInitData(initData string, botToken string) error {
	if initData == "" {
		return errors.New("initData is empty")
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

	// Проверяем auth_date (срок действия 1 час)
	authDateStr := values.Get("auth_date")
	if authDateStr == "" {
		return ErrInvalidAuthDate
	}

	authDate, err := strconv.ParseInt(authDateStr, 10, 64)
	if err != nil {
		return ErrInvalidAuthDate
	}

	// Проверяем, что auth_date не старше 1 часа (3600 секунд)
	now := time.Now().Unix()
	if now-authDate > 3600 {
		return ErrAuthDateExpired
	}

	// Create data-check-string
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

	// Create secret key
	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(botToken))

	// Calculate hash
	calculatedHash := hmac.New(sha256.New, secretKey.Sum(nil))
	calculatedHash.Write([]byte(dataCheckString))
	calculatedHashStr := hex.EncodeToString(calculatedHash.Sum(nil))

	// Compare hashes
	if calculatedHashStr != hash {
		return ErrInvalidHash
	}

	return nil
}

// ParseUserFromInitData parses user data from initData query string
func ParseUserFromInitData(initData string) (*TelegramUser, error) {
	values, err := url.ParseQuery(initData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse initData: %w", err)
	}

	// Получаем JSON объект user из параметра (он URL-encoded)
	userJSON := values.Get("user")
	if userJSON == "" {
		return nil, ErrMissingUserData
	}

	// Декодируем URL-encoded JSON и парсим
	var user TelegramUser
	if err := json.Unmarshal([]byte(userJSON), &user); err != nil {
		return nil, fmt.Errorf("failed to parse user JSON: %w", err)
	}

	return &user, nil
}

// ValidateAndParseInitData validates initData and returns user data
// Это комбинированная функция для удобства
func ValidateAndParseInitData(initData string, botToken string) (*TelegramUser, error) {
	// Сначала валидируем
	if err := ValidateInitData(initData, botToken); err != nil {
		return nil, err
	}

	// Затем парсим данные пользователя
	user, err := ParseUserFromInitData(initData)
	if err != nil {
		return nil, err
	}

	return user, nil
}
