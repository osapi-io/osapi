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
	"strconv"
	"strings"
	"time"

	"github.com/osapi-io/osapi/pkg/sdk/client"
)

// traceDetail are the attributes a span's DETAIL cell shows, the first one it
// has: what the span acted on. A KV key is shown by what it means.
var traceDetail = []string{
	"file.path",
	"nats.object.name",
	"messaging.destination.name",
	"db.namespace",
	"job.id",
	"http.response.status_code",
}

// TraceRows returns the rows of a trace's span tree, each with its SPAN,
// HOSTNAME, DURATION, STATUS and DETAIL. Without all, it leaves out the
// spans of NATS calls, the bookkeeping of the queue, and NATSCalls counts
// them. A child is under its parent, in
// start order, behind tree lines. A span whose parent is not in the trace,
// such as one under the span of a CLI that sent none, is a root.
//
// HOSTNAME is the agent of the nearest job.process at or above the span, so
// each step on an agent names its host. STATUS is the job's status on a
// job.process, changed or unchanged on a behavior write, and ok or error on
// other spans. DETAIL is the error of a failed span, or else what it acted
// on: a behavior's recorded inputs, what a key is for, a file or a subject.
func TraceRows(
	spans []client.TraceSpan,
	all bool,
) [][]string {
	byID := make(map[string]client.TraceSpan, len(spans))
	for _, s := range spans {
		byID[s.ID] = s
	}
	// parentOf is the nearest ancestor that is shown: a span under a hidden
	// NATS span, such as job.process under nats.process, moves up to the
	// span above it.
	parentOf := func(s client.TraceSpan) string {
		p := s.ParentID
		for {
			ps, ok := byID[p]
			if !ok || all || !IsNATSSpan(ps.Name) {
				if ok {
					return p
				}

				return ""
			}
			p = ps.ParentID
		}
	}
	children := map[string][]client.TraceSpan{}
	var roots []client.TraceSpan
	for _, s := range spans {
		if !all && IsNATSSpan(s.Name) {
			continue
		}
		p := parentOf(s)
		if p == "" {
			roots = append(roots, s)
			continue
		}
		children[p] = append(children[p], s)
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
	// A behavior write says whether it changed the host.
	switch s.Attributes["osapi.behavior.changed"] {
	case "true":
		return "changed"
	case "false":
		return "unchanged"
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
	if inputs := behaviorInputs(s.Attributes); inputs != "" {
		return inputs
	}
	if key := s.Attributes["nats.kv.key"]; key != "" {
		return keyMeaning(key, s.Attributes["db.namespace"])
	}
	// job.list records how many jobs its page has.
	if n, err := strconv.Atoi(s.Attributes["job.count"]); err == nil {
		return Plural(n, "job", "jobs")
	}
	if s.Name == "kv.keys" {
		return "list keys in " + s.Attributes["db.namespace"]
	}
	// The request span has the size too, from otelecho; it shows its status.
	if size, err := strconv.Atoi(s.Attributes["http.response.body.size"]); err == nil &&
		s.Name == "response.write" {
		return FormatBytes(size)
	}
	// A job's subject names a host by its machine ID; the operation and the
	// target that the caller gave say the same thing in words.
	if op, target := s.Attributes["job.type"], s.Attributes["osapi.target"]; op != "" &&
		target != "" {
		return op + " to " + target
	}
	for _, k := range traceDetail {
		if v := s.Attributes[k]; v != "" {
			return v
		}
	}

	return ""
}

// behaviorInputs are the inputs a behavior recorded on its span, as
// key=vm.swappiness, in key order.
func behaviorInputs(
	attrs map[string]string,
) string {
	const prefix = "osapi.behavior.input."
	var parts []string
	for k, v := range attrs {
		if name, ok := strings.CutPrefix(k, prefix); ok {
			parts = append(parts, name+"="+v)
		}
	}
	sort.Strings(parts)

	return strings.Join(parts, " ")
}

// keyMeaning says what an osapi key in bucket is for, from the shape of the
// key: status: started for status.<job>.started.<host>.<time>. A key of no
// known shape is itself.
func keyMeaning(
	key string,
	bucket string,
) string {
	parts := strings.Split(key, ".")
	switch {
	case parts[0] == "jobs" && len(parts) == 2:
		return "job"
	case parts[0] == "status" && len(parts) >= 3:
		return "status: " + parts[2]
	case parts[0] == "responses" && strings.HasSuffix(key, ".>"):
		return "answers"
	case parts[0] == "responses":
		return "response"
	case parts[0] == "agents":
		return "registration"
	case parts[0] == "facts":
		return "facts"
	case parts[0] == "drain":
		return "drain flag"
	case parts[0] == "timeline" && len(parts) >= 3:
		return "timeline: " + parts[2]
	case strings.HasSuffix(bucket, "file-state"):
		return "file record"
	default:
		return key
	}
}

// IsNATSSpan reports whether name is the span of a NATS call, which
// nats-client makes: kv.*, objectstore.*, stream.* and nats.*.
func IsNATSSpan(
	name string,
) bool {
	for _, p := range []string{"kv.", "objectstore.", "stream.", "nats."} {
		if strings.HasPrefix(name, p) {
			return true
		}
	}

	return false
}

// NATSCalls counts the spans of NATS calls, and the time they took.
func NATSCalls(
	spans []client.TraceSpan,
) (int, time.Duration) {
	var n int
	var d time.Duration
	for _, s := range spans {
		if IsNATSSpan(s.Name) {
			n++
			d += s.Duration
		}
	}

	return n, d
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
