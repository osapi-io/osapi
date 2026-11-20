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

package telemetry_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/controller/endpoints/telemetry"
)

type TracePublicTestSuite struct {
	suite.Suite
}

// traceSpans is a request with a job on one agent, a failed command under
// it, and a span whose parent is not in the trace.
func traceSpans() []telemetry.TraceSpan {
	at := func(ms int) time.Time {
		return time.Date(2026, 11, 15, 12, 0, 0, 0, time.UTC).
			Add(time.Duration(ms) * time.Millisecond)
	}

	return []telemetry.TraceSpan{
		{
			SpanID:       "c",
			ParentSpanID: "b",
			Name:         "job.process",
			Service:      "osapi-agent",
			StartTime:    at(2),
			DurationMs:   20,
			Attributes: map[string]string{
				"osapi.agent.hostname": "web-01",
				"job.status":           "failed",
			},
		},
		{
			SpanID:     "a",
			Name:       "GET /api/x",
			Service:    "osapi-controller",
			StartTime:  at(0),
			DurationMs: 50,
			Attributes: map[string]string{
				"http.response.status_code": "200", "http.response.body.size": "1536",
			},
		},
		{
			SpanID:       "b",
			ParentSpanID: "a",
			Name:         "job.publish",
			Service:      "osapi-controller",
			StartTime:    at(1),
			DurationMs:   30,
			Attributes:   map[string]string{"messaging.destination.name": "jobs.query.web-01"},
		},
		{
			SpanID:        "d",
			ParentSpanID:  "c",
			Name:          "exec apt-get",
			Service:       "osapi-agent",
			StartTime:     at(3),
			DurationMs:    10,
			Status:        "error",
			StatusMessage: "exit status 100",
			Attributes:    map[string]string{},
		},
		{
			SpanID:       "e",
			ParentSpanID: "c",
			Name:         "file.write",
			Service:      "osapi-agent",
			StartTime:    at(4),
			DurationMs:   1,
			Attributes:   map[string]string{"file.path": "/etc/x.conf"},
		},
		{
			SpanID:       "f",
			ParentSpanID: "a",
			Name:         "job.collect",
			Service:      "osapi-controller",
			StartTime:    at(31),
			DurationMs:   19,
			Attributes:   map[string]string{},
		},
		{
			SpanID:       "h",
			ParentSpanID: "a",
			Name:         "response.encode",
			Service:      "osapi-controller",
			StartTime:    at(50),
			DurationMs:   1,
			Attributes:   map[string]string{},
		},
		{
			SpanID:       "i",
			ParentSpanID: "a",
			Name:         "response.write",
			Service:      "osapi-controller",
			StartTime:    at(51),
			DurationMs:   2,
			Attributes:   map[string]string{"http.response.body.size": "1536"},
		},
		{
			SpanID:       "g",
			ParentSpanID: "missing",
			Name:         "orphan",
			Service:      "osapi-cli",
			StartTime:    at(60),
			DurationMs:   1,
			Status:       "error",
			Attributes:   map[string]string{},
		},
	}
}

func (s *TracePublicTestSuite) TestTraceRows() {
	s.Equal([][]string{
		{"GET /api/x", "controller", "50ms", "ok", "status=200"},
		{"├─ job.publish", "controller", "30ms", "ok", "subject=jobs.query.web-01"},
		{"│  └─ job.process", "agent", "20ms", "failed", "host=web-01"},
		{"│     ├─ exec apt-get", "agent", "10ms", "error", `error="exit status 100"`},
		{"│     └─ file.write", "agent", "1ms", "ok", "path=/etc/x.conf"},
		{"├─ job.collect", "controller", "19ms", "ok", ""},
		{"├─ response.encode", "controller", "1ms", "ok", ""},
		{"└─ response.write", "controller", "2ms", "ok", "size=1536"},
		{"orphan", "cli", "1ms", "error", ""},
	}, telemetry.ExportRows(traceSpans(), true))
}

