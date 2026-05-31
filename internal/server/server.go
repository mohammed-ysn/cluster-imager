package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mohammed-ysn/cluster-imager/internal/handlers"
	"github.com/mohammed-ysn/cluster-imager/internal/processors"
	"github.com/mohammed-ysn/cluster-imager/pkg/logging"
	"github.com/mohammed-ysn/cluster-imager/pkg/middleware"
)

func StartServer() {
	logger := logging.NewLogger(slog.LevelInfo)
	logger.Info("initializing server")

	registry := processors.DefaultRegistry()
	h := handlers.New(logger, registry)

	mux := http.NewServeMux()
	mux.HandleFunc("/crop", h.CropHandler)
	mux.HandleFunc("/resize", h.ResizeHandler)

	handler := middleware.RequestLogging(logger)(mux)

	srv := &http.Server{
		Addr:           ":8080",
		Handler:        handler,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	serverErrors := make(chan error, 1)

	go func() {
		logger.Info("server started", "addr", srv.Addr)
		serverErrors <- srv.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		logger.Error("server error", "error", err)
		os.Exit(1)
	case sig := <-shutdown:
		logger.Info("shutdown signal received", "signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			logger.Error("graceful shutdown failed", "error", err)
			if err := srv.Close(); err != nil {
				logger.Error("forced shutdown failed", "error", err)
			}
		}
		logger.Info("server stopped")
	}
}
