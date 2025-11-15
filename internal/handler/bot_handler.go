package handler

import (
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/space/backend/internal/models"
	"github.com/space/backend/internal/service"
	"github.com/space/backend/pkg/response"
	"github.com/space/backend/pkg/utils"
)

type BotHandler struct {
	bookingService      *service.BookingService
	notificationService *service.NotificationService
}

func NewBotHandler(bookingService *service.BookingService, notificationService *service.NotificationService) *BotHandler {
	return &BotHandler{
		bookingService:      bookingService,
		notificationService: notificationService,
	}
}

// CreateBooking creates a booking on behalf of a user
// POST /api/bot/bookings
func (h *BotHandler) CreateBooking(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		response.Unauthorized(c, service.ErrNotAuthorized)
		return
	}
	user := userInterface.(*models.User)

	var req struct {
		RoomID                uint      `json:"room_id" binding:"required"`
		StartTime             time.Time `json:"start_time" binding:"required"`
		EndTime               time.Time `json:"end_time" binding:"required"`
		Title                 string    `json:"title" binding:"required"`
		Description           string    `json:"description"`
		EstimatedParticipants int       `json:"estimated_participants"`
		IsJoinable            bool      `json:"is_joinable"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err)
		return
	}

	booking, err := h.bookingService.CreateSimpleBooking(
		req.RoomID,
		user.ID,
		req.StartTime,
		req.EndTime,
		req.Title,
		req.Description,
		req.EstimatedParticipants,
		req.IsJoinable,
	)

	if err != nil {
		log.Printf("ERROR: Bot failed to create booking: %v", err)
		response.InternalServerError(c, err)
		return
	}

	log.Printf("INFO: Bot created booking ID %d for user %d (TelegramID: %d)", booking.ID, user.ID, user.TelegramID)

	// Получаем подписчиков для уведомлений
	subscribers, err := h.notificationService.GetRoomSubscribers(req.RoomID)
	if err != nil {
		log.Printf("WARNING: Failed to get subscribers for room %d: %v", req.RoomID, err)
	} else {
		log.Printf("INFO: Found %d subscribers for room %d", len(subscribers), req.RoomID)
	}

	response.Created(c, gin.H{
		"booking":     booking,
		"subscribers": subscribers, // Бот использует это для отправки уведомлений
	})
}

// Subscribe subscribes a user to room notifications
// POST /api/bot/notifications/subscribe
func (h *BotHandler) Subscribe(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		response.Unauthorized(c, service.ErrNotAuthorized)
		return
	}
	user := userInterface.(*models.User)

	var req struct {
		RoomID uint `json:"room_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err)
		return
	}

	err := h.notificationService.Subscribe(user.ID, req.RoomID)
	if err != nil {
		log.Printf("ERROR: Bot failed to subscribe user %d to room %d: %v", user.ID, req.RoomID, err)
		response.InternalServerError(c, err)
		return
	}

	log.Printf("INFO: User %d (TelegramID: %d) subscribed to room %d", user.ID, user.TelegramID, req.RoomID)
	response.Success(c, gin.H{"message": "subscribed successfully"})
}

// Unsubscribe unsubscribes a user from room notifications
// POST /api/bot/notifications/unsubscribe
func (h *BotHandler) Unsubscribe(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		response.Unauthorized(c, service.ErrNotAuthorized)
		return
	}
	user := userInterface.(*models.User)

	var req struct {
		RoomID uint `json:"room_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err)
		return
	}

	err := h.notificationService.Unsubscribe(user.ID, req.RoomID)
	if err != nil {
		log.Printf("ERROR: Bot failed to unsubscribe user %d from room %d: %v", user.ID, req.RoomID, err)
		response.InternalServerError(c, err)
		return
	}

	log.Printf("INFO: User %d (TelegramID: %d) unsubscribed from room %d", user.ID, user.TelegramID, req.RoomID)
	response.Success(c, gin.H{"message": "unsubscribed successfully"})
}

// GetSubscriptions returns all rooms a user is subscribed to
// GET /api/bot/notifications/subscriptions
func (h *BotHandler) GetSubscriptions(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		response.Unauthorized(c, service.ErrNotAuthorized)
		return
	}
	user := userInterface.(*models.User)

	subscriptions, err := h.notificationService.GetUserSubscriptions(user.ID)
	if err != nil {
		log.Printf("ERROR: Bot failed to get subscriptions for user %d: %v", user.ID, err)
		response.InternalServerError(c, err)
		return
	}

	response.Success(c, subscriptions)
}

// GetUserBookings returns bookings for a specific user
// GET /api/bot/bookings/user/:telegram_id
func (h *BotHandler) GetUserBookings(c *gin.Context) {
	telegramIDStr := c.Param("telegram_id")
	telegramID, err := strconv.ParseInt(telegramIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	// Здесь можно было бы добавить проверку прав доступа
	// (например, только свои бронирования или админ)

	bookings, err := h.bookingService.GetUserBookingsByTelegramID(telegramID)
	if err != nil {
		log.Printf("ERROR: Bot failed to get bookings for Telegram user %d: %v", telegramID, err)
		response.InternalServerError(c, err)
		return
	}

	response.Success(c, bookings)
}

// GetRoomBookings returns all bookings for a specific room
// GET /api/bot/rooms/:id/bookings
func (h *BotHandler) GetRoomBookings(c *gin.Context) {
	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	// Parse query parameters for date filtering
	// Support both "date" parameter (for single day) and "start"/"end" parameters (for range)
	dateStr := c.Query("date")
	startTimeStr := c.Query("start")
	endTimeStr := c.Query("end")

	var startTime, endTime time.Time

	// If "date" parameter is provided, use it for a single day
	if dateStr != "" {
		t, err := utils.ParseFlexibleTime(dateStr)
		if err != nil {
			response.BadRequest(c, err)
			return
		}
		// Set start to beginning of the day (00:00:00)
		startTime = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		// Set end to end of the day (23:59:59)
		endTime = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
	} else {
		// Use "start" and "end" parameters
		// Default to current time and 30 days ahead if not specified
		if startTimeStr != "" {
			t, err := utils.ParseFlexibleTime(startTimeStr)
			if err != nil {
				response.BadRequest(c, err)
				return
			}
			startTime = t
		} else {
			startTime = time.Now()
		}

		if endTimeStr != "" {
			t, err := utils.ParseFlexibleTime(endTimeStr)
			if err != nil {
				response.BadRequest(c, err)
				return
			}
			endTime = t
		} else {
			endTime = startTime.AddDate(0, 0, 30) // 30 days from start
		}
	}

	bookings, err := h.bookingService.GetRoomBookings(uint(roomID), startTime, endTime)
	if err != nil {
		log.Printf("ERROR: Bot failed to get bookings for room %d: %v", roomID, err)
		response.InternalServerError(c, err)
		return
	}

	response.Success(c, bookings)
}

