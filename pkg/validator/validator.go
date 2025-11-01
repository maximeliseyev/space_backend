package validator

import (
	"errors"
	"regexp"
	"strings"
)

var (
	// Регулярное выражение для безопасных имён (буквы, пробелы, дефисы)
	nameRegex = regexp.MustCompile(`^[a-zA-Zа-яА-ЯёЁ\s\-']+$`)

	// Регулярное выражение для username (буквы, цифры, подчеркивания)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
)

// SanitizeSearchQuery санитизирует поисковый запрос
func SanitizeSearchQuery(query string) (string, error) {
	// Убираем пробелы по краям
	query = strings.TrimSpace(query)

	// Проверяем минимальную длину
	if len(query) < 2 {
		return "", errors.New("search query must be at least 2 characters")
	}

	// Ограничиваем максимальную длину
	if len(query) > 100 {
		query = query[:100]
	}

	return query, nil
}

// EscapeLike экранирует специальные символы LIKE для PostgreSQL
func EscapeLike(s string) string {
	// Экранируем обратный слеш
	s = strings.ReplaceAll(s, "\\", "\\\\")
	// Экранируем %
	s = strings.ReplaceAll(s, "%", "\\%")
	// Экранируем _
	s = strings.ReplaceAll(s, "_", "\\_")
	return s
}

// ValidateName проверяет, что имя содержит только допустимые символы
func ValidateName(name string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}

	if len(name) > 50 {
		return errors.New("name is too long (max 50 characters)")
	}

	if !nameRegex.MatchString(name) {
		return errors.New("name contains invalid characters (only letters, spaces, hyphens allowed)")
	}

	return nil
}

// ValidateUsername проверяет username
func ValidateUsername(username string) error {
	if username == "" {
		return nil // Username может быть пустым
	}

	if len(username) > 32 {
		return errors.New("username is too long (max 32 characters)")
	}

	if !usernameRegex.MatchString(username) {
		return errors.New("username contains invalid characters")
	}

	return nil
}

// SanitizeText общая санитизация текстовых полей
func SanitizeText(text string, maxLength int) (string, error) {
	// Убираем пробелы по краям
	text = strings.TrimSpace(text)

	// Ограничиваем длину
	if len(text) > maxLength {
		text = text[:maxLength]
	}

	// Можно добавить дополнительную очистку здесь

	return text, nil
}