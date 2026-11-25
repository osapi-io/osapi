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
	"encoding/base64"
	"fmt"
	"time"

	"github.com/osapi-io/osapi/internal/controller/admission"
	"github.com/osapi-io/osapi/internal/pki"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/osapi-io/osapi/internal/controller/api/apierr"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/internal/job"
)

// List is the agents that report.
func List(
	ctx context.Context,
	d Deps,
	_ ListInput,
) (ListAgentsResponse, error) {
	agents, err := d.Jobs.ListAgents(ctx)
	if err != nil {
		return ListAgentsResponse{}, err
	}

	infos := make([]AgentInfo, 0, len(agents))
	for i := range agents {
		infos = append(infos, buildAgentInfo(&agents[i]))
	}

	return ListAgentsResponse{Items: infos, TotalItems: len(infos)}, nil
}

// Get is one agent.
func Get(
	ctx context.Context,
	d Deps,
	in GetInput,
) (AgentInfo, error) {
	info, err := d.Jobs.GetAgent(ctx, in.Hostname)
	if err != nil {
		if apierr.IsMissing(err) {
			return AgentInfo{}, endpoint.Errorf(ErrNotFound, "%s", err.Error())
		}

		return AgentInfo{}, err
	}

	return buildAgentInfo(info), nil
}

// Drain stops an agent taking new jobs. A failure of the registry or of a
// write is a 500; only an agent already in the state is a 409.
func Drain(
	ctx context.Context,
	d Deps,
	in DrainInput,
) (struct{}, error) {
	info, err := lookup(ctx, d, in.Hostname)
	if err != nil {
		return struct{}{}, err
	}
	if info.State == types.AgentDraining || info.State == types.AgentCordoned {
		return struct{}{}, endpoint.Errorf(ErrState,
			"agent %s is already in %s state", in.Hostname, info.State)
	}
	if err := d.Jobs.SetDrainFlag(ctx, info.MachineID); err != nil {
		return struct{}{}, fmt.Errorf("failed to set drain flag: %w", err)
	}
	if err := d.Jobs.WriteAgentTimelineEvent(
		ctx, in.Hostname, "drain", "Drain initiated via API",
	); err != nil {
		return struct{}{}, err
	}

	return struct{}{}, nil
}

// Undrain lets a drained agent take jobs again.
func Undrain(
	ctx context.Context,
	d Deps,
	in UndrainInput,
) (struct{}, error) {
	info, err := lookup(ctx, d, in.Hostname)
	if err != nil {
		return struct{}{}, err
	}
	if info.State != types.AgentDraining && info.State != types.AgentCordoned {
		return struct{}{}, endpoint.Errorf(ErrState,
			"agent %s is not in draining or cordoned state (current: %s)",
			in.Hostname, info.State)
	}
	if err := d.Jobs.DeleteDrainFlag(ctx, info.MachineID); err != nil {
		return struct{}{}, fmt.Errorf("failed to delete drain flag: %w", err)
	}
	if err := d.Jobs.WriteAgentTimelineEvent(
		ctx, in.Hostname, "undrain", "Undrain initiated via API",
	); err != nil {
		return struct{}{}, err
	}

	return struct{}{}, nil
}

// lookup is the agent with the hostname: ErrNotFound when the registry has
// none, and the registry's own error, a 500, when it cannot be read.
func lookup(
	ctx context.Context,
	d Deps,
	hostname string,
) (*job.AgentInfo, error) {
	info, err := d.Jobs.GetAgent(ctx, hostname)
	if err != nil {
		if apierr.IsMissing(err) {
			return nil, endpoint.Errorf(ErrNotFound, "agent not found: %s", hostname)
		}

		return nil, err
	}

	return info, nil
}

