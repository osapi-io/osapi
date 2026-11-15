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
			Attributes: map[string]string{"http.response.status_code": "200"},
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
		{"└─ job.collect", "", "19ms", "ok", ""},
		{"orphan", "", "1ms", "error", ""},
	}, cli.TraceRows(traceSpans()))
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
