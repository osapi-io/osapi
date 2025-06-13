# OSAPI Job System Architecture

**Date:** June 2025  
**Status:** Implemented  
**Author:** System Architecture Team

## Overview

The OSAPI Job System implements a **KV-first, stream-notification architecture** using NATS JetStream for distributed job processing. This system provides asynchronous operation execution with persistent job state, intelligent worker routing, and comprehensive job lifecycle management.

## Architecture Principles

- **KV-First Storage**: Job state lives in NATS KV for persistence and direct access
- **Stream Notifications**: Workers receive job notifications via JetStream subjects
- **Hierarchical Routing**: Operations use dot-notation for intelligent worker targeting
- **REST-Compatible**: Supports standard HTTP polling patterns for API integration
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
                    │  Subject Routing        │ <--- Intelligent Dispatch
                    └─────────────────────────┘
```

### Job Storage (KV Store)

**Bucket**: `job-queue`  
**Purpose**: Persistent job state and result storage  
**Key Structure**: `{status}.{job-uuid}` for efficient filtering

#### Key Format
```
unprocessed.{uuid}    # New jobs waiting for processing
processing.{uuid}     # Jobs currently being worked on  
completed.{uuid}      # Successfully completed jobs
failed.{uuid}         # Jobs that encountered errors
```

#### Job Data Structure
```json
{
  "id": "uuid-12345",
  "status": "unprocessed|processing|completed|failed",
  "created": "2025-06-14T10:00:00Z",
  "updated_at": "2025-06-14T10:01:30Z",
  "subject": "jobs.query._any.system.hostname.get",  // NATS subject for routing
  "operation": {
    "type": "system.hostname.get",
    "data": {}
  },
  "result": { ... },        // Present when completed
  "error": "...",           // Present when failed
  "status_history": [       // Track all status transitions
    {
      "status": "unprocessed",
      "timestamp": "2025-06-14T10:00:00Z"
    },
    {
      "status": "processing", 
      "timestamp": "2025-06-14T10:01:00Z"
    },
    {
      "status": "completed",
      "timestamp": "2025-06-14T10:01:30Z"
    }
  ]
}
```

#### Queue Statistics Structure
```json
{
  "total_jobs": 42,
  "status_counts": {
    "unprocessed": 5,
    "processing": 2, 
    "completed": 30,
    "failed": 5
  },
  "operation_counts": {
    "system.hostname.get": 15,
    "system.status.get": 19,
    "system.disk.get": 12,
    "network.dns.get": 8,
    "network.dns.update": 5,
    "network.ping.do": 23
  }
}
```

#### Performance Benefits
- **Efficient filtering**: `Watch("failed.*")` only retrieves failed jobs
- **Scalable listing**: No need to load all jobs to filter by status
- **Active queue focus**: Default views show only `unprocessed.*` and `processing.*`
- **Historical access**: Completed jobs remain accessible via UUID

### Worker Notification (Streams)

**Stream**: `JOBS`  
**Subjects**: `jobs.>` (all job-related subjects)  
**Purpose**: Notify workers of new jobs without consuming job data

## Subject Routing Hierarchy

### Subject Format
```
jobs.{type}.{hostname}.{category}.{operation}
```

### Operation Types
- **Query Operations**: `jobs.query.*` - Read-only operations and actions
- **Modify Operations**: `jobs.modify.*` - State-changing operations

### Operation Suffix Routing
The job client automatically routes operations based on their suffix:

**Query Suffixes** (→ `jobs.query.*`):
- `.get` - Retrieve current state (DNS config, hostname, disk usage)
- `.query` - Query for information  
- `.read` - Read data
- `.status` - Get status information
- `.do` - Perform actions/execution (ping, tests, measurements)

**Modify Suffixes** (→ `jobs.modify.*`):
- `.update` - Update configuration
- `.set` - Set values
- `.create` - Create resources
- `.delete` - Delete resources

**System Category Exception**: All `system.*` operations are automatically routed to query subjects regardless of suffix.

### Hostname Routing
- **`_any`**: Load-balanced across available workers (queue groups)
- **`_all`**: Broadcast to all workers (no queue groups)
- **`specific-host`**: Direct targeting to named hosts
- **`*`**: Wildcard matching for subscription patterns

### Example Subjects
```bash
# System queries (read-only)
jobs.query._any.system.hostname.get
jobs.query._any.system.status.get
jobs.query._any.system.uptime.get
jobs.query._any.system.disk.get
jobs.query._any.system.memory.get
jobs.query._any.system.load.get
jobs.query._any.system.os.get

