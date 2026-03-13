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

package client

import (
	"testing"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/pkg/sdk/client/gen"
)

type DockerTypesTestSuite struct {
	suite.Suite
}

func (suite *DockerTypesTestSuite) TestDockerResultCollectionFromGen() {
	testUUID := openapi_types.UUID{
		0x55, 0x0e, 0x84, 0x00,
		0xe2, 0x9b, 0x41, 0xd4,
		0xa7, 0x16, 0x44, 0x66,
		0x55, 0x44, 0x00, 0x00,
	}

	tests := []struct {
		name         string
		input        *gen.DockerResultCollectionResponse
		validateFunc func(Collection[DockerResult])
	}{
		{
			name: "when all fields are populated",
			input: func() *gen.DockerResultCollectionResponse {
				id := "abc123"
				name := "my-nginx"
				image := "nginx:latest"
				state := "running"
				created := "2026-01-01T00:00:00Z"
				changed := true

				return &gen.DockerResultCollectionResponse{
					JobId: &testUUID,
					Results: []gen.DockerResponse{
						{
							Hostname: "web-01",
							Id:       &id,
							Name:     &name,
							Image:    &image,
							State:    &state,
							Created:  &created,
							Changed:  &changed,
						},
					},
				}
			}(),
			validateFunc: func(c Collection[DockerResult]) {
				suite.Equal("550e8400-e29b-41d4-a716-446655440000", c.JobID)
				suite.Require().Len(c.Results, 1)

				r := c.Results[0]
				suite.Equal("web-01", r.Hostname)
				suite.Equal("abc123", r.ID)
				suite.Equal("my-nginx", r.Name)
				suite.Equal("nginx:latest", r.Image)
				suite.Equal("running", r.State)
				suite.Equal("2026-01-01T00:00:00Z", r.Created)
				suite.True(r.Changed)
				suite.Empty(r.Error)
			},
		},
		{
			name: "when minimal with error",
			input: func() *gen.DockerResultCollectionResponse {
				errMsg := "image not found"

				return &gen.DockerResultCollectionResponse{
					Results: []gen.DockerResponse{
						{
							Hostname: "web-01",
							Error:    &errMsg,
						},
					},
				}
			}(),
			validateFunc: func(c Collection[DockerResult]) {
				suite.Empty(c.JobID)
				suite.Require().Len(c.Results, 1)

				r := c.Results[0]
				suite.Equal("web-01", r.Hostname)
				suite.Equal("image not found", r.Error)
				suite.Empty(r.ID)
				suite.Empty(r.Name)
				suite.Empty(r.Image)
				suite.Empty(r.State)
				suite.Empty(r.Created)
				suite.False(r.Changed)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := dockerResultCollectionFromGen(tc.input)
			tc.validateFunc(result)
		})
	}
}

