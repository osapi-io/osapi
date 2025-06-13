---
sidebar_position: 7
---

# Job System Architecture

**Date:** June 2025 **Status:** Implemented **Author:** @retr0h

## Overview

The OSAPI Job System implements a **KV-first, stream-notification architecture**
using NATS JetStream for distributed job processing. This system provides
asynchronous operation execution with persistent job state, intelligent worker
routing, and comprehensive job lifecycle management.

## Architecture Principles

- **KV-First Storage**: Job state lives in NATS KV for persistence and direct
  access
- **Stream Notifications**: Workers receive job notifications via JetStream
  subjects
- **Hierarchical Routing**: Operations use dot-notation for intelligent worker
  targeting
- **REST-Compatible**: Supports standard HTTP polling patterns for API
  integration
- **CLI Management**: Direct job queue inspection and management tools

## System Components

### Core Components

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   REST API      │    │   Jobs CLI      │    │   Job Workers   │
│                 │    │                 │    │                 │
│ • Create Jobs   │    │ • Add Jobs      │    │ • Process Jobs  │
│ • Query Status  │    │ • List Jobs     │    │ • Update Status │
│ • Return Results│    │ • Get Details   │    │ • Store Results │
└─────────────────┘    │ • Status View   │    └─────────────────┘
         │              └─────────────────┘                │
         │                       │                         │
         └───────────────────────┼─────────────────────────┘
                                 │
                                 v
                    ┌─────────────────────────┐
                    │     Job Client Layer    │
                    │                         │
                    │ • CreateJob()           │ <--- Business Logic
                    │ • GetQueueStats()       │ <--- Abstraction
                    │ • GetJobStatus()        │ <--- Type Safety
                    │ • ListJobs()            │
                    └─────────────────────────┘
                                 │
                                 v
                    ┌─────────────────────────┐
                    │     NATS JetStream      │
                    │                         │
                    │  KV Store (job-queue)   │ <--- Job Persistence
                    │  Stream (JOBS)          │ <--- Worker Notifications
                    │  KV Store (job-results) │ <--- Result Storage
                    └─────────────────────────┘
```

### Job Flow Diagram

```
1. Job Creation
   API/CLI → Job Client → KV Store → Stream Notification

2. Job Processing
   Worker ← Stream Notification
   Worker → Get Job from KV
   Worker → Update Status in KV
   Worker → Process Operation
   Worker → Store Result in KV

3. Status Query
   API/CLI → Job Client → Read from KV
```

## NATS Configuration

### KV Buckets

1. **job-queue**: Primary job storage

   - Key format: `{status}.{uuid}`
   - Status prefixes: `unprocessed`, `processing`, `completed`, `failed`
   - TTL: 24 hours for completed/failed jobs
   - History: 5 versions

2. **job-responses**: Result storage
   - Key format: `{sanitized_request_id}`
   - TTL: 24 hours
   - Used for worker-to-client result passing

### JetStream Configuration

```yaml
Stream: JOBS
Subjects:
  - jobs.query.> # Read operations
  - jobs.modify.> # Write operations

Consumer: jobs-worker
Durable: true
AckPolicy: Explicit
MaxDeliver: 3
AckWait: 30s
```

## Subject Hierarchy

The system uses hierarchical subjects for intelligent routing:

```
jobs.{type}.{hostname}.{category}.{operation}

Examples:
- jobs.query._any.system.hostname.get
- jobs.query.server1.network.dns.get
- jobs.modify._all.network.dns.update
- jobs.modify._any.network.ping.do
```

### Semantic Routing Rules

Operations are automatically routed based on their suffix:

- **Query operations** (read-only):

  - `.get` - Retrieve current state
  - `.query` - Query information
  - `.read` - Read configuration
  - `.status` - Get status information
  - `.do` - Perform read-only actions (e.g., ping)

- **Modify operations** (state-changing):
  - `.update` - Update configuration
  - `.set` - Set new values
  - `.create` - Create resources
  - `.delete` - Remove resources
  - `.execute` - Execute commands

### Special Hostnames

- `_any`: Route to any available worker
- `_all`: Route to all workers (broadcast)
- `{hostname}`: Route to specific worker

## Job Lifecycle

### 1. Job Submission

```go
// Via API
POST /api/v1/jobs
{
  "operation": {
    "type": "network.dns.get",
    "data": {"interface": "eth0"}
  },
  "target_hostname": "_any"
}

