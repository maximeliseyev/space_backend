package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/space/backend/internal/models"
	"github.com/space/backend/internal/service"
	"github.com/space/backend/pkg/response"
	"github.com/space/backend/pkg/telegram"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetProfile godoc
// @Summary Get current user profile
// @Tags users
// @Produce json
// @Success 200 {object} models.User
// @Router /api/users/me [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, service.ErrNotAuthorized)
		return
	}

	user, err := h.userService.GetUser(userID.(uint))
	if err != nil {
		response.NotFound(c, err)
		return
	}

	response.Success(c, user)
}

// UpdateProfile godoc
// @Summary Update current user profile
// @Tags users
// @Accept json
// @Produce json
// @Param profile body service.UpdateProfileRequest true "Profile data"
// @Success 200 {object} models.User
// @Router /api/users/me [patch]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, service.ErrNotAuthorized)
		return
	}

	var req service.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err)
		return
	}

	user, err := h.userService.UpdateProfile(userID.(uint), req)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.Success(c, user)
}

// GetPhonebook godoc
// @Summary Get phonebook (all users with name and phone)
// @Tags users
// @Produce json
// @Param q query string false "Search query"
// @Success 200 {array} models.User
// @Router /api/users/phonebook [get]
func (h *UserHandler) GetPhonebook(c *gin.Context) {
	query := c.Query("q")

	users, err := h.userService.SearchPhonebook(query)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.Success(c, users)
}

// SyncFromTelegram godoc
// @Summary Sync user profile from Telegram
// @Description Updates user's profile with current data from Telegram (name, username, etc.)
// @Description User must provide fresh Telegram initData in X-Telegram-Init-Data header
// @Tags users
// @Produce json
// @Success 200 {object} models.User
// @Router /api/users/me/sync-telegram [post]
func (h *UserHandler) SyncFromTelegram(c *gin.Context) {
	// Получаем текущего пользователя из контекста
	userInterface, exists := c.Get("user")
	if !exists {
		response.Unauthorized(c, service.ErrNotAuthorized)
		return
	}

	user := userInterface.(*models.User)

	// Получаем данные из Telegram, которые пришли в текущем запросе
	// Middleware уже валидировал initData и извлек telegramUser
	telegramUserInterface, exists := c.Get("telegramUser")
	if !exists {
		response.BadRequest(c, errors.New("telegram user data not found in request context"))
		return
	}

	telegramUser := telegramUserInterface.(*telegram.TelegramUser)

	// Синхронизируем данные из Telegram
	updatedUser, err := h.userService.SyncUserFromTelegram(
		telegramUser.ID,
		telegramUser.Username,
		telegramUser.FirstName,
		telegramUser.LastName,
		telegramUser.LanguageCode,
	)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	// Проверяем, что это тот же пользователь
	if updatedUser.ID != user.ID {
		response.BadRequest(c, errors.New("cannot sync different user's data"))
		return
	}

	response.Success(c, updatedUser)
}
