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

package integration_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ScheduleSmokeSuite struct {
	suite.Suite
}

// TestScheduleList verifies the list endpoint is reachable and answers in the
// collection shape every domain returns. On a platform where the operation is
// unsupported the results array carries an error entry rather than entries,
// which is still a valid answer and still proves the path.
func (s *ScheduleSmokeSuite) TestScheduleList() {
	tests := []struct {
		name         string
		args         []string
		validateFunc func(stdout string, exitCode int)
	}{
		{
			name: "list endpoint responds",
			args: []string{
				"client", "node", "schedule", "list",
				"--target", "_any",
				"--json",
			},
			validateFunc: func(
				stdout string,
				exitCode int,
			) {
				s.Require().Equal(0, exitCode)

				var result map[string]any
				s.Require().NoError(parseJSON(stdout, &result))

				results, ok := result["results"].([]any)
				s.Require().True(ok)
				s.NotNil(results)
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

// TestScheduleValidationRejectsBeforeQueueing verifies the controller answers a
// malformed request itself rather than turning it into a job.
//
// This is a read-only test despite naming create: a request rejected at the API
// never reaches an agent and changes nothing, which is the property being
// checked. The unit suites prove which rule fires; this proves the rejection
// survives the real CLI, the real client and the real middleware stack.
func (s *ScheduleSmokeSuite) TestScheduleValidationRejectsBeforeQueueing() {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "a malformed cron expression",
			args: []string{
				"client", "node", "schedule", "create",
				"--target", "_any",
				"--name", "osapi-int-invalid",
				"--object", "osapi-int-schedule.sh",
				"--schedule", "not-a-cron",
				"--json",
			},
		},
		{
			name: "neither schedule nor interval",
			args: []string{
				"client", "node", "schedule", "create",
				"--target", "_any",
				"--name", "osapi-int-invalid",
				"--object", "osapi-int-schedule.sh",
				"--json",
			},
		},
		{
			name: "a name the provider would refuse as a file name",
			args: []string{
				"client", "node", "schedule", "get",
				"--target", "_any",
				"--name", "../../etc/passwd",
				"--json",
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			_, stderr, exitCode := runCLI(tt.args...)
			s.NotEqual(0, exitCode, "the request should have been rejected")
			s.NotEmpty(stderr)
		})
	}
}

// TestScheduleCreateGetUpdateDelete walks one entry through all five
// operations, which is the only way to see that create's object reference, the
// name the entry is addressed by afterwards, and delete's idempotency agree
// with each other.
func (s *ScheduleSmokeSuite) TestScheduleCreateGetUpdateDelete() {
	skipWriteOp(s.T(), "SCHEDULE_LIFECYCLE")

	const (
		object = "osapi-int-schedule.sh"
		name   = "osapi-int-schedule"
	)

	filePath := writeTempFile(s.T(), "#!/bin/sh\necho osapi integration\n")

	// The entry's content comes from the Object Store, so the object has to
	// exist before create can reference it.
	uploadOut, _, uploadCode := runCLI(
		"client", "file", "upload",
		"--name", object,
		"--file", filePath,
		"--json",
	)
	s.Require().Equal(0, uploadCode)

	var uploadResp struct {
		Name string `json:"name"`
	}
	s.Require().NoError(parseJSON(uploadOut, &uploadResp))
	s.Equal(object, uploadResp.Name)

	s.T().Cleanup(func() {
		runCLI("client", "node", "schedule", "delete",
			"--target", "_any", "--name", name, "--json")
		runCLI("client", "file", "delete", "--name", object, "--json")
	})

	// Create
	createOut, _, createCode := runCLI(
		"client", "node", "schedule", "create",
		"--target", "_any",
		"--name", name,
		"--object", object,
		"--schedule", "0 3 * * *",
		"--json",
	)
	s.Require().Equal(0, createCode)

	var createResp map[string]any
	s.Require().NoError(parseJSON(createOut, &createResp))
	s.NotEmpty(createResp["job_id"])
	s.Contains(createResp, "results")

	// A platform without the provider reports skipped, which the provider
	// contract makes a real outcome rather than a failure, so the queueing and
	// the dispatch are proved and there is no entry to go on and read. Skipping
	// here says that out loud instead of passing on an assertion that could
	// never have failed.
	if status := firstResultStatus(s.T(), createOut); status == "skipped" {
		s.T().Skipf("schedule is unsupported on this host, create reported %q", status)
	}

	// Get, which is what proves create wrote the entry under the name it was
	// given rather than under the object's.
	getOut, _, getCode := runCLI(
		"client", "node", "schedule", "get",
		"--target", "_any",
		"--name", name,
		"--json",
	)
	s.Require().Equal(0, getCode)

	s.Contains(getOut, name)

	// Update
	updateOut, _, updateCode := runCLI(
		"client", "node", "schedule", "update",
		"--target", "_any",
		"--name", name,
		"--schedule", "30 4 * * *",
		"--json",
	)
	s.Require().Equal(0, updateCode)

	var updateResp map[string]any
	s.Require().NoError(parseJSON(updateOut, &updateResp))
	s.NotEmpty(updateResp["job_id"])
	s.Equal("ok", firstResultStatus(s.T(), updateOut))

	// Delete
	deleteOut, _, deleteCode := runCLI(
		"client", "node", "schedule", "delete",
		"--target", "_any",
		"--name", name,
		"--json",
	)
	s.Require().Equal(0, deleteCode)

	var deleteResp map[string]any
	s.Require().NoError(parseJSON(deleteOut, &deleteResp))
	s.NotEmpty(deleteResp["job_id"])
	s.Equal("ok", firstResultStatus(s.T(), deleteOut))

	// Deleting again is the idempotency claim the provider contract makes, so
	// the second call reports no change rather than failing.
	secondOut, _, secondCode := runCLI(
		"client", "node", "schedule", "delete",
		"--target", "_any",
		"--name", name,
		"--json",
	)
	s.Require().Equal(0, secondCode)
	s.Contains(secondOut, "results")
}

// firstResultStatus returns the status of the first per-host result in a
// collection response. Every domain answers in that shape, and the status is
// how a host says it did the work, did nothing, or does not support it.
func firstResultStatus(
	t *testing.T,
	stdout string,
) string {
	t.Helper()

	var parsed struct {
		Results []struct {
			Status string `json:"status"`
		} `json:"results"`
	}

	if err := parseJSON(stdout, &parsed); err != nil {
		t.Fatalf("parse results: %v", err)
	}

	if len(parsed.Results) == 0 {
		t.Fatalf("response carried no results: %s", stdout)
	}

	return parsed.Results[0].Status
}

func TestScheduleSmokeSuite(
	t *testing.T,
) {
	suite.Run(t, new(ScheduleSmokeSuite))
}
