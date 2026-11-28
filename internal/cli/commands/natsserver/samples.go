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
	"log/slog"

	"github.com/osapi-io/osapi/internal/cli/command"
)

// StartLogs are the lines that the NATS server writes at the info level when
// it starts with the default config, which the docs page of the command
// shows as the log handler writes them. osapi start writes them too.
var StartLogs = []command.LogLine{
	{
		Level: slog.LevelInfo,
		Msg:   "nats server started successfully",
		Attrs: []slog.Attr{slog.String("component", "nats")},
	},
	{
		Level: slog.LevelInfo,
		Msg:   "embedded NATS server running",
		Attrs: []slog.Attr{
			slog.String("component", "nats"),
			slog.String("addr", "localhost:4222"),
			slog.String("store_dir", ".nats/jetstream/"),
			slog.String("namespace", "osapi"),
			slog.String("auth.type", "none"),
		},
	},
	{
		Level: slog.LevelInfo,
		Msg:   "metrics server started",
		Attrs: []slog.Attr{
			slog.String("component", "nats"),
			slog.String("subsystem", "metrics"),
			slog.String("addr", "0.0.0.0:9092"),
		},
	},
	{
		Level: slog.LevelInfo,
		Msg:   "http(s) server started",
		Attrs: []slog.Attr{
			slog.String("component", "nats"),
			slog.String("subsystem", "metrics"),
			slog.String("address", "[::]:9092"),
		},
	},
	{
		Level: slog.LevelInfo,
		Msg:   "registered in component registry",
		Attrs: []slog.Attr{
			slog.String("component", "nats"),
			slog.String("subsystem", "nats.heartbeat"),
			slog.String("hostname", "web-03"),
			slog.String("type", "nats"),
			slog.String("key", "nats.web-03"),
		},
	},
}
