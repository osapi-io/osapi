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

package agent_test

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/avfs/avfs"

	"github.com/osapi-io/osapi/internal/agent"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/exec"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/job/client"
	"github.com/osapi-io/osapi/internal/provider/command"
	dockerProv "github.com/osapi-io/osapi/internal/provider/container/docker"
	fileProv "github.com/osapi-io/osapi/internal/provider/file"
	"github.com/osapi-io/osapi/internal/provider/node/disk"
	nodeHost "github.com/osapi-io/osapi/internal/provider/node/host"
	"github.com/osapi-io/osapi/internal/provider/node/load"
	"github.com/osapi-io/osapi/internal/provider/node/memory"
	ntpProv "github.com/osapi-io/osapi/internal/provider/node/ntp"
	cronProv "github.com/osapi-io/osapi/internal/provider/schedule/cron"
	"github.com/osapi-io/osapi/internal/telemetry/process"
)

// newTestAgentParams holds all provider dependencies for building a test agent.
// This mirrors the old agent.New() parameter list, making test migration mechanical.
type newTestAgentParams struct {
	appFs           avfs.VFS
	appConfig       config.Config
	logger          *slog.Logger
	jobClient       client.JobClient
	streamName      string
	hostProvider    nodeHost.Provider
	diskProvider    disk.Provider
	memoryProvider  memory.Provider
	loadProvider    load.Provider
	commandProvider command.Provider
	fileProvider    fileProv.Provider
	dockerProvider  dockerProv.Provider
	cronProvider    cronProv.Provider
	ntpProvider     ntpProv.Provider
	processProvider process.Provider
	execManager     exec.Manager
	registryBucket  string
	factsBucket     string
	natsClient      agent.NATSClient
}

// newTestAgent builds a ProviderRegistry from the supplied providers and
// constructs an Agent — replacing the old 19-argument agent.New() call used
// across all test suites.
func newTestAgent(
	p newTestAgentParams,
) *agent.Agent {
	logger := p.logger
	if logger == nil {
		logger = slog.Default()
	}

	streamName := p.streamName
	if streamName == "" {
		streamName = "test-stream"
	}

	registry := agent.NewProviderRegistry()

	registry.Register(
		"node",
		agent.NewNodeProcessor(
			p.ntpProvider,
			nil,
			nil,
			logger,
		),
		p.hostProvider, p.diskProvider, p.memoryProvider, p.loadProvider,
	)

	// The handler tests send network.ping.do as a sample job. The network
	// domain is a behavior now, so a test processor answers it.
	registry.Register(
		"network",
		func(context.Context, job.Request) (json.RawMessage, error) {
			return json.RawMessage(`{"packets_sent":3,"packets_received":3,"packet_loss":0}`), nil
		},
	)

	registry.Register(
		"command",
		agent.NewCommandProcessor(p.commandProvider, true, logger),
		p.commandProvider,
	)

	registry.Register(
		"file",
		agent.NewFileProcessor(p.fileProvider, logger),
		p.fileProvider,
	)

	registry.Register(
		"docker",
		agent.NewDockerProcessor(p.dockerProvider, true, logger),
		p.dockerProvider,
	)

	registry.Register(
		"schedule",
		agent.NewScheduleProcessor(p.cronProvider, logger),
		p.cronProvider,
	)

	return agent.New(
		p.appFs,
		p.appConfig,
		logger,
		p.jobClient,
		streamName,
		p.hostProvider,
		p.diskProvider,
		p.memoryProvider,
		p.loadProvider,
		p.processProvider,
		registry,
		p.registryBucket,
		p.factsBucket,
		p.execManager,
		p.natsClient,
	)
}
