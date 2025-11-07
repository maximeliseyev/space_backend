package service

import (
	"log"

	"github.com/space/backend/internal/models"
	"github.com/space/backend/internal/repository"
	"github.com/space/backend/pkg/telegram"
)

// UserService handles user business logic
type UserService struct {
	userRepo *repository.UserRepository
	botToken string // Нужен для получения фото профиля из Telegram
}

// NewUserService creates a new user service
func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// SetBotToken sets the bot token for Telegram API calls
func (s *UserService) SetBotToken(botToken string) {
	s.botToken = botToken
}

// SyncTelegramUser syncs a user from Telegram (get or create)
// NOTE: This does NOT update existing users automatically
func (s *UserService) SyncTelegramUser(telegramID int64, username, firstName, lastName, languageCode string) (*models.User, error) {
	user, err := s.userRepo.GetOrCreate(telegramID, username, firstName, lastName, languageCode)
	if err != nil {
		return nil, err
	}

	// Асинхронно обновляем userpic из Telegram (не блокируем запрос)
	if s.botToken != "" {
		go s.syncUserpicAsync(telegramID)
	}

	return user, nil
}

// syncUserpicAsync асинхронно обновляет userpic пользователя из Telegram
func (s *UserService) syncUserpicAsync(telegramID int64) {
	userpicURL, err := telegram.GetUserProfilePhotoURL(telegramID, s.botToken)
	if err != nil {
		log.Printf("WARNING: Failed to get userpic for user %d: %v", telegramID, err)
		return
	}

	// Обновляем только если фото есть
	if userpicURL != "" {
		if err := s.userRepo.SyncUserpic(telegramID, userpicURL); err != nil {
			log.Printf("WARNING: Failed to sync userpic for user %d: %v", telegramID, err)
		}
	}
}

// SyncUserFromTelegram explicitly updates user data from Telegram
// Use this when user wants to sync their Telegram profile changes
func (s *UserService) SyncUserFromTelegram(telegramID int64, username, firstName, lastName, languageCode string) (*models.User, error) {
	return s.userRepo.SyncFromTelegram(telegramID, username, firstName, lastName, languageCode)
}

// GetUser gets a user by ID
func (s *UserService) GetUser(id uint) (*models.User, error) {
	return s.userRepo.GetByID(id)
}

// GetUserByTelegramID gets a user by Telegram ID
func (s *UserService) GetUserByTelegramID(telegramID int64) (*models.User, error) {
	return s.userRepo.GetByTelegramID(telegramID)
}

// UpdateProfileRequest represents a request to update user profile
type UpdateProfileRequest struct {
	FirstName   *string `json:"first_name"`
	LastName    *string `json:"last_name"`
	PhoneNumber *string `json:"phone_number"`
	About       *string `json:"about"` // Новое поле
}

// UpdateProfile updates user profile
func (s *UserService) UpdateProfile(userID uint, req UpdateProfileRequest) (*models.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	// Обновляем поля
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.PhoneNumber != nil {
		user.PhoneNumber = *req.PhoneNumber
	}
	if req.About != nil {
		user.About = *req.About
	}

	err = s.userRepo.Update(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// CanEditUser проверяет, может ли пользователь редактировать другого пользователя
func (s *UserService) CanEditUser(currentUser *models.User, targetUserID uint) bool {
	// Админ может редактировать всех
	if currentUser.IsAdmin() {
		return true
	}
	// Пользователь может редактировать только себя
	return currentUser.ID == targetUserID
}

// GetPhonebook gets all users in the phonebook
func (s *UserService) GetPhonebook() ([]models.User, error) {
	return s.userRepo.GetPhonebook()
}

// SearchPhonebook searches users in the phonebook
func (s *UserService) SearchPhonebook(query string) ([]models.User, error) {
	if query == "" {
		return s.GetPhonebook()
	}
	return s.userRepo.Search(query)
}
