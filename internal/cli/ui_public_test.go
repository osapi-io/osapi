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
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/cli"
	"github.com/retr0h/osapi/internal/client/gen"
)

type UITestSuite struct {
	suite.Suite
}

func TestUITestSuite(t *testing.T) {
	suite.Run(t, new(UITestSuite))
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

func (suite *UITestSuite) TestBuildBroadcastTable() {
	errMsg := "connection refused"

	tests := []struct {
		name         string
		results      []cli.ResultRow
		fieldHeaders []string
		wantHeaders  []string
		wantRows     [][]string
	}{
		{
			name:         "when no results returns empty",
			results:      []cli.ResultRow{},
			fieldHeaders: nil,
			wantHeaders:  []string{"HOSTNAME"},
			wantRows:     [][]string{},
		},
		{
			name: "when all results succeed omits status and error columns",
			results: []cli.ResultRow{
				{Hostname: "web-01", Fields: []string{"val1"}},
				{Hostname: "web-02", Fields: []string{"val2"}},
			},
			fieldHeaders: []string{"DATA"},
			wantHeaders:  []string{"HOSTNAME", "DATA"},
			wantRows: [][]string{
				{"web-01", "val1"},
				{"web-02", "val2"},
			},
		},
		{
			name: "when some results have errors adds status and error columns",
			results: []cli.ResultRow{
				{Hostname: "web-01", Fields: []string{"val1"}},
				{Hostname: "web-02", Error: &errMsg, Fields: []string{""}},
			},
			fieldHeaders: []string{"DATA"},
			wantHeaders:  []string{"HOSTNAME", "STATUS", "ERROR", "DATA"},
			wantRows: [][]string{
				{"web-01", "ok", "", "val1"},
				{"web-02", "failed", "connection refused", ""},
			},
		},
		{
			name: "when no field headers works with hostname only",
			results: []cli.ResultRow{
				{Hostname: "web-01"},
			},
			fieldHeaders: nil,
			wantHeaders:  []string{"HOSTNAME"},
			wantRows: [][]string{
				{"web-01"},
			},
		},
		{
			name: "when all results have errors all show failed",
			results: []cli.ResultRow{
				{Hostname: "web-01", Error: &errMsg, Fields: []string{""}},
				{Hostname: "web-02", Error: &errMsg, Fields: []string{""}},
			},
			fieldHeaders: []string{"DATA"},
			wantHeaders:  []string{"HOSTNAME", "STATUS", "ERROR", "DATA"},
			wantRows: [][]string{
				{"web-01", "failed", "connection refused", ""},
				{"web-02", "failed", "connection refused", ""},
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			headers, rows := cli.BuildBroadcastTable(tc.results, tc.fieldHeaders)

			assert.Equal(suite.T(), tc.wantHeaders, headers)
			assert.Equal(suite.T(), tc.wantRows, rows)
		})
	}
}

func (suite *UITestSuite) TestFormatLabels() {
	tests := []struct {
		name   string
		labels *map[string]string
		want   string
	}{
		{
			name:   "when nil returns empty",
			labels: nil,
			want:   "",
		},
		{
			name:   "when empty map returns empty",
			labels: &map[string]string{},
			want:   "",
		},
		{
			name:   "when single label formats correctly",
			labels: &map[string]string{"group": "web"},
			want:   "group:web",
		},
		{
			name:   "when multiple labels sorts by key",
			labels: &map[string]string{"group": "web", "env": "prod", "az": "us-east"},
			want:   "az:us-east, env:prod, group:web",
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := cli.FormatLabels(tc.labels)
			assert.Equal(suite.T(), tc.want, result)
		})
	}
}

func (suite *UITestSuite) TestBuildMutationTable() {
	errMsg := "interface not found"

	tests := []struct {
		name         string
		results      []cli.MutationResultRow
		fieldHeaders []string
		wantHeaders  []string
		wantRows     [][]string
	}{
		{
			name:         "when no results returns empty",
			results:      []cli.MutationResultRow{},
			fieldHeaders: nil,
			wantHeaders:  []string{"HOSTNAME", "STATUS", "ERROR"},
			wantRows:     [][]string{},
		},
		{
			name: "when all succeed shows ok status with empty error",
			results: []cli.MutationResultRow{
				{Hostname: "web-01", Status: "ok"},
				{Hostname: "web-02", Status: "ok"},
			},
			fieldHeaders: nil,
			wantHeaders:  []string{"HOSTNAME", "STATUS", "ERROR"},
			wantRows: [][]string{
				{"web-01", "ok", ""},
				{"web-02", "ok", ""},
			},
		},
		{
			name: "when some fail shows error message",
			results: []cli.MutationResultRow{
				{Hostname: "web-01", Status: "ok"},
				{Hostname: "web-02", Status: "failed", Error: &errMsg},
			},
			fieldHeaders: nil,
			wantHeaders:  []string{"HOSTNAME", "STATUS", "ERROR"},
			wantRows: [][]string{
				{"web-01", "ok", ""},
				{"web-02", "failed", "interface not found"},
			},
		},
		{
			name: "when field headers provided appends extra columns",
			results: []cli.MutationResultRow{
				{Hostname: "web-01", Status: "ok", Fields: []string{"extra"}},
			},
			fieldHeaders: []string{"DETAIL"},
			wantHeaders:  []string{"HOSTNAME", "STATUS", "ERROR", "DETAIL"},
			wantRows: [][]string{
				{"web-01", "ok", "", "extra"},
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			headers, rows := cli.BuildMutationTable(tc.results, tc.fieldHeaders)

			assert.Equal(suite.T(), tc.wantHeaders, headers)
			assert.Equal(suite.T(), tc.wantRows, rows)
		})
	}
}

