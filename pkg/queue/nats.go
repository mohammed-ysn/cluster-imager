package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mohammed-ysn/cluster-imager/pkg/job"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// NATSQueue implements Queue interface using NATS JetStream
type NATSQueue struct {
	nc       *nats.Conn
	js       jetstream.JetStream
	stream   jetstream.Stream
	consumer jetstream.Consumer
	config   Config
}

// NewNATSQueue creates a new NATS JetStream queue
func NewNATSQueue(config Config) (*NATSQueue, error) {
	// Connect to NATS
	nc, err := nats.Connect(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	// Create JetStream context
	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	// Create or get stream
	stream, err := createOrGetStream(js, config)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("failed to create/get stream: %w", err)
	}

	return &NATSQueue{
		nc:     nc,
		js:     js,
		stream: stream,
		config: config,
	}, nil
}

// createOrGetStream creates or gets an existing stream
func createOrGetStream(js jetstream.JetStream, config Config) (jetstream.Stream, error) {
	streamConfig := jetstream.StreamConfig{
		Name:     config.Stream,
		Subjects: []string{config.Subject},
		Storage:  jetstream.FileStorage,
		Replicas: 1,
		MaxAge:   24 * time.Hour,
		MaxMsgs:  -1,
		MaxBytes: -1,
	}

	// Try to get existing stream
	stream, err := js.Stream(context.Background(), config.Stream)
	if err == nil {
		return stream, nil
	}

	// Create new stream
	stream, err = js.CreateStream(context.Background(), streamConfig)
	if err != nil {
		return nil, err
	}

	return stream, nil
}

// Publish publishes a job to the queue
func (q *NATSQueue) Publish(ctx context.Context, job *job.Job) error {
	// Marshal job to JSON
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	// Publish to stream
	_, err = q.js.Publish(ctx, q.config.Subject, data)
	if err != nil {
		return fmt.Errorf("failed to publish job: %w", err)
	}

	return nil
}

// Subscribe subscribes to jobs from the queue
func (q *NATSQueue) Subscribe(ctx context.Context, handler JobHandler) error {
	// Create or get consumer
	consumer, err := q.createOrGetConsumer()
	if err != nil {
		return fmt.Errorf("failed to create/get consumer: %w", err)
	}
	q.consumer = consumer

	// Consume messages
	msgs, err := consumer.Messages()
	if err != nil {
		return fmt.Errorf("failed to get message channel: %w", err)
	}

	// Process messages
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-msgs:
				if msg == nil {
					continue
				}

				// Unmarshal job
				var j job.Job
				if err := json.Unmarshal(msg.Data(), &j); err != nil {
					// Bad message, acknowledge and continue
					msg.Ack()
					continue
				}

				// Process job
				if err := handler(ctx, &j); err != nil {
					// Handle retry logic
					if j.Metadata.RetryCount < q.config.MaxRetry {
						// Negative acknowledgment for retry
						msg.Nak()
					} else {
						// Max retries reached, acknowledge to remove from queue
						// In production, we'd send to a dead letter queue
						msg.Ack()
					}
				} else {
					// Success, acknowledge
					msg.Ack()
				}
			}
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()
	return nil
}

// createOrGetConsumer creates or gets an existing consumer
func (q *NATSQueue) createOrGetConsumer() (jetstream.Consumer, error) {
	consumerConfig := jetstream.ConsumerConfig{
		Name:          q.config.Consumer,
		Durable:       q.config.Consumer,
		AckPolicy:     jetstream.AckExplicitPolicy,
		MaxDeliver:    q.config.MaxRetry + 1,
		FilterSubject: q.config.Subject,
		AckWait:       30 * time.Second,
		MaxAckPending: 100,
	}

	// Try to get existing consumer
	consumer, err := q.stream.Consumer(context.Background(), q.config.Consumer)
	if err == nil {
		return consumer, nil
	}

	// Create new consumer
	consumer, err = q.stream.CreateConsumer(context.Background(), consumerConfig)
	if err != nil {
		return nil, err
	}

	return consumer, nil
}

// Close closes the queue connections
func (q *NATSQueue) Close() error {
	if q.nc != nil {
		q.nc.Close()
	}
	return nil
}