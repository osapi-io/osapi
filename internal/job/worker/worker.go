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
	"log/slog"
	"os"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	natsclient "github.com/osapi-io/nats-client/pkg/client"
	"github.com/spf13/afero"

	"github.com/retr0h/osapi/internal/config"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/job/client"
	"github.com/retr0h/osapi/internal/messaging"
)

// New creates a new job worker instance.
func New(
	appFs afero.Fs,
	appConfig config.Config,
	logger *slog.Logger,
	natsClient messaging.NATSClient,
	jobClient client.JobClient,
) *Worker {
	return &Worker{
		logger:     logger,
		appConfig:  appConfig,
		appFs:      appFs,
		natsClient: natsClient,
		jobClient:  jobClient,
		done:       make(chan struct{}),
	}
}

// Start starts the worker and runs until the context is canceled.
func (w *Worker) Start(ctx context.Context) {
	w.logger.Info("starting job worker")
	w.run(ctx)
	w.logger.Info("job worker stopped")
}

// Stop gracefully shuts down the worker with timeout, following the API server pattern.
func (w *Worker) Stop(ctx context.Context) {
	w.logger.Info("stopping job worker")

	if w.cancel != nil {
		w.cancel()
	}

	select {
	case <-w.done:
		w.logger.Info("job worker stopped gracefully")
	case <-ctx.Done():
		w.logger.Error("job worker shutdown timeout")
	}

	// Wait for all goroutines to finish
	w.wg.Wait()
}

// run contains the main worker loop.
func (w *Worker) run(ctx context.Context) {
	w.logger.Info("job worker started successfully")

	// Determine worker hostname
	hostname := w.appConfig.Job.Worker.Hostname
	if hostname == "" {
		if h, err := os.Hostname(); err == nil {
			hostname = h
		} else {
			hostname = "unknown"
		}
	}

	w.logger.Info(
		"worker configuration",
		slog.String("hostname", hostname),
		slog.String("queue_group", w.appConfig.Job.Worker.QueueGroup),
		slog.Int("max_jobs", w.appConfig.Job.Worker.MaxJobs),
	)

	// Start consuming messages for different job types
	w.wg.Add(2)

	// Consumer for query jobs (read operations)
	go func() {
		defer w.wg.Done()
		if err := w.consumeQueryJobs(ctx, hostname); err != nil && err != context.Canceled {
			w.logger.Error(
				"query job consumer error",
				slog.String("error", err.Error()),
			)
		}
	}()

	// Consumer for modify jobs (write operations)
	go func() {
		defer w.wg.Done()
		if err := w.consumeModifyJobs(ctx, hostname); err != nil && err != context.Canceled {
			w.logger.Error(
				"modify job consumer error",
				slog.String("error", err.Error()),
			)
		}
	}()

	// Wait for cancellation
	<-ctx.Done()
	w.logger.Info("job worker shutting down")
}

// consumeQueryJobs handles read-only job operations using JetStream consumers.
func (w *Worker) consumeQueryJobs(
	ctx context.Context,
	hostname string,
) error {
	streamName := w.appConfig.Job.StreamName

	// Sanitize hostname for consumer names (alphanumeric and underscores only)
	sanitizedHostname := job.SanitizeHostname(hostname)

	// Create consumers for different query patterns
	consumers := []struct {
		name       string
		filter     string
		queueGroup string
	}{
		{
			name:       "query_any_" + sanitizedHostname,
			filter:     "jobs.query._any.>",
			queueGroup: w.appConfig.Job.Worker.QueueGroup,
		},
		{
			name:   "query_all_" + sanitizedHostname,
			filter: "jobs.query._all.>",
		},
		{
			name:   "query_direct_" + sanitizedHostname,
			filter: "jobs.query." + hostname + ".>",
		},
	}

	for _, consumer := range consumers {
		// Create the consumer first
		if err := w.createConsumer(ctx, streamName, consumer.name, consumer.filter); err != nil {
			w.logger.Error(
				"failed to create query consumer",
				slog.String("consumer", consumer.name),
				slog.String("error", err.Error()),
			)
			continue
		}

		w.wg.Add(1)
		go func(c struct {
			name       string
			filter     string
			queueGroup string
		},
		) {
			defer w.wg.Done()

			opts := &natsclient.ConsumeOptions{
				QueueGroup:  c.queueGroup,
				MaxInFlight: w.appConfig.Job.Worker.MaxJobs,
			}

			err := w.natsClient.ConsumeMessages(ctx, streamName, c.name, w.handleJobMessageJS, opts)
			if err != nil && err != context.Canceled {
				w.logger.Error(
					"error consuming query messages",
					slog.String("consumer", c.name),
					slog.String("error", err.Error()),
				)
			}
		}(consumer)
	}

	return nil
}