func (suite *DockerTypesTestSuite) TestDockerListCollectionFromGen() {
	testUUID := openapi_types.UUID{
		0x55, 0x0e, 0x84, 0x00,
		0xe2, 0x9b, 0x41, 0xd4,
		0xa7, 0x16, 0x44, 0x66,
		0x55, 0x44, 0x00, 0x00,
	}

	tests := []struct {
		name         string
		input        *gen.DockerListCollectionResponse
		validateFunc func(Collection[DockerListResult])
	}{
		{
			name: "when containers are populated",
			input: func() *gen.DockerListCollectionResponse {
				changed := false
				id := "abc123"
				name := "my-nginx"
				image := "nginx:latest"
				state := "running"
				created := "2026-01-01T00:00:00Z"
				containers := []gen.DockerSummary{
					{
						Id:      &id,
						Name:    &name,
						Image:   &image,
						State:   &state,
						Created: &created,
					},
				}

				return &gen.DockerListCollectionResponse{
					JobId: &testUUID,
					Results: []gen.DockerListItem{
						{
							Hostname:   "web-01",
							Changed:    &changed,
							Containers: &containers,
						},
					},
				}
			}(),
			validateFunc: func(c Collection[DockerListResult]) {
				suite.Equal("550e8400-e29b-41d4-a716-446655440000", c.JobID)
				suite.Require().Len(c.Results, 1)

				r := c.Results[0]
				suite.Equal("web-01", r.Hostname)
				suite.False(r.Changed)
				suite.Empty(r.Error)
				suite.Require().Len(r.Containers, 1)
				suite.Equal("abc123", r.Containers[0].ID)
				suite.Equal("my-nginx", r.Containers[0].Name)
				suite.Equal("nginx:latest", r.Containers[0].Image)
				suite.Equal("running", r.Containers[0].State)
				suite.Equal("2026-01-01T00:00:00Z", r.Containers[0].Created)
			},
		},
		{
			name: "when containers is nil",
			input: &gen.DockerListCollectionResponse{
				Results: []gen.DockerListItem{
					{Hostname: "web-01"},
				},
			},
			validateFunc: func(c Collection[DockerListResult]) {
				suite.Empty(c.JobID)
				suite.Require().Len(c.Results, 1)
				suite.Equal("web-01", c.Results[0].Hostname)
				suite.False(c.Results[0].Changed)
				suite.Nil(c.Results[0].Containers)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := dockerListCollectionFromGen(tc.input)
			tc.validateFunc(result)
		})
	}
}

func (suite *DockerTypesTestSuite) TestDockerDetailCollectionFromGen() {
	testUUID := openapi_types.UUID{
		0x55, 0x0e, 0x84, 0x00,
		0xe2, 0x9b, 0x41, 0xd4,
		0xa7, 0x16, 0x44, 0x66,
		0x55, 0x44, 0x00, 0x00,
	}

	tests := []struct {
		name         string
		input        *gen.DockerDetailCollectionResponse
		validateFunc func(Collection[DockerDetailResult])
	}{
		{
			name: "when all fields are populated",
			input: func() *gen.DockerDetailCollectionResponse {
				id := "abc123"
				name := "my-nginx"
				image := "nginx:latest"
				state := "running"
				created := "2026-01-01T00:00:00Z"
				health := "healthy"
				changed := false
				ports := []string{"80/tcp", "443/tcp"}
				mounts := []string{"/data:/data"}
				env := []string{"FOO=bar", "BAZ=qux"}
				networkSettings := map[string]string{"ip": "172.17.0.2"}

				return &gen.DockerDetailCollectionResponse{
					JobId: &testUUID,
					Results: []gen.DockerDetailResponse{
						{
							Hostname:        "web-01",
							Id:              &id,
							Name:            &name,
							Image:           &image,
							State:           &state,
							Created:         &created,
							Health:          &health,
							Changed:         &changed,
							Ports:           &ports,
							Mounts:          &mounts,
							Env:             &env,
							NetworkSettings: &networkSettings,
						},
					},
				}
			}(),
			validateFunc: func(c Collection[DockerDetailResult]) {
				suite.Equal("550e8400-e29b-41d4-a716-446655440000", c.JobID)
				suite.Require().Len(c.Results, 1)

				r := c.Results[0]
				suite.Equal("web-01", r.Hostname)
				suite.Equal("abc123", r.ID)
				suite.Equal("my-nginx", r.Name)
				suite.Equal("nginx:latest", r.Image)
				suite.Equal("running", r.State)
				suite.Equal("2026-01-01T00:00:00Z", r.Created)
				suite.Equal("healthy", r.Health)
				suite.False(r.Changed)
				suite.Empty(r.Error)
				suite.Equal([]string{"80/tcp", "443/tcp"}, r.Ports)
				suite.Equal([]string{"/data:/data"}, r.Mounts)
				suite.Equal([]string{"FOO=bar", "BAZ=qux"}, r.Env)
				suite.Equal(map[string]string{"ip": "172.17.0.2"}, r.NetworkSettings)
			},
		},
		{
			name: "when optional fields are nil",
			input: &gen.DockerDetailCollectionResponse{
				Results: []gen.DockerDetailResponse{
					{Hostname: "web-01"},
				},
			},
			validateFunc: func(c Collection[DockerDetailResult]) {
				suite.Empty(c.JobID)
				suite.Require().Len(c.Results, 1)

				r := c.Results[0]
				suite.Equal("web-01", r.Hostname)
				suite.Empty(r.ID)
				suite.Empty(r.Name)
				suite.Empty(r.Image)
				suite.Empty(r.State)
				suite.Empty(r.Created)
				suite.Empty(r.Health)
				suite.False(r.Changed)
				suite.Empty(r.Error)
				suite.Nil(r.Ports)
				suite.Nil(r.Mounts)
				suite.Nil(r.Env)
				suite.Nil(r.NetworkSettings)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := dockerDetailCollectionFromGen(tc.input)
			tc.validateFunc(result)
		})
	}
}

