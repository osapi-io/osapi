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
	"context"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/behaviors/disk"
	"github.com/osapi-io/osapi/internal/behaviors/hostname"
	"github.com/osapi-io/osapi/internal/behaviors/load"
	"github.com/osapi-io/osapi/internal/behaviors/memory"
	osb "github.com/osapi-io/osapi/internal/behaviors/os"
	"github.com/osapi-io/osapi/internal/behaviors/uptime"
)

// reader reads the status with the providers of the behaviors it
// requires, each built for the same host.
type reader struct {
	hostname hostname.Provider
	uptime   uptime.Provider
	os       osb.Provider
	load     load.Provider
	memory   memory.Provider
	disk     disk.Provider
}

func newReader(
	h host.Host,
) (Provider, error) {
	return &reader{
		hostname: host.Require[hostname.Provider](h, hostname.Behavior),
		uptime:   host.Require[uptime.Provider](h, uptime.Behavior),
		os:       host.Require[osb.Provider](h, osb.Behavior),
		load:     host.Require[load.Provider](h, load.Behavior),
		memory:   host.Require[memory.Provider](h, memory.Behavior),
		disk:     host.Require[disk.Provider](h, disk.Behavior),
	}, nil
}

// Status runs the get of each behavior. A read that fails is in Errors,
// and the others still answer: one failure is not a reason to answer
// nothing.
func (r *reader) Status(
	ctx context.Context,
) (Entry, error) {
	var e Entry
	note := func(name string, err error) {
		if err != nil {
			e.Errors = append(e.Errors, name+": "+err.Error())
		}
	}
	name, err := behavior.Call(
		ctx,
		hostname.Behavior,
		hostname.Get,
		r.hostname,
		hostname.GetInput{},
	)
	e.Name = name.Name
	note("hostname", err)
	up, err := behavior.Call(ctx, uptime.Behavior, uptime.Get, r.uptime, uptime.GetInput{})
	e.Uptime = up
	note("uptime", err)
	release, err := behavior.Call(ctx, osb.Behavior, osb.Get, r.os, osb.GetInput{})
	e.OSInfo = release.OSInfo
	note("os", err)
	avg, err := behavior.Call(ctx, load.Behavior, load.Get, r.load, load.GetInput{})
	e.LoadAverage = avg.LoadAverage
	note("load", err)
	mem, err := behavior.Call(ctx, memory.Behavior, memory.Get, r.memory, memory.GetInput{})
	e.Memory = mem.Memory
	note("memory", err)
	disks, err := behavior.Call(ctx, disk.Behavior, disk.Get, r.disk, disk.GetInput{})
	e.Disks = disks.Disks
	note("disk", err)

	return e, nil
}
