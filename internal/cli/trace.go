// Copyright (c) 2024 John Dewey

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

package cli

import (
	"slices"
	"sort"
	"time"

	"github.com/osapi-io/osapi/pkg/sdk/client"
)

// traceDetail are the attributes a span's DETAIL cell shows, the first one it
// has: what the span acted on.
var traceDetail = []string{
	"file.path",
	"nats.kv.key",
	"nats.object.name",
	"messaging.destination.name",
	"db.namespace",
	"http.response.status_code",
}

// TraceRows returns the rows of a trace's span tree, each with its SPAN,
// HOSTNAME, DURATION, STATUS and DETAIL. A child is under its parent, in
// start order, behind tree lines. A span whose parent is not in the trace,
// such as one under the span of a CLI that sent none, is a root.
//
// HOSTNAME is the agent of the nearest job.process at or above the span, so
// each step on an agent names its host. STATUS is the job's status on a
// job.process, and ok or error on other spans. DETAIL is the error of a
// failed span, or else what it acted on.
func TraceRows(
	spans []client.TraceSpan,
) [][]string {
	ids := make(map[string]bool, len(spans))
	for _, s := range spans {
		ids[s.ID] = true
	}
	children := map[string][]client.TraceSpan{}
	var roots []client.TraceSpan
	for _, s := range spans {
		if s.ParentID == "" || !ids[s.ParentID] {
			roots = append(roots, s)
			continue
		}
		children[s.ParentID] = append(children[s.ParentID], s)
	}
	byStart := func(list []client.TraceSpan) {
		sort.SliceStable(list, func(a, b int) bool { return list[a].Start.Before(list[b].Start) })
	}
	byStart(roots)

	var rows [][]string
	var walk func(s client.TraceSpan, lead, branch, host string)
	walk = func(s client.TraceSpan, lead, branch, host string) {
		if h := s.Attributes["osapi.agent.hostname"]; s.Name == "job.process" && h != "" {
			host = h
		}
		rows = append(rows, []string{
			lead + branch + s.Name,
			host,
			FormatDuration(s.Duration.Round(time.Microsecond).String()),
			spanStatus(s),
			spanDetail(s),
		})

		kids := children[s.ID]
		byStart(kids)
		next := lead
		switch branch {
		case "├─ ":
			next += "│  "
		case "└─ ":
			next += "   "
		}
		for i, k := range kids {
			b := "├─ "
			if i == len(kids)-1 {
				b = "└─ "
			}
			walk(k, next, b, host)
		}
	}
	for _, r := range roots {
		walk(r, "", "", "")
	}

	return rows
}

// spanStatus is the job's status on a job.process, and ok or error on other
// spans.
func spanStatus(
	s client.TraceSpan,
) string {
	if st := s.Attributes["job.status"]; s.Name == "job.process" && st != "" {
		return st
	}
	if s.Failed {
		return "error"
	}

	return "ok"
}

// spanDetail is the error of a failed span, or else what it acted on.
func spanDetail(
	s client.TraceSpan,
) string {
	if s.Failed && s.StatusMessage != "" {
		return s.StatusMessage
	}
	for _, k := range traceDetail {
		if v := s.Attributes[k]; v != "" {
			return v
		}
	}

	return ""
}

// TraceServices are the services of the spans, in the order each first
// appears.
func TraceServices(
	spans []client.TraceSpan,
) []string {
	var out []string
	for _, s := range spans {
		if s.Service != "" && !slices.Contains(out, s.Service) {
			out = append(out, s.Service)
		}
	}

	return out
}
