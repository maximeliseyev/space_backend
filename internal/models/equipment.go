package models

import (
	"time"

	"gorm.io/gorm"
)

// Equipment represents equipment available in a room
type Equipment struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	RoomID      uint   `gorm:"not null;index" json:"room_id"`
	Name        string `gorm:"not null" json:"name"`        // Название оборудования (проектор, сканер, проигрыватель и т.д.)
	Description string `gorm:"type:text" json:"description"` // Описание оборудования
	IsAvailable bool   `gorm:"default:true" json:"is_available"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Связи
	Room         Room         `gorm:"foreignKey:RoomID" json:"room,omitempty"`
	Instructions []Instruction `gorm:"foreignKey:EquipmentID" json:"instructions,omitempty"`
}
