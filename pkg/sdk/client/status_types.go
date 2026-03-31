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

// NodeStatus represents full node status from a single agent.
type NodeStatus struct {
	Hostname    string       `json:"hostname"`
	Status      string       `json:"status"`
	Uptime      string       `json:"uptime,omitempty"`
	Error       string       `json:"error,omitempty"`
	Changed     bool         `json:"changed"`
	Disks       []Disk       `json:"disks,omitempty"`
	LoadAverage *LoadAverage `json:"load_average,omitempty"`
	Memory      *Memory      `json:"memory,omitempty"`
	OSInfo      *OSInfo      `json:"os_info,omitempty"`
}

// nodeStatusCollectionFromGen converts a gen.NodeStatusCollectionResponse to a Collection[NodeStatus].
func nodeStatusCollectionFromGen(
	g *gen.NodeStatusCollectionResponse,
) Collection[NodeStatus] {
	results := make([]NodeStatus, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, NodeStatus{
			Hostname:    r.Hostname,
			Status:      string(r.Status),
			Uptime:      derefString(r.Uptime),
			Error:       derefString(r.Error),
			Changed:     derefBool(r.Changed),
			Disks:       disksFromGen(r.Disks),
			LoadAverage: loadAverageFromGen(r.LoadAverage),
			Memory:      memoryFromGen(r.Memory),
			OSInfo:      osInfoFromGen(r.OsInfo),
		})
	}

	return Collection[NodeStatus]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}
