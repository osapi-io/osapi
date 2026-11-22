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

package connectivity

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/osapi-io/osapi/internal/behavior/host"
)

// pinger runs the platform's ping, with the arguments that args gives, and
// reads its summary. The ping of Linux (iputils) and the ping of macOS print
// the same two summary lines, so one parser reads both.
type pinger struct {
	exec host.Exec
	args func(address string) (string, []string)
}

// newLinux is the provider on Linux: iputils ping, which reaches an
// address without root when net.ipv4.ping_group_range admits the agent's
// group, or through its own file capability.
func newLinux(
	h host.Host,
) (Provider, error) {
	return &pinger{exec: h.Exec, args: linuxArgs}, nil
}

// newDarwin is the provider on macOS: ping, or ping6 for an IPv6 address.
func newDarwin(
	h host.Host,
) (Provider, error) {
	return &pinger{exec: h.Exec, args: darwinArgs}, nil
}

// linuxArgs sends three packets, waits two seconds for each reply and four
// seconds in all, and prints numbers, not names.
func linuxArgs(
	address string,
) (string, []string) {
	return "ping", []string{"-n", "-c", "3", "-W", "2", "-w", "4", address}
}

// darwinArgs is the same on macOS, where -t is the time in all and ping6
// pings an IPv6 address.
func darwinArgs(
	address string,
) (string, []string) {
	if strings.Contains(address, ":") {
		return "ping6", []string{"-n", "-c", "3", address}
	}

	return "ping", []string{"-n", "-c", "3", "-t", "4", address}
}

var (
	// countsLine is "3 packets transmitted, 3 received, 0% packet loss" on
	// Linux, and "3 packets transmitted, 3 packets received, 0.0% packet
	// loss" on macOS.
	countsLine = regexp.MustCompile(
		`(\d+) packets transmitted, (\d+) (?:packets )?received.*?([\d.]+)% packet loss`)
	// rttLine is "rtt min/avg/max/mdev = 0.04/0.05/0.06/0.01 ms" on Linux,
	// and "round-trip min/avg/max/stddev = ..." on macOS.
	rttLine = regexp.MustCompile(`= ([\d.]+)/([\d.]+)/([\d.]+)/[\d.]+ ms`)
)

// Ping runs ping and reads its summary. ping exits 1 when no reply came
// back, with the summary still printed, so that is a result with 100% loss.
// It is an error only when ping printed no summary, such as for an address
// the host has no route to.
func (p *pinger) Ping(
	ctx context.Context,
	address string,
) (PingResult, error) {
	name, args := p.args(address)
	out, err := p.exec.RunCmd(ctx, name, args)
	counts := countsLine.FindStringSubmatch(out)
	if counts == nil {
		if err != nil {
			return PingResult{}, fmt.Errorf("ping %s: %w: %s", address, err, strings.TrimSpace(out))
		}

		return PingResult{}, fmt.Errorf("ping %s: no summary in its output", address)
	}
	sent, _ := strconv.Atoi(counts[1])
	received, _ := strconv.Atoi(counts[2])
	loss, _ := strconv.ParseFloat(counts[3], 64)
	r := PingResult{PacketsSent: sent, PacketsReceived: received, PacketLoss: loss}
	if rtt := rttLine.FindStringSubmatch(out); rtt != nil {
		r.MinRTT, r.AvgRTT, r.MaxRTT = millis(rtt[1]), millis(rtt[2]), millis(rtt[3])
	}

	return r, nil
}

// millis is a time in milliseconds with two decimals, such as 14.64ms.
func millis(
	ms string,
) string {
	v, _ := strconv.ParseFloat(ms, 64)

	return strconv.FormatFloat(v, 'f', 2, 64) + "ms"
}
