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
	"github.com/retr0h/osapi/pkg/sdk/client/gen"
)

// DockerCreateOpts contains options for creating a container.
type DockerCreateOpts struct {
	// Image is the container image reference (required).
	Image string
	// Name is an optional container name.
	Name string
	// Hostname sets the container hostname.
	Hostname string
	// DNS sets custom DNS servers for the container.
	DNS []string
	// Command overrides the image's default command.
	Command []string
	// Env is environment variables in KEY=VALUE format.
	Env []string
	// Ports is port mappings in host_port:container_port format.
	Ports []string
	// Volumes is volume mounts in host_path:container_path format.
	Volumes []string
	// AutoStart starts the container after creation (default true).
	AutoStart *bool
}

// DockerStopOpts contains options for stopping a container.
type DockerStopOpts struct {
	// Timeout is seconds to wait before killing. Zero uses default.
	Timeout int
}

// DockerListParams contains parameters for listing containers.
type DockerListParams struct {
	// State filters by state: "running", "stopped", "all".
	State string
	// Limit caps the number of results.
	Limit int
}

// DockerRemoveParams contains parameters for removing a container.
type DockerRemoveParams struct {
	// Force forces removal of a running container.
	Force bool
}

// DockerPullOpts contains options for pulling an image.
type DockerPullOpts struct {
	// Image is the image reference to pull (required).
	Image string
}

// DockerImageRemoveParams contains parameters for removing an image.
type DockerImageRemoveParams struct {
	// Force forces removal even if the image is in use.
	Force bool
}

// DockerExecOpts contains options for executing a command in a container.
type DockerExecOpts struct {
	// Command is the command and arguments to execute (required).
	Command []string
	// Env is additional environment variables in KEY=VALUE format.
	Env []string
	// WorkingDir is the working directory inside the container.
	WorkingDir string
}

// DockerResult represents a docker container create result from a single agent.
type DockerResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	ID       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Image    string `json:"image,omitempty"`
	State    string `json:"state,omitempty"`
	Created  string `json:"created,omitempty"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

// DockerListResult represents a docker container list result from a single agent.
type DockerListResult struct {
	Hostname   string              `json:"hostname"`
	Status     string              `json:"status"`
	Containers []DockerSummaryItem `json:"containers,omitempty"`
	Changed    bool                `json:"changed"`
	Error      string              `json:"error,omitempty"`
}

// DockerSummaryItem represents a brief docker container summary.
type DockerSummaryItem struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Image   string `json:"image"`
	State   string `json:"state"`
	Created string `json:"created"`
}

// DockerDetailResult represents a docker container inspect result from a single agent.
type DockerDetailResult struct {
	Hostname        string            `json:"hostname"`
	Status          string            `json:"status"`
	ID              string            `json:"id,omitempty"`
	Name            string            `json:"name,omitempty"`
	Image           string            `json:"image,omitempty"`
	State           string            `json:"state,omitempty"`
	Created         string            `json:"created,omitempty"`
	Ports           []string          `json:"ports,omitempty"`
	Mounts          []string          `json:"mounts,omitempty"`
	Env             []string          `json:"env,omitempty"`
	NetworkSettings map[string]string `json:"network_settings,omitempty"`
	Health          string            `json:"health,omitempty"`
	Changed         bool              `json:"changed"`
	Error           string            `json:"error,omitempty"`
}

// DockerActionResult represents a docker container lifecycle action result from a single agent.
type DockerActionResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	ID       string `json:"id,omitempty"`
	Message  string `json:"message,omitempty"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

// DockerExecResult represents a docker container exec result from a single agent.
type DockerExecResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Stdout   string `json:"stdout,omitempty"`
	Stderr   string `json:"stderr,omitempty"`
	ExitCode int    `json:"exit_code"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

// DockerPullResult represents a docker image pull result from a single agent.
type DockerPullResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	ImageID  string `json:"image_id,omitempty"`
	Tag      string `json:"tag,omitempty"`
	Size     int64  `json:"size"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

