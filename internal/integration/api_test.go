//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"mime/multipart"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/mohammed-ysn/cluster-imager/pkg/job"
)

var baseURL = func() string {
	if u := os.Getenv("API_URL"); u != "" {
		return u
	}
	return "http://localhost:8080"
}()

func TestHealthLive(t *testing.T) {
	resp, err := http.Get(baseURL + "/health/live")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestHealthReady(t *testing.T) {
	resp, err := http.Get(baseURL + "/health/ready")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestResizeJob_Completed(t *testing.T) {
	jobID := submitImage(t, "/api/v1/resize?width=50&height=50", 100, 100)
	j := pollJob(t, jobID, 5*time.Second)

	if j.Status != job.StatusCompleted {
		t.Errorf("expected completed, got %s (error: %s)", j.Status, j.Error)
	}
	if j.Result == nil {
		t.Fatal("expected result, got nil")
	}
	if j.Result.Width != 50 || j.Result.Height != 50 {
		t.Errorf("expected 50x50, got %dx%d", j.Result.Width, j.Result.Height)
	}
}

func TestCropJob_Completed(t *testing.T) {
	jobID := submitImage(t, "/api/v1/crop?x=0&y=0&width=40&height=40", 100, 100)
	j := pollJob(t, jobID, 5*time.Second)

	if j.Status != job.StatusCompleted {
		t.Errorf("expected completed, got %s (error: %s)", j.Status, j.Error)
	}
	if j.Result == nil {
		t.Fatal("expected result, got nil")
	}
	if j.Result.Width != 40 || j.Result.Height != 40 {
		t.Errorf("expected 40x40, got %dx%d", j.Result.Width, j.Result.Height)
	}
}

func TestResizeJob_InvalidImage(t *testing.T) {
	jobID := submitRaw(t, "/api/v1/resize?width=50&height=50", []byte("not an image"))
	j := pollJob(t, jobID, 5*time.Second)

	if j.Status != job.StatusFailed {
		t.Errorf("expected failed, got %s", j.Status)
	}
	if j.Error == "" {
		t.Error("expected error message, got empty")
	}
}

func TestJobStatus_NotFound(t *testing.T) {
	resp, err := http.Get(baseURL + "/api/v1/jobs/nonexistent-id")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestResizeJob_MissingParam(t *testing.T) {
	body, ct := buildMultipart(t, makeJPEG(t, 100, 100))
	resp, err := http.Post(baseURL+"/api/v1/resize?width=50", ct, body)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func submitImage(t *testing.T, path string, w, h int) string {
	t.Helper()
	return submitRaw(t, path, makeJPEG(t, w, h))
}

func submitRaw(t *testing.T, path string, data []byte) string {
	t.Helper()
	body, ct := buildMultipart(t, data)
	resp, err := http.Post(baseURL+path, ct, body)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", resp.StatusCode)
	}
	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	jobID := result["job_id"]
	if jobID == "" {
		t.Fatal("empty job_id in response")
	}
	return jobID
}

func pollJob(t *testing.T, jobID string, timeout time.Duration) *job.Job {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(fmt.Sprintf("%s/api/v1/jobs/%s", baseURL, jobID))
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		var j job.Job
		if err := json.NewDecoder(resp.Body).Decode(&j); err != nil {
			resp.Body.Close()
			t.Fatalf("failed to decode job: %v", err)
		}
		resp.Body.Close()
		if j.Status == job.StatusCompleted || j.Status == job.StatusFailed {
			return &j
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatalf("job %s did not finish within %s", jobID, timeout)
	return nil
}

func buildMultipart(t *testing.T, data []byte) (*bytes.Buffer, string) {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	fw, err := w.CreateFormFile("image", "test.jpg")
	if err != nil {
		t.Fatal(err)
	}
	fw.Write(data)
	w.Close()
	return &body, w.FormDataContentType()
}

func makeJPEG(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
