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

package bus

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	natsclient "github.com/osapi-io/nats-client/pkg/client"
	natsembedded "github.com/osapi-io/nats-server/pkg/server"

	"github.com/osapi-io/osapi/internal/config"
)

// Server is the embedded NATS server that osapi runs.
type Server struct {
	server *natsembedded.Server
}

// NewServer starts the embedded NATS server that cfg.Server describes, then
// creates osapi's streams and buckets on it with Setup. If Setup fails,
// NewServer stops the server and returns the error.
func NewServer(
	ctx context.Context,
	log *slog.Logger,
	cfg config.NATS,
) (*Server, error) {
	sc := cfg.Server

	opts, err := ServerOptions(sc.Host, sc.Port, sc.StoreDir, sc.Auth)
	if err != nil {
		return nil, fmt.Errorf("NATS server auth: %w", err)
	}

	s := natsembedded.New(log, &natsembedded.Options{
		Options:      opts,
		ReadyTimeout: 5 * time.Second,
	})
	if err := s.Start(); err != nil {
		return nil, fmt.Errorf("start embedded NATS server: %w", err)
	}

	if err := Prepare(ctx, log, &natsclient.Options{
		Host: sc.Host,
		Port: sc.Port,
		Auth: ServerClientAuth(sc.Auth),
		Name: "osapi-nats-setup",
	}, sc.Namespace, cfg); err != nil {
		s.Stop()

		return nil, fmt.Errorf("set up JetStream: %w", err)
	}

	log.Info(
		"embedded NATS server running",
		"host", sc.Host,
		"port", sc.Port,
		"store_dir", sc.StoreDir,
		"namespace", sc.Namespace,
		"auth.type", sc.Auth.Type,
	)

	return &Server{server: s}, nil
}

// Start does nothing. NewServer starts the server, because Setup needs it to
// run first.
func (s *Server) Start() {}

// Stop stops the server.
func (s *Server) Stop(
	_ context.Context,
) {
	s.server.Stop()
}
