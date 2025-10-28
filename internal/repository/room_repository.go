package repository

import (
	"github.com/space/backend/internal/models"
	"gorm.io/gorm"
)

// RoomRepository handles database operations for rooms
type RoomRepository struct {
	db *gorm.DB
}

// NewRoomRepository creates a new room repository
func NewRoomRepository(db *gorm.DB) *RoomRepository {
	return &RoomRepository{db: db}
}

// Create creates a new room
func (r *RoomRepository) Create(room *models.Room) error {
	return r.db.Create(room).Error
}

// GetByID gets a room by ID with its equipment
func (r *RoomRepository) GetByID(id uint) (*models.Room, error) {
	var room models.Room
	err := r.db.Preload("Equipment").First(&room, id).Error
	if err != nil {
		return nil, err
	}
	return &room, nil
}

// GetAll gets all active rooms
func (r *RoomRepository) GetAll() ([]models.Room, error) {
	var rooms []models.Room
	err := r.db.Where("is_active = ?", true).Preload("Equipment").Order("name").Find(&rooms).Error
	return rooms, err
}

// GetAllWithEquipment gets all active rooms with their equipment
func (r *RoomRepository) GetAllWithEquipment() ([]models.Room, error) {
	var rooms []models.Room
	err := r.db.Where("is_active = ?", true).
		Preload("Equipment").
		Preload("Equipment.Instructions").
		Order("name").
		Find(&rooms).Error
	return rooms, err
}

// Update updates a room
func (r *RoomRepository) Update(room *models.Room) error {
	return r.db.Save(room).Error
}

// Delete soft deletes a room
func (r *RoomRepository) Delete(id uint) error {
	return r.db.Delete(&models.Room{}, id).Error
}

// GetByName gets a room by name
func (r *RoomRepository) GetByName(name string) (*models.Room, error) {
	var room models.Room
	err := r.db.Where("name = ?", name).First(&room).Error
	if err != nil {
		return nil, err
	}
	return &room, nil
}
