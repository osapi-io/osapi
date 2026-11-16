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

package cli_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/pkg/sdk/client"
)

type TracePublicTestSuite struct {
	suite.Suite
}

// traceSpans is a request with a job on one agent, a failed command under
// it, and a span whose parent is not in the trace.
func traceSpans() []client.TraceSpan {
	at := func(ms int) time.Time {
		return time.Date(2026, 11, 15, 12, 0, 0, 0, time.UTC).
			Add(time.Duration(ms) * time.Millisecond)
	}

	return []client.TraceSpan{
		{
			ID: "c", ParentID: "b", Name: "job.process", Service: "osapi-agent", Start: at(2), Duration: 20 * time.Millisecond,
			Attributes: map[string]string{"osapi.agent.hostname": "web-01", "job.status": "failed"},
		},
		{
			ID: "a", Name: "GET /api/x", Service: "osapi-controller", Start: at(0), Duration: 50 * time.Millisecond,
			Attributes: map[string]string{
				"http.response.status_code": "200", "http.response.body.size": "1536",
			},
		},
		{
			ID: "b", ParentID: "a", Name: "job.publish", Service: "osapi-controller", Start: at(1), Duration: 30 * time.Millisecond,
			Attributes: map[string]string{"messaging.destination.name": "jobs.query.web-01"},
		},
		{
			ID: "d", ParentID: "c", Name: "exec apt-get", Service: "osapi-agent", Start: at(3), Duration: 10 * time.Millisecond,
			Failed: true, StatusMessage: "exit status 100", Attributes: map[string]string{},
		},
		{
			ID: "e", ParentID: "c", Name: "file.write", Service: "osapi-agent", Start: at(4), Duration: time.Millisecond,
			Attributes: map[string]string{"file.path": "/etc/x.conf"},
		},
		{
			ID: "f", ParentID: "a", Name: "job.collect", Service: "osapi-controller", Start: at(31), Duration: 19 * time.Millisecond,
			Attributes: map[string]string{},
		},
		{
			ID: "h", ParentID: "a", Name: "response.encode", Service: "osapi-controller", Start: at(50), Duration: time.Millisecond,
			Attributes: map[string]string{},
		},
		{
			ID: "i", ParentID: "a", Name: "response.write", Service: "osapi-controller", Start: at(51), Duration: 2 * time.Millisecond,
			Attributes: map[string]string{"http.response.body.size": "1536"},
		},
		{
			ID: "g", ParentID: "missing", Name: "orphan", Service: "osapi-cli", Start: at(60), Duration: time.Millisecond,
			Failed: true, Attributes: map[string]string{},
		},
	}
}

func (s *TracePublicTestSuite) TestTraceRows() {
	s.Equal([][]string{
		{"GET /api/x", "", "50ms", "ok", "200"},
		{"├─ job.publish", "", "30ms", "ok", "jobs.query.web-01"},
		{"│  └─ job.process", "web-01", "20ms", "failed", ""},
		{"│     ├─ exec apt-get", "web-01", "10ms", "error", "exit status 100"},
		{"│     └─ file.write", "web-01", "1ms", "ok", "/etc/x.conf"},
		{"├─ job.collect", "", "19ms", "ok", ""},
		{"├─ response.encode", "", "1ms", "ok", ""},
		{"└─ response.write", "", "2ms", "ok", "1.5 KB"},
		{"orphan", "", "1ms", "error", ""},
	}, cli.TraceRows(traceSpans(), true))
}

// natsTrace is a job on one agent through the queue, with NATS calls of each
// key shape, a behavior write that requires another, and a command.
func natsTrace() []client.TraceSpan {
	at := func(ms int) time.Time {
		return time.Date(2026, 11, 15, 12, 0, 0, 0, time.UTC).
			Add(time.Duration(ms) * time.Millisecond)
	}
	span := func(id, parent, name string, ms int, attrs map[string]string) client.TraceSpan {
		return client.TraceSpan{
			ID: id, ParentID: parent, Name: name, Start: at(ms),
			Duration: time.Millisecond, Attributes: attrs,
		}
	}
	kv := func(key string) map[string]string { return map[string]string{"nats.kv.key": key} }

	return []client.TraceSpan{
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
		rows := cli.TraceRows([]client.TraceSpan{{
			ID: "p", Name: "job.publish", Start: time.Now(), Attributes: attrs,
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
			want: "behavior.sysctl.get to debian-1",
		},
		{
			name: "when an older trace has only the subject it shows the subject",
			attrs: map[string]string{
				"messaging.destination.name": "osapi.jobs.query._all",
			},
			want: "osapi.jobs.query._all",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, publish(tc.attrs))
		})
	}
}

