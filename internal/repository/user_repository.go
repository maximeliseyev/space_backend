package repository

import (
	"log"

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
// NOTE: This method does NOT update existing users. Use SyncFromTelegram() for that.
func (r *UserRepository) GetOrCreate(telegramID int64, username, firstName, lastName, languageCode string) (*models.User, error) {
	user, err := r.GetByTelegramID(telegramID)
	if err == nil {
		// Пользователь существует - возвращаем без изменений
		// Данные пользователя могут быть отредактированы вручную, не перезаписываем их
		log.Printf("DEBUG: Found existing user (ID: %d, TelegramID: %d) - keeping user-managed data", user.ID, user.TelegramID)
		return user, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Создаём нового пользователя с данными из Telegram
	log.Printf("DEBUG: Creating NEW user (TelegramID: %d, username: %s)", telegramID, username)

	user = &models.User{
		TelegramID:   telegramID,
		Username:     username,
		FirstName:    firstName,
		LastName:     lastName,
		LanguageCode: languageCode,
		Role:         models.RoleUser, // По умолчанию обычный пользователь (админ назначается вручную через SQL)
		// Userpic будет установлен через SyncUserpic после создания
	}

	err = r.Create(user)
	if err != nil {
		return nil, err
	}

	log.Printf("DEBUG: Created user with ID: %d, Role: %s", user.ID, user.Role)
	return user, nil
}

// SyncFromTelegram explicitly updates user data from Telegram
// Use this when you want to sync user's Telegram profile changes
func (r *UserRepository) SyncFromTelegram(telegramID int64, username, firstName, lastName, languageCode string) (*models.User, error) {
	user, err := r.GetByTelegramID(telegramID)
	if err != nil {
		return nil, err
	}

	log.Printf("DEBUG: Syncing user ID %d from Telegram", user.ID)
	updated := false

	if user.Username != username {
		log.Printf("DEBUG: Username changed: '%s' -> '%s'", user.Username, username)
		user.Username = username
		updated = true
	}
	if user.FirstName != firstName {
		log.Printf("DEBUG: FirstName changed: '%s' -> '%s'", user.FirstName, firstName)
		user.FirstName = firstName
		updated = true
	}
	if user.LastName != lastName {
		log.Printf("DEBUG: LastName changed: '%s' -> '%s'", user.LastName, lastName)
		user.LastName = lastName
		updated = true
	}
	if user.LanguageCode != languageCode {
		log.Printf("DEBUG: LanguageCode changed: '%s' -> '%s'", user.LanguageCode, languageCode)
		user.LanguageCode = languageCode
		updated = true
	}

	if updated {
		log.Printf("DEBUG: Updating user ID %d with Telegram data", user.ID)
		if err := r.Update(user); err != nil {
			return nil, err
		}
	} else {
		log.Printf("DEBUG: No changes for user ID %d", user.ID)
	}

	return user, nil
}

// SyncUserpic updates user's profile picture URL
func (r *UserRepository) SyncUserpic(telegramID int64, userpicURL string) error {
	user, err := r.GetByTelegramID(telegramID)
	if err != nil {
		return err
	}

	// Обновляем только если URL изменился
	if user.Userpic != userpicURL {
		log.Printf("DEBUG: Updating userpic for user ID %d: '%s' -> '%s'", user.ID, user.Userpic, userpicURL)
		user.Userpic = userpicURL
		return r.Update(user)
	}

	log.Printf("DEBUG: Userpic unchanged for user ID %d", user.ID)
	return nil
}

// UpdateAbout updates user's about/bio field
func (r *UserRepository) UpdateAbout(userID uint, about string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("about", about).Error
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
