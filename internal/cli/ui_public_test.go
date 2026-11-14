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
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"k8s.io/utils/ptr"

	"github.com/google/uuid"
	"github.com/osapi-io/osapi/pkg/sdk/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli"
)

type UIPublicTestSuite struct {
	suite.Suite
}

func (suite *UIPublicTestSuite) TestSummarize() {
	tests := []struct {
		name string
		v    any
		want string
	}{
		{name: "when there is nothing", v: nil, want: "none"},
		{name: "when it is a list", v: []any{1, 2}, want: "2 items"},
		{name: "when it is an object", v: map[string]any{"a": 1}, want: "1 field"},
		{name: "when it is a whole number", v: float64(3), want: "3"},
		{name: "when it is a fraction", v: 0.25, want: "0.25"},
		{name: "when it is text", v: "ok", want: "ok"},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.Equal(tc.want, cli.Summarize(tc.v))
		})
	}
}

func (suite *UIPublicTestSuite) TestPlural() {
	suite.Equal("1 key", cli.Plural(1, "key", "keys"))
	suite.Equal("2 keys", cli.Plural(2, "key", "keys"))
}

func (suite *UIPublicTestSuite) TestFormatTime() {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "when the time parses it is local",
			in:   "2026-11-13T02:15:13.549Z",
			want: time.Date(2026, 11, 13, 2, 15, 13, 0, time.UTC).
				Local().
				Format("2006-01-02 15:04:05"),
		},
		{name: "when it does not parse it is kept", in: "yesterday", want: "yesterday"},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.Equal(tc.want, cli.FormatTime(tc.in))
		})
	}
}

func (suite *UIPublicTestSuite) TestFormatDuration() {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "when under a millisecond", in: "123.456µs", want: "123µs"},
		{name: "when under a second", in: "6.168625ms", want: "6ms"},
		{name: "when under a minute", in: "1.234567s", want: "1.2s"},
		{name: "when a minute or more", in: "184.4s", want: "3m4s"},
		{name: "when it does not parse it is kept", in: "soon", want: "soon"},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.Equal(tc.want, cli.FormatDuration(tc.in))
		})
	}
}

func (suite *UIPublicTestSuite) TestPrintObject() {
	tests := []struct {
		name string
		v    any
		want string
	}{
		{
			name: "when an object",
			v:    map[string]any{"type": "node.sysctl.get", "data": nil},
			want: "type",
		},
		{name: "when not an object", v: []any{1}, want: "Value"},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			out := cli.Capture(func() { cli.ExportPrintObject("Request", tc.v) })
			suite.Contains(out, tc.want)
			suite.Contains(out, "--json prints all of it.")
		})
	}
}

// update rewrites the golden files from the current renderer:
//
//	go test ./internal/cli -run TestUIPublicTestSuite/TestShapes -update
var update = flag.Bool("update", false, "rewrite the golden output files")

// TestShapes renders every shape of command output the style guide defines
// and compares it, colors included, with a committed file. A change to the
// renderer that alters any shape fails here, and the diff shows how.
func (suite *UIPublicTestSuite) TestShapes() {
	ok := func(host string, fields ...string) cli.ResultRow {
		return cli.ResultRow{Hostname: host, Status: "ok", Fields: fields}
	}
	skipped := func(host string) cli.ResultRow {
		return cli.ResultRow{
			Hostname: host,
			Status:   "skipped",
			Error:    ptr.To("operation not supported on this OS family"),
		}
	}
	failed := func(host string) cli.ResultRow {
		return cli.ResultRow{Hostname: host, Status: "failed", Error: ptr.To("permission denied")}
	}
	results := func(rows []cli.ResultRow, headers ...string) func() {
		return func() {
			cli.PrintJobID("550e8400-e29b-41d4-a716-446655440000")
			tr := cli.BuildBroadcastTable(rows, headers)
			cli.PrintCompactTable(
				[]cli.Section{{Headers: tr.Headers, Rows: tr.Rows, Errors: tr.Errors}},
			)
		}
	}

	tests := []struct {
		name   string
		render func()
	}{
		{
			name:   "one host",
			render: results([]cli.ResultRow{ok("web-01", "vm.swappiness", "10")}, "KEY", "VALUE"),
		},
		{
			name: "several hosts in name order",
			render: results([]cli.ResultRow{
				ok("web-02", "vm.swappiness", "10"),
				ok("web-01", "vm.swappiness", "10"),
				ok("db-01", "vm.swappiness", "1"),
			}, "KEY", "VALUE"),
		},
		{
			name: "a list across hosts grouped by host",
			render: results([]cli.ResultRow{
				ok("web-01", "net.ipv4.ip_forward", "1"),
				ok("web-01", "vm.swappiness", "10"),
				ok("web-02", "net.ipv4.ip_forward", "1"),
				ok("web-02", "vm.swappiness", "10"),
			}, "KEY", "VALUE"),
		},
		{
			name: "mixed ok skipped and failed",
			render: results([]cli.ResultRow{
				ok("web-01", "vm.swappiness", "10"),
				skipped("mac-01"),
				failed("web-02"),
			}, "KEY", "VALUE"),
		},
		{
			name:   "every host skipped",
			render: results([]cli.ResultRow{skipped("mac-01"), skipped("mac-02")}, "KEY", "VALUE"),
		},
		{
			name: "an empty list",
			render: func() {
				cli.PrintCompactTable([]cli.Section{{
					Title:     "Files (0)",
					Headers:   []string{"NAME", "SIZE"},
					Reference: true,
					Empty:     "None. Upload one with osapi client file upload.",
				}})
			},
		},
		{
			name: "a key-value block",
			render: func() {
				cli.PrintHeading("Agent")
				cli.PrintKV("Hostname", "web-01", "Status", "Ready")
				cli.PrintKV("Primary interface", "eth0")
				cli.PrintKV("OS", "debian 13.6")
			},
		},
		{
			name: "a reference table with routes",
			render: func() {
				cli.PrintCompactTable([]cli.Section{{
					Title:     "Methods (2)",
					Headers:   []string{"NAME", "VERB", "ROUTE", "PERMISSION"},
					Reference: true,
					Rows: [][]string{
						{"get", "read", "GET /api/node/{hostname}/sysctl/{key}", "sysctl:read"},
						{"create", "write", "POST /api/node/{hostname}/sysctl", "sysctl:write"},
					},
				}})
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			got := cli.Capture(tt.render)
			path := filepath.Join("testdata", "golden", filepath.Clean(tt.name)+".ansi")

			if *update {
				suite.Require().NoError(os.MkdirAll(filepath.Dir(path), 0o755))
				suite.Require().NoError(os.WriteFile(path, []byte(got), 0o644))

				return
			}

			want, err := os.ReadFile(path)
			suite.Require().NoError(err, "no golden file; run with -update")
			suite.Equal(string(want), got)
		})
	}
}

func TestUIPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(UIPublicTestSuite))
}

func captureStdout(
	fn func(),
) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	_ = w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = old

	return string(out)
}

func (suite *UIPublicTestSuite) TestBoolToSafeString() {
	tests := []struct {
		name         string
		b            *bool
		validateFunc func(string)
	}{
		{
			name: "when true returns true",
			b:    ptr.To(true),
			validateFunc: func(got string) {
				assert.Equal(suite.T(), "true", got)
			},
		},
		{
			name: "when false returns false",
			b:    ptr.To(false),
			validateFunc: func(got string) {
				assert.Equal(suite.T(), "false", got)
			},
		},
		{
			name: "when nil returns empty",
			b:    nil,
			validateFunc: func(got string) {
				assert.Equal(suite.T(), "", got)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			got := cli.BoolToSafeString(tc.b)

			tc.validateFunc(got)
		})
	}
}

func (suite *UIPublicTestSuite) TestBuildBroadcastTable() {
	errMsg := "connection refused"

	tests := []struct {
		name         string
		results      []cli.ResultRow
		fieldHeaders []string
		wantHeaders  []string
		validateFunc func([][]string)
	}{
		{
			name:         "when no results returns hostname and status headers",
			results:      []cli.ResultRow{},
			fieldHeaders: nil,
			wantHeaders:  []string{"HOSTNAME", "STATUS"},
			validateFunc: func(got [][]string) {
				assert.Equal(suite.T(), [][]string{}, got)
			},
		},
		{
			name: "when all results succeed shows ok status",
			results: []cli.ResultRow{
				{Hostname: "web-01", Fields: []string{"val1"}},
				{Hostname: "web-02", Fields: []string{"val2"}},
			},
			fieldHeaders: []string{"DATA"},
			wantHeaders:  []string{"HOSTNAME", "STATUS", "DATA"},
			validateFunc: func(got [][]string) {
				assert.Equal(suite.T(), [][]string{
					{"web-01", "ok", "val1"},
					{"web-02", "ok", "val2"},
				}, got)
			},
		},
		{
			name: "when some results have errors leaves them out of the rows",
			results: []cli.ResultRow{
				{Hostname: "web-01", Status: "ok", Fields: []string{"val1"}},
				{Hostname: "web-02", Status: "failed", Error: &errMsg, Fields: []string{""}},
			},
			fieldHeaders: []string{"DATA"},
			wantHeaders:  []string{"HOSTNAME", "STATUS", "DATA"},
			validateFunc: func(got [][]string) {
				assert.Equal(suite.T(), [][]string{
					{"web-01", "ok", "val1"},
				}, got)
			},
		},
		{
			name: "when a host is skipped leaves it out of the rows",
			results: []cli.ResultRow{
				{Hostname: "web-01", Status: "ok", Fields: []string{"val1"}},
				{Hostname: "web-02", Status: "skipped", Error: &errMsg},
			},
			fieldHeaders: []string{"DATA"},
			wantHeaders:  []string{"HOSTNAME", "STATUS", "DATA"},
			validateFunc: func(got [][]string) {
				assert.Equal(suite.T(), [][]string{
					{"web-01", "ok", "val1"},
				}, got)
			},
		},
		{
			name: "when a skipped host has fields still leaves it out of the rows",
			results: []cli.ResultRow{
				{Hostname: "web-01", Status: "ok", Fields: []string{"val1"}},
				{Hostname: "web-02", Status: "skipped", Fields: []string{"partial"}},
			},
			fieldHeaders: []string{"DATA"},
			wantHeaders:  []string{"HOSTNAME", "STATUS", "DATA"},
			validateFunc: func(got [][]string) {
				assert.Equal(suite.T(), [][]string{
					{"web-01", "ok", "val1"},
				}, got)
			},
		},
		{
			name: "when single host shows hostname and status columns",
			results: []cli.ResultRow{
				{Hostname: "web-01", Fields: []string{"val1"}},
			},
			fieldHeaders: []string{"DATA"},
			wantHeaders:  []string{"HOSTNAME", "STATUS", "DATA"},
			validateFunc: func(got [][]string) {
				assert.Equal(suite.T(), [][]string{
					{"web-01", "ok", "val1"},
				}, got)
			},
		},
		{
			name: "when every host failed has no rows",
			results: []cli.ResultRow{
				{Hostname: "web-01", Status: "failed", Error: &errMsg, Fields: []string{""}},
				{Hostname: "web-02", Status: "failed", Error: &errMsg, Fields: []string{""}},
			},
			fieldHeaders: []string{"DATA"},
			wantHeaders:  []string{"HOSTNAME", "STATUS", "DATA"},
			validateFunc: func(got [][]string) {
				assert.Equal(suite.T(), [][]string{}, got)
			},
		},
		{
			name: "when changed is true shows changed status",
			results: []cli.ResultRow{
				{Hostname: "web-01", Changed: ptr.To(true), Fields: []string{"val1"}},
				{Hostname: "web-02", Changed: ptr.To(false), Fields: []string{"val2"}},
			},
			fieldHeaders: []string{"DATA"},
			wantHeaders:  []string{"HOSTNAME", "STATUS", "DATA"},
			validateFunc: func(got [][]string) {
				assert.Equal(suite.T(), [][]string{
					{"web-01", "changed", "val1"},
					{"web-02", "ok", "val2"},
				}, got)
			},
		},
		{
			name: "when changed and errors both present shows only the changed row",
			results: []cli.ResultRow{
				{
					Hostname: "web-01",
					Status:   "ok",
					Changed:  ptr.To(true),
					Fields:   []string{"val1"},
				},
				{
					Hostname: "web-02",
					Status:   "failed",
					Changed:  ptr.To(false),
					Error:    &errMsg,
					Fields:   []string{""},
				},
			},
			fieldHeaders: []string{"DATA"},
			wantHeaders:  []string{"HOSTNAME", "STATUS", "DATA"},
			validateFunc: func(got [][]string) {
				assert.Equal(suite.T(), [][]string{
					{"web-01", "changed", "val1"},
				}, got)
			},
		},
		{
			name: "when no field headers shows only hostname and status",
			results: []cli.ResultRow{
				{Hostname: "web-01", Changed: ptr.To(true)},
				{Hostname: "web-02"},
			},
			fieldHeaders: nil,
			wantHeaders:  []string{"HOSTNAME", "STATUS"},
			validateFunc: func(got [][]string) {
				assert.Equal(suite.T(), [][]string{
					{"web-01", "changed"},
					{"web-02", "ok"},
				}, got)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tr := cli.BuildBroadcastTable(tc.results, tc.fieldHeaders)

			assert.Equal(suite.T(), tc.wantHeaders, tr.Headers)
			tc.validateFunc(tr.Rows)
		})
	}
}

