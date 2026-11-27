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
	"log/slog"

	"github.com/osapi-io/osapi/internal/cli/command"
)

// StartLogs are the lines that the agent writes at the info level when it
// starts on Linux with the default config, which the docs page of the
// command shows as the log handler writes them. osapi start writes them
// too.
var StartLogs = []command.LogLine{
	{
		Level: slog.LevelInfo,
		Msg:   "starting node agent",
		Attrs: []slog.Attr{
			slog.String("component", "agent"),
			slog.String("subsystem", "agent"),
		},
	},
	{
		Level: slog.LevelInfo,
		Msg:   "agent configuration",
		Attrs: []slog.Attr{
			slog.String("component", "agent"),
			slog.String("subsystem", "agent"),
			slog.String("machine_id", "d530664edc1c431eb5ef1e9dfd6617a5"),
			slog.String("hostname", "web-03"),
			slog.Int("max_jobs", 10),
			slog.Any("labels", map[string]string{"group": "web.dev.us-east"}),
		},
	},
	{
		Level: slog.LevelInfo,
		Msg:   "registered in agent registry",
		Attrs: []slog.Attr{
			slog.String("component", "agent"),
			slog.String("subsystem", "agent.heartbeat"),
			slog.String("machine_id", "d530664edc1c431eb5ef1e9dfd6617a5"),
			slog.String("hostname", "web-03"),
			slog.String("key", "agents.d530664edc1c431eb5ef1e9dfd6617a5"),
			slog.String("ttl", "30s"),
		},
	},
	{
		Level: slog.LevelInfo,
		Msg:   "node agent started successfully",
		Attrs: []slog.Attr{
			slog.String("component", "agent"),
			slog.String("subsystem", "agent"),
		},
	},
	{
		Level: slog.LevelInfo,
		Msg:   "metrics server started",
		Attrs: []slog.Attr{
			slog.String("component", "agent"),
			slog.String("subsystem", "metrics"),
			slog.String("addr", "0.0.0.0:9091"),
		},
	},
	{
		Level: slog.LevelInfo,
		Msg:   "http(s) server started",
		Attrs: []slog.Attr{
			slog.String("component", "agent"),
			slog.String("subsystem", "metrics"),
			slog.String("address", "[::]:9091"),
		},
	},
}
