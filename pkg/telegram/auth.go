package telegram

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

var (
	ErrInvalidHash     = errors.New("invalid hash")
	ErrMissingHash     = errors.New("missing hash parameter")
	ErrMissingUserData = errors.New("missing user data")
)

// TelegramUser represents Telegram user data from initData
type TelegramUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
	PhotoURL  string `json:"photo_url,omitempty"`
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

	// В реальном initData от Telegram пользователь передается как JSON в параметре "user"
	// Для упрощения можем использовать отдельные поля
	userJSON := values.Get("user")
	if userJSON == "" {
		return nil, ErrMissingUserData
	}

	// TODO: Здесь нужно распарсить JSON пользователя
	// Для простоты используем прямые поля (это нужно будет доработать)

	return &TelegramUser{
		// Временная заглушка - в реальности парсить из JSON
	}, nil
}
