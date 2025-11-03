package telegram

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"testing"
	"time"
)

// Helper function to create valid initData for testing
func createValidInitData(botToken string) string {
	authDate := time.Now().Unix()
	userJSON := `{"id":12345,"first_name":"Test","last_name":"User","username":"testuser","language_code":"en"}`

	// Create data pairs
	data := map[string]string{
		"auth_date": fmt.Sprintf("%d", authDate),
		"user":      userJSON,
	}

	// Create data-check-string
	var keys []string
	for key := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var pairs []string
	for _, key := range keys {
		pairs = append(pairs, fmt.Sprintf("%s=%s", key, data[key]))
	}
	dataCheckString := strings.Join(pairs, "\n")

	// Create secret key
	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(botToken))

	// Calculate hash
	calculatedHash := hmac.New(sha256.New, secretKey.Sum(nil))
	calculatedHash.Write([]byte(dataCheckString))
	hash := hex.EncodeToString(calculatedHash.Sum(nil))

	// Build query string
	values := url.Values{}
	values.Set("auth_date", data["auth_date"])
	values.Set("user", data["user"])
	values.Set("hash", hash)

	return values.Encode()
}

func TestValidateInitData_Success(t *testing.T) {
	botToken := "test_bot_token_123456"
	initData := createValidInitData(botToken)

	err := ValidateInitData(initData, botToken, 3600) // 1 hour TTL
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestValidateInitData_EmptyInitData(t *testing.T) {
	err := ValidateInitData("", "test_token", 3600)
	if err == nil {
		t.Error("Expected error for empty initData")
	}
}

func TestValidateInitData_MissingHash(t *testing.T) {
	initData := "auth_date=1234567890&user={}"
	err := ValidateInitData(initData, "test_token", 3600)
	if err != ErrMissingHash {
		t.Errorf("Expected ErrMissingHash, got: %v", err)
	}
}

func TestValidateInitData_InvalidHash(t *testing.T) {
	authDate := time.Now().Unix()
	initData := fmt.Sprintf("auth_date=%d&user={}&hash=invalid", authDate)
	err := ValidateInitData(initData, "test_token", 3600)
	if err != ErrInvalidHash {
		t.Errorf("Expected ErrInvalidHash, got: %v", err)
	}
}

func TestValidateInitData_ExpiredAuthDate(t *testing.T) {
	// Create initData with expired auth_date (2 hours ago)
	expiredAuthDate := time.Now().Unix() - 7200
	userJSON := `{"id":12345,"first_name":"Test"}`

	data := map[string]string{
		"auth_date": fmt.Sprintf("%d", expiredAuthDate),
		"user":      userJSON,
	}

	// Calculate valid hash for expired data
	botToken := "test_token"
	var keys []string
	for key := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var pairs []string
	for _, key := range keys {
		pairs = append(pairs, fmt.Sprintf("%s=%s", key, data[key]))
	}
	dataCheckString := strings.Join(pairs, "\n")

	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(botToken))

	calculatedHash := hmac.New(sha256.New, secretKey.Sum(nil))
	calculatedHash.Write([]byte(dataCheckString))
	hash := hex.EncodeToString(calculatedHash.Sum(nil))

	values := url.Values{}
	values.Set("auth_date", data["auth_date"])
	values.Set("user", data["user"])
	values.Set("hash", hash)

	initData := values.Encode()

	err := ValidateInitData(initData, botToken, 3600) // 1 hour TTL
	if err != ErrAuthDateExpired {
		t.Errorf("Expected ErrAuthDateExpired, got: %v", err)
	}
}

func TestParseUserFromInitData_Success(t *testing.T) {
	userJSON := `{"id":12345,"first_name":"Test","last_name":"User","username":"testuser","language_code":"en"}`
	values := url.Values{}
	values.Set("user", userJSON)
	initData := values.Encode()

	user, err := ParseUserFromInitData(initData)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if user.ID != 12345 {
		t.Errorf("Expected ID 12345, got: %d", user.ID)
	}
	if user.FirstName != "Test" {
		t.Errorf("Expected FirstName 'Test', got: %s", user.FirstName)
	}
	if user.LastName != "User" {
		t.Errorf("Expected LastName 'User', got: %s", user.LastName)
	}
	if user.Username != "testuser" {
		t.Errorf("Expected Username 'testuser', got: %s", user.Username)
	}
	if user.LanguageCode != "en" {
		t.Errorf("Expected LanguageCode 'en', got: %s", user.LanguageCode)
	}
}

func TestParseUserFromInitData_MissingUser(t *testing.T) {
	initData := "auth_date=1234567890"
	_, err := ParseUserFromInitData(initData)
	if err != ErrMissingUserData {
		t.Errorf("Expected ErrMissingUserData, got: %v", err)
	}
}

func TestValidateAndParseInitData_Success(t *testing.T) {
	botToken := "test_bot_token_123456"
	initData := createValidInitData(botToken)

	user, err := ValidateAndParseInitData(initData, botToken, 3600) // 1 hour TTL
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if user.ID != 12345 {
		t.Errorf("Expected ID 12345, got: %d", user.ID)
	}
	if user.FirstName != "Test" {
		t.Errorf("Expected FirstName 'Test', got: %s", user.FirstName)
	}
}

func TestValidateAndParseInitData_InvalidHash(t *testing.T) {
	authDate := time.Now().Unix()
	initData := fmt.Sprintf("auth_date=%d&user={}&hash=invalid", authDate)
	_, err := ValidateAndParseInitData(initData, "test_token", 3600)
	if err != ErrInvalidHash {
		t.Errorf("Expected ErrInvalidHash, got: %v", err)
	}
}
