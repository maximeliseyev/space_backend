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
