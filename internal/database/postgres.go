package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
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

	// Создаем кастомный dialer который использует только IPv4
	ipv4Dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		// Принудительно используем только IPv4
		FallbackDelay: -1, // Отключаем fallback на IPv6
	}

	// Парсим connection string
	config, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Устанавливаем кастомный DialFunc который использует только IPv4
	config.DialFunc = func(ctx context.Context, network, addr string) (net.Conn, error) {
		// Принудительно используем tcp4 вместо tcp (который может быть tcp6)
		return ipv4Dialer.DialContext(ctx, "tcp4", addr)
	}

	// Регистрируем драйвер с кастомной конфигурацией
	connStr := stdlib.RegisterConnConfig(config)

	// Открываем подключение через sql.Open
	sqlDB, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Настраиваем connection pool
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Проверяем подключение
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Создаем GORM DB из существующего подключения
	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("✅ Successfully connected to database (IPv4)")
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
