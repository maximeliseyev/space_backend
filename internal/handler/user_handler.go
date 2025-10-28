package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/space/backend/internal/service"
	"github.com/space/backend/pkg/response"
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