func (suite *DockerTypesTestSuite) TestDockerActionCollectionFromGen() {
	testUUID := openapi_types.UUID{
		0x55, 0x0e, 0x84, 0x00,
		0xe2, 0x9b, 0x41, 0xd4,
		0xa7, 0x16, 0x44, 0x66,
		0x55, 0x44, 0x00, 0x00,
	}

	tests := []struct {
		name         string
		input        *gen.DockerActionCollectionResponse
		validateFunc func(Collection[DockerActionResult])
	}{
		{
			name: "when all fields are populated",
			input: func() *gen.DockerActionCollectionResponse {
				id := "abc123"
				message := "container started"
				changed := true

				return &gen.DockerActionCollectionResponse{
					JobId: &testUUID,
					Results: []gen.DockerActionResultItem{
						{
							Hostname: "web-01",
							Id:       &id,
							Message:  &message,
							Changed:  &changed,
						},
					},
				}
			}(),
			validateFunc: func(c Collection[DockerActionResult]) {
				suite.Equal("550e8400-e29b-41d4-a716-446655440000", c.JobID)
				suite.Require().Len(c.Results, 1)

				r := c.Results[0]
				suite.Equal("web-01", r.Hostname)
				suite.Equal("abc123", r.ID)
				suite.Equal("container started", r.Message)
				suite.True(r.Changed)
				suite.Empty(r.Error)
			},
		},
		{
			name: "when minimal with error",
			input: func() *gen.DockerActionCollectionResponse {
				errMsg := "container not found"

				return &gen.DockerActionCollectionResponse{
					Results: []gen.DockerActionResultItem{
						{
							Hostname: "web-01",
							Error:    &errMsg,
						},
					},
				}
			}(),
			validateFunc: func(c Collection[DockerActionResult]) {
				suite.Empty(c.JobID)
				suite.Require().Len(c.Results, 1)

				r := c.Results[0]
				suite.Equal("web-01", r.Hostname)
				suite.Equal("container not found", r.Error)
				suite.Empty(r.ID)
				suite.Empty(r.Message)
				suite.False(r.Changed)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := dockerActionCollectionFromGen(tc.input)
			tc.validateFunc(result)
		})
	}
}

