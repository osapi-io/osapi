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

package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/retr0h/osapi/internal/job"
)

// writeStatusEvent writes an append-only status event for a job.
// This eliminates race conditions by never updating existing keys.
func (w *Worker) writeStatusEvent(
	jobID string,
	event string,
	data map[string]interface{},
) error {
	// Get hostname for this worker (GetWorkerHostname always succeeds)
	hostname, _ := job.GetWorkerHostname(w.appConfig.Job.Worker.Hostname)

	// Use job client to write status event
	return w.jobClient.WriteStatusEvent(context.Background(), jobID, event, hostname, data)
}

// handleJobMessage processes incoming job messages from NATS.
func (w *Worker) handleJobMessage(
	msg *nats.Msg,
) error {
	// Extract the key (job ID) from the message data
	jobKey := string(msg.Data)

	w.logger.Info(
		"received job notification",
		slog.String("subject", msg.Subject),
		slog.String("job_key", jobKey),
	)

	w.logger.Debug(
		"processing job message",
		slog.String("subject", msg.Subject),
		slog.String("job_key", jobKey),
		slog.String("raw_data", string(msg.Data)),
	)

	// Parse subject to extract prefix and hostname
	prefix, _, err := job.ParseSubject(msg.Subject)
	if err != nil {
		return fmt.Errorf("failed to parse subject %s: %w", msg.Subject, err)
	}

	// Get the immutable job data
	jobDataKey := "jobs." + jobKey
	jobDataBytes, err := w.jobClient.GetJobData(context.Background(), jobDataKey)
	if err != nil {
		return fmt.Errorf("job not found: %s", jobKey)
	}

	// Parse the job data
	var jobData map[string]interface{}
	if err := json.Unmarshal(jobDataBytes, &jobData); err != nil {
		return fmt.Errorf("failed to parse job data: %w", err)
	}

	// Extract the RequestID from top-level job data
	requestID, ok := jobData["id"].(string)
	if !ok {
		return fmt.Errorf("invalid job format: missing id")
	}

	// Extract the operation data
	operationData, ok := jobData["operation"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid job format: missing operation")
	}

	// Extract operation type from the operation data
	operationType, ok := operationData["type"].(string)
	if !ok {
		return fmt.Errorf("invalid operation format: missing type field")
	}

	// Parse operation type to extract category and operation
	parts := strings.Split(operationType, ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid operation type format: %s", operationType)
	}

	category := parts[0]
	operation := strings.Join(parts[1:], ".")

	// Convert operation data to job.Request
	// operationData came from json.Unmarshal, so Marshal always succeeds.
	operationJSON, _ := json.Marshal(operationData)

	var jobRequest job.Request
	// operationJSON is valid JSON, Unmarshal into a struct always succeeds.
	_ = json.Unmarshal(operationJSON, &jobRequest)

	// Set the RequestID and extract category/operation from operation data
	jobRequest.RequestID = requestID
	jobRequest.Category = category
	jobRequest.Operation = operation

	// Determine Type from subject prefix (jobs.query vs jobs.modify)
	if strings.HasPrefix(prefix, "jobs.query") {
		jobRequest.Type = job.TypeQuery
	} else if strings.HasPrefix(prefix, "jobs.modify") {
		jobRequest.Type = job.TypeModify
	}

	// Write acknowledged event
	if err := w.writeStatusEvent(jobKey, "acknowledged", map[string]interface{}{
		"subject":   msg.Subject,
		"category":  jobRequest.Category,
		"operation": jobRequest.Operation,
	}); err != nil {
		w.logger.Error("failed to write acknowledged event", slog.String("error", err.Error()))
	}

	// Process the job
	w.logger.Info(
		"processing job",
		slog.String("job_id", jobKey),
		slog.String("type", string(jobRequest.Type)),
		slog.String("category", jobRequest.Category),
		slog.String("operation", jobRequest.Operation),
	)

	// Write started event
	startTime := time.Now()
	if err := w.writeStatusEvent(jobKey, "started", map[string]interface{}{
		"worker_version": "1.0.0", // TODO: get from config or build info
		"pid":            os.Getpid(),
	}); err != nil {
		w.logger.Error("failed to write started event", slog.String("error", err.Error()))
	}

	// Get worker hostname (GetWorkerHostname always succeeds)
	hostname, _ := job.GetWorkerHostname(w.appConfig.Job.Worker.Hostname)

	// Create job response
	response := job.Response{
		RequestID: jobRequest.RequestID,
		Status:    job.StatusProcessing,
		Hostname:  hostname,
		Timestamp: time.Now(),
	}

	// Process based on category and operation
	result, err := w.processJobOperation(jobRequest)
	if err != nil {
		w.logger.Error(
			"job processing failed",
			slog.String("job_id", jobKey),
			slog.String("category", jobRequest.Category),
			slog.String("operation", jobRequest.Operation),
			slog.String("error", err.Error()),
		)
		response.Status = job.StatusFailed
		response.Error = err.Error()

		// Write failed event
		if err := w.writeStatusEvent(jobKey, "failed", map[string]interface{}{
			"error":       err.Error(),
			"duration_ms": time.Since(startTime).Milliseconds(),
		}); err != nil {
			w.logger.Error("failed to write failed event", slog.String("error", err.Error()))
		}
	} else {
		response.Status = job.StatusCompleted
		response.Data = result

		// Write completed event
		if err := w.writeStatusEvent(jobKey, "completed", map[string]interface{}{
			"duration_ms": time.Since(startTime).Milliseconds(),
			"result_size": len(result),
		}); err != nil {
			w.logger.Error("failed to write completed event", slog.String("error", err.Error()))
		}
	}

	response.Timestamp = time.Now()

	// Store response using job client
	var errorMsg string
	if response.Error != "" {
		errorMsg = response.Error
	}

	err = w.jobClient.WriteJobResponse(context.Background(), jobKey, hostname,
		response.Data, string(response.Status), errorMsg)
	if err != nil {
		return fmt.Errorf("failed to store job response: %w", err)
	}

	// NOTE: We no longer update the original job - it remains immutable.
	// Status is now tracked through append-only events to avoid race conditions.

	w.logger.Info(
		"job processing completed",
		slog.String("job_id", jobKey),
		slog.String("status", string(response.Status)),
	)

	// Return error if job failed so message won't be acknowledged and will retry
	if response.Status == job.StatusFailed {
		return fmt.Errorf("job processing failed: %s", response.Error)
	}

	return nil
}
