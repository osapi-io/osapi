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
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/osfs"
	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"

	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/telemetry/tracing"
)

var (
	appConfig  config.Config
	appFs      avfs.VFS = osfs.New()
	logger              = slog.New(slog.NewTextHandler(os.Stdout, nil))
	jsonOutput bool

	// skipConfigCmds lists subcommands that don't need a config file. Matched
	// against the resolved command and its parents, so `completion bash` is
	// covered by `completion`.
	skipConfigCmds = map[string]bool{
		"version":    true,
		"help":       true,
		"completion": true,
		"config":     true,
		"roles":      true,
	}
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "osapi",
	Short: "A CRUD API for managing Linux systems.",
	Long: `A CRUD API for managing Linux systems, responsible for ensuring that
the system's configuration matches the desired state.

https://github.com/osapi-io/osapi`,
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

	cli.GuardGroups(rootCmd)

	cmd, err := rootCmd.ExecuteContextC(ctx)
	cli.FlushKV()
	if err != nil {
		cli.PrintError(os.Stderr, cmd, err, jsonOutput)
		os.Exit(cli.ExitCode(err))
	}
}

func init() {
	cobra.OnInitialize(func() { cli.SetJSONOutput(jsonOutput) }, initConfig, initLogger)
	cli.ApplyTheme(rootCmd, displayVersion())

	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable or disable debug mode")
	rootCmd.PersistentFlags().BoolVarP(&jsonOutput, "json", "j", false, "Enable JSON output")

	rootCmd.PersistentFlags().
		StringP("osapi-file", "f", "/etc/osapi/osapi.yaml", "Path to config file")

	_ = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	_ = viper.BindPFlag("osapiFile", rootCmd.PersistentFlags().Lookup("osapi-file"))

	// The environment is read before any command resolves its config file, so
	// OSAPI_OSAPIFILE picks the file as -f does. config.Read sets the same up
	// again for the file's keys, but by then the path has been chosen.
	viper.SetEnvPrefix(config.EnvPrefix)
	viper.AutomaticEnv()
}

// skipConfig reports whether the command being run needs no config file.
//
// It asks Cobra which command the arguments resolve to rather than reading
// os.Args[1] directly. Reading the raw argument missed `help` and `completion`,
// which Cobra adds itself, so both exited 1 on a host with no config file. It
// also misses a flag before the subcommand, and a subcommand of one that is
// skipped, which is why the resolved command's parents are checked too.
func skipConfig() bool {
	cmd, _, err := rootCmd.Find(os.Args[1:])
	if err != nil || cmd == nil {
		return false
	}

	for c := cmd; c != nil; c = c.Parent() {
		if skipConfigCmds[c.Name()] {
			return true
		}
	}

	return false
}

func initConfig() {
	// Commands that don't need a config file.
	if skipConfig() {
		return
	}

	if err := config.Load(viper.GetViper(), viper.GetString("osapiFile"), &appConfig); err != nil {
		cli.LogFatal(logger, "invalid config", err, "osapiFile", viper.GetString("osapiFile"))
	}

	// Auto-enable tracing in debug mode so trace_id appears in log lines.
	// No exporter is set — just log correlation, no span dumps.
	if appConfig.Debug && !appConfig.Telemetry.Tracing.Enabled {
		appConfig.Telemetry.Tracing.Enabled = true
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
		handler = tint.NewTextHandler(os.Stderr, &tint.Options{
			Level:      logLevel,
			TimeFormat: time.Kitchen,
			NoColor:    !term.IsTerminal(int(os.Stdout.Fd())),
		})
	}

	handler = tracing.NewTraceHandler(handler)
	logger = slog.New(handler)
	cli.SetJSONOutput(jsonOutput)
}

// displayVersion is the version help reports: the release version, or "dev"
// for a build without one.
func displayVersion() string {
	if version == "" {
		return "dev"
	}

	return version
}