// natsTrace is a job on one agent through the queue, with NATS calls of each
// key shape, a behavior write that requires another, and a command.
func natsTrace() []telemetry.TraceSpan {
	at := func(ms int) time.Time {
		return time.Date(2026, 11, 15, 12, 0, 0, 0, time.UTC).
			Add(time.Duration(ms) * time.Millisecond)
	}
	span := func(id, parent, name string, ms int, attrs map[string]string) telemetry.TraceSpan {
		return telemetry.TraceSpan{
			SpanID: id, ParentSpanID: parent, Name: name, StartTime: at(ms),
			DurationMs: 1, Attributes: attrs,
		}
	}
	kv := func(key string) map[string]string { return map[string]string{"nats.kv.key": key} }

	return []telemetry.TraceSpan{
		span("root", "", "POST /api/x", 0, map[string]string{"http.response.status_code": "200"}),
		span("pub", "root", "job.publish", 1, map[string]string{}),
		span("put", "pub", "kv.put", 2, kv("jobs.j1")),
		span("reg", "pub", "kv.get", 3, kv("agents.m1")),
		span("fct", "pub", "kv.get", 4, kv("facts.m1")),
		span("drn", "pub", "kv.get", 5, kv("drain.m1")),
		span("np", "pub", "nats.publish", 6, map[string]string{}),
		span("rcv", "np", "nats.process", 7, map[string]string{}),
		span(
			"job",
			"rcv",
			"job.process",
			8,
			map[string]string{"osapi.agent.hostname": "web-01", "job.status": "completed"},
		),
		span("keys", "job", "kv.keys", 9, map[string]string{"db.namespace": "osapi-job-queue"}),
		span("st", "job", "kv.put", 10, kv("status.j1.started.web-01.123")),
		span("tl", "job", "kv.put", 11, kv("timeline.web-01.drain.123")),
		span("beh", "job", "behavior.sysctl.create", 12, map[string]string{
			"osapi.behavior.changed": "true", "osapi.behavior.input.key": "vm.swappiness",
			"osapi.behavior.input.value": "10",
		}),
		span(
			"mf",
			"beh",
			"behavior.managedfile.get",
			13,
			map[string]string{"osapi.behavior.changed": "false"},
		),
		span(
			"rec",
			"mf",
			"kv.put",
			14,
			map[string]string{"nats.kv.key": "web-01.abc", "db.namespace": "osapi-file-state"},
		),
		span("odd", "job", "kv.get", 15, kv("other.key")),
		span("ans", "job", "kv.put", 16, kv("responses.j1.web-01.123")),
		span("col", "root", "job.collect", 17, map[string]string{}),
		span("w", "col", "kv.watch", 18, kv("responses.j1.>")),
	}
}

func (s *TracePublicTestSuite) TestJobPublishDetail() {
	publish := func(attrs map[string]string) string {
		rows := telemetry.ExportRows([]telemetry.TraceSpan{{
			SpanID: "p", Name: "job.publish", StartTime: time.Now(), Attributes: attrs,
		}}, true)

		return rows[0][4]
	}

	tests := []struct {
		name  string
		attrs map[string]string
		want  string
	}{
		{
			name: "when the job has its operation and target they say where it went",
			attrs: map[string]string{
				"job.type":                   "behavior.sysctl.get",
				"osapi.target":               "debian-1",
				"messaging.destination.name": "osapi.jobs.query.host.d530664edc1c431eb5ef1e9dfd6617a5",
			},
			want: "op=behavior.sysctl.get target=debian-1",
		},
		{
			name: "when an older trace has only the subject it shows the subject",
			attrs: map[string]string{
				"messaging.destination.name": "osapi.jobs.query._all",
			},
			want: "subject=osapi.jobs.query._all",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, publish(tc.attrs))
		})
	}
}

// TestJobReadDetail shows what a job client read acted on: the jobs a
// job.list page has, and the job of a job.get.
func (s *TracePublicTestSuite) TestJobReadDetail() {
	detail := func(name string, attrs map[string]string) string {
		rows := telemetry.ExportRows([]telemetry.TraceSpan{{
			SpanID: "r", Name: name, StartTime: time.Now(), Attributes: attrs,
		}}, true)

		return rows[0][4]
	}

	tests := []struct {
		name  string
		span  string
		attrs map[string]string
		want  string
	}{
		{
			name:  "when job.list has two jobs it says so",
			span:  "job.list",
			attrs: map[string]string{"job.count": "2", "job.status_filter": ""},
			want:  "jobs=2",
		},
		{
			name:  "when job.list has one job it counts it",
			span:  "job.list",
			attrs: map[string]string{"job.count": "1"},
			want:  "jobs=1",
		},
		{
			name:  "when agent.list counts the agents it says so",
			span:  "agent.list",
			attrs: map[string]string{"osapi.agent.count": "3"},
			want:  "agents=3",
		},
		{
			name:  "when audit.list counts one entry it counts it",
			span:  "audit.list",
			attrs: map[string]string{"osapi.audit.count": "1"},
			want:  "entries=1",
		},
		{
			name:  "when admission.admit has a machine ID it shows it",
			span:  "admission.admit",
			attrs: map[string]string{"osapi.agent.machine_id": "m-1"},
			want:  "machine=m-1",
		},
		{
			name:  "when validate.target has a target it shows it",
			span:  "validate.target",
			attrs: map[string]string{"osapi.target": "_all"},
			want:  "target=_all",
		},
		{
			name:  "when job.process has a job ID it does not show it",
			span:  "job.process",
			attrs: map[string]string{"job.id": "0162195c"},
			want:  "",
		},
		{
			name:  "when job.get has a job ID it shows it",
			span:  "job.get",
			attrs: map[string]string{"job.id": "0162195c"},
			want:  "job=0162195c",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, detail(tc.span, tc.attrs))
		})
	}
}