func (suite *UITestSuite) TestFormatList() {
	tests := []struct {
		name string
		list []string
		want string
	}{
		{
			name: "when empty returns None",
			list: []string{},
			want: "None",
		},
		{
			name: "when single item returns it",
			list: []string{"alpha"},
			want: "alpha",
		},
		{
			name: "when multiple items joins with comma",
			list: []string{"alpha", "beta", "gamma"},
			want: "alpha, beta, gamma",
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			got := cli.FormatList(tc.list)

			assert.Equal(suite.T(), tc.want, got)
		})
	}
}

func (suite *UITestSuite) TestCalculateColumnWidths() {
	tests := []struct {
		name       string
		headers    []string
		rows       [][]string
		minPadding int
		want       []int
	}{
		{
			name:       "when empty headers returns empty",
			headers:    []string{},
			rows:       nil,
			minPadding: 1,
			want:       []int{},
		},
		{
			name:       "when headers wider than rows uses header width",
			headers:    []string{"HOSTNAME", "STATUS"},
			rows:       [][]string{{"a", "b"}},
			minPadding: 1,
			want:       []int{10, 8},
		},
		{
			name:       "when rows wider than headers uses row width",
			headers:    []string{"A", "B"},
			rows:       [][]string{{"longvalue", "anotherlongvalue"}},
			minPadding: 1,
			want:       []int{11, 18},
		},
		{
			name:       "when multi-line content uses longest line width",
			headers:    []string{"DATA"},
			rows:       [][]string{{"short\nvery long line here"}},
			minPadding: 0,
			want:       []int{19},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			got := cli.CalculateColumnWidths(tc.headers, tc.rows, tc.minPadding)

			assert.Equal(suite.T(), tc.want, got)
		})
	}
}

func (suite *UITestSuite) TestGetMaxLineWidth() {
	tests := []struct {
		name string
		text string
		want int
	}{
		{
			name: "when single line returns its length",
			text: "hello",
			want: 5,
		},
		{
			name: "when multi-line returns longest",
			text: "short\na much longer line\nmed",
			want: 18,
		},
		{
			name: "when empty returns zero",
			text: "",
			want: 0,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			got := cli.GetMaxLineWidth(tc.text)

			assert.Equal(suite.T(), tc.want, got)
		})
	}
}

func (suite *UITestSuite) TestSafeString() {
	str := "hello"

	tests := []struct {
		name string
		s    *string
		want string
	}{
		{
			name: "when non-nil returns value",
			s:    &str,
			want: "hello",
		},
		{
			name: "when nil returns empty",
			s:    nil,
			want: "",
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			got := cli.SafeString(tc.s)

			assert.Equal(suite.T(), tc.want, got)
		})
	}
}

func (suite *UITestSuite) TestSafeUUID() {
	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		name string
		u    *uuid.UUID
		want string
	}{
		{
			name: "when non-nil returns string",
			u:    &id,
			want: "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name: "when nil returns empty",
			u:    nil,
			want: "",
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			got := cli.SafeUUID(tc.u)

			assert.Equal(suite.T(), tc.want, got)
		})
	}
}

func (suite *UITestSuite) TestFloat64ToSafeString() {
	val := 3.14

	tests := []struct {
		name string
		f    *float64
		want string
	}{
		{
			name: "when non-nil returns formatted float",
			f:    &val,
			want: "3.140000",
		},
		{
			name: "when nil returns N/A",
			f:    nil,
			want: "N/A",
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			got := cli.Float64ToSafeString(tc.f)

			assert.Equal(suite.T(), tc.want, got)
		})
	}
}

func (suite *UITestSuite) TestIntToSafeString() {
	val := 42

	tests := []struct {
		name string
		i    *int
		want string
	}{
		{
			name: "when non-nil returns formatted int",
			i:    &val,
			want: "42",
		},
		{
			name: "when nil returns N/A",
			i:    nil,
			want: "N/A",
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			got := cli.IntToSafeString(tc.i)

			assert.Equal(suite.T(), tc.want, got)
		})
	}
}

