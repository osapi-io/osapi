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

// Package instruments is each metric that osapi records: its name, kind,
// unit, labels, the component that records it, and its doc. The code that
// makes an instrument takes its name and doc from here, and the docs site's
// metrics reference is written from All. Thus the two cannot differ.
package instruments

import (
	"fmt"
	"strings"
)

// Namespace is what the Prometheus exporter puts before each name, with an
// underscore. A Name does not have it.
const Namespace = "osapi"

// The kinds of an instrument, as Prometheus shows them.
const (
	KindCounter   = "counter"
	KindGauge     = "gauge"
	KindHistogram = "histogram"
)

// Instrument is one metric.
type Instrument struct {
	// Name is the instrument's name, without the namespace, such as
	// jobs_created_total.
	Name string
	// Kind is counter, gauge or histogram.
	Kind string
	// Unit is the unit, such as s, or "".
	Unit string
	// Labels are the labels of each value, as Prometheus names them.
	Labels []string
	// Component is what records it: controller, agent, or all.
	Component string
	// Doc says what it counts or measures.
	Doc string
}

// Exported is the name that Prometheus shows: the namespace, then Name.
func (i Instrument) Exported() string {
	return Namespace + "_" + i.Name
}

// The instruments, in the order the reference lists them.
var (
	ComponentUp = Instrument{
		Name: "component_up", Kind: KindGauge, Component: "all",
		Doc: "Whether the component is ready (1) or not (0).",
	}
	SubsystemUp = Instrument{
		Name: "subsystem_up", Kind: KindGauge, Component: "all",
		Labels: []string{"subsystem"},
		Doc:    "Whether the config turns a subsystem on (1) or off (0).",
	}
	HeartbeatAge = Instrument{
		Name: "heartbeat_age_seconds", Kind: KindGauge, Unit: "s", Component: "agent",
		Doc: "Seconds since last successful heartbeat write.",
	}
	JobStoreBytes = Instrument{
		Name: "job_store_bytes", Kind: KindGauge, Unit: "By", Component: "controller",
		Doc: "Bytes that the jobs bucket holds.",
	}
	JobStoreMaxBytes = Instrument{
		Name:      "job_store_max_bytes",
		Kind:      KindGauge,
		Unit:      "By",
		Component: "controller",
		Doc:       "The max_bytes of the jobs bucket: at this size the controller removes the oldest finished jobs.",
	}
	JobsCreated = Instrument{
		Name: "jobs_created_total", Kind: KindCounter, Component: "controller",
		Doc: "Total jobs submitted via API.",
	}
	JobsProcessed = Instrument{
		Name: "jobs_processed_total", Kind: KindCounter, Component: "agent",
		Labels: []string{"status"},
		Doc:    "Total jobs processed.",
	}
	JobsActive = Instrument{
		Name: "jobs_active", Kind: KindGauge, Component: "agent",
		Doc: "Currently executing jobs.",
	}
	JobDuration = Instrument{
		Name: "job_duration_seconds", Kind: KindHistogram, Unit: "s", Component: "agent",
		Doc: "Job execution duration in seconds.",
	}
	BehaviorCalls = Instrument{
		Name: "behavior_calls_total", Kind: KindCounter, Component: "agent",
		Labels: []string{
			"osapi_behavior_name", "osapi_behavior_method", "osapi_behavior_verb",
			"osapi_behavior_outcome",
		},
		Doc: "Behavior method calls, by behavior, method and outcome.",
	}
	BehaviorDuration = Instrument{
		Name: "behavior_duration_seconds", Kind: KindHistogram, Unit: "s", Component: "agent",
		Labels: []string{"osapi_behavior_name", "osapi_behavior_method", "osapi_behavior_verb"},
		Doc:    "Behavior method duration in seconds.",
	}
	BehaviorChanges = Instrument{
		Name: "behavior_changes_total", Kind: KindCounter, Component: "agent",
		Labels: []string{"osapi_behavior_name", "osapi_behavior_method", "osapi_behavior_verb"},
		Doc:    "Behavior writes that changed the host.",
	}
)

// All are the instruments that osapi records.
func All() []Instrument {
	return []Instrument{
		ComponentUp, SubsystemUp, HeartbeatAge,
		JobStoreBytes, JobStoreMaxBytes,
		JobsCreated, JobsProcessed, JobsActive, JobDuration,
		BehaviorCalls, BehaviorDuration, BehaviorChanges,
	}
}

// Table is each instrument as a row of a markdown table, for the metrics
// reference of the docs site: the name that Prometheus shows, the kind, the
// labels, the component and the doc.
func Table() string {
	var sb strings.Builder
	sb.WriteString("| Metric | Type | Labels | Component | Description |\n")
	sb.WriteString("| --- | --- | --- | --- | --- |\n")
	for _, i := range All() {
		fmt.Fprintf(&sb, "| `%s` | %s | %s | %s | %s |\n",
			i.Exported(), i.Kind, strings.Join(i.Labels, ", "), i.Component, i.Doc)
	}

	return strings.TrimSuffix(sb.String(), "\n")
}
