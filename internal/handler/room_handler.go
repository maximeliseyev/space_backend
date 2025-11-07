package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/space/backend/internal/service"
	"github.com/space/backend/pkg/response"
)

// RoomHandler handles room-related HTTP requests
type RoomHandler struct {
	roomService *service.RoomService
}

// NewRoomHandler creates a new room handler
func NewRoomHandler(roomService *service.RoomService) *RoomHandler {
	return &RoomHandler{roomService: roomService}
}

// GetAllRooms godoc
// @Summary Get all rooms
// @Tags rooms
// @Produce json
// @Param with_equipment query bool false "Include equipment"
// @Success 200 {array} models.Room
// @Router /api/rooms [get]
func (h *RoomHandler) GetAllRooms(c *gin.Context) {
	withEquipment := c.Query("with_equipment") == "true"

	var rooms interface{}
	var err error

	if withEquipment {
		rooms, err = h.roomService.GetAllRoomsWithEquipment()
	} else {
		rooms, err = h.roomService.GetAllRooms()
	}

	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.Success(c, rooms)
}

// GetRoom godoc
// @Summary Get room by ID
// @Tags rooms
// @Produce json
// @Param id path int true "Room ID"
// @Success 200 {object} models.Room
// @Router /api/rooms/{id} [get]
func (h *RoomHandler) GetRoom(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	room, err := h.roomService.GetRoom(uint(id))
	if err != nil {
		response.NotFound(c, err)
		return
	}

	response.Success(c, room)
}

// GetRoomEquipment godoc
// @Summary Get room equipment
// @Tags rooms
// @Produce json
// @Param id path int true "Room ID"
// @Success 200 {array} models.Equipment
// @Router /api/rooms/{id}/equipment [get]
func (h *RoomHandler) GetRoomEquipment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	equipment, err := h.roomService.GetRoomEquipment(uint(id))
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.Success(c, equipment)
}

// CreateRoom godoc
// @Summary Create a new room (admin only)
// @Tags rooms
// @Accept json
// @Produce json
// @Param room body service.CreateRoomRequest true "Room data"
// @Success 201 {object} models.Room
// @Router /api/rooms [post]
func (h *RoomHandler) CreateRoom(c *gin.Context) {
	var req service.CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err)
		return
	}

	room, err := h.roomService.CreateRoom(req)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.Created(c, room)
}

// UpdateRoom godoc
// @Summary Update a room (admin only)
// @Tags rooms
// @Accept json
// @Produce json
// @Param id path int true "Room ID"
// @Param room body service.UpdateRoomRequest true "Room data"
// @Success 200 {object} models.Room
// @Router /api/rooms/{id} [patch]
func (h *RoomHandler) UpdateRoom(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	var req service.UpdateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err)
		return
	}

	room, err := h.roomService.UpdateRoom(uint(id), req)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.Success(c, room)
}

// DeleteRoom godoc
// @Summary Delete a room (admin only)
// @Tags rooms
// @Param id path int true "Room ID"
// @Success 204
// @Router /api/rooms/{id} [delete]
func (h *RoomHandler) DeleteRoom(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	err = h.roomService.DeleteRoom(uint(id))
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.NoContent(c)
}
