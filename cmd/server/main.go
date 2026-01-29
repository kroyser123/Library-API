package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"libraryapi/internal/Storage"
	"libraryapi/internal/api/handlers"
	"libraryapi/internal/api/router"
	"libraryapi/internal/pkg/cache"
	"libraryapi/internal/pkg/logger"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Warn().Msg("Warning: .env file not found, using environment variables")
	}
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "debug"
	}
	logPretty := os.Getenv("LOG_PRETTY") == "true"
	logger.Init(logLevel, logPretty)

	log.Info().Msg("Starting Library API server")
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
	storage := Storage.NewMemory()
	bookHandler := handlers.NewBookHandler(storage, redisCache)
	bookHandler.AddTestBooks()
	mux := router.SetupRouter(bookHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
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
