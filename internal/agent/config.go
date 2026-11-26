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
	"log/slog"
	"maps"
	"sort"
	"strings"
)

// applyConfig reads the config that the controller holds for the agent,
// and applies a new version: its labels replace the labels of the agent's
// config file with the same keys, and its max jobs replaces the file's
// when it is set. A change of labels or of max jobs restarts the job
// consumers, on the subjects of the labels and with max jobs in flight,
// under lifecycleMu, as a drain does. The heartbeat calls it, so a change applies within one heartbeat.
func (a *Agent) applyConfig(
	ctx context.Context,
) {
	cfg, err := a.jobClient.GetAgentConfig(ctx, a.machineID)
	if err != nil {
		a.logger.Warn("failed to read the agent's config", slog.String("error", err.Error()))

		return
	}
	if cfg.Version == a.configVersion {
		return
	}
	labels := maps.Clone(a.fileLabels)
	if labels == nil {
		labels = map[string]string{}
	}
	maps.Copy(labels, cfg.Labels)
	maxJobs := a.fileMaxJobs
	if cfg.MaxJobs > 0 {
		maxJobs = cfg.MaxJobs
	}

	a.lifecycleMu.Lock()
	a.configVersion = cfg.Version
	changed := !maps.Equal(labels, a.appConfig.Agent.Labels) ||
		maxJobs != a.appConfig.Agent.MaxJobs
	if changed {
		running := a.consumerCancel != nil
		if running {
			a.cancelConsumersLocked()
		}
		a.appConfig.Agent.Labels = labels
		a.appConfig.Agent.MaxJobs = maxJobs
		if running {
			a.startConsumersLocked()
		}
	}
	a.lifecycleMu.Unlock()

	a.logger.Info("applied the agent's config",
		slog.Int("version", cfg.Version), slog.Any("labels", labels),
		slog.Int("max_jobs", maxJobs))
	if changed {
		a.lifecycle(ctx, "config_applied",
			fmt.Sprintf("config version %d, labels %s, max jobs %d",
				cfg.Version, labelString(labels), maxJobs))
	}
}

// labelString is labels as k=v, in order.
func labelString(
	labels map[string]string,
) string {
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+labels[k])
	}

	return strings.Join(parts, ",")
}
