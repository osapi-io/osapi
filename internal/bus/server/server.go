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
	"os"
	"path/filepath"
	"time"

	"github.com/avfs/avfs/vfs/osfs"
	"github.com/osapi-io/osapi/internal/natsauth"

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
	// In operator mode with no resolver_file, the server makes the
	// operator, the account and a credential for each role at its first
	// start, as osapi nats auth operator create does (osapi-io/osapi#1413).
	if sc.Auth.Type == "operator" && sc.Auth.ResolverFile == "" {
		auth, err := selfMadeOperator(ctx, cfg)
		if err != nil {
			return nil, fmt.Errorf("NATS server: %w", err)
		}
		sc.Auth = auth
	}

	opts, err := bus.ServerOptions(sc.Host, sc.Port, sc.StoreDir, sc.Auth, sc.TLS, sc.KeyDir)
	if err != nil {
		return nil, fmt.Errorf("NATS server: %w", err)
	}
	if err := bus.ApplyRoles(opts, sc.Auth, sc.Namespace, cfg); err != nil {
		return nil, fmt.Errorf("NATS server: %w", err)
	}
	// The server's own connection trusts its own certificate.
	selfTLS := config.SelfClient(opts.TLSConfig)
	opts.JetStreamKey, err = bus.JetStreamKey(sc.JetStream)
	if err != nil {
		return nil, err
	}

	// The server logs its debug lines when osapi logs at debug. It never
	// traces: a trace line is each protocol message, with its payload, such
	// as a sealed job or a credential.
	opts.Debug = log.Enabled(ctx, slog.LevelDebug)
	opts.Trace = false

	s := natsembedded.New(log, &natsembedded.Options{
		Options:      opts,
		ReadyTimeout: 5 * time.Second,
	})
	if err := s.Start(ctx); err != nil {
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

// selfMadeOperator is the operator auth of the credentials in
// config.NATSAuthDir, which it makes when they do not exist.
func selfMadeOperator(
	ctx context.Context,
	cfg config.NATS,
) (config.NATSServerAuth, error) {
	dir := config.NATSAuthDir(cfg.Server.KeyDir)
	resolver := filepath.Join(dir, "resolver.conf")
	if _, err := os.Stat(resolver); err != nil {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return config.NATSServerAuth{}, fmt.Errorf("make %s: %w", dir, err)
		}
		if _, err := natsauthOperator(
			ctx,
			osfs.NewWithNoIdm(),
			dir,
			cfg.Server.Namespace,
			cfg,
		); err != nil {
			return config.NATSServerAuth{}, fmt.Errorf("make the operator: %w", err)
		}
	}

	return config.NATSServerAuth{
		Type:         "operator",
		ResolverFile: resolver,
		CredsFile:    filepath.Join(dir, "setup.creds"),
	}, nil
}

// natsauthOperator makes the operator credentials. Tests swap it.
var natsauthOperator = natsauth.Operator
