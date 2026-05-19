package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Siargio/shortener/internal/handler"
	"github.com/Siargio/shortener/internal/repository"
	"github.com/Siargio/shortener/internal/service"
	"github.com/Siargio/shortener/pkg/cache"
	"github.com/Siargio/shortener/pkg/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctxRoot := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config:", err)
	}

	// Подключение к PostgreSQL
	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	ctxDB, cancelDB := context.WithTimeout(ctxRoot, 5*time.Second)
	dbPool, err := pgxpool.New(ctxDB, dbURL)
	if err != nil {
		log.Fatal("failed to connect to db:", err)
	}
	cancelDB()
	defer dbPool.Close()

	// Подключение к Redis
	redisAddr := fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort)
	redisCache := cache.NewRedisCache(redisAddr, cfg.RedisPassword)
	defer redisCache.Close()

	// Инициализация слоёв
	linkRepo := repository.NewPostgresLinkRepo(dbPool)
	linkService := service.NewLinkService(linkRepo, redisCache, cfg.BaseURL)
	linkHandler := handler.NewLinkHandler(linkService)

	// Маршруты
	mux := http.NewServeMux()
	mux.HandleFunc("POST /shorten", linkHandler.Shorten)
	mux.HandleFunc("GET /{code}", linkHandler.Redirect)
	mux.HandleFunc("GET /stats/{code}", linkHandler.Stats)

	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: mux,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Server starting on :%s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server failed:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(ctxRoot, 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("server shutdown failed:", err)
	}
	log.Println("Server stopped")
}
