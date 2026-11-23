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

package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli"
)

type LogPublicTestSuite struct {
	suite.Suite
}

func (suite *LogPublicTestSuite) TearDownTest() {
	cli.ResetOsExit()
}

func (suite *LogPublicTestSuite) TestLogFatal() {
	tests := []struct {
		name         string
		message      string
		err          error
		kvPairs      []any
		wantCode     int
		validateFunc func(string)
	}{
		{
			name:     "when error is provided logs error",
			message:  "something failed",
			err:      fmt.Errorf("connection refused"),
			wantCode: 1,
			validateFunc: func(output string) {
				for _, want := range []string{"something failed", "connection refused"} {
					assert.Contains(suite.T(), output, want)
				}
			},
		},
		{
			name:     "when error is nil logs without error key",
			message:  "fatal event",
			err:      nil,
			wantCode: 1,
			validateFunc: func(output string) {
				for _, want := range []string{"fatal event"} {
					assert.Contains(suite.T(), output, want)
				}
			},
		},
		{
			name:     "when extra kv pairs are provided logs them",
			message:  "startup failed",
			err:      fmt.Errorf("bad config"),
			kvPairs:  []any{"host", "localhost"},
			wantCode: 1,
			validateFunc: func(output string) {
				for _, want := range []string{"startup failed", "bad config", "host", "localhost"} {
					assert.Contains(suite.T(), output, want)
				}
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			var exitCode int
			cli.SetOsExit(func(code int) { exitCode = code })

			defer cli.ResetErrOut()

			// With --json the error is one JSON document, and no log.
			var buf bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&buf, nil))
			var jsonBuf bytes.Buffer
			cli.SetErrOut(&jsonBuf)
			cli.SetJSONOutput(true)
			cli.LogFatal(logger, tc.message, tc.err, tc.kvPairs...)
			assert.Equal(suite.T(), tc.wantCode, exitCode)
			tc.validateFunc(jsonBuf.String())
			var doc cli.ErrorDocument
			suite.Require().NoError(json.Unmarshal(jsonBuf.Bytes(), &doc))
			assert.Equal(suite.T(), cli.ReasonFatal, doc.Reason)
			assert.Empty(suite.T(), buf.String())

			// Without it the error is a line for a person, details beneath.
			exitCode = 0
			buf.Reset()
			var errBuf bytes.Buffer
			cli.SetErrOut(&errBuf)
			cli.SetJSONOutput(false)
			cli.LogFatal(logger, tc.message, tc.err, tc.kvPairs...)
			assert.Equal(suite.T(), tc.wantCode, exitCode)
			tc.validateFunc(errBuf.String())
			assert.True(suite.T(), strings.HasPrefix(errBuf.String(), "[err] "+tc.message))
			assert.Empty(suite.T(), buf.String())
		})
	}
}

func (suite *LogPublicTestSuite) TestLogHandler() {
	at := time.Date(2026, time.January, 15, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		name   string
		level  slog.Level
		asJSON bool
		color  bool
		want   string
	}{
		{
			name:  "when color is on, the time and keys are faint and the level is green",
			level: slog.LevelInfo,
			color: true,
			want:  "\x1b[2m10:00AM\x1b[0m \x1b[92mINF\x1b[0m started \x1b[2mport=\x1b[0m4222\n",
		},
		{
			name:  "when color is off, plain text",
			level: slog.LevelInfo,
			want:  "10:00AM INF started port=4222\n",
		},
		{
			name:  "when the level is above the line's, nothing",
			level: slog.LevelWarn,
			want:  "",
		},
		{
			name:   "when --json is set, a JSON object",
			level:  slog.LevelInfo,
			asJSON: true,
			want:   `{"time":"2026-01-15T10:00:00Z","level":"INFO","msg":"started","port":4222}` + "\n",
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			var buf bytes.Buffer
			h := cli.LogHandler(&buf, tc.level, tc.asJSON, tc.color)
			r := slog.NewRecord(at, slog.LevelInfo, "started", 0)
			r.AddAttrs(slog.Int("port", 4222))
			if h.Enabled(context.Background(), r.Level) {
				suite.Require().NoError(h.Handle(context.Background(), r))
			}
			suite.Equal(tc.want, buf.String())
		})
	}
}

func TestLogPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(LogPublicTestSuite))
}
