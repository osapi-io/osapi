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
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/osapi-io/osapi/pkg/sdk/types"

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
	consumers := []consumerSpec{
		{
			name:   consumerPrefix + "query_any_" + sanitizedMachineID,
			filter: job.JobsQueryPrefix + "._any",
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
			consumers = append(consumers, consumerSpec{
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
		a.startConsumer(ctx, streamName, consumer, types.JobTypeQuery)
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
	consumers := []consumerSpec{
		{
			name:   consumerPrefix + "modify_any_" + sanitizedMachineID,
			filter: job.JobsModifyPrefix + "._any",
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
			consumers = append(consumers, consumerSpec{
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
		a.startConsumer(ctx, streamName, consumer, types.JobTypeModify)
	}

	return nil
}

// consumerSpec is one durable job consumer of the agent: its name and the
// subject it filters on.
type consumerSpec struct {
	name   string
	filter string
}

// consumerRetryInterval is how long the agent waits before it tries again
// to create or bind a job consumer that failed.
var consumerRetryInterval = 10 * time.Second

// startConsumer creates the consumer c and consumes its jobs in a goroutine
// of the running set. A consumer that cannot be created or bound, such as
// when the stream has reached its limit of consumers, is recorded for the
// ConsumersUnavailable condition, and the goroutine tries again each
// consumerRetryInterval until it runs or ctx ends. The first try is not in
// the goroutine, so a caller sees each consumer created that can be.
func (a *Agent) startConsumer(
	ctx context.Context,
	streamName string,
	c consumerSpec,
	kind types.JobType,
) {
	err := a.createConsumer(ctx, streamName, c.name, c.filter)
	if err != nil {
		a.consumerFailed(c.name, "failed to create "+string(kind)+" consumer", err)
	}

	opts := &natsclient.ConsumeOptions{
		MaxInFlight: a.appConfig.Agent.MaxJobs,
	}
	retry := consumerRetryInterval

	a.consumerWg.Add(1)
	set := a.consumerSetWg
	if set != nil {
		set.Add(1)
	}
	go func() {
		defer a.consumerWg.Done()
		if set != nil {
			defer set.Done()
		}
		// A consumer that stops is not a failed one: a drain or a stop
		// clears it from the condition.
		defer a.consumerRecovered(c.name)

		for {
			if err == nil {
				err = a.jobClient.ConsumeJobs(ctx, streamName, c.name, a.handleJobMessageJS, opts)
				if err == nil || errors.Is(err, context.Canceled) {
					return
				}
				a.consumerFailed(c.name, "error consuming "+string(kind)+" messages", err)
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(retry):
			}

			err = a.createConsumer(ctx, streamName, c.name, c.filter)
			if err != nil {
				a.consumerFailed(c.name, "failed to create "+string(kind)+" consumer", err)

				continue
			}
			a.consumerRecovered(c.name)
		}
	}()
}

// consumerFailed logs the failure of the consumer name and records it for
// the ConsumersUnavailable condition.
func (a *Agent) consumerFailed(
	name string,
	msg string,
	err error,
) {
	a.logger.Error(
		msg,
		slog.String("consumer", name),
		slog.String("error", err.Error()),
	)

	a.consumerErrMu.Lock()
	defer a.consumerErrMu.Unlock()
	if a.consumerErrs == nil {
		a.consumerErrs = make(map[string]string)
	}
	a.consumerErrs[name] = err.Error()
}

// consumerRecovered removes the consumer name from the failed consumers.
func (a *Agent) consumerRecovered(
	name string,
) {
	a.consumerErrMu.Lock()
	defer a.consumerErrMu.Unlock()
	delete(a.consumerErrs, name)
}

// consumersCondition is the ConsumersUnavailable condition of the agent,
// for the heartbeat. It holds while a job consumer of the running set
// cannot be created or bound, so that the agent can miss jobs for it. The
// scheduling state stays as it is: it is the drain state that an operator
// sets, and the condition is what says the agent is not ready.
func (a *Agent) consumersCondition() job.Condition {
	a.consumerErrMu.Lock()
	failed := len(a.consumerErrs)
	names := make([]string, 0, failed)
	for name := range a.consumerErrs {
		names = append(names, name)
	}
	sort.Strings(names)
	var reason string
	if failed > 0 {
		reason = fmt.Sprintf(
			"%d job consumers do not run; %s: %s",
			failed, names[0], a.consumerErrs[names[0]],
		)
	}
	a.consumerErrMu.Unlock()

	return job.Condition{
		Type:   types.ConditionConsumersUnavailable,
		Status: failed > 0,
		Reason: reason,
		LastTransitionTime: transitionTime(
			types.ConditionConsumersUnavailable, failed > 0, a.prevConditions,
		),
	}
}

// State returns the agent's scheduling state. Read by the heartbeat on its own
// goroutine while a drain or an enrollment acceptance writes it.
func (a *Agent) State() types.AgentState {
	a.lifecycleMu.RLock()
	defer a.lifecycleMu.RUnlock()

	return a.state
}

// setState records a state that carries no consumer change with it. A
// transition that does own one takes lifecycleMu itself and sets the field
// under it, so the two cannot be observed apart.
func (a *Agent) setState(
	state types.AgentState,
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

	if a.state != types.AgentReady {
		a.logger.Info(
			"job consumers not started",
			slog.String("state", string(a.state)),
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

// handleJobMessageJS is the JetStream handler of a job message. ctx is the
// consumer set's: a config apply and a drain cancel the set so that it takes
// no new job. A job that started runs on, with the set's trace and values
// but not its cancel, and sends its result (osapi-io/osapi#1487). Stop waits
// for it, up to the deadline that it has.
func (a *Agent) handleJobMessageJS(
	ctx context.Context,
	msg jetstream.Msg,
) error {
	return a.handleJobMessage(context.WithoutCancel(ctx), msg)
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

	// An idle consumer, such as one of a host that has a new machine ID,
	// is removed by NATS after this time. Empty keeps it.
	inactiveThreshold, _ := time.ParseDuration(a.appConfig.Agent.Consumer.InactiveThreshold)

	// Parse replay policy
	var replayPolicy jetstream.ReplayPolicy
	if a.appConfig.Agent.Consumer.ReplayPolicy == "original" {
		replayPolicy = jetstream.ReplayOriginalPolicy
	} else {
		replayPolicy = jetstream.ReplayInstantPolicy
	}

	// A consumer gets only the jobs that come after NATS makes it. A new
	// consumer, such as one of a new host or a new label, thus does not run
	// the jobs that the stream kept from before it existed. A durable
	// consumer keeps its position, so an agent that starts again gets the
	// jobs that came while it was gone. NATS does not change the deliver
	// policy of a consumer that exists.
	consumerConfig := jetstream.ConsumerConfig{
		Durable:           consumerName,
		FilterSubject:     filterSubject,
		AckPolicy:         jetstream.AckExplicitPolicy,
		DeliverPolicy:     jetstream.DeliverNewPolicy,
		MaxDeliver:        a.appConfig.Agent.Consumer.MaxDeliver,
		AckWait:           ackWait,
		BackOff:           backOff,
		MaxAckPending:     a.appConfig.Agent.Consumer.MaxAckPending,
		ReplayPolicy:      replayPolicy,
		InactiveThreshold: inactiveThreshold,
	}

	return a.jobClient.CreateOrUpdateConsumer(ctx, streamName, consumerConfig)
}
