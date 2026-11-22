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

// Package connectivity is the connectivity behavior: whether the host
// reaches an address, by ICMP ping.
package connectivity

import "github.com/osapi-io/osapi/internal/behavior"

// The ping commands and what each prints, for the examples.
const (
	linuxPing   = "ping -n -c 3 -W 2 -w 4 127.0.0.1"
	linuxOutput = `PING 127.0.0.1 (127.0.0.1) 56(84) bytes of data.
64 bytes from 127.0.0.1: icmp_seq=1 ttl=64 time=0.045 ms
64 bytes from 127.0.0.1: icmp_seq=2 ttl=64 time=0.052 ms
64 bytes from 127.0.0.1: icmp_seq=3 ttl=64 time=0.060 ms

--- 127.0.0.1 ping statistics ---
3 packets transmitted, 3 received, 0% packet loss, time 2003ms
rtt min/avg/max/mdev = 0.045/0.052/0.060/0.006 ms
`
	linuxLost       = "ping -n -c 3 -W 2 -w 4 192.0.2.1"
	linuxLostOutput = `PING 192.0.2.1 (192.0.2.1) 56(84) bytes of data.

--- 192.0.2.1 ping statistics ---
3 packets transmitted, 0 received, 100% packet loss, time 2040ms
`
	darwinPing   = "ping -n -c 3 -t 4 127.0.0.1"
	darwinOutput = `PING 127.0.0.1 (127.0.0.1): 56 data bytes
64 bytes from 127.0.0.1: icmp_seq=0 ttl=64 time=0.081 ms
64 bytes from 127.0.0.1: icmp_seq=1 ttl=64 time=0.173 ms
64 bytes from 127.0.0.1: icmp_seq=2 ttl=64 time=0.152 ms

--- 127.0.0.1 ping statistics ---
3 packets transmitted, 3 packets received, 0.0% packet loss
round-trip min/avg/max/stddev = 0.081/0.135/0.173/0.040 ms
`
)

// Behavior is the connectivity definition. The routes, the agent's
// dispatch, the CLI commands, the spec, the docs pages and the tests of
// every layer above the provider are built from it. osapi behavior spec get
// lists every option.
var Behavior = behavior.Register(behavior.Define(
	"connectivity", "Whether the host reaches an address, by ICMP ping",
	behavior.Node,
	behavior.Under(behavior.GroupNetwork),
	behavior.API,
	behavior.Area(behavior.AreaNetworking),
	behavior.CLI,
	behavior.Platform(behavior.PlatformLinux, newLinux),
	behavior.Platform(behavior.PlatformDarwin, newDarwin),
	behavior.Runs("ping"),
	behavior.RunsOnPlatform(behavior.PlatformDarwin, "ping6"),
	behavior.Action("ping", Ping).
		Describe("Ping an address from the host").
		Explain("Sends three ICMP echo requests from the host and reports how many came "+
			"back, and their round-trip times. A host that does not answer is 100% loss, "+
			"not an error. It changes nothing, so it reports changed: false. On Linux "+
			"the agent pings without root when net.ipv4.ping_group_range admits its group.").
		Example("A host that answers", PingInput{Address: "127.0.0.1"},
			PingResult{
				PacketsSent: 3, PacketsReceived: 3,
				MinRTT: "0.04ms", AvgRTT: "0.05ms", MaxRTT: "0.06ms",
			},
			behavior.OnPlatform(behavior.PlatformLinux),
			behavior.Prints(linuxPing, linuxOutput),
		).
		Example("A host that answers, on macOS", PingInput{Address: "127.0.0.1"},
			PingResult{
				PacketsSent: 3, PacketsReceived: 3,
				MinRTT: "0.08ms", AvgRTT: "0.14ms", MaxRTT: "0.17ms",
			},
			behavior.OnPlatform(behavior.PlatformDarwin),
			behavior.Prints(darwinPing, darwinOutput),
		).
		Example("A host that does not answer", PingInput{Address: "192.0.2.1"},
			PingResult{PacketsSent: 3, PacketLoss: 100},
			behavior.OnPlatform(behavior.PlatformLinux),
			behavior.Prints(linuxLost, linuxLostOutput),
		).
		Fails("A name, not an address", PingInput{Address: "example.com"}, "ip_or_fact"),
))
