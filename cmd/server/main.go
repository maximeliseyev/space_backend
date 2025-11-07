package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/space/backend/internal/config"
	"github.com/space/backend/internal/database"
	"github.com/space/backend/internal/repository"
	"github.com/space/backend/internal/router"
	"github.com/space/backend/internal/service"
	"github.com/space/backend/pkg/telegram"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Starting Space Backend API in %s mode...", cfg.Environment)

	// Запускаем фоновую очистку кэша членства в группе
	telegram.GlobalCache.StartCleanupRoutine(12 * time.Hour)
	log.Println("Membership cache cleanup routine started")

	// Подключаемся к базе данных
	debugMode := cfg.Environment == "development"
	db, err := database.Connect(cfg.DatabaseURL, debugMode)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Запускаем миграции
	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Инициализируем репозитории
	userRepo := repository.NewUserRepository(db)
	roomRepo := repository.NewRoomRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	equipmentRepo := repository.NewEquipmentRepository(db)
	instructionRepo := repository.NewInstructionRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)

	log.Println("Repositories initialized")

	// Инициализируем сервисы
	userService := service.NewUserService(userRepo)
	userService.SetBotToken(cfg.TelegramBotToken) // Устанавливаем bot token для синхронизации userpic
	roomService := service.NewRoomService(roomRepo, equipmentRepo)
	bookingService := service.NewBookingService(bookingRepo, roomRepo, userRepo)
	notificationService := service.NewNotificationService(notificationRepo, roomRepo)

	log.Println("Services initialized")

	// Настраиваем роутер
	r := router.SetupRouter(
		cfg.TelegramBotToken,
		cfg.BotAPIToken,
		cfg.AllowedChatID,
		cfg.AllowedOrigins,
		cfg.Environment,
		cfg.AuthDateTTLMiniApp,
		cfg.AuthDateTTLLoginWidget,
		userService,
		roomService,
		bookingService,
		notificationService,
	)

	log.Printf("Router configured")

	// Создаем канал для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем сервер в горутине
	go func() {
		addr := ":" + cfg.ServerPort
		log.Printf("Server is starting on http://localhost%s", addr)
		log.Printf("Health check available at http://localhost%s/health", addr)
		log.Printf("API endpoints available at http://localhost%s/api", addr)

		if err := r.Run(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Ожидаем сигнал завершения
	<-quit
	log.Println("Shutting down server...")

	// Закрываем подключение к базе данных
	if err := database.Close(db); err != nil {
		log.Printf("Error closing database: %v", err)
	}

	log.Println("Server gracefully stopped")

	// Используем переменные чтобы избежать ошибки "unused"
	_ = instructionRepo
}
