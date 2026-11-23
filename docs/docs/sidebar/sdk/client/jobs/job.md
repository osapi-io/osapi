---
sidebar_position: 2
---

# Job

Async job queue operations.

## Methods

| Method                                  | Returns                       | Permission  | Description    |
| --------------------------------------- | ----------------------------- | ----------- | -------------- |
| `c.Job.List(ctx, params JobListParams)` | `Response[ListJobsResponse]`  | `job:read`  | List jobs      |
| `c.Job.Get(ctx, id)`                    | `Response[JobDetailResponse]` | `job:read`  | Get job detail |
| `c.Job.Delete(ctx, id)`                 | `error`                       | `job:write` | Delete a job   |
| `c.Job.Retry(ctx, input JobRetryInput)` | `Response[CreateJobResponse]` | `job:write` | Retry a job    |

## Types

The fields of each type are the properties of its schema in the API's spec.

### JobListParams

| Field    | Type     | Description                              |
| -------- | -------- | ---------------------------------------- |
| `Status` | `string` | Filter jobs by status.                   |
| `Limit`  | `int`    | Maximum number of jobs per page (1-100). |
| `Offset` | `int`    | Number of jobs to skip for pagination.   |

### ListJobsResponse

| Field          | Type                  | Description                                                                                                                                                   |
| -------------- | --------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Items`        | `[]JobDetailResponse` | Items is the property items.                                                                                                                                  |
| `StatusCounts` | `map[string]int`      | Count of all jobs by status (submitted, processing, completed, failed, partial_failure). Derived from key names during the listing pass, with no extra reads. |
| `TotalItems`   | `int`                 | Total number of jobs matching the filter.                                                                                                                     |

### JobDetailResponse

| Field         | Type                             | Description                                                                                                                                                      |
| ------------- | -------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `AgentStates` | `[]JobDetailResponseAgentStates` | How far each agent got, by hostname.                                                                                                                             |
| `Chain`       | `[]string`                       | The calls of other behaviors that the method made on each host, one line for each host, such as web-01: netplan update (changed) → managedfile update (changed). |
| `Changed`     | `bool`                           | Whether the operation modified system state. Null for pending/failed jobs; true/false for completed jobs.                                                        |
| `Created`     | `string`                         | Creation timestamp.                                                                                                                                              |
| `Error`       | `string`                         | Error message if failed.                                                                                                                                         |
| `Hostname`    | `string`                         | Agent hostname that processed the job.                                                                                                                           |
| `ID`          | `string`                         | Unique identifier of the job.                                                                                                                                    |
| `Operation`   | `JobDetailResponseOperation`     | What the job asked for, with each secret masked.                                                                                                                 |
| `Responses`   | `[]JobDetailResponseResponses`   | What each agent answered, by hostname.                                                                                                                           |
| `Result`      | `any`                            | The result data if completed.                                                                                                                                    |
| `Status`      | `string`                         | Current status of the job.                                                                                                                                       |
| `Target`      | `string`                         | The target the request gave: _any, _all, a hostname, or a label such as group:web.dev. A job stored before targets were kept has the host that answered.         |
| `Timeline`    | `[]JobDetailResponseTimeline`    | Chronological sequence of job lifecycle events.                                                                                                                  |
| `UpdatedAt`   | `string`                         | Last update timestamp.                                                                                                                                           |

### JobRetryInput

| Field            | Type     | Description                                                                              |
| ---------------- | -------- | ---------------------------------------------------------------------------------------- |
| `ID`             | `string` | The id of the path: UUID of the job to retry.                                            |
| `TargetHostname` | `string` | The target of the retried job: _any, _all, a hostname, or a label such as group:web.dev. |

### CreateJobResponse

| Field       | Type     | Description                            |
| ----------- | -------- | -------------------------------------- |
| `JobID`     | `string` | Unique identifier for the created job. |
| `Revision`  | `int64`  | The KV revision number.                |
| `Status`    | `string` | Initial status of the job.             |
| `Timestamp` | `string` | Creation timestamp.                    |

## Usage

```go
// List completed jobs
resp, err := c.Job.List(ctx, client.JobListParams{
    Status: "completed",
    Limit:  20,
})

// Retry a failed job
resp, err := c.Job.Retry(ctx, client.JobRetryInput{
    ID: "uuid-string",
    TargetHostname: "_any",
})
```

## Example

Refer to
[`examples/sdk/client/job.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/job.go)
for a complete working example.

## Permissions

The Permission column of [Methods](#methods) gives the permission of each
method.
