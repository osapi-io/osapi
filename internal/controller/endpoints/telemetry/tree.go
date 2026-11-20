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

import (
	"sort"
	"strconv"
	"strings"
	"time"
)

// tree is the answer of a trace from its spans: each span under its
// parent, in start order, with what ran it, its status and what it acted
// on, and the summary of the trace. Without all, it leaves out the spans of
// NATS calls, the bookkeeping of the queue, and counts them: a span under
// one moves up to the span above it. A span whose parent is not in the
// trace, such as one under the span of a CLI that sent none, is a root.
func tree(
	id string,
	spans []TraceSpan,
	all bool,
) TraceResponse {
	out := TraceResponse{TraceID: id, Spans: []TraceNode{}}
	if len(spans) == 0 {
		return out
	}

	byID := make(map[string]TraceSpan, len(spans))
	start, end := spans[0].StartTime, spans[0].StartTime
	var hidden time.Duration
	for _, s := range spans {
		byID[s.SpanID] = s
		if s.StartTime.Before(start) {
			start = s.StartTime
		}
		if e := s.StartTime.Add(duration(s)); e.After(end) {
			end = e
		}
		if s.Status == "error" {
			out.FailedSpans++
		}
		if s.Service != "" && !contains(out.Services, s.Service) {
			out.Services = append(out.Services, s.Service)
		}
		if isNATSSpan(s.Name) {
			out.NATSCalls++
			hidden += duration(s)
		}
	}
	out.Started = start
	out.Duration = end.Sub(start).String()
	if all {
		out.NATSCalls, hidden = 0, 0
	}
	if out.NATSCalls > 0 {
		out.NATSTime = hidden.Round(time.Microsecond).String()
	}

	// parentOf is the nearest ancestor that is shown.
	parentOf := func(s TraceSpan) string {
		p := s.ParentSpanID
		for {
			ps, ok := byID[p]
			if !ok {
				return ""
			}
			if all || !isNATSSpan(ps.Name) {
				return p
			}
			p = ps.ParentSpanID
		}
	}
	children := map[string][]TraceSpan{}
	var roots []TraceSpan
	for _, s := range spans {
		if !all && isNATSSpan(s.Name) {
			continue
		}
		if p := parentOf(s); p != "" {
			children[p] = append(children[p], s)

			continue
		}
		roots = append(roots, s)
	}

	var node func(s TraceSpan) TraceNode
	node = func(s TraceSpan) TraceNode {
		n := TraceNode{
			Name:     s.Name,
			On:       spanOn(s),
			Duration: duration(s).Round(time.Microsecond).String(),
			Status:   spanStatus(s),
			Detail:   spanDetail(s),
			Span:     s,
		}
		kids := children[s.SpanID]
		byStart(kids)
		for _, k := range kids {
			n.Children = append(n.Children, node(k))
		}

		return n
	}
	byStart(roots)
	for _, r := range roots {
		out.Spans = append(out.Spans, node(r))
	}

	return out
}

// duration is how long a span took.
func duration(
	s TraceSpan,
) time.Duration {
	return time.Duration(s.DurationMs * float64(time.Millisecond))
}

// byStart sorts spans by their start, the first first.
func byStart(
	list []TraceSpan,
) {
	sort.SliceStable(list, func(a, b int) bool {
		return list[a].StartTime.Before(list[b].StartTime)
	})
}

// contains reports whether list has s.
func contains(
	list []string,
	s string,
) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}

	return false
}

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
	"job.get":          {"job.id"},
	"job.retry":        {"job.id"},
	"job.delete":       {"job.id"},
	"agent.get":        {"osapi.target"},
	"agent.timeline":   {"osapi.agent.hostname"},
	"admission.get":    {"osapi.agent.machine_id"},
	"admission.admit":  {"osapi.agent.machine_id"},
	"admission.reject": {"osapi.agent.machine_id"},
	"admission.revoke": {"osapi.agent.machine_id"},
	"admission.delete": {"osapi.agent.machine_id"},
	"audit.get":        {"osapi.audit.id"},
	"traces.read":      {"osapi.trace.id"},
	"validate.target":  {"osapi.target"},
	"target.resolve":   {"osapi.target"},
}

// opCounts are the attributes that count what an operation read, and the
// key of the count in DETAIL: job.list's job.count is jobs=2.
var opCounts = []struct{ attr, key string }{
	{"job.count", "jobs"},
	{"osapi.agent.count", "agents"},
	{"osapi.audit.count", "entries"},
	{"osapi.trace.spans", "spans"},
}

// spanStatus is the job's status on a job.process, and ok or error on other
// spans.
func spanStatus(
	s TraceSpan,
) string {
	if st := s.Attributes["job.status"]; s.Name == "job.process" && st != "" {
		return st
	}
	if s.Status == "error" {
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
	s TraceSpan,
) string {
	return strings.TrimPrefix(s.Service, "osapi-")
}

// spanDetail is what a span acted on, as key=value pairs, so each value
// says what it is: target=_all, agents=3, host=debian-1. A failed span
// leads with error=. A value with a space is quoted, as in logfmt.
func spanDetail(
	s TraceSpan,
) string {
	var pairs [][2]string
	if s.Status == "error" && s.StatusMessage != "" {
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
	s TraceSpan,
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
		return [][2]string{{"size", strconv.Itoa(size)}}
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
	"osapi.agent.machine_id":     "machine",
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

// isNATSSpan reports whether name is the span of a NATS call, which
// nats-client makes: kv.*, objectstore.*, stream.* and nats.*.
func isNATSSpan(
	name string,
) bool {
	for _, p := range []string{"kv.", "objectstore.", "stream.", "nats."} {
		if strings.HasPrefix(name, p) {
			return true
		}
	}

	return false
}
