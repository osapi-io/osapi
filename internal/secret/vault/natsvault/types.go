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

package natsvault

import (
	"context"
	"time"

	"github.com/avfs/avfs"
	"github.com/nats-io/nats.go/jetstream"
)

// KV is the part of a NATS client that the vault uses. A
// *natsclient.Client is one.
type KV interface {
	KVGet(ctx context.Context, bucket string, key string) (jetstream.KeyValueEntry, error)
	KVCreate(ctx context.Context, bucket string, key string, value []byte) (uint64, error)
	KVUpdate(
		ctx context.Context,
		bucket string,
		key string,
		value []byte,
		revision uint64,
	) (uint64, error)
	KVDelete(ctx context.Context, bucket string, key string) error
	KVKeys(ctx context.Context, bucket string, filters ...string) ([]string, error)
	KVStatus(ctx context.Context, bucket string) (jetstream.KeyValueStatus, error)
}

// Config is the config block of a nats vault. It has no setting: the
// bucket is nats.secrets, and the key is key_file or key_env of the vault.
type Config struct{}

// Vault is a vault of type nats: each value is encrypted with AES-256-GCM
// under the vault key, in the KV bucket of nats.secrets, under
// <vault>.<key>.
type Vault struct {
	name   string
	kv     KV
	bucket string
	keys   Keys
}

// Keys is where the vault key is: a file, or an environment variable. A
// vault reads it at each call, so a rotation applies with no restart.
type Keys struct {
	FS avfs.VFS
	// File is key_file of the vault.
	File string
	// Env is key_env of the vault, when File is empty.
	Env string
	// LookupEnv reads Env, as os.LookupEnv does.
	LookupEnv func(key string) (string, bool)
}

// Keyring is the vault keys, each 32 bytes. The first one encrypts. Each
// one decrypts, so a rotation keeps the old key until each value has the
// new one.
type Keyring struct {
	keys [][]byte
}

// record is a value in the bucket: its metadata, the ID of the key that
// encrypts it, and the encrypted value, with its nonce in front.
type record struct {
	Version uint64    `json:"version"`
	Changed time.Time `json:"changed"`
	By      string    `json:"by,omitempty"`
	KeyID   string    `json:"key_id"`
	Data    []byte    `json:"data"`
}