func (suite *UIPublicTestSuite) TestFormatLabels() {
	tests := []struct {
		name         string
		labels       map[string]string
		validateFunc func(string)
	}{
		{
			name:   "when nil returns empty",
			labels: nil,
			validateFunc: func(got string) {
				assert.Equal(suite.T(), "", got)
			},
		},
		{
			name:   "when empty map returns empty",
			labels: map[string]string{},
			validateFunc: func(got string) {
				assert.Equal(suite.T(), "", got)
			},
		},
		{
			name:   "when single label formats correctly",
			labels: map[string]string{"group": "web"},
			validateFunc: func(got string) {
				assert.Equal(suite.T(), "group:web", got)
			},
		},
		{
			name:   "when multiple labels sorts by key",
			labels: map[string]string{"group": "web", "env": "prod", "az": "us-east"},
			validateFunc: func(got string) {
				assert.Equal(suite.T(), "az:us-east, env:prod, group:web", got)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := cli.FormatLabels(tc.labels)
			tc.validateFunc(result)
		})
	}
}

func (suite *UIPublicTestSuite) TestBuildMutationTable() {
	errMsg := "interface not found"

	tests := []struct {
		name         string
		results      []cli.ResultRow
		fieldHeaders []string
		wantHeaders  []string
		validateFunc func([][]string)
	}{
		{
			name:         "when no results returns hostname and status headers",
			results:      []cli.ResultRow{},
			fieldHeaders: nil,
			wantHeaders:  []string{"HOSTNAME", "STATUS"},
			validateFunc: func(got [][]string) {
				assert.Equal(suite.T(), [][]string{}, got)
			},
		},
		{
			name: "when all succeed shows ok status",
			results: []cli.ResultRow{
				{Hostname: "web-01", Status: "ok"},
				{Hostname: "web-02", Status: "ok"},
			},
			fieldHeaders: nil,
			wantHeaders:  []string{"HOSTNAME", "STATUS"},
			validateFunc: func(got [][]string) {
				assert.Equal(suite.T(), [][]string{
					{"web-01", "ok"},
					{"web-02", "ok"},
				}, got)
			},
		},
		{
			name: "when some fail shows err status",
			results: []cli.ResultRow{
				{Hostname: "web-01", Status: "ok"},
				{Hostname: "web-02", Status: "failed", Error: &errMsg},
			},
			fieldHeaders: nil,
			wantHeaders:  []string{"HOSTNAME", "STATUS"},
			validateFunc: func(got [][]string) {
				assert.Equal(suite.T(), [][]string{
					{"web-01", "ok"},
				}, got)
			},
		},
		{
			name: "when single host shows hostname and status",
			results: []cli.ResultRow{
				{Hostname: "web-01", Status: "ok", Fields: []string{"extra"}},
			},
			fieldHeaders: []string{"DETAIL"},
			wantHeaders:  []string{"HOSTNAME", "STATUS", "DETAIL"},
			validateFunc: func(got [][]string) {
				assert.Equal(suite.T(), [][]string{
					{"web-01", "ok", "extra"},
				}, got)
			},
		},
		{
			name: "when changed is true shows changed status",
			results: []cli.ResultRow{
				{Hostname: "web-01", Status: "ok", Changed: ptr.To(true)},
				{Hostname: "web-02", Status: "ok", Changed: ptr.To(true)},
			},
			fieldHeaders: nil,
			wantHeaders:  []string{"HOSTNAME", "STATUS"},
			validateFunc: func(got [][]string) {
				assert.Equal(suite.T(), [][]string{
					{"web-01", "changed"},
					{"web-02", "changed"},
				}, got)
			},
		},
		{
			name: "when changed is false shows ok status",
			results: []cli.ResultRow{
				{Hostname: "web-01", Status: "ok", Changed: ptr.To(false)},
				{Hostname: "web-02", Status: "ok", Changed: ptr.To(false)},
			},
			fieldHeaders: nil,
			wantHeaders:  []string{"HOSTNAME", "STATUS"},
			validateFunc: func(got [][]string) {
				assert.Equal(suite.T(), [][]string{
					{"web-01", "ok"},
					{"web-02", "ok"},
				}, got)
			},
		},
		{
			name: "when field headers provided includes them after status",
			results: []cli.ResultRow{
				{Hostname: "web-01", Changed: ptr.To(true), Fields: []string{"val1"}},
				{Hostname: "web-02", Fields: []string{"val2"}},
			},
			fieldHeaders: []string{"DATA"},
			wantHeaders:  []string{"HOSTNAME", "STATUS", "DATA"},
			validateFunc: func(got [][]string) {
				assert.Equal(suite.T(), [][]string{
					{"web-01", "changed", "val1"},
					{"web-02", "ok", "val2"},
				}, got)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tr := cli.BuildMutationTable(tc.results, tc.fieldHeaders)

			assert.Equal(suite.T(), tc.wantHeaders, tr.Headers)
			tc.validateFunc(tr.Rows)
		})
	}
}

