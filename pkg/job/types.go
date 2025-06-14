package job

import (
	"time"
)

// Status represents the status of a job
type Status string

const (
	StatusQueued     Status = "queued"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
)

// Type represents the type of image processing job
type Type string

const (
	TypeResize Type = "resize"
	TypeCrop   Type = "crop"
)

// Job represents an image processing job
type Job struct {
	ID         string                 `json:"job_id"`
	Type       Type                   `json:"type"`
	Status     Status                 `json:"status"`
	Parameters map[string]interface{} `json:"parameters"`
	Input      Input                  `json:"input"`
	Result     *Result                `json:"result,omitempty"`
	Error      string                 `json:"error,omitempty"`
	Metadata   Metadata               `json:"metadata"`
}

// Input represents the input for a job
type Input struct {
	StorageKey string `json:"storage_key"`
	MimeType   string `json:"mime_type"`
	Size       int64  `json:"size"`
}

// Result represents the result of a completed job
type Result struct {
	StorageKey string `json:"storage_key"`
	URL        string `json:"url"`
	MimeType   string `json:"mime_type"`
	Size       int64  `json:"size"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
}

// Metadata contains job metadata
type Metadata struct {
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	StartedAt   time.Time `json:"started_at,omitempty"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
	RetryCount  int       `json:"retry_count"`
	RequestID   string    `json:"request_id"`
}

// ResizeParams represents parameters for resize operation
type ResizeParams struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// CropParams represents parameters for crop operation
type CropParams struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}