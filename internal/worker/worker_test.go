package worker

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/jpeg"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/mohammed-ysn/cluster-imager/internal/processors"
	"github.com/mohammed-ysn/cluster-imager/pkg/job"
	"github.com/mohammed-ysn/cluster-imager/pkg/logging"
)

type mockJobStore struct {
	jobs    map[string]*job.Job
	updates []job.Status
	err     error
}

func newMockJobStore(j *job.Job) *mockJobStore {
	m := &mockJobStore{jobs: make(map[string]*job.Job)}
	if j != nil {
		m.jobs[j.ID] = j
	}
	return m
}

func (m *mockJobStore) Create(_ context.Context, j *job.Job) error   { return m.err }
func (m *mockJobStore) Get(_ context.Context, id string) (*job.Job, error) {
	if m.err != nil {
		return nil, m.err
	}
	j, ok := m.jobs[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return j, nil
}
func (m *mockJobStore) Update(_ context.Context, j *job.Job) error { return m.err }
func (m *mockJobStore) UpdateStatus(_ context.Context, id string, status job.Status, _ *job.Result, _ string) error {
	if m.err != nil {
		return m.err
	}
	m.updates = append(m.updates, status)
	if j, ok := m.jobs[id]; ok {
		j.Status = status
	}
	return nil
}
func (m *mockJobStore) List(_ context.Context, _ job.Filter) ([]*job.Job, error) { return nil, m.err }
func (m *mockJobStore) Delete(_ context.Context, _ string) error                  { return m.err }

type mockStorage struct {
	data map[string][]byte
	err  error
}

func newMockStorage() *mockStorage {
	return &mockStorage{data: make(map[string][]byte)}
}

func (m *mockStorage) Upload(_ context.Context, key string, r io.Reader, _ string) error {
	if m.err != nil {
		return m.err
	}
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	m.data[key] = b
	return nil
}

func (m *mockStorage) Download(_ context.Context, key string) (io.ReadCloser, error) {
	if m.err != nil {
		return nil, m.err
	}
	b, ok := m.data[key]
	if !ok {
		return nil, errors.New("not found: " + key)
	}
	return io.NopCloser(bytes.NewReader(b)), nil
}

func (m *mockStorage) Delete(_ context.Context, _ string) error { return m.err }
func (m *mockStorage) GetURL(_ context.Context, _ string, _ time.Duration) (string, error) {
	return "", m.err
}
func (m *mockStorage) Exists(_ context.Context, _ string) (bool, error) { return false, m.err }

func minimalJPEG(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func newWorker(jobs *mockJobStore, stor *mockStorage) *Worker {
	logger := logging.NewLogger(slog.LevelError)
	return New(nil, jobs, stor, processors.DefaultRegistry(), logger)
}

func TestHandle_ResizeSuccess(t *testing.T) {
	stor := newMockStorage()
	stor.data["inputs/job1"] = minimalJPEG(t, 100, 100)

	j := &job.Job{
		ID:     "job1",
		Type:   job.TypeResize,
		Status: job.StatusQueued,
		Input:  job.Input{StorageKey: "inputs/job1"},
		Parameters: map[string]any{
			"width": 50, "height": 50,
		},
	}
	jobs := newMockJobStore(j)
	w := newWorker(jobs, stor)

	if err := w.handle(context.Background(), j); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if j.Status != job.StatusCompleted {
		t.Errorf("expected status completed, got %s", j.Status)
	}
	if _, ok := stor.data["results/job1.jpg"]; !ok {
		t.Error("expected result file in storage")
	}
}

func TestHandle_CropSuccess(t *testing.T) {
	stor := newMockStorage()
	stor.data["inputs/job2"] = minimalJPEG(t, 100, 100)

	j := &job.Job{
		ID:     "job2",
		Type:   job.TypeCrop,
		Status: job.StatusQueued,
		Input:  job.Input{StorageKey: "inputs/job2"},
		Parameters: map[string]any{
			"x": 0, "y": 0, "width": 50, "height": 50,
		},
	}
	jobs := newMockJobStore(j)
	w := newWorker(jobs, stor)

	if err := w.handle(context.Background(), j); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if j.Status != job.StatusCompleted {
		t.Errorf("expected status completed, got %s", j.Status)
	}
}

func TestHandle_StorageDownloadFails(t *testing.T) {
	stor := newMockStorage()
	stor.err = errors.New("storage down")

	j := &job.Job{
		ID:    "job3",
		Type:  job.TypeResize,
		Input: job.Input{StorageKey: "inputs/job3"},
		Parameters: map[string]any{"width": 50, "height": 50},
	}
	jobs := newMockJobStore(j)
	w := newWorker(jobs, stor)

	if err := w.handle(context.Background(), j); err == nil {
		t.Error("expected error, got nil")
	}
	if j.Status != job.StatusFailed {
		t.Errorf("expected status failed, got %s", j.Status)
	}
}

func TestHandle_UnknownProcessor(t *testing.T) {
	stor := newMockStorage()
	stor.data["inputs/job4"] = minimalJPEG(t, 10, 10)

	j := &job.Job{
		ID:         "job4",
		Type:       job.Type("unknown"),
		Input:      job.Input{StorageKey: "inputs/job4"},
		Parameters: map[string]any{},
	}
	jobs := newMockJobStore(j)
	w := newWorker(jobs, stor)

	if err := w.handle(context.Background(), j); err == nil {
		t.Error("expected error for unknown processor")
	}
	if j.Status != job.StatusFailed {
		t.Errorf("expected status failed, got %s", j.Status)
	}
}
