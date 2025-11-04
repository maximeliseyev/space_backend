package models

import (
	"time"

	"gorm.io/gorm"
)

// NotificationSubscription represents a user's subscription to room booking notifications
type NotificationSubscription struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null;index:idx_user_room" json:"user_id"`
	RoomID    uint           `gorm:"not null;index:idx_user_room" json:"room_id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Room *Room `gorm:"foreignKey:RoomID" json:"room,omitempty"`
}

// TableName specifies the table name for NotificationSubscription
func (NotificationSubscription) TableName() string {
	return "notification_subscriptions"
}
