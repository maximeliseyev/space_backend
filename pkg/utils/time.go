package utils

import (
	"fmt"
	"strings"
	"time"
)

// FlexibleTime поддерживает парсинг нескольких форматов времени
type FlexibleTime struct {
	time.Time
}

// UnmarshalJSON парсит несколько форматов времени
func (ft *FlexibleTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")

	if s == "null" || s == "" {
		ft.Time = time.Time{}
		return nil
	}

	// Список поддерживаемых форматов
	formats := []string{
		time.RFC3339,                 // 2006-01-02T15:04:05Z07:00
		time.RFC3339Nano,             // 2006-01-02T15:04:05.999999999Z07:00
		"2006-01-02T15:04:05",        // БЕЗ таймзоны
		"2006-01-02 15:04:05",        // Формат PostgreSQL
		"2006-01-02 15:04:05-07:00",  // PostgreSQL с таймзоной
		"2006-01-02 15:04:05+00",     // Формат из БД
		"2006-01-02",                 // Только дата
	}

	var parseErr error
	for _, format := range formats {
		t, err := time.Parse(format, s)
		if err == nil {
			ft.Time = t
			return nil
		}
		parseErr = err
	}

	return fmt.Errorf("unable to parse time %q: %w", s, parseErr)
}

// MarshalJSON возвращает время в формате RFC3339
func (ft FlexibleTime) MarshalJSON() ([]byte, error) {
	if ft.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", ft.Time.Format(time.RFC3339))), nil
}

// ParseFlexibleTime парсит строку времени в нескольких форматах
func ParseFlexibleTime(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05-07:00",
		"2006-01-02 15:04:05+00",
		"2006-01-02",
	}

	var parseErr error
	for _, format := range formats {
		t, err := time.Parse(format, s)
		if err == nil {
			return t, nil
		}
		parseErr = err
	}

	return time.Time{}, fmt.Errorf("unable to parse time %q: %w", s, parseErr)
}
