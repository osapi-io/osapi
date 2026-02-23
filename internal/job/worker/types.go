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
	"context"
	"log/slog"
	"sync"

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

// Worker implements job processing with clean lifecycle management.
type Worker struct {
	logger     *slog.Logger
	appConfig  config.Config
	appFs      afero.Fs
	jobClient  client.JobClient
	streamName string

	// System providers
	hostProvider host.Provider
	diskProvider disk.Provider
	memProvider  mem.Provider
	loadProvider load.Provider

	// Network providers
	dnsProvider  dns.Provider
	pingProvider ping.Provider

	// Command provider
	commandProvider command.Provider

	// Lifecycle management
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// JobContext contains the context and data for a single job execution.
type JobContext struct {
	// JobID from the original job request
	JobID string
	// WorkerHostname identifies which worker is processing this job
	WorkerHostname string
	// JobData contains the raw job request data
	JobData []byte
	// ResponseKV is the key-value bucket for storing responses
	ResponseKV string
}
