package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mohammed-ysn/cluster-imager/internal/config"
	"github.com/mohammed-ysn/cluster-imager/internal/handlers"
	"github.com/mohammed-ysn/cluster-imager/internal/processors"
	"github.com/mohammed-ysn/cluster-imager/internal/worker"
	"github.com/mohammed-ysn/cluster-imager/pkg/job"
	"github.com/mohammed-ysn/cluster-imager/pkg/logging"
	"github.com/mohammed-ysn/cluster-imager/pkg/middleware"
	"github.com/mohammed-ysn/cluster-imager/pkg/queue"
	"github.com/mohammed-ysn/cluster-imager/pkg/storage"
)

func StartServer() {
	cfg := config.Load()
	logger := logging.NewLogger(slog.LevelInfo)
	logger.Info("initializing server")

	stor, err := storage.NewLocalStorage(cfg.Storage.LocalPath)
	if err != nil {
		logger.Error("failed to init storage", "error", err)
		os.Exit(1)
	}

	jobStore, err := job.NewRedisStore(cfg.Redis.URL, "jobs", cfg.Job.TTL)
	if err != nil {
		logger.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}
	defer jobStore.Close()

	q, err := queue.NewNATSQueue(queue.Config{
		URL:      cfg.NATS.URL,
		Stream:   cfg.NATS.Stream,
		Subject:  cfg.NATS.Subject,
		Consumer: cfg.NATS.Consumer,
		MaxRetry: cfg.NATS.MaxRetry,
	})
	if err != nil {
		logger.Error("failed to connect to nats", "error", err)
		_ = jobStore.Close()
		os.Exit(1)
	}
	defer q.Close()

	registry := processors.DefaultRegistry()

	w := worker.New(q, jobStore, stor, registry, logger)

	h := handlers.New(logger, registry, jobStore, stor, q)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/live", h.LiveHandler)
	mux.HandleFunc("GET /health/ready", h.ReadyHandler)
	mux.HandleFunc("POST /api/v1/crop", h.CropHandler)
	mux.HandleFunc("POST /api/v1/resize", h.ResizeHandler)
	mux.HandleFunc("GET /api/v1/jobs/{id}", h.JobStatusHandler)

	handler := middleware.RequestLogging(logger)(mux)

	srv := &http.Server{
		Addr:           fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:        handler,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := w.Run(ctx); err != nil && ctx.Err() == nil {
			logger.Error("worker error", "error", err)
		}
	}()

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
		cancel()

		shutCtx, shutCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutCancel()

		if err := srv.Shutdown(shutCtx); err != nil {
			logger.Error("graceful shutdown failed", "error", err)
			if err := srv.Close(); err != nil {
				logger.Error("forced close failed", "error", err)
			}
		}
		logger.Info("server stopped")
	}
}
