package telegram

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// UserProfilePhotos represents user profile photos response from Telegram API
type UserProfilePhotos struct {
	Ok     bool `json:"ok"`
	Result struct {
		TotalCount int           `json:"total_count"`
		Photos     [][]PhotoSize `json:"photos"`
	} `json:"result"`
}

// PhotoSize represents a photo size
type PhotoSize struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	FileSize     int    `json:"file_size,omitempty"`
}

// GetFileResponse represents get file response from Telegram API
type GetFileResponse struct {
	Ok     bool `json:"ok"`
	Result struct {
		FileID       string `json:"file_id"`
		FileUniqueID string `json:"file_unique_id"`
		FileSize     int    `json:"file_size"`
		FilePath     string `json:"file_path"`
	} `json:"result"`
}

// GetUserProfilePhotoURL получает URL последней фотографии профиля пользователя из Telegram
// Возвращает URL или пустую строку если фото нет
func GetUserProfilePhotoURL(telegramUserID int64, botToken string) (string, error) {
	// Получаем список фотографий профиля (limit=1 для получения только последней)
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getUserProfilePhotos?user_id=%d&limit=1", botToken, telegramUserID)

	resp, err := http.Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("failed to get user profile photos: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var photos UserProfilePhotos
	if err := json.Unmarshal(body, &photos); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if !photos.Ok {
		return "", fmt.Errorf("telegram API returned ok=false")
	}

	// Если фотографий нет, возвращаем пустую строку
	if photos.Result.TotalCount == 0 || len(photos.Result.Photos) == 0 {
		log.Printf("DEBUG: User %d has no profile photos", telegramUserID)
		return "", nil
	}

	// Получаем последнюю фотографию (первая в массиве)
	// В каждой фотографии есть несколько размеров, берем самый большой (последний в массиве)
	photoSizes := photos.Result.Photos[0]
	if len(photoSizes) == 0 {
		return "", nil
	}

	// Берем самый большой размер (последний элемент)
	largestPhoto := photoSizes[len(photoSizes)-1]

	// Получаем file_path для построения URL
	fileInfo, err := getFileInfo(largestPhoto.FileID, botToken)
	if err != nil {
		return "", fmt.Errorf("failed to get file info: %w", err)
	}

	// Строим публичный URL для фотографии
	photoURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", botToken, fileInfo.Result.FilePath)

	log.Printf("DEBUG: Got profile photo URL for user %d: %s", telegramUserID, photoURL)
	return photoURL, nil
}

// getFileInfo получает информацию о файле по file_id
func getFileInfo(fileID, botToken string) (*GetFileResponse, error) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getFile?file_id=%s", botToken, fileID)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var fileInfo GetFileResponse
	if err := json.Unmarshal(body, &fileInfo); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !fileInfo.Ok {
		return nil, fmt.Errorf("telegram API returned ok=false")
	}

	return &fileInfo, nil
}
