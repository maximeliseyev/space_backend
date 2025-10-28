package repository

import (
	"github.com/space/backend/internal/models"
	"gorm.io/gorm"
)

// InstructionRepository handles database operations for instructions
type InstructionRepository struct {
	db *gorm.DB
}

// NewInstructionRepository creates a new instruction repository
func NewInstructionRepository(db *gorm.DB) *InstructionRepository {
	return &InstructionRepository{db: db}
}

// Create creates a new instruction
func (r *InstructionRepository) Create(instruction *models.Instruction) error {
	return r.db.Create(instruction).Error
}

// GetByID gets an instruction by ID
func (r *InstructionRepository) GetByID(id uint) (*models.Instruction, error) {
	var instruction models.Instruction
	err := r.db.Preload("Equipment").Preload("Equipment.Room").First(&instruction, id).Error
	if err != nil {
		return nil, err
	}
	return &instruction, nil
}

// GetByEquipmentID gets all instructions for specific equipment
func (r *InstructionRepository) GetByEquipmentID(equipmentID uint) ([]models.Instruction, error) {
	var instructions []models.Instruction
	err := r.db.Where("equipment_id = ?", equipmentID).Order("\"order\" ASC").Find(&instructions).Error
	return instructions, err
}

// Update updates an instruction
func (r *InstructionRepository) Update(instruction *models.Instruction) error {
	return r.db.Save(instruction).Error
}

// Delete soft deletes an instruction
func (r *InstructionRepository) Delete(id uint) error {
	return r.db.Delete(&models.Instruction{}, id).Error
}

// GetAll gets all instructions
func (r *InstructionRepository) GetAll() ([]models.Instruction, error) {
	var instructions []models.Instruction
	err := r.db.Preload("Equipment").Order("equipment_id, \"order\"").Find(&instructions).Error
	return instructions, err
}
