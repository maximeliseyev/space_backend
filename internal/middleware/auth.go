package middleware

import (
	"errors"

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

		// Development mode - пропускаем валидацию
		if initData == "dev_mode" {
			// Создаем тестового пользователя для разработки
			user, err := userService.SyncTelegramUser(12345, "devuser", "Dev", "User", "en")
			if err != nil {
				response.InternalServerError(c, err)
				c.Abort()
				return
			}

			// Сохраняем пользователя в контекст
			c.Set("userID", user.ID)
			c.Set("user", user)
			c.Next()
			return
		}

		// Production mode - валидируем и парсим initData
		telegramUser, err := telegram.ValidateAndParseInitData(initData, botToken)
		if err != nil {
			response.Unauthorized(c, err)
			c.Abort()
			return
		}

		// Получаем или создаем пользователя с полными данными из Telegram
		user, err := userService.SyncTelegramUser(
			telegramUser.ID,
			telegramUser.Username,
			telegramUser.FirstName,
			telegramUser.LastName,
			telegramUser.LanguageCode,
		)
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

// CORS middleware with security restrictions
// allowedOrigins: список разрешённых доменов (из конфигурации)
func CORS(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Проверяем, разрешён ли origin
		isAllowed := false
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				isAllowed = true
				break
			}
		}

		// Устанавливаем заголовки только для разрешённых origins
		if isAllowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Telegram-Init-Data, X-Telegram-User-ID, X-Telegram-Username")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")
		c.Writer.Header().Set("Access-Control-Max-Age", "43200") // 12 hours

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
