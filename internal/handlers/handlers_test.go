package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mohammed-ysn/cluster-imager/internal/processors"
	"github.com/mohammed-ysn/cluster-imager/pkg/job"
	"github.com/mohammed-ysn/cluster-imager/pkg/logging"
	"log/slog"
)

type mockJobStore struct {
	created []*job.Job
	jobs    map[string]*job.Job
	err     error
}

func newMockJobStore() *mockJobStore {
	return &mockJobStore{jobs: make(map[string]*job.Job)}
}

func (m *mockJobStore) Create(_ context.Context, j *job.Job) error {
	if m.err != nil {
		return m.err
	}
	m.created = append(m.created, j)
	m.jobs[j.ID] = j
	return nil
}

func (m *mockJobStore) Get(_ context.Context, id string) (*job.Job, error) {
	if m.err != nil {
		return nil, m.err
	}
	j, ok := m.jobs[id]
	if !ok {
		return nil, errors.New("job not found")
	}
	return j, nil
}

func (m *mockJobStore) Update(_ context.Context, j *job.Job) error          { return m.err }
func (m *mockJobStore) UpdateStatus(_ context.Context, _ string, _ job.Status, _ *job.Result, _ string) error {
	return m.err
}
func (m *mockJobStore) List(_ context.Context, _ job.Filter) ([]*job.Job, error) {
	return nil, m.err
}
func (m *mockJobStore) Delete(_ context.Context, _ string) error { return m.err }

type mockStorage struct {
	err error
}

func (m *mockStorage) Upload(_ context.Context, _ string, _ io.Reader, _ string) error {
	return m.err
}
func (m *mockStorage) Download(_ context.Context, _ string) (io.ReadCloser, error) {
	return nil, m.err
}
func (m *mockStorage) Delete(_ context.Context, _ string) error { return m.err }
func (m *mockStorage) GetURL(_ context.Context, _ string, _ time.Duration) (string, error) {
	return "", m.err
}
func (m *mockStorage) Exists(_ context.Context, _ string) (bool, error) { return false, m.err }

type mockQueue struct {
	published []*job.Job
	err       error
}

func (m *mockQueue) Publish(_ context.Context, j *job.Job) error {
	if m.err != nil {
		return m.err
	}
	m.published = append(m.published, j)
	return nil
}

func (m *mockQueue) Close() error { return nil }

func newHandlers(jobs *mockJobStore, stor *mockStorage, q *mockQueue) *Handlers {
	logger := logging.NewLogger(slog.LevelError)
	registry := processors.DefaultRegistry()
	return New(logger, registry, jobs, stor, q)
}