# Network queries (read-only and actions)
jobs.query._any.network.dns.get
jobs.query._any.network.ping.do

# Network modifications (state-changing)  
jobs.modify._any.network.dns.update
jobs.modify.server1.network.dns.update
```

## Supported Operations

### System Operations (Query)
All system operations are routed to `jobs.query.*` subjects:

- **`system.hostname.get`** - Get system hostname
- **`system.status.get`** - Get comprehensive system status (all providers)
- **`system.uptime.get`** - Get system uptime
- **`system.os.get`** - Get operating system information
- **`system.disk.get`** - Get disk usage statistics
- **`system.memory.get`** - Get memory statistics  
- **`system.load.get`** - Get load average statistics

### Network Operations

**Query/Action Operations** (`jobs.query.*`):
- **`network.dns.get`** - Get DNS configuration for interface
- **`network.ping.do`** - Execute ping to target address

**Modify Operations** (`jobs.modify.*`):
- **`network.dns.update`** - Update DNS servers and search domains

### Provider Mapping
- **System Host**: hostname, uptime, OS info
- **System Disk**: disk usage statistics
- **System Memory**: memory usage statistics
- **System Load**: load average statistics
- **Network DNS**: DNS configuration (get/update)
- **Network Ping**: ping execution and statistics

## Data Flow

### 1. Job Creation Flow

```
Client/API ---|
              | 1. Create Job Request
              v
      ┌─────────────┐
      │  Job Creator│ 2. Generate UUID
      │  (API/CLI)  │ 3. Build operation data
      └─────────────┘ 4. Initialize status_history
              | 5. Store with status prefix
              v
      ┌─────────────┐ 6. PUT unprocessed.{uuid} ┌─────────────┐
      │ KV Storage  │<--------------------------│ NATS Client │
      │ (job-queue) │                           │             │
      └─────────────┘                           └─────────────┘
                                                        | 7. Publish notification
                                                        v
                                               ┌─────────────┐
                                               │   Stream    │
                                               │   (JOBS)    │
                                               └─────────────┘
                                                        | 8. Notify workers
                                                        v
                                               ┌─────────────┐
                                               │   Workers   │
                                               │ (Subscribe) │
                                               └─────────────┘
```

### 2. Job Processing Flow

```
Worker ---|
          | 1. Receive notification (job ID only)
          v
 ┌─────────────┐ 2. Find job by status ┌─────────────┐
 │   Worker    │ prefixes             │ KV Storage  │
 │  Process    │<---------------------│ (job-queue) │
 └─────────────┘                      └─────────────┘
          | 3. Move: unprocessed.{uuid} → processing.{uuid}
          | 4. Add status_history entry
          v
 ┌─────────────┐ 5. PUT processing.{uuid} ┌─────────────┐
 │   Worker    │------------------------->│ KV Storage  │
 │  Update     │ 6. DELETE unprocessed.*  │ (job-queue) │
 └─────────────┘                          └─────────────┘
          | 7. Execute operation
          v
 ┌─────────────┐
 │  Provider   │ 8. Perform work
 │  Services   │ 9. Return result
 └─────────────┘
          | 10. Store result + final status
          v
 ┌─────────────┐ 11. Move: processing.{uuid} → completed.{uuid}
 │   Worker    │ 12. Add final status_history
 │  Complete   │ 13. Preserve original created timestamp
 └─────────────┘
          | 14. PUT completed.{uuid}
          | 15. DELETE processing.{uuid}
          v
 ┌─────────────┐
 │ KV Storage  │ Final job state with complete audit trail
 │ (job-queue) │
 └─────────────┘
```

### 3. Job Query Flow (API Integration)

```
HTTP Client ---|
               | 1. GET /api/jobs/{uuid}
               v
       ┌─────────────┐ 2. Search status prefixes:  ┌─────────────┐
       │  REST API   │    - unprocessed.{uuid}     │ KV Storage  │
       │  Server     │    - processing.{uuid}      │ (job-queue) │
       └─────────────┘    - completed.{uuid}       └─────────────┘
               │          - failed.{uuid}
               │ 3. Return job state + results + status_history
               v
       ┌─────────────┐
       │ HTTP Client │ 4. Complete job data with audit trail
       │  Response   │
       └─────────────┘
