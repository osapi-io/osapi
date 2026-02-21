// Copyright (c) 2025 John Dewey

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/job"
)

// CreateJob creates a new job from operation data and stores it in the KV bucket.
func (c *Client) CreateJob(
	ctx context.Context,
	operationData map[string]interface{},
	targetHostname string,
) (*CreateJobResult, error) {
	// Extract operation type for semantic routing
	operationType, ok := operationData["type"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid operation format: missing type field")
	}

	// Default to load-balanced routing if no specific hostname provided
	if targetHostname == "" {
		targetHostname = job.AnyHost
	}

	// Build the notification subject using semantic routing based on operation type.
	// Route based on operation name - operations ending in common read patterns are queries.
	var prefix string
	if strings.HasSuffix(operationType, ".get") ||
		strings.HasSuffix(operationType, ".query") ||
		strings.HasSuffix(operationType, ".read") ||
		strings.HasSuffix(operationType, ".status") ||
		strings.HasSuffix(operationType, ".do") ||
		strings.HasPrefix(operationType, "system.") {
		prefix = job.JobsQueryPrefix
	} else {
		prefix = job.JobsModifyPrefix
	}

	notificationSubject := job.BuildSubjectFromTarget(prefix, targetHostname)

	// Generate job ID
	jobID := uuid.New().String()

	// Create job with status metadata
	createdTime := time.Now().Format(time.RFC3339)
	jobWithStatus := map[string]interface{}{
		"id":        jobID,
		"status":    "unprocessed",
		"created":   createdTime,
		"subject":   notificationSubject,
		"operation": operationData,
		"status_history": []interface{}{
			map[string]interface{}{
				"status":    "unprocessed",
				"timestamp": createdTime,
			},
		},
	}

	jobWithStatusJSON, err := json.Marshal(jobWithStatus)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal job with status: %w", err)
	}

	// Store job in KV with immutable key format
	kvKey := "jobs." + jobID
	c.logger.DebugContext(ctx, "storing job and sending notification",
		slog.String("kv_bucket", c.kv.Bucket()),
		slog.String("key", kvKey),
		slog.String("subject", notificationSubject),
		slog.String("job_id", jobID),
	)

	revision, err := c.kv.Put(kvKey, jobWithStatusJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to store job in KV: %w", err)
	}

	// Write initial submitted status event
	statusKey := fmt.Sprintf("status.%s.submitted._api.%d", jobID, time.Now().UnixNano())
	statusEvent := map[string]interface{}{
		"job_id":    jobID,
		"event":     "submitted",
		"hostname":  "_api", // API server hostname could be added here
		"timestamp": createdTime,
		"unix_nano": time.Now().UnixNano(),
		"data": map[string]interface{}{
			"target_hostname": targetHostname,
			"operation_type":  operationType,
		},
	}
	statusEventJSON, _ := json.Marshal(statusEvent)
	if _, err := c.kv.Put(statusKey, statusEventJSON); err != nil {
		c.logger.ErrorContext(
			ctx,
			"failed to write submitted event",
			slog.String("error", err.Error()),
		)
	}

	// Send notification via stream.
	// Trace context is propagated via NATS message headers automatically by nats-client.
	err = c.natsClient.Publish(ctx, notificationSubject, []byte(jobID))
	if err != nil {
		return nil, fmt.Errorf("failed to send notification: %w", err)
	}

	c.logger.DebugContext(ctx, "job created successfully",
		slog.String("job_id", jobID),
		slog.Uint64("revision", revision),
		slog.String("subject", notificationSubject),
		slog.String("target_hostname", targetHostname),
	)

	return &CreateJobResult{
		JobID:     jobID,
		Status:    "created",
		Revision:  revision,
		Timestamp: createdTime,
	}, nil
}

