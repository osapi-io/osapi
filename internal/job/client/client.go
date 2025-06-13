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
	"github.com/nats-io/nats.go"
	natsclient "github.com/osapi-io/nats-client/pkg/client"

	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/messaging"
)

// Client provides methods for publishing job requests and retrieving responses.
type Client struct {
	logger     *slog.Logger
	natsClient messaging.NATSClient
	kv         nats.KeyValue
	timeout    time.Duration
}

// Options configures the jobs client.
type Options struct {
	// Timeout for waiting for job responses (default: 30s)
	Timeout time.Duration
	// KVBucket for job storage (required)
	KVBucket nats.KeyValue
}

// New creates a new jobs client using an existing NATS client.
func New(
	logger *slog.Logger,
	natsClient messaging.NATSClient,
	opts *Options,
) (*Client, error) {
	if opts == nil {
		return nil, fmt.Errorf("options cannot be nil")
	}
	if opts.KVBucket == nil {
		return nil, fmt.Errorf("KVBucket cannot be nil")
	}

	return &Client{
		logger:     logger,
		natsClient: natsClient,
		kv:         opts.KVBucket,
		timeout:    opts.Timeout,
	}, nil
}

// publishAndWait publishes a job request and waits for the response.
func (c *Client) publishAndWait(
	ctx context.Context,
	subject string,
	req *job.Request,
) (*job.Response, error) {
	// Generate request ID if not provided
	if req.RequestID == "" {
		req.RequestID = uuid.New().String()
	}
	req.Timestamp = time.Now()

	// Marshal request
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	c.logger.Info("publishing job request",
		slog.String("request_id", req.RequestID),
		slog.String("subject", subject),
		slog.String("type", string(req.Type)),
	)

	// Use nats-client's PublishAndWaitKV
	opts := &natsclient.RequestReplyOptions{
		RequestID: req.RequestID,
		Timeout:   c.timeout,
	}

	responseData, err := c.natsClient.PublishAndWaitKV(ctx, subject, data, c.kv, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to publish and wait: %w", err)
	}

	// Unmarshal response
	var response job.Response
	if err := json.Unmarshal(responseData, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	c.logger.Info("received job response",
		slog.String("request_id", req.RequestID),
		slog.String("status", string(response.Status)),
	)

	return &response, nil
}

// CreateJobResult represents the result of creating a job.
type CreateJobResult struct {
	JobID     string `json:"job_id"`
	Status    string `json:"status"`
	Revision  uint64 `json:"revision"`
	Timestamp string `json:"timestamp"`
}

// CreateJob creates a new job from operation data and stores it in the KV bucket.
func (c *Client) CreateJob(
	ctx context.Context,
	operationData map[string]interface{},
	targetHostname string,
) (*CreateJobResult, error) {
	// Extract operation type for subject routing
	operationType, ok := operationData["type"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid operation format: missing type field")
	}

	// Parse operation type to extract category and operation
	// Expected format: "category.operation" (e.g., "system.hostname", "network.dns")
	parts := strings.Split(operationType, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid operation type format: %s", operationType)
	}

	category := parts[0]
	operation := strings.Join(parts[1:], ".")

	// Default to load-balanced routing if no specific hostname provided
	if targetHostname == "" {
		targetHostname = job.AnyHost
	}

	// Build the notification subject using jobs routing based on operation semantics
	var notificationSubject string

	// Route based on operation name - operations ending in common read patterns are queries
	if category == "system" ||
		strings.HasSuffix(operation, ".get") ||
		strings.HasSuffix(operation, ".query") ||
		strings.HasSuffix(operation, ".read") ||
		strings.HasSuffix(operation, ".status") ||
		strings.HasSuffix(operation, ".do") {
		notificationSubject = job.BuildQuerySubject(targetHostname, category, operation)
	} else {
		notificationSubject = job.BuildModifySubject(targetHostname, category, operation)
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

	// Store job in KV with status prefix
	kvKey := "unprocessed." + jobID
	c.logger.Info("storing job and sending notification",
		slog.String("kv_bucket", c.kv.Bucket()),
		slog.String("key", kvKey),
		slog.String("subject", notificationSubject),
		slog.String("job_id", jobID),
	)

	revision, err := c.kv.Put(kvKey, jobWithStatusJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to store job in KV: %w", err)
	}

	// Send notification via stream
	if nc, ok := c.natsClient.(*natsclient.Client); ok {
		_, err = nc.ExtJS.Publish(ctx, notificationSubject, []byte(jobID))
		if err != nil {
			return nil, fmt.Errorf("failed to send notification: %w", err)
		}
	} else {
		return nil, fmt.Errorf("failed to access JetStream: natsClient is not of expected type")
	}

	c.logger.Info("job created successfully",
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
					"unprocessed": 0,
					"processing":  0,
					"completed":   0,
					"failed":      0,
				},
				OperationCounts: map[string]int{},
				DLQCount:        0,
			}, nil
		}
		return nil, fmt.Errorf("error fetching jobs: %w", err)
	}

	statusCounts := map[string]int{
		"unprocessed": 0,
		"processing":  0,
		"completed":   0,
		"failed":      0,
	}

	operationCounts := map[string]int{}
	uniqueJobIDs := make(map[string]bool)
	jobOperations := make(map[string]string) // jobUUID -> operationType

	// Count jobs by status and operation type
	for _, key := range keys {
		// Extract job UUID from key (format: "status.uuid")
		keyParts := strings.SplitN(key, ".", 2)
		if len(keyParts) != 2 {
			continue // Skip malformed keys
		}
		jobUUID := keyParts[1]

		entry, err := c.kv.Get(key)
		if err != nil {
			continue // Skip jobs we can't read
		}

		var jobData map[string]interface{}
		if err := json.Unmarshal(entry.Value(), &jobData); err != nil {
			continue // Skip jobs we can't parse
		}

		// Track unique job IDs
		uniqueJobIDs[jobUUID] = true

		// Count by status
		if status, ok := jobData["status"].(string); ok {
			statusCounts[status]++
		}

		// Track operation type for each unique job (avoid double-counting)
		if jobOperationData, ok := jobData["operation"].(map[string]interface{}); ok {
			if operationType, ok := jobOperationData["type"].(string); ok {
				jobOperations[jobUUID] = operationType
			}
		} else if jobTaskData, ok := jobData["task"].(map[string]interface{}); ok {
			// Fallback for old format
			if taskType, ok := jobTaskData["type"].(string); ok {
				jobOperations[jobUUID] = taskType
			}
		}
	}

	// Count operations based on unique jobs only
	for _, operationType := range jobOperations {
		operationCounts[operationType]++
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
		TotalJobs:       len(uniqueJobIDs),
		StatusCounts:    statusCounts,
		OperationCounts: operationCounts,
		DLQCount:        dlqCount,
	}, nil
}

