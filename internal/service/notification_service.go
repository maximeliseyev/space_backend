package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/space/backend/internal/config"
	"github.com/space/backend/internal/models"
	"github.com/space/backend/internal/repository"
)

type NotificationService struct {
	notificationRepo *repository.NotificationRepository
	roomRepo         *repository.RoomRepository
	config           *config.Config
}

func NewNotificationService(notificationRepo *repository.NotificationRepository, roomRepo *repository.RoomRepository, cfg *config.Config) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		roomRepo:         roomRepo,
		config:           cfg,
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

// BookingWebhookData represents booking data for webhook
type BookingWebhookData struct {
	BookingID         uint      `json:"booking_id"`
	RoomID            uint      `json:"room_id"`
	RoomName          string    `json:"room_name"`
	StartTime         time.Time `json:"start_time"`
	EndTime           time.Time `json:"end_time"`
	Title             string    `json:"title"`
	CreatorName       string    `json:"creator_name"`
	CreatorTelegramID *int64    `json:"creator_telegram_id,omitempty"`
}

// SubscriberWebhookData represents subscriber data for webhook
type SubscriberWebhookData struct {
	TelegramID int64   `json:"telegram_id"`
	Username   *string `json:"username,omitempty"`
	FirstName  *string `json:"first_name,omitempty"`
}

// BookingCreatedWebhook represents the webhook payload for booking creation
type BookingCreatedWebhook struct {
	Event       string                  `json:"event"`
	Booking     BookingWebhookData      `json:"booking"`
	Subscribers []SubscriberWebhookData `json:"subscribers"`
}

// NotifyBookingCreated sends a webhook notification to the bot about a new booking
func (s *NotificationService) NotifyBookingCreated(booking *models.Booking) error {
	// Получаем подписчиков на комнату
	subscriptions, err := s.GetRoomSubscribers(booking.RoomID)
	if err != nil {
		log.Printf("Failed to get room subscribers: %v", err)
		return err
	}

	// Если нет подписчиков, не отправляем уведомление
	if len(subscriptions) == 0 {
		log.Printf("No subscribers for room %d, skipping notification", booking.RoomID)
		return nil
	}

	// Формируем данные о бронировании
	creatorName := booking.Creator.FirstName
	if booking.Creator.LastName != "" {
		creatorName += " " + booking.Creator.LastName
	}

	var creatorTelegramID *int64
	if booking.Creator.TelegramID != 0 {
		creatorTelegramID = &booking.Creator.TelegramID
	}

	webhookBooking := BookingWebhookData{
		BookingID:         booking.ID,
		RoomID:            booking.RoomID,
		RoomName:          booking.Room.Name,
		StartTime:         booking.StartTime,
		EndTime:           booking.EndTime,
		Title:             booking.Title,
		CreatorName:       creatorName,
		CreatorTelegramID: creatorTelegramID,
	}

	// Формируем список подписчиков
	subscribers := make([]SubscriberWebhookData, 0, len(subscriptions))
	for _, sub := range subscriptions {
		if sub.User != nil && sub.User.TelegramID != 0 {
			var username *string
			if sub.User.Username != "" {
				username = &sub.User.Username
			}

			var firstName *string
			if sub.User.FirstName != "" {
				firstName = &sub.User.FirstName
			}

			subscribers = append(subscribers, SubscriberWebhookData{
				TelegramID: sub.User.TelegramID,
				Username:   username,
				FirstName:  firstName,
			})
		}
	}

	// Создаем webhook payload
	webhook := BookingCreatedWebhook{
		Event:       "booking.created",
		Booking:     webhookBooking,
		Subscribers: subscribers,
	}

	// Отправляем webhook
	return s.sendWebhook(webhook)
}

// sendWebhook sends webhook data to the bot
func (s *NotificationService) sendWebhook(webhook BookingCreatedWebhook) error {
	// Формируем URL
	webhookURL := fmt.Sprintf("%s/webhook/booking/created", s.config.BotWebhookURL)

	// Сериализуем данные в JSON
	jsonData, err := json.Marshal(webhook)
	if err != nil {
		log.Printf("Failed to marshal webhook data: %v", err)
		return fmt.Errorf("failed to marshal webhook data: %w", err)
	}

	// Создаем HTTP запрос
	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to create webhook request: %v", err)
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	// Устанавливаем заголовки
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Bot-Token", s.config.BotAPIToken)

	// Отправляем запрос
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to send webhook: %v", err)
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("Webhook returned non-success status: %d", resp.StatusCode)
		return fmt.Errorf("webhook returned non-success status: %d", resp.StatusCode)
	}

	log.Printf("Successfully sent booking notification to bot for booking #%d", webhook.Booking.BookingID)
	return nil
}
