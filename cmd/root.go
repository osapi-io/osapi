// Copyright (c) 2024 John Dewey

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

// Package cmd provides CLI commands for OSAPI.
package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/osfs"
	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"

	"github.com/retr0h/osapi/internal/cli"
	"github.com/retr0h/osapi/internal/config"
	"github.com/retr0h/osapi/internal/telemetry/tracing"
	"github.com/retr0h/osapi/internal/validation"
)

var (
	appConfig  config.Config
	appFs      avfs.VFS = osfs.New()
	logger              = slog.New(slog.NewTextHandler(os.Stdout, nil))
	jsonOutput bool

	// skipConfigCmds lists subcommands that don't need a config file.
	skipConfigCmds = map[string]bool{"version": true}
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "osapi",
	Short: "A CRUD API for managing Linux systems.",
	Long: `A CRUD API for managing Linux systems, responsible for ensuring that
the system's configuration matches the desired state.

┌─┐┌─┐┌─┐┌─┐┬
│ │└─┐├─┤├─┘│
└─┘└─┘┴ ┴┴  ┴

https://github.com/retr0h/osapi
`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig, initLogger)

	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable or disable debug mode")
	rootCmd.PersistentFlags().BoolVarP(&jsonOutput, "json", "j", false, "Enable JSON output")

	rootCmd.PersistentFlags().
		StringP("osapi-file", "f", "/etc/osapi/osapi.yaml", "Path to config file")

	_ = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	_ = viper.BindPFlag("osapiFile", rootCmd.PersistentFlags().Lookup("osapi-file"))
}

func initConfig() {
	// Commands that don't need a config file.
	if len(os.Args) > 1 && skipConfigCmds[os.Args[1]] {
		return
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("osapi")

	// Controller defaults
	viper.SetDefault("controller.api.port", 8080)
	viper.SetDefault("controller.api.job_timeout", "30s")
	viper.SetDefault("controller.metrics.port", 9090)

	// NATS server defaults
	viper.SetDefault("nats.api.port", 4222)
	viper.SetDefault("nats.api.host", "localhost")

	// NATS connection defaults (controller + agent)
	viper.SetDefault("controller.api.nats.port", 4222)
	viper.SetDefault("controller.api.nats.host", "localhost")
	viper.SetDefault("agent.nats.port", 4222)
	viper.SetDefault("agent.nats.host", "localhost")

	// NATS infrastructure defaults
	viper.SetDefault("nats.stream.name", "JOBS")
	viper.SetDefault("nats.stream.subjects", "jobs.>")
	viper.SetDefault("nats.kv.bucket", "job-queue")
	viper.SetDefault("nats.kv.response_bucket", "job-responses")
	viper.SetDefault("nats.registry.bucket", "agent-registry")
	viper.SetDefault("nats.registry.ttl", "30s")

	// Agent defaults
	viper.SetDefault("agent.max_jobs", 10)
	viper.SetDefault("agent.queue_group", "job-agents")
	viper.SetDefault("agent.facts.interval", "60s")
	viper.SetDefault("agent.conditions.memory_pressure_threshold", 90)
	viper.SetDefault("agent.conditions.high_load_multiplier", 2.0)
	viper.SetDefault("agent.conditions.disk_pressure_threshold", 90)
	viper.SetDefault("agent.consumer.max_deliver", 5)
	viper.SetDefault("agent.consumer.ack_wait", "2m")
	viper.SetDefault("agent.consumer.max_ack_pending", 1000)
	viper.SetDefault("agent.consumer.replay_policy", "instant")

	// PKI defaults.
	viper.SetDefault("agent.pki.enabled", false)
	viper.SetDefault("agent.pki.key_dir", "/etc/osapi/pki")
	viper.SetDefault("controller.pki.enabled", false)
	viper.SetDefault("controller.pki.key_dir", "/etc/osapi/pki")
	viper.SetDefault("controller.pki.auto_accept", false)
	viper.SetDefault("controller.pki.rotation_grace_period", "24h")

	// Enrollment KV defaults.
	viper.SetDefault("nats.enrollment.bucket", "agent-enrollment")
	viper.SetDefault("nats.enrollment.storage", "file")
	viper.SetDefault("nats.enrollment.replicas", 1)

	viper.SetConfigFile(viper.GetString("osapiFile"))

	if err := viper.ReadInConfig(); err != nil {
		cli.LogFatal(logger, "failed to read config", err, "osapiFile", viper.ConfigFileUsed())
	}

	if err := viper.Unmarshal(&appConfig); err != nil {
		cli.LogFatal(logger, "failed to unmarshal config", err, "osapiFile", viper.ConfigFileUsed())
	}

	if errMsg, ok := validation.Struct(appConfig); !ok {
		cli.LogFatal(
			logger,
			"invalid config",
			fmt.Errorf("%s", errMsg),
			"osapiFile",
			viper.ConfigFileUsed(),
		)
	}

	// Auto-enable tracing in debug mode so trace_id appears in log lines.
	// No exporter is set — just log correlation, no span dumps.
	if appConfig.Debug && !appConfig.Telemetry.Tracing.Enabled {
		appConfig.Telemetry.Tracing.Enabled = true
	}

	err := config.Validate(&appConfig)
	if err != nil {
		cli.LogFatal(logger, "validation failed", err, "osapiFile", viper.ConfigFileUsed())
	}
}

func initLogger() {
	logLevel := slog.LevelInfo
	if viper.GetBool("debug") {
		logLevel = slog.LevelDebug
	}

	var handler slog.Handler
	if jsonOutput {
		handler = slog.NewJSONHandler(os.Stderr, nil)
	} else {
		handler = tint.NewHandler(os.Stderr, &tint.Options{
			Level:      logLevel,
			TimeFormat: time.Kitchen,
			NoColor:    !term.IsTerminal(int(os.Stdout.Fd())),
		})
	}

	handler = tracing.NewTraceHandler(handler)
	logger = slog.New(handler)
}
