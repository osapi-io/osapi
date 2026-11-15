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

package exec_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	"github.com/osapi-io/osapi/internal/exec"
)

type SpanPublicTestSuite struct {
	suite.Suite

	recorder *tracetest.SpanRecorder
	previous trace.TracerProvider
}

func TestSpanPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(SpanPublicTestSuite))
}

func (s *SpanPublicTestSuite) SetupTest() {
	s.recorder = tracetest.NewSpanRecorder()
	s.previous = otel.GetTracerProvider()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(s.recorder)))
}

func (s *SpanPublicTestSuite) TearDownTest() {
	otel.SetTracerProvider(s.previous)
}

func (s *SpanPublicTestSuite) TestRunCmd() {
	tests := []struct {
		name     string
		program  string
		args     []string
		validate func(span sdktrace.ReadOnlySpan)
	}{
		{
			name:    "when the command succeeds records its exit code",
			program: "sh",
			args:    []string{"-c", "exit 0"},
			validate: func(span sdktrace.ReadOnlySpan) {
				s.Equal("exec sh", span.Name())
				s.Equal(codes.Unset, span.Status().Code)
				s.Contains(attrMap(span), "process.exit.code=0")
			},
		},
		{
			name:    "when the command fails marks the span as an error",
			program: "sh",
			args:    []string{"-c", "exit 3"},
			validate: func(span sdktrace.ReadOnlySpan) {
				s.Equal(codes.Error, span.Status().Code)
				s.Contains(attrMap(span), "process.exit.code=3")
			},
		},
		{
			name:    "when the program does not exist has no exit code",
			program: "osapi-no-such-program",
			validate: func(span sdktrace.ReadOnlySpan) {
				s.Equal(codes.Error, span.Status().Code)
				for _, kv := range span.Attributes() {
					s.NotEqual("process.exit.code", string(kv.Key))
				}
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()
			e := exec.New(slog.Default(), false)

			_, _ = e.RunCmd(context.Background(), tc.program, tc.args)

			s.Require().Len(s.recorder.Ended(), 1)
			tc.validate(s.recorder.Ended()[0])
		})
	}
}

func (s *SpanPublicTestSuite) TestStartSpan() {
	tests := []struct {
		name     string
		program  string
		args     []string
		wantName string
		wantSudo bool
	}{
		{
			name:     "when sudo runs a program names the span for the program",
			program:  "sudo",
			args:     []string{"sysctl", "-p", "/etc/sysctl.d/osapi.conf"},
			wantName: "exec sysctl",
			wantSudo: true,
		},
		{
			name:     "when a program runs without sudo names the span for it",
			program:  "chronyc",
			args:     []string{"reload", "sources"},
			wantName: "exec chronyc",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()

			_, span := exec.ExportStartSpan(context.Background(), tc.program, tc.args)
			span.End()

			got := s.recorder.Ended()[0]
			s.Equal(tc.wantName, got.Name())
			s.Contains(
				attrMap(got),
				"osapi.exec.sudo="+map[bool]string{true: "true", false: "false"}[tc.wantSudo],
			)
			for _, kv := range got.Attributes() {
				s.NotContains(kv.Value.String(), "-p", "the arguments stay out of the span")
			}
		})
	}
}

// attrMap is the attributes of span as key=value strings.
func attrMap(
	span sdktrace.ReadOnlySpan,
) []string {
	out := make([]string, 0, len(span.Attributes()))
	for _, kv := range span.Attributes() {
		out = append(out, string(kv.Key)+"="+kv.Value.String())
	}

	return out
}
