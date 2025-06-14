package job

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStore implements Store interface using Redis
type RedisStore struct {
	client *redis.Client
	prefix string
	ttl    time.Duration
}

// NewRedisStore creates a new Redis job store
func NewRedisStore(url string, prefix string, ttl time.Duration) (*RedisStore, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opt)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisStore{
		client: client,
		prefix: prefix,
		ttl:    ttl,
	}, nil
}

// Create creates a new job
func (s *RedisStore) Create(ctx context.Context, job *Job) error {
	key := s.key(job.ID)
	
	// Set metadata
	job.Metadata.CreatedAt = time.Now()
	job.Metadata.UpdatedAt = time.Now()

	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	if err := s.client.Set(ctx, key, data, s.ttl).Err(); err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	// Add to status index
	if err := s.addToIndex(ctx, job); err != nil {
		return fmt.Errorf("failed to add to index: %w", err)
	}

	return nil
}

// Get retrieves a job by ID
func (s *RedisStore) Get(ctx context.Context, id string) (*Job, error) {
	key := s.key(id)

	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("job not found")
		}
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	var job Job
	if err := json.Unmarshal(data, &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	return &job, nil
}

// Update updates a job
func (s *RedisStore) Update(ctx context.Context, job *Job) error {
	// Remove from old status index
	oldJob, err := s.Get(ctx, job.ID)
	if err == nil {
		s.removeFromIndex(ctx, oldJob)
	}

	// Update timestamp
	job.Metadata.UpdatedAt = time.Now()

	// Save job
	key := s.key(job.ID)
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	if err := s.client.Set(ctx, key, data, s.ttl).Err(); err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	// Add to new status index
	if err := s.addToIndex(ctx, job); err != nil {
		return fmt.Errorf("failed to update index: %w", err)
	}

	return nil
}

// UpdateStatus updates the status of a job
func (s *RedisStore) UpdateStatus(ctx context.Context, id string, status Status, result *Result, errMsg string) error {
	job, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	job.Status = status
	job.Result = result
	job.Error = errMsg

	// Update timestamps
	if status == StatusProcessing {
		job.Metadata.StartedAt = time.Now()
	} else if status == StatusCompleted || status == StatusFailed {
		job.Metadata.CompletedAt = time.Now()
	}

	return s.Update(ctx, job)
}

// List lists jobs with optional filters
func (s *RedisStore) List(ctx context.Context, filter Filter) ([]*Job, error) {
	// For simplicity, we'll use status-based indices
	// In production, consider using Redis sorted sets for time-based queries
	
	var keys []string
	
	if filter.Status != "" {
		indexKey := s.indexKey(string(filter.Status))
		jobIDs, err := s.client.SMembers(ctx, indexKey).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to get job IDs from index: %w", err)
		}
		
		for _, id := range jobIDs {
			keys = append(keys, s.key(id))
		}
	} else {
		// Get all jobs (expensive operation, use with caution)
		pattern := fmt.Sprintf("%s:*", s.prefix)
		iter := s.client.Scan(ctx, 0, pattern, 0).Iterator()
		for iter.Next(ctx) {
			keys = append(keys, iter.Val())
		}
		if err := iter.Err(); err != nil {
			return nil, fmt.Errorf("failed to scan keys: %w", err)
		}
	}

	// Apply limit
	if filter.Limit > 0 && len(keys) > filter.Limit {
		keys = keys[filter.Offset : filter.Offset+filter.Limit]
	}

	// Get jobs
	var jobs []*Job
	for _, key := range keys {
		data, err := s.client.Get(ctx, key).Bytes()
		if err != nil {
			continue // Skip missing jobs
		}

		var job Job
		if err := json.Unmarshal(data, &job); err != nil {
			continue // Skip invalid jobs
		}

		// Apply time filters
		if !filter.Since.IsZero() && job.Metadata.CreatedAt.Before(filter.Since) {
			continue
		}
		if !filter.Until.IsZero() && job.Metadata.CreatedAt.After(filter.Until) {
			continue
		}

		jobs = append(jobs, &job)
	}

	return jobs, nil
}

// Delete deletes a job
func (s *RedisStore) Delete(ctx context.Context, id string) error {
	job, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	// Remove from index
	if err := s.removeFromIndex(ctx, job); err != nil {
		return fmt.Errorf("failed to remove from index: %w", err)
	}

	// Delete job
	key := s.key(id)
	if err := s.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete job: %w", err)
	}

	return nil
}

// Helper methods

func (s *RedisStore) key(id string) string {
	return fmt.Sprintf("%s:%s", s.prefix, id)
}

func (s *RedisStore) indexKey(status string) string {
	return fmt.Sprintf("%s:status:%s", s.prefix, status)
}

func (s *RedisStore) addToIndex(ctx context.Context, job *Job) error {
	indexKey := s.indexKey(string(job.Status))
	return s.client.SAdd(ctx, indexKey, job.ID).Err()
}

func (s *RedisStore) removeFromIndex(ctx context.Context, job *Job) error {
	indexKey := s.indexKey(string(job.Status))
	return s.client.SRem(ctx, indexKey, job.ID).Err()
}

// Close closes the Redis connection
func (s *RedisStore) Close() error {
	return s.client.Close()
}