package service

import (
	"github.com/space/backend/internal/models"
	"github.com/space/backend/internal/repository"
)

type NotificationService struct {
	notificationRepo *repository.NotificationRepository
	roomRepo         *repository.RoomRepository
}

func NewNotificationService(notificationRepo *repository.NotificationRepository, roomRepo *repository.RoomRepository) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		roomRepo:         roomRepo,
	}
}

// Subscribe subscribes a user to room notifications
func (s *NotificationService) Subscribe(userID uint, roomID uint) error {
	// Проверяем что комната существует
	_, err := s.roomRepo.GetByID(roomID)
	if err != nil {
		return err
	}

	return s.notificationRepo.Subscribe(userID, roomID)
}

// Unsubscribe unsubscribes a user from room notifications
func (s *NotificationService) Unsubscribe(userID uint, roomID uint) error {
	return s.notificationRepo.Unsubscribe(userID, roomID)
}

// GetUserSubscriptions returns all rooms a user is subscribed to
func (s *NotificationService) GetUserSubscriptions(userID uint) ([]models.NotificationSubscription, error) {
	return s.notificationRepo.GetUserSubscriptions(userID)
}

// GetRoomSubscribers returns all users subscribed to a room
func (s *NotificationService) GetRoomSubscribers(roomID uint) ([]models.NotificationSubscription, error) {
	return s.notificationRepo.GetRoomSubscribers(roomID)
}

// IsSubscribed checks if a user is subscribed to a room
func (s *NotificationService) IsSubscribed(userID uint, roomID uint) (bool, error) {
	return s.notificationRepo.IsSubscribed(userID, roomID)
}