func multipartImageRequest(t *testing.T, url string) *http.Request {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	fw, err := w.CreateFormFile("image", "test.jpg")
	if err != nil {
		t.Fatal(err)
	}
	fw.Write([]byte{
		0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 0x4a, 0x46, 0x49, 0x46, 0x00, 0x01,
		0x01, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0xff, 0xdb, 0x00, 0x43,
		0x00, 0x08, 0x06, 0x06, 0x07, 0x06, 0x05, 0x08, 0x07, 0x07, 0x07, 0x09,
		0x09, 0x08, 0x0a, 0x0c, 0x14, 0x0d, 0x0c, 0x0b, 0x0b, 0x0c, 0x19, 0x12,
		0x13, 0x0f, 0x14, 0x1d, 0x1a, 0x1f, 0x1e, 0x1d, 0x1a, 0x1c, 0x1c, 0x20,
		0x24, 0x2e, 0x27, 0x20, 0x22, 0x2c, 0x23, 0x1c, 0x1c, 0x28, 0x37, 0x29,
		0x2c, 0x30, 0x31, 0x34, 0x34, 0x34, 0x1f, 0x27, 0x39, 0x3d, 0x38, 0x32,
		0x3c, 0x2e, 0x33, 0x34, 0x32, 0xff, 0xc0, 0x00, 0x0b, 0x08, 0x00, 0x01,
		0x00, 0x01, 0x01, 0x01, 0x11, 0x00, 0xff, 0xc4, 0x00, 0x1f, 0x00, 0x00,
		0x01, 0x05, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0xff, 0xc4, 0x00, 0xb5, 0x10, 0x00, 0x02, 0x01, 0x03,
		0x03, 0x02, 0x04, 0x03, 0x05, 0x05, 0x04, 0x04, 0x00, 0x00, 0x01, 0x7d,
		0x01, 0x02, 0x03, 0x00, 0x04, 0x11, 0x05, 0x12, 0x21, 0x31, 0x41, 0x06,
		0x13, 0x51, 0x61, 0x07, 0x22, 0x71, 0x14, 0x32, 0x81, 0x91, 0xa1, 0x08,
		0x23, 0x42, 0xb1, 0xc1, 0x15, 0x52, 0xd1, 0xf0, 0x24, 0x33, 0x62, 0x72,
		0x82, 0x09, 0x0a, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x25, 0x26, 0x27, 0x28,
		0x29, 0x2a, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3a, 0x43, 0x44, 0x45,
		0x46, 0x47, 0x48, 0x49, 0x4a, 0x53, 0x54, 0x55, 0x56, 0x57, 0x58, 0x59,
		0x5a, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x6a, 0x73, 0x74, 0x75,
		0x76, 0x77, 0x78, 0x79, 0x7a, 0x83, 0x84, 0x85, 0x86, 0x87, 0x88, 0x89,
		0x8a, 0x93, 0x94, 0x95, 0x96, 0x97, 0x98, 0x99, 0x9a, 0xa2, 0xa3, 0xa4,
		0xa5, 0xa6, 0xa7, 0xa8, 0xa9, 0xaa, 0xb2, 0xb3, 0xb4, 0xb5, 0xb6, 0xb7,
		0xb8, 0xb9, 0xba, 0xc2, 0xc3, 0xc4, 0xc5, 0xc6, 0xc7, 0xc8, 0xc9, 0xca,
		0xd2, 0xd3, 0xd4, 0xd5, 0xd6, 0xd7, 0xd8, 0xd9, 0xda, 0xe1, 0xe2, 0xe3,
		0xe4, 0xe5, 0xe6, 0xe7, 0xe8, 0xe9, 0xea, 0xf1, 0xf2, 0xf3, 0xf4, 0xf5,
		0xf6, 0xf7, 0xf8, 0xf9, 0xfa, 0xff, 0xda, 0x00, 0x08, 0x01, 0x01, 0x00,
		0x00, 0x3f, 0x00, 0xfb, 0xd6, 0xff, 0xd9,
	})
	w.Close()

	req := httptest.NewRequest(http.MethodPost, url, &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

func TestLiveHandler(t *testing.T) {
	h := newHandlers(newMockJobStore(), &mockStorage{}, &mockQueue{})
	rr := httptest.NewRecorder()
	h.LiveHandler(rr, httptest.NewRequest(http.MethodGet, "/health/live", nil))
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestReadyHandler(t *testing.T) {
	h := newHandlers(newMockJobStore(), &mockStorage{}, &mockQueue{})
	rr := httptest.NewRecorder()
	h.ReadyHandler(rr, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestCropHandler_Success(t *testing.T) {
	jobs := newMockJobStore()
	q := &mockQueue{}
	h := newHandlers(jobs, &mockStorage{}, q)

	req := multipartImageRequest(t, "/api/v1/crop?x=0&y=0&width=1&height=1")
	rr := httptest.NewRecorder()
	h.CropHandler(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d: %s", rr.Code, rr.Body.String())
	}
	if len(jobs.created) != 1 {
		t.Errorf("expected 1 job created, got %d", len(jobs.created))
	}
	if len(q.published) != 1 {
		t.Errorf("expected 1 job published, got %d", len(q.published))
	}
	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp["job_id"] == "" {
		t.Error("expected job_id in response")
	}
}

func TestCropHandler_WrongMethod(t *testing.T) {
	h := newHandlers(newMockJobStore(), &mockStorage{}, &mockQueue{})
	rr := httptest.NewRecorder()
	h.CropHandler(rr, httptest.NewRequest(http.MethodGet, "/api/v1/crop?x=0&y=0&width=1&height=1", nil))
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

func TestCropHandler_MissingParams(t *testing.T) {
	h := newHandlers(newMockJobStore(), &mockStorage{}, &mockQueue{})
	req := multipartImageRequest(t, "/api/v1/crop?x=0&y=0&width=1")
	rr := httptest.NewRecorder()
	h.CropHandler(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestResizeHandler_Success(t *testing.T) {
	jobs := newMockJobStore()
	q := &mockQueue{}
	h := newHandlers(jobs, &mockStorage{}, q)

	req := multipartImageRequest(t, "/api/v1/resize?width=10&height=10")
	rr := httptest.NewRecorder()
	h.ResizeHandler(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d: %s", rr.Code, rr.Body.String())
	}
	if len(jobs.created) != 1 {
		t.Errorf("expected 1 job created, got %d", len(jobs.created))
	}
}

func TestResizeHandler_WrongMethod(t *testing.T) {
	h := newHandlers(newMockJobStore(), &mockStorage{}, &mockQueue{})
	rr := httptest.NewRecorder()
	h.ResizeHandler(rr, httptest.NewRequest(http.MethodGet, "/api/v1/resize?width=10&height=10", nil))
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

func TestResizeHandler_MissingParams(t *testing.T) {
	h := newHandlers(newMockJobStore(), &mockStorage{}, &mockQueue{})
	req := multipartImageRequest(t, "/api/v1/resize?width=10")
	rr := httptest.NewRecorder()
	h.ResizeHandler(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestJobStatusHandler_Found(t *testing.T) {
	jobs := newMockJobStore()
	jobs.jobs["abc123"] = &job.Job{ID: "abc123", Status: job.StatusQueued}
	h := newHandlers(jobs, &mockStorage{}, &mockQueue{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/jobs/abc123", nil)
	req.SetPathValue("id", "abc123")
	rr := httptest.NewRecorder()
	h.JobStatusHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	var j job.Job
	if err := json.NewDecoder(rr.Body).Decode(&j); err != nil {
		t.Fatal(err)
	}
	if j.ID != "abc123" {
		t.Errorf("expected job_id abc123, got %s", j.ID)
	}
}

func TestJobStatusHandler_NotFound(t *testing.T) {
	h := newHandlers(newMockJobStore(), &mockStorage{}, &mockQueue{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/jobs/nope", nil)
	req.SetPathValue("id", "nope")
	rr := httptest.NewRecorder()
	h.JobStatusHandler(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestJobStatusHandler_MissingID(t *testing.T) {
	h := newHandlers(newMockJobStore(), &mockStorage{}, &mockQueue{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/jobs/", nil)
	rr := httptest.NewRecorder()
	h.JobStatusHandler(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}
