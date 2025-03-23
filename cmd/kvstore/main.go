package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"testKvStore/internal/config"
	"testKvStore/internal/handlers"
	"testKvStore/internal/logger"
	"testKvStore/internal/storage"
	"testKvStore/pkg/middleware"

	"github.com/go-chi/chi/v5"
)

func main() {
	log := logger.Init()
	cfg := config.LoadConfig()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	store, err := storage.NewTarantoolStorage(cfg.TarantoolHost, cfg.TarantoolUser, cfg.TarantoolPass)
	if err != nil {
		log.Fatalf("Не удалось инициализировать TarantoolStorage: %v", err)
	}
	defer store.Close()

	router := chi.NewRouter()
	router.Use(middleware.RequestLogger)
	router.Use(middleware.Timeout(60 * time.Second))
	router.Use(middleware.Recoverer)

	handlers.RegisterRoutes(router, store)

	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("Сервер слушает на %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка ListenAndServe: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Завершение работы, идёт остановка сервера...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Ошибка при остановке сервера: %v", err)
	}

	log.Println("Сервер остановлен корректно.")
}
