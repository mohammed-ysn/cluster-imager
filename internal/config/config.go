package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server  ServerConfig
	Redis   RedisConfig
	NATS    NATSConfig
	Storage StorageConfig
	Job     JobConfig
}

type ServerConfig struct {
	Port string
}

type RedisConfig struct {
	URL string
}

type NATSConfig struct {
	URL      string
	Stream   string
	Subject  string
	Consumer string
	MaxRetry int
}

type StorageConfig struct {
	Type      string
	LocalPath string
}

type JobConfig struct {
	TTL time.Duration
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
		},
		Redis: RedisConfig{
			URL: getEnv("REDIS_URL", "redis://localhost:6379"),
		},
		NATS: NATSConfig{
			URL:      getEnv("NATS_URL", "nats://localhost:4222"),
			Stream:   getEnv("NATS_STREAM", "IMAGES"),
			Subject:  getEnv("NATS_SUBJECT", "images.process"),
			Consumer: getEnv("NATS_CONSUMER", "worker"),
			MaxRetry: getEnvInt("NATS_MAX_RETRY", 3),
		},
		Storage: StorageConfig{
			Type:      getEnv("STORAGE_TYPE", "local"),
			LocalPath: getEnv("STORAGE_LOCAL_PATH", "/tmp/cluster-imager"),
		},
		Job: JobConfig{
			TTL: getEnvDuration("JOB_TTL", 24*time.Hour),
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
