// Copyright (c) 2025 John Dewey

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

package worker

import (
	"log/slog"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/spf13/afero"

	"github.com/retr0h/osapi/internal/config"
	"github.com/retr0h/osapi/internal/job/client"
	"github.com/retr0h/osapi/internal/provider/command"
	"github.com/retr0h/osapi/internal/provider/network/dns"
	"github.com/retr0h/osapi/internal/provider/network/ping"
	"github.com/retr0h/osapi/internal/provider/system/disk"
	"github.com/retr0h/osapi/internal/provider/system/host"
	"github.com/retr0h/osapi/internal/provider/system/load"
	"github.com/retr0h/osapi/internal/provider/system/mem"
)

// New creates a new job worker instance.
func New(
	appFs afero.Fs,
	appConfig config.Config,
	logger *slog.Logger,
	jobClient client.JobClient,
	streamName string,
	hostProvider host.Provider,
	diskProvider disk.Provider,
	memProvider mem.Provider,
	loadProvider load.Provider,
	dnsProvider dns.Provider,
	pingProvider ping.Provider,
	commandProvider command.Provider,
	registryKV jetstream.KeyValue,
) *Worker {
	return &Worker{
		logger:          logger,
		appConfig:       appConfig,
		appFs:           appFs,
		jobClient:       jobClient,
		streamName:      streamName,
		hostProvider:    hostProvider,
		diskProvider:    diskProvider,
		memProvider:     memProvider,
		loadProvider:    loadProvider,
		dnsProvider:     dnsProvider,
		pingProvider:    pingProvider,
		commandProvider: commandProvider,
		registryKV:      registryKV,
	}
}
