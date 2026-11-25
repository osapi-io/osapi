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

// Package server runs the embedded NATS server that osapi starts for itself,
// and writes its heartbeat.
package server

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	natsclient "github.com/osapi-io/nats-client/pkg/client"
	natsembedded "github.com/osapi-io/nats-server/pkg/server"

	"github.com/osapi-io/osapi/internal/bus"
	"github.com/osapi-io/osapi/internal/config"
)

// dialFn connects to NATS (injectable for testing).
var dialFn = bus.Dial

// Server is the embedded NATS server. It keeps one client connection to
// itself, for its heartbeat.
type Server struct {
	log    *slog.Logger
	cfg    config.NATSServer
	server *natsembedded.Server
	nc     *natsclient.Client
	// registryBucket is the KV bucket of the registry, or "" when the config
	// names none.
	registryBucket string
}

// New starts the embedded NATS server that cfg.Server describes, connects to
// it, and creates osapi's streams and buckets with bus.Setup. If a step fails,
// New stops the server and returns the error.
func New(
	ctx context.Context,
	log *slog.Logger,
	cfg config.NATS,
	mp *sdkmetric.MeterProvider,
) (*Server, error) {
	sc := cfg.Server

	opts, err := bus.ServerOptions(sc.Host, sc.Port, sc.StoreDir, sc.Auth, sc.TLS)
	if err != nil {
		return nil, fmt.Errorf("NATS server: %w", err)
	}
	// The server's own connection trusts its own certificate.
	selfTLS := config.SelfClient(opts.TLSConfig)
	opts.JetStreamKey, err = bus.JetStreamKey(sc.JetStream)
	if err != nil {
		return nil, err
	}

	s := natsembedded.New(log, &natsembedded.Options{
		Options:      opts,
		ReadyTimeout: 5 * time.Second,
	})
	if err := s.Start(); err != nil {
		return nil, fmt.Errorf("start embedded NATS server: %w", err)
	}

	nc, err := dialFn(log, bus.Telemetry(&natsclient.Options{
		Host: sc.Host,
		Port: sc.Port,
		Auth: bus.ServerClientAuth(sc.Auth),
		Name: "osapi-nats-server",
		TLS:  selfTLS,
	}, mp))
	if err != nil {
		s.Stop()

		return nil, err
	}

	if err := bus.Setup(ctx, nc, sc.Namespace, cfg); err != nil {
		bus.Close(nc)
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

	return &Server{
		log:            log,
		cfg:            sc,
		server:         s,
		nc:             nc,
		registryBucket: registryBucket(sc.Namespace, cfg),
	}, nil
}

// Start does nothing. New starts the server, because bus.Setup needs it to
// run first.
func (s *Server) Start() {}

// Stop closes the connection of the server, then stops the server.
func (s *Server) Stop(
	_ context.Context,
) {
	bus.Close(s.nc)
	s.server.Stop()
}

// registryBucket is the name of the registry KV bucket of cfg in namespace,
// or "" when cfg names none.
func registryBucket(
	namespace string,
	cfg config.NATS,
) string {
	if cfg.Registry.Bucket == "" {
		return ""
	}

	return bus.RegistryKV(namespace, cfg.Registry).Bucket
}
