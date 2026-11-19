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
	"sync"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	natsclient "github.com/osapi-io/nats-client/pkg/client"

	"github.com/osapi-io/osapi/internal/job"
)

// consumeQueryJobs handles read-only job operations using JetStream consumers.
func (a *Agent) consumeQueryJobs(
	ctx context.Context,
	machineID string,
) error {
	streamName := a.streamName

	// Sanitize machine ID for consumer names (alphanumeric and underscores only).
	sanitizedMachineID := job.SanitizeHostname(machineID)

	// Use configured consumer name as prefix, falling back to empty string
	consumerPrefix := a.consumerNamePrefix()

	// Create consumers for different query patterns.
	// Direct targeting uses machine ID (permanent), not hostname (mutable).
	consumers := []struct {
		name       string
		filter     string
		queueGroup string
	}{
		{
			name:       consumerPrefix + "query_any_" + sanitizedMachineID,
			filter:     job.JobsQueryPrefix + "._any",
			queueGroup: a.appConfig.Agent.QueueGroup,
		},
		{
			name:   consumerPrefix + "query_all_" + sanitizedMachineID,
			filter: job.JobsQueryPrefix + "._all",
		},
		{
			name:   consumerPrefix + "query_direct_" + sanitizedMachineID,
			filter: job.JobsQueryPrefix + ".host." + sanitizedMachineID,
		},
	}

	// Add label-based consumers with hierarchical prefix subscriptions.
	// For "group: web.dev.us-east", creates consumers for each prefix level:
	//   jobs.query.label.group.web
	//   jobs.query.label.group.web.dev
	//   jobs.query.label.group.web.dev.us-east
	for key, value := range a.appConfig.Agent.Labels {
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
					"%squery_label_%s_%s_%s",
					consumerPrefix,
					key,
					sanitizedPrefix,
					sanitizedMachineID,
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

		a.consumerWg.Add(1)
		set := a.consumerSetWg
		if set != nil {
			set.Add(1)
		}
		go func(c struct {
			name       string
			filter     string
			queueGroup string
		},
		) {
			defer a.consumerWg.Done()
			if set != nil {
				defer set.Done()
			}

			opts := &natsclient.ConsumeOptions{
				QueueGroup:  c.queueGroup,
				MaxInFlight: a.appConfig.Agent.MaxJobs,
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
	machineID string,
) error {
	streamName := a.streamName

	// Sanitize machine ID for consumer names (alphanumeric and underscores only).
	sanitizedMachineID := job.SanitizeHostname(machineID)

	// Use configured consumer name as prefix, falling back to empty string
	consumerPrefix := a.consumerNamePrefix()

	// Create consumers for different modify patterns.
	// Direct targeting uses machine ID (permanent), not hostname (mutable).
	consumers := []struct {
		name       string
		filter     string
		queueGroup string
	}{
		{
			name:       consumerPrefix + "modify_any_" + sanitizedMachineID,
			filter:     job.JobsModifyPrefix + "._any",
			queueGroup: a.appConfig.Agent.QueueGroup,
		},
		{
			name:   consumerPrefix + "modify_all_" + sanitizedMachineID,
			filter: job.JobsModifyPrefix + "._all",
		},
		{
			name:   consumerPrefix + "modify_direct_" + sanitizedMachineID,
			filter: job.JobsModifyPrefix + ".host." + sanitizedMachineID,
		},
	}

	// Add label-based consumers with hierarchical prefix subscriptions.
	for key, value := range a.appConfig.Agent.Labels {
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
					"%smodify_label_%s_%s_%s",
					consumerPrefix,
					key,
					sanitizedPrefix,
					sanitizedMachineID,
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

		a.consumerWg.Add(1)
		set := a.consumerSetWg
		if set != nil {
			set.Add(1)
		}
		go func(c struct {
			name       string
			filter     string
			queueGroup string
		},
		) {
			defer a.consumerWg.Done()
			if set != nil {
				defer set.Done()
			}

			opts := &natsclient.ConsumeOptions{
				QueueGroup:  c.queueGroup,
				MaxInFlight: a.appConfig.Agent.MaxJobs,
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

// State returns the agent's scheduling state. Read by the heartbeat on its own
// goroutine while a drain or an enrollment acceptance writes it.
func (a *Agent) State() string {
	a.lifecycleMu.RLock()
	defer a.lifecycleMu.RUnlock()

	return a.state
}

// setState records a state that carries no consumer change with it. A
// transition that does own one takes lifecycleMu itself and sets the field
// under it, so the two cannot be observed apart.
func (a *Agent) setState(
	state string,
) {
	a.lifecycleMu.Lock()
	defer a.lifecycleMu.Unlock()

	a.state = state
}

// startConsumers starts the job consumers at startup, when the agent is
// Ready. The heartbeat starts before it and can find the drain flag first:
// then the agent is already Cordoned, and it must not take jobs. Starting
// the consumers anyway left them running while cordoned, and the undrain
// started a second set over them, whose cancel replaced the first's; the
// next drain then waited for the first set forever (#876).
func (a *Agent) startConsumers() {
	a.lifecycleMu.Lock()
	defer a.lifecycleMu.Unlock()

	if a.state != job.AgentStateReady {
		a.logger.Info(
			"job consumers not started",
			slog.String("state", a.state),
		)

		return
	}
	a.startConsumersLocked()
}

// startConsumersLocked is startConsumers for a caller that already holds
// lifecycleMu, so a state change and the consumers it implies are one step.
// When the consumers already run, it does nothing: a second set would
// replace the cancel of the first, and cancelConsumersLocked could not stop
// it.
func (a *Agent) startConsumersLocked() {
	if a.consumerCancel != nil {
		return
	}
	a.consumerCtx, a.consumerCancel = context.WithCancel(a.ctx)
	a.consumerSetWg = &sync.WaitGroup{}
	_ = a.consumeQueryJobs(a.consumerCtx, a.machineID)
	_ = a.consumeModifyJobs(a.consumerCtx, a.machineID)
}

// cancelConsumersLocked cancels the running set of consumers and returns
// its wait group, or nil when no set runs. It does not wait: a drain waits
// for the set in its own goroutine, so a consumer that is slow to stop
// never holds lifecycleMu or the heartbeat. The caller holds lifecycleMu.
func (a *Agent) cancelConsumersLocked() *sync.WaitGroup {
	if a.consumerCancel == nil {
		return nil
	}
	a.consumerCancel()
	a.consumerCancel = nil
	set := a.consumerSetWg
	a.consumerSetWg = nil

	return set
}

// handleJobMessageJS is the JetStream handler of a job message.
func (a *Agent) handleJobMessageJS(
	ctx context.Context,
	msg jetstream.Msg,
) error {
	return a.handleJobMessage(ctx, msg)
}

// consumerNamePrefix returns the configured consumer name followed by an
// underscore separator. When no consumer name is configured, an empty string
// is returned so existing consumer names are unchanged.
func (a *Agent) consumerNamePrefix() string {
	if name := a.appConfig.Agent.Consumer.Name; name != "" {
		return name + "_"
	}

	return ""
}

// createConsumer creates a durable JetStream consumer for the agent.
func (a *Agent) createConsumer(
	ctx context.Context,
	streamName, consumerName, filterSubject string,
) error {
	// Parse AckWait duration from config
	ackWait, _ := time.ParseDuration(a.appConfig.Agent.Consumer.AckWait)

	// Parse BackOff durations from config
	var backOff []time.Duration
	for _, duration := range a.appConfig.Agent.Consumer.BackOff {
		if d, err := time.ParseDuration(duration); err == nil {
			backOff = append(backOff, d)
		}
	}

	// Parse replay policy
	var replayPolicy jetstream.ReplayPolicy
	if a.appConfig.Agent.Consumer.ReplayPolicy == "original" {
		replayPolicy = jetstream.ReplayOriginalPolicy
	} else {
		replayPolicy = jetstream.ReplayInstantPolicy
	}

	consumerConfig := jetstream.ConsumerConfig{
		Durable:       consumerName,
		FilterSubject: filterSubject,
		AckPolicy:     jetstream.AckExplicitPolicy,
		DeliverPolicy: jetstream.DeliverAllPolicy,
		MaxDeliver:    a.appConfig.Agent.Consumer.MaxDeliver,
		AckWait:       ackWait,
		BackOff:       backOff,
		MaxAckPending: a.appConfig.Agent.Consumer.MaxAckPending,
		ReplayPolicy:  replayPolicy,
	}

	return a.jobClient.CreateOrUpdateConsumer(ctx, streamName, consumerConfig)
}