func (suite *UIPublicTestSuite) TestBuildBroadcastTableResult() {
	errMsg := "connection refused"
	skipMsg := "unsupported"

	tests := []struct {
		name         string
		results      []cli.ResultRow
		fieldHdrs    []string
		wantHeaders  []string
		wantRows     [][]string
		validateFunc func([]cli.ErrorEntry)
	}{
		{
			name: "when errors exist they appear in errors field",
			results: []cli.ResultRow{
				{Hostname: "web-01", Fields: []string{"val1"}},
				{Hostname: "web-02", Error: &errMsg, Fields: []string{""}},
			},
			fieldHdrs:   []string{"DATA"},
			wantHeaders: []string{"HOSTNAME", "STATUS", "DATA"},
			wantRows: [][]string{
				{"web-01", "ok", "val1"},
			},
			validateFunc: func(got []cli.ErrorEntry) {
				assert.Equal(suite.T(), []cli.ErrorEntry{
					{Hostname: "web-02", Message: "connection refused", Status: "err"},
				}, got)
			},
		},
		{
			name: "when no errors the errors field is nil",
			results: []cli.ResultRow{
				{Hostname: "web-01", Fields: []string{"val1"}},
				{Hostname: "web-02", Fields: []string{"val2"}},
			},
			fieldHdrs:   []string{"DATA"},
			wantHeaders: []string{"HOSTNAME", "STATUS", "DATA"},
			wantRows: [][]string{
				{"web-01", "ok", "val1"},
				{"web-02", "ok", "val2"},
			},
			validateFunc: func(got []cli.ErrorEntry) {
				assert.Equal(suite.T(), []cli.ErrorEntry(nil), got)
			},
		},
		{
			name: "when skipped host without fields appears in errors",
			results: []cli.ResultRow{
				{Hostname: "web-01", Fields: []string{"val1"}},
				{Hostname: "web-02", Status: "skipped", Error: &skipMsg},
			},
			fieldHdrs:   []string{"DATA"},
			wantHeaders: []string{"HOSTNAME", "STATUS", "DATA"},
			wantRows: [][]string{
				{"web-01", "ok", "val1"},
			},
			validateFunc: func(got []cli.ErrorEntry) {
				assert.Equal(suite.T(), []cli.ErrorEntry{
					{Hostname: "web-02", Message: "unsupported", Status: "skip"},
				}, got)
			},
		},
		{
			name: "when multiple errors collects all",
			results: []cli.ResultRow{
				{Hostname: "web-01", Error: &errMsg},
				{Hostname: "web-02", Error: &skipMsg},
				{Hostname: "web-03", Fields: []string{"ok"}},
			},
			fieldHdrs:   []string{"DATA"},
			wantHeaders: []string{"HOSTNAME", "STATUS", "DATA"},
			wantRows: [][]string{
				{"web-03", "ok", "ok"},
			},
			validateFunc: func(got []cli.ErrorEntry) {
				assert.Equal(suite.T(), []cli.ErrorEntry{
					{Hostname: "web-01", Message: "connection refused", Status: "err"},
					{Hostname: "web-02", Message: "unsupported", Status: "err"},
				}, got)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := cli.BuildBroadcastTable(tc.results, tc.fieldHdrs)

			assert.Equal(suite.T(), tc.wantHeaders, result.Headers)
			assert.Equal(suite.T(), tc.wantRows, result.Rows)
			tc.validateFunc(result.Errors)
		})
	}
}

func (suite *UIPublicTestSuite) TestBuildMutationTableResult() {
	errMsg := "permission denied"

	tests := []struct {
		name         string
		results      []cli.ResultRow
		fieldHdrs    []string
		wantHeaders  []string
		wantRows     [][]string
		validateFunc func([]cli.ErrorEntry)
	}{
		{
			name: "when mutation errors exist they appear in errors field",
			results: []cli.ResultRow{
				{Hostname: "web-01", Changed: ptr.To(true)},
				{Hostname: "web-02", Error: &errMsg},
			},
			fieldHdrs:   nil,
			wantHeaders: []string{"HOSTNAME", "STATUS"},
			wantRows: [][]string{
				{"web-01", "changed"},
			},
			validateFunc: func(got []cli.ErrorEntry) {
				assert.Equal(suite.T(), []cli.ErrorEntry{
					{Hostname: "web-02", Message: "permission denied", Status: "err"},
				}, got)
			},
		},
		{
			name: "when no mutation errors the errors field is nil",
			results: []cli.ResultRow{
				{Hostname: "web-01", Changed: ptr.To(true)},
				{Hostname: "web-02", Changed: ptr.To(false)},
			},
			fieldHdrs:   nil,
			wantHeaders: []string{"HOSTNAME", "STATUS"},
			wantRows: [][]string{
				{"web-01", "changed"},
				{"web-02", "ok"},
			},
			validateFunc: func(got []cli.ErrorEntry) {
				assert.Equal(suite.T(), []cli.ErrorEntry(nil), got)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := cli.BuildMutationTable(tc.results, tc.fieldHdrs)

			assert.Equal(suite.T(), tc.wantHeaders, result.Headers)
			assert.Equal(suite.T(), tc.wantRows, result.Rows)
			tc.validateFunc(result.Errors)
		})
	}
}

func (suite *UIPublicTestSuite) TestFormatList() {
	tests := []struct {
		name         string
		list         []string
		validateFunc func(string)
	}{
		{
			name: "when empty returns None",
			list: []string{},
			validateFunc: func(got string) {
				assert.Equal(suite.T(), "None", got)
			},
		},
		{
			name: "when single item returns it",
			list: []string{"alpha"},
			validateFunc: func(got string) {
				assert.Equal(suite.T(), "alpha", got)
			},
		},
		{
			name: "when multiple items joins with comma",
			list: []string{"alpha", "beta", "gamma"},
			validateFunc: func(got string) {
				assert.Equal(suite.T(), "alpha, beta, gamma", got)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			got := cli.FormatList(tc.list)

			tc.validateFunc(got)
		})
	}
}

func (suite *UIPublicTestSuite) TestCalculateColumnWidths() {
	tests := []struct {
		name         string
		headers      []string
		rows         [][]string
		minPadding   int
		validateFunc func([]int)
	}{
		{
			name:       "when empty headers returns empty",
			headers:    []string{},
			rows:       nil,
			minPadding: 1,
			validateFunc: func(got []int) {
				assert.Equal(suite.T(), []int{}, got)
			},
		},
		{
			name:       "when headers wider than rows uses header width",
			headers:    []string{"HOSTNAME", "STATUS"},
			rows:       [][]string{{"a", "b"}},
			minPadding: 1,
			validateFunc: func(got []int) {
				assert.Equal(suite.T(), []int{10, 8}, got)
			},
		},
		{
			name:       "when rows wider than headers uses row width",
			headers:    []string{"A", "B"},
			rows:       [][]string{{"longvalue", "anotherlongvalue"}},
			minPadding: 1,
			validateFunc: func(got []int) {
				assert.Equal(suite.T(), []int{11, 18}, got)
			},
		},
		{
			name:       "when multi-line content uses longest line width",
			headers:    []string{"DATA"},
			rows:       [][]string{{"short\nvery long line here"}},
			minPadding: 0,
			validateFunc: func(got []int) {
				assert.Equal(suite.T(), []int{19}, got)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			got := cli.CalculateColumnWidths(tc.headers, tc.rows, tc.minPadding)

			tc.validateFunc(got)
		})
	}
}