func (suite *UITestSuite) TestHandleAuthError() {
	tests := []struct {
		name      string
		jsonError *gen.ErrorResponse
		code      int
		wantInLog string
	}{
		{
			name:      "when error response is nil logs unknown error",
			jsonError: nil,
			code:      401,
			wantInLog: "unknown error",
		},
		{
			name:      "when error field is nil logs unknown error",
			jsonError: &gen.ErrorResponse{Error: nil},
			code:      403,
			wantInLog: "unknown error",
		},
		{
			name: "when error response provided logs message",
			jsonError: func() *gen.ErrorResponse {
				msg := "insufficient permissions"
				return &gen.ErrorResponse{Error: &msg}
			}(),
			code:      403,
			wantInLog: "insufficient permissions",
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			var buf bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&buf, nil))

			cli.HandleAuthError(tc.jsonError, tc.code, logger)

			assert.Contains(suite.T(), buf.String(), tc.wantInLog)
		})
	}
}

func (suite *UITestSuite) TestHandleUnknownError() {
	tests := []struct {
		name      string
		jsonError *gen.ErrorResponse
		code      int
		wantInLog string
	}{
		{
			name:      "when error response is nil logs unknown error",
			jsonError: nil,
			code:      500,
			wantInLog: "unknown error",
		},
		{
			name:      "when error field is nil logs unknown error",
			jsonError: &gen.ErrorResponse{Error: nil},
			code:      500,
			wantInLog: "unknown error",
		},
		{
			name: "when error response provided logs message",
			jsonError: func() *gen.ErrorResponse {
				msg := "internal server error"
				return &gen.ErrorResponse{Error: &msg}
			}(),
			code:      500,
			wantInLog: "internal server error",
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			var buf bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&buf, nil))

			cli.HandleUnknownError(tc.jsonError, tc.code, logger)

			assert.Contains(suite.T(), buf.String(), tc.wantInLog)
		})
	}
}

func (suite *UITestSuite) TestPrintKV() {
	tests := []struct {
		name       string
		pairs      []string
		wantOutput bool
	}{
		{
			name:       "when valid pairs prints output",
			pairs:      []string{"Key", "Value"},
			wantOutput: true,
		},
		{
			name:       "when multiple pairs prints all",
			pairs:      []string{"Name", "test", "Status", "ok"},
			wantOutput: true,
		},
		{
			name:       "when odd number of pairs prints nothing",
			pairs:      []string{"Key"},
			wantOutput: false,
		},
		{
			name:       "when empty prints nothing",
			pairs:      []string{},
			wantOutput: false,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			output := captureStdout(func() {
				cli.PrintKV(tc.pairs...)
			})

			if tc.wantOutput {
				assert.NotEmpty(suite.T(), output)
			} else {
				assert.Empty(suite.T(), output)
			}
		})
	}
}

func (suite *UITestSuite) TestPrintStyledTable() {
	wideRow := []string{
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		"ccccccccccccccccccccccccccccccc",
		"ddddddddddddddddddddddddddddd",
		"eeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
	}

	tests := []struct {
		name     string
		sections []cli.Section
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
		},
		{
			name: "when section without title renders table",
			sections: []cli.Section{
				{
					Headers: []string{"COL1"},
					Rows:    [][]string{{"a"}},
				},
			},
		},
		{
			name: "when table exceeds terminal width scales columns",
			sections: []cli.Section{
				{
					Title:   "Wide",
					Headers: []string{"A", "B", "C", "D", "E"},
					Rows:    [][]string{wideRow},
				},
			},
		},
		{
			name: "when scaled column drops below minimum enforces floor",
			sections: []cli.Section{
				{
					Headers: []string{
						"X", "Y", "Z",
						"LONG-HEADER-1", "LONG-HEADER-2",
						"LONG-HEADER-3", "LONG-HEADER-4",
						"LONG-HEADER-5", "LONG-HEADER-6",
					},
					Rows: [][]string{{
						"a", "b", "c",
						"aaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbb",
						"cccccccccccccccccccc", "dddddddddddddddddddd",
						"eeeeeeeeeeeeeeeeeeee", "ffffffffffffffffffff",
					}},
				},
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			output := captureStdout(func() {
				cli.PrintStyledTable(tc.sections)
			})

			assert.NotEmpty(suite.T(), output)
		})
	}
}

