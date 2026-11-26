// Copyright (c) 2026 John Dewey

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
	"fmt"
	"maps"
	"strings"

	"github.com/osapi-io/osapi/internal/job"
)

// seenState is what the agent last reported, which a change of it is a
// timeline event against. After a start, its first value is the agent's
// previous registration and facts, which the buckets still hold.
type seenState struct {
	regLoaded   bool
	os          string
	labels      map[string]string
	conditions  map[string]bool
	factsLoaded bool
	address     string
	// configVersion is the version of the agent's config at the last note:
	// a change of the labels with a new version is config_applied's, not a
	// labels_changed of its own.
	configVersion int
}

// osName is the OS of reg as the timeline shows it, such as ubuntu 24.04.
func osName(
	reg job.AgentRegistration,
) string {
	if reg.OSInfo == nil {
		return ""
	}

	return strings.TrimSpace(reg.OSInfo.Distribution + " " + reg.OSInfo.Version)
}

// holding is the type of each condition that holds in conditions.
func holding(
	conditions []job.Condition,
) map[string]bool {
	out := map[string]bool{}
	for _, c := range conditions {
		if c.Status {
			out[string(c.Type)] = true
		}
	}

	return out
}

// noteRegistration writes a timeline event for each change of reg against
// what the agent last reported: its OS, its labels, and each condition that
// starts or clears. A change of the labels that the controller's config
// made is in the config_applied event, so it has no event of its own.
func (a *Agent) noteRegistration(
	ctx context.Context,
	reg job.AgentRegistration,
) {
	if !a.seen.regLoaded {
		a.seen.regLoaded = true
		a.loadSeenRegistration(ctx, reg.MachineID)
	}
	if now := osName(reg); now != "" && a.seen.os != "" && now != a.seen.os {
		a.lifecycle(ctx, "os_changed", a.seen.os+" → "+now)
	}
	if osName(reg) != "" {
		a.seen.os = osName(reg)
	}
	byConfig := a.configVersion != a.seen.configVersion
	a.seen.configVersion = a.configVersion
	if a.seen.labels != nil && !byConfig && !maps.Equal(a.seen.labels, reg.Labels) {
		a.lifecycle(ctx, "labels_changed", labelString(a.seen.labels)+" → "+labelString(reg.Labels))
	}
	a.seen.labels = maps.Clone(reg.Labels)
	if a.seen.labels == nil {
		a.seen.labels = map[string]string{}
	}

	now := holding(reg.Conditions)
	known := a.seen.conditions != nil
	for _, c := range reg.Conditions {
		if !known {
			break
		}
		t := string(c.Type)
		switch {
		case c.Status && !a.seen.conditions[t]:
			a.lifecycle(ctx, "condition_started", strings.TrimSpace(t+": "+c.Reason))
		case !c.Status && a.seen.conditions[t]:
			a.lifecycle(ctx, "condition_cleared", t)
		}
	}
	a.seen.conditions = now
}

// loadSeenRegistration takes the agent's previous registration as what it
// last reported. With none, the first heartbeat is the baseline.
func (a *Agent) loadSeenRegistration(
	ctx context.Context,
	machineID string,
) {
	if a.registryBucket == "" {
		return
	}
	entry, err := a.natsClient.KVGet(ctx, a.registryBucket, registryKey(machineID))
	if err != nil {
		return
	}
	var prev job.AgentRegistration
	if unmarshalJSON(entry.Value(), &prev) != nil {
		return
	}
	a.seen.os = osName(prev)
	a.seen.labels = prev.Labels
	if a.seen.labels == nil {
		a.seen.labels = map[string]string{}
	}
	a.seen.conditions = holding(prev.Conditions)
}

// primaryAddress is the IPv4 address of the interface of the default route,
// as the timeline shows it, such as eth0 10.0.0.5.
func primaryAddress(
	f job.FactsRegistration,
) string {
	for _, i := range f.Interfaces {
		if i.Name == f.PrimaryInterface && i.IPv4 != "" {
			return fmt.Sprintf("%s %s", i.Name, i.IPv4)
		}
	}

	return ""
}

// noteFacts writes a timeline event when the primary address of f is not
// the one that the agent last reported.
func (a *Agent) noteFacts(
	ctx context.Context,
	machineID string,
	f job.FactsRegistration,
) {
	if !a.seen.factsLoaded {
		a.seen.factsLoaded = true
		if a.factsBucket != "" {
			if entry, err := a.natsClient.KVGet(
				ctx,
				a.factsBucket,
				factsKey(machineID),
			); err == nil {
				var prev job.FactsRegistration
				if unmarshalJSON(entry.Value(), &prev) == nil {
					a.seen.address = primaryAddress(prev)
				}
			}
		}
	}
	now := primaryAddress(f)
	if now == "" {
		return
	}
	if a.seen.address != "" && now != a.seen.address {
		a.lifecycle(ctx, "address_changed", a.seen.address+" → "+now)
	}
	a.seen.address = now
}
