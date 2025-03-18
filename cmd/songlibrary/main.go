// cmd/songlibrary/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"

	"songlibrary/config"
	"songlibrary/internal/api/handlers/songs"
	"songlibrary/internal/lib/logger/utils"
	"songlibrary/internal/musicapi"
	"songlibrary/internal/service"
	"songlibrary/internal/storage/postgres"
	_ "songlibrary/swagger" // Import generated swagger docs

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
)

// @title Online Library API
// @version 1.0
// @description This is a sample online library API for songs.

// @host localhost:8080
// @BasePath /
// @schemes http

func main() {
	// 1. Инициализация логгера
	if err := utils.InitLogger(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer utils.Logger.Sync()

	utils.Logger.Info("Starting Song Library API")

	// 2. Загрузка конфигурации
	godotenv.Load() // Загрузка .env файла
	cfg, err := config.LoadConfig()
	if err != nil {
		utils.Logger.Fatal("Config load failed", zap.Error(err))
		return
	}
	utils.Logger.Debug("Configuration loaded", zap.Any("config", cfg))

	// 3. Подключение к БД и запуск миграций
	conn, err := pgx.Connect(context.Background(), cfg.DBURL)
	if err != nil {
		utils.Logger.Fatal("Database connection failed", zap.Error(err))
		return
	}
	defer conn.Close(context.Background())
	utils.Logger.Info("Database connected")

	if err := runMigrations(cfg.DBURL); err != nil {
		utils.Logger.Fatal("Database migration failed", zap.Error(err))
		return
	}
	utils.Logger.Info("Database migrations completed successfully")

	// 4. Инициализация хранилища, music API клиента и сервиса
	pgStorage := postgres.NewPgStorage(conn)
	musicAPIClient := musicapi.NewMusicAPIClient(cfg.APIURL)
	songService := service.NewSongService(pgStorage, musicAPIClient)

	// 5. Инициализация обработчиков API
	songHandlers := songs.NewSongHandlers(songService)

	// 6. Настройка роутера
	router := mux.NewRouter()

	// Регистрация эндпоинтов
	router.HandleFunc("/health", songHandlers.HealthCheckHandler).Methods("GET")
	router.HandleFunc("/songs", songHandlers.GetSongsHandler).Methods("GET")
	router.HandleFunc("/songs", songHandlers.AddSongHandler).Methods("POST")
	router.HandleFunc("/songs/{id}/text", songHandlers.GetSongTextHandler).Methods("GET")
	router.HandleFunc("/songs/{id}", songHandlers.UpdateSongHandler).Methods("PUT")
	router.HandleFunc("/songs/{id}", songHandlers.DeleteSongHandler).Methods("DELETE")

	// Swagger documentation
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// 7. Запуск сервера
	serverAddr := fmt.Sprintf(":%d", cfg.ServerPort)
	utils.Logger.Info("Server starting", zap.String("address", serverAddr))
	log.Fatal(http.ListenAndServe(serverAddr, router))
}

func runMigrations(dbURL string) error {
	migrationSourceURL := "file://internal/migrations"
	m, err := migrate.New(migrationSourceURL, dbURL)
	if err != nil {
		return fmt.Errorf("failed to initialize migration: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
