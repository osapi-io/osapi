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

import (
	"context"
	"log/slog"
	"time"

	"github.com/avfs/avfs"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/job"
)

// Agent is the node agent, which agent.Setup makes.
type Agent interface {
	command.Component
	IsReady() error
	SetMeterProvider(mp *sdkmetric.MeterProvider)
	SubComponents() map[string]job.SubComponentInfo
	LastHeartbeatTime() time.Time
	SetVersion(version string)
	SetLogLevel(level *slog.LevelVar)
}

// Deps is what start uses: the config, the log, the version of the binary,
// the file system, and what checks the host, starts the tracer and makes
// the agent.
type Deps struct {
	Config     config.Config
	ConfigFile string
	Logger     *slog.Logger
	// LogLevel is the level of Logger, which the agent changes when the
	// controller's config of the agent sets a log level. Nil leaves it.
	LogLevel *slog.LevelVar
	Version  string
	FS       avfs.VFS
	// CheckHost is an error when the agent does not support the host.
	CheckHost func() error
	// Tracer starts the tracer of the service name, and returns what
	// flushes and stops it.
	Tracer func(ctx context.Context, name string) (func(context.Context) error, error)
	// NewAgent makes the agent, and what closes its NATS connection.
	NewAgent func(
		ctx context.Context,
		log *slog.Logger,
		fs avfs.VFS,
		cfg config.Config,
		mp *sdkmetric.MeterProvider,
	) (Agent, func(), error)
}
