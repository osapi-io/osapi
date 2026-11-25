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
package natsvault_test

import (
	"strings"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/osapi-io/osapi/internal/secret/vault/natsvault"
)

// TestSpans: each operation of the vault opens its span, which names the
// vault and the key, and never the value (osapi-io/osapi#1426).
func (s *NATSVaultPublicTestSuite) TestSpans() {
	recorder := tracetest.NewSpanRecorder()
	previous := otel.GetTracerProvider()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder)))
	defer otel.SetTracerProvider(previous)

	// A span opens under the request's, as in the controller.
	ctx, root := otel.Tracer("test").Start(s.ctx, "request")
	defer root.End()
	nc := startNATS(s.T())
	_, keys := keyFile(s.T())
	v, err := natsvault.Open(s.ctx, nc, testBucket, "default", keys)
	s.Require().NoError(err)
	secret := "hunter2-do-not-trace"

	_, err = v.Put(ctx, "db-password", []byte(secret))
	s.Require().NoError(err)
	_, _, err = v.Get(ctx, "db-password")
	s.Require().NoError(err)
	ring, err := keys.Load()
	s.Require().NoError(err)
	_, err = v.Rotate(ctx, ring)
	s.Require().NoError(err)
	s.Require().NoError(v.Delete(ctx, "db-password"))

	seen := map[string]bool{}
	for _, sp := range recorder.Ended() {
		seen[sp.Name()] = true
		for _, a := range sp.Attributes() {
			s.False(
				strings.Contains(a.Value.String(), secret),
				"%s holds the value in %s",
				sp.Name(),
				a.Key,
			)
		}
	}
	for _, name := range []string{"vault.put", "vault.get", "vault.rotate", "vault.delete"} {
		s.True(seen[name], name)
	}
}
