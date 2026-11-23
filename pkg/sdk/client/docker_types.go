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
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/osapi-io/osapi/pkg/sdk/client/gen"
)

// DockerCreateOpts contains options for creating a container.
type DockerCreateOpts struct {
	// Name is the container's name. Required.
	Name string
	// Image is the image reference, such as alpine:3.20. Required.
	Image string
	// Hostname is the hostname inside the container.
	Hostname string
	// DNS is the DNS servers of the container.
	DNS []string
	// Command replaces the image's command.
	Command []string
	// Env is the container's environment. Its values go to docker on
	// standard input and are never returned.
	Env map[string]string
	// Ports is the port mappings, each host_port:container_port.
	Ports []string
	// Volumes is the bind mounts, each /host/path:/container/path.
	Volumes []string
	// AutoStart starts the container once it is made.
	AutoStart bool
}

// DockerExecOpts contains options for running a command in a container.
type DockerExecOpts struct {
	// Command is the command and its arguments. Required.
	Command []string
	// Env is the command's environment.
	Env map[string]string
	// WorkingDir is the absolute directory the command runs in.
	WorkingDir string
}

// DockerResult is one container on one host, from a list or a get. A host
// that failed or skipped has a row with its status and no container.
type DockerResult struct {
	Hostname string   `json:"hostname"`
	Status   string   `json:"status"`
	ID       string   `json:"id,omitempty"`
	Name     string   `json:"name,omitempty"`
	Image    string   `json:"image,omitempty"`
	State    string   `json:"state,omitempty"`
	Created  string   `json:"created,omitempty"`
	Health   string   `json:"health,omitempty"`
	Ports    []string `json:"ports,omitempty"`
	Mounts   []string `json:"mounts,omitempty"`
	Networks []string `json:"networks,omitempty"`
	EnvNames []string `json:"env_names,omitempty"`
	Error    string   `json:"error,omitempty"`
}

// DockerActionResult is the result of a create, a remove, a start or a stop
// on one host.
type DockerActionResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Name     string `json:"name,omitempty"`
	ID       string `json:"id,omitempty"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

// DockerExecResult is what a command in a container wrote, and its exit
// code, on one host.
type DockerExecResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Stdout   string `json:"stdout,omitempty"`
	Stderr   string `json:"stderr,omitempty"`
	ExitCode int    `json:"exit_code"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

// DockerImageResult is one image on one host, from an image list.
type DockerImageResult struct {
	Hostname   string `json:"hostname"`
	Status     string `json:"status"`
	ID         string `json:"id,omitempty"`
	Repository string `json:"repository,omitempty"`
	Tag        string `json:"tag,omitempty"`
	Size       string `json:"size,omitempty"`
	Error      string `json:"error,omitempty"`
}

// DockerPullResult is the image that a pull left on one host.
type DockerPullResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Image    string `json:"image,omitempty"`
	ImageID  string `json:"image_id,omitempty"`
	Size     int64  `json:"size,omitempty"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

// DockerImageRemoveResult is the result of an image remove on one host.
type DockerImageRemoveResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Image    string `json:"image,omitempty"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

// collect converts the rows of a response with conv.
func collect[G, R any](
	rows []G,
	jobID *openapi_types.UUID,
	conv func(G) R,
) Collection[R] {
	results := make([]R, 0, len(rows))
	for _, r := range rows {
		results = append(results, conv(r))
	}

	return Collection[R]{Results: results, JobID: jobIDFromGen(jobID)}
}

func dockerFromGen(
	r gen.ContainerDockerEntry,
) DockerResult {
	return DockerResult{
		Hostname: r.Hostname, Status: string(r.Status),
		ID: derefString(r.Id), Name: derefString(r.Name), Image: derefString(r.Image),
		State: derefString(r.State), Created: derefString(r.Created), Health: derefString(r.Health),
		Ports: derefStringSlice(r.Ports), Mounts: derefStringSlice(r.Mounts),
		Networks: derefStringSlice(r.Networks), EnvNames: derefStringSlice(r.EnvNames),
		Error: derefString(r.Error),
	}
}

func dockerActionFromGen(
	r gen.ContainerDockerResult,
) DockerActionResult {
	return DockerActionResult{
		Hostname: r.Hostname, Status: string(r.Status), Name: derefString(r.Name),
		ID: derefString(r.Id), Changed: derefBool(r.Changed), Error: derefString(r.Error),
	}
}

func dockerExecFromGen(
	r gen.ContainerDockerExecResult,
) DockerExecResult {
	return DockerExecResult{
		Hostname: r.Hostname, Status: string(r.Status), Stdout: derefString(r.Stdout),
		Stderr: derefString(r.Stderr), ExitCode: derefInt(r.ExitCode),
		Changed: derefBool(r.Changed), Error: derefString(r.Error),
	}
}

func dockerImageFromGen(
	r gen.ContainerImageEntry,
) DockerImageResult {
	return DockerImageResult{
		Hostname: r.Hostname, Status: string(r.Status), ID: derefString(r.Id),
		Repository: derefString(r.Repository), Tag: derefString(r.Tag),
		Size: derefString(r.Size), Error: derefString(r.Error),
	}
}

func dockerPullFromGen(
	r gen.ContainerImagePullResult,
) DockerPullResult {
	return DockerPullResult{
		Hostname: r.Hostname, Status: string(r.Status), Image: derefString(r.Image),
		ImageID: derefString(r.ImageId), Size: derefInt64(r.Size),
		Changed: derefBool(r.Changed), Error: derefString(r.Error),
	}
}

func dockerImageRemoveFromGen(
	r gen.ContainerImageResult,
) DockerImageRemoveResult {
	return DockerImageRemoveResult{
		Hostname: r.Hostname, Status: string(r.Status), Image: derefString(r.Image),
		Changed: derefBool(r.Changed), Error: derefString(r.Error),
	}
}
