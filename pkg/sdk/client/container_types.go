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

// ContainerResult represents a container create result from a single agent.
type ContainerResult struct {
	Hostname string
	ID       string
	Name     string
	Image    string
	State    string
	Created  string
	Changed  bool
	Error    string
}

// ContainerListResult represents a container list result from a single agent.
type ContainerListResult struct {
	Hostname   string
	Containers []ContainerSummaryItem
	Changed    bool
	Error      string
}

// ContainerSummaryItem represents a brief container summary.
type ContainerSummaryItem struct {
	ID      string
	Name    string
	Image   string
	State   string
	Created string
}

// ContainerDetailResult represents a container inspect result from a single agent.
type ContainerDetailResult struct {
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

// ContainerActionResult represents a container lifecycle action result from a single agent.
type ContainerActionResult struct {
	Hostname string
	ID       string
	Message  string
	Changed  bool
	Error    string
}

// ContainerExecResult represents a container exec result from a single agent.
type ContainerExecResult struct {
	Hostname string
	Stdout   string
	Stderr   string
	ExitCode int
	Changed  bool
	Error    string
}

// ContainerPullResult represents an image pull result from a single agent.
type ContainerPullResult struct {
	Hostname string
	ImageID  string
	Tag      string
	Size     int64
	Changed  bool
	Error    string
}

// containerResultCollectionFromGen converts a gen.ContainerResultCollectionResponse
// to a Collection[ContainerResult].
func containerResultCollectionFromGen(
	g *gen.ContainerResultCollectionResponse,
) Collection[ContainerResult] {
	results := make([]ContainerResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, ContainerResult{
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

	return Collection[ContainerResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// containerListCollectionFromGen converts a gen.ContainerListCollectionResponse
// to a Collection[ContainerListResult].
func containerListCollectionFromGen(
	g *gen.ContainerListCollectionResponse,
) Collection[ContainerListResult] {
	results := make([]ContainerListResult, 0, len(g.Results))
	for _, r := range g.Results {
		item := ContainerListResult{
			Hostname: r.Hostname,
			Changed:  derefBool(r.Changed),
			Error:    derefString(r.Error),
		}

		if r.Containers != nil {
			containers := make([]ContainerSummaryItem, 0, len(*r.Containers))
			for _, c := range *r.Containers {
				containers = append(containers, ContainerSummaryItem{
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

	return Collection[ContainerListResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// containerDetailCollectionFromGen converts a gen.ContainerDetailCollectionResponse
// to a Collection[ContainerDetailResult].
func containerDetailCollectionFromGen(
	g *gen.ContainerDetailCollectionResponse,
) Collection[ContainerDetailResult] {
	results := make([]ContainerDetailResult, 0, len(g.Results))
	for _, r := range g.Results {
		item := ContainerDetailResult{
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

	return Collection[ContainerDetailResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// containerActionCollectionFromGen converts a gen.ContainerActionCollectionResponse
// to a Collection[ContainerActionResult].
func containerActionCollectionFromGen(
	g *gen.ContainerActionCollectionResponse,
) Collection[ContainerActionResult] {
	results := make([]ContainerActionResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, ContainerActionResult{
			Hostname: r.Hostname,
			ID:       derefString(r.Id),
			Message:  derefString(r.Message),
			Changed:  derefBool(r.Changed),
			Error:    derefString(r.Error),
		})
	}

	return Collection[ContainerActionResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// containerExecCollectionFromGen converts a gen.ContainerExecCollectionResponse
// to a Collection[ContainerExecResult].
func containerExecCollectionFromGen(
	g *gen.ContainerExecCollectionResponse,
) Collection[ContainerExecResult] {
	results := make([]ContainerExecResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, ContainerExecResult{
			Hostname: r.Hostname,
			Stdout:   derefString(r.Stdout),
			Stderr:   derefString(r.Stderr),
			ExitCode: derefInt(r.ExitCode),
			Changed:  derefBool(r.Changed),
			Error:    derefString(r.Error),
		})
	}

	return Collection[ContainerExecResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// containerPullCollectionFromGen converts a gen.ContainerPullCollectionResponse
// to a Collection[ContainerPullResult].
func containerPullCollectionFromGen(
	g *gen.ContainerPullCollectionResponse,
) Collection[ContainerPullResult] {
	results := make([]ContainerPullResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, ContainerPullResult{
			Hostname: r.Hostname,
			ImageID:  derefString(r.ImageId),
			Tag:      derefString(r.Tag),
			Size:     derefInt64(r.Size),
			Changed:  derefBool(r.Changed),
			Error:    derefString(r.Error),
		})
	}

	return Collection[ContainerPullResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}
