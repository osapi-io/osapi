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

package metrics

import (
	"log/slog"

	"github.com/osapi-io/osapi/internal/config"
)

// FromConfig makes the metrics server that cfg turns on, and does not start
// it. It is nil when cfg turns it off.
func FromConfig(
	cfg config.MetricsServer,
	log *slog.Logger,
) *Server {
	if !cfg.Enabled {
		return nil
	}
	// A metrics server whose certificate does not load does not start: it
	// does not fall back to plain HTTP.
	tlsConfig, err := cfg.TLS.Config()
	if err != nil {
		log.Error("metrics server: tls", slog.String("error", err.Error()))

		return nil
	}
	s := New(cfg.Host, cfg.Port, log)
	if s != nil {
		s.tls = tlsConfig
	}

	return s
}