// GetQueueStats returns statistics about the job queue.
func (c *Client) GetQueueStats(
	ctx context.Context,
) (*job.QueueStats, error) {
	c.logger.Debug("kv.keys",
		slog.String("operation", "get_queue_stats"),
	)
	// Get all job keys from KV store
	keys, err := c.kv.Keys()
	if err != nil {
		if err.Error() == "nats: no keys found" {
			return &job.QueueStats{
				TotalJobs: 0,
				StatusCounts: map[string]int{
					"submitted":  0,
					"processing": 0,
					"completed":  0,
					"failed":     0,
				},
				OperationCounts: map[string]int{},
				DLQCount:        0,
			}, nil
		}
		return nil, fmt.Errorf("error fetching jobs: %w", err)
	}

	statusCounts := map[string]int{
		"submitted":       0,
		"processing":      0,
		"completed":       0,
		"failed":          0,
		"partial_failure": 0,
	}

	operationCounts := map[string]int{}
	jobCount := 0

	// Process immutable jobs and compute status from events
	for _, key := range keys {
		// Only process "jobs." prefixed keys
		if !strings.HasPrefix(key, "jobs.") {
			continue
		}

		entry, err := c.kv.Get(key)
		if err != nil {
			continue
		}

		var jobData map[string]interface{}
		if err := json.Unmarshal(entry.Value(), &jobData); err != nil {
			continue
		}

		jobID := strings.TrimPrefix(key, "jobs.")
		jobCount++

		// Compute status from events (reuse already-fetched keys)
		computedStatus := c.computeStatusFromEvents(keys, jobID)
		statusCounts[computedStatus.Status]++

		// Track operation type
		if jobOperationData, ok := jobData["operation"].(map[string]interface{}); ok {
			if operationType, ok := jobOperationData["type"].(string); ok {
				operationCounts[operationType]++
			}
		}
	}

	// Get DLQ count
	dlqCount := 0
	dlqInfo, err := c.natsClient.GetStreamInfo(ctx, "JOBS-DLQ")
	if err != nil {
		// DLQ might not exist, which is fine
		c.logger.Debug("failed to get DLQ info", slog.String("error", err.Error()))
	} else {
		dlqCount = int(dlqInfo.State.Msgs)
	}

	return &job.QueueStats{
		TotalJobs:       jobCount,
		StatusCounts:    statusCounts,
		OperationCounts: operationCounts,
		DLQCount:        dlqCount,
	}, nil
}

