package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/space/backend/internal/service"
	"github.com/space/backend/pkg/response"
	"github.com/space/backend/pkg/utils"
)

// BookingHandler handles booking-related HTTP requests
type BookingHandler struct {
	bookingService *service.BookingService
}

// NewBookingHandler creates a new booking handler
func NewBookingHandler(bookingService *service.BookingService) *BookingHandler {
	return &BookingHandler{bookingService: bookingService}
}

// CreateBooking godoc
// @Summary Create a new booking
// @Tags bookings
// @Accept json
// @Produce json
// @Param booking body service.CreateBookingRequest true "Booking data"
// @Success 201 {object} models.Booking
// @Router /api/bookings [post]
func (h *BookingHandler) CreateBooking(c *gin.Context) {
	var req service.CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err)
		return
	}

	// Получаем ID пользователя из контекста (устанавливается middleware)
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, service.ErrNotAuthorized)
		return
	}

	booking, err := h.bookingService.CreateBooking(userID.(uint), req)
	if err != nil {
		// Проверяем, является ли это ошибкой конфликта с деталями
		if conflictErr, ok := err.(*service.BookingConflictError); ok {
			response.ConflictWithData(c, conflictErr.Message, conflictErr.ConflictingBookings)
			return
		}

		switch err {
		case service.ErrBookingConflict:
			response.Conflict(c, err)
		case service.ErrInvalidTime, service.ErrPastBooking:
			response.BadRequest(c, err)
		case service.ErrRoomNotFound:
			response.NotFound(c, err)
		default:
			response.InternalServerError(c, err)
		}
		return
	}

	response.Created(c, booking)
}

// GetBooking godoc
// @Summary Get booking by ID
// @Tags bookings
// @Produce json
// @Param id path int true "Booking ID"
// @Success 200 {object} models.Booking
// @Router /api/bookings/{id} [get]
func (h *BookingHandler) GetBooking(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	booking, err := h.bookingService.GetBooking(uint(id))
	if err != nil {
		response.NotFound(c, err)
		return
	}

	response.Success(c, booking)
}

// GetUserBookings godoc
// @Summary Get current user's bookings
// @Tags bookings
// @Produce json
// @Success 200 {array} models.Booking
// @Router /api/bookings/my [get]
func (h *BookingHandler) GetUserBookings(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, service.ErrNotAuthorized)
		return
	}

	bookings, err := h.bookingService.GetUserBookings(userID.(uint))
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.Success(c, bookings)
}

// GetCalendarEvents godoc
// @Summary Get calendar events
// @Tags bookings
// @Produce json
// @Param start query string true "Start date (RFC3339)"
// @Param end query string true "End date (RFC3339)"
// @Success 200 {array} map[string]interface{}
// @Router /api/bookings/calendar [get]
func (h *BookingHandler) GetCalendarEvents(c *gin.Context) {
	startStr := c.Query("start")
	endStr := c.Query("end")

	if startStr == "" || endStr == "" {
		response.BadRequest(c, service.ErrInvalidTime)
		return
	}

	start, err := utils.ParseFlexibleTime(startStr)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	end, err := utils.ParseFlexibleTime(endStr)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	bookings, err := h.bookingService.GetCalendarEvents(start, end)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	// Форматируем для FullCalendar
	events := make([]map[string]interface{}, len(bookings))
	for i, booking := range bookings {
		events[i] = service.FormatBookingForCalendar(&booking)
	}

	response.Success(c, events)
}

// CancelBooking godoc
// @Summary Cancel a booking
// @Tags bookings
// @Param id path int true "Booking ID"
// @Success 204
// @Router /api/bookings/{id} [delete]
func (h *BookingHandler) CancelBooking(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, service.ErrNotAuthorized)
		return
	}

	err = h.bookingService.CancelBooking(uint(id), userID.(uint))
	if err != nil {
		switch err {
		case service.ErrNotAuthorized:
			response.Forbidden(c, err)
		default:
			response.InternalServerError(c, err)
		}
		return
	}

	response.NoContent(c)
}

// JoinBooking godoc
// @Summary Join a booking
// @Tags bookings
// @Param id path int true "Booking ID"
// @Success 200
// @Router /api/bookings/{id}/join [post]
func (h *BookingHandler) JoinBooking(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, service.ErrNotAuthorized)
		return
	}

	err = h.bookingService.JoinBooking(uint(id), userID.(uint))
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	response.SuccessWithMessage(c, nil, "Successfully joined booking")
}

// LeaveBooking godoc
// @Summary Leave a booking
// @Tags bookings
// @Param id path int true "Booking ID"
// @Success 200
// @Router /api/bookings/{id}/leave [post]
func (h *BookingHandler) LeaveBooking(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, service.ErrNotAuthorized)
		return
	}

	err = h.bookingService.LeaveBooking(uint(id), userID.(uint))
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	response.SuccessWithMessage(c, nil, "Successfully left booking")
}

// UpdateBooking godoc
// @Summary Update a booking
// @Tags bookings
// @Accept json
// @Produce json
// @Param id path int true "Booking ID"
// @Param booking body service.UpdateBookingRequest true "Booking data"
// @Success 200 {object} models.Booking
// @Router /api/bookings/{id} [patch]
func (h *BookingHandler) UpdateBooking(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	var req service.UpdateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err)
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, service.ErrNotAuthorized)
		return
	}

	booking, err := h.bookingService.UpdateBooking(uint(id), userID.(uint), req)
	if err != nil {
		// Проверяем, является ли это ошибкой конфликта с деталями
		if conflictErr, ok := err.(*service.BookingConflictError); ok {
			response.ConflictWithData(c, conflictErr.Message, conflictErr.ConflictingBookings)
			return
		}

		switch err {
		case service.ErrNotAuthorized:
			response.Forbidden(c, err)
		case service.ErrBookingConflict:
			response.Conflict(c, err)
		case service.ErrInvalidTime:
			response.BadRequest(c, err)
		default:
			response.InternalServerError(c, err)
		}
		return
	}

	response.Success(c, booking)
}
