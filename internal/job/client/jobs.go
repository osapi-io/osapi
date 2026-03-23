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
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go/jetstream"

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
		strings.HasPrefix(operationType, "node.") {
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

	revision, err := c.kv.Put(ctx, kvKey, jobWithStatusJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to store job in KV: %w", err)
	}

	// Write initial submitted status event
	statusKey := fmt.Sprintf("status.%s.submitted._api.%d", jobID, time.Now().UnixNano())
	statusEvent := map[string]interface{}{
		"job_id":    jobID,
		"event":     string(job.StatusSubmitted),
		"hostname":  "_api", // API server hostname could be added here
		"timestamp": createdTime,
		"unix_nano": time.Now().UnixNano(),
		"data": map[string]interface{}{
			"target_hostname": targetHostname,
			"operation_type":  operationType,
		},
	}
	statusEventJSON, _ := json.Marshal(statusEvent)
	if _, err := c.kv.Put(ctx, statusKey, statusEventJSON); err != nil {
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

// GetQueueSummary returns job queue statistics derived from KV key names
// only — no entry reads.
func (c *Client) GetQueueSummary(
	ctx context.Context,
) (*job.QueueStats, error) {
	keys, err := c.kv.Keys(ctx)
	if err != nil {
		if errors.Is(err, jetstream.ErrNoKeysFound) {
			return &job.QueueStats{
				StatusCounts: map[string]int{
					string(job.StatusSubmitted):  0,
					string(job.StatusProcessing): 0,
					string(job.StatusCompleted):  0,
					string(job.StatusFailed):     0,
					string(job.StatusSkipped):    0,
				},
			}, nil
		}
		return nil, fmt.Errorf("error fetching keys: %w", err)
	}

	_, jobStatuses := computeStatusFromKeyNames(keys)

	statusCounts := map[string]int{
		string(job.StatusSubmitted):      0,
		string(job.StatusProcessing):     0,
		string(job.StatusCompleted):      0,
		string(job.StatusFailed):         0,
		string(job.StatusPartialFailure): 0,
	}

	for _, info := range jobStatuses {
		statusCounts[info.Status]++
	}

	total := len(jobStatuses)

	// DLQ count
	dlqCount := 0
	dlqName := c.streamName + "-DLQ"
	dlqInfo, err := c.natsClient.GetStreamInfo(ctx, dlqName)
	if err == nil {
		dlqCount = int(dlqInfo.State.Msgs)
	}

	return &job.QueueStats{
		TotalJobs:    total,
		StatusCounts: statusCounts,
		DLQCount:     dlqCount,
	}, nil
}

// GetJobStatus returns information about a specific job.
func (c *Client) GetJobStatus(
	ctx context.Context,
	jobID string,
) (*job.QueuedJob, error) {
	// Get the immutable job data
	jobKey := "jobs." + jobID
	c.logger.Debug("kv.get",
		slog.String("key", jobKey),
	)
	entry, err := c.kv.Get(ctx, jobKey)
	if err != nil {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	var jobData map[string]interface{}
	if err := json.Unmarshal(entry.Value(), &jobData); err != nil {
		return nil, fmt.Errorf("failed to parse job data: %w", err)
	}

	// Get status event keys for this job (server-side filtered)
	statusKeys, err := collectKeys(ctx, c.kv, "status."+jobID+".>")
	if err != nil && !errors.Is(err, jetstream.ErrNoKeysFound) {
		return nil, fmt.Errorf("failed to get status events: %w", err)
	}

	// Get response keys for this job (server-side filtered)
	responseKeys, err := collectKeys(ctx, c.kv, "responses."+jobID+".>")
	if err != nil && !errors.Is(err, jetstream.ErrNoKeysFound) {
		return nil, fmt.Errorf("failed to get response keys: %w", err)
	}

	// Compute current status from events
	computedStatus := c.computeStatusFromEvents(ctx, statusKeys, jobID)

	// Get response data for this job
	responses := c.getJobResponses(ctx, responseKeys, jobID)

	queuedJob := &job.QueuedJob{
		ID:          jobID,
		Status:      computedStatus.Status,
		Created:     getStringField(jobData, "created"),
		Subject:     getStringField(jobData, "subject"),
		Hostname:    computedStatus.Hostname,
		Error:       computedStatus.Error,
		UpdatedAt:   computedStatus.UpdatedAt,
		AgentStates: computedStatus.AgentStates,
		Timeline:    computedStatus.Timeline,
		Responses:   responses,
	}

	if operation, ok := jobData["operation"].(map[string]interface{}); ok {
		queuedJob.Operation = operation
	}

	// Populate Result and Changed from responses.
	switch len(responses) {
	case 1:
		for _, resp := range responses {
			queuedJob.Result = resp.Data
			queuedJob.Changed = resp.Changed
		}
	default:
		// Broadcast: aggregate Changed as OR of all per-host values.
		anyChanged := false
		for _, resp := range responses {
			if resp.Changed != nil && *resp.Changed {
				anyChanged = true

				break
			}
		}

		if anyChanged {
			queuedJob.Changed = &anyChanged
		}
	}

	return queuedJob, nil
}

// ListJobs returns jobs filtered by status with server-side pagination.
// Uses a two-pass approach: Pass 1 derives status from key names only (fast),
// Pass 2 fetches full details for the paginated page only.
// Jobs are returned newest-first (reverse insertion order).
func (c *Client) ListJobs(
	ctx context.Context,
	statusFilter string,
	limit int,
	offset int,
) (*ListJobsResult, error) {
	// Cap limit to MaxPageSize
	if limit <= 0 || limit > MaxPageSize {
		limit = DefaultPageSize
	}

	c.logger.Debug("kv.keys",
		slog.String("operation", "list_jobs"),
		slog.String("status_filter", statusFilter),
		slog.Int("limit", limit),
		slog.Int("offset", offset),
	)

	allKeys, err := c.kv.Keys(ctx)
	if err != nil {
		if errors.Is(err, jetstream.ErrNoKeysFound) {
			return &ListJobsResult{
				Jobs:       []*job.QueuedJob{},
				TotalCount: 0,
			}, nil
		}
		return nil, fmt.Errorf("error fetching jobs: %w", err)
	}

	// Pass 1: Light — key names only, no kv.Get()
	orderedJobIDs, jobStatuses := computeStatusFromKeyNames(allKeys)

	// Compute status counts from key-name-derived statuses (free — no extra reads)
	statusCounts := map[string]int{
		string(job.StatusSubmitted):      0,
		string(job.StatusProcessing):     0,
		string(job.StatusCompleted):      0,
		string(job.StatusFailed):         0,
		string(job.StatusPartialFailure): 0,
	}
	for _, info := range jobStatuses {
		statusCounts[info.Status]++
	}
	// Jobs with no status events are "submitted" but not in jobStatuses
	submittedOnly := len(orderedJobIDs) - len(jobStatuses)
	if submittedOnly > 0 {
		statusCounts[string(job.StatusSubmitted)] += submittedOnly
	}

	// Filter + count (fast, no reads)
	var matchingIDs []string
	for _, id := range orderedJobIDs {
		if statusFilter != "" {
			info, exists := jobStatuses[id]
			if !exists {
				info = lightJobInfo{Status: string(job.StatusSubmitted)}
			}
			if info.Status != statusFilter {
				continue
			}
		}
		matchingIDs = append(matchingIDs, id)
	}
	totalCount := len(matchingIDs)

	// Apply offset
	if offset >= len(matchingIDs) {
		return &ListJobsResult{
			Jobs:         []*job.QueuedJob{},
			TotalCount:   totalCount,
			StatusCounts: statusCounts,
		}, nil
	}
	matchingIDs = matchingIDs[offset:]

	// Apply limit
	if len(matchingIDs) > limit {
		matchingIDs = matchingIDs[:limit]
	}

	c.logger.Debug("kv.keys",
		slog.Int("total_matching", totalCount),
		slog.Int("page_size", len(matchingIDs)),
		slog.Int("all_keys", len(allKeys)),
	)

	// Pass 2: Heavy — kv.Get() only for the page
	var jobs []*job.QueuedJob
	for _, id := range matchingIDs {
		jobKey := "jobs." + id
		jobInfo, err := c.getJobStatusFromKeys(ctx, allKeys, jobKey, id)
		if err != nil {
			c.logger.Debug("failed to get job status during list",
				slog.String("job_id", id),
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
		Jobs:         jobs,
		TotalCount:   totalCount,
		StatusCounts: statusCounts,
	}, nil
}

// getJobStatusFromKeys builds a QueuedJob using pre-fetched keys (no inner kv.Keys() call).
func (c *Client) getJobStatusFromKeys(
	ctx context.Context,
	allKeys []string,
	jobKey string,
	jobID string,
) (*job.QueuedJob, error) {
	entry, err := c.kv.Get(ctx, jobKey)
	if err != nil {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	var jobData map[string]interface{}
	if err := json.Unmarshal(entry.Value(), &jobData); err != nil {
		return nil, fmt.Errorf("failed to parse job data: %w", err)
	}

	computedStatus := c.computeStatusFromEvents(ctx, allKeys, jobID)
	responses := c.getJobResponses(ctx, allKeys, jobID)

	queuedJob := &job.QueuedJob{
		ID:          jobID,
		Status:      computedStatus.Status,
		Created:     getStringField(jobData, "created"),
		Subject:     getStringField(jobData, "subject"),
		Hostname:    computedStatus.Hostname,
		Error:       computedStatus.Error,
		UpdatedAt:   computedStatus.UpdatedAt,
		AgentStates: computedStatus.AgentStates,
		Timeline:    computedStatus.Timeline,
		Responses:   responses,
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

// collectKeys wraps ListKeysFiltered and drains the KeyLister channel into a []string.
func collectKeys(
	ctx context.Context,
	kv jetstream.KeyValue,
	filter string,
) ([]string, error) {
	lister, err := kv.ListKeysFiltered(ctx, filter)
	if err != nil {
		return nil, err
	}

	var keys []string
	for key := range lister.Keys() {
		keys = append(keys, key)
	}

	return keys, nil
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
	ctx context.Context,
	eventKeys []string,
	jobID string,
) computedJobStatus {
	result := computedJobStatus{
		Status:      string(job.StatusSubmitted), // Default if no events
		AgentStates: make(map[string]job.AgentState),
		Timeline:    []job.TimelineEvent{},
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

	// Track status by agent
	agentStates := make(map[string]string)
	agentStartTimes := make(map[string]time.Time)
	agentEndTimes := make(map[string]time.Time)
	agentErrors := make(map[string]string)
	var latestEvent time.Time
	var latestError string
	var timeline []job.TimelineEvent

	// Process each event
	for _, key := range relevantKeys {
		entry, err := c.kv.Get(ctx, key)
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
		case string(job.StatusSubmitted):
			timelineEvent.Message = "Job submitted to queue"
		case string(job.StatusAcknowledged):
			timelineEvent.Message = fmt.Sprintf("Job acknowledged by agent %s", hostname)
		case string(job.StatusStarted):
			timelineEvent.Message = fmt.Sprintf("Job processing started on %s", hostname)
		case string(job.StatusCompleted):
			timelineEvent.Message = fmt.Sprintf("Job completed successfully on %s", hostname)
		case string(job.StatusSkipped):
			timelineEvent.Message = fmt.Sprintf(
				"Job skipped on %s (unsupported OS family)",
				hostname,
			)
			if data, ok := event["data"].(map[string]interface{}); ok {
				if errMsg, ok := data["error"].(string); ok {
					timelineEvent.Error = errMsg
					agentErrors[hostname] = errMsg
					latestError = errMsg
				}
			}
		case string(job.StatusFailed):
			timelineEvent.Message = fmt.Sprintf("Job failed on %s", hostname)
			if data, ok := event["data"].(map[string]interface{}); ok {
				if errMsg, ok := data["error"].(string); ok {
					timelineEvent.Error = errMsg
					agentErrors[hostname] = errMsg
					latestError = errMsg
				}
			}
		case string(job.StatusRetried):
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

		// Track agent state
		if hostname != "" && hostname != "_api" {
			agentStates[hostname] = eventType

			// Track first start time for duration calculation
			if eventType == string(job.StatusStarted) {
				if _, exists := agentStartTimes[hostname]; !exists {
					agentStartTimes[hostname] = t
				}
			}

			// Track last end time for duration calculation
			if eventType == string(job.StatusCompleted) || eventType == string(job.StatusFailed) {
				agentEndTimes[hostname] = t
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

	// Build agent states with detailed information
	for hostname, state := range agentStates {
		agentState := job.AgentState{
			Status: state,
		}

		// Add error if agent failed
		if errorMsg, hasError := agentErrors[hostname]; hasError {
			agentState.Error = errorMsg
		}

		// Calculate duration from first start to last end
		if startTime, hasStart := agentStartTimes[hostname]; hasStart {
			agentState.StartTime = startTime

			if endTime, hasEnd := agentEndTimes[hostname]; hasEnd {
				agentState.EndTime = endTime
				duration := endTime.Sub(startTime)
				agentState.Duration = duration.String()
			}
		}

		result.AgentStates[hostname] = agentState
	}

	// Compute overall status based on agent states
	completed := 0
	failed := 0
	skippedCount := 0
	processing := 0
	acknowledged := 0

	for _, state := range agentStates {
		switch state {
		case string(job.StatusCompleted):
			completed++
		case string(job.StatusFailed):
			failed++
		case string(job.StatusSkipped):
			skippedCount++
		case string(job.StatusStarted):
			processing++
		case string(job.StatusAcknowledged):
			acknowledged++
		}
	}

	totalAgents := len(agentStates)

	// Determine overall status
	if totalAgents == 0 {
		result.Status = string(job.StatusSubmitted)
	} else if processing > 0 || acknowledged > 0 {
		result.Status = string(job.StatusProcessing)
	} else if completed+failed+skippedCount == totalAgents {
		if failed > 0 && completed > 0 {
			result.Status = string(job.StatusPartialFailure)
		} else if failed > 0 {
			result.Status = string(job.StatusFailed)
			result.Error = latestError
		} else if skippedCount > 0 && completed == 0 {
			result.Status = string(job.StatusSkipped)
			result.Error = latestError
		} else {
			result.Status = string(job.StatusCompleted)
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

	entry, err := c.kv.Get(ctx, jobKey)
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
		"event":     string(job.StatusRetried),
		"hostname":  "_api",
		"timestamp": time.Now().Format(time.RFC3339Nano),
		"unix_nano": time.Now().UnixNano(),
		"data": map[string]interface{}{
			"new_job_id":      result.JobID,
			"target_hostname": targetHostname,
		},
	}
	retriedEventJSON, _ := json.Marshal(retriedEvent)
	if _, err := c.kv.Put(ctx, retriedKey, retriedEventJSON); err != nil {
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
	ctx context.Context,
	jobID string,
) error {
	jobKey := "jobs." + jobID
	c.logger.Debug("kv.delete",
		slog.String("key", jobKey),
	)
	_, err := c.kv.Get(ctx, jobKey)
	if err != nil {
		return fmt.Errorf("job not found: %s", jobID)
	}

	if err := c.kv.Delete(ctx, jobKey); err != nil {
		return fmt.Errorf("failed to delete job: %w", err)
	}

	return nil
}

// computeStatusFromKeyNames derives job statuses from KV key names only — no
// kv.Get() calls. It returns ordered job IDs (newest-first from "jobs.*" keys)
// and a map of per-job status computed using the same multi-agent logic as
// computeStatusFromEvents but parsed entirely from key names.
//
// Key format: status.<jobID>.<event>.<hostname>.<timestamp>
func computeStatusFromKeyNames(
	keys []string,
) ([]string, map[string]lightJobInfo) {
	// Extract job IDs from jobs.* keys
	var orderedJobIDs []string
	for _, key := range keys {
		if strings.HasPrefix(key, "jobs.") {
			jobID := strings.TrimPrefix(key, "jobs.")
			if jobID != "" {
				orderedJobIDs = append(orderedJobIDs, jobID)
			}
		}
	}

	// Reverse for newest-first (KV insertion order is chronological)
	for i, j := 0, len(orderedJobIDs)-1; i < j; i, j = i+1, j-1 {
		orderedJobIDs[i], orderedJobIDs[j] = orderedJobIDs[j], orderedJobIDs[i]
	}

	// Parse status keys and track highest-priority event per (jobID, hostname)
	statusPriority := map[string]int{
		string(job.StatusSubmitted):    0,
		string(job.StatusAcknowledged): 1,
		string(job.StatusStarted):      2,
		string(job.StatusFailed):       3,
		string(job.StatusSkipped):      3,
		string(job.StatusCompleted):    4,
		string(job.StatusRetried):      4,
	}

	// agentStates[jobID][hostname] = highest-priority event
	agentStates := make(map[string]map[string]string)

	for _, key := range keys {
		if !strings.HasPrefix(key, "status.") {
			continue
		}
		parts := strings.SplitN(key, ".", 5)
		if len(parts) < 4 {
			continue
		}
		jobID := parts[1]
		event := parts[2]
		hostname := parts[3]

		if _, exists := agentStates[jobID]; !exists {
			agentStates[jobID] = make(map[string]string)
		}

		cur, exists := agentStates[jobID][hostname]
		if !exists || statusPriority[event] > statusPriority[cur] {
			agentStates[jobID][hostname] = event
		}
	}

	// Compute overall status per job using multi-agent logic
	jobStatuses := make(map[string]lightJobInfo, len(agentStates))

	for jobID, agents := range agentStates {
		completed := 0
		failed := 0
		skipped := 0
		processing := 0
		acknowledged := 0

		for hostname, state := range agents {
			if hostname == "_api" {
				continue
			}
			switch state {
			case string(job.StatusCompleted), string(job.StatusRetried):
				completed++
			case string(job.StatusFailed):
				failed++
			case string(job.StatusSkipped):
				skipped++
			case string(job.StatusStarted):
				processing++
			case string(job.StatusAcknowledged):
				acknowledged++
			}
		}

		totalAgents := completed + failed + skipped + processing + acknowledged

		var status string
		switch {
		case totalAgents == 0:
			status = string(job.StatusSubmitted)
		case processing > 0 || acknowledged > 0:
			status = string(job.StatusProcessing)
		case failed > 0 && completed > 0:
			status = string(job.StatusPartialFailure)
		case failed > 0:
			status = string(job.StatusFailed)
		case skipped > 0 && completed == 0:
			status = string(job.StatusSkipped)
		default:
			status = string(job.StatusCompleted)
		}

		jobStatuses[jobID] = lightJobInfo{Status: status}
	}

	return orderedJobIDs, jobStatuses
}

// getJobResponses retrieves response data for a specific job
func (c *Client) getJobResponses(
	ctx context.Context,
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

		entry, err := c.kv.Get(ctx, key)
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