// GetJobStatus returns information about a specific job.
func (c *Client) GetJobStatus(
	ctx context.Context,
	jobID string,
) (*job.QueuedJob, error) {
	// Try different status prefixes to find the job
	statusPrefixes := []string{"unprocessed.", "processing.", "completed.", "failed."}

	for _, prefix := range statusPrefixes {
		entry, err := c.kv.Get(prefix + jobID)
		if err != nil {
			continue // Try next prefix
		}

		var jobData map[string]interface{}
		if err := json.Unmarshal(entry.Value(), &jobData); err != nil {
			return nil, fmt.Errorf("failed to parse job data: %w", err)
		}

		queuedJob := &job.QueuedJob{
			ID:        jobID,
			Status:    getStringField(jobData, "status"),
			Created:   getStringField(jobData, "created"),
			Subject:   getStringField(jobData, "subject"),
			Error:     getStringField(jobData, "error"),
			UpdatedAt: getStringField(jobData, "updated_at"),
		}

		if operation, ok := jobData["operation"].(map[string]interface{}); ok {
			queuedJob.Operation = operation
		}

		if statusHistory, ok := jobData["status_history"].([]interface{}); ok {
			queuedJob.StatusHistory = statusHistory
		}

		if result, ok := jobData["result"]; ok {
			if resultBytes, err := json.Marshal(result); err == nil {
				queuedJob.Result = resultBytes
			}
		}

		return queuedJob, nil
	}

	return nil, fmt.Errorf("job not found: %s", jobID)
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
		// Extract status from key prefix
		keyParts := strings.SplitN(key, ".", 2)
		if len(keyParts) != 2 {
			continue
		}

		keyStatus := keyParts[0]
		jobID := keyParts[1]

		// Filter by status if specified
		if statusFilter != "" && keyStatus != statusFilter {
			continue
		}

		entry, err := c.kv.Get(key)
		if err != nil {
			continue // Skip jobs we can't read
		}

		var jobData map[string]interface{}
		if err := json.Unmarshal(entry.Value(), &jobData); err != nil {
			continue // Skip jobs we can't parse
		}

		queuedJob := &job.QueuedJob{
			ID:        jobID,
			Status:    getStringField(jobData, "status"),
			Created:   getStringField(jobData, "created"),
			Subject:   getStringField(jobData, "subject"),
			Error:     getStringField(jobData, "error"),
			UpdatedAt: getStringField(jobData, "updated_at"),
		}

		if operation, ok := jobData["operation"].(map[string]interface{}); ok {
			queuedJob.Operation = operation
		}

		if statusHistory, ok := jobData["status_history"].([]interface{}); ok {
			queuedJob.StatusHistory = statusHistory
		}

		if result, ok := jobData["result"]; ok {
			if resultBytes, err := json.Marshal(result); err == nil {
				queuedJob.Result = resultBytes
			}
		}

		jobs = append(jobs, queuedJob)
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
