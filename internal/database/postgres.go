package database

import (
	"fmt"
	"log"

	"github.com/space/backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect creates a connection to PostgreSQL database
func Connect(databaseURL string, debug bool) (*gorm.DB, error) {
	// Настройка логгера
	logLevel := logger.Silent
	if debug {
		logLevel = logger.Info
	}

	// Открываем подключение к PostgreSQL
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Получаем базовый sql.DB для настройки connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Настройка connection pool
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)

	log.Println("✅ Successfully connected to database")
	return db, nil
}

// Migrate runs database migrations
func Migrate(db *gorm.DB) error {
	log.Println("Running database migrations...")

	err := db.AutoMigrate(
		&models.User{},
		&models.Room{},
		&models.Equipment{},
		&models.Instruction{},
		&models.Booking{},
	)

	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Migrations completed successfully")
	return nil
}

// Close closes the database connection
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