func (suite *UIPublicTestSuite) TestGetMaxLineWidth() {
	tests := []struct {
		name         string
		text         string
		validateFunc func(int)
	}{
		{
			name: "when single line returns its length",
			text: "hello",
			validateFunc: func(got int) {
				assert.Equal(suite.T(), 5, got)
			},
		},
		{
			name: "when multi-line returns longest",
			text: "short\na much longer line\nmed",
			validateFunc: func(got int) {
				assert.Equal(suite.T(), 18, got)
			},
		},
		{
			name: "when empty returns zero",
			text: "",
			validateFunc: func(got int) {
				assert.Equal(suite.T(), 0, got)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			got := cli.GetMaxLineWidth(tc.text)

			tc.validateFunc(got)
		})
	}
}

func (suite *UIPublicTestSuite) TestSafeString() {
	str := "hello"

	tests := []struct {
		name         string
		s            *string
		validateFunc func(string)
	}{
		{
			name: "when non-nil returns value",
			s:    &str,
			validateFunc: func(got string) {
				assert.Equal(suite.T(), "hello", got)
			},
		},
		{
			name: "when nil returns empty",
			s:    nil,
			validateFunc: func(got string) {
				assert.Equal(suite.T(), "", got)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			got := cli.SafeString(tc.s)

			tc.validateFunc(got)
		})
	}
}

func (suite *UIPublicTestSuite) TestSafeUUID() {
	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		name         string
		u            *uuid.UUID
		validateFunc func(string)
	}{
		{
			name: "when non-nil returns string",
			u:    &id,
			validateFunc: func(got string) {
				assert.Equal(suite.T(), "550e8400-e29b-41d4-a716-446655440000", got)
			},
		},
		{
			name: "when nil returns empty",
			u:    nil,
			validateFunc: func(got string) {
				assert.Equal(suite.T(), "", got)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			got := cli.SafeUUID(tc.u)

			tc.validateFunc(got)
		})
	}
}

func (suite *UIPublicTestSuite) TestFloat64ToSafeString() {
	val := 3.14

	tests := []struct {
		name         string
		f            *float64
		validateFunc func(string)
	}{
		{
			name: "when non-nil returns formatted float",
			f:    &val,
			validateFunc: func(got string) {
				assert.Equal(suite.T(), "3.140000", got)
			},
		},
		{
			name: "when nil returns N/A",
			f:    nil,
			validateFunc: func(got string) {
				assert.Equal(suite.T(), "N/A", got)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			got := cli.Float64ToSafeString(tc.f)

			tc.validateFunc(got)
		})
	}
}

func (suite *UIPublicTestSuite) TestIntToSafeString() {
	val := 42

	tests := []struct {
		name         string
		i            *int
		validateFunc func(string)
	}{
		{
			name: "when non-nil returns formatted int",
			i:    &val,
			validateFunc: func(got string) {
				assert.Equal(suite.T(), "42", got)
			},
		},
		{
			name: "when nil returns N/A",
			i:    nil,
			validateFunc: func(got string) {
				assert.Equal(suite.T(), "N/A", got)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			got := cli.IntToSafeString(tc.i)

			tc.validateFunc(got)
		})
	}
}

func (suite *UIPublicTestSuite) TestHandleError() {
	tests := []struct {
		name         string
		err          error
		validateFunc func(string, int)
	}{
		{
			name: "when auth error logs api error with status code",
			err: &client.AuthError{
				APIError: client.APIError{StatusCode: 403, Message: "insufficient permissions"},
			},
			validateFunc: func(log string, exitCode int) {
				assert.Contains(suite.T(), log, "insufficient permissions")
				assert.Equal(suite.T(), 1, exitCode)
			},
		},
		{
			name: "when not found error logs api error with status code",
			err: &client.NotFoundError{
				APIError: client.APIError{StatusCode: 404, Message: "job not found"},
			},
			validateFunc: func(log string, exitCode int) {
				assert.Contains(suite.T(), log, "job not found")
				assert.Equal(suite.T(), 1, exitCode)
			},
		},
		{
			name: "when validation error logs api error with status code",
			err: &client.ValidationError{
				APIError: client.APIError{StatusCode: 400, Message: "invalid input"},
			},
			validateFunc: func(log string, exitCode int) {
				assert.Contains(suite.T(), log, "invalid input")
				assert.Equal(suite.T(), 1, exitCode)
			},
		},
		{
			name: "when server error logs api error with status code",
			err: &client.ServerError{
				APIError: client.APIError{StatusCode: 500, Message: "internal server error"},
			},
			validateFunc: func(log string, exitCode int) {
				assert.Contains(suite.T(), log, "internal server error")
				assert.Equal(suite.T(), 1, exitCode)
			},
		},
		{
			name: "when generic error logs error message",
			err:  fmt.Errorf("connection refused"),
			validateFunc: func(log string, exitCode int) {
				assert.Contains(suite.T(), log, "connection refused")
				assert.Equal(suite.T(), 1, exitCode)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			// Prevent os.Exit from killing the test process.
			var exitCode int
			cli.SetOsExit(func(code int) { exitCode = code })
			defer cli.ResetOsExit()

			defer cli.ResetErrOut()

			// With --json the error is one JSON document, and no log.
			var logBuf bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&logBuf, nil))
			var jsonBuf bytes.Buffer
			cli.SetErrOut(&jsonBuf)
			cli.SetJSONOutput(true)
			cli.HandleError(tc.err, logger)
			tc.validateFunc(jsonBuf.String(), exitCode)
			var doc cli.ErrorDocument
			suite.Require().NoError(json.Unmarshal(jsonBuf.Bytes(), &doc))
			assert.NotEmpty(suite.T(), doc.Code)
			assert.Empty(suite.T(), logBuf.String())

			// Without it the error is one line for a person, and no log.
			exitCode = 0
			logBuf.Reset()
			var errBuf bytes.Buffer
			cli.SetErrOut(&errBuf)
			cli.SetJSONOutput(false)
			cli.HandleError(tc.err, logger)
			tc.validateFunc(errBuf.String(), exitCode)
			assert.True(suite.T(), strings.HasPrefix(errBuf.String(), "[err] "))
			assert.Empty(suite.T(), logBuf.String())

			var apiErr *client.APIError
			if errors.As(tc.err, &apiErr) {
				assert.Contains(
					suite.T(),
					errBuf.String(),
					fmt.Sprintf("api error (%d): ", apiErr.StatusCode),
				)
			}
		})
	}
}

