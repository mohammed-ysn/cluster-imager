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
	// Initialize logger
	logger := logging.NewLogger(slog.LevelInfo)
	logger.Info("initializing server")

	// Create processor registry
	registry := processors.DefaultRegistry()

	// Create handlers
	h := handlers.New(logger, registry)

	// Create a new mux instead of using default
	mux := http.NewServeMux()
	mux.HandleFunc("/crop", h.CropHandler)
	mux.HandleFunc("/resize", h.ResizeHandler)

	// Apply middleware
	handler := middleware.RequestLogging(logger)(mux)

	server := &http.Server{
		Addr:           ":8080",
		Handler:        handler,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB max header size
	}

	// Setup graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	serverErrors := make(chan error, 1)

	// Start the server
	go func() {
		logger.Info("server started", "addr", server.Addr)
		serverErrors <- server.ListenAndServe()
	}()

	// Wait for shutdown or error
	select {
	case err := <-serverErrors:
		logger.Error("server error", "error", err)
		os.Exit(1)
	case sig := <-shutdown:
		logger.Info("shutdown signal received", "signal", sig)

		// Give outstanding requests 5 seconds to complete
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Error("graceful shutdown failed", "error", err)
			if err := server.Close(); err != nil {
				logger.Error("forced shutdown failed", "error", err)
			}
		}
		logger.Info("server stopped")
	}
}
