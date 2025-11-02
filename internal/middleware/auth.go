package middleware

import (
	"errors"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/space/backend/internal/models"
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

// RequireChatMembership проверяет, что пользователь является участником разрешенной группы
func RequireChatMembership(botToken string, allowedChatID int64, environment string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// В development режиме без настроенного ALLOWED_CHAT_ID пропускаем проверку
		if allowedChatID == 0 {
			if environment != "production" {
				log.Println("WARNING: ALLOWED_CHAT_ID not set, skipping membership check in development mode")
				c.Next()
				return
			}
			// В production требуем настройку
			response.InternalServerError(c, errors.New("server configuration error: ALLOWED_CHAT_ID not configured"))
			c.Abort()
			return
		}

		// Получаем user из контекста (должен быть установлен предыдущим middleware)
		userInterface, exists := c.Get("user")
		if !exists {
			response.Unauthorized(c, errors.New("user not authenticated"))
			c.Abort()
			return
		}

		// Приводим к типу модели User
		userModel, ok := userInterface.(*models.User)
		if !ok {
			response.InternalServerError(c, errors.New("invalid user data type"))
			c.Abort()
			return
		}

		telegramUserID := userModel.TelegramID

		// Проверяем кэш сначала (TTL 5 минут)
		if isMember, cached := telegram.GlobalCache.Get(telegramUserID); cached {
			if !isMember {
				log.Printf("INFO: User %d denied access - not a group member (cached)", telegramUserID)
				response.Forbidden(c, errors.New("access denied. You must be a member of the authorized group"))
				c.Abort()
				return
			}
			// Пользователь в кэше и является участником
			c.Next()
			return
		}

		// Проверяем членство через API
		isMember, err := telegram.CheckUserInChat(telegramUserID, allowedChatID, botToken)
		if err != nil {
			log.Printf("ERROR: Failed to check membership for user %d: %v", telegramUserID, err)

			// В production блокируем при ошибке проверки
			if environment == "production" {
				response.InternalServerError(c, errors.New("failed to verify membership"))
				c.Abort()
				return
			}
			// В dev разрешаем
			log.Println("WARNING: Membership check failed in development mode, allowing access")
			c.Next()
			return
		}

		// Сохраняем результат в кэш (5 минут)
		telegram.GlobalCache.Set(telegramUserID, isMember, 5*time.Minute)

		if !isMember {
			log.Printf("INFO: User %d denied access - not a group member", telegramUserID)
			response.Forbidden(c, errors.New("access denied. You must be a member of the authorized group"))
			c.Abort()
			return
		}

		log.Printf("INFO: User %d authorized - group member", telegramUserID)
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
