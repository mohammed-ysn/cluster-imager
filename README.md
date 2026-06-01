# cluster-imager

Distributed image processing system built with Go and Kubernetes. Accepts image upload requests via HTTP, queues jobs through NATS JetStream, processes them asynchronously in a worker pool, and stores results on disk (swappable for S3/MinIO).

## Architecture

```
HTTP API  ->  NATS JetStream  ->  Worker
   |                                 |
Redis (job store)          Local/S3 Storage
```

- **API** - accepts requests, stores input, creates job record, publishes to queue, returns job ID
- **Worker** - consumes queue, processes image, writes result, updates job status
- **Redis** - job status and metadata
- **NATS JetStream** - durable message queue with retry

## Running locally

Requires Docker.

```bash
docker compose up --build
```

API available at `http://localhost:8080`.

## API

### Resize

```
POST /api/v1/resize?width=<int>&height=<int>
Content-Type: multipart/form-data
Body: image=<file>
```

```bash
curl -X POST "http://localhost:8080/api/v1/resize?width=200&height=200" \
  -F "image=@photo.jpg"
```

```json
{"job_id": "3f2a1b4c-..."}
```

### Crop

```
POST /api/v1/crop?x=<int>&y=<int>&width=<int>&height=<int>
Content-Type: multipart/form-data
Body: image=<file>
```

```bash
curl -X POST "http://localhost:8080/api/v1/crop?x=0&y=0&width=100&height=100" \
  -F "image=@photo.jpg"
```

### Job status

```
GET /api/v1/jobs/{job_id}
```

```bash
curl http://localhost:8080/api/v1/jobs/3f2a1b4c-...
```

```json
{
  "job_id": "3f2a1b4c-...",
  "type": "resize",
  "status": "completed",
  "result": {
    "storage_key": "results/3f2a1b4c-....jpg",
    "mime_type": "image/jpeg",
    "size": 4120,
    "width": 200,
    "height": 200
  }
}
```

Job status values: `queued` -> `processing` -> `completed` or `failed`

### Health

```
GET /health/live
GET /health/ready
```

## Configuration

Configured via environment variables. Copy `.env.example` to `.env` and adjust as needed. See [`internal/config/config.go`](internal/config/config.go) for all variables and their defaults.

## Testing

```bash
# Unit tests
go test ./...

# Integration tests (requires running stack)
go test -tags integration ./internal/integration/
```

## Kubernetes

Manifests in `k8s/`. Update image references and secret values before applying.

```bash
kubectl apply -f k8s/
```
