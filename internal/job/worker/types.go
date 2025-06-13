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
	"github.com/retr0h/osapi/internal/messaging"
)

// Worker implements job processing with clean lifecycle management.
type Worker struct {
	logger     *slog.Logger
	appConfig  config.Config
	appFs      afero.Fs
	natsClient messaging.NATSClient
	jobClient  client.JobClient

	// Lifecycle management
	cancel context.CancelFunc
	done   chan struct{}
	wg     sync.WaitGroup
}

// JobContext contains the context and data for a single job execution.
type JobContext struct {
	// RequestID from the original job request
	RequestID string
	// WorkerHostname identifies which worker is processing this job
	WorkerHostname string
	// JobData contains the raw job request data
	JobData []byte
	// ResponseKV is the key-value bucket for storing responses
	ResponseKV string
}
