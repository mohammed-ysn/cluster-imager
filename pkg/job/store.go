package job

import (
	"context"
	"time"
)

// Store defines the interface for job storage
type Store interface {
	// Create creates a new job
	Create(ctx context.Context, job *Job) error
	
	// Get retrieves a job by ID
	Get(ctx context.Context, id string) (*Job, error)
	
	// Update updates a job
	Update(ctx context.Context, job *Job) error
	
	// UpdateStatus updates the status of a job
	UpdateStatus(ctx context.Context, id string, status Status, result *Result, errMsg string) error
	
	// List lists jobs with optional filters
	List(ctx context.Context, filter Filter) ([]*Job, error)
	
	// Delete deletes a job
	Delete(ctx context.Context, id string) error
}

// Filter represents job listing filters
type Filter struct {
	Status    Status
	Type      Type
	Since     time.Time
	Until     time.Time
	Limit     int
	Offset    int
}