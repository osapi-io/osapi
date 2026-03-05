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
	"fmt"
	"time"

	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/provider/node/disk"
	"github.com/retr0h/osapi/internal/provider/node/load"
	"github.com/retr0h/osapi/internal/provider/node/mem"
)

// findPrevCondition returns the previous condition of the given type,
// or nil if not found.
func findPrevCondition(
	condType string,
	prev []job.Condition,
) *job.Condition {
	for i := range prev {
		if prev[i].Type == condType {
			return &prev[i]
		}
	}
	return nil
}

// transitionTime returns the previous LastTransitionTime if status
// hasn't changed, otherwise returns now.
func transitionTime(
	condType string,
	newStatus bool,
	prev []job.Condition,
) time.Time {
	if p := findPrevCondition(condType, prev); p != nil {
		if p.Status == newStatus {
			return p.LastTransitionTime
		}
	}
	return time.Now()
}

func evaluateMemoryPressure(
	stats *mem.Stats,
	threshold int,
	prev []job.Condition,
) job.Condition {
	c := job.Condition{Type: job.ConditionMemoryPressure}
	if stats == nil || stats.Total == 0 {
		c.LastTransitionTime = transitionTime(c.Type, false, prev)
		return c
	}
	available := stats.Free + stats.Cached
	used := stats.Total - available
	pct := float64(used) / float64(stats.Total) * 100
	c.Status = pct > float64(threshold)
	if c.Status {
		c.Reason = fmt.Sprintf(
			"memory %.0f%% used (%.1f/%.1f GB)",
			pct,
			float64(used)/1024/1024/1024,
			float64(stats.Total)/1024/1024/1024,
		)
	}
	c.LastTransitionTime = transitionTime(c.Type, c.Status, prev)
	return c
}

func evaluateHighLoad(
	loadAvg *load.AverageStats,
	cpuCount int,
	multiplier float64,
	prev []job.Condition,
) job.Condition {
	c := job.Condition{Type: job.ConditionHighLoad}
	if loadAvg == nil || cpuCount == 0 {
		c.LastTransitionTime = transitionTime(c.Type, false, prev)
		return c
	}
	threshold := float64(cpuCount) * multiplier
	c.Status = float64(loadAvg.Load1) > threshold
	if c.Status {
		c.Reason = fmt.Sprintf(
			"load %.2f, threshold %.2f for %d CPUs",
			loadAvg.Load1, threshold, cpuCount,
		)
	}
	c.LastTransitionTime = transitionTime(c.Type, c.Status, prev)
	return c
}

func evaluateDiskPressure(
	disks []disk.UsageStats,
	threshold int,
	prev []job.Condition,
) job.Condition {
	c := job.Condition{Type: job.ConditionDiskPressure}
	if len(disks) == 0 {
		c.LastTransitionTime = transitionTime(c.Type, false, prev)
		return c
	}
	for _, d := range disks {
		if d.Total == 0 {
			continue
		}
		pct := float64(d.Used) / float64(d.Total) * 100
		if pct > float64(threshold) {
			c.Status = true
			c.Reason = fmt.Sprintf(
				"%s %.0f%% used (%.1f/%.1f GB)",
				d.Name, pct,
				float64(d.Used)/1024/1024/1024,
				float64(d.Total)/1024/1024/1024,
			)
			break
		}
	}
	c.LastTransitionTime = transitionTime(c.Type, c.Status, prev)
	return c
}