```

**Note**: Job retrieval by UUID requires checking up to 4 keys (one per status), which is acceptable performance for individual job queries while enabling efficient bulk operations.

## Worker Subscription Patterns

### Load-Balanced Workers (Queue Groups)
```go
// DNS specialist workers
consumer.Subscribe("jobs.*._any.network.dns.>", 
                  nats.Queue("dns-workers"))

// System monitoring workers  
consumer.Subscribe("jobs.*._any.system.>",
                  nats.Queue("system-workers"))
```

### Direct Host Workers
```go
// Worker on specific host
hostname, _ := os.Hostname()
consumer.Subscribe(fmt.Sprintf("jobs.*.%s.>", hostname))
```

### Broadcast Workers (No Queue Groups)
```go
// All workers receive urgent notifications
consumer.Subscribe("jobs.*._all.>")
```

## Operation Type Definitions

### System Operations (Query)
```go
const (
    OperationSystemHostnameGet = "system.hostname.get"
    OperationSystemStatusGet   = "system.status.get"
    OperationSystemUptimeGet   = "system.uptime.get"
    OperationSystemLoadGet     = "system.load.get"
    OperationSystemMemoryGet   = "system.memory.get"
    OperationSystemDiskGet     = "system.disk.get"
)
```

### Network Operations (Modify)
```go
const (
    OperationNetworkDNSGet      = "network.dns.get"
    OperationNetworkDNSUpdate   = "network.dns.update"
    OperationNetworkPingExecute = "network.ping.execute"
)
```

### System Operations (Modify)
```go
const (
    OperationSystemShutdown = "system.shutdown.execute"
    OperationSystemReboot   = "system.reboot.execute"
)
```

## REST API Integration

### Job Creation Endpoint
```http
POST /api/jobs
Content-Type: application/json

{
  "type": "system.hostname.get",
  "data": {},
  "target_hostname": "_any"
}
```

**Response:**
```json
{
  "job_id": "uuid-12345",
  "status": "created",
  "revision": 1
}
```

### Job Status Endpoint
```http
GET /api/jobs/{job_id}
```

**Response (Processing):**
```json
{
  "job_id": "uuid-12345",
  "status": "processing",
  "created": "2025-06-14T10:00:00Z",
  "operation": {
    "type": "system.hostname.get",
    "data": {}
  }
}
```

**Response (Completed):**
```json
{
  "job_id": "uuid-12345", 
  "status": "completed",
  "created": "2025-06-14T10:00:00Z",
  "operation": {
    "type": "system.hostname.get",
    "data": {}
  },
  "result": {
    "hostname": "server-01"
  }
}
```

## CLI Management Interface

### Available Commands
```bash
# Create jobs directly
osapi client job add --json-file operation.json --target-hostname myserver

# View real-time queue status (TUI)
osapi client job status

# View queue status (JSON)
osapi client job status --json

# Get job details (searches all status prefixes)
osapi client job get --job-id uuid-12345

# List jobs by status (efficient prefix filtering) 
osapi client job list --status failed
osapi client job list --status completed --limit 20

# Delete jobs
osapi client job delete --job-id uuid-12345
```

### Configuration Options
```bash
# NATS connection
--nats-host localhost
--client-name osapi-jobs-cli

# KV bucket selection
--kv-bucket job-queue

