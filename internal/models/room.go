package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Room represents a bookable room in the coworking space
type Room struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"uniqueIndex;not null" json:"name"` // Название комнаты
	Description string `gorm:"type:text" json:"description"`     // Описание
	Capacity    int    `gorm:"default:1" json:"capacity"`        // Вместимость
	IsActive    bool   `gorm:"default:true" json:"is_active"`    // Активна ли комната

	// Дополнительные параметры в виде JSON
	// Например: {"color": "#FF5733", "location": "2 этаж", "area_sqm": 25}
	Attributes datatypes.JSON `json:"attributes,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Связи
	Equipment []Equipment `gorm:"foreignKey:RoomID" json:"equipment,omitempty"`
	Bookings  []Booking   `gorm:"foreignKey:RoomID" json:"bookings,omitempty"`
}
