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
	"context"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/memfs"
	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go/jetstream"
	natsclient "github.com/osapi-io/nats-client/pkg/client"
	"github.com/stretchr/testify/require"

	"github.com/osapi-io/osapi/internal/secret/vault/natsvault"
)

// testBucket is the secrets bucket of the server that startNATS runs.
const testBucket = "test-secrets"

// startNATS runs a NATS server with JetStream and the secrets bucket, and
// is a client of it.
func startNATS(
	t *testing.T,
) *natsclient.Client {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := l.Addr().(*net.TCPAddr).Port
	require.NoError(t, l.Close())

	ns, err := natsserver.NewServer(&natsserver.Options{
		Host:      "127.0.0.1",
		Port:      port,
		JetStream: true,
		StoreDir:  t.TempDir(),
	})
	require.NoError(t, err)
	ns.Start()
	t.Cleanup(ns.Shutdown)
	require.True(t, ns.ReadyForConnections(5*time.Second))

	nc := natsclient.New(slog.New(slog.DiscardHandler), &natsclient.Options{
		Host: "127.0.0.1",
		Port: port,
		Auth: natsclient.AuthOptions{AuthType: natsclient.NoAuth},
		Name: "test",
	})
	require.NoError(t, nc.Connect())
	t.Cleanup(nc.Close)
	require.NoError(t, nc.KVCreateOrUpdate(context.Background(), jetstream.KeyValueConfig{
		Bucket:  testBucket,
		History: 5,
		Storage: jetstream.MemoryStorage,
	}))

	return nc
}

// keyFile is a file system with a new vault key in /etc/osapi/secrets.key,
// and the keys of the vault that read it.
func keyFile(
	t *testing.T,
) (avfs.VFS, natsvault.Keys) {
	fs := memfs.New()
	_, err := natsvault.CreateKey(context.Background(), fs, "/etc/osapi/secrets.key")
	require.NoError(t, err)

	return fs, natsvault.Keys{FS: fs, File: "/etc/osapi/secrets.key"}
}
