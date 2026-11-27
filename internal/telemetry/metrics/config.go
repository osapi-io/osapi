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
	"fmt"
	"log/slog"

	"github.com/osapi-io/osapi/internal/config"
)

// FromConfig makes the metrics server that cfg turns on, and does not start
// it. It is nil when cfg turns it off. A server that cfg turns on and that
// cannot be made is the error: the process does not run without the metrics
// that its config asks for. With TLS on and no cert_file, it
// serves the self-signed certificate name in keyDir, which it makes.
// keyDirKey is the config key of keyDir, such as nats.server.key_dir: the
// error names it when the certificate cannot be made there
// (osapi-io/osapi#1551).
func FromConfig(
	cfg config.MetricsServer,
	keyDir string,
	keyDirKey string,
	name string,
	log *slog.Logger,
) (*Server, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	// A metrics server whose certificate does not load does not start: it
	// does not fall back to plain HTTP.
	// Nothing pins a metrics certificate, so a renamed host makes a new one
	// for its new name (osapi-io/osapi#1493).
	t, err := cfg.TLS.WithCurrentSelfSigned(keyDir, name)
	if err != nil {
		return nil, fmt.Errorf("metrics server: tls: %w: set %s to a directory that this user "+
			"can write, or give the metrics server a tls.cert_file", err, keyDirKey)
	}
	tlsConfig, err := t.Config()
	if err != nil {
		return nil, fmt.Errorf("metrics server: tls: %w", err)
	}
	s, err := New(cfg.Host, cfg.Port, log)
	if err != nil {
		return nil, err
	}
	s.tls = tlsConfig

	return s, nil
}
