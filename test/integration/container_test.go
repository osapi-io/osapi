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

type ContainerSmokeSuite struct {
	suite.Suite
}

func (s *ContainerSmokeSuite) TestContainerPull() {
	skipWriteOp(s.T(), "CONTAINER_PULL")

	tests := []struct {
		name         string
		args         []string
		validateFunc func(stdout string, exitCode int)
	}{
		{
			name: "pulls alpine image and returns image id",
			args: []string{
				"client", "container", "pull",
				"--image", "alpine:latest",
				"--json",
			},
			validateFunc: func(
				stdout string,
				exitCode int,
			) {
				s.Require().Equal(0, exitCode)

				var result map[string]any
				s.Require().NoError(parseJSON(stdout, &result))

				s.NotEmpty(result["job_id"])

				results, ok := result["results"].([]any)
				s.Require().True(ok)
				s.GreaterOrEqual(len(results), 1)

				first, ok := results[0].(map[string]any)
				s.Require().True(ok)
				s.NotEmpty(first["hostname"])
				s.NotEmpty(first["image_id"])
				s.Contains(first, "changed")
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

func (s *ContainerSmokeSuite) TestContainerCreateListInspectStopRemove() {
	skipWriteOp(s.T(), "CONTAINER_LIFECYCLE")

	// Pull the image first so create does not fail.
	pullOut, _, pullCode := runCLI(
		"client", "container", "pull",
		"--image", "alpine:latest",
		"--json",
	)
	s.Require().Equal(0, pullCode)

	var pullResp map[string]any
	s.Require().NoError(parseJSON(pullOut, &pullResp))
	s.NotEmpty(pullResp["job_id"])

	// Create
	createOut, _, createCode := runCLI(
		"client", "container", "create",
		"--image", "alpine:latest",
		"--name", "integration-test-container",
		"--auto-start=false",
		"--json",
	)
	s.Require().Equal(0, createCode)

	var createResp map[string]any
	s.Require().NoError(parseJSON(createOut, &createResp))
	s.NotEmpty(createResp["job_id"])

	createResults, ok := createResp["results"].([]any)
	s.Require().True(ok)
	s.GreaterOrEqual(len(createResults), 1)

	firstCreate, ok := createResults[0].(map[string]any)
	s.Require().True(ok)
	s.NotEmpty(firstCreate["id"])
	s.Contains(firstCreate, "changed")

	// List
	listOut, _, listCode := runCLI(
		"client", "container", "list",
		"--state", "all",
		"--json",
	)
	s.Require().Equal(0, listCode)

	var listResp map[string]any
	s.Require().NoError(parseJSON(listOut, &listResp))
	s.NotEmpty(listResp["job_id"])

	listResults, ok := listResp["results"].([]any)
	s.Require().True(ok)
	s.GreaterOrEqual(len(listResults), 1)

	firstList, ok := listResults[0].(map[string]any)
	s.Require().True(ok)
	s.Contains(firstList, "containers")

	// Inspect
	inspectOut, _, inspectCode := runCLI(
		"client", "container", "inspect",
		"--id", "integration-test-container",
		"--json",
	)
	s.Require().Equal(0, inspectCode)

	var inspectResp map[string]any
	s.Require().NoError(parseJSON(inspectOut, &inspectResp))
	s.NotEmpty(inspectResp["job_id"])

	inspectResults, ok := inspectResp["results"].([]any)
	s.Require().True(ok)
	s.GreaterOrEqual(len(inspectResults), 1)

	firstInspect, ok := inspectResults[0].(map[string]any)
	s.Require().True(ok)
	s.NotEmpty(firstInspect["id"])
	s.NotEmpty(firstInspect["image"])
	s.Contains(firstInspect, "state")

	// Remove (force, in case the container is running)
	removeOut, _, removeCode := runCLI(
		"client", "container", "remove",
		"--id", "integration-test-container",
		"--force",
		"--json",
	)
	s.Require().Equal(0, removeCode)

	var removeResp map[string]any
	s.Require().NoError(parseJSON(removeOut, &removeResp))
	s.NotEmpty(removeResp["job_id"])

	removeResults, ok := removeResp["results"].([]any)
	s.Require().True(ok)
	s.GreaterOrEqual(len(removeResults), 1)

	firstRemove, ok := removeResults[0].(map[string]any)
	s.Require().True(ok)
	s.NotEmpty(firstRemove["message"])
	s.Contains(firstRemove, "changed")
}

func (s *ContainerSmokeSuite) TestContainerList() {
	tests := []struct {
		name         string
		args         []string
		validateFunc func(stdout string, exitCode int)
	}{
		{
			name: "returns container list with results",
			args: []string{
				"client", "container", "list",
				"--state", "all",
				"--json",
			},
			validateFunc: func(
				stdout string,
				exitCode int,
			) {
				s.Require().Equal(0, exitCode)

				var result map[string]any
				s.Require().NoError(parseJSON(stdout, &result))
				s.NotEmpty(result["job_id"])
				s.Contains(result, "results")
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

func TestContainerSmokeSuite(
	t *testing.T,
) {
	suite.Run(t, new(ContainerSmokeSuite))
}