func (suite *UIPublicTestSuite) TestPrintKV() {
	tests := []struct {
		name  string
		pairs []string
		// noHeading resets the heading color, as before any heading prints.
		noHeading    bool
		validateFunc func(string)
	}{
		{
			name:  "when valid pairs prints output",
			pairs: []string{"Key", "Value"},
			validateFunc: func(output string) {
				assert.NotEmpty(suite.T(), output)
			},
		},
		{
			name:  "when multiple pairs prints all",
			pairs: []string{"Name", "test", "Status", "ok"},
			validateFunc: func(output string) {
				assert.NotEmpty(suite.T(), output)
			},
		},
		{
			name:      "when no heading came before, the label still prints",
			pairs:     []string{"Name", "x"},
			noHeading: true,
			validateFunc: func(output string) {
				suite.Contains(output, "Name:")
			},
		},
		{
			name:  "when odd number of pairs prints nothing",
			pairs: []string{"Key"},
			validateFunc: func(output string) {
				assert.Empty(suite.T(), output)
			},
		},
		{
			name:  "when empty prints nothing",
			pairs: []string{},
			validateFunc: func(output string) {
				assert.Empty(suite.T(), output)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			if tc.noHeading {
				cli.ResetTint()
			}
			tc.validateFunc(captureStdout(func() {
				cli.PrintKV(tc.pairs...)
				cli.FlushKV()
			}))
		})
	}
}

func (suite *UIPublicTestSuite) TestPrintCompactTable() {
	tests := []struct {
		name         string
		sections     []cli.Section
		validateFunc func(string)
	}{
		{
			name: "when section with title renders table",
			sections: []cli.Section{
				{
					Title:   "Test",
					Headers: []string{"COL1", "COL2"},
					Rows:    [][]string{{"a", "b"}},
				},
			},
			validateFunc: func(output string) {
				assert.NotEmpty(suite.T(), output)
				for _, h := range []string{"COL1", "COL2"} {
					assert.Contains(suite.T(), output, h)
				}
			},
		},
		{
			name: "when section without title renders table",
			sections: []cli.Section{
				{
					Headers: []string{"COL1"},
					Rows:    [][]string{{"a"}},
				},
			},
			validateFunc: func(output string) {
				assert.NotEmpty(suite.T(), output)
				for _, h := range []string{"COL1"} {
					assert.Contains(suite.T(), output, h)
				}
			},
		},
		{
			name: "when wide data aligns columns",
			sections: []cli.Section{
				{
					Title:   "Wide",
					Headers: []string{"A", "B", "C", "D", "E"},
					Rows: [][]string{{
						"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
						"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
						"ccccccccccccccccccccccccccccccc",
						"ddddddddddddddddddddddddddddd",
						"eeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
					}},
				},
			},
			validateFunc: func(output string) {
				assert.NotEmpty(suite.T(), output)
				for _, h := range []string{"A", "B", "C", "D", "E"} {
					assert.Contains(suite.T(), output, h)
				}
			},
		},
		{
			name: "when many columns renders all headers",
			sections: []cli.Section{
				{
					Headers: []string{
						"X", "Y", "Z",
						"LONG-HEADER-1", "LONG-HEADER-2",
					},
					Rows: [][]string{{
						"a", "b", "c",
						"aaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbb",
					}},
				},
			},
			validateFunc: func(output string) {
				assert.NotEmpty(suite.T(), output)
				for _, h := range []string{"X", "Y", "Z", "LONG-HEADER-1", "LONG-HEADER-2"} {
					assert.Contains(suite.T(), output, h)
				}
			},
		},
		{
			name: "when multiple rows alternates colors",
			sections: []cli.Section{
				{
					Headers: []string{"NAME", "VALUE"},
					Rows: [][]string{
						{"row0", "even"},
						{"row1", "odd"},
						{"row2", "even"},
					},
				},
			},
			validateFunc: func(output string) {
				assert.NotEmpty(suite.T(), output)
				for _, h := range []string{"NAME", "VALUE"} {
					assert.Contains(suite.T(), output, h)
				}
			},
		},
		{
			name: "when column exceeds max width truncates to cap",
			sections: []cli.Section{
				{
					Headers: []string{"A", "B"},
					Rows: [][]string{{
						strings.Repeat("x", 90),
						"short",
					}},
				},
			},
			validateFunc: func(output string) {
				assert.NotEmpty(suite.T(), output)
				for _, h := range []string{"A", "B"} {
					assert.Contains(suite.T(), output, h)
				}
			},
		},
		{
			name: "when cell exceeds column width shows ellipsis",
			sections: []cli.Section{
				{
					Headers: []string{"A"},
					Rows: [][]string{
						{strings.Repeat("x", 90)},
					},
				},
			},
			validateFunc: func(output string) {
				assert.NotEmpty(suite.T(), output)
				for _, h := range []string{"A"} {
					assert.Contains(suite.T(), output, h)
				}
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tc.validateFunc(captureStdout(func() {
				cli.PrintCompactTable(tc.sections)
			}))
		})
	}
}

func (suite *UIPublicTestSuite) TestFormatAge() {
	tests := []struct {
		name         string
		d            time.Duration
		validateFunc func(string)
	}{
		{
			name: "when zero returns empty",
			d:    0,
			validateFunc: func(got string) {
				assert.Equal(suite.T(), "", got)
			},
		},
		{
			name: "when negative returns empty",
			d:    -1 * time.Hour,
			validateFunc: func(got string) {
				assert.Equal(suite.T(), "", got)
			},
		},
		{
			name: "when days and hours formats as days",
			d:    3*24*time.Hour + 4*time.Hour,
			validateFunc: func(got string) {
				assert.Equal(suite.T(), "3d 4h", got)
			},
		},
		{
			name: "when hours and minutes formats as hours",
			d:    12*time.Hour + 30*time.Minute,
			validateFunc: func(got string) {
				assert.Equal(suite.T(), "12h 30m", got)
			},
		},
		{
			name: "when only minutes formats as minutes",
			d:    45 * time.Minute,
			validateFunc: func(got string) {
				assert.Equal(suite.T(), "45m", got)
			},
		},
		{
			name: "when only seconds formats as seconds",
			d:    30 * time.Second,
			validateFunc: func(got string) {
				assert.Equal(suite.T(), "30s", got)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			got := cli.FormatAge(tc.d)

			tc.validateFunc(got)
		})
	}
}

