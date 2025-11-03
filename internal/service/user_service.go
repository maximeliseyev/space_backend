package service

import (
	"github.com/space/backend/internal/models"
	"github.com/space/backend/internal/repository"
)

// UserService handles user business logic
type UserService struct {
	userRepo *repository.UserRepository
}

// NewUserService creates a new user service
func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// SyncTelegramUser syncs a user from Telegram (get or create)
// NOTE: This does NOT update existing users automatically
func (s *UserService) SyncTelegramUser(telegramID int64, username, firstName, lastName, languageCode string) (*models.User, error) {
	return s.userRepo.GetOrCreate(telegramID, username, firstName, lastName, languageCode)
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

	err = s.userRepo.Update(user)
	if err != nil {
		return nil, err
	}

	return user, nil
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