func (suite *DockerTypesTestSuite) TestDockerExecCollectionFromGen() {
	testUUID := openapi_types.UUID{
		0x55, 0x0e, 0x84, 0x00,
		0xe2, 0x9b, 0x41, 0xd4,
		0xa7, 0x16, 0x44, 0x66,
		0x55, 0x44, 0x00, 0x00,
	}

	tests := []struct {
		name         string
		input        *gen.DockerExecCollectionResponse
		validateFunc func(Collection[DockerExecResult])
	}{
		{
			name: "when all fields are populated",
			input: func() *gen.DockerExecCollectionResponse {
				stdout := "hello world\n"
				stderr := "warning: something\n"
				exitCode := 0
				changed := true

				return &gen.DockerExecCollectionResponse{
					JobId: &testUUID,
					Results: []gen.DockerExecResultItem{
						{
							Hostname: "web-01",
							Stdout:   &stdout,
							Stderr:   &stderr,
							ExitCode: &exitCode,
							Changed:  &changed,
						},
					},
				}
			}(),
			validateFunc: func(c Collection[DockerExecResult]) {
				suite.Equal("550e8400-e29b-41d4-a716-446655440000", c.JobID)
				suite.Require().Len(c.Results, 1)

				r := c.Results[0]
				suite.Equal("web-01", r.Hostname)
				suite.Equal("hello world\n", r.Stdout)
				suite.Equal("warning: something\n", r.Stderr)
				suite.Equal(0, r.ExitCode)
				suite.True(r.Changed)
				suite.Empty(r.Error)
			},
		},
		{
			name: "when minimal with error",
			input: func() *gen.DockerExecCollectionResponse {
				errMsg := "exec failed"
				exitCode := 1

				return &gen.DockerExecCollectionResponse{
					Results: []gen.DockerExecResultItem{
						{
							Hostname: "web-01",
							Error:    &errMsg,
							ExitCode: &exitCode,
						},
					},
				}
			}(),
			validateFunc: func(c Collection[DockerExecResult]) {
				suite.Empty(c.JobID)
				suite.Require().Len(c.Results, 1)

				r := c.Results[0]
				suite.Equal("web-01", r.Hostname)
				suite.Equal("exec failed", r.Error)
				suite.Equal(1, r.ExitCode)
				suite.Empty(r.Stdout)
				suite.Empty(r.Stderr)
				suite.False(r.Changed)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := dockerExecCollectionFromGen(tc.input)
			tc.validateFunc(result)
		})
	}
}

func (suite *DockerTypesTestSuite) TestDockerPullCollectionFromGen() {
	testUUID := openapi_types.UUID{
		0x55, 0x0e, 0x84, 0x00,
		0xe2, 0x9b, 0x41, 0xd4,
		0xa7, 0x16, 0x44, 0x66,
		0x55, 0x44, 0x00, 0x00,
	}

	tests := []struct {
		name         string
		input        *gen.DockerPullCollectionResponse
		validateFunc func(Collection[DockerPullResult])
	}{
		{
			name: "when all fields are populated",
			input: func() *gen.DockerPullCollectionResponse {
				imageID := "sha256:abc123"
				tag := "latest"
				size := int64(52428800)
				changed := true

				return &gen.DockerPullCollectionResponse{
					JobId: &testUUID,
					Results: []gen.DockerPullResultItem{
						{
							Hostname: "web-01",
							ImageId:  &imageID,
							Tag:      &tag,
							Size:     &size,
							Changed:  &changed,
						},
					},
				}
			}(),
			validateFunc: func(c Collection[DockerPullResult]) {
				suite.Equal("550e8400-e29b-41d4-a716-446655440000", c.JobID)
				suite.Require().Len(c.Results, 1)

				r := c.Results[0]
				suite.Equal("web-01", r.Hostname)
				suite.Equal("sha256:abc123", r.ImageID)
				suite.Equal("latest", r.Tag)
				suite.Equal(int64(52428800), r.Size)
				suite.True(r.Changed)
				suite.Empty(r.Error)
			},
		},
		{
			name: "when minimal with error",
			input: func() *gen.DockerPullCollectionResponse {
				errMsg := "pull failed: image not found"

				return &gen.DockerPullCollectionResponse{
					Results: []gen.DockerPullResultItem{
						{
							Hostname: "web-01",
							Error:    &errMsg,
						},
					},
				}
			}(),
			validateFunc: func(c Collection[DockerPullResult]) {
				suite.Empty(c.JobID)
				suite.Require().Len(c.Results, 1)

				r := c.Results[0]
				suite.Equal("web-01", r.Hostname)
				suite.Equal("pull failed: image not found", r.Error)
				suite.Empty(r.ImageID)
				suite.Empty(r.Tag)
				suite.Zero(r.Size)
				suite.False(r.Changed)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := dockerPullCollectionFromGen(tc.input)
			tc.validateFunc(result)
		})
	}
}

func TestDockerTypesTestSuite(t *testing.T) {
	suite.Run(t, new(DockerTypesTestSuite))
}