func (suite *UIPublicTestSuite) TestFormatBytes() {
	tests := []struct {
		name         string
		b            int
		validateFunc func(string)
	}{
		{
			name: "when bytes returns bytes",
			b:    512,
			validateFunc: func(got string) {
				assert.Equal(suite.T(), "512 B", got)
			},
		},
		{
			name: "when kilobytes returns KB",
			b:    5 * 1024,
			validateFunc: func(got string) {
				assert.Equal(suite.T(), "5.0 KB", got)
			},
		},
		{
			name: "when megabytes returns MB",
			b:    3 * 1024 * 1024,
			validateFunc: func(got string) {
				assert.Equal(suite.T(), "3.0 MB", got)
			},
		},
		{
			name: "when gigabytes returns GB",
			b:    2 * 1024 * 1024 * 1024,
			validateFunc: func(got string) {
				assert.Equal(suite.T(), "2.0 GB", got)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			got := cli.FormatBytes(tc.b)

			tc.validateFunc(got)
		})
	}
}

func (suite *UIPublicTestSuite) TestDisplayJobDetail() {
	tests := []struct {
		name         string
		resp         *client.JobDetail
		validateFunc func(string)
	}{
		{
			name: "when minimal response displays job info",
			resp: &client.JobDetail{
				Status: "completed",
			},
			validateFunc: func(output string) {
				assert.NotEmpty(suite.T(), output)
			},
		},
		{
			name: "when full response displays all sections",
			resp: &client.JobDetail{
				ID:        "550e8400-e29b-41d4-a716-446655440000",
				Status:    "completed",
				Hostname:  "web-01",
				Created:   "2026-01-01T00:00:00Z",
				UpdatedAt: "2026-01-01T00:01:00Z",
				Error:     "timeout",
				Operation: map[string]any{"type": "node.hostname"},
				Result:    map[string]any{"hostname": "web-01"},
				Timeline: []client.TimelineEvent{
					{
						Event:     "completed",
						Timestamp: "2026-01-01T00:01:00Z",
						Hostname:  "web-01",
						Message:   "job completed",
					},
				},
				AgentStates: map[string]client.AgentState{
					"web-01": {
						Status:   "completed",
						Duration: "1.5s",
					},
				},
				Responses: map[string]client.AgentJobResponse{
					"web-01": {
						Status: "ok",
						Data:   map[string]string{"key": "val"},
					},
				},
			},
			validateFunc: func(output string) {
				assert.NotEmpty(suite.T(), output)
			},
		},
		{
			name: "when agent states with multiple agents shows summary",
			resp: &client.JobDetail{
				Status: "completed",
				AgentStates: map[string]client.AgentState{
					"web-01": {Status: "completed", Duration: "1s"},
					"web-02": {Status: "failed", Duration: "1s", Error: "error"},
					"web-03": {Status: "started", Duration: "1s"},
				},
			},
			validateFunc: func(output string) {
				assert.NotEmpty(suite.T(), output)
			},
		},
		{
			name: "when response has nil data shows no data placeholder",
			resp: &client.JobDetail{
				Status: "completed",
				Responses: map[string]client.AgentJobResponse{
					"web-01": {
						Status: "ok",
						Data:   nil,
					},
				},
			},
			validateFunc: func(output string) {
				assert.NotEmpty(suite.T(), output)
			},
		},
		{
			name: "when response has error shows error message",
			resp: &client.JobDetail{
				Status: "failed",
				Responses: map[string]client.AgentJobResponse{
					"web-01": {
						Status: "failed",
						Error:  "timeout",
					},
				},
			},
			validateFunc: func(output string) {
				assert.NotEmpty(suite.T(), output)
			},
		},
		{
			name: "when timeline has error shows error message",
			resp: &client.JobDetail{
				Status: "failed",
				Timeline: []client.TimelineEvent{
					{
						Event:     "failed",
						Timestamp: "2026-01-01T00:01:00Z",
						Error:     "connection refused",
					},
				},
			},
			validateFunc: func(output string) {
				assert.NotEmpty(suite.T(), output)
			},
		},
		{
			name: "when agent states contain skipped status",
			resp: &client.JobDetail{
				Status: "completed",
				AgentStates: map[string]client.AgentState{
					"web-01": {Status: "completed", Duration: "1s"},
					"web-02": {Status: "skipped", Duration: "0s"},
					"web-03": {Status: "completed", Duration: "2s"},
				},
			},
			validateFunc: func(output string) {
				assert.NotEmpty(suite.T(), output)
			},
		},
		{
			name: "when multiple agents skipped shows skipped in summary",
			resp: &client.JobDetail{
				Status: "skipped",
				AgentStates: map[string]client.AgentState{
					"web-01": {Status: "skipped"},
					"web-02": {Status: "skipped"},
					"web-03": {Status: "completed", Duration: "1s"},
				},
			},
			validateFunc: func(output string) {
				assert.NotEmpty(suite.T(), output)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tc.validateFunc(captureStdout(func() {
				cli.DisplayJobDetail(tc.resp)
			}))
		})
	}
}

func (suite *UIPublicTestSuite) TestStatusWeight() {
	tests := []struct {
		name         string
		status       string
		validateFunc func(int)
	}{
		{
			name:   "when ok returns 0",
			status: "ok",
			validateFunc: func(got int) {
				assert.Equal(suite.T(), 0, got)
			},
		},
		{
			name:   "when changed returns 1",
			status: "changed",
			validateFunc: func(got int) {
				assert.Equal(suite.T(), 1, got)
			},
		},
		{
			name:   "when skip returns 2",
			status: "skip",
			validateFunc: func(got int) {
				assert.Equal(suite.T(), 2, got)
			},
		},
		{
			name:   "when err returns 3",
			status: "err",
			validateFunc: func(got int) {
				assert.Equal(suite.T(), 3, got)
			},
		},
		{
			name:   "when unknown returns 0",
			status: "unknown",
			validateFunc: func(got int) {
				assert.Equal(suite.T(), 0, got)
			},
		},
		{
			name:   "when empty returns 0",
			status: "",
			validateFunc: func(got int) {
				assert.Equal(suite.T(), 0, got)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			got := cli.ExportStatusWeight(tc.status)
			tc.validateFunc(got)
		})
	}
}

