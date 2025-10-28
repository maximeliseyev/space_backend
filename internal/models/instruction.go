package models

import (
	"time"

	"gorm.io/gorm"
)

// InstructionType определяет тип инструкции
type InstructionType string

const (
	InstructionTypeDocument InstructionType = "document" // PDF, DOCX и т.д.
	InstructionTypeVideo    InstructionType = "video"    // Видео инструкция
	InstructionTypeText     InstructionType = "text"     // Текстовая инструкция
	InstructionTypeLink     InstructionType = "link"     // Ссылка на внешний ресурс
)

// Instruction represents instructions for using equipment
type Instruction struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	EquipmentID uint   `gorm:"not null;index" json:"equipment_id"`
	Title       string `gorm:"not null" json:"title"`             // Название инструкции
	Description string `gorm:"type:text" json:"description"`      // Краткое описание
	Type        InstructionType `gorm:"type:varchar(50);not null" json:"type"` // Тип инструкции

	// Путь к файлу в storage или URL
	FilePath string `json:"file_path,omitempty"` // Для document/video
	URL      string `json:"url,omitempty"`       // Для link
	Content  string `gorm:"type:text" json:"content,omitempty"` // Для text

	// Метаданные
	FileSize int64  `json:"file_size,omitempty"` // Размер файла в байтах
	MimeType string `json:"mime_type,omitempty"` // MIME тип файла

	Order int `gorm:"default:0" json:"order"` // Порядок отображения

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Связи
	Equipment Equipment `gorm:"foreignKey:EquipmentID" json:"equipment,omitempty"`
}
