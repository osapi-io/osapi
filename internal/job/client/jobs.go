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
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"

	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/telemetry/opspan"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// CreateJob creates a new job of jobType from operation data and stores it
// in the KV bucket. The type picks the queue: a query or a modify.
func (c *Client) CreateJob(
	ctx context.Context,
	jobType types.JobType,
	operationData map[string]interface{},
	targetHostname string,
) (*CreateJobResult, error) {
	operationType, ok := operationData["type"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid operation format: missing type field")
	}

	// Default to load-balanced routing if no specific hostname provided
	if targetHostname == "" {
		targetHostname = job.AnyHost
	}

	var prefix string
	switch jobType {
	case types.JobTypeQuery:
		prefix = job.JobsQueryPrefix
	case types.JobTypeModify:
		prefix = job.JobsModifyPrefix
	default:
		return nil, fmt.Errorf("invalid job type %q", jobType)
	}

	notificationSubject := job.BuildSubjectFromTarget(prefix, targetHostname)

	// Generate job ID
	jobID := uuid.New().String()

	// Create job with status metadata
	createdTime := time.Now().Format(time.RFC3339)
	jobWithStatus := map[string]interface{}{
		"id":        jobID,
		"status":    types.JobStatusSubmitted,
		"created":   createdTime,
		"subject":   notificationSubject,
		"target":    targetHostname,
		"operation": operationData,
		"status_history": []interface{}{
			map[string]interface{}{
				"status":    types.JobStatusSubmitted,
				"timestamp": createdTime,
			},
		},
	}

	jobWithStatusJSON, err := json.Marshal(jobWithStatus)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal job with status: %w", err)
	}

	// Sign the job data with the key of the controller.
	kvPayload, err := c.sign(jobWithStatusJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to sign job data: %w", err)
	}

	// Store job in KV with immutable key format
	kvKey := "jobs." + jobID
	c.logger.DebugContext(
		ctx, "storing job and sending notification",
		slog.String("kv_bucket", c.jobsBucket),
		slog.String("key", kvKey),
		slog.String("subject", notificationSubject),
		slog.String("job_id", jobID),
	)

	revision, err := c.putJob(ctx, kvKey, kvPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to store job in KV: %w", err)
	}

	// Write initial submitted status event
	statusKey := fmt.Sprintf("status.%s.submitted._api.%d", jobID, time.Now().UnixNano())
	statusEvent := map[string]interface{}{
		"job_id":    jobID,
		"event":     types.AgentJobStatusSubmitted,
		"hostname":  "_api", // API server hostname could be added here
		"timestamp": createdTime,
		"unix_nano": time.Now().UnixNano(),
		"data": map[string]interface{}{
			"target_hostname": targetHostname,
			"operation_type":  operationType,
		},
	}
	statusEventJSON, _ := json.Marshal(statusEvent)
	if _, err := c.natsClient.KVPut(ctx, c.jobsBucket, statusKey, statusEventJSON); err != nil {
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

	c.logger.DebugContext(
		ctx, "job created successfully",
		slog.String("job_id", jobID),
		slog.Uint64("revision", revision),
		slog.String("subject", notificationSubject),
		slog.String("target_hostname", targetHostname),
	)

	return &CreateJobResult{
		JobID:     jobID,
		Status:    types.JobStatusSubmitted,
		Revision:  revision,
		Timestamp: createdTime,
	}, nil
}

