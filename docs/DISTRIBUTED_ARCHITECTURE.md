# Distributed Architecture Design

## Overview
Transform the monolithic image processing service into a scalable, distributed system that can handle high volumes of image processing requests across multiple worker nodes.

## Architecture Components

### 1. API Service
- **Purpose**: HTTP API gateway that accepts image processing requests
- **Responsibilities**:
  - Accept HTTP requests (resize, crop)
  - Validate input parameters
  - Create job messages
  - Publish jobs to queue
  - Return job ID to client
  - Provide job status endpoint

### 2. Job Queue (NATS JetStream)
- **Purpose**: Reliable message queue for job distribution
- **Features**:
  - Persistent message storage
  - At-least-once delivery
  - Consumer groups for worker scaling
  - Message replay capability
  - Dead letter queue for failed jobs

### 3. Worker Service
- **Purpose**: Process image jobs from the queue
- **Responsibilities**:
  - Subscribe to job queue
  - Process images (resize, crop)
  - Upload results to storage
  - Update job status
  - Handle failures with retry

### 4. Storage Service
- **Purpose**: Store processed images
- **Options**:
  - MinIO (S3-compatible, self-hosted)
  - AWS S3 (cloud)
  - Local volume (development)

### 5. Job Status Store
- **Purpose**: Track job progress and results
- **Options**:
  - Redis (fast, simple)
  - PostgreSQL (durable, queryable)

## Job Flow

```
1. Client uploads image to API
2. API creates job and publishes to queue
3. API returns job ID to client
4. Worker picks up job from queue
5. Worker processes image
6. Worker uploads result to storage
7. Worker updates job status
8. Client polls/gets notified of completion
```

## Message Format

```json
{
  "job_id": "uuid",
  "type": "resize|crop",
  "parameters": {
    "width": 400,
    "height": 300,
    // or for crop:
    "x": 100,
    "y": 100,
    "width": 200,
    "height": 200
  },
  "input": {
    "storage_key": "uploads/uuid/original.jpg"
  },
  "metadata": {
    "created_at": "2024-01-01T00:00:00Z",
    "retry_count": 0,
    "request_id": "uuid"
  }
}
```

## API Changes

### New Endpoints

#### Async Processing
```
POST /api/v1/resize
POST /api/v1/crop
Returns: { "job_id": "uuid", "status": "queued" }
```

#### Job Status
```
GET /api/v1/jobs/{job_id}
Returns: {
  "job_id": "uuid",
  "status": "queued|processing|completed|failed",
  "result": {
    "url": "https://storage/processed/uuid.jpg"
  },
  "error": "error message if failed"
}
```

#### Health Checks
```
GET /health/live
GET /health/ready
```

## Kubernetes Resources

### Deployments
1. **API Deployment**
   - 2-3 replicas
   - CPU: 100m-500m
   - Memory: 128Mi-512Mi

2. **Worker Deployment**
   - 3-10 replicas (HPA)
   - CPU: 200m-1000m
   - Memory: 256Mi-1Gi

3. **NATS Deployment**
   - StatefulSet with 3 replicas
   - Persistent volumes

4. **MinIO Deployment**
   - StatefulSet with 4 replicas
   - Persistent volumes

### Services
- API Service (LoadBalancer/Ingress)
- NATS Service (ClusterIP)
- MinIO Service (ClusterIP)

### ConfigMaps & Secrets
- App configuration
- NATS credentials
- Storage credentials

## Scaling Strategy

### Horizontal Scaling
- **API**: Scale based on CPU/memory
- **Workers**: Scale based on queue depth
- **Storage**: Scale based on capacity

### Metrics for Scaling
- Queue depth
- Processing time per job
- Worker utilization
- API response time

## Error Handling

### Retry Strategy
1. Immediate retry (1x)
2. Exponential backoff (3x)
3. Dead letter queue
4. Manual intervention

### Failure Scenarios
- Worker crash → Job returned to queue
- Storage failure → Retry with backoff
- Invalid image → Mark job as failed
- Timeout → Kill job, retry

## Monitoring & Observability

### Metrics (Prometheus)
- Job queue depth
- Processing rate
- Success/failure rate
- Processing duration
- Storage usage

### Logging
- Structured JSON logs
- Correlation IDs
- Distributed tracing

### Dashboards (Grafana)
- System overview
- Job processing pipeline
- Error rates
- Performance metrics

## Development Phases

### Phase 2.1: Core Queue System
- Implement NATS JetStream
- Create job publisher/consumer
- Basic worker implementation

### Phase 2.2: Service Separation
- Split API and worker
- Add storage service
- Implement job status

### Phase 2.3: Kubernetes Deployment
- Create manifests
- Add health checks
- Configure autoscaling

### Phase 2.4: Monitoring
- Add Prometheus metrics
- Create Grafana dashboards
- Implement alerting

## Configuration

### Environment Variables
```env
# API Service
NATS_URL=nats://nats:4222
STORAGE_TYPE=minio
STORAGE_ENDPOINT=minio:9000
REDIS_URL=redis://redis:6379

# Worker Service
WORKER_CONCURRENCY=5
MAX_RETRIES=3
JOB_TIMEOUT=30s

# Storage
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
BUCKET_NAME=processed-images
```

## Security Considerations

1. **API Security**
   - Rate limiting per IP
   - API key authentication
   - Request size limits

2. **Queue Security**
   - TLS for NATS connections
   - Authentication required

3. **Storage Security**
   - Signed URLs for uploads/downloads
   - Bucket policies
   - Encryption at rest

4. **Network Policies**
   - Restrict pod-to-pod communication
   - Egress rules for external services