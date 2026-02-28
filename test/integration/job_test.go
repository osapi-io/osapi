//go:build integration

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

package integration_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type JobSmokeSuite struct {
	suite.Suite
}

func (s *JobSmokeSuite) TestJobList() {
	tests := []struct {
		name         string
		args         []string
		validateFunc func(stdout string, exitCode int)
	}{
		{
			name: "returns jobs list",
			args: []string{"client", "job", "list", "--json"},
			validateFunc: func(
				stdout string,
				exitCode int,
			) {
				s.Require().Equal(0, exitCode)

				var result map[string]any
				s.Require().NoError(parseJSON(stdout, &result))
				s.Contains(result, "jobs")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			stdout, _, exitCode := runCLI(tt.args...)
			tt.validateFunc(stdout, exitCode)
		})
	}
}

func (s *JobSmokeSuite) TestJobGet() {
	triggerOut, _, triggerCode := runCLI(
		"client", "node", "command", "shell",
		"--command", "echo job-test",
		"--json",
	)
	s.Require().Equal(0, triggerCode)

	var triggerResp struct {
		JobID string `json:"job_id"`
	}
	s.Require().NoError(parseJSON(triggerOut, &triggerResp))
	s.Require().NotEmpty(triggerResp.JobID, "shell command must return a job_id")

	jobID := triggerResp.JobID

	tests := []struct {
		name         string
		args         []string
		validateFunc func(stdout string, exitCode int)
	}{
		{
			name: "returns job details for known job id",
			args: []string{"client", "job", "get", "--job-id", jobID, "--json"},
			validateFunc: func(
				stdout string,
				exitCode int,
			) {
				s.Require().Equal(0, exitCode)

				var result struct {
					ID string `json:"id"`
				}
				s.Require().NoError(parseJSON(stdout, &result))
				s.Equal(jobID, result.ID)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			stdout, _, exitCode := runCLI(tt.args...)
			tt.validateFunc(stdout, exitCode)
		})
	}
}

func (s *JobSmokeSuite) TestJobStatus() {
	tests := []struct {
		name         string
		args         []string
		validateFunc func(stdout string, exitCode int)
	}{
		{
			name: "returns queue stats with total_jobs",
			args: []string{"client", "job", "status", "--json"},
			validateFunc: func(
				stdout string,
				exitCode int,
			) {
				s.Require().Equal(0, exitCode)

				var result map[string]any
				s.Require().NoError(parseJSON(stdout, &result))
				s.Contains(result, "total_jobs")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			stdout, _, exitCode := runCLI(tt.args...)
			tt.validateFunc(stdout, exitCode)
		})
	}
}

func (s *JobSmokeSuite) TestJobDelete() {
	skipWrite(s.T())

	triggerOut, _, triggerCode := runCLI(
		"client", "node", "command", "shell",
		"--command", "echo delete-test",
		"--json",
	)
	s.Require().Equal(0, triggerCode)

	var triggerResp struct {
		JobID string `json:"job_id"`
	}
	s.Require().NoError(parseJSON(triggerOut, &triggerResp))
	s.Require().NotEmpty(triggerResp.JobID)

	tests := []struct {
		name         string
		args         []string
		validateFunc func(stdout string, exitCode int)
	}{
		{
			name: "deletes a job by id",
			args: []string{"client", "job", "delete", "--job-id", triggerResp.JobID, "--json"},
			validateFunc: func(
				stdout string,
				exitCode int,
			) {
				s.Require().Equal(0, exitCode)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			stdout, _, exitCode := runCLI(tt.args...)
			tt.validateFunc(stdout, exitCode)
		})
	}
}

func (s *JobSmokeSuite) TestJobAdd() {
	skipWrite(s.T())

	jobFile := writeJobFile(s.T(), map[string]any{
		"operation": "node.hostname",
		"target":    "_any",
	})

	tests := []struct {
		name         string
		args         []string
		validateFunc func(stdout string, exitCode int)
	}{
		{
			name: "adds a job to the queue",
			args: []string{"client", "job", "add", "--json-file", jobFile, "--json"},
			validateFunc: func(
				stdout string,
				exitCode int,
			) {
				s.Require().Equal(0, exitCode)

				var result struct {
					JobID  string `json:"job_id"`
					Status string `json:"status"`
				}
				s.Require().NoError(parseJSON(stdout, &result))
				s.NotEmpty(result.JobID)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			stdout, _, exitCode := runCLI(tt.args...)
			tt.validateFunc(stdout, exitCode)
		})
	}
}

func (s *JobSmokeSuite) TestJobRun() {
	skipWrite(s.T())

	jobFile := writeJobFile(s.T(), map[string]any{
		"operation": "node.hostname",
		"target":    "_any",
	})

	tests := []struct {
		name         string
		args         []string
		validateFunc func(stdout string, exitCode int)
	}{
		{
			name: "submits and polls job to completion",
			args: []string{
				"client", "job", "run",
				"--json-file", jobFile,
				"--timeout", "30",
				"--poll-interval", "1",
				"--json",
			},
			validateFunc: func(
				stdout string,
				exitCode int,
			) {
				s.Require().Equal(0, exitCode)

				var result struct {
					ID     string  `json:"id"`
					Status *string `json:"status"`
				}
				s.Require().NoError(parseJSON(stdout, &result))
				s.NotEmpty(result.ID)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			stdout, _, exitCode := runCLI(tt.args...)
			tt.validateFunc(stdout, exitCode)
		})
	}
}

func (s *JobSmokeSuite) TestJobRetry() {
	skipWrite(s.T())

	triggerOut, _, triggerCode := runCLI(
		"client", "node", "command", "shell",
		"--command", "echo retry-test",
		"--json",
	)
	s.Require().Equal(0, triggerCode)

	var triggerResp struct {
		JobID string `json:"job_id"`
	}
	s.Require().NoError(parseJSON(triggerOut, &triggerResp))
	s.Require().NotEmpty(triggerResp.JobID)

	tests := []struct {
		name         string
		args         []string
		validateFunc func(stdout string, exitCode int)
	}{
		{
			name: "retries a job by id",
			args: []string{"client", "job", "retry", "--job-id", triggerResp.JobID, "--json"},
			validateFunc: func(
				stdout string,
				exitCode int,
			) {
				s.Require().Equal(0, exitCode)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			stdout, _, exitCode := runCLI(tt.args...)
			tt.validateFunc(stdout, exitCode)
		})
	}
}

func TestJobSmokeSuite(
	t *testing.T,
) {
	suite.Run(t, new(JobSmokeSuite))
}
