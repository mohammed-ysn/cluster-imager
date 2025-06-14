package logging

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
)

// ContextKey is a type for context keys
type ContextKey string

const (
	// RequestIDKey is the context key for request ID
	RequestIDKey ContextKey = "request_id"
)

// Logger wraps slog.Logger with additional functionality
type Logger struct {
	*slog.Logger
}

// NewLogger creates a new structured logger
func NewLogger(level slog.Level) *Logger {
	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Customize time format
			if a.Key == slog.TimeKey {
				if t, ok := a.Value.Any().(time.Time); ok {
					a.Value = slog.StringValue(t.Format(time.RFC3339))
				}
			}
			return a
		},
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)

	return &Logger{Logger: logger}
}

// WithRequestID adds a request ID to the logger context
func (l *Logger) WithRequestID(ctx context.Context) *Logger {
	requestID := GetRequestID(ctx)
	return &Logger{
		Logger: l.With("request_id", requestID),
	}
}

// WithContext returns a logger with context values
func (l *Logger) WithContext(ctx context.Context) *Logger {
	return l.WithRequestID(ctx)
}

// GenerateRequestID generates a new request ID
func GenerateRequestID() string {
	return uuid.New().String()
}

// GetRequestID retrieves the request ID from context
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return "unknown"
}

// WithRequestID adds a request ID to the context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}