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

// PingInput is ping's input: the address to ping from the host.
type PingInput struct {
	Address string `json:"address" validate:"required,ip_or_fact" doc:"The IPv4 or IPv6 address to ping, or an @fact. reference that the agent resolves" telemetry:"attr"`
}

// PingResult is ping's output: what came back of the three packets.
type PingResult struct {
	PacketsSent     int     `json:"packets_sent"      doc:"Number of packets sent"`
	PacketsReceived int     `json:"packets_received"  doc:"Number of packets received"`
	PacketLoss      float64 `json:"packet_loss"       doc:"Percentage of packet loss"`
	MinRTT          string  `json:"min_rtt,omitempty" doc:"Minimum round-trip time in milliseconds, with two decimals. Absent when no reply came back."`
	AvgRTT          string  `json:"avg_rtt,omitempty" doc:"Average round-trip time in milliseconds, with two decimals. Absent when no reply came back."`
	MaxRTT          string  `json:"max_rtt,omitempty" doc:"Maximum round-trip time in milliseconds, with two decimals. Absent when no reply came back."`
	Changed         bool    `json:"changed"           doc:"Always false: a ping changes nothing on the host"`
}
