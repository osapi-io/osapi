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
	Name     string `json:"name"`
	Schedule string `json:"schedule"`
	User     string `json:"user"`
	Command  string `json:"command"`
	Error    string `json:"error,omitempty"`
}

// CronMutationResult represents the result of a cron create/update/delete operation.
type CronMutationResult struct {
	JobID   string `json:"job_id"`
	Name    string `json:"name"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// CronCreateOpts contains options for creating a cron entry.
type CronCreateOpts struct {
	// Name is the cron drop-in entry name (required).
	Name string
	// Schedule is the cron expression (required).
	Schedule string
	// Command is the command to execute (required).
	Command string
	// User is the user to run the command as (optional, defaults to root).
	User string
}

// CronUpdateOpts contains options for updating a cron entry.
type CronUpdateOpts struct {
	// Schedule is the cron expression (optional).
	Schedule string
	// Command is the command to execute (optional).
	Command string
	// User is the user to run the command as (optional).
	User string
}

// cronEntryCollectionFromGen converts a gen.CronCollectionResponse
// to a Collection[CronEntryResult].
func cronEntryCollectionFromGen(
	g *gen.CronCollectionResponse,
) Collection[CronEntryResult] {
	results := make([]CronEntryResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, CronEntryResult{
			Name:     derefString(r.Name),
			Schedule: derefString(r.Schedule),
			User:     derefString(r.User),
			Command:  derefString(r.Command),
		})
	}

	return Collection[CronEntryResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// cronEntryFromGen converts a gen.CronEntryResponse to a CronEntryResult.
func cronEntryFromGen(
	g *gen.CronEntryResponse,
) CronEntryResult {
	return CronEntryResult{
		Name:     derefString(g.Name),
		Schedule: derefString(g.Schedule),
		User:     derefString(g.User),
		Command:  derefString(g.Command),
		Error:    derefString(g.Error),
	}
}

// cronMutationFromCreate converts a gen.CronCreateResponse to a CronMutationResult.
func cronMutationFromCreate(
	g *gen.CronCreateResponse,
) CronMutationResult {
	return CronMutationResult{
		JobID:   jobIDFromGen(g.JobId),
		Name:    derefString(g.Name),
		Changed: derefBool(g.Changed),
		Error:   derefString(g.Error),
	}
}

// cronMutationFromUpdate converts a gen.CronUpdateResponse to a CronMutationResult.
func cronMutationFromUpdate(
	g *gen.CronUpdateResponse,
) CronMutationResult {
	return CronMutationResult{
		JobID:   jobIDFromGen(g.JobId),
		Name:    derefString(g.Name),
		Changed: derefBool(g.Changed),
		Error:   derefString(g.Error),
	}
}

// cronMutationFromDelete converts a gen.CronDeleteResponse to a CronMutationResult.
func cronMutationFromDelete(
	g *gen.CronDeleteResponse,
) CronMutationResult {
	return CronMutationResult{
		JobID:   jobIDFromGen(g.JobId),
		Name:    derefString(g.Name),
		Changed: derefBool(g.Changed),
		Error:   derefString(g.Error),
	}
}
