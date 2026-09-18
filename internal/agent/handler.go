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

package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/provider"
	"github.com/osapi-io/osapi/internal/telemetry/tracing"
)

// Errors returned by unwrapJobEnvelope. Each names a distinct rejection
// reason so an operator reading agent logs or job termination reasons can
// tell an unenrolled agent apart from a forged or malformed job.
var (
	// ErrJobEnvelopeMissing means PKI is enabled but the job data is not a
	// signed envelope at all, or is missing a required envelope field.
	ErrJobEnvelopeMissing = errors.New("job data is not a signed envelope")
	// ErrControllerKeyUnknown means PKI is enabled but this agent has not
	// completed enrollment, so it holds no controller public key to verify
	// against.
	ErrControllerKeyUnknown = errors.New("no cached controller public key: agent not enrolled")
	// ErrJobSignatureInvalid means the envelope signature did not verify
	// against the cached controller key (or the previous key, during a
	// rotation grace period).
	ErrJobSignatureInvalid = errors.New("invalid controller signature on job data")
)

// extractChanged parses the processor result JSON and extracts the "changed"
// boolean field. Returns nil if the field is absent.
func extractChanged(
	data json.RawMessage,
) *bool {
	if len(data) == 0 {
		return nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil
	}
	v, ok := m["changed"]
	if !ok {
		return nil
	}
	b, ok := v.(bool)
	if !ok {
		return nil
	}
	return &b
}

// unwrapJobEnvelope unwraps a SignedEnvelope from job data and verifies its
// signature. When PKI is disabled (pkiManager is nil), the raw data passes
// through unchanged — job payloads are never wrapped in an envelope in that
// case. When PKI is enabled, verification is fail-closed: a missing or
// malformed envelope, an agent with no cached controller key, and a bad
// signature are each rejected with a distinct error rather than executed.
// Returns the inner payload or one of the Err* sentinels above.
func (a *Agent) unwrapJobEnvelope(
	data []byte,
) ([]byte, error) {
	if a.pkiManager == nil {
		return data, nil
	}

	var envelope job.SignedEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil ||
		len(envelope.Payload) == 0 || len(envelope.Signature) == 0 || envelope.Fingerprint == "" {
		return nil, ErrJobEnvelopeMissing
	}

	// An agent that has not finished enrollment holds no controller key
	// and so cannot verify anything. That is distinct from a bad
	// signature: the former means "not yet trusted", the latter means
	// "this job was tampered with or forged".
	if len(a.pkiManager.ControllerPublicKey()) == 0 {
		return nil, ErrControllerKeyUnknown
	}

	// VerifyWithGrace checks both the current and previous controller
	// key, so a key rotation does not reject jobs signed just before it.
	if !a.pkiManager.VerifyWithGrace(envelope.Payload, envelope.Signature) {
		return nil, ErrJobSignatureInvalid
	}

	a.logger.Debug(
		"verified job signature",
		slog.String("fingerprint", envelope.Fingerprint),
	)

	return envelope.Payload, nil
}

// writeStatusEvent writes an append-only status event for a job.
// This eliminates race conditions by never updating existing keys.
func (a *Agent) writeStatusEvent(
	ctx context.Context,
	jobID string,
	event string,
	data map[string]interface{},
) error {
	// Use cached hostname resolved at startup.
	return a.jobClient.WriteStatusEvent(ctx, jobID, event, a.hostname, data)
}

// defaultAckWait is used to size the InProgress keepalive interval when the
// configured AckWait is missing or unparsable. It matches the default set in
// cmd/root.go.
const defaultAckWait = 2 * time.Minute

// inProgressInterval overrides the computed InProgress keepalive interval.
// Zero means derive it from the consumer's AckWait, halved. Exposed for
// tests via export_test.go.
var inProgressInterval time.Duration

// startInProgressKeepAlive periodically calls msg.InProgress() to reset the
// server's redelivery timer while a long-running operation executes. It
// returns a stop function that must be called exactly once when the
// operation finishes; stop blocks until the keepalive goroutine has exited,
// so no goroutine outlives the call.
func (a *Agent) startInProgressKeepAlive(
	ctx context.Context,
	msg jetstream.Msg,
) func() {
	interval := inProgressInterval
	if interval <= 0 {
		ackWait, err := time.ParseDuration(a.appConfig.Agent.Consumer.AckWait)
		if err != nil || ackWait <= 0 {
			ackWait = defaultAckWait
		}
		interval = ackWait / 2
	}

	done := make(chan struct{})
	stop := make(chan struct{})

	go func() {
		defer close(done)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-stop:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := msg.InProgress(); err != nil {
					a.logger.WarnContext(
						ctx,
						"failed to extend job ack deadline",
						slog.String("error", err.Error()),
					)
				}
			}
		}
	}()

	return func() {
		close(stop)
		<-done
	}
}

