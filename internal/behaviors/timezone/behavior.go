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

// Package timezone is the timezone behavior: the timezone of the host.
package timezone

import (
	"strings"

	"github.com/osapi-io/osapi/internal/area"
	"github.com/osapi-io/osapi/internal/behavior"
)

// The timedatectl query of the update's current timezone.
const show = "timedatectl show -p Timezone --value"

// Behavior is the timezone of the host. gohai reads it on every platform;
// timedatectl sets it on the Debian family.
var Behavior = behavior.Register(behavior.Define(
	"timezone", "The timezone of the host",
	behavior.Node,
	behavior.API,
	behavior.Area(area.Config),
	behavior.CLI,
	behavior.Platform(behavior.PlatformLinux, newReader),
	behavior.Platform(behavior.PlatformDarwin, newReader),
	behavior.Read("get", Get).
		Describe("Get the timezone of the host").
		Explain("Reads the IANA name of the timezone and its offset from UTC now: the "+
			"zone that /etc/localtime links to, or /etc/timezone on Linux.").
		Example("A Linux host on UTC", GetInput{}, Entry{Timezone: "UTC", UTCOffset: "+00:00"},
			behavior.OnPlatform(behavior.PlatformLinux),
			behavior.Exists("/etc/timezone", "UTC\n"),
		).
		Example("A Mac on UTC", GetInput{}, Entry{Timezone: "UTC", UTCOffset: "+00:00"},
			behavior.OnPlatform(behavior.PlatformDarwin),
			behavior.Links("/etc/localtime", "/var/db/timezone/zoneinfo/UTC"),
		),
	behavior.Write("update", Update).
		On(behavior.PlatformDebian).
		Runs("timedatectl").
		Sudo("timedatectl").
		Describe("Set the timezone of the host").
		Explain("Sets the timezone with timedatectl set-timezone. A host that already "+
			"has the timezone does not change and reports changed: false. It runs on "+
			"the Debian family only, and elsewhere reports skipped.").
		Example("Set a new timezone", UpdateInput{Timezone: "America/New_York"},
			Result{Timezone: "America/New_York", Changed: true},
			behavior.Prints(show, "UTC\n"),
			behavior.PrintsAfter("timedatectl set-timezone America/New_York", show,
				"America/New_York\n"),
			behavior.Ran("timedatectl", "set-timezone", "America/New_York"),
			behavior.Reads(GetInput{}, Entry{Timezone: "America/New_York"}),
		).
		// After sets the timezone first, as a live host can run it: the
		// integration test runs these on CI's runner and puts its timezone
		// back.
		Example("A host that already has the timezone", UpdateInput{Timezone: "UTC"},
			Result{Timezone: "UTC"},
			behavior.After(UpdateInput{Timezone: "UTC"}),
			behavior.Prints(show, "America/New_York\n"),
			behavior.PrintsAfter("timedatectl set-timezone UTC", show, "UTC\n"),
			behavior.Reads(GetInput{}, Entry{Timezone: "UTC"}),
		).
		Fails("A name that is not a timezone", UpdateInput{Timezone: "../../etc/passwd"},
			"not a valid timezone name").
		Fails("A name that starts with a dash", UpdateInput{Timezone: "-UTC"},
			"not a valid timezone name").
		Fails("A name longer than 64 characters",
			UpdateInput{Timezone: strings.Repeat("a", 65)}, "timezone"),
))
