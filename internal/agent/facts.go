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
	"time"

	"github.com/retr0h/osapi/internal/job"
)

// factsInterval controls the fact refresh period.
var factsInterval = 60 * time.Second

// startFacts writes the initial facts, spawns a goroutine that
// refreshes the entry on a ticker, and stops on ctx.Done().
func (a *Agent) startFacts(
	ctx context.Context,
	hostname string,
) {
	if a.factsKV == nil {
		return
	}

	a.writeFacts(ctx, hostname)

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()

		ticker := time.NewTicker(factsInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				a.writeFacts(ctx, hostname)
			}
		}
	}()
}

// writeFacts collects system facts from providers and writes them to the
// facts KV bucket. Provider errors are non-fatal; the facts entry is still
// written with whatever data was gathered.
func (a *Agent) writeFacts(
	ctx context.Context,
	hostname string,
) {
	reg := job.FactsRegistration{}

	// Call providers — errors are non-fatal
	if arch, err := a.hostProvider.GetArchitecture(); err == nil {
		reg.Architecture = arch
	}

	if kv, err := a.hostProvider.GetKernelVersion(); err == nil {
		reg.KernelVersion = kv
	}

	if fqdn, err := a.hostProvider.GetFQDN(); err == nil {
		reg.FQDN = fqdn
	}

	if count, err := a.hostProvider.GetCPUCount(); err == nil {
		reg.CPUCount = count
	}

	if mgr, err := a.hostProvider.GetServiceManager(); err == nil {
		reg.ServiceMgr = mgr
	}

	if mgr, err := a.hostProvider.GetPackageManager(); err == nil {
		reg.PackageMgr = mgr
	}

	if ifaces, err := a.netinfoProvider.GetInterfaces(); err == nil {
		reg.Interfaces = ifaces
	}

	data, err := marshalJSON(reg)
	if err != nil {
		a.logger.Warn(
			"failed to marshal facts",
			slog.String("hostname", hostname),
			slog.String("error", err.Error()),
		)
		return
	}

	key := factsKey(hostname)
	if _, err := a.factsKV.Put(ctx, key, data); err != nil {
		a.logger.Warn(
			"failed to write facts",
			slog.String("hostname", hostname),
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
	}
}

// factsKey returns the KV key for an agent's facts entry.
func factsKey(
	hostname string,
) string {
	return "facts." + job.SanitizeHostname(hostname)
}
