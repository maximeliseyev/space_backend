package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/space/backend/internal/models"
	"github.com/space/backend/internal/repository"
	"gorm.io/gorm"
)

var (
	ErrBookingConflict = errors.New("booking conflict: room is already booked for this time")
	ErrInvalidTime     = errors.New("invalid time: end time must be after start time")
	ErrPastBooking     = errors.New("cannot create booking in the past")
	ErrRoomNotFound    = errors.New("room not found")
	ErrNotAuthorized   = errors.New("not authorized to perform this action")
)

// BookingConflictError represents a conflict error with details about conflicting bookings
type BookingConflictError struct {
	Message            string            `json:"message"`
	ConflictingBookings []models.Booking `json:"conflicting_bookings"`
}

func (e *BookingConflictError) Error() string {
	return e.Message
}

// BookingService handles booking business logic
type BookingService struct {
	bookingRepo         *repository.BookingRepository
	roomRepo            *repository.RoomRepository
	userRepo            *repository.UserRepository
	notificationService *NotificationService
}

// NewBookingService creates a new booking service
func NewBookingService(
	bookingRepo *repository.BookingRepository,
	roomRepo *repository.RoomRepository,
	userRepo *repository.UserRepository,
	notificationService *NotificationService,
) *BookingService {
	return &BookingService{
		bookingRepo:         bookingRepo,
		roomRepo:            roomRepo,
		userRepo:            userRepo,
		notificationService: notificationService,
	}
}

// CreateBookingRequest represents a request to create a booking
type CreateBookingRequest struct {
	RoomID                uint      `json:"room_id" binding:"required"`
	StartTime             time.Time `json:"start_time" binding:"required"`
	EndTime               time.Time `json:"end_time" binding:"required"`
	Title                 string    `json:"title" binding:"required"`
	Description           string    `json:"description"`
	EstimatedParticipants int       `json:"estimated_participants"`
	IsJoinable            bool      `json:"is_joinable"`
	ParticipantIDs        []uint    `json:"participant_ids"`
}

// CreateBooking creates a new booking with validation
func (s *BookingService) CreateBooking(creatorID uint, req CreateBookingRequest) (*models.Booking, error) {
	// Валидация времени
	if !req.EndTime.After(req.StartTime) {
		return nil, ErrInvalidTime
	}

	// Проверка что бронирование не в прошлом
	if req.StartTime.Before(time.Now()) {
		return nil, ErrPastBooking
	}

	// Проверка существования комнаты
	room, err := s.roomRepo.GetByID(req.RoomID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrRoomNotFound
		}
		return nil, err
	}

	if !room.IsActive {
		return nil, errors.New("room is not active")
	}

	// Проверка на конфликты
	conflictingBookings, err := s.bookingRepo.GetConflictingBookings(req.RoomID, req.StartTime, req.EndTime, nil)
	if err != nil {
		return nil, err
	}
	if len(conflictingBookings) > 0 {
		return nil, &BookingConflictError{
			Message:            "booking conflict: room is already booked for this time",
			ConflictingBookings: conflictingBookings,
		}
	}

	// Получаем участников если они указаны
	var participants []models.User
	if len(req.ParticipantIDs) > 0 {
		participants, err = s.userRepo.GetByIDs(req.ParticipantIDs)
		if err != nil {
			return nil, err
		}
	}

	// Создаем бронирование
	booking := &models.Booking{
		RoomID:                req.RoomID,
		CreatorID:             creatorID,
		StartTime:             req.StartTime,
		EndTime:               req.EndTime,
		Title:                 req.Title,
		Description:           req.Description,
		EstimatedParticipants: req.EstimatedParticipants,
		IsJoinable:            req.IsJoinable,
		Status:                models.BookingStatusConfirmed,
		Participants:          participants,
	}

	err = s.bookingRepo.Create(booking)
	if err != nil {
		return nil, err
	}

	// Загружаем полную информацию о бронировании
	fullBooking, err := s.bookingRepo.GetByID(booking.ID)
	if err != nil {
		return nil, err
	}

	// Отправляем уведомление боту о новом бронировании (асинхронно, не блокируя создание)
	if s.notificationService != nil {
		go func() {
			if err := s.notificationService.NotifyBookingCreated(fullBooking); err != nil {
				// Логируем ошибку, но не прерываем процесс создания бронирования
				fmt.Printf("Failed to send booking notification: %v\n", err)
			}
		}()
	}

	return fullBooking, nil
}

// GetBooking gets a booking by ID
func (s *BookingService) GetBooking(id uint) (*models.Booking, error) {
	return s.bookingRepo.GetByID(id)
}

// GetUserBookings gets all bookings for a user
func (s *BookingService) GetUserBookings(userID uint) ([]models.Booking, error) {
	return s.bookingRepo.GetByUserID(userID)
}

// GetUserBookingsByTelegramID gets all bookings for a user by Telegram ID
func (s *BookingService) GetUserBookingsByTelegramID(telegramID int64) ([]models.Booking, error) {
	user, err := s.userRepo.GetByTelegramID(telegramID)
	if err != nil {
		return nil, err
	}
	return s.bookingRepo.GetByUserID(user.ID)
}

// CreateSimpleBooking creates a new booking (simplified version for bot API)
func (s *BookingService) CreateSimpleBooking(
	roomID uint,
	creatorID uint,
	startTime time.Time,
	endTime time.Time,
	title string,
	description string,
	estimatedParticipants int,
	isJoinable bool,
) (*models.Booking, error) {
	req := CreateBookingRequest{
		RoomID:                roomID,
		StartTime:             startTime,
		EndTime:               endTime,
		Title:                 title,
		Description:           description,
		EstimatedParticipants: estimatedParticipants,
		IsJoinable:            isJoinable,
	}
	return s.CreateBooking(creatorID, req)
}

