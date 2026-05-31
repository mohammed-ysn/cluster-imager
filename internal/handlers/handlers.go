package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/mohammed-ysn/cluster-imager/internal/processors"
	"github.com/mohammed-ysn/cluster-imager/pkg/job"
	"github.com/mohammed-ysn/cluster-imager/pkg/logging"
	"github.com/mohammed-ysn/cluster-imager/pkg/queue"
	"github.com/mohammed-ysn/cluster-imager/pkg/storage"
)

type Handlers struct {
	logger   *logging.Logger
	registry *processors.Registry
	jobs     job.Store
	storage  storage.Storage
	queue    queue.Publisher
}

func New(logger *logging.Logger, registry *processors.Registry, jobs job.Store, stor storage.Storage, q queue.Publisher) *Handlers {
	return &Handlers{
		logger:   logger,
		registry: registry,
		jobs:     jobs,
		storage:  stor,
		queue:    q,
	}
}

func (h *Handlers) LiveHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handlers) ReadyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handlers) CropHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()
	x, err := strconv.Atoi(q.Get("x"))
	if err != nil {
		http.Error(w, "invalid value for 'x'", http.StatusBadRequest)
		return
	}
	y, err := strconv.Atoi(q.Get("y"))
	if err != nil {
		http.Error(w, "invalid value for 'y'", http.StatusBadRequest)
		return
	}
	width, err := strconv.Atoi(q.Get("width"))
	if err != nil {
		http.Error(w, "invalid value for 'width'", http.StatusBadRequest)
		return
	}
	height, err := strconv.Atoi(q.Get("height"))
	if err != nil {
		http.Error(w, "invalid value for 'height'", http.StatusBadRequest)
		return
	}

	h.enqueue(w, r, job.TypeCrop, map[string]any{
		"x": x, "y": y, "width": width, "height": height,
	})
}

func (h *Handlers) ResizeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()
	width, err := strconv.Atoi(q.Get("width"))
	if err != nil {
		http.Error(w, "invalid value for 'width'", http.StatusBadRequest)
		return
	}
	height, err := strconv.Atoi(q.Get("height"))
	if err != nil {
		http.Error(w, "invalid value for 'height'", http.StatusBadRequest)
		return
	}

	h.enqueue(w, r, job.TypeResize, map[string]any{
		"width": width, "height": height,
	})
}

func (h *Handlers) JobStatusHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing job id", http.StatusBadRequest)
		return
	}

	j, err := h.jobs.Get(r.Context(), id)
	if err != nil {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(j)
}

func (h *Handlers) enqueue(w http.ResponseWriter, r *http.Request, jobType job.Type, params map[string]any) {
	logger := h.logger.WithContext(r.Context())

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "no image file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	jobID := uuid.New().String()
	storageKey := "inputs/" + jobID

	if err := h.storage.Upload(r.Context(), storageKey, file, header.Header.Get("Content-Type")); err != nil {
		logger.Error("failed to upload image", "error", err)
		http.Error(w, "failed to store image", http.StatusInternalServerError)
		return
	}

	j := &job.Job{
		ID:         jobID,
		Type:       jobType,
		Status:     job.StatusQueued,
		Parameters: params,
		Input: job.Input{
			StorageKey: storageKey,
			MimeType:   header.Header.Get("Content-Type"),
			Size:       header.Size,
		},
		Metadata: job.Metadata{
			RequestID: logging.GetRequestID(r.Context()),
		},
	}

	if err := h.jobs.Create(r.Context(), j); err != nil {
		logger.Error("failed to create job", "error", err)
		http.Error(w, "failed to create job", http.StatusInternalServerError)
		return
	}

	if err := h.queue.Publish(r.Context(), j); err != nil {
		logger.Error("failed to publish job", "error", err)
		http.Error(w, "failed to queue job", http.StatusInternalServerError)
		return
	}

	logger.Info("job queued", "job_id", jobID, "type", jobType)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"job_id": jobID})
}
