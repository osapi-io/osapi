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

package status

import (
	"github.com/osapi-io/osapi/internal/behaviors/disk"
	"github.com/osapi-io/osapi/internal/behaviors/load"
	"github.com/osapi-io/osapi/internal/behaviors/memory"
	osb "github.com/osapi-io/osapi/internal/behaviors/os"
	"github.com/osapi-io/osapi/internal/behaviors/uptime"
)

// GetInput takes nothing: the status is of the whole host.
type GetInput struct{}

// Entry is the status summary of one host: what the hostname, uptime, os,
// load, memory and disk behaviors read.
type Entry struct {
	Name        string        `json:"name"             doc:"The hostname of the host, as the hostname command prints it"`
	Uptime      uptime.Entry  `json:"uptime"           doc:"How long the host has run"`
	OSInfo      osb.Release   `json:"os_info"          doc:"The operating system"`
	LoadAverage load.Averages `json:"load_average"     doc:"The load averages of the host"`
	Memory      memory.Usage  `json:"memory"           doc:"The memory of the host"`
	Disks       []disk.Disk   `json:"disks"            doc:"Each local filesystem"`
	Errors      []string      `json:"errors,omitempty" doc:"Each read that failed, as name: error. The other fields still answer."`
}
