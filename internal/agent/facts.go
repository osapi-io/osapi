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
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/osapi-io/gohai/pkg/gohai"
	"github.com/osapi-io/gohai/pkg/gohai/collectors/network"

	"github.com/osapi-io/osapi/internal/job"
)

// factCollectors are the gohai collectors whose facts the agent stores.
var factCollectors = []string{
	"kernel", "hostname", "cpu", "init", "package_mgr", "network", "virtualization",
}

// collectFactsFn collects the facts of the host (injectable for testing).
var collectFactsFn = collectFacts

// collectFacts collects the facts of the host with gohai's factCollectors.
func collectFacts() (*gohai.Facts, error) {
	return gohaiFacts(gohai.WithCollectors(factCollectors...))
}

// defaultFactsInterval is the fallback fact refresh period when no config
// value is set or the configured value is unparseable.
var defaultFactsInterval = 60 * time.Second

// startFacts writes the initial facts, spawns a goroutine that
// refreshes the entry on a ticker, and stops on ctx.Done().
func (a *Agent) startFacts(
	ctx context.Context,
	machineID string,
	hostname string,
) {
	if a.factsBucket == "" {
		return
	}

	a.writeFacts(ctx, machineID, hostname)

	interval := defaultFactsInterval
	if cfgInterval := a.appConfig.Agent.Facts.Interval; cfgInterval != "" {
		if parsed, err := time.ParseDuration(cfgInterval); err == nil {
			interval = parsed
		}
	}

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				a.writeFacts(ctx, machineID, hostname)
			}
		}
	}()
}

// writeFacts collects system facts from providers and writes them to the
// facts KV bucket. Provider errors are non-fatal; the facts entry is still
// written with whatever data was gathered.
func (a *Agent) writeFacts(
	ctx context.Context,
	machineID string,
	hostname string,
) {
	reg := job.FactsRegistration{}
	if f, err := collectFactsFn(); err == nil {
		reg = factsFromGohai(f)
		a.cpuCount = reg.CPUCount
	} else {
		a.factsLogger.Warn(
			"failed to collect facts",
			slog.String("machine_id", machineID),
			slog.String("error", err.Error()),
		)
	}

	a.cachedFacts.Store(&reg)
	// A change of the primary address is an event of the timeline.
	a.noteFacts(ctx, machineID, reg)

	data, err := marshalJSON(reg)
	if err != nil {
		a.factsLogger.Warn(
			"failed to marshal facts",
			slog.String("machine_id", machineID),
			slog.String("hostname", hostname),
			slog.String("error", err.Error()),
		)
		return
	}

	key := factsKey(machineID)
	if _, err := a.natsClient.KVPut(ctx, a.factsBucket, key, data); err != nil {
		a.factsLogger.Warn(
			"failed to write facts",
			slog.String("machine_id", machineID),
			slog.String("hostname", hostname),
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
	}
}

// GetFacts returns the agent's current facts as a flat map suitable for
// template rendering. Returns nil if facts haven't been collected yet.
// Uses JSON round-trip so the map automatically includes all fields
// from FactsRegistration without hardcoding field names.
func (a *Agent) GetFacts() map[string]any {
	facts := a.cachedFacts.Load()
	if facts == nil {
		return nil
	}

	data, err := marshalJSON(facts)
	if err != nil {
		return nil
	}

	var result map[string]any
	if err := unmarshalJSON(data, &result); err != nil {
		return nil
	}

	return result
}

// agentFacts gives a behavior's host the facts of an agent that is made
// after the host. Its get returns nil until agent is set.
type agentFacts struct {
	agent *Agent
}

// get returns the facts of the agent, or nil when the agent is not set.
func (f *agentFacts) get() map[string]any {
	if f.agent == nil {
		return nil
	}

	return f.agent.GetFacts()
}

// factsKey returns the KV key for an agent's facts entry.
// Uses machineID as the key component for stable identity across hostname changes.
func factsKey(
	machineID string,
) string {
	return "facts." + job.SanitizeHostname(machineID)
}

// factsFromGohai is the facts entry of gohai's facts. A collector that
// failed leaves its fields empty.
func factsFromGohai(
	f *gohai.Facts,
) job.FactsRegistration {
	reg := job.FactsRegistration{Containerized: f.Virtualization.InContainer()}
	if f.Kernel != nil {
		reg.Architecture, reg.KernelVersion = f.Kernel.Machine, f.Kernel.Release
	}
	if f.Hostname != nil {
		reg.FQDN = f.Hostname.FQDN
	}
	if f.CPU != nil {
		reg.CPUCount = f.CPU.Count
	}
	if f.Init != nil {
		reg.ServiceMgr = f.Init.Name
	}
	if f.PackageMgr != nil {
		reg.PackageMgr = f.PackageMgr.Name
	}
	if f.Network != nil {
		reg.Interfaces = interfacesFromGohai(f.Network.Interfaces)
		reg.Routes = routesFromGohai(f.Network.Routes)
		reg.PrimaryInterface = f.Network.DefaultInterface
	}

	return reg
}

// interfacesFromGohai is each interface that is up and not loopback, with
// its first IPv4 and first IPv6 address.
func interfacesFromGohai(
	ifs []network.Interface,
) []job.NetworkInterface {
	var out []job.NetworkInterface
	for _, i := range ifs {
		if i.State != "up" || slices.Contains(i.Flags, "loopback") {
			continue
		}
		ni := job.NetworkInterface{Name: i.Name, MAC: i.MAC}
		for _, a := range i.Addresses {
			switch {
			case a.Family == "inet" && ni.IPv4 == "":
				ni.IPv4 = a.Addr
			case a.Family == "inet6" && ni.IPv6 == "":
				ni.IPv6 = a.Addr
			}
		}
		switch {
		case ni.IPv4 != "" && ni.IPv6 != "":
			ni.Family = "dual"
		case ni.IPv6 != "":
			ni.Family = "inet6"
		case ni.IPv4 != "":
			ni.Family = "inet"
		}
		out = append(out, ni)
	}

	return out
}

// routesFromGohai is each IPv4 route, its destination split from its
// mask: "10.0.0.0/24" is "10.0.0.0" and "/24", and the default route is
// "0.0.0.0" and "/0".
func routesFromGohai(
	routes []network.Route,
) []job.Route {
	var out []job.Route
	for _, r := range routes {
		if r.Family != "inet" {
			continue
		}
		dest, mask := r.Destination, ""
		if dest == "default" {
			dest, mask = "0.0.0.0", "/0"
		} else if d, bits, ok := strings.Cut(dest, "/"); ok {
			dest, mask = d, "/"+bits
		}
		out = append(out, job.Route{
			Destination: dest, Gateway: r.Gateway, Interface: r.Interface,
			Mask: mask, Metric: r.Metric,
		})
	}

	return out
}
