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
	"encoding/json"
	"log/slog"
	"time"

	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/cryptoreport"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/pki"
	"github.com/osapi-io/osapi/internal/provider/node/disk"
	"github.com/osapi-io/osapi/internal/provider/node/load"
	"github.com/osapi-io/osapi/internal/provider/node/memory"
	"github.com/osapi-io/osapi/internal/telemetry/process"
)

// heartbeatInterval is the interval between heartbeat refreshes.
var heartbeatInterval = 10 * time.Second

// seenInterval is how often the heartbeat records when the agent last
// reported, for admission: often enough to say when an offline agent was
// last seen, and a sixth of the writes of the heartbeat itself.
var seenInterval = time.Minute

// getAgentHostnameFn is a package-level variable for testing hostname re-reads.
var getAgentHostnameFn = job.GetAgentHostname

// marshalJSON is a package-level variable for testing the marshal error path.
var marshalJSON = json.Marshal

// unmarshalJSON is a package-level variable for testing the unmarshal error path.
var unmarshalJSON = json.Unmarshal

// startHeartbeat writes the initial registration, spawns a goroutine that
// refreshes the entry on a ticker, and deregisters on ctx.Done().
func (a *Agent) startHeartbeat(
	ctx context.Context,
	machineID string,
	hostname string,
) {
	if a.registryBucket == "" {
		return
	}

	ttl := a.appConfig.NATS.Registry.TTL
	key := registryKey(machineID)

	a.writeRegistration(ctx, machineID, hostname)

	a.heartbeatLogger.Info(
		"registered in agent registry",
		slog.String("machine_id", machineID),
		slog.String("hostname", hostname),
		slog.String("key", key),
		slog.String("ttl", ttl),
	)

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()

		ticker := time.NewTicker(heartbeatInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				a.deregister(machineID)
				return
			case done := <-a.renamed:
				hostname = a.refreshHostname(hostname)
				a.writeRegistration(ctx, machineID, hostname)
				close(done)
			case <-ticker.C:
				hostname = a.refreshHostname(hostname)
				a.writeRegistration(ctx, machineID, hostname)
				a.retryEnrollmentIfPending()

				a.heartbeatLogger.Info(
					"heartbeat refreshed",
					slog.String("machine_id", machineID),
					slog.String("hostname", hostname),
					slog.String("key", key),
					slog.String("next_in", heartbeatInterval.String()),
				)
			}
		}
	}()
}

// currentHostname is the host's name as the agent last read it, or "" before
// Start reads it.
func (a *Agent) currentHostname() string {
	if h := a.hostname.Load(); h != nil {
		return *h
	}

	return ""
}

// setHostname keeps hostname as the host's name.
func (a *Agent) setHostname(
	hostname string,
) {
	a.hostname.Store(&hostname)
}

// refreshHostname is the host's hostname now, and logs and keeps it when it
// is not hostname. Consumers use the machine ID, which does not change, so
// nothing resubscribes: only the registry and the logs take the new name. A
// failed read keeps hostname, so a host does not answer as "unknown" for a
// moment.
func (a *Agent) refreshHostname(
	hostname string,
) string {
	newHostname, err := getAgentHostnameFn(a.appConfig.Agent.Hostname)
	if err != nil {
		a.heartbeatLogger.Warn(
			"failed to read the hostname, keeping the last one",
			slog.String("hostname", hostname),
			slog.String("error", err.Error()),
		)

		return hostname
	}
	if newHostname != hostname {
		a.heartbeatLogger.Info(
			"hostname changed",
			slog.String("old_hostname", hostname),
			slog.String("new_hostname", newHostname),
		)
		a.setHostname(newHostname)
	}

	return newHostname
}

