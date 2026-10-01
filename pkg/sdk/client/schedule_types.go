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
	"github.com/osapi-io/osapi/pkg/sdk/client/gen"
)

// ScheduleEntryResult represents a cron entry from a single agent.
type ScheduleEntryResult struct {
	Hostname string `json:"hostname,omitempty"`
	Status   string `json:"status"`
	Name     string `json:"name"`
	Object   string `json:"object,omitempty"`
	Schedule string `json:"schedule,omitempty"`
	Interval string `json:"interval,omitempty"`
	Source   string `json:"source,omitempty"`
	User     string `json:"user,omitempty"`
	Error    string `json:"error,omitempty"`
}

// ScheduleMutationResult represents the result of a cron create/update/delete operation.
type ScheduleMutationResult struct {
	Hostname string `json:"hostname,omitempty"`
	Status   string `json:"status"`
	Name     string `json:"name"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

// ScheduleCreateOpts contains options for creating a cron entry.
type ScheduleCreateOpts struct {
	// Name is the cron drop-in entry name (required).
	Name string
	// Object is the name of the uploaded file in the object store (required).
	Object string
	// Schedule is the cron expression (mutually exclusive with Interval).
	Schedule string
	// Interval is the periodic interval: hourly, daily, weekly, monthly
	// (mutually exclusive with Schedule).
	Interval string
	// User is the user to run the command as (optional).
	User string
	// ContentType is "raw" or "template" (optional, defaults to raw).
	ContentType string
	// Vars contains template variables (optional).
	Vars map[string]any
}

// ScheduleUpdateOpts contains options for updating a cron entry.
type ScheduleUpdateOpts struct {
	// Object is the new object to deploy (optional).
	Object string
	// Schedule is the cron expression (optional).
	Schedule string
	// User is the user to run the command as (optional).
	User string
	// ContentType is "raw" or "template" (optional).
	ContentType string
	// Vars contains template variables (optional).
	Vars map[string]any
}

// cronEntryCollectionFromGen converts a gen.ScheduleCollectionResponse
// to a Collection[ScheduleEntryResult].
func cronEntryCollectionFromGen(
	g *gen.ScheduleCollectionResponse,
) Collection[ScheduleEntryResult] {
	results := make([]ScheduleEntryResult, 0, len(g.Results))
	for _, r := range g.Results {
		var interval string
		if r.Interval != nil {
			interval = string(*r.Interval)
		}

		results = append(results, ScheduleEntryResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Name:     derefString(r.Name),
			Object:   derefString(r.Object),
			Schedule: derefString(r.Schedule),
			Interval: interval,
			Source:   derefString(r.Source),
			User:     derefString(r.User),
			Error:    derefString(r.Error),
		})
	}

	return Collection[ScheduleEntryResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// cronGetCollectionFromGen converts a gen.ScheduleGetResponse
// to a Collection[ScheduleEntryResult].
func cronGetCollectionFromGen(
	g *gen.ScheduleGetResponse,
) Collection[ScheduleEntryResult] {
	results := make([]ScheduleEntryResult, 0, len(g.Results))
	for _, r := range g.Results {
		var interval string
		if r.Interval != nil {
			interval = string(*r.Interval)
		}

		results = append(results, ScheduleEntryResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Name:     derefString(r.Name),
			Object:   derefString(r.Object),
			Schedule: derefString(r.Schedule),
			Interval: interval,
			Source:   derefString(r.Source),
			User:     derefString(r.User),
			Error:    derefString(r.Error),
		})
	}

	return Collection[ScheduleEntryResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// cronMutationCollectionFromCreate converts a gen.ScheduleCreateResponse
// to a Collection[ScheduleMutationResult].
func cronMutationCollectionFromCreate(
	g *gen.ScheduleCreateResponse,
) Collection[ScheduleMutationResult] {
	results := make([]ScheduleMutationResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, ScheduleMutationResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Name:     derefString(r.Name),
			Changed:  derefBool(r.Changed),
			Error:    derefString(r.Error),
		})
	}

	return Collection[ScheduleMutationResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// cronMutationCollectionFromUpdate converts a gen.ScheduleUpdateResponse
// to a Collection[ScheduleMutationResult].
func cronMutationCollectionFromUpdate(
	g *gen.ScheduleUpdateResponse,
) Collection[ScheduleMutationResult] {
	results := make([]ScheduleMutationResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, ScheduleMutationResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Name:     derefString(r.Name),
			Changed:  derefBool(r.Changed),
			Error:    derefString(r.Error),
		})
	}

	return Collection[ScheduleMutationResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// cronMutationCollectionFromDelete converts a gen.ScheduleDeleteResponse
// to a Collection[ScheduleMutationResult].
func cronMutationCollectionFromDelete(
	g *gen.ScheduleDeleteResponse,
) Collection[ScheduleMutationResult] {
	results := make([]ScheduleMutationResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, ScheduleMutationResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Name:     derefString(r.Name),
			Changed:  derefBool(r.Changed),
			Error:    derefString(r.Error),
		})
	}

	return Collection[ScheduleMutationResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}