// handleJobMessageJS wraps the existing handleJobMessage for JetStream compatibility.
func (w *Worker) handleJobMessageJS(msg jetstream.Msg) error {
	// Convert JetStream message to NATS message for compatibility with existing handler
	natsMsg := &nats.Msg{
		Subject: msg.Subject(),
		Data:    msg.Data(),
	}

	// Call the handler and check if job processing actually succeeded
	err := w.handleJobMessage(natsMsg)
	if err != nil {
		// Don't acknowledge - let it retry
		return err
	}

	// Only acknowledge if job processing succeeded
	return nil
}

// createConsumer creates a durable JetStream consumer for the worker.
func (w *Worker) createConsumer(
	ctx context.Context,
	streamName, consumerName, filterSubject string,
) error {
	// Parse AckWait duration from config
	ackWait, _ := time.ParseDuration(w.appConfig.Job.Consumer.AckWait)

	// Parse BackOff durations from config
	var backOff []time.Duration
	for _, duration := range w.appConfig.Job.Consumer.BackOff {
		if d, err := time.ParseDuration(duration); err == nil {
			backOff = append(backOff, d)
		}
	}

	// Parse replay policy
	var replayPolicy jetstream.ReplayPolicy
	if w.appConfig.Job.Consumer.ReplayPolicy == "original" {
		replayPolicy = jetstream.ReplayOriginalPolicy
	} else {
		replayPolicy = jetstream.ReplayInstantPolicy
	}

	consumerConfig := jetstream.ConsumerConfig{
		Durable:       consumerName,
		FilterSubject: filterSubject,
		AckPolicy:     jetstream.AckExplicitPolicy,
		DeliverPolicy: jetstream.DeliverAllPolicy,
		MaxDeliver:    w.appConfig.Job.Consumer.MaxDeliver,
		AckWait:       ackWait,
		BackOff:       backOff,
		MaxAckPending: w.appConfig.Job.Consumer.MaxAckPending,
		ReplayPolicy:  replayPolicy,
	}

	return w.natsClient.CreateOrUpdateConsumerWithConfig(ctx, streamName, consumerConfig)
}

// consumeModifyJobs handles write job operations using JetStream consumers.
func (w *Worker) consumeModifyJobs(
	ctx context.Context,
	hostname string,
) error {
	streamName := w.appConfig.Job.StreamName

	// Sanitize hostname for consumer names (alphanumeric and underscores only)
	sanitizedHostname := job.SanitizeHostname(hostname)

	// Create consumers for different modify patterns
	consumers := []struct {
		name       string
		filter     string
		queueGroup string
	}{
		{
			name:       "modify_any_" + sanitizedHostname,
			filter:     "jobs.modify._any.>",
			queueGroup: w.appConfig.Job.Worker.QueueGroup,
		},
		{
			name:   "modify_all_" + sanitizedHostname,
			filter: "jobs.modify._all.>",
		},
		{
			name:   "modify_direct_" + sanitizedHostname,
			filter: "jobs.modify." + hostname + ".>",
		},
	}

	for _, consumer := range consumers {
		// Create the consumer first
		if err := w.createConsumer(ctx, streamName, consumer.name, consumer.filter); err != nil {
			w.logger.Error(
				"failed to create modify consumer",
				slog.String("consumer", consumer.name),
				slog.String("error", err.Error()),
			)
			continue
		}

		w.wg.Add(1)
		go func(c struct {
			name       string
			filter     string
			queueGroup string
		},
		) {
			defer w.wg.Done()

			opts := &natsclient.ConsumeOptions{
				QueueGroup:  c.queueGroup,
				MaxInFlight: w.appConfig.Job.Worker.MaxJobs,
			}

			err := w.natsClient.ConsumeMessages(ctx, streamName, c.name, w.handleJobMessageJS, opts)
			if err != nil && err != context.Canceled {
				w.logger.Error(
					"error consuming modify messages",
					slog.String("consumer", c.name),
					slog.String("error", err.Error()),
				)
			}
		}(consumer)
	}

	return nil
}