// Via CLI
osapi client job add --json-file dns-query.json --target-hostname _any
```

### 2. Job States

```
unprocessed → processing → completed
                    ↓
                 failed
```

Each state transition updates:

- Job status in KV
- Status history with timestamps
- Updated_at timestamp

### 3. Job Polling

```go
// REST API polling
GET /api/v1/jobs/{job-id}

// Returns
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "created": "2024-01-10T10:00:00Z",
  "operation": {...},
  "result": {...}
}
```

## Worker Implementation

### Processing Flow

1. **Receive notification** from JetStream
2. **Fetch job** from KV using key from notification
3. **Update status** to "processing"
4. **Execute operation** based on category/operation
5. **Store result** in job-responses KV
6. **Update job** with final status and result
7. **ACK message** to JetStream

### Provider Pattern

Workers use platform-specific providers:

```go
// Provider selection based on platform
switch platform {
case "ubuntu":
    provider = dns.NewUbuntuProvider()
default:
    provider = dns.NewLinuxProvider()
}
```

## Operation Examples

### System Operations

```json
// Get hostname
{
  "type": "system.hostname.get",
  "data": {}
}

// Get system status
{
  "type": "system.status.get",
  "data": {}
}

// Get uptime
{
  "type": "system.uptime.get",
  "data": {}
}
```

### Network Operations

```json
// Query DNS configuration
{
  "type": "network.dns.get",
  "data": {"interface": "eth0"}
}

// Update DNS servers
{
  "type": "network.dns.update",
  "data": {
    "servers": ["8.8.8.8", "1.1.1.1"],
    "interface": "eth0"
  }
}

// Execute ping
{
  "type": "network.ping.do",
  "data": {
    "address": "google.com"
  }
}
```

## CLI Commands

### Job Management

```bash
# Add a job
osapi client job add --json-file operation.json --target-hostname _any

# List jobs
osapi client job list --status unprocessed --limit 10

# Get job details
osapi client job get --job-id 550e8400-e29b-41d4-a716-446655440000

# Run job and wait for completion
osapi client job run --json-file operation.json --timeout 60

# Monitor queue status
osapi client job status --poll-interval-seconds 5
```

### Direct Worker Testing

```bash
# Start a worker
osapi worker start

# Worker will:
# - Connect to NATS
# - Subscribe to job streams
# - Process jobs based on platform capabilities
```

## Security Considerations

1. **Authentication**: NATS authentication via environment variables
2. **Authorization**: Subject-based permissions for workers
3. **Input Validation**: All job data validated before processing
4. **Result Sanitization**: Sensitive data filtered from responses

## Performance Optimizations

1. **Batch Operations**: Workers can fetch multiple jobs per poll
2. **Connection Pooling**: Reuse NATS connections
3. **KV Caching**: Local caching of frequently accessed jobs
4. **Stream Filtering**: Workers only receive relevant job types

## Error Handling

1. **Retry Logic**: Failed jobs retry up to MaxDeliver times
2. **Dead Letter Queue**: Jobs failing after max retries
3. **Timeout Handling**: Jobs timeout after AckWait period
4. **Graceful Degradation**: Workers continue on provider errors

## Monitoring

Key metrics to track:

- Queue depth by status
- Job processing time
- Worker availability
- DLQ message count
- Stream consumer lag

## Future Enhancements

1. **Job Dependencies**: Chain multiple operations
2. **Scheduled Jobs**: Cron-like job scheduling
3. **Job Priorities**: High/medium/low priority queues
4. **Result Streaming**: Stream large results via NATS
5. **Worker Autoscaling**: Dynamic worker pool sizing
