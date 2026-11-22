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
// Package telemetry is the endpoint group for what the controller reads
// from the telemetry backends, such as the spans of a trace.
package telemetry

import "time"

// ExportRows is the rows of the tree of spans: the tree lines and the name,
// what ran it, its duration, its status and its detail.
func ExportRows(
	spans []TraceSpan,
	all bool,
) [][]string {
	var rows [][]string
	var walk func(nodes []TraceNode, lead string, root bool)
	walk = func(nodes []TraceNode, lead string, root bool) {
		for i, n := range nodes {
			last := i == len(nodes)-1
			mark, under := "", ""
			if !root {
				mark, under = "├─ ", "│  "
				if last {
					mark, under = "└─ ", "   "
				}
			}
			rows = append(
				rows,
				[]string{lead + mark + n.Name, n.On, n.Duration, n.Status, n.Detail},
			)
			walk(n.Children, lead+under, false)
		}
	}
	walk(tree("t", spans, all).Spans, "", true)

	return rows
}

// ExportTree exposes tree for testing.
var ExportTree = tree

// ExportIsNATSSpan exposes isNATSSpan for testing.
var ExportIsNATSSpan = isNATSSpan

// ExportServices is the services of the trace of spans.
func ExportServices(
	spans []TraceSpan,
) []string {
	return tree("t", spans, false).Services
}

// ExportNATSCalls is the count of the NATS calls of spans, left out, and the
// time they took.
func ExportNATSCalls(
	spans []TraceSpan,
) (int, time.Duration) {
	t := tree("t", spans, false)
	d, _ := time.ParseDuration(t.NATSTime)

	return t.NATSCalls, d
}
