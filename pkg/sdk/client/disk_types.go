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

// Disk represents disk usage information.
type Disk struct {
	Name  string `json:"name"`
	Total int    `json:"total"`
	Used  int    `json:"used"`
	Free  int    `json:"free"`
}

// DiskResult represents disk query result from a single agent.
type DiskResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
	Changed  bool   `json:"changed"`
	Disks    []Disk `json:"disks,omitempty"`
}

// disksFromGen converts a gen.DisksResponse to a slice of Disk.
func disksFromGen(
	g *gen.DisksResponse,
) []Disk {
	if g == nil {
		return nil
	}

	disks := make([]Disk, 0, len(*g))
	for _, d := range *g {
		disks = append(disks, Disk{
			Name:  d.Name,
			Total: d.Total,
			Used:  d.Used,
			Free:  d.Free,
		})
	}

	return disks
}

// diskCollectionFromGen converts a gen.DiskCollectionResponse to a Collection[DiskResult].
func diskCollectionFromGen(
	g *gen.DiskCollectionResponse,
) Collection[DiskResult] {
	results := make([]DiskResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, DiskResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Error:    derefString(r.Error),
			Changed:  derefBool(r.Changed),
			Disks:    disksFromGen(r.Disks),
		})
	}

	return Collection[DiskResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}
