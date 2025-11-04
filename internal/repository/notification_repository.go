package repository

import (
	"github.com/space/backend/internal/models"
	"gorm.io/gorm"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Subscribe creates a subscription for a user to receive notifications about a room
func (r *NotificationRepository) Subscribe(userID uint, roomID uint) error {
	// Проверяем что подписка не существует
	var existing models.NotificationSubscription
	err := r.db.Where("user_id = ? AND room_id = ?", userID, roomID).First(&existing).Error

	if err == nil {
		// Подписка уже существует
		return nil
	}

	if err != gorm.ErrRecordNotFound {
		return err
	}

	// Создаём новую подписку
	subscription := models.NotificationSubscription{
		UserID: userID,
		RoomID: roomID,
	}

	return r.db.Create(&subscription).Error
}

// Unsubscribe removes a subscription
func (r *NotificationRepository) Unsubscribe(userID uint, roomID uint) error {
	return r.db.Where("user_id = ? AND room_id = ?", userID, roomID).
		Delete(&models.NotificationSubscription{}).Error
}

// GetUserSubscriptions returns all rooms a user is subscribed to
func (r *NotificationRepository) GetUserSubscriptions(userID uint) ([]models.NotificationSubscription, error) {
	var subscriptions []models.NotificationSubscription
	err := r.db.Preload("Room").Where("user_id = ?", userID).Find(&subscriptions).Error
	return subscriptions, err
}

// GetRoomSubscribers returns all users subscribed to a room
func (r *NotificationRepository) GetRoomSubscribers(roomID uint) ([]models.NotificationSubscription, error) {
	var subscriptions []models.NotificationSubscription
	err := r.db.Preload("User").Where("room_id = ?", roomID).Find(&subscriptions).Error
	return subscriptions, err
}

// IsSubscribed checks if a user is subscribed to a room
func (r *NotificationRepository) IsSubscribed(userID uint, roomID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.NotificationSubscription{}).
		Where("user_id = ? AND room_id = ?", userID, roomID).
		Count(&count).Error
	return count > 0, err
}