// dockerResultCollectionFromGen converts a gen.DockerResultCollectionResponse
// to a Collection[DockerResult].
func dockerResultCollectionFromGen(
	g *gen.DockerResultCollectionResponse,
) Collection[DockerResult] {
	results := make([]DockerResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, DockerResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			ID:       derefString(r.Id),
			Name:     derefString(r.Name),
			Image:    derefString(r.Image),
			State:    derefString(r.State),
			Created:  derefString(r.Created),
			Changed:  derefBool(r.Changed),
			Error:    derefString(r.Error),
		})
	}

	return Collection[DockerResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// dockerListCollectionFromGen converts a gen.DockerListCollectionResponse
// to a Collection[DockerListResult].
func dockerListCollectionFromGen(
	g *gen.DockerListCollectionResponse,
) Collection[DockerListResult] {
	results := make([]DockerListResult, 0, len(g.Results))
	for _, r := range g.Results {
		item := DockerListResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Changed:  derefBool(r.Changed),
			Error:    derefString(r.Error),
		}

		if r.Containers != nil {
			containers := make([]DockerSummaryItem, 0, len(*r.Containers))
			for _, c := range *r.Containers {
				containers = append(containers, DockerSummaryItem{
					ID:      derefString(c.Id),
					Name:    derefString(c.Name),
					Image:   derefString(c.Image),
					State:   derefString(c.State),
					Created: derefString(c.Created),
				})
			}

			item.Containers = containers
		}

		results = append(results, item)
	}

	return Collection[DockerListResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// dockerDetailCollectionFromGen converts a gen.DockerDetailCollectionResponse
// to a Collection[DockerDetailResult].
func dockerDetailCollectionFromGen(
	g *gen.DockerDetailCollectionResponse,
) Collection[DockerDetailResult] {
	results := make([]DockerDetailResult, 0, len(g.Results))
	for _, r := range g.Results {
		item := DockerDetailResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			ID:       derefString(r.Id),
			Name:     derefString(r.Name),
			Image:    derefString(r.Image),
			State:    derefString(r.State),
			Created:  derefString(r.Created),
			Health:   derefString(r.Health),
			Changed:  derefBool(r.Changed),
			Error:    derefString(r.Error),
		}

		if r.Ports != nil {
			item.Ports = *r.Ports
		}

		if r.Mounts != nil {
			item.Mounts = *r.Mounts
		}

		if r.Env != nil {
			item.Env = *r.Env
		}

		if r.NetworkSettings != nil {
			item.NetworkSettings = *r.NetworkSettings
		}

		results = append(results, item)
	}

	return Collection[DockerDetailResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// dockerActionCollectionFromGen converts a gen.DockerActionCollectionResponse
// to a Collection[DockerActionResult].
func dockerActionCollectionFromGen(
	g *gen.DockerActionCollectionResponse,
) Collection[DockerActionResult] {
	results := make([]DockerActionResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, DockerActionResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			ID:       derefString(r.Id),
			Message:  derefString(r.Message),
			Changed:  derefBool(r.Changed),
			Error:    derefString(r.Error),
		})
	}

	return Collection[DockerActionResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// dockerExecCollectionFromGen converts a gen.DockerExecCollectionResponse
// to a Collection[DockerExecResult].
func dockerExecCollectionFromGen(
	g *gen.DockerExecCollectionResponse,
) Collection[DockerExecResult] {
	results := make([]DockerExecResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, DockerExecResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Stdout:   derefString(r.Stdout),
			Stderr:   derefString(r.Stderr),
			ExitCode: derefInt(r.ExitCode),
			Changed:  derefBool(r.Changed),
			Error:    derefString(r.Error),
		})
	}

	return Collection[DockerExecResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// dockerPullCollectionFromGen converts a gen.DockerPullCollectionResponse
// to a Collection[DockerPullResult].
func dockerPullCollectionFromGen(
	g *gen.DockerPullCollectionResponse,
) Collection[DockerPullResult] {
	results := make([]DockerPullResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, DockerPullResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			ImageID:  derefString(r.ImageId),
			Tag:      derefString(r.Tag),
			Size:     derefInt64(r.Size),
			Changed:  derefBool(r.Changed),
			Error:    derefString(r.Error),
		})
	}

	return Collection[DockerPullResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}