func (suite *UITestSuite) TestDisplayJobDetailResponse() {
	tests := []struct {
		name string
		resp *gen.JobDetailResponse
	}{
		{
			name: "when minimal response displays job info",
			resp: func() *gen.JobDetailResponse {
				status := "completed"
				return &gen.JobDetailResponse{
					Status: &status,
				}
			}(),
		},
		{
			name: "when full response displays all sections",
			resp: func() *gen.JobDetailResponse {
				id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
				status := "completed"
				hostname := "web-01"
				created := "2026-01-01T00:00:00Z"
				updatedAt := "2026-01-01T00:01:00Z"
				errMsg := "timeout"
				operation := map[string]interface{}{"type": "system.hostname"}
				result := map[string]interface{}{"hostname": "web-01"}
				event := "completed"
				timestamp := "2026-01-01T00:01:00Z"
				message := "job completed"
				workerStatus := "completed"
				duration := "1.5s"
				respStatus := "ok"

				return &gen.JobDetailResponse{
					Id:        &id,
					Status:    &status,
					Hostname:  &hostname,
					Created:   &created,
					UpdatedAt: &updatedAt,
					Error:     &errMsg,
					Operation: &operation,
					Result:    result,
					Timeline: &[]struct {
						Error     *string `json:"error,omitempty"`
						Event     *string `json:"event,omitempty"`
						Hostname  *string `json:"hostname,omitempty"`
						Message   *string `json:"message,omitempty"`
						Timestamp *string `json:"timestamp,omitempty"`
					}{
						{
							Event:     &event,
							Timestamp: &timestamp,
							Hostname:  &hostname,
							Message:   &message,
						},
					},
					WorkerStates: &map[string]struct {
						Duration *string `json:"duration,omitempty"`
						Error    *string `json:"error,omitempty"`
						Status   *string `json:"status,omitempty"`
					}{
						"web-01": {
							Status:   &workerStatus,
							Duration: &duration,
						},
					},
					Responses: &map[string]struct {
						Data     interface{} `json:"data,omitempty"`
						Error    *string     `json:"error,omitempty"`
						Hostname *string     `json:"hostname,omitempty"`
						Status   *string     `json:"status,omitempty"`
					}{
						"web-01": {
							Status: &respStatus,
							Data:   map[string]string{"key": "val"},
						},
					},
				}
			}(),
		},
		{
			name: "when worker states with multiple workers shows summary",
			resp: func() *gen.JobDetailResponse {
				status := "completed"
				completed := "completed"
				failed := "failed"
				started := "started"
				duration := "1s"
				errMsg := "error"

				return &gen.JobDetailResponse{
					Status: &status,
					WorkerStates: &map[string]struct {
						Duration *string `json:"duration,omitempty"`
						Error    *string `json:"error,omitempty"`
						Status   *string `json:"status,omitempty"`
					}{
						"web-01": {Status: &completed, Duration: &duration},
						"web-02": {Status: &failed, Duration: &duration, Error: &errMsg},
						"web-03": {Status: &started, Duration: &duration},
					},
				}
			}(),
		},
		{
			name: "when response has nil data shows no data placeholder",
			resp: func() *gen.JobDetailResponse {
				status := "completed"
				respStatus := "ok"

				return &gen.JobDetailResponse{
					Status: &status,
					Responses: &map[string]struct {
						Data     interface{} `json:"data,omitempty"`
						Error    *string     `json:"error,omitempty"`
						Hostname *string     `json:"hostname,omitempty"`
						Status   *string     `json:"status,omitempty"`
					}{
						"web-01": {
							Status: &respStatus,
							Data:   nil,
						},
					},
				}
			}(),
		},
		{
			name: "when response has error shows error message",
			resp: func() *gen.JobDetailResponse {
				status := "failed"
				respStatus := "failed"
				errMsg := "timeout"

				return &gen.JobDetailResponse{
					Status: &status,
					Responses: &map[string]struct {
						Data     interface{} `json:"data,omitempty"`
						Error    *string     `json:"error,omitempty"`
						Hostname *string     `json:"hostname,omitempty"`
						Status   *string     `json:"status,omitempty"`
					}{
						"web-01": {
							Status: &respStatus,
							Error:  &errMsg,
						},
					},
				}
			}(),
		},
		{
			name: "when timeline has error shows error message",
			resp: func() *gen.JobDetailResponse {
				status := "failed"
				event := "failed"
				timestamp := "2026-01-01T00:01:00Z"
				errMsg := "connection refused"

				return &gen.JobDetailResponse{
					Status: &status,
					Timeline: &[]struct {
						Error     *string `json:"error,omitempty"`
						Event     *string `json:"event,omitempty"`
						Hostname  *string `json:"hostname,omitempty"`
						Message   *string `json:"message,omitempty"`
						Timestamp *string `json:"timestamp,omitempty"`
					}{
						{
							Event:     &event,
							Timestamp: &timestamp,
							Error:     &errMsg,
						},
					},
				}
			}(),
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			output := captureStdout(func() {
				cli.DisplayJobDetailResponse(tc.resp)
			})

			assert.NotEmpty(suite.T(), output)
		})
	}
}