func (suite *UIPublicTestSuite) TestResolveStatus() {
	errMsg := "failed"

	tests := []struct {
		name         string
		row          cli.ResultRow
		validateFunc func(string)
	}{
		{
			name: "when no error and no change returns ok",
			row:  cli.ResultRow{Hostname: "h1"},
			validateFunc: func(got string) {
				assert.Equal(suite.T(), "ok", got)
			},
		},
		{
			name: "when changed true returns changed",
			row:  cli.ResultRow{Hostname: "h1", Changed: ptr.To(true)},
			validateFunc: func(got string) {
				assert.Equal(suite.T(), "changed", got)
			},
		},
		{
			name: "when changed false returns ok",
			row:  cli.ResultRow{Hostname: "h1", Changed: ptr.To(false)},
			validateFunc: func(got string) {
				assert.Equal(suite.T(), "ok", got)
			},
		},
		{
			name: "when error returns err",
			row:  cli.ResultRow{Hostname: "h1", Error: &errMsg},
			validateFunc: func(got string) {
				assert.Equal(suite.T(), "err", got)
			},
		},
		{
			name: "when status skipped returns skip even with error",
			row:  cli.ResultRow{Hostname: "h1", Status: "skipped", Error: &errMsg},
			validateFunc: func(got string) {
				assert.Equal(suite.T(), "skip", got)
			},
		},
		{
			name: "when status skip returns skip",
			row:  cli.ResultRow{Hostname: "h1", Status: "skip"},
			validateFunc: func(got string) {
				assert.Equal(suite.T(), "skip", got)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			got := cli.ExportResolveStatus(tc.row)
			tc.validateFunc(got)
		})
	}
}

func (suite *UIPublicTestSuite) TestPrintSummary() {
	tests := []struct {
		name         string
		section      cli.Section
		validateFunc func(output string)
	}{
		{
			name: "when mixed statuses shows counts",
			section: cli.Section{
				Rows: [][]string{
					{"web-01", "ok", "val"},
					{"web-02", "skip"},
					{"web-03", "err"},
				},
			},
			validateFunc: func(output string) {
				assert.Contains(suite.T(), output, "3 hosts:")
				assert.Contains(suite.T(), output, "1 ok")
				assert.Contains(suite.T(), output, "1 skipped")
				assert.Contains(suite.T(), output, "1 failed")
			},
		},
		{
			name: "when all ok shows only ok count",
			section: cli.Section{
				Rows: [][]string{
					{"web-01", "ok", "val"},
					{"web-02", "ok", "val"},
				},
			},
			validateFunc: func(output string) {
				assert.Contains(suite.T(), output, "2 hosts:")
				assert.Contains(suite.T(), output, "2 ok")
			},
		},
		{
			name: "when changed shows changed count",
			section: cli.Section{
				Rows: [][]string{
					{"web-01", "changed"},
					{"web-02", "ok"},
				},
			},
			validateFunc: func(output string) {
				assert.Contains(suite.T(), output, "2 hosts:")
				assert.Contains(suite.T(), output, "1 changed")
				assert.Contains(suite.T(), output, "1 ok")
			},
		},
		{
			name: "when duration set shows duration",
			section: cli.Section{
				Duration: "286ms",
				Rows: [][]string{
					{"web-01", "ok"},
				},
			},
			validateFunc: func(output string) {
				assert.Contains(suite.T(), output, "286ms")
			},
		},
		{
			name: "when reference section does not print summary line",
			section: cli.Section{
				Reference: true,
				Rows: [][]string{
					{"sysctl", "list"},
				},
			},
			validateFunc: func(output string) {
				assert.NotContains(suite.T(), output, "hosts:")
			},
		},
		{
			name: "when no rows does not print summary line",
			section: cli.Section{
				Rows: [][]string{},
			},
			validateFunc: func(output string) {
				assert.NotContains(suite.T(), output, "hosts:")
			},
		},
		{
			name: "when error host not in rows counts from errors",
			section: cli.Section{
				Rows: [][]string{
					{"web-01", "ok"},
				},
				Errors: []cli.ErrorEntry{
					{Hostname: "web-02", Message: "timeout", Status: "err"},
				},
			},
			validateFunc: func(output string) {
				assert.Contains(suite.T(), output, "2 hosts:")
				assert.Contains(suite.T(), output, "1 ok")
				assert.Contains(suite.T(), output, "1 failed")
			},
		},
		{
			name: "when same host has multiple rows uses worst status",
			section: cli.Section{
				Rows: [][]string{
					{"web-01", "ok", "val1"},
					{"web-01", "err"},
				},
			},
			validateFunc: func(output string) {
				assert.Contains(suite.T(), output, "1 host:")
				assert.Contains(suite.T(), output, "1 failed")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			output := captureStdout(func() {
				cli.PrintCompactTable([]cli.Section{tc.section})
			})
			tc.validateFunc(output)
		})
	}
}

func (suite *UIPublicTestSuite) TestPrintErrors() {
	tests := []struct {
		name         string
		errors       []cli.ErrorEntry
		validateFunc func(output string)
	}{
		{
			name:   "when no errors prints nothing",
			errors: nil,
			validateFunc: func(output string) {
				assert.Empty(suite.T(), output)
			},
		},
		{
			name: "when error shows red details",
			errors: []cli.ErrorEntry{
				{Hostname: "web-01", Message: "connection refused", Status: "err"},
			},
			validateFunc: func(output string) {
				assert.Contains(suite.T(), output, "Details:")
				assert.Contains(suite.T(), output, "web-01")
				assert.Contains(suite.T(), output, "connection refused")
			},
		},
		{
			name: "when skip shows details",
			errors: []cli.ErrorEntry{
				{Hostname: "mac-01", Message: "unsupported OS", Status: "skip"},
			},
			validateFunc: func(output string) {
				assert.Contains(suite.T(), output, "Details:")
				assert.Contains(suite.T(), output, "mac-01")
				assert.Contains(suite.T(), output, "unsupported OS")
			},
		},
		{
			name: "when multiple entries shows all",
			errors: []cli.ErrorEntry{
				{Hostname: "web-01", Message: "timeout", Status: "err"},
				{Hostname: "mac-01", Message: "unsupported", Status: "skip"},
			},
			validateFunc: func(output string) {
				assert.Contains(suite.T(), output, "web-01")
				assert.Contains(suite.T(), output, "mac-01")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			output := captureStdout(func() {
				cli.PrintErrors(tc.errors)
			})
			tc.validateFunc(output)
		})
	}
}
