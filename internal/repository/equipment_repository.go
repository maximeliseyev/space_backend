package repository

import (
	"github.com/space/backend/internal/models"
	"gorm.io/gorm"
)

// EquipmentRepository handles database operations for equipment
type EquipmentRepository struct {
	db *gorm.DB
}

// NewEquipmentRepository creates a new equipment repository
func NewEquipmentRepository(db *gorm.DB) *EquipmentRepository {
	return &EquipmentRepository{db: db}
}

// Create creates new equipment
func (r *EquipmentRepository) Create(equipment *models.Equipment) error {
	return r.db.Create(equipment).Error
}

// GetByID gets equipment by ID with instructions
func (r *EquipmentRepository) GetByID(id uint) (*models.Equipment, error) {
	var equipment models.Equipment
	err := r.db.Preload("Instructions", func(db *gorm.DB) *gorm.DB {
		return db.Order("\"order\" ASC")
	}).Preload("Room").First(&equipment, id).Error
	if err != nil {
		return nil, err
	}
	return &equipment, nil
}

// GetByRoomID gets all equipment for a specific room
func (r *EquipmentRepository) GetByRoomID(roomID uint) ([]models.Equipment, error) {
	var equipment []models.Equipment
	err := r.db.Preload("Instructions", func(db *gorm.DB) *gorm.DB {
		return db.Order("\"order\" ASC")
	}).Where("room_id = ?", roomID).Order("name").Find(&equipment).Error
	return equipment, err
}

// Update updates equipment
func (r *EquipmentRepository) Update(equipment *models.Equipment) error {
	return r.db.Save(equipment).Error
}

// Delete soft deletes equipment
func (r *EquipmentRepository) Delete(id uint) error {
	return r.db.Delete(&models.Equipment{}, id).Error
}

// GetAll gets all equipment
func (r *EquipmentRepository) GetAll() ([]models.Equipment, error) {
	var equipment []models.Equipment
	err := r.db.Preload("Room").Preload("Instructions").Order("name").Find(&equipment).Error
	return equipment, err
}
