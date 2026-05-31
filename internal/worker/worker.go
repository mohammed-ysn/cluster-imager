package worker

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"log/slog"

	"github.com/mohammed-ysn/cluster-imager/internal/processors"
	"github.com/mohammed-ysn/cluster-imager/pkg/job"
	"github.com/mohammed-ysn/cluster-imager/pkg/logging"
	"github.com/mohammed-ysn/cluster-imager/pkg/queue"
	"github.com/mohammed-ysn/cluster-imager/pkg/storage"
)

type Worker struct {
	queue    queue.Consumer
	jobs     job.Store
	storage  storage.Storage
	registry *processors.Registry
	logger   *logging.Logger
}

func New(q queue.Consumer, jobs job.Store, stor storage.Storage, registry *processors.Registry, logger *logging.Logger) *Worker {
	return &Worker{
		queue:    q,
		jobs:     jobs,
		storage:  stor,
		registry: registry,
		logger:   logger,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	w.logger.Info("worker started")
	return w.queue.Subscribe(ctx, w.handle)
}

func (w *Worker) handle(ctx context.Context, j *job.Job) error {
	log := w.logger.WithContext(ctx)
	log.Info("processing job", "job_id", j.ID, "type", j.Type)

	if err := w.jobs.UpdateStatus(ctx, j.ID, job.StatusProcessing, nil, ""); err != nil {
		log.Error("failed to update job status", slog.String("job_id", j.ID), slog.Any("error", err))
	}

	result, err := w.process(ctx, j)
	if err != nil {
		log.Error("job failed", "job_id", j.ID, "error", err)
		if updateErr := w.jobs.UpdateStatus(ctx, j.ID, job.StatusFailed, nil, err.Error()); updateErr != nil {
			log.Error("failed to update job status", "job_id", j.ID, "error", updateErr)
		}
		return err
	}

	if err := w.jobs.UpdateStatus(ctx, j.ID, job.StatusCompleted, result, ""); err != nil {
		log.Error("failed to update job status", "job_id", j.ID, "error", err)
	}

	log.Info("job completed", "job_id", j.ID)
	return nil
}

func (w *Worker) process(ctx context.Context, j *job.Job) (*job.Result, error) {
	rc, err := w.storage.Download(ctx, j.Input.StorageKey)
	if err != nil {
		return nil, fmt.Errorf("download input: %w", err)
	}
	defer rc.Close()

	img, _, err := image.Decode(rc)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	proc, err := w.registry.Get(string(j.Type))
	if err != nil {
		return nil, fmt.Errorf("unknown processor %q: %w", j.Type, err)
	}

	out, err := proc.Process(img, j.Parameters)
	if err != nil {
		return nil, fmt.Errorf("process image: %w", err)
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, out, nil); err != nil {
		return nil, fmt.Errorf("encode result: %w", err)
	}

	resultKey := "results/" + j.ID + ".jpg"
	if err := w.storage.Upload(ctx, resultKey, &buf, "image/jpeg"); err != nil {
		return nil, fmt.Errorf("upload result: %w", err)
	}

	bounds := out.Bounds()
	return &job.Result{
		StorageKey: resultKey,
		MimeType:   "image/jpeg",
		Size:       int64(buf.Len()),
		Width:      bounds.Dx(),
		Height:     bounds.Dy(),
	}, nil
}
