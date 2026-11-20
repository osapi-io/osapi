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

package controllercmd

import (
	"context"
	"log/slog"

	"github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller"
	"github.com/osapi-io/osapi/internal/job"
)

// Controller is the controller, which controller.New makes.
type Controller interface {
	command.Component
	CheckHealth(ctx context.Context) error
	SubComponents() map[string]job.SubComponentInfo
}

// Deps is what start uses: the config, the log, the version of the binary,
// and what starts the tracer and makes the controller.
type Deps struct {
	Config     config.Config
	ConfigFile string
	Logger     *slog.Logger
	Version    string
	// Tracer starts the tracer of the service name, and returns what
	// flushes and stops it.
	Tracer        func(ctx context.Context, name string) (func(context.Context) error, error)
	NewController func(
		ctx context.Context,
		log *slog.Logger,
		cfg config.Config,
		opts ...controller.Option,
	) (Controller, error)
}
