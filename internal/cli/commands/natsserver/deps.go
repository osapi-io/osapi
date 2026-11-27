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

package natsserver

import (
	"context"
	"log/slog"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/cryptoreport"
)

// Server is the embedded NATS server, which server.New makes.
type Server interface {
	Start() error
	Stop(ctx context.Context)
	StartHeartbeat(
		ctx context.Context,
		version string,
		conditions config.ProcessConditions,
		crypto *cryptoreport.Report,
	)
}

// Deps is what start uses: the config, the log, the version of the binary,
// and what makes the server.
type Deps struct {
	Config     config.Config
	ConfigFile string
	Logger     *slog.Logger
	Version    string
	NewServer  func(
		ctx context.Context,
		log *slog.Logger,
		cfg config.NATS,
		mp *sdkmetric.MeterProvider,
	) (Server, error)
}