// terminateMessage tells JetStream not to redeliver msg, then returns err
// unchanged. Used for pre-execution failures that redelivery can never fix:
// a malformed payload, an unparsable subject, or a failed signature.
func (a *Agent) terminateMessage(
	msg jetstream.Msg,
	reason string,
	err error,
) error {
	if termErr := msg.TermWithReason(reason); termErr != nil {
		a.logger.Warn(
			"failed to terminate undeliverable message",
			slog.String("reason", reason),
			slog.String("error", termErr.Error()),
		)
	}
	return err
}

// handleJobMessage processes incoming job messages from NATS.
func (a *Agent) handleJobMessage(
	msg jetstream.Msg,
) error {
	// Extract the key (job ID) from the message data
	jobKey := string(msg.Data())

	a.logger.Info(
		"received job notification",
		slog.String("subject", msg.Subject()),
		slog.String("job_key", jobKey),
	)

	a.logger.Debug(
		"processing job message",
		slog.String("subject", msg.Subject()),
		slog.String("job_key", jobKey),
		slog.String("raw_data", string(msg.Data())),
	)

	// A redelivered message for a job this agent already answered must not
	// re-execute the operation — the response and terminal status are
	// already recorded. Checked before touching the job data itself, since
	// a message that gets this far has already run to completion once.
	// The check is scoped to this agent's hostname, so broadcast jobs
	// (where every targeted agent answers independently) are unaffected.
	// A failed check fails open: proceed with execution rather than risk
	// silently dropping a job this agent has not actually answered.
	if answered, err := a.jobClient.HasJobResponse(context.Background(), jobKey, a.hostname); err != nil {
		a.logger.Warn(
			"failed to check for existing job response; proceeding with execution",
			slog.String("job_id", jobKey),
			slog.String("error", err.Error()),
		)
	} else if answered {
		a.logger.Debug(
			"job already answered by this agent; skipping redelivered execution",
			slog.String("job_id", jobKey),
			slog.String("hostname", a.hostname),
		)
		return nil
	}

	// Parse subject to extract prefix and hostname
	prefix, _, err := job.ParseSubject(msg.Subject())
	if err != nil {
		return a.terminateMessage(msg, "unparsable subject",
			fmt.Errorf("failed to parse subject %s: %w", msg.Subject(), err))
	}

	// Get the immutable job data
	jobDataKey := "jobs." + jobKey
	jobDataBytes, err := a.jobClient.GetJobData(context.Background(), jobDataKey)
	if err != nil {
		// Redelivery can help here: the job's KV write and this
		// notification are not transactional, so a very early delivery
		// may briefly race the write.
		return fmt.Errorf("job not found: %s", jobKey)
	}

	// Unwrap and verify the signed envelope when PKI is enabled. Each
	// rejection reason is distinct so an operator can tell an unenrolled
	// agent apart from a forged or malformed job in logs and in the
	// terminated message's reason.
	jobDataBytes, err = a.unwrapJobEnvelope(jobDataBytes)
	if err != nil {
		reason := "job signature verification failed"
		switch {
		case errors.Is(err, ErrControllerKeyUnknown):
			reason = "agent not enrolled: no cached controller key"
		case errors.Is(err, ErrJobEnvelopeMissing):
			reason = "missing or malformed job signature"
		case errors.Is(err, ErrJobSignatureInvalid):
			reason = "invalid job signature"
		}
		return a.terminateMessage(msg, reason,
			fmt.Errorf("job signature verification failed: %w", err))
	}

	// Parse the job data
	var jobData map[string]interface{}
	if err := json.Unmarshal(jobDataBytes, &jobData); err != nil {
		return a.terminateMessage(msg, "malformed job data",
			fmt.Errorf("failed to parse job data: %w", err))
	}

	// Extract trace context from NATS message headers and create a processing span
	ctx := tracing.ExtractTraceContextFromHeader(context.Background(), http.Header(msg.Headers()))
	ctx, span := otel.Tracer("osapi-agent").Start(ctx, "job.process")
	defer span.End()

	// Extract the job ID from top-level job data
	jobID, ok := jobData["id"].(string)
	if !ok {
		return a.terminateMessage(msg, "missing job id",
			fmt.Errorf("invalid job format: missing id"))
	}

	// Extract the operation data
	operationData, ok := jobData["operation"].(map[string]interface{})
	if !ok {
		return a.terminateMessage(msg, "missing operation",
			fmt.Errorf("invalid job format: missing operation"))
	}

	// Extract operation type from the operation data
	operationType, ok := operationData["type"].(string)
	if !ok {
		return a.terminateMessage(msg, "missing operation type",
			fmt.Errorf("invalid operation format: missing type field"))
	}

	// Parse operation type to extract category and operation
	parts := strings.Split(operationType, ".")
	if len(parts) < 2 {
		return a.terminateMessage(msg, "invalid operation type",
			fmt.Errorf("invalid operation type format: %s", operationType))
	}

	category := parts[0]
	operation := strings.Join(parts[1:], ".")

	span.SetAttributes(
		attribute.String("job.id", jobID),
		attribute.String("job.category", category),
		attribute.String("job.operation", operation),
		attribute.String("job.type", operationType),
	)

	// Convert operation data to job.Request
	// operationData came from json.Unmarshal, so Marshal always succeeds.
	operationJSON, _ := json.Marshal(operationData)

	var jobRequest job.Request
	// operationJSON is valid JSON, Unmarshal into a struct always succeeds.
	_ = json.Unmarshal(operationJSON, &jobRequest)

	// Set the job ID and extract category/operation from operation data
	jobRequest.JobID = jobID
	jobRequest.Category = category
	jobRequest.Operation = operation

	// Determine Type from subject prefix
	if strings.HasPrefix(prefix, job.JobsQueryPrefix) {
		jobRequest.Type = job.TypeQuery
	} else if strings.HasPrefix(prefix, job.JobsModifyPrefix) {
		jobRequest.Type = job.TypeModify
	}

	// Write acknowledged event
	if err := a.writeStatusEvent(ctx, jobKey, string(job.StatusAcknowledged), map[string]interface{}{
		"subject":   msg.Subject(),
		"category":  jobRequest.Category,
		"operation": jobRequest.Operation,
	}); err != nil {
		a.logger.ErrorContext(
			ctx,
			"failed to write acknowledged event",
			slog.String("error", err.Error()),
		)
	}

	// Resolve @fact.X references in job request data.
	// Always attempt resolution so @fact. strings never pass through as literals.
	// ResolveFacts handles nil cachedFacts with a clear "facts not available" error.
	var resolveFactsErr error
	if len(jobRequest.Data) > 0 {
		var dataMap map[string]any
		if err := json.Unmarshal(jobRequest.Data, &dataMap); err == nil {
			resolved, err := ResolveFacts(dataMap, a.cachedFacts, a.hostname)
			if err != nil {
				resolveFactsErr = fmt.Errorf("failed to resolve fact references: %w", err)
			} else if resolved != nil {
				if resolvedJSON, err := json.Marshal(resolved); err == nil {
					jobRequest.Data = resolvedJSON
				}
			}
		}
	}

	// Process the job
	a.logger.InfoContext(
		ctx,
		"processing job",
		slog.String("job_id", jobKey),
		slog.String("type", string(jobRequest.Type)),
		slog.String("category", jobRequest.Category),
		slog.String("operation", jobRequest.Operation),
	)

	// Write started event
	startTime := time.Now()
	if err := a.writeStatusEvent(ctx, jobKey, string(job.StatusStarted), map[string]interface{}{
		"agent_version": "1.0.0", // TODO: get from config or build info
		"pid":           os.Getpid(),
	}); err != nil {
		a.logger.ErrorContext(
			ctx,
			"failed to write started event",
			slog.String("error", err.Error()),
		)
	}

	if a.jobsActive != nil {
		a.jobsActive.Add(ctx, 1)
	}

	// Create job response using cached hostname resolved at startup.
	hostname := a.hostname
	response := job.Response{
		JobID:     jobRequest.JobID,
		Status:    job.StatusProcessing,
		Hostname:  hostname,
		Timestamp: time.Now(),
	}

	// Process based on category and operation.
	// Fact resolution errors flow through the same path as processing errors
	// so the error is written to KV and clients get it instead of timing out.
	var result json.RawMessage
	if resolveFactsErr != nil {
		err = resolveFactsErr
	} else {
		// Extend the ack deadline periodically while the operation runs, so
		// an operation that outlives AckWait is not redelivered mid-flight.
		stopKeepAlive := a.startInProgressKeepAlive(ctx, msg)
		result, err = a.processJobOperation(jobRequest)
		stopKeepAlive()
	}
	if err != nil && errors.Is(err, provider.ErrUnsupported) {
		a.logger.WarnContext(
			ctx,
			"job skipped: unsupported on this OS family",
			slog.String("job_id", jobKey),
			slog.String("category", jobRequest.Category),
			slog.String("operation", jobRequest.Operation),
			slog.String("error", err.Error()),
		)
		response.Status = job.StatusSkipped
		response.Error = err.Error()
	} else if err != nil {
		a.logger.ErrorContext(
			ctx,
			"job processing failed",
			slog.String("job_id", jobKey),
			slog.String("category", jobRequest.Category),
			slog.String("operation", jobRequest.Operation),
			slog.String("error", err.Error()),
		)
		response.Status = job.StatusFailed
		response.Error = err.Error()
	} else {
		response.Status = job.StatusCompleted
		response.Data = result

		// Extract "changed" from result data for mutation operations
		if jobRequest.Type == job.TypeModify {
			response.Changed = extractChanged(result)
		}
	}

	if a.jobsActive != nil {
		a.jobsActive.Add(ctx, -1)
	}
	if a.jobDuration != nil {
		a.jobDuration.Record(ctx, time.Since(startTime).Seconds())
	}
	if a.jobsProcessed != nil {
		status := "completed"
		if response.Status == job.StatusFailed {
			status = "failed"
		}
		a.jobsProcessed.Add(ctx, 1,
			metric.WithAttributes(attribute.String("status", status)))
	}

	response.Timestamp = time.Now()

	// Store response before writing the status event. Polling clients
	// use the status event to decide completion — the response data
	// must already be in KV when they read it.
	var errorMsg string
	if response.Error != "" {
		errorMsg = response.Error
	}

	err = a.jobClient.WriteJobResponse(ctx, jobKey, hostname,
		response.Data, string(response.Status), errorMsg, response.Changed)
	if err != nil {
		// The operation already ran. Returning an error here would leave
		// the message unacked, and JetStream would redeliver it — but
		// HasJobResponse would find nothing, since this write is what
		// failed, and the operation would run a second time. That is the
		// exact double-execution this handler exists to prevent, so the
		// failure is recorded as best-effort and the message is still
		// acked; the caller sees a failed or timed-out job rather than a
		// second execution.
		a.logger.ErrorContext(
			ctx,
			"failed to store job response",
			slog.String("job_id", jobKey),
			slog.String("error", err.Error()),
		)
		if statusErr := a.writeStatusEvent(ctx, jobKey, string(job.StatusFailed), map[string]interface{}{
			"error":       fmt.Sprintf("failed to store job response: %s", err.Error()),
			"duration_ms": time.Since(startTime).Milliseconds(),
		}); statusErr != nil {
			a.logger.ErrorContext(
				ctx,
				"failed to write failed event after response storage failure",
				slog.String("error", statusErr.Error()),
			)
		}
		return nil
	}

	// Write terminal status event after the response is persisted.
	switch response.Status {
	case job.StatusFailed:
		if err := a.writeStatusEvent(ctx, jobKey, string(job.StatusFailed), map[string]interface{}{
			"error":       response.Error,
			"duration_ms": time.Since(startTime).Milliseconds(),
		}); err != nil {
			a.logger.ErrorContext(
				ctx,
				"failed to write failed event",
				slog.String("error", err.Error()),
			)
		}
	case job.StatusSkipped:
		if err := a.writeStatusEvent(ctx, jobKey, string(job.StatusSkipped), map[string]interface{}{
			"error":       response.Error,
			"duration_ms": time.Since(startTime).Milliseconds(),
		}); err != nil {
			a.logger.ErrorContext(
				ctx,
				"failed to write skipped event",
				slog.String("error", err.Error()),
			)
		}
	default:
		if err := a.writeStatusEvent(ctx, jobKey, string(job.StatusCompleted), map[string]interface{}{
			"duration_ms": time.Since(startTime).Milliseconds(),
			"result_size": len(result),
		}); err != nil {
			a.logger.ErrorContext(
				ctx,
				"failed to write completed event",
				slog.String("error", err.Error()),
			)
		}
	}

	// NOTE: We no longer update the original job - it remains immutable.
	// Status is now tracked through append-only events to avoid race conditions.

	a.logger.InfoContext(
		ctx,
		"job processing completed",
		slog.String("job_id", jobKey),
		slog.String("status", string(response.Status)),
	)

	// The operation has run and its outcome — success or failure — is
	// durably recorded above. The job is terminal from this agent's
	// perspective: acknowledge the message so JetStream does not redeliver
	// and re-execute it. A client that wants another attempt uses
	// `job retry`, which creates a new job rather than relying on
	// redelivery of this one.
	return nil
}
