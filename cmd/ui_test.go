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

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UITestSuite struct {
	suite.Suite
}

func TestUITestSuite(t *testing.T) {
	suite.Run(t, new(UITestSuite))
}

func (suite *UITestSuite) TestBuildBroadcastTable() {
	errMsg := "connection refused"

	tests := []struct {
		name         string
		results      []resultRow
		fieldHeaders []string
		wantHeaders  []string
		wantRows     [][]string
	}{
		{
			name:         "when no results returns empty",
			results:      []resultRow{},
			fieldHeaders: nil,
			wantHeaders:  []string{"HOSTNAME"},
			wantRows:     [][]string{},
		},
		{
			name: "when all results succeed omits status and error columns",
			results: []resultRow{
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
			results: []resultRow{
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
			results: []resultRow{
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
			results: []resultRow{
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
			headers, rows := buildBroadcastTable(tc.results, tc.fieldHeaders)

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
			result := formatLabels(tc.labels)
			assert.Equal(suite.T(), tc.want, result)
		})
	}
}

func (suite *UITestSuite) TestBuildMutationTable() {
	errMsg := "interface not found"

	tests := []struct {
		name         string
		results      []mutationResultRow
		fieldHeaders []string
		wantHeaders  []string
		wantRows     [][]string
	}{
		{
			name:         "when no results returns empty",
			results:      []mutationResultRow{},
			fieldHeaders: nil,
			wantHeaders:  []string{"HOSTNAME", "STATUS", "ERROR"},
			wantRows:     [][]string{},
		},
		{
			name: "when all succeed shows ok status with empty error",
			results: []mutationResultRow{
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
			results: []mutationResultRow{
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
			results: []mutationResultRow{
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
			headers, rows := buildMutationTable(tc.results, tc.fieldHeaders)

			assert.Equal(suite.T(), tc.wantHeaders, headers)
			assert.Equal(suite.T(), tc.wantRows, rows)
		})
	}
}
