package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	TelegramID  int64          `gorm:"uniqueIndex;not null" json:"telegram_id"`
	Username    string         `gorm:"index" json:"username"`
	FirstName   string         `json:"first_name,omitempty"`
	LastName    string         `json:"last_name,omitempty"`
	PhoneNumber string         `gorm:"index" json:"phone_number,omitempty"`

	// Телефонная книга - пользователь показывается только если заполнены имя/фамилия и телефон
	IsInPhoneBook bool `gorm:"default:false" json:"is_in_phonebook"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Связи
	Bookings      []Booking      `gorm:"foreignKey:CreatorID" json:"bookings,omitempty"`
	ParticipatedBookings []Booking `gorm:"many2many:booking_participants;" json:"-"`
}

// BeforeSave hook для автоматической установки флага IsInPhoneBook
func (u *User) BeforeSave(tx *gorm.DB) error {
	// Пользователь попадает в телефонную книгу только если указал ФИО и телефон
	if u.FirstName != "" && u.LastName != "" && u.PhoneNumber != "" {
		u.IsInPhoneBook = true
	} else {
		u.IsInPhoneBook = false
	}
	return nil
}
