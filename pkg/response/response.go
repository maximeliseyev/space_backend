package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// Success sends a successful JSON response
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, SuccessResponse{
		Data: data,
	})
}

// SuccessWithMessage sends a successful JSON response with a message
func SuccessWithMessage(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusOK, SuccessResponse{
		Data:    data,
		Message: message,
	})
}

// Created sends a 201 Created response
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, SuccessResponse{
		Data: data,
	})
}

// NoContent sends a 204 No Content response
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Error sends an error JSON response
func Error(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, ErrorResponse{
		Error: err.Error(),
	})
}

// ErrorWithMessage sends an error JSON response with a custom message
func ErrorWithMessage(c *gin.Context, statusCode int, err error, message string) {
	c.JSON(statusCode, ErrorResponse{
		Error:   err.Error(),
		Message: message,
	})
}

// BadRequest sends a 400 Bad Request response
func BadRequest(c *gin.Context, err error) {
	Error(c, http.StatusBadRequest, err)
}

// Unauthorized sends a 401 Unauthorized response
func Unauthorized(c *gin.Context, err error) {
	Error(c, http.StatusUnauthorized, err)
}

// Forbidden sends a 403 Forbidden response
func Forbidden(c *gin.Context, err error) {
	Error(c, http.StatusForbidden, err)
}

// NotFound sends a 404 Not Found response
func NotFound(c *gin.Context, err error) {
	Error(c, http.StatusNotFound, err)
}

// Conflict sends a 409 Conflict response
func Conflict(c *gin.Context, err error) {
	Error(c, http.StatusConflict, err)
}

// ConflictWithData sends a 409 Conflict response with additional data
func ConflictWithData(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusConflict, gin.H{
		"error": message,
		"data":  data,
	})
}

// InternalServerError sends a 500 Internal Server Error response
func InternalServerError(c *gin.Context, err error) {
	Error(c, http.StatusInternalServerError, err)
}