func (s *TracePublicTestSuite) TestTraceRowsWithoutNATS() {
	s.Equal([][]string{
		{"POST /api/x", "", "1ms", "ok", "200"},
		{"├─ job.publish", "", "1ms", "ok", ""},
		{"│  └─ job.process", "web-01", "1ms", "completed", ""},
		{
			"│     └─ behavior.sysctl.create",
			"web-01",
			"1ms",
			"changed",
			"key=vm.swappiness value=10",
		},
		{"│        └─ behavior.managedfile.get", "web-01", "1ms", "unchanged", ""},
		{"└─ job.collect", "", "1ms", "ok", ""},
	}, cli.TraceRows(natsTrace(), false), "a span under a NATS span moves up to the span above it")

	n, took := cli.NATSCalls(natsTrace())
	s.Equal(13, n)
	s.Equal(13*time.Millisecond, took)
}

func (s *TracePublicTestSuite) TestKeyMeanings() {
	details := map[string]string{}
	for _, r := range cli.TraceRows(natsTrace(), true) {
		details[strings.TrimLeft(r[0], "│├└─ ")+" "+r[4]] = r[4]
	}
	for _, want := range []string{
		"kv.put job", "kv.get registration", "kv.get facts", "kv.get drain flag",
		"kv.keys list keys in osapi-job-queue", "kv.put status: started", "kv.put timeline: drain",
		"kv.put file record", "kv.get other.key", "kv.put response", "kv.watch answers",
	} {
		s.Contains(details, want)
	}
}

func (s *TracePublicTestSuite) TestIsNATSSpan() {
	for _, name := range []string{"kv.get", "objectstore.put", "stream.read", "nats.publish"} {
		s.True(cli.IsNATSSpan(name), name)
	}
	for _, name := range []string{"job.process", "file.write", "exec kv"} {
		s.False(cli.IsNATSSpan(name), name)
	}
}

func (s *TracePublicTestSuite) TestTraceServices() {
	s.Equal(
		[]string{"osapi-agent", "osapi-controller", "osapi-cli"},
		cli.TraceServices(append(traceSpans(), client.TraceSpan{Name: "no service"})),
	)
}

func (s *TracePublicTestSuite) TestSpanCell() {
	s.Equal("job.publish", ansi(cli.ExportRenderCell("SPAN", "job.publish")))
	s.Equal("│  └─ job.process", ansi(cli.ExportRenderCell("SPAN", "│  └─ job.process")))
}

func (s *TracePublicTestSuite) TestTreeSpacingIsKept() {
	out := cli.Capture(func() {
		cli.PrintCompactTable([]cli.Section{{
			Title:     "Spans",
			Headers:   []string{"SPAN", "DETAIL"},
			Rows:      [][]string{{"│     └─ file.write", "a  b"}, {"two\nlines", "x"}},
			Reference: true,
		}})
	})

	plain := ansi(out)
	s.Contains(plain, "│     └─ file.write")
	s.Contains(plain, "a b", "other cells are still flattened")
	s.Contains(plain, "two lines")
}

// ansi drops the color codes of s.
func ansi(
	s string,
) string {
	var b strings.Builder
	skip := false
	for _, r := range s {
		switch {
		case r == '\x1b':
			skip = true
		case skip && r == 'm':
			skip = false
		case !skip:
			b.WriteRune(r)
		}
	}

	return b.String()
}

func TestTracePublicTestSuite(t *testing.T) {
	suite.Run(t, new(TracePublicTestSuite))
}
