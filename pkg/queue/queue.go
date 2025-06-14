package queue

import (
	"context"

	"github.com/mohammed-ysn/cluster-imager/pkg/job"
)

// Publisher publishes jobs to the queue
type Publisher interface {
	// Publish publishes a job to the queue
	Publish(ctx context.Context, job *job.Job) error
	
	// Close closes the publisher connection
	Close() error
}

// Consumer consumes jobs from the queue
type Consumer interface {
	// Subscribe subscribes to jobs from the queue
	// The handler function is called for each job
	Subscribe(ctx context.Context, handler JobHandler) error
	
	// Close closes the consumer connection
	Close() error
}

// JobHandler processes a job
type JobHandler func(ctx context.Context, job *job.Job) error

// Queue combines publisher and consumer interfaces
type Queue interface {
	Publisher
	Consumer
}

// Config represents queue configuration
type Config struct {
	URL      string
	Stream   string
	Subject  string
	Consumer string
	MaxRetry int
}