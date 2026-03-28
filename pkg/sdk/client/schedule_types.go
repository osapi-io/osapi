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

// CronEntryResult represents a cron entry from a single agent.
type CronEntryResult struct {
	Hostname string `json:"hostname,omitempty"`
	Name     string `json:"name"`
	Object   string `json:"object,omitempty"`
	Schedule string `json:"schedule,omitempty"`
	Interval string `json:"interval,omitempty"`
	Source   string `json:"source,omitempty"`
	User     string `json:"user,omitempty"`
	Error    string `json:"error,omitempty"`
}

// CronMutationResult represents the result of a cron create/update/delete operation.
type CronMutationResult struct {
	Hostname string `json:"hostname,omitempty"`
	Name     string `json:"name"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

// CronCreateOpts contains options for creating a cron entry.
type CronCreateOpts struct {
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

// CronUpdateOpts contains options for updating a cron entry.
type CronUpdateOpts struct {
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

// cronEntryCollectionFromGen converts a gen.CronCollectionResponse
// to a Collection[CronEntryResult].
func cronEntryCollectionFromGen(
	g *gen.CronCollectionResponse,
) Collection[CronEntryResult] {
	results := make([]CronEntryResult, 0, len(g.Results))
	for _, r := range g.Results {
		var interval string
		if r.Interval != nil {
			interval = string(*r.Interval)
		}

		results = append(results, CronEntryResult{
			Hostname: r.Hostname,
			Name:     derefString(r.Name),
			Object:   derefString(r.Object),
			Schedule: derefString(r.Schedule),
			Interval: interval,
			Source:   derefString(r.Source),
			User:     derefString(r.User),
			Error:    derefString(r.Error),
		})
	}

	return Collection[CronEntryResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// cronGetCollectionFromGen converts a gen.CronGetResponse
// to a Collection[CronEntryResult].
func cronGetCollectionFromGen(
	g *gen.CronGetResponse,
) Collection[CronEntryResult] {
	results := make([]CronEntryResult, 0, len(g.Results))
	for _, r := range g.Results {
		var interval string
		if r.Interval != nil {
			interval = string(*r.Interval)
		}

		results = append(results, CronEntryResult{
			Hostname: r.Hostname,
			Name:     derefString(r.Name),
			Object:   derefString(r.Object),
			Schedule: derefString(r.Schedule),
			Interval: interval,
			Source:   derefString(r.Source),
			User:     derefString(r.User),
			Error:    derefString(r.Error),
		})
	}

	return Collection[CronEntryResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// cronMutationCollectionFromCreate converts a gen.CronCreateResponse
// to a Collection[CronMutationResult].
func cronMutationCollectionFromCreate(
	g *gen.CronCreateResponse,
) Collection[CronMutationResult] {
	results := make([]CronMutationResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, CronMutationResult{
			Hostname: r.Hostname,
			Name:     derefString(r.Name),
			Changed:  derefBool(r.Changed),
			Error:    derefString(r.Error),
		})
	}

	return Collection[CronMutationResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// cronMutationCollectionFromUpdate converts a gen.CronUpdateResponse
// to a Collection[CronMutationResult].
func cronMutationCollectionFromUpdate(
	g *gen.CronUpdateResponse,
) Collection[CronMutationResult] {
	results := make([]CronMutationResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, CronMutationResult{
			Hostname: r.Hostname,
			Name:     derefString(r.Name),
			Changed:  derefBool(r.Changed),
			Error:    derefString(r.Error),
		})
	}

	return Collection[CronMutationResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// cronMutationCollectionFromDelete converts a gen.CronDeleteResponse
// to a Collection[CronMutationResult].
func cronMutationCollectionFromDelete(
	g *gen.CronDeleteResponse,
) Collection[CronMutationResult] {
	results := make([]CronMutationResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, CronMutationResult{
			Hostname: r.Hostname,
			Name:     derefString(r.Name),
			Changed:  derefBool(r.Changed),
			Error:    derefString(r.Error),
		})
	}

	return Collection[CronMutationResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}
