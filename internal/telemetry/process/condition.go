package process

import (
	"fmt"
	"time"

	"github.com/retr0h/osapi/internal/job"
)

// ConditionThresholds holds thresholds for process-level conditions.
type ConditionThresholds struct {
	// MemoryPressureBytes is the RSS threshold in bytes.
	MemoryPressureBytes int64
	// HighCPUPercent is the CPU usage threshold as a percentage.
	HighCPUPercent float64
}

// EvaluateProcessConditions evaluates process-level conditions against
// thresholds and returns the condition list. Tracks transition times
// using the previous condition state.
func EvaluateProcessConditions(
	metrics *Metrics,
	thresholds ConditionThresholds,
	prev []job.Condition,
) []job.Condition {
	if metrics == nil {
		return nil
	}

	var conditions []job.Condition

	// ProcessMemoryPressure
	if thresholds.MemoryPressureBytes > 0 {
		memStatus := metrics.RSSBytes > thresholds.MemoryPressureBytes
		reason := ""
		if memStatus {
			reason = fmt.Sprintf(
				"process RSS %d bytes exceeds threshold %d bytes",
				metrics.RSSBytes,
				thresholds.MemoryPressureBytes,
			)
		}

		conditions = append(conditions, job.Condition{
			Type:               "ProcessMemoryPressure",
			Status:             memStatus,
			Reason:             reason,
			LastTransitionTime: transitionTime("ProcessMemoryPressure", memStatus, prev),
		})
	}

	// ProcessHighCPU
	if thresholds.HighCPUPercent > 0 {
		cpuStatus := metrics.CPUPercent > thresholds.HighCPUPercent
		reason := ""
		if cpuStatus {
			reason = fmt.Sprintf(
				"process CPU %.1f%% exceeds threshold %.1f%%",
				metrics.CPUPercent,
				thresholds.HighCPUPercent,
			)
		}

		conditions = append(conditions, job.Condition{
			Type:               "ProcessHighCPU",
			Status:             cpuStatus,
			Reason:             reason,
			LastTransitionTime: transitionTime("ProcessHighCPU", cpuStatus, prev),
		})
	}

	return conditions
}

// transitionTime returns the previous LastTransitionTime if the status
// hasn't changed, otherwise returns now.
func transitionTime(
	condType string,
	newStatus bool,
	prev []job.Condition,
) time.Time {
	for i := range prev {
		if prev[i].Type == condType && prev[i].Status == newStatus {
			return prev[i].LastTransitionTime
		}
	}

	return time.Now()
}
