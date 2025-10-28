package service

import (
	"github.com/space/backend/internal/models"
	"github.com/space/backend/internal/repository"
)

// RoomService handles room business logic
type RoomService struct {
	roomRepo      *repository.RoomRepository
	equipmentRepo *repository.EquipmentRepository
}

// NewRoomService creates a new room service
func NewRoomService(roomRepo *repository.RoomRepository, equipmentRepo *repository.EquipmentRepository) *RoomService {
	return &RoomService{
		roomRepo:      roomRepo,
		equipmentRepo: equipmentRepo,
	}
}

// GetAllRooms gets all active rooms
func (s *RoomService) GetAllRooms() ([]models.Room, error) {
	return s.roomRepo.GetAll()
}

// GetAllRoomsWithEquipment gets all rooms with their equipment and instructions
func (s *RoomService) GetAllRoomsWithEquipment() ([]models.Room, error) {
	return s.roomRepo.GetAllWithEquipment()
}

// GetRoom gets a room by ID with equipment
func (s *RoomService) GetRoom(id uint) (*models.Room, error) {
	return s.roomRepo.GetByID(id)
}

// GetRoomEquipment gets all equipment for a specific room
func (s *RoomService) GetRoomEquipment(roomID uint) ([]models.Equipment, error) {
	return s.equipmentRepo.GetByRoomID(roomID)
}

// CreateRoomRequest represents a request to create a room
type CreateRoomRequest struct {
	Name        string      `json:"name" binding:"required"`
	Description string      `json:"description"`
	Capacity    int         `json:"capacity"`
	Attributes  interface{} `json:"attributes"`
}

// CreateRoom creates a new room (admin only)
func (s *RoomService) CreateRoom(req CreateRoomRequest) (*models.Room, error) {
	room := &models.Room{
		Name:        req.Name,
		Description: req.Description,
		Capacity:    req.Capacity,
		IsActive:    true,
	}

	err := s.roomRepo.Create(room)
	if err != nil {
		return nil, err
	}

	return s.roomRepo.GetByID(room.ID)
}

// UpdateRoomRequest represents a request to update a room
type UpdateRoomRequest struct {
	Name        *string     `json:"name"`
	Description *string     `json:"description"`
	Capacity    *int        `json:"capacity"`
	IsActive    *bool       `json:"is_active"`
	Attributes  interface{} `json:"attributes"`
}

// UpdateRoom updates a room (admin only)
func (s *RoomService) UpdateRoom(id uint, req UpdateRoomRequest) (*models.Room, error) {
	room, err := s.roomRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		room.Name = *req.Name
	}
	if req.Description != nil {
		room.Description = *req.Description
	}
	if req.Capacity != nil {
		room.Capacity = *req.Capacity
	}
	if req.IsActive != nil {
		room.IsActive = *req.IsActive
	}

	err = s.roomRepo.Update(room)
	if err != nil {
		return nil, err
	}

	return s.roomRepo.GetByID(id)
}

// DeleteRoom soft deletes a room (admin only)
func (s *RoomService) DeleteRoom(id uint) error {
	return s.roomRepo.Delete(id)
}
