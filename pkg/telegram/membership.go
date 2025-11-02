package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ChatMemberResponse represents the Telegram API response for getChatMember
type ChatMemberResponse struct {
	OK     bool   `json:"ok"`
	Result struct {
		Status string `json:"status"`
		User   struct {
			ID        int64  `json:"id"`
			FirstName string `json:"first_name"`
			Username  string `json:"username"`
		} `json:"user"`
	} `json:"result"`
}

// CheckUserInChat проверяет, является ли пользователь участником чата
func CheckUserInChat(userID int64, chatID int64, botToken string) (bool, error) {
	// Создаем HTTP клиент с таймаутом
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	url := fmt.Sprintf(
		"https://api.telegram.org/bot%s/getChatMember?chat_id=%d&user_id=%d",
		botToken,
		chatID,
		userID,
	)

	// Создаем запрос с контекстом
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to check membership: %w", err)
	}
	defer resp.Body.Close()

	var result ChatMemberResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.OK {
		return false, nil
	}

	// Разрешенные статусы
	allowedStatuses := map[string]bool{
		"creator":       true,
		"administrator": true,
		"member":        true,
	}

	// Запрещенные: left, kicked, restricted
	return allowedStatuses[result.Result.Status], nil
}