// buildAgentInfo is the registry's record of an agent in the API's shape.
// An empty or zero value is left out, as the hand-written handler left it
// out; key_stored and verified are always there.
func buildAgentInfo(
	a *job.AgentInfo,
) AgentInfo {
	keyStored, verified := a.KeyStored, a.Verified
	info := AgentInfo{
		MachineID:        a.MachineID,
		Fingerprint:      a.Fingerprint,
		Version:          a.AgentVersion,
		Hostname:         a.Hostname,
		RegisteredAt:     a.RegisteredAt,
		StartedAt:        a.StartedAt,
		Architecture:     a.Architecture,
		KernelVersion:    a.KernelVersion,
		CPUCount:         a.CPUCount,
		FQDN:             a.FQDN,
		ServiceMgr:       a.ServiceMgr,
		PackageMgr:       a.PackageMgr,
		PrimaryInterface: a.PrimaryInterface,
		State:            job.EffectiveState(a.State, false),
		KeyStored:        &keyStored,
		Verified:         &verified,
	}
	if len(a.Labels) > 0 {
		info.Labels = a.Labels
	}
	if len(a.Facts) > 0 {
		info.Facts = a.Facts
	}
	if len(a.MissingPrograms) > 0 {
		info.MissingPrograms = a.MissingPrograms
	}
	if a.Uptime > 0 {
		info.Uptime = formatDuration(a.Uptime)
	}
	if a.OSInfo != nil {
		info.OSInfo = &OSInfoResponse{
			Distribution: a.OSInfo.Distribution,
			Version:      a.OSInfo.Version,
			Family:       a.OSInfo.Family,
		}
	}
	if a.LoadAverages != nil {
		info.LoadAverage = &LoadAverageResponse{
			OneMin:     a.LoadAverages.Load1,
			FiveMin:    a.LoadAverages.Load5,
			FifteenMin: a.LoadAverages.Load15,
		}
	}
	if a.MemoryStats != nil {
		info.Memory = &MemoryResponse{
			Total: uint64ToInt(a.MemoryStats.Total),
			Free:  uint64ToInt(a.MemoryStats.Free),
			Used:  uint64ToInt(a.MemoryStats.Used()),
		}
	}
	for _, ni := range a.Interfaces {
		info.Interfaces = append(info.Interfaces, NetworkInterfaceResponse{
			Name:   ni.Name,
			IPv4:   ni.IPv4,
			IPv6:   ni.IPv6,
			MAC:    ni.MAC,
			Family: types.AddressFamily(ni.Family),
		})
	}
	for _, r := range a.Routes {
		info.Routes = append(info.Routes, RouteResponse{
			Destination: r.Destination, Gateway: r.Gateway, Interface: r.Interface,
			Mask: r.Mask, Metric: r.Metric,
		})
	}
	// A condition that does not hold is not in the answer. The registry
	// keeps it, for the transition time and the notifier.
	for _, c := range a.Conditions {
		if !c.Status {
			continue
		}
		info.Conditions = append(info.Conditions, NodeCondition{
			Type: c.Type, Reason: c.Reason, LastTransitionTime: c.LastTransitionTime,
		})
	}
	for _, te := range a.Timeline {
		info.Timeline = append(info.Timeline, TimelineEvent{
			Timestamp: te.Timestamp, Event: te.Event, Hostname: te.Hostname,
			Message: te.Message, Error: te.Error,
		})
	}

	return info
}

// formatDuration is an uptime as days, hours and minutes.
func formatDuration(
	d time.Duration,
) string {
	totalMinutes := int(d.Minutes())
	days := totalMinutes / (24 * 60)
	hours := (totalMinutes % (24 * 60)) / 60
	minutes := totalMinutes % 60

	return fmt.Sprintf("%d %s, %d %s, %d %s",
		days, plural(days, "day"), hours, plural(hours, "hour"), minutes, plural(minutes, "minute"))
}

func plural(
	n int,
	word string,
) string {
	if n == 1 {
		return word
	}

	return word + "s"
}

// uint64ToInt is value as an int, at most the largest int.
func uint64ToInt(
	value uint64,
) int {
	maxInt := int(^uint(0) >> 1)
	if value > uint64(maxInt) {
		return maxInt
	}

	return int(value)
}

// Join admits the agent of in at once, when its token admits, and answers the
// controller's public key.
func Join(
	ctx context.Context,
	d Deps,
	in JoinInput,
) (JoinResponse, error) {
	if d.Joiner == nil {
		return JoinResponse{}, admission.ErrNoJoinTokens
	}
	// The input's base64 rule refuses a key or a signature that is not.
	pub, _ := base64.StdEncoding.DecodeString(in.Body.PublicKey)
	sig, _ := base64.StdEncoding.DecodeString(in.Body.Signature)
	key, err := d.Joiner.Join(ctx, pki.JoinRequest{
		Token:         string(in.Body.Token),
		MachineID:     in.Body.MachineID,
		Hostname:      in.Body.Hostname,
		PublicKey:     pub,
		EncryptionKey: in.Body.EncryptionKey,
		Signature:     sig,
	})
	if err != nil {
		return JoinResponse{}, err
	}

	return JoinResponse{ControllerPublicKey: base64.StdEncoding.EncodeToString(key)}, nil
}
