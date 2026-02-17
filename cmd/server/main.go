package main

import (
	"context"
	"fmt"
	storage "libraryapi/internal/Storage/postgres"
	"libraryapi/internal/api/handlers"
	"libraryapi/internal/api/router"
	"libraryapi/internal/pkg/cache"
	"libraryapi/internal/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Warn().Msg("Warning: .env file not found, using environment variables")
	}

	// Инициализация логгера
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "debug"
	}
	logPretty := os.Getenv("LOG_PRETTY") == "true"
	logger.Init(logLevel, logPretty)

	log.Info().Msg("Starting Library API server")

	// 1. Инициализация Redis
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}
	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}
	redisPassword := os.Getenv("REDIS_PASSWORD")

	redisCache := cache.NewRedisCache(redisHost, redisPort, redisPassword, 0)
	defer func() {
		if err := redisCache.Close(); err != nil {
			log.Error().Err(err).Msg("Error closing Redis connection")
		}
	}()

	// 2. Инициализация PostgreSQL
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "postgres"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "library"
	}
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	log.Info().Str("connection", connStr).Msg("Connecting to PostgreSQL")

	storage, err := storage.NewPostgres(connStr)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to PostgreSQL")
	}
	log.Info().Msg("Successfully connected to PostgreSQL")

	// 3. Инициализация обработчиков
	bookHandler := handlers.NewBookHandler(storage, redisCache)

	// 4. Настройка роутера
	mux := router.SetupRouter(bookHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// 5. shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Info().Str("port", port).Msg("Server starting")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	<-stop
	log.Info().Msg("Shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server shutdown error")
	}

	log.Info().Msg("Server stopped gracefully")
}