func (s *TracePublicTestSuite) TestTraceRowsWithoutNATS() {
	s.Equal([][]string{
		{"POST /api/x", "", "1ms", "ok", "status=200"},
		{"├─ job.publish", "", "1ms", "ok", ""},
		{"│  └─ job.process", "", "1ms", "completed", "host=web-01"},
		{
			"│     └─ behavior.sysctl.create",
			"",
			"1ms",
			"changed",
			"key=vm.swappiness value=10",
		},
		{"│        └─ behavior.managedfile.get", "", "1ms", "unchanged", ""},
		{"└─ job.collect", "", "1ms", "ok", ""},
	}, telemetry.ExportRows(natsTrace(), false), "a span under a NATS span moves up to the span above it")

	n, took := telemetry.ExportNATSCalls(natsTrace())
	s.Equal(13, n)
	s.Equal(13*time.Millisecond, took)
}

// TestSpanOn shows what ran each span, the role of its service, and the
// host of an agent's job as host= in its DETAIL, so each column says one
// kind of thing.
func (s *TracePublicTestSuite) TestSpanOn() {
	at := time.Now()
	rows := telemetry.ExportRows([]telemetry.TraceSpan{
		{
			SpanID:     "a",
			Name:       "POST /api/x",
			Service:    "osapi-controller",
			StartTime:  at,
			Attributes: map[string]string{},
		},
		{
			SpanID:       "b",
			ParentSpanID: "a",
			Name:         "nats.publish",
			Service:      "osapi-controller",
			StartTime:    at.Add(1),
			Attributes:   map[string]string{},
		},
		{
			SpanID:       "c",
			ParentSpanID: "b",
			Name:         "nats.process",
			Service:      "osapi-agent",
			StartTime:    at.Add(2),
			Attributes:   map[string]string{},
		},
		{
			SpanID:       "d",
			ParentSpanID: "c",
			Name:         "job.process",
			Service:      "osapi-agent",
			StartTime:    at.Add(3),
			Attributes:   map[string]string{"osapi.agent.hostname": "web-01"},
		},
		{
			SpanID:     "e",
			Name:       "osapi client node hostname get",
			Service:    "osapi-cli",
			StartTime:  at.Add(4),
			Attributes: map[string]string{},
		},
		{
			SpanID:     "f",
			Name:       "other",
			Service:    "jaeger",
			StartTime:  at.Add(5),
			Attributes: map[string]string{},
		},
	}, true)

	on, detail := map[string]string{}, map[string]string{}
	for _, r := range rows {
		name := strings.TrimLeft(r[0], "├└│─ ")
		on[name], detail[name] = r[1], r[4]
	}
	s.Equal("controller", on["POST /api/x"])
	s.Equal("controller", on["nats.publish"])
	s.Equal("agent", on["nats.process"])
	s.Equal("agent", on["job.process"])
	s.Equal("host=web-01", detail["job.process"], "the host is a key=value in DETAIL")
	s.Equal("cli", on["osapi client node hostname get"])
	s.Equal("jaeger", on["other"], "a service that is not osapi's keeps its name")
}

func (s *TracePublicTestSuite) TestKeyMeanings() {
	details := map[string]string{}
	for _, r := range telemetry.ExportRows(natsTrace(), true) {
		details[strings.TrimLeft(r[0], "│├└─ ")+" "+r[4]] = r[4]
	}
	for _, want := range []string{
		"kv.put key=job", "kv.get key=registration", "kv.get key=facts", `kv.get key="drain flag"`,
		"kv.keys bucket=osapi-job-queue", `kv.put key="status: started"`, `kv.put key="timeline: drain"`,
		`kv.put key="file record"`, "kv.get key=other.key", "kv.put key=response", "kv.watch key=answers",
	} {
		s.Contains(details, want)
	}
}

func (s *TracePublicTestSuite) TestIsNATSSpan() {
	for _, name := range []string{"kv.get", "objectstore.put", "stream.read", "nats.publish"} {
		s.True(telemetry.ExportIsNATSSpan(name), name)
	}
	for _, name := range []string{"job.process", "file.write", "exec kv"} {
		s.False(telemetry.ExportIsNATSSpan(name), name)
	}
}

func (s *TracePublicTestSuite) TestTraceServices() {
	s.Equal(
		[]string{"osapi-agent", "osapi-controller", "osapi-cli"},
		telemetry.ExportServices(append(traceSpans(), telemetry.TraceSpan{Name: "no service"})),
	)
}

func TestTracePublicTestSuite(t *testing.T) {
	suite.Run(t, new(TracePublicTestSuite))
}
