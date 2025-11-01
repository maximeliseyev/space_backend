package repository

import (
	"github.com/space/backend/internal/models"
	"github.com/space/backend/pkg/validator"
	"gorm.io/gorm"
)

// UserRepository handles database operations for users
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// GetByID gets a user by ID
func (r *UserRepository) GetByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByTelegramID gets a user by Telegram ID
func (r *UserRepository) GetByTelegramID(telegramID int64) (*models.User, error) {
	var user models.User
	err := r.db.Where("telegram_id = ?", telegramID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetOrCreate gets a user by Telegram ID or creates a new one
func (r *UserRepository) GetOrCreate(telegramID int64, username string) (*models.User, error) {
	user, err := r.GetByTelegramID(telegramID)
	if err == nil {
		return user, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Create new user
	user = &models.User{
		TelegramID: telegramID,
		Username:   username,
	}

	err = r.Create(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Update updates a user
func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

// GetPhonebook gets all users in the phonebook
func (r *UserRepository) GetPhonebook() ([]models.User, error) {
	var users []models.User
	err := r.db.Where("is_in_phone_book = ?", true).Order("last_name, first_name").Find(&users).Error
	return users, err
}

// Search searches users by name or username
func (r *UserRepository) Search(query string) ([]models.User, error) {
	var users []models.User
	// Экранируем специальные символы LIKE для безопасности
	escapedQuery := validator.EscapeLike(query)
	searchPattern := "%" + escapedQuery + "%"
	err := r.db.Where(
		"is_in_phone_book = ? AND (first_name ILIKE ? OR last_name ILIKE ? OR username ILIKE ?)",
		true, searchPattern, searchPattern, searchPattern,
	).Order("last_name, first_name").Find(&users).Error
	return users, err
}

// GetByIDs gets multiple users by their IDs
func (r *UserRepository) GetByIDs(ids []uint) ([]models.User, error) {
	var users []models.User
	err := r.db.Where("id IN ?", ids).Find(&users).Error
	return users, err
}
