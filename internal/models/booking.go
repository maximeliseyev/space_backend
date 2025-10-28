package models

import (
	"time"

	"gorm.io/gorm"
)

// BookingStatus определяет статус бронирования
type BookingStatus string

const (
	BookingStatusConfirmed BookingStatus = "confirmed" // Подтверждено
	BookingStatusCancelled BookingStatus = "cancelled" // Отменено
	BookingStatusCompleted BookingStatus = "completed" // Завершено
)

// Booking represents a room booking
type Booking struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	RoomID    uint   `gorm:"not null;index" json:"room_id"`
	CreatorID uint   `gorm:"not null;index" json:"creator_id"` // Кто создал бронирование

	// Обязательные параметры
	StartTime time.Time `gorm:"not null;index" json:"start_time"` // Время начала
	EndTime   time.Time `gorm:"not null;index" json:"end_time"`   // Время окончания

	// Информация о мероприятии
	Title       string `gorm:"not null" json:"title"`                       // Название мероприятия
	Description string `gorm:"type:text" json:"description,omitempty"`      // Описание

	// Дополнительные параметры
	EstimatedParticipants int  `gorm:"default:1" json:"estimated_participants"` // Предполагаемое количество участников
	IsJoinable           bool `gorm:"default:false" json:"is_joinable"`        // Можно ли присоединиться к мероприятию

	Status BookingStatus `gorm:"type:varchar(20);default:'confirmed'" json:"status"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Связи
	Room         Room   `gorm:"foreignKey:RoomID" json:"room,omitempty"`
	Creator      User   `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
	Participants []User `gorm:"many2many:booking_participants;" json:"participants,omitempty"` // Другие участники
}

// BeforeCreate hook для валидации
func (b *Booking) BeforeCreate(tx *gorm.DB) error {
	// Проверка что время окончания позже времени начала
	if !b.EndTime.After(b.StartTime) {
		return gorm.ErrInvalidData
	}
	return nil
}

// BeforeUpdate hook для валидации
func (b *Booking) BeforeUpdate(tx *gorm.DB) error {
	// Проверка что время окончания позже времени начала
	if !b.EndTime.After(b.StartTime) {
		return gorm.ErrInvalidData
	}
	return nil
}