// GetJobStatus returns information about a specific job.
func (c *Client) GetJobStatus(
	_ context.Context,
	jobID string,
) (*job.QueuedJob, error) {
	// Get the immutable job data
	jobKey := "jobs." + jobID
	c.logger.Debug("kv.get",
		slog.String("key", jobKey),
	)
	entry, err := c.kv.Get(jobKey)
	if err != nil {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	var jobData map[string]interface{}
	if err := json.Unmarshal(entry.Value(), &jobData); err != nil {
		return nil, fmt.Errorf("failed to parse job data: %w", err)
	}

	// Get all status events for this job
	statusKeys, err := c.kv.Keys()
	if err != nil && err.Error() != "nats: no keys found" {
		return nil, fmt.Errorf("failed to get status events: %w", err)
	}

	// Compute current status from events
	computedStatus := c.computeStatusFromEvents(statusKeys, jobID)

	// Get response data for this job
	responses := c.getJobResponses(statusKeys, jobID)

	queuedJob := &job.QueuedJob{
		ID:           jobID,
		Status:       computedStatus.Status,
		Created:      getStringField(jobData, "created"),
		Subject:      getStringField(jobData, "subject"),
		Hostname:     computedStatus.Hostname,
		Error:        computedStatus.Error,
		UpdatedAt:    computedStatus.UpdatedAt,
		WorkerStates: computedStatus.WorkerStates,
		Timeline:     computedStatus.Timeline,
		Responses:    responses,
	}

	if operation, ok := jobData["operation"].(map[string]interface{}); ok {
		queuedJob.Operation = operation
	}

	// Populate Result from single-worker response
	if len(responses) == 1 {
		for _, resp := range responses {
			queuedJob.Result = resp.Data
			break
		}
	}

	return queuedJob, nil
}

// ListJobs returns jobs filtered by status with server-side pagination.
// limit=0 means no limit. offset=0 means start from beginning.
// Jobs are returned newest-first (reverse insertion order).
func (c *Client) ListJobs(
	_ context.Context,
	statusFilter string,
	limit int,
	offset int,
) (*ListJobsResult, error) {
	c.logger.Debug("kv.keys",
		slog.String("operation", "list_jobs"),
		slog.String("status_filter", statusFilter),
		slog.Int("limit", limit),
		slog.Int("offset", offset),
	)
	allKeys, err := c.kv.Keys()
	if err != nil {
		if err.Error() == "nats: no keys found" {
			return &ListJobsResult{
				Jobs:       []*job.QueuedJob{},
				TotalCount: 0,
			}, nil
		}
		return nil, fmt.Errorf("error fetching jobs: %w", err)
	}

	// Extract job keys and reverse for newest-first ordering
	var jobKeys []string
	for _, key := range allKeys {
		if strings.HasPrefix(key, "jobs.") {
			jobID := strings.TrimPrefix(key, "jobs.")
			if jobID != "" {
				jobKeys = append(jobKeys, key)
			}
		}
	}

	// Reverse for newest-first (KV insertion order is chronological)
	for i, j := 0, len(jobKeys)-1; i < j; i, j = i+1, j-1 {
		jobKeys[i], jobKeys[j] = jobKeys[j], jobKeys[i]
	}

	c.logger.Debug("kv.keys",
		slog.Int("total_job_keys", len(jobKeys)),
		slog.Int("all_keys", len(allKeys)),
	)

	// No status filter: we know the total count immediately
	if statusFilter == "" {
		return c.listJobsNoFilter(allKeys, jobKeys, limit, offset), nil
	}

	// With status filter: scan all jobs to count matches and collect page
	return c.listJobsWithFilter(allKeys, jobKeys, statusFilter, limit, offset), nil
}

// listJobsNoFilter handles pagination when no status filter is applied.
func (c *Client) listJobsNoFilter(
	allKeys []string,
	jobKeys []string,
	limit int,
	offset int,
) *ListJobsResult {
	totalCount := len(jobKeys)

	// Apply offset
	if offset >= len(jobKeys) {
		return &ListJobsResult{
			Jobs:       []*job.QueuedJob{},
			TotalCount: totalCount,
		}
	}
	jobKeys = jobKeys[offset:]

	// Apply limit
	if limit > 0 && len(jobKeys) > limit {
		jobKeys = jobKeys[:limit]
	}

	var jobs []*job.QueuedJob
	for _, key := range jobKeys {
		jobID := strings.TrimPrefix(key, "jobs.")
		jobInfo, err := c.getJobStatusFromKeys(allKeys, key, jobID)
		if err != nil {
			c.logger.Debug("failed to get job status during list",
				slog.String("job_id", jobID),
				slog.String("error", err.Error()),
			)
			continue
		}
		jobs = append(jobs, jobInfo)
	}

	if jobs == nil {
		jobs = []*job.QueuedJob{}
	}

	return &ListJobsResult{
		Jobs:       jobs,
		TotalCount: totalCount,
	}
}

// listJobsWithFilter handles pagination when a status filter is applied.
func (c *Client) listJobsWithFilter(
	allKeys []string,
	jobKeys []string,
	statusFilter string,
	limit int,
	offset int,
) *ListJobsResult {
	var jobs []*job.QueuedJob
	totalCount := 0
	skipped := 0

	for _, key := range jobKeys {
		jobID := strings.TrimPrefix(key, "jobs.")
		jobInfo, err := c.getJobStatusFromKeys(allKeys, key, jobID)
		if err != nil {
			c.logger.Debug("failed to get job status during list",
				slog.String("job_id", jobID),
				slog.String("error", err.Error()),
			)
			continue
		}

		if jobInfo.Status != statusFilter {
			continue
		}

		totalCount++

		// Skip offset items
		if skipped < offset {
			skipped++
			continue
		}

		// Collect items up to limit
		if limit == 0 || len(jobs) < limit {
			jobs = append(jobs, jobInfo)
		}
	}

	if jobs == nil {
		jobs = []*job.QueuedJob{}
	}

	return &ListJobsResult{
		Jobs:       jobs,
		TotalCount: totalCount,
	}
}

// getJobStatusFromKeys builds a QueuedJob using pre-fetched keys (no inner kv.Keys() call).
func (c *Client) getJobStatusFromKeys(
	allKeys []string,
	jobKey string,
	jobID string,
) (*job.QueuedJob, error) {
	entry, err := c.kv.Get(jobKey)
	if err != nil {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	var jobData map[string]interface{}
	if err := json.Unmarshal(entry.Value(), &jobData); err != nil {
		return nil, fmt.Errorf("failed to parse job data: %w", err)
	}

	computedStatus := c.computeStatusFromEvents(allKeys, jobID)
	responses := c.getJobResponses(allKeys, jobID)

	queuedJob := &job.QueuedJob{
		ID:           jobID,
		Status:       computedStatus.Status,
		Created:      getStringField(jobData, "created"),
		Subject:      getStringField(jobData, "subject"),
		Hostname:     computedStatus.Hostname,
		Error:        computedStatus.Error,
		UpdatedAt:    computedStatus.UpdatedAt,
		WorkerStates: computedStatus.WorkerStates,
		Timeline:     computedStatus.Timeline,
		Responses:    responses,
	}

	if operation, ok := jobData["operation"].(map[string]interface{}); ok {
		queuedJob.Operation = operation
	}

	if len(responses) == 1 {
		for _, resp := range responses {
			queuedJob.Result = resp.Data
			break
		}
	}

	return queuedJob, nil
}

// getStringField safely extracts a string field from a map.
func getStringField(
	data map[string]interface{},
	field string,
) string {
	if value, ok := data[field].(string); ok {
		return value
	}
	return ""
}

// computeStatusFromEvents computes the current job status from status events
func (c *Client) computeStatusFromEvents(
	eventKeys []string,
	jobID string,
) computedJobStatus {
	result := computedJobStatus{
		Status:       "submitted", // Default if no events
		WorkerStates: make(map[string]job.WorkerState),
		Timeline:     []job.TimelineEvent{},
	}

	// Filter keys to only status events for this job
	statusPrefix := "status." + jobID + "."
	var relevantKeys []string
	for _, key := range eventKeys {
		if strings.HasPrefix(key, statusPrefix) {
			relevantKeys = append(relevantKeys, key)
		}
	}

	// If no events, return default
	if len(relevantKeys) == 0 {
		return result
	}

	// Track status by worker
	workerStates := make(map[string]string)
	workerStartTimes := make(map[string]time.Time)
	workerEndTimes := make(map[string]time.Time)
	workerErrors := make(map[string]string)
	var latestEvent time.Time
	var latestError string
	var timeline []job.TimelineEvent

	// Process each event
	for _, key := range relevantKeys {
		entry, err := c.kv.Get(key)
		if err != nil {
			continue
		}

		var event map[string]interface{}
		if err := json.Unmarshal(entry.Value(), &event); err != nil {
			continue
		}

		hostname := getStringField(event, "hostname")
		eventType := getStringField(event, "event")
		timestamp := getStringField(event, "timestamp")

		// Parse timestamp (supports both RFC3339 and RFC3339Nano)
		t, timestampErr := time.Parse(time.RFC3339Nano, timestamp)
		if timestampErr != nil {
			continue
		}

		// Add to timeline
		timelineEvent := job.TimelineEvent{
			Timestamp: t,
			Event:     eventType,
			Hostname:  hostname,
		}

		// Set message based on event type
		switch eventType {
		case "submitted":
			timelineEvent.Message = "Job submitted to queue"
		case "acknowledged":
			timelineEvent.Message = fmt.Sprintf("Job acknowledged by worker %s", hostname)
		case "started":
			timelineEvent.Message = fmt.Sprintf("Job processing started on %s", hostname)
		case "completed":
			timelineEvent.Message = fmt.Sprintf("Job completed successfully on %s", hostname)
		case "failed":
			timelineEvent.Message = fmt.Sprintf("Job failed on %s", hostname)
			if data, ok := event["data"].(map[string]interface{}); ok {
				if errMsg, ok := data["error"].(string); ok {
					timelineEvent.Error = errMsg
					workerErrors[hostname] = errMsg
					latestError = errMsg
				}
			}
		case "retried":
			if data, ok := event["data"].(map[string]interface{}); ok {
				if newJobID, ok := data["new_job_id"].(string); ok {
					timelineEvent.Message = fmt.Sprintf("Job retried as %s", newJobID)
				}
			}
			if timelineEvent.Message == "" {
				timelineEvent.Message = "Job retried"
			}
		}

		timeline = append(timeline, timelineEvent)

		// Track worker state
		if hostname != "" && hostname != "_api" {
			workerStates[hostname] = eventType

			// Track first start time for duration calculation
			if eventType == "started" {
				if _, exists := workerStartTimes[hostname]; !exists {
					workerStartTimes[hostname] = t
				}
			}

			// Track last end time for duration calculation
			if eventType == "completed" || eventType == "failed" {
				workerEndTimes[hostname] = t
				result.Hostname = hostname
			}
		}

		// Track latest timestamp
		if t.After(latestEvent) {
			latestEvent = t
			result.UpdatedAt = timestamp
		}
	}

	// Sort timeline by timestamp
	for i := 0; i < len(timeline)-1; i++ {
		for j := i + 1; j < len(timeline); j++ {
			if timeline[i].Timestamp.After(timeline[j].Timestamp) {
				timeline[i], timeline[j] = timeline[j], timeline[i]
			}
		}
	}
	result.Timeline = timeline

	// Build WorkerStates with detailed information
	for hostname, state := range workerStates {
		workerState := job.WorkerState{
			Status: state,
		}

		// Add error if worker failed
		if errorMsg, hasError := workerErrors[hostname]; hasError {
			workerState.Error = errorMsg
		}

		// Calculate duration from first start to last end
		if startTime, hasStart := workerStartTimes[hostname]; hasStart {
			workerState.StartTime = startTime

			if endTime, hasEnd := workerEndTimes[hostname]; hasEnd {
				workerState.EndTime = endTime
				duration := endTime.Sub(startTime)
				workerState.Duration = duration.String()
			}
		}

		result.WorkerStates[hostname] = workerState
	}

	// Compute overall status based on worker states
	completed := 0
	failed := 0
	processing := 0
	acknowledged := 0

	for _, state := range workerStates {
		switch state {
		case "completed":
			completed++
		case "failed":
			failed++
		case "started":
			processing++
		case "acknowledged":
			acknowledged++
		}
	}

	totalWorkers := len(workerStates)

	// Determine overall status
	if totalWorkers == 0 {
		result.Status = "submitted"
	} else if processing > 0 || acknowledged > 0 {
		result.Status = "processing"
	} else if completed+failed == totalWorkers {
		if failed > 0 && completed > 0 {
			result.Status = "partial_failure"
		} else if failed > 0 {
			result.Status = "failed"
			result.Error = latestError
		} else {
			result.Status = "completed"
		}
	}

	return result
}

// RetryJob creates a new job using the same operation data as an existing job.
// The original job is preserved. A "retried" status event is written to the
// original job's timeline linking to the new job.
func (c *Client) RetryJob(
	ctx context.Context,
	jobID string,
	targetHostname string,
) (*CreateJobResult, error) {
	// Read original job from KV
	jobKey := "jobs." + jobID
	c.logger.Debug("kv.get",
		slog.String("key", jobKey),
		slog.String("operation", "retry_job"),
	)

	entry, err := c.kv.Get(jobKey)
	if err != nil {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	var jobData map[string]interface{}
	if err := json.Unmarshal(entry.Value(), &jobData); err != nil {
		return nil, fmt.Errorf("failed to parse job data: %w", err)
	}

	// Extract operation data
	operationData, ok := jobData["operation"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("job has no operation data: %s", jobID)
	}

	// Create new job with the same operation data
	result, err := c.CreateJob(ctx, operationData, targetHostname)
	if err != nil {
		return nil, fmt.Errorf("failed to create retry job: %w", err)
	}

	// Write "retried" status event on the original job's timeline
	retriedKey := fmt.Sprintf("status.%s.retried._api.%d", jobID, time.Now().UnixNano())
	retriedEvent := map[string]interface{}{
		"job_id":    jobID,
		"event":     "retried",
		"hostname":  "_api",
		"timestamp": time.Now().Format(time.RFC3339Nano),
		"unix_nano": time.Now().UnixNano(),
		"data": map[string]interface{}{
			"new_job_id":      result.JobID,
			"target_hostname": targetHostname,
		},
	}
	retriedEventJSON, _ := json.Marshal(retriedEvent)
	if _, err := c.kv.Put(retriedKey, retriedEventJSON); err != nil {
		c.logger.Error("failed to write retried event",
			slog.String("error", err.Error()),
		)
	}

	c.logger.Debug("job retried successfully",
		slog.String("original_job_id", jobID),
		slog.String("new_job_id", result.JobID),
		slog.String("target_hostname", targetHostname),
	)

	return result, nil
}

// DeleteJob deletes a job from the KV store by its ID.
func (c *Client) DeleteJob(
	_ context.Context,
	jobID string,
) error {
	jobKey := "jobs." + jobID
	c.logger.Debug("kv.delete",
		slog.String("key", jobKey),
	)
	_, err := c.kv.Get(jobKey)
	if err != nil {
		return fmt.Errorf("job not found: %s", jobID)
	}

	if err := c.kv.Delete(jobKey); err != nil {
		return fmt.Errorf("failed to delete job: %w", err)
	}

	return nil
}

// getJobResponses retrieves response data for a specific job
func (c *Client) getJobResponses(
	allKeys []string,
	jobID string,
) map[string]job.Response {
	responses := make(map[string]job.Response)

	// Filter keys to only response keys for this job
	responsePrefix := "responses." + jobID + "."
	for _, key := range allKeys {
		if !strings.HasPrefix(key, responsePrefix) {
			continue
		}

		entry, err := c.kv.Get(key)
		if err != nil {
			continue
		}

		var response job.Response
		if err := json.Unmarshal(entry.Value(), &response); err != nil {
			continue
		}

		// Extract hostname from key: responses.{job-id}.{hostname}.{timestamp}
		keyParts := strings.Split(key, ".")
		if len(keyParts) >= 3 {
			hostname := keyParts[2]
			responses[hostname] = response
		}
	}

	return responses
}
