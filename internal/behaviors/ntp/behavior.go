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

// Package ntp is the ntp behavior: the NTP servers that chrony uses, and
// its sync status.
package ntp

import (
	"regexp"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behaviors/managedfile"
	"github.com/osapi-io/osapi/internal/validation/rule"
)

// serverPattern is an NTP server as chrony's server directive takes it: a
// hostname, an IPv4 address or an IPv6 address. A space or a line break is
// not one, so a server cannot add a directive of its own to the file.
const serverPattern = `^[A-Za-z0-9][A-Za-z0-9.:-]{0,252}$`

var serverRE = regexp.MustCompile(serverPattern)

// The chronyc output of a host that is synchronized to time.cloudflare.com.
const (
	tracking = "Reference ID    : A29FC801 (time.cloudflare.com)\n" +
		"Stratum         : 3\n" +
		"System time     : 0.000003422 seconds fast of NTP time\n" +
		"Leap status     : Normal\n"
	sources = "^,*,time.cloudflare.com,3,6,377,12,0.000001,0.000002,0.001\n" +
		"^,+,0.pool.ntp.org,2,6,377,13,0.000004,0.000005,0.002\n"
)

// Behavior is the ntp definition. chronyc reads the sync status, which is
// the live state of chrony and not a fact of the host; the servers that
// osapi sets go to one file under /etc/chrony/sources.d through
// managedfile.
var Behavior = behavior.Register(behavior.Define(
	"ntp", "The NTP servers of chrony, and its sync status",
	behavior.Node,
	behavior.API,
	behavior.Area(behavior.AreaConfig),
	behavior.CLI,
	behavior.Requires(managedfile.Behavior),
	behavior.Platform(behavior.PlatformDebian, newDebian),
	behavior.Runs("chronyc"),
	behavior.Sudo("chronyc"),
	behavior.Writes(sourcesFile),
	behavior.Validator("ntp_server", serverRE.MatchString,
		func(t rule.Target) { t.Pattern(serverPattern) },
		"is not a hostname or an IP address"),
	behavior.Read("get", Get).
		NotAFact("the sync status and the sources are the live state of chrony, not a fact of the host").
		Describe("Get the sync status of chrony and its sources").
		Explain("Reads chronyc tracking, for whether chrony is synchronized, its stratum, "+
			"its offset from NTP time and the source that it follows, and chronyc "+
			"sources, for each source that it has.").
		Example("A host synchronized to time.cloudflare.com", GetInput{},
			Status{
				Synchronized:  true,
				Stratum:       3,
				Offset:        "+0.000003422s",
				CurrentSource: "time.cloudflare.com",
				Servers:       []string{"time.cloudflare.com", "0.pool.ntp.org"},
			},
			behavior.Prints("chronyc tracking", tracking),
			behavior.Prints("chronyc sources -c", sources),
		).
		Example("A host that chrony does not synchronize", GetInput{},
			Status{Offset: "-0.000001834s"},
			behavior.Prints("chronyc tracking",
				"System time     : 0.000001834 seconds slow of NTP time\n"+
					"Leap status     : Not synchronised\n"),
			behavior.Prints("chronyc sources -c", ""),
		),
	behavior.Write("create", Create).
		Describe("Set the NTP servers that chrony uses").
		Explain("Writes a server line for each server to "+sourcesFile+", and reloads "+
			"the sources with chronyc reload sources. A file that has the same servers "+
			"does not change, and reports changed: false.").
		Example("Set two servers", CreateInput{Servers: []string{"0.pool.ntp.org", "1.pool.ntp.org"}},
			Result{Changed: true},
			behavior.Wrote(sourcesFile, "server 0.pool.ntp.org iburst\nserver 1.pool.ntp.org iburst\n"),
			behavior.Ran("chronyc", "reload", "sources"),
		).
		Example("The servers that the host has", CreateInput{Servers: []string{"0.pool.ntp.org"}},
			Result{},
			behavior.After(CreateInput{Servers: []string{"0.pool.ntp.org"}}),
		).
		Fails("No servers", CreateInput{}, "required").
		Fails("An empty list of servers", CreateInput{Servers: []string{}}, "min").
		Fails("A server with a directive after it",
			CreateInput{Servers: []string{"0.pool.ntp.org iburst\nallow all"}},
			"is not a hostname or an IP address").
		Fails("More than 16 servers", CreateInput{Servers: make17()}, "max"),
	behavior.Write("update", Update).
		Describe("Replace the NTP servers that osapi set").
		Explain("The same as create, for servers that osapi set already. It fails with "+
			"not found when osapi set none. The servers that the file has already "+
			"report changed: false.").
		Example("Replace the servers", UpdateInput{Servers: []string{"time.cloudflare.com"}},
			Result{Changed: true},
			behavior.After(CreateInput{Servers: []string{"0.pool.ntp.org"}}),
			behavior.Wrote(sourcesFile, "server time.cloudflare.com iburst\n"),
			behavior.Ran("chronyc", "reload", "sources"),
		).
		Example("The servers that the file has", UpdateInput{Servers: []string{"0.pool.ntp.org"}},
			Result{},
			behavior.After(CreateInput{Servers: []string{"0.pool.ntp.org"}}),
		).
		Fails("When osapi set no servers", UpdateInput{Servers: []string{"0.pool.ntp.org"}},
			"osapi manages no servers").
		Fails("No servers", UpdateInput{}, "required").
		Fails("An empty list of servers", UpdateInput{Servers: []string{}}, "min").
		Fails("A server with a directive after it",
			UpdateInput{Servers: []string{"0.pool.ntp.org iburst\nallow all"}},
			"is not a hostname or an IP address").
		Fails("More than 16 servers", UpdateInput{Servers: make17()}, "max"),
	behavior.Write("delete", Delete).
		Describe("Remove the NTP servers that osapi set").
		Explain("Removes "+sourcesFile+" and reloads the sources, so chrony uses its "+
			"other sources. When osapi set no servers, it reports changed: false.").
		Example("Remove the servers", DeleteInput{}, Result{Changed: true},
			behavior.After(CreateInput{Servers: []string{"0.pool.ntp.org"}}),
			behavior.Removed(sourcesFile),
			behavior.Ran("chronyc", "reload", "sources"),
		).
		Example("When osapi set no servers", DeleteInput{}, Result{}),
))

// make17 is 17 servers, one more than a request may have.
func make17() []string {
	out := make([]string, 17)
	for i := range out {
		out[i] = "0.pool.ntp.org"
	}

	return out
}
