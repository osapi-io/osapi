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
package jointoken_test

import (
	"context"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// TestSpans: each operation of the store opens its span, which names the
// token's ID, never the token (osapi-io/osapi#1426).
func (s *JoinTokenPublicTestSuite) TestSpans() {
	recorder := tracetest.NewSpanRecorder()
	previous := otel.GetTracerProvider()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder)))
	defer otel.SetTracerProvider(previous)
	// A span opens under the request's, as in the controller.
	ctx, root := otel.Tracer("test").Start(context.Background(), "request")
	defer root.End()

	token, t, err := s.store.Create(ctx, time.Hour, 2, nil)
	s.Require().NoError(err)
	_, err = s.store.Consume(ctx, token)
	s.Require().NoError(err)
	_, err = s.store.List(ctx)
	s.Require().NoError(err)
	s.Require().NoError(s.store.Revoke(ctx, t.ID))

	secret := token[strings.Index(token, ".")+1:]
	seen := map[string]bool{}
	for _, sp := range recorder.Ended() {
		seen[sp.Name()] = true
		for _, a := range sp.Attributes() {
			s.False(
				strings.Contains(a.Value.String(), secret),
				"%s holds the token in %s",
				sp.Name(),
				a.Key,
			)
		}
	}
	for _, name := range []string{"jointoken.create", "jointoken.consume", "jointoken.list", "jointoken.revoke"} {
		s.True(seen[name], name)
	}
}
