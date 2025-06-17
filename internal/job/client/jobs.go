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

	// Build the notification subject using simplified routing based on operation semantics
	var notificationSubject string

	// Route based on operation name - operations ending in common read patterns are queries
	if strings.HasSuffix(operationType, ".get") ||
		strings.HasSuffix(operationType, ".query") ||
		strings.HasSuffix(operationType, ".read") ||
		strings.HasSuffix(operationType, ".status") ||
		strings.HasSuffix(operationType, ".do") ||
		strings.HasPrefix(operationType, "system.") {
		notificationSubject = job.BuildQuerySubject(targetHostname)
	} else {
		notificationSubject = job.BuildModifySubject(targetHostname)
	}

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
	c.logger.Debug("storing job and sending notification",
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
		c.logger.Error("failed to write submitted event", slog.String("error", err.Error()))
	}

	// Send notification via stream
	err = c.natsClient.Publish(ctx, notificationSubject, []byte(jobID))
	if err != nil {
		return nil, fmt.Errorf("failed to send notification: %w", err)
	}

	c.logger.Debug("job created successfully",
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

		// Get status events for this job
		statusKeys, _ := c.kv.Keys()

		// Compute status from events
		computedStatus := c.computeStatusFromEvents(statusKeys, jobID)
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
		Error:        computedStatus.Error,
		UpdatedAt:    computedStatus.UpdatedAt,
		WorkerStates: computedStatus.WorkerStates,
		Timeline:     computedStatus.Timeline,
		Responses:    responses,
	}

	if operation, ok := jobData["operation"].(map[string]interface{}); ok {
		queuedJob.Operation = operation
	}

	return queuedJob, nil
}

// ListJobs returns jobs filtered by status.
func (c *Client) ListJobs(
	ctx context.Context,
	statusFilter string,
) ([]*job.QueuedJob, error) {
	keys, err := c.kv.Keys()
	if err != nil {
		if err.Error() == "nats: no keys found" {
			return []*job.QueuedJob{}, nil
		}
		return nil, fmt.Errorf("error fetching jobs: %w", err)
	}

	var jobs []*job.QueuedJob

	for _, key := range keys {
		// Only process job keys (format: jobs.{job-id})
		if !strings.HasPrefix(key, "jobs.") {
			continue
		}

		// Extract job ID from job key
		jobID := strings.TrimPrefix(key, "jobs.")
		if jobID == "" {
			continue
		}

		// Get the full job status (including computed status from events)
		jobInfo, err := c.GetJobStatus(ctx, jobID)
		if err != nil {
			c.logger.Debug("failed to get job status during list",
				slog.String("job_id", jobID),
				slog.String("error", err.Error()),
			)
			continue
		}

		// Apply status filter if specified
		if statusFilter != "" && jobInfo.Status != statusFilter {
			continue
		}

		jobs = append(jobs, jobInfo)
	}

	return jobs, nil
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

		// Parse timestamp
		t, timestampErr := time.Parse(time.RFC3339, timestamp)
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
		}

		timeline = append(timeline, timelineEvent)

		// Track worker state
		if hostname != "" && hostname != "_api" {
			workerStates[hostname] = eventType

			// Track start times for duration calculation
			if eventType == "started" {
				workerStartTimes[hostname] = t
			}

			// Track the processing hostname
			if eventType == "completed" || eventType == "failed" {
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

		// Calculate duration if we have start time
		if startTime, hasStart := workerStartTimes[hostname]; hasStart {
			workerState.StartTime = startTime

			// Find end time from timeline events
			for _, event := range timeline {
				if event.Hostname == hostname &&
					(event.Event == "completed" || event.Event == "failed") {
					workerState.EndTime = event.Timestamp
					duration := event.Timestamp.Sub(startTime)
					workerState.Duration = duration.String()
					break
				}
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
