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

package cmd

import (
	"context"
	"log/slog"

	"github.com/avfs/avfs"
	"github.com/spf13/viper"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/osapi-io/osapi/internal/agent"
	"github.com/osapi-io/osapi/internal/bus"
	"github.com/osapi-io/osapi/internal/bus/server"
	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/cli/commands/agentcmd"
	"github.com/osapi-io/osapi/internal/cli/commands/controllercmd"
	"github.com/osapi-io/osapi/internal/cli/commands/natsserver"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller"
	"github.com/osapi-io/osapi/internal/telemetry/tracing"
)

// The deps of the processes: what each start uses, made after the config is
// read, when its command runs.

func natsDeps() natsserver.Deps {
	return natsserver.Deps{
		Config:     appConfig,
		ConfigFile: viper.ConfigFileUsed(),
		Logger:     logger,
		Version:    appVersion(),
		NewServer: func(
			ctx context.Context,
			log *slog.Logger,
			cfg config.NATS,
			mp *sdkmetric.MeterProvider,
		) (natsserver.Server, error) {
			return server.New(ctx, log, cfg, mp)
		},
	}
}

func agentDeps() agentcmd.Deps {
	return agentcmd.Deps{
		Config:     appConfig,
		ConfigFile: viper.ConfigFileUsed(),
		Logger:     logger,
		Version:    appVersion(),
		FS:         appFs,
		CheckHost:  func() error { return cli.ValidateDistribution(logger) },
		Tracer:     startTracer,
		NewAgent: func(
			ctx context.Context,
			log *slog.Logger,
			fs avfs.VFS,
			cfg config.Config,
			mp *sdkmetric.MeterProvider,
		) (agentcmd.Agent, func(), error) {
			a, conn, err := agent.Setup(ctx, log, fs, cfg, mp)
			if err != nil {
				return nil, nil, err
			}

			return a, func() { bus.Close(conn.NATS) }, nil
		},
	}
}

func controllerDeps() controllercmd.Deps {
	return controllercmd.Deps{
		Config:     appConfig,
		ConfigFile: viper.ConfigFileUsed(),
		Logger:     logger,
		Version:    appVersion(),
		Tracer:     startTracer,
		NewController: func(
			ctx context.Context,
			log *slog.Logger,
			cfg config.Config,
			opts ...controller.Option,
		) (controllercmd.Controller, error) {
			return controller.New(ctx, log, cfg, opts...)
		},
	}
}

// startTracer starts the tracer of the service name, with the version of
// this binary.
func startTracer(
	ctx context.Context,
	name string,
) (func(context.Context) error, error) {
	return tracing.InitTracer(ctx, name, appConfig.Telemetry.Tracing, versionAttr())
}
