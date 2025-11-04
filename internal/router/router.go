package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/space/backend/internal/handler"
	"github.com/space/backend/internal/middleware"
	"github.com/space/backend/internal/service"
)

// SetupRouter configures all routes for the application
func SetupRouter(
	botToken string,
	botAPIToken string,
	allowedChatID int64,
	allowedOrigins []string,
	environment string,
	authDateTTLMiniApp int64,
	authDateTTLLoginWidget int64,
	userService *service.UserService,
	roomService *service.RoomService,
	bookingService *service.BookingService,
	notificationService *service.NotificationService,
) *gin.Engine {
	r := gin.Default()

	// Global middleware - безопасность
	// 1. Security Headers - должны быть первыми
	r.Use(middleware.SecurityHeaders())

	// 2. HTTPS Enforcement - перенаправление на HTTPS в production
	r.Use(middleware.HTTPSEnforcement(environment))

	// 3. CORS с ограничением по доменам
	r.Use(middleware.CORS(allowedOrigins))

	// 4. Rate Limiting - 100 запросов в минуту с одного IP
	rateLimiter := middleware.NewRateLimiter(100, 1*time.Minute)
	r.Use(rateLimiter.RateLimit())

	// 5. Логирование подозрительных запросов
	r.Use(middleware.SecurityLogger(allowedOrigins))

	// 6. Проверка Referer (только для защищённых эндпоинтов)
	r.Use(middleware.RefererCheck(allowedOrigins))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "Space Backend API",
		})
	})

	// API group
	api := r.Group("/api")

	// Public routes (no auth required)
	public := api.Group("/public")
	{
		roomHandler := handler.NewRoomHandler(roomService)
		public.GET("/rooms", roomHandler.GetAllRooms)
		public.GET("/rooms/:id", roomHandler.GetRoom)
	}

	// Protected routes (require Telegram auth and group membership)
	protected := api.Group("")
	protected.Use(middleware.TelegramAuthMiddleware(botToken, userService, authDateTTLMiniApp, authDateTTLLoginWidget))
	protected.Use(middleware.RequireChatMembership(botToken, allowedChatID, environment))
	{
		// User routes
		userHandler := handler.NewUserHandler(userService)
		users := protected.Group("/users")
		{
			users.GET("/me", userHandler.GetProfile)
			users.PATCH("/me", userHandler.UpdateProfile)
			users.POST("/me/sync-telegram", userHandler.SyncFromTelegram) // Синхронизация данных из Telegram
			users.GET("/phonebook", userHandler.GetPhonebook)
		}

		// Room routes
		roomHandler := handler.NewRoomHandler(roomService)
		rooms := protected.Group("/rooms")
		{
			rooms.GET("", roomHandler.GetAllRooms)
			rooms.GET("/:id", roomHandler.GetRoom)
			rooms.GET("/:id/equipment", roomHandler.GetRoomEquipment)
		}

		// Booking routes
		bookingHandler := handler.NewBookingHandler(bookingService)
		bookings := protected.Group("/bookings")
		{
			bookings.POST("", bookingHandler.CreateBooking)
			bookings.GET("/my", bookingHandler.GetUserBookings)
			bookings.GET("/calendar", bookingHandler.GetCalendarEvents)
			bookings.GET("/:id", bookingHandler.GetBooking)
			bookings.PATCH("/:id", bookingHandler.UpdateBooking)
			bookings.DELETE("/:id", bookingHandler.CancelBooking)
			bookings.POST("/:id/join", bookingHandler.JoinBooking)
			bookings.POST("/:id/leave", bookingHandler.LeaveBooking)
		}
	}

	// Bot API routes (require bot authentication)
	botAPI := api.Group("/bot")
	botAPI.Use(middleware.BotAuthMiddleware(botAPIToken, botToken, allowedChatID, environment, userService))
	{
		botHandler := handler.NewBotHandler(bookingService, notificationService)

		// Booking endpoints for bot
		botAPI.POST("/bookings", botHandler.CreateBooking)
		botAPI.GET("/bookings/user/:telegram_id", botHandler.GetUserBookings)
		botAPI.GET("/rooms/:id/bookings", botHandler.GetRoomBookings)

		// Notification subscription endpoints
		botAPI.POST("/notifications/subscribe", botHandler.Subscribe)
		botAPI.POST("/notifications/unsubscribe", botHandler.Unsubscribe)
		botAPI.GET("/notifications/subscriptions", botHandler.GetSubscriptions)

		roomBotHandler := handler.NewRoomHandler(roomService)
		rooms := botAPI.Group("/rooms")
		{
			rooms.GET("", roomBotHandler.GetAllRooms)
			rooms.GET("/:id", roomBotHandler.GetRoom)
		}
	}

	return r
}
