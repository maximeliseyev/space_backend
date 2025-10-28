package repository

import (
	"time"

	"github.com/space/backend/internal/models"
	"gorm.io/gorm"
)

// BookingRepository handles database operations for bookings
type BookingRepository struct {
	db *gorm.DB
}

// NewBookingRepository creates a new booking repository
func NewBookingRepository(db *gorm.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

// Create creates a new booking
func (r *BookingRepository) Create(booking *models.Booking) error {
	return r.db.Create(booking).Error
}

// GetByID gets a booking by ID with all relations
func (r *BookingRepository) GetByID(id uint) (*models.Booking, error) {
	var booking models.Booking
	err := r.db.Preload("Room").
		Preload("Creator").
		Preload("Participants").
		First(&booking, id).Error
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

// GetByUserID gets all bookings for a user (created or participating)
func (r *BookingRepository) GetByUserID(userID uint) ([]models.Booking, error) {
	var bookings []models.Booking

	// Получаем бронирования где пользователь - создатель или участник
	err := r.db.Preload("Room").
		Preload("Creator").
		Preload("Participants").
		Where("creator_id = ?", userID).
		Or("id IN (SELECT booking_id FROM booking_participants WHERE user_id = ?)", userID).
		Order("start_time DESC").
		Find(&bookings).Error

	return bookings, err
}

// GetByRoomAndTimeRange gets bookings for a room in a time range
func (r *BookingRepository) GetByRoomAndTimeRange(roomID uint, start, end time.Time) ([]models.Booking, error) {
	var bookings []models.Booking
	err := r.db.Preload("Room").
		Preload("Creator").
		Preload("Participants").
		Where("room_id = ? AND status != ? AND start_time < ? AND end_time > ?",
			roomID, models.BookingStatusCancelled, end, start).
		Order("start_time").
		Find(&bookings).Error
	return bookings, err
}

// CheckConflict checks if there's a booking conflict
func (r *BookingRepository) CheckConflict(roomID uint, start, end time.Time, excludeBookingID *uint) (bool, error) {
	var count int64
	query := r.db.Model(&models.Booking{}).
		Where("room_id = ? AND status != ? AND start_time < ? AND end_time > ?",
			roomID, models.BookingStatusCancelled, end, start)

	// Исключаем конкретное бронирование (для обновления)
	if excludeBookingID != nil {
		query = query.Where("id != ?", *excludeBookingID)
	}

	err := query.Count(&count).Error
	return count > 0, err
}

// GetUpcoming gets upcoming bookings
func (r *BookingRepository) GetUpcoming(limit int) ([]models.Booking, error) {
	var bookings []models.Booking
	now := time.Now()

	err := r.db.Preload("Room").
		Preload("Creator").
		Preload("Participants").
		Where("start_time > ? AND status = ?", now, models.BookingStatusConfirmed).
		Order("start_time ASC").
		Limit(limit).
		Find(&bookings).Error

	return bookings, err
}

// GetForCalendar gets all bookings in a time range for calendar view
func (r *BookingRepository) GetForCalendar(start, end time.Time) ([]models.Booking, error) {
	var bookings []models.Booking
	err := r.db.Preload("Room").
		Preload("Creator").
		Preload("Participants").
		Where("status != ? AND start_time < ? AND end_time > ?",
			models.BookingStatusCancelled, end, start).
		Order("start_time").
		Find(&bookings).Error
	return bookings, err
}

// Update updates a booking
func (r *BookingRepository) Update(booking *models.Booking) error {
	return r.db.Save(booking).Error
}

// Delete soft deletes a booking
func (r *BookingRepository) Delete(id uint) error {
	return r.db.Delete(&models.Booking{}, id).Error
}

// Cancel cancels a booking (sets status to cancelled)
func (r *BookingRepository) Cancel(id uint) error {
	return r.db.Model(&models.Booking{}).
		Where("id = ?", id).
		Update("status", models.BookingStatusCancelled).Error
}

// AddParticipant adds a participant to a booking
func (r *BookingRepository) AddParticipant(bookingID, userID uint) error {
	return r.db.Exec(
		"INSERT INTO booking_participants (booking_id, user_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
		bookingID, userID,
	).Error
}

// RemoveParticipant removes a participant from a booking
func (r *BookingRepository) RemoveParticipant(bookingID, userID uint) error {
	return r.db.Exec(
		"DELETE FROM booking_participants WHERE booking_id = ? AND user_id = ?",
		bookingID, userID,
	).Error
}