// GetUpcomingBookings gets upcoming bookings
func (s *BookingService) GetUpcomingBookings(limit int) ([]models.Booking, error) {
	return s.bookingRepo.GetUpcoming(limit)
}

// GetCalendarEvents gets bookings for calendar view
func (s *BookingService) GetCalendarEvents(start, end time.Time) ([]models.Booking, error) {
	return s.bookingRepo.GetForCalendar(start, end)
}

// CancelBooking cancels a booking (creator or admin can cancel)
func (s *BookingService) CancelBooking(bookingID, userID uint) error {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return err
	}

	// Получаем пользователя для проверки прав
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	// Проверка прав доступа: создатель или админ
	if booking.CreatorID != userID && !user.IsAdmin() {
		return ErrNotAuthorized
	}

	return s.bookingRepo.Cancel(bookingID)
}

// JoinBooking allows a user to join a joinable booking
func (s *BookingService) JoinBooking(bookingID, userID uint) error {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return err
	}

	if !booking.IsJoinable {
		return errors.New("this booking is not joinable")
	}

	if booking.Status != models.BookingStatusConfirmed {
		return errors.New("cannot join cancelled or completed booking")
	}

	return s.bookingRepo.AddParticipant(bookingID, userID)
}

// LeaveBooking allows a participant to leave a booking
func (s *BookingService) LeaveBooking(bookingID, userID uint) error {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return err
	}

	// Создатель не может покинуть бронирование, только отменить
	if booking.CreatorID == userID {
		return errors.New("creator cannot leave booking, use cancel instead")
	}

	return s.bookingRepo.RemoveParticipant(bookingID, userID)
}

// CheckAvailability checks if a room is available for a time period
func (s *BookingService) CheckAvailability(roomID uint, start, end time.Time) (bool, error) {
	hasConflict, err := s.bookingRepo.CheckConflict(roomID, start, end, nil)
	if err != nil {
		return false, err
	}
	return !hasConflict, nil
}

// GetRoomBookings gets all bookings for a specific room in a time range
func (s *BookingService) GetRoomBookings(roomID uint, start, end time.Time) ([]models.Booking, error) {
	return s.bookingRepo.GetByRoomAndTimeRange(roomID, start, end)
}

// UpdateBookingRequest represents a request to update a booking
type UpdateBookingRequest struct {
	StartTime             *time.Time `json:"start_time"`
	EndTime               *time.Time `json:"end_time"`
	Title                 *string    `json:"title"`
	Description           *string    `json:"description"`
	EstimatedParticipants *int       `json:"estimated_participants"`
	IsJoinable            *bool      `json:"is_joinable"`
}

// UpdateBooking updates a booking (creator or admin can update)
func (s *BookingService) UpdateBooking(bookingID, userID uint, req UpdateBookingRequest) (*models.Booking, error) {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, err
	}

	// Получаем пользователя для проверки прав
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	// Проверка прав доступа: создатель или админ
	if booking.CreatorID != userID && !user.IsAdmin() {
		return nil, ErrNotAuthorized
	}

	// Обновляем поля
	if req.StartTime != nil {
		booking.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		booking.EndTime = *req.EndTime
	}
	if req.Title != nil {
		booking.Title = *req.Title
	}
	if req.Description != nil {
		booking.Description = *req.Description
	}
	if req.EstimatedParticipants != nil {
		booking.EstimatedParticipants = *req.EstimatedParticipants
	}
	if req.IsJoinable != nil {
		booking.IsJoinable = *req.IsJoinable
	}

	// Валидация времени
	if !booking.EndTime.After(booking.StartTime) {
		return nil, ErrInvalidTime
	}

	// Проверка на конфликты (исключая текущее бронирование)
	conflictingBookings, err := s.bookingRepo.GetConflictingBookings(booking.RoomID, booking.StartTime, booking.EndTime, &bookingID)
	if err != nil {
		return nil, err
	}
	if len(conflictingBookings) > 0 {
		return nil, &BookingConflictError{
			Message:            "booking conflict: room is already booked for this time",
			ConflictingBookings: conflictingBookings,
		}
	}

	err = s.bookingRepo.Update(booking)
	if err != nil {
		return nil, err
	}

	return s.bookingRepo.GetByID(bookingID)
}

// FormatBookingForCalendar formats booking for FullCalendar
func FormatBookingForCalendar(booking *models.Booking) map[string]interface{} {
	// Формируем информацию о создателе
	creatorInfo := map[string]interface{}{
		"id":         booking.Creator.ID,
		"first_name": booking.Creator.FirstName,
		"last_name":  booking.Creator.LastName,
		"username":   booking.Creator.Username,
	}

	return map[string]interface{}{
		"id":         fmt.Sprintf("%d", booking.ID),
		"title":      booking.Title,
		"start":      booking.StartTime.Format(time.RFC3339),
		"end":        booking.EndTime.Format(time.RFC3339),
		"resourceId": fmt.Sprintf("%d", booking.RoomID),
		"calendarId": fmt.Sprintf("room_%d", booking.RoomID),
		"creator":    creatorInfo, // Полная информация о создателе
		"extendedProps": map[string]interface{}{
			"roomName":     booking.Room.Name,
			"description":  booking.Description,
			"participants": booking.Participants,
			"allow_join":   booking.IsJoinable,
			"status":       booking.Status,
		},
	}
}
