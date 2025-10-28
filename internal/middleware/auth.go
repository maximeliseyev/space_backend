package middleware

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/space/backend/internal/service"
	"github.com/space/backend/pkg/response"
	"github.com/space/backend/pkg/telegram"
)

var (
	ErrMissingAuthHeader = errors.New("missing authorization header")
	ErrInvalidAuthHeader = errors.New("invalid authorization header")
)

// TelegramAuthMiddleware validates Telegram Mini App authentication
func TelegramAuthMiddleware(botToken string, userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем initData из заголовка
		initData := c.GetHeader("X-Telegram-Init-Data")
		if initData == "" {
			response.Unauthorized(c, ErrMissingAuthHeader)
			c.Abort()
			return
		}

		// Валидируем initData
		err := telegram.ValidateInitData(initData, botToken)
		if err != nil {
			response.Unauthorized(c, err)
			c.Abort()
			return
		}

		// Извлекаем Telegram ID из query параметров initData
		// В реальном initData это будет JSON объект user
		// Для упрощения можем передавать telegram_id отдельно
		telegramIDStr := c.GetHeader("X-Telegram-User-ID")
		if telegramIDStr == "" {
			response.Unauthorized(c, errors.New("missing telegram user ID"))
			c.Abort()
			return
		}

		telegramID, err := strconv.ParseInt(telegramIDStr, 10, 64)
		if err != nil {
			response.Unauthorized(c, errors.New("invalid telegram user ID"))
			c.Abort()
			return
		}

		username := c.GetHeader("X-Telegram-Username")

		// Получаем или создаем пользователя
		user, err := userService.SyncTelegramUser(telegramID, username)
		if err != nil {
			response.InternalServerError(c, err)
			c.Abort()
			return
		}

		// Сохраняем пользователя в контекст
		c.Set("userID", user.ID)
		c.Set("user", user)

		c.Next()
	}
}

// CORS middleware
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Telegram-Init-Data, X-Telegram-User-ID, X-Telegram-Username")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
