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

// DockerResult represents a docker container create result from a single agent.
type DockerResult struct {
	Hostname string
	ID       string
	Name     string
	Image    string
	State    string
	Created  string
	Changed  bool
	Error    string
}

// DockerListResult represents a docker container list result from a single agent.
type DockerListResult struct {
	Hostname   string
	Containers []DockerSummaryItem
	Changed    bool
	Error      string
}

// DockerSummaryItem represents a brief docker container summary.
type DockerSummaryItem struct {
	ID      string
	Name    string
	Image   string
	State   string
	Created string
}

// DockerDetailResult represents a docker container inspect result from a single agent.
type DockerDetailResult struct {
	Hostname        string
	ID              string
	Name            string
	Image           string
	State           string
	Created         string
	Ports           []string
	Mounts          []string
	Env             []string
	NetworkSettings map[string]string
	Health          string
	Changed         bool
	Error           string
}

// DockerActionResult represents a docker container lifecycle action result from a single agent.
type DockerActionResult struct {
	Hostname string
	ID       string
	Message  string
	Changed  bool
	Error    string
}

// DockerExecResult represents a docker container exec result from a single agent.
type DockerExecResult struct {
	Hostname string
	Stdout   string
	Stderr   string
	ExitCode int
	Changed  bool
	Error    string
}

// DockerPullResult represents a docker image pull result from a single agent.
type DockerPullResult struct {
	Hostname string
	ImageID  string
	Tag      string
	Size     int64
	Changed  bool
	Error    string
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