// GetQueueSummary returns job queue statistics derived from KV key names
// only — no entry reads.
func (c *Client) queueSummary(
	ctx context.Context,
) (*job.QueueStats, error) {
	keys, err := c.natsClient.KVKeys(ctx, c.jobsBucket)
	if err != nil {
		return nil, fmt.Errorf("error fetching keys: %w", err)
	}

	_, jobStatuses := computeStatusFromKeyNames(keys)

	statusCounts := map[types.JobStatus]int{
		types.JobStatusSubmitted:      0,
		types.JobStatusProcessing:     0,
		types.JobStatusCompleted:      0,
		types.JobStatusFailed:         0,
		types.JobStatusPartialFailure: 0,
	}

	for _, info := range jobStatuses {
		statusCounts[info.Status]++
	}

	total := len(jobStatuses)

	// DLQ count
	dlqCount := 0
	dlqName := c.streamName + "-DLQ"
	dlqInfo, err := c.natsClient.StreamInfo(ctx, dlqName)
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
func (c *Client) jobStatus(
	ctx context.Context,
	jobID string,
) (*job.QueuedJob, error) {
	// Get the immutable job data
	jobKey := "jobs." + jobID
	c.logger.Debug(
		"kv.get",
		slog.String("key", jobKey),
	)
	entry, err := c.natsClient.KVGet(ctx, c.jobsBucket, jobKey)
	if err != nil {
		return nil, readError("job", jobID, err)
	}

	jobData, err := c.jobRecord(entry.Value())
	if err != nil {
		return nil, err
	}

	// Get status event keys for this job (server-side filtered)
	statusKeys, err := c.natsClient.KVKeys(ctx, c.jobsBucket, "status."+jobID+".>")
	if err != nil {
		return nil, fmt.Errorf("failed to get status events: %w", err)
	}

	// Get response keys for this job (server-side filtered)
	responseKeys, err := c.natsClient.KVKeys(ctx, c.jobsBucket, "responses."+jobID+".>")
	if err != nil {
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
		Target:      getStringField(jobData, "target"),
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
//
// A non-nil keep drops each job whose target it rejects before the jobs are
// counted and paged, so the totals count only the jobs that keep accepts. It
// costs a read of each job, which a nil keep does not.
func (c *Client) listJobs(
	ctx context.Context,
	statusFilter types.JobStatus,
	limit int,
	offset int,
	keep func(target string) bool,
) (*ListJobsResult, error) {
	// Cap limit to MaxPageSize
	if limit <= 0 || limit > MaxPageSize {
		limit = DefaultPageSize
	}

	c.logger.Debug(
		"kv.keys",
		slog.String("operation", "list_jobs"),
		slog.String("status_filter", string(statusFilter)),
		slog.Int("limit", limit),
		slog.Int("offset", offset),
	)

	allKeys, err := c.natsClient.KVKeys(ctx, c.jobsBucket)
	if err != nil {
		return nil, fmt.Errorf("error fetching jobs: %w", err)
	}

	// Pass 1: Light — key names only, no kv.Get()
	orderedJobIDs, jobStatuses := computeStatusFromKeyNames(allKeys)
	if keep != nil {
		orderedJobIDs = c.keepJobs(ctx, orderedJobIDs, keep)
	}

	// Compute status counts from key-name-derived statuses (free — no extra reads)
	statusCounts := map[types.JobStatus]int{
		types.JobStatusSubmitted:      0,
		types.JobStatusProcessing:     0,
		types.JobStatusCompleted:      0,
		types.JobStatusFailed:         0,
		types.JobStatusPartialFailure: 0,
	}
	for _, id := range orderedJobIDs {
		// Jobs with no status events are "submitted" but not in jobStatuses
		info, exists := jobStatuses[id]
		if !exists {
			info = lightJobInfo{Status: types.JobStatusSubmitted}
		}
		statusCounts[info.Status]++
	}

	// Filter + count (fast, no reads)
	var matchingIDs []string
	for _, id := range orderedJobIDs {
		if statusFilter != "" {
			info, exists := jobStatuses[id]
			if !exists {
				info = lightJobInfo{Status: types.JobStatusSubmitted}
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

	c.logger.Debug(
		"kv.keys",
		slog.Int("total_matching", totalCount),
		slog.Int("page_size", len(matchingIDs)),
		slog.Int("all_keys", len(allKeys)),
	)

	// Pass 2: Heavy — kv.Get() only for the page
	jobs := make([]*job.QueuedJob, 0, len(matchingIDs))
	for _, id := range matchingIDs {
		jobKey := "jobs." + id
		jobInfo, err := c.getJobStatusFromKeys(ctx, allKeys, jobKey, id)
		if err != nil {
			// The page shows each job that it counts: a job that cannot be
			// read is listed with its status from the key names and the
			// reason, so it can be seen and deleted.
			c.logger.Warn(
				"failed to read a job for the list",
				slog.String("job_id", id),
				slog.String("error", err.Error()),
			)
			jobInfo = unreadableJob(id, jobStatuses, err)
		}
		jobs = append(jobs, jobInfo)
	}

	return &ListJobsResult{
		Jobs:         jobs,
		TotalCount:   totalCount,
		StatusCounts: statusCounts,
	}, nil
}

// keepJobs is the IDs of ids whose job has a target that keep accepts, in
// the same order. A job that cannot be read is dropped.
func (c *Client) keepJobs(
	ctx context.Context,
	ids []string,
	keep func(target string) bool,
) []string {
	kept := make([]string, 0, len(ids))
	for _, id := range ids {
		entry, err := c.natsClient.KVGet(ctx, c.jobsBucket, "jobs."+id)
		if err != nil {
			continue
		}
		jobData, err := c.jobRecord(entry.Value())
		if err != nil {
			continue
		}
		if keep(getStringField(jobData, "target")) {
			kept = append(kept, id)
		}
	}

	return kept
}

// getJobStatusFromKeys builds a QueuedJob using pre-fetched keys (no inner kv.Keys() call).
func (c *Client) getJobStatusFromKeys(
	ctx context.Context,
	allKeys []string,
	jobKey string,
	jobID string,
) (*job.QueuedJob, error) {
	entry, err := c.natsClient.KVGet(ctx, c.jobsBucket, jobKey)
	if err != nil {
		return nil, readError("job", jobID, err)
	}

	jobData, err := c.jobRecord(entry.Value())
	if err != nil {
		return nil, err
	}

	computedStatus := c.computeStatusFromEvents(ctx, allKeys, jobID)
	responses := c.getJobResponses(ctx, allKeys, jobID)

	queuedJob := &job.QueuedJob{
		ID:          jobID,
		Status:      computedStatus.Status,
		Created:     getStringField(jobData, "created"),
		Subject:     getStringField(jobData, "subject"),
		Target:      getStringField(jobData, "target"),
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
		Status:      types.JobStatusSubmitted, // Default if no events
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

	// Track status by agent. Events are processed in alphabetical key
	// order (e.g., "completed" before "started") so we use priority to
	// ensure only higher-priority events overwrite lower ones.
	agentStates := make(map[string]types.AgentJobStatus)
	agentStartTimes := make(map[string]time.Time)
	agentEndTimes := make(map[string]time.Time)
	agentErrors := make(map[string]string)
	var latestEvent time.Time
	var latestError string
	var timeline []job.TimelineEvent

	// Process each event
	for _, key := range relevantKeys {
		entry, err := c.natsClient.KVGet(ctx, c.jobsBucket, key)
		if err != nil {
			continue
		}

		var event map[string]interface{}
		if err := json.Unmarshal(entry.Value(), &event); err != nil {
			continue
		}

		hostname := getStringField(event, "hostname")
		eventType := types.AgentJobStatus(getStringField(event, "event"))
		timestamp := getStringField(event, "timestamp")

		// Parse timestamp (supports both RFC3339 and RFC3339Nano)
		t, timestampErr := time.Parse(time.RFC3339Nano, timestamp)
		if timestampErr != nil {
			continue
		}

		// Add to timeline
		timelineEvent := job.TimelineEvent{
			Timestamp: t,
			Event:     string(eventType),
			Hostname:  hostname,
		}

		// Set message based on event type
		switch eventType {
		case types.AgentJobStatusSubmitted:
			timelineEvent.Message = "Job submitted to queue"
		case types.AgentJobStatusAcknowledged:
			timelineEvent.Message = fmt.Sprintf("Job acknowledged by agent %s", hostname)
		case types.AgentJobStatusStarted:
			timelineEvent.Message = fmt.Sprintf("Job processing started on %s", hostname)
		case types.AgentJobStatusCompleted:
			timelineEvent.Message = fmt.Sprintf("Job completed successfully on %s", hostname)
		case types.AgentJobStatusSkipped:
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
		case types.AgentJobStatusFailed:
			timelineEvent.Message = fmt.Sprintf("Job failed on %s", hostname)
			if data, ok := event["data"].(map[string]interface{}); ok {
				if errMsg, ok := data["error"].(string); ok {
					timelineEvent.Error = errMsg
					agentErrors[hostname] = errMsg
					latestError = errMsg
				}
			}
		case types.AgentJobStatusRetried:
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

		// Track agent state — only upgrade to higher-priority events.
		if hostname != "" && hostname != "_api" {
			cur, exists := agentStates[hostname]
			if !exists || job.StatusPriority[eventType] > job.StatusPriority[cur] {
				agentStates[hostname] = eventType
			}

			// Track first start time for duration calculation
			if eventType == types.AgentJobStatusStarted {
				if _, exists := agentStartTimes[hostname]; !exists {
					agentStartTimes[hostname] = t
				}
			}

			// Track last end time for duration calculation
			if eventType == types.AgentJobStatusCompleted ||
				eventType == types.AgentJobStatusFailed {
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

	// Sort timeline by timestamp.
	sort.SliceStable(timeline, func(i, j int) bool {
		return timeline[i].Timestamp.Before(timeline[j].Timestamp)
	})
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
		case types.AgentJobStatusCompleted:
			completed++
		case types.AgentJobStatusFailed, types.AgentJobStatusRefused:
			failed++
		case types.AgentJobStatusSkipped:
			skippedCount++
		case types.AgentJobStatusStarted:
			processing++
		case types.AgentJobStatusAcknowledged:
			acknowledged++
		}
	}

	totalAgents := len(agentStates)

	// Determine overall status
	if totalAgents == 0 {
		result.Status = types.JobStatusSubmitted
	} else if processing > 0 || acknowledged > 0 {
		result.Status = types.JobStatusProcessing
	} else if completed+failed+skippedCount == totalAgents {
		if failed > 0 && completed > 0 {
			result.Status = types.JobStatusPartialFailure
		} else if failed > 0 {
			result.Status = types.JobStatusFailed
			result.Error = latestError
		} else if skippedCount > 0 && completed == 0 {
			result.Status = types.JobStatusSkipped
			result.Error = latestError
		} else {
			result.Status = types.JobStatusCompleted
		}
	}

	return result
}

// RetryJob creates a new job using the same operation data as an existing job.
// The original job is preserved. A "retried" status event is written to the
// original job's timeline linking to the new job.
func (c *Client) retryJob(
	ctx context.Context,
	jobID string,
	targetHostname string,
) (*CreateJobResult, error) {
	// Read original job from KV
	jobKey := "jobs." + jobID
	c.logger.Debug(
		"kv.get",
		slog.String("key", jobKey),
		slog.String("operation", "retry_job"),
	)

	entry, err := c.natsClient.KVGet(ctx, c.jobsBucket, jobKey)
	if err != nil {
		return nil, readError("job", jobID, err)
	}

	jobData, err := c.jobRecord(entry.Value())
	if err != nil {
		return nil, err
	}

	// Extract operation data
	operationData, ok := jobData["operation"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("job %s: %w", jobID, job.ErrNoOperationData)
	}
	// The secrets of a sealed job are sealed to its own job ID, and the
	// controller does not keep their values, so a new job cannot carry
	// them.
	if _, sealed := operationData["sealed_keys"]; sealed {
		return nil, fmt.Errorf("job %s: %w", jobID, job.ErrSealedJob)
	}
	// The new job goes to the queue of the old one, which its subject
	// names, not to a queue that the operation's name suggests.
	jobType, err := queueOf(getStringField(jobData, "subject"))
	if err != nil {
		return nil, fmt.Errorf("job %s: %w", jobID, err)
	}

	// Create new job with the same operation data
	result, err := c.CreateJob(ctx, jobType, operationData, targetHostname)
	if err != nil {
		return nil, fmt.Errorf("failed to create retry job: %w", err)
	}

	// Write "retried" status event on the original job's timeline
	retriedKey := fmt.Sprintf("status.%s.retried._api.%d", jobID, time.Now().UnixNano())
	retriedEvent := map[string]interface{}{
		"job_id":    jobID,
		"event":     types.AgentJobStatusRetried,
		"hostname":  "_api",
		"timestamp": time.Now().Format(time.RFC3339Nano),
		"unix_nano": time.Now().UnixNano(),
		"data": map[string]interface{}{
			"new_job_id":      result.JobID,
			"target_hostname": targetHostname,
		},
	}
	retriedEventJSON, _ := json.Marshal(retriedEvent)
	if _, err := c.natsClient.KVPut(ctx, c.jobsBucket, retriedKey, retriedEventJSON); err != nil {
		c.logger.Error(
			"failed to write retried event",
			slog.String("error", err.Error()),
		)
	}

	c.logger.Debug(
		"job retried successfully",
		slog.String("original_job_id", jobID),
		slog.String("new_job_id", result.JobID),
		slog.String("target_hostname", targetHostname),
	)

	return result, nil
}

// DeleteJob deletes a job from the KV store by its ID.
func (c *Client) deleteJob(
	ctx context.Context,
	jobID string,
) error {
	jobKey := "jobs." + jobID
	c.logger.Debug(
		"kv.delete",
		slog.String("key", jobKey),
	)
	_, err := c.natsClient.KVGet(ctx, c.jobsBucket, jobKey)
	if err != nil {
		return readError("job", jobID, err)
	}

	if err := c.natsClient.KVDelete(ctx, c.jobsBucket, jobKey); err != nil {
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
	// agentStates[jobID][hostname] = highest-priority event
	agentStates := make(map[string]map[string]types.AgentJobStatus)

	for _, key := range keys {
		if !strings.HasPrefix(key, "status.") {
			continue
		}
		parts := strings.SplitN(key, ".", 5)
		if len(parts) < 4 {
			continue
		}
		jobID := parts[1]
		event := types.AgentJobStatus(parts[2])
		hostname := parts[3]

		if _, exists := agentStates[jobID]; !exists {
			agentStates[jobID] = make(map[string]types.AgentJobStatus)
		}

		cur, exists := agentStates[jobID][hostname]
		if !exists || job.StatusPriority[event] > job.StatusPriority[cur] {
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
			case types.AgentJobStatusCompleted, types.AgentJobStatusRetried:
				completed++
			case types.AgentJobStatusFailed, types.AgentJobStatusRefused:
				failed++
			case types.AgentJobStatusSkipped:
				skipped++
			case types.AgentJobStatusStarted:
				processing++
			case types.AgentJobStatusAcknowledged:
				acknowledged++
			}
		}

		totalAgents := completed + failed + skipped + processing + acknowledged

		var status types.JobStatus
		switch {
		case totalAgents == 0:
			status = types.JobStatusSubmitted
		case processing > 0 || acknowledged > 0:
			status = types.JobStatusProcessing
		case failed > 0 && completed > 0:
			status = types.JobStatusPartialFailure
		case failed > 0:
			status = types.JobStatusFailed
		case skipped > 0 && completed == 0:
			status = types.JobStatusSkipped
		default:
			status = types.JobStatusCompleted
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

		entry, err := c.natsClient.KVGet(ctx, c.jobsBucket, key)
		if err != nil {
			continue
		}

		// A response that does not verify against the stored key of the
		// agent is not reported.
		response, verifyErr := c.verifyResponse(ctx, entry.Value())
		if verifyErr != nil {
			continue
		}

		// Extract hostname from key: responses.{job-id}.{hostname}.{timestamp}
		keyParts := strings.Split(key, ".")
		if len(keyParts) >= 3 {
			hostname := keyParts[2]
			responses[hostname] = *response
		}
	}

	return responses
}

// jobRecord is the stored record of a job. The controller writes the
// record inside a signed envelope, so the record is the envelope's
// payload, its signature checked with the controller's key. A job client
// with no signer refuses each record.
func (c *Client) jobRecord(
	value []byte,
) (map[string]interface{}, error) {
	if c.pkiSigner == nil {
		return nil, fmt.Errorf("failed to verify job data: %w", ErrNoSigner)
	}
	payload, err := unwrapStoredEnvelope(value, c.pkiSigner.StoredRecordKeys())
	if err != nil {
		return nil, fmt.Errorf("failed to verify job data: %w", err)
	}
	value = payload

	var jobData map[string]interface{}
	if err := json.Unmarshal(value, &jobData); err != nil {
		return nil, fmt.Errorf("failed to parse job data: %w", err)
	}

	return jobData, nil
}

// GetQueueSummary returns the counts of the queue, in a job.summary span.
func (c *Client) GetQueueSummary(
	ctx context.Context,
) (*job.QueueStats, error) {
	ctx, span := opspan.Start(ctx, tracerName, "job.summary")
	stats, err := c.queueSummary(ctx)

	return stats, opspan.End(span, err)
}

// GetJobStatus returns one job and its computed status, in a job.get span.
func (c *Client) GetJobStatus(
	ctx context.Context,
	jobID string,
) (*job.QueuedJob, error) {
	ctx, span := opspan.Start(ctx, tracerName, "job.get", attribute.String("job.id", jobID))
	qj, err := c.jobStatus(ctx, jobID)

	return qj, opspan.End(span, err)
}

// ListJobs returns a page of jobs, in a job.list span that records how many
// jobs the page has. A non-nil keep drops each job whose target it rejects
// before the jobs are counted and paged.
func (c *Client) ListJobs(
	ctx context.Context,
	statusFilter types.JobStatus,
	limit int,
	offset int,
	keep func(target string) bool,
) (*ListJobsResult, error) {
	ctx, span := opspan.Start(ctx, tracerName, "job.list",
		attribute.String("job.status_filter", string(statusFilter)),
	)
	result, err := c.listJobs(ctx, statusFilter, limit, offset, keep)
	if err == nil {
		span.SetAttributes(attribute.Int("job.count", len(result.Jobs)))
	}

	return result, opspan.End(span, err)
}

// RetryJob creates a job with the operation of another, in a job.retry span.
func (c *Client) RetryJob(
	ctx context.Context,
	jobID string,
	targetHostname string,
) (*CreateJobResult, error) {
	ctx, span := opspan.Start(ctx, tracerName, "job.retry", attribute.String("job.id", jobID))
	result, err := c.retryJob(ctx, jobID, targetHostname)

	return result, opspan.End(span, err)
}

// DeleteJob deletes a job and its events, in a job.delete span.
func (c *Client) DeleteJob(
	ctx context.Context,
	jobID string,
) error {
	ctx, span := opspan.Start(ctx, tracerName, "job.delete", attribute.String("job.id", jobID))

	return opspan.End(span, c.deleteJob(ctx, jobID))
}

// unreadableJob is the entry of a job that the list counts but cannot
// read: its ID, its status from the key names, and why it cannot be read.
func unreadableJob(
	id string,
	statuses map[string]lightJobInfo,
	err error,
) *job.QueuedJob {
	status := types.JobStatusSubmitted
	if info, ok := statuses[id]; ok {
		status = info.Status
	}

	return &job.QueuedJob{
		ID:     id,
		Status: status,
		Error:  "the job's record cannot be read: " + err.Error(),
	}
}

// queueOf is the job type of the queue that subject is on.
func queueOf(
	subject string,
) (types.JobType, error) {
	switch {
	case strings.HasPrefix(subject, job.JobsQueryPrefix+"."):
		return types.JobTypeQuery, nil
	case strings.HasPrefix(subject, job.JobsModifyPrefix+"."):
		return types.JobTypeModify, nil
	default:
		return "", fmt.Errorf("%q: %w", subject, job.ErrUnknownQueue)
	}
}
