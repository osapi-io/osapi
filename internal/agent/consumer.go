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

// Package agent provides the node agent implementation.
package agent

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	natsclient "github.com/osapi-io/nats-client/pkg/client"

	"github.com/retr0h/osapi/internal/job"
)

// consumeQueryJobs handles read-only job operations using JetStream consumers.
func (a *Agent) consumeQueryJobs(
	ctx context.Context,
	hostname string,
) error {
	streamName := a.streamName

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
			filter:     job.JobsQueryPrefix + "._any",
			queueGroup: a.appConfig.Node.Agent.QueueGroup,
		},
		{
			name:   "query_all_" + sanitizedHostname,
			filter: job.JobsQueryPrefix + "._all",
		},
		{
			name:   "query_direct_" + sanitizedHostname,
			filter: job.JobsQueryPrefix + ".host." + sanitizedHostname,
		},
	}

	// Add label-based consumers with hierarchical prefix subscriptions.
	// For "group: web.dev.us-east", creates consumers for each prefix level:
	//   jobs.query.label.group.web
	//   jobs.query.label.group.web.dev
	//   jobs.query.label.group.web.dev.us-east
	for key, value := range a.appConfig.Node.Agent.Labels {
		segments := strings.Split(value, ".")
		for i := range segments {
			prefix := strings.Join(segments[:i+1], ".")
			sanitizedPrefix := job.SanitizeHostname(prefix)
			consumers = append(consumers, struct {
				name       string
				filter     string
				queueGroup string
			}{
				name: fmt.Sprintf(
					"query_label_%s_%s_%s",
					key,
					sanitizedPrefix,
					sanitizedHostname,
				),
				filter: job.JobsQueryPrefix + ".label." + key + "." + prefix,
			})
		}
	}

	for _, consumer := range consumers {
		// Create the consumer first
		if err := a.createConsumer(ctx, streamName, consumer.name, consumer.filter); err != nil {
			a.logger.Error(
				"failed to create query consumer",
				slog.String("consumer", consumer.name),
				slog.String("error", err.Error()),
			)
			continue
		}

		a.wg.Add(1)
		go func(c struct {
			name       string
			filter     string
			queueGroup string
		},
		) {
			defer a.wg.Done()

			opts := &natsclient.ConsumeOptions{
				QueueGroup:  c.queueGroup,
				MaxInFlight: a.appConfig.Node.Agent.MaxJobs,
			}

			err := a.jobClient.ConsumeJobs(ctx, streamName, c.name, a.handleJobMessageJS, opts)
			if err != nil && err != context.Canceled {
				a.logger.Error(
					"error consuming query messages",
					slog.String("consumer", c.name),
					slog.String("error", err.Error()),
				)
			}
		}(consumer)
	}

	return nil
}

// consumeModifyJobs handles write job operations using JetStream consumers.
func (a *Agent) consumeModifyJobs(
	ctx context.Context,
	hostname string,
) error {
	streamName := a.streamName

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
			filter:     job.JobsModifyPrefix + "._any",
			queueGroup: a.appConfig.Node.Agent.QueueGroup,
		},
		{
			name:   "modify_all_" + sanitizedHostname,
			filter: job.JobsModifyPrefix + "._all",
		},
		{
			name:   "modify_direct_" + sanitizedHostname,
			filter: job.JobsModifyPrefix + ".host." + sanitizedHostname,
		},
	}

	// Add label-based consumers with hierarchical prefix subscriptions.
	for key, value := range a.appConfig.Node.Agent.Labels {
		segments := strings.Split(value, ".")
		for i := range segments {
			prefix := strings.Join(segments[:i+1], ".")
			sanitizedPrefix := job.SanitizeHostname(prefix)
			consumers = append(consumers, struct {
				name       string
				filter     string
				queueGroup string
			}{
				name: fmt.Sprintf(
					"modify_label_%s_%s_%s",
					key,
					sanitizedPrefix,
					sanitizedHostname,
				),
				filter: job.JobsModifyPrefix + ".label." + key + "." + prefix,
			})
		}
	}

	for _, consumer := range consumers {
		// Create the consumer first
		if err := a.createConsumer(ctx, streamName, consumer.name, consumer.filter); err != nil {
			a.logger.Error(
				"failed to create modify consumer",
				slog.String("consumer", consumer.name),
				slog.String("error", err.Error()),
			)
			continue
		}

		a.wg.Add(1)
		go func(c struct {
			name       string
			filter     string
			queueGroup string
		},
		) {
			defer a.wg.Done()

			opts := &natsclient.ConsumeOptions{
				QueueGroup:  c.queueGroup,
				MaxInFlight: a.appConfig.Node.Agent.MaxJobs,
			}

			err := a.jobClient.ConsumeJobs(ctx, streamName, c.name, a.handleJobMessageJS, opts)
			if err != nil && err != context.Canceled {
				a.logger.Error(
					"error consuming modify messages",
					slog.String("consumer", c.name),
					slog.String("error", err.Error()),
				)
			}
		}(consumer)
	}

	return nil
}

// handleJobMessageJS wraps the existing handleJobMessage for JetStream compatibility.
func (a *Agent) handleJobMessageJS(
	msg jetstream.Msg,
) error {
	err := a.handleJobMessage(msg)
	if err != nil {
		return err
	}

	return nil
}

// createConsumer creates a durable JetStream consumer for the agent.
func (a *Agent) createConsumer(
	ctx context.Context,
	streamName, consumerName, filterSubject string,
) error {
	// Parse AckWait duration from config
	ackWait, _ := time.ParseDuration(a.appConfig.Node.Agent.Consumer.AckWait)

	// Parse BackOff durations from config
	var backOff []time.Duration
	for _, duration := range a.appConfig.Node.Agent.Consumer.BackOff {
		if d, err := time.ParseDuration(duration); err == nil {
			backOff = append(backOff, d)
		}
	}

	// Parse replay policy
	var replayPolicy jetstream.ReplayPolicy
	if a.appConfig.Node.Agent.Consumer.ReplayPolicy == "original" {
		replayPolicy = jetstream.ReplayOriginalPolicy
	} else {
		replayPolicy = jetstream.ReplayInstantPolicy
	}

	consumerConfig := jetstream.ConsumerConfig{
		Durable:       consumerName,
		FilterSubject: filterSubject,
		AckPolicy:     jetstream.AckExplicitPolicy,
		DeliverPolicy: jetstream.DeliverAllPolicy,
		MaxDeliver:    a.appConfig.Node.Agent.Consumer.MaxDeliver,
		AckWait:       ackWait,
		BackOff:       backOff,
		MaxAckPending: a.appConfig.Node.Agent.Consumer.MaxAckPending,
		ReplayPolicy:  replayPolicy,
	}

	return a.jobClient.CreateOrUpdateConsumer(ctx, streamName, consumerConfig)
}