// writeRegistration marshals an agent registration and puts it to the registry KV.
// Provider errors are non-fatal; the heartbeat still writes with whatever data it gathered.
func (a *Agent) writeRegistration(
	ctx context.Context,
	machineID string,
	hostname string,
) {
	a.handleDrainDetection(ctx, machineID, hostname)
	a.reloadKeys()
	a.applyConfig(ctx)
	if a.pkiManager != nil {
		a.followControllerKey(ctx)
		a.followNATSCert(ctx)
	}
	a.writeSeen(ctx, machineID)
	a.refreshPrograms()

	certs := a.certificates()
	report := cryptoreport.Agent(a.appConfig, certs)
	reg := job.AgentRegistration{
		MachineID:       machineID,
		Hostname:        hostname,
		Labels:          a.appConfig.Agent.Labels,
		RegisteredAt:    time.Now(),
		StartedAt:       a.startedAt,
		State:           a.State(),
		SubComponents:   a.subComponents,
		Fingerprint:     a.pkiFingerprint(),
		AgentVersion:    a.version,
		ConfigVersion:   a.configVersion,
		MissingPrograms: a.missingPrograms,
		Certificates:    certs,
		Crypto:          &report,
	}
	reg.ControllerKey, reg.ControllerKeyNext = a.controllerKeys()
	reg.NATSCertNext = a.natsCertNext

	if info, err := a.hostProvider.GetOSInfo(); err == nil {
		info.Family = detectFn()
		reg.OSInfo = info
	}

	if uptime, err := a.hostProvider.GetUptime(); err == nil {
		reg.Uptime = uptime
	}

	var loadAvg *load.Result
	if avg, err := a.loadProvider.GetAverageStats(); err == nil {
		loadAvg = avg
		reg.LoadAverages = avg
	}

	var memStats *memory.Result
	if stats, err := a.memoryProvider.GetStats(); err == nil {
		memStats = stats
		reg.MemoryStats = stats
	}

	var diskStats []disk.Result
	if stats, err := a.diskProvider.GetLocalUsageStats(); err == nil {
		diskStats = stats
	}

	conditions := []job.Condition{
		evaluateMemoryPressure(
			memStats,
			a.appConfig.Agent.Conditions.MemoryPressureThreshold,
			a.prevConditions,
		),
		evaluateHighLoad(
			loadAvg,
			int(a.cpuCount.Load()),
			a.appConfig.Agent.Conditions.HighLoadMultiplier,
			a.prevConditions,
		),
		evaluateDiskPressure(
			diskStats,
			a.appConfig.Agent.Conditions.DiskPressureThreshold,
			a.prevConditions,
		),
		a.drainCondition(),
		a.consumersCondition(),
	}
	a.prevConditions = conditions

	if a.processProvider != nil {
		if pm, err := a.processProvider.GetMetrics(); err == nil {
			reg.Process = &job.ProcessMetrics{
				CPUPercent: pm.CPUPercent,
				RSSBytes:   pm.RSSBytes,
				Goroutines: pm.Goroutines,
			}

			processConditions := process.EvaluateProcessConditions(
				pm,
				process.ConditionThresholds{
					MemoryPressureBytes: a.appConfig.Agent.ProcessConditions.MemoryPressureBytes,
					HighCPUPercent:      a.appConfig.Agent.ProcessConditions.HighCPUPercent,
				},
				a.prevProcessConditions,
			)
			a.prevProcessConditions = processConditions
			conditions = append(conditions, processConditions...)
		}
	}

	reg.Conditions = conditions
	// A change of the OS, the labels or a condition is an event of the
	// agent's timeline.
	a.noteRegistration(ctx, reg)

	// Sign the routing fields. Without this the controller cannot tell this
	// registration from one written by anything else that can reach the
	// bucket, which is what lets a hostname be claimed by a machine that
	// never enrolled under it. Unsigned when PKI is disabled, which leaves
	// behaviour as it was.
	if a.pkiManager != nil {
		reg.Signature = a.pkiManager.Sign(job.RegistrationSigningBytes(&reg))
	}

	data, err := marshalJSON(reg)
	if err != nil {
		a.heartbeatLogger.Warn(
			"failed to marshal agent registration",
			slog.String("machine_id", machineID),
			slog.String("hostname", hostname),
			slog.String("error", err.Error()),
		)
		return
	}

	key := registryKey(machineID)
	if _, err := a.natsClient.KVPut(ctx, a.registryBucket, key, data); err != nil {
		a.heartbeatLogger.Warn(
			"failed to write agent registration",
			slog.String("machine_id", machineID),
			slog.String("hostname", hostname),
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
	} else {
		a.lastHeartbeatTime.Store(time.Now())
	}
}

// deregister deletes the agent's registration key on clean shutdown.
func (a *Agent) deregister(
	machineID string,
) {
	key := registryKey(machineID)
	if err := a.natsClient.KVDelete(context.Background(), a.registryBucket, key); err != nil {
		a.heartbeatLogger.Warn(
			"failed to deregister agent",
			slog.String("machine_id", machineID),
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
		return
	}

	a.heartbeatLogger.Info(
		"agent deregistered",
		slog.String("machine_id", machineID),
		slog.String("key", key),
	)
}

// pkiFingerprint returns the agent's PKI public key fingerprint,
// or empty string when PKI is disabled.
func (a *Agent) pkiFingerprint() string {
	if a.pkiManager != nil {
		return a.pkiManager.Fingerprint()
	}

	return ""
}

// registryKey returns the KV key for an agent's registration entry.
// Uses machineID as the key component for stable identity across hostname changes.
func registryKey(
	machineID string,
) string {
	return "agents." + job.SanitizeHostname(machineID)
}

// writeSeen records when the agent last reported, at most once each
// seenInterval. A failed write is logged and tried again at the next
// heartbeat: the registration is what keeps the agent in the fleet.
func (a *Agent) writeSeen(
	ctx context.Context,
	machineID string,
) {
	if time.Since(a.seenWrittenAt) < seenInterval {
		return
	}
	if err := a.jobClient.WriteAgentSeen(ctx, machineID); err != nil {
		a.heartbeatLogger.Debug(
			"failed to record last seen",
			slog.String("error", err.Error()),
		)

		return
	}
	a.seenWrittenAt = time.Now()
}

// certificates are the certificates that the agent serves, as its
// registration reports them, or none for an agent with no file system.
func (a *Agent) certificates() []pki.Cert {
	if a.appFs == nil {
		return nil
	}

	return pki.ReadCerts(a.appFs, config.CertSourcesOf(a.appConfig, "agent"), time.Now())
}