# Output format
--json
```

## Operational Benefits

### Scalability
- **Horizontal worker scaling**: Add workers to handle increased load
- **Geographic distribution**: Route jobs to workers by location/capability
- **Load balancing**: Automatic distribution via NATS queue groups
- **Efficient filtering**: Status-based key prefixes enable fast queries at scale

### Reliability  
- **Job persistence**: Survives system restarts and network partitions
- **Result storage**: Completed job data available for extended periods
- **Status tracking**: Complete job lifecycle visibility with audit trails
- **Atomic operations**: Create-before-delete pattern prevents data loss
- **Metadata preservation**: Original timestamps maintained across status changes

### Monitoring & Observability
- **Job metrics**: Count by status using efficient prefix queries
- **Performance tracking**: Status history shows execution timing
- **Queue inspection**: Focus on active jobs without loading completed history
- **Audit trails**: Complete transition history with timestamps

### Security & Authorization
- **Subject-based ACLs**: Control access to operation types
- **Host targeting**: Restrict operations to authorized hosts  
- **Complete audit trails**: Full job lifecycle history in KV store
- **Data integrity**: Status transitions are logged and preserved

## Package Architecture

### Type Organization
```
internal/job/
├── types.go              # Core domain types (Request, Response, QueuedJob, QueueStats)
├── subjects.go           # Subject routing and pattern generation  
├── config.go            # Configuration structures
├── client/              # High-level job operations for API embedding
│   ├── client.go        # Job client with CreateJob, GetQueueStats, etc.
│   ├── query.go         # Query operations (system status, hostname, etc.)
│   ├── modify.go        # Modify operations (DNS updates, ping, etc.)
│   └── types.go         # Client-specific types and interfaces
└── worker/              # Job processing and worker lifecycle
    ├── worker.go        # Worker implementation and lifecycle management
    ├── processor.go     # Job processing logic and provider integration
    ├── manager.go       # Worker manager interface
    └── types.go         # Worker-specific types and context
```

### Separation of Concerns
- **`internal/job/`**: Core domain types shared across all components
- **`internal/job/client/`**: High-level operations for API integration
- **`internal/job/worker/`**: Job processing and worker lifecycle management
- **No type duplication**: All packages use shared types from main job package

## Implementation Details

### Job Client Layer
```go
// High-level job client that abstracts NATS operations
jobClient, err := jobclient.New(logger, natsClient, &jobclient.Options{
    Timeout:  30 * time.Second,
    KVBucket: jobsKV,
})

// Create a job with business logic abstraction
result, err := jobClient.CreateJob(ctx, operationData, targetHostname)

// Get queue statistics efficiently
stats, err := jobClient.GetQueueStats(ctx)

// Retrieve individual job status
jobInfo, err := jobClient.GetJobStatus(ctx, jobID)
```

### NATS Client Configuration
```go
natsClient := natsclient.New(logger, &natsclient.Options{
    Host: "localhost",
    Port: 4222,
    Auth: natsclient.AuthOptions{
        AuthType: natsclient.NoAuth,
    },
    Name: "osapi-jobs",
})
```

### Stream Setup
```go
streamConfig := &nats.StreamConfig{
    Name:     "JOBS",
    Subjects: []string{"jobs.>"},
}
```

### KV Bucket Creation
```go
jobsKV, err := natsClient.CreateKVBucket("job-queue")
```

### Job Storage Pattern
```go
// Job creation with status prefix
jobID := uuid.New().String()
key := "unprocessed." + jobID
jobData := map[string]interface{}{
    "id": jobID,
    "status": "unprocessed", 
    "created": time.Now().Format(time.RFC3339),
    "status_history": []interface{}{
        map[string]interface{}{
            "status": "unprocessed",
            "timestamp": time.Now().Format(time.RFC3339),
        },
    },
    "operation": operationData,
}
_, err := jobsKV.Put(key, jobJSON)
```

### Job Retrieval Pattern
```go
// Search across status prefixes for UUID
statuses := []string{"unprocessed", "processing", "completed", "failed"}
for _, status := range statuses {
    key := status + "." + jobID
    entry, err := jobsKV.Get(key)
    if err == nil {
        return entry // Found the job
    }
}
```

### Efficient Status Filtering
```go
// List only failed jobs
watcher, err := jobsKV.Watch("failed.*")
for entry := range watcher.Updates() {
    if entry == nil { break } // End of initial values
    // Process failed job...
}
```

## Migration Path

1. **✅ Phase 1**: Deploy job system alongside existing task system *(Complete)*
2. **✅ Phase 2**: Implement job client abstraction layer *(Complete)*
3. **🔄 Phase 3**: Integrate job client into REST API *(In Progress)*
4. **📋 Phase 4**: Migrate GET endpoints to use job system
5. **📋 Phase 5**: Deprecate legacy task system  
6. **📋 Phase 6**: Full transition to jobs-based architecture

## Future Enhancements

- **Job scheduling**: Time-based job execution
- **Job dependencies**: Workflow orchestration
- **Job retries**: Automatic retry with backoff
- **Job prioritization**: Priority queues for critical operations
- **Job cancellation**: Ability to cancel running jobs
- **Distributed workers**: Cross-datacenter job execution