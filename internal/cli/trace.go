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
	"http.response.status_code",
}

// opDetail is the attribute that a store or client operation's DETAIL shows,
// by span name: what it acted on. Its first attribute that is set wins.
var opDetail = map[string][]string{
	"job.get":           {"job.id"},
	"job.retry":         {"job.id"},
	"job.delete":        {"job.id"},
	"agent.get":         {"osapi.target"},
	"agent.timeline":    {"osapi.agent.hostname"},
	"enrollment.accept": {"osapi.agent.hostname", "osapi.agent.fingerprint"},
	"enrollment.reject": {"osapi.agent.hostname"},
	"enrollment.remove": {"osapi.agent.hostname"},
	"audit.get":         {"osapi.audit.id"},
	"traces.read":       {"osapi.trace.id"},
	"validate.target":   {"osapi.target"},
	"target.resolve":    {"osapi.target"},
}

// opCounts are the attributes that count what an operation read, and the
// key of the count in DETAIL: job.list's job.count is jobs=2.
var opCounts = []struct{ attr, key string }{
	{"job.count", "jobs"},
	{"osapi.agent.count", "agents"},
	{"osapi.audit.count", "entries"},
	{"osapi.trace.spans", "spans"},
}

// TraceRows returns the rows of a trace's span tree, each with its SPAN,
// ON, DURATION, STATUS and DETAIL. Without all, it leaves out the
// spans of NATS calls, the bookkeeping of the queue, and NATSCalls counts
// them. A child is under its parent, in
// start order, behind tree lines. A span whose parent is not in the trace,
// such as one under the span of a CLI that sent none, is a root.
//
// ON is what ran the span, the role of its service: controller, agent or
// cli. STATUS is the job's status on a job.process, changed or unchanged on
// a behavior write, and ok or error on other spans. DETAIL is what the span
// acted on as key=value pairs, led by error= on a failed span: host= on a
// job.process, a behavior's recorded inputs, a key, a file or a subject.
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
	var walk func(s client.TraceSpan, lead, branch string)
	walk = func(s client.TraceSpan, lead, branch string) {
		rows = append(rows, []string{
			lead + branch + s.Name,
			spanOn(s),
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
			walk(k, next, b)
		}
	}
	for _, r := range roots {
		walk(r, "", "")
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

// spanOn is what ran a span: the role of the service that made it, such as
// controller for osapi-controller, agent or cli. The host an agent span ran
// on is host= in its job.process's DETAIL, not here, so each column says one
// kind of thing.
func spanOn(
	s client.TraceSpan,
) string {
	return strings.TrimPrefix(s.Service, "osapi-")
}

// spanDetail is what a span acted on, as key=value pairs, so each value
// says what it is: target=_all, agents=3, host=debian-1. A failed span
// leads with error=. A value with a space is quoted, as in logfmt.
func spanDetail(
	s client.TraceSpan,
) string {
	var pairs [][2]string
	if s.Failed && s.StatusMessage != "" {
		pairs = append(pairs, [2]string{"error", s.StatusMessage})
	}
	pairs = append(pairs, actedOn(s)...)

	parts := make([]string, 0, len(pairs))
	for _, p := range pairs {
		parts = append(parts, p[0]+"="+logfmtValue(p[1]))
	}

	return strings.Join(parts, " ")
}

// actedOn is what a span acted on, as key and value, by the first rule that
// fits it.
func actedOn(
	s client.TraceSpan,
) [][2]string {
	a := s.Attributes
	if inputs := behaviorInputs(a); len(inputs) > 0 {
		return inputs
	}
	if s.Name == "job.process" {
		if h := a["osapi.agent.hostname"]; h != "" {
			return [][2]string{{"host", h}}
		}
	}
	if key := a["nats.kv.key"]; key != "" {
		return [][2]string{{"key", keyMeaning(key, a["db.namespace"])}}
	}
	for _, c := range opCounts {
		if _, err := strconv.Atoi(a[c.attr]); err == nil {
			return [][2]string{{c.key, a[c.attr]}}
		}
	}
	for _, k := range opDetail[s.Name] {
		if v := a[k]; v != "" {
			return [][2]string{{attrKeys[k], v}}
		}
	}
	if s.Name == "kv.keys" {
		return [][2]string{{"bucket", a["db.namespace"]}}
	}
	// The request span has the size too, from otelecho; it shows its status.
	if size, err := strconv.Atoi(a["http.response.body.size"]); err == nil &&
		s.Name == "response.write" {
		return [][2]string{{"size", FormatBytes(size)}}
	}
	// A job's subject names a host by its machine ID; the operation and the
	// target that the caller gave say the same thing in words.
	if op, target := a["job.type"], a["osapi.target"]; op != "" && target != "" {
		return [][2]string{{"op", op}, {"target", target}}
	}
	for _, k := range traceDetail {
		if v := a[k]; v != "" {
			return [][2]string{{attrKeys[k], v}}
		}
	}

	return nil
}

// attrKeys is the key that each attribute has in DETAIL: the word a reader
// uses for it, not the OpenTelemetry name.
var attrKeys = map[string]string{
	"file.path":                  "path",
	"nats.object.name":           "object",
	"messaging.destination.name": "subject",
	"db.namespace":               "bucket",
	"http.response.status_code":  "status",
	"job.id":                     "job",
	"osapi.target":               "target",
	"osapi.agent.hostname":       "host",
	"osapi.agent.fingerprint":    "fingerprint",
	"osapi.audit.id":             "entry",
	"osapi.trace.id":             "trace",
}

// logfmtValue quotes a value that has a space, an equals sign or a quote,
// as logfmt does, so a reader can tell where it ends.
func logfmtValue(
	v string,
) string {
	if v == "" || strings.ContainsAny(v, " =\"") {
		return strconv.Quote(v)
	}

	return v
}

// behaviorInputs are the inputs a behavior recorded on its span, as key
// and value, in key order: key=vm.swappiness value=10.
func behaviorInputs(
	attrs map[string]string,
) [][2]string {
	const prefix = "osapi.behavior.input."
	var pairs [][2]string
	for k, v := range attrs {
		if name, ok := strings.CutPrefix(k, prefix); ok {
			pairs = append(pairs, [2]string{name, v})
		}
	}
	sort.Slice(pairs, func(a, b int) bool { return pairs[a][0] < pairs[b][0] })

	return pairs
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
