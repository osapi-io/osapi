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

// NtpStatusResult represents NTP status from a query operation.
type NtpStatusResult struct {
	Hostname      string   `json:"hostname"`
	Status        string   `json:"status"`
	Synchronized  bool     `json:"synchronized,omitempty"`
	Stratum       int      `json:"stratum,omitempty"`
	Offset        string   `json:"offset,omitempty"`
	CurrentSource string   `json:"current_source,omitempty"`
	Servers       []string `json:"servers,omitempty"`
	Error         string   `json:"error,omitempty"`
}

// NtpMutationResult represents the result of an NTP create, update, or delete.
type NtpMutationResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

// NtpCreateOpts contains options for creating NTP configuration.
type NtpCreateOpts struct {
	// Servers is the list of NTP server addresses to configure. Required.
	Servers []string
}

// NtpUpdateOpts contains options for updating NTP configuration.
type NtpUpdateOpts struct {
	// Servers is the list of NTP server addresses to configure. Required.
	Servers []string
}

// ntpStatusCollectionFromGen converts a gen.NtpCollectionResponse
// to a Collection[NtpStatusResult].
func ntpStatusCollectionFromGen(
	g *gen.NtpCollectionResponse,
) Collection[NtpStatusResult] {
	results := make([]NtpStatusResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, NtpStatusResult{
			Hostname:      r.Hostname,
			Status:        string(r.Status),
			Synchronized:  derefBool(r.Synchronized),
			Stratum:       derefInt(r.Stratum),
			Offset:        derefString(r.Offset),
			CurrentSource: derefString(r.CurrentSource),
			Servers:       derefStringSlice(r.Servers),
			Error:         derefString(r.Error),
		})
	}

	return Collection[NtpStatusResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// ntpMutationCollectionFromCreate converts a gen.NtpCreateResponse
// to a Collection[NtpMutationResult].
func ntpMutationCollectionFromCreate(
	g *gen.NtpCreateResponse,
) Collection[NtpMutationResult] {
	results := make([]NtpMutationResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, NtpMutationResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Changed:  derefBool(r.Changed),
			Error:    derefString(r.Error),
		})
	}

	return Collection[NtpMutationResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// ntpMutationCollectionFromUpdate converts a gen.NtpUpdateResponse
// to a Collection[NtpMutationResult].
func ntpMutationCollectionFromUpdate(
	g *gen.NtpUpdateResponse,
) Collection[NtpMutationResult] {
	results := make([]NtpMutationResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, NtpMutationResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Changed:  derefBool(r.Changed),
			Error:    derefString(r.Error),
		})
	}

	return Collection[NtpMutationResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// ntpMutationCollectionFromDelete converts a gen.NtpDeleteResponse
// to a Collection[NtpMutationResult].
func ntpMutationCollectionFromDelete(
	g *gen.NtpDeleteResponse,
) Collection[NtpMutationResult] {
	results := make([]NtpMutationResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, NtpMutationResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Changed:  derefBool(r.Changed),
			Error:    derefString(r.Error),
		})
	}

	return Collection[NtpMutationResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// derefStringSlice returns the value pointed to by s, or nil if s is nil.
func derefStringSlice(
	s *[]string,
) []string {
	if s == nil {
		return nil
	}

	return *s
}
