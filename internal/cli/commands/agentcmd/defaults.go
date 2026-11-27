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

package agentcmd

import "github.com/osapi-io/osapi/internal/config"

// Defaults are the defaults of the agent section of the config, which agent
// start owns. osapi config default get prints them.
var Defaults = config.AgentConfig{
	NATS: config.NATSConnection{
		Host:       "localhost",
		Port:       4222,
		ClientName: "osapi-agent",
		Namespace:  "osapi",
		Auth:       config.NATSAuth{Type: "none"},
		TLS:        config.ClientTLS{Enabled: true},
	},
	Consumer: config.AgentConsumer{
		Name:              "jobs-agent",
		MaxDeliver:        5,
		AckWait:           "2m",
		MaxAckPending:     1000,
		ReplayPolicy:      "instant",
		BackOff:           []string{"30s", "2m", "5m", "15m", "30m"},
		InactiveThreshold: "24h",
	},
	MaxJobs: 10,
	Facts:   config.AgentFacts{Interval: "60s"},
	Conditions: config.AgentConditions{
		MemoryPressureThreshold: 90,
		HighLoadMultiplier:      2.0,
		DiskPressureThreshold:   90,
	},
	Metrics: config.MetricsServer{
		Enabled: true, Host: "0.0.0.0", Port: 9091, TLS: config.ServerTLS{Enabled: true},
	},
	// The agent keeps its keys apart from the controller's and the NATS
	// server's, which share /etc/osapi/pki on a host that runs each.
	PKI: config.AgentPKI{KeyDir: "/etc/osapi/pki/agent"},
}
