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
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/nats-io/nats.go/jetstream"
	natsclient "github.com/osapi-io/nats-client/pkg/client"

	"github.com/osapi-io/osapi/internal/job"
)

// consumeJobs starts the job consumers of kind: one for the jobs of any
// agent, one for the jobs of all agents, one for the jobs of the agent's
// machine ID, and one for each prefix of each label. For "group:
// web.dev.us-east", the label consumers filter on
// jobs.<kind>.label.group.web, jobs.<kind>.label.group.web.dev and
// jobs.<kind>.label.group.web.dev.us-east. Direct targeting uses the
// machine ID, which is permanent, not the hostname, which can change.
func (a *Agent) consumeJobs(
	ctx context.Context,
	machineID string,
	kind types.JobType,
) {
	for _, c := range a.jobConsumers(machineID, kind) {
		a.startConsumer(ctx, a.streamName, c, kind)
	}
}

// jobConsumers is the consumers of kind that consumeJobs starts, in a
// fixed order: the labels by key.
func (a *Agent) jobConsumers(
	machineID string,
	kind types.JobType,
) []consumerSpec {
	subjectPrefix := job.JobsQueryPrefix
	if kind == types.JobTypeModify {
		subjectPrefix = job.JobsModifyPrefix
	}
	// A consumer name has only letters, digits and underscores.
	id := job.SanitizeHostname(machineID)
	name := a.consumerNamePrefix() + string(kind) + "_"

	consumers := []consumerSpec{
		{name: name + "any_" + id, filter: subjectPrefix + "._any"},
		{name: name + "all_" + id, filter: subjectPrefix + "._all"},
		{name: name + "direct_" + id, filter: subjectPrefix + ".host." + id},
	}
	labels := a.appConfig.Agent.Labels
	for _, key := range slices.Sorted(maps.Keys(labels)) {
		segments := strings.Split(labels[key], ".")
		for i := range segments {
			label := key + "." + strings.Join(segments[:i+1], ".")
			consumers = append(consumers, consumerSpec{
				name:   name + "label_" + labelConsumerName(label) + "_" + id,
				filter: subjectPrefix + ".label." + label,
			})
		}
	}

	return consumers
}

// labelConsumerName is the part of a consumer name for the label subject
// key.prefix. The sanitized text keeps the name readable, and a hash of the
// subject keeps two labels that sanitize to the same text, such as
// group.web.dev and group.web-dev, from sharing one consumer: a
// CreateOrUpdate of one would replace the filter of the other.
func labelConsumerName(
	label string,
) string {
	sum := sha256.Sum256([]byte(label))

	return job.SanitizeHostname(label) + "_" + hex.EncodeToString(sum[:4])
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
// consumerRetryInterval until it runs or ctx ends. Each try, the first
// too, is in the goroutine: the caller can hold lifecycleMu, and a request
// to a slow NATS must not hold State, a drain or a config apply.
func (a *Agent) startConsumer(
	ctx context.Context,
	streamName string,
	c consumerSpec,
	kind types.JobType,
) {
	// The caller holds lifecycleMu, or no set runs yet.
	gen := a.consumerGen
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
		defer a.consumerRecovered(c.name, gen)

		err := a.createConsumer(ctx, streamName, c.name, c.filter)
		if err != nil {
			a.consumerFailed(c.name, gen, "failed to create "+string(kind)+" consumer", err)
		}
		for {
			if err == nil {
				err = a.jobClient.ConsumeJobs(ctx, streamName, c.name, a.handleJobMessageJS, opts)
				if ctx.Err() != nil {
					return
				}
				// The set still runs, so the consumer must too: a consume
				// that ended with no error is a failure, and it is tried
				// again.
				if err == nil {
					err = errConsumeEnded
				}
				a.consumerFailed(c.name, gen, "error consuming "+string(kind)+" messages", err)
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(retry):
			}

			err = a.createConsumer(ctx, streamName, c.name, c.filter)
			if err != nil {
				a.consumerFailed(c.name, gen, "failed to create "+string(kind)+" consumer", err)

				continue
			}
			a.consumerRecovered(c.name, gen)
		}
	}()
}

// errConsumeEnded is the failure of a consume that ended with no error
// while its set still runs.
var errConsumeEnded = errors.New("the consume ended")

// consumerErr is the last failure of a job consumer, with the generation of
// the set that it belongs to.
type consumerErr struct {
	gen uint64
	msg string
}

// consumerFailed logs the failure of the consumer name of the set gen, and
// records it for the ConsumersUnavailable condition.
func (a *Agent) consumerFailed(
	name string,
	gen uint64,
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
		a.consumerErrs = make(map[string]consumerErr)
	}
	if prev, ok := a.consumerErrs[name]; ok && prev.gen > gen {
		// A newer set owns the name.
		return
	}
	a.consumerErrs[name] = consumerErr{gen: gen, msg: err.Error()}
}

// consumerRecovered removes the consumer name of the set gen from the failed
// consumers. A failure of a newer set with the same name stays.
func (a *Agent) consumerRecovered(
	name string,
	gen uint64,
) {
	a.consumerErrMu.Lock()
	defer a.consumerErrMu.Unlock()
	if prev, ok := a.consumerErrs[name]; ok && prev.gen <= gen {
		delete(a.consumerErrs, name)
	}
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
			failed, names[0], a.consumerErrs[names[0]].msg,
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
	a.consumerGen++
	a.consumeJobs(a.consumerCtx, a.machineID, types.JobTypeQuery)
	a.consumeJobs(a.consumerCtx, a.machineID, types.JobTypeModify)
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
