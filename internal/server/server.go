package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"servicesubs/internal/api"
	"servicesubs/internal/config"
	"syscall"
	"time"
)

func Start(cfg *config.Config) {

	mux := http.NewServeMux()

	api.Init(mux, cfg)

	// slog.Debug("Start server")
	// err := http.ListenAndServe(":"+cfg.HTTP.Port, mux)
	// if err != nil {
	// 	slog.Error(err.Error())
	// 	os.Exit(1)
	// }

	srv := &http.Server{
		Addr:         ":" + cfg.HTTP.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Запускаем сервер в горутине (не блокируем)
	go func() {
		slog.Info("Starting server", "port", cfg.HTTP.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Ожидаем сигнал остановки (Ctrl+C или SIGTERM)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server gracefully...")

	// Даем 30 секунд на завершение текущих запросов
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Плавное завершение
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	// Если дошли сюда - сервер завершился корректно
	slog.Info("Server stopped gracefully")
}
