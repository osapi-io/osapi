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

package vault

import (
	"context"
	"reflect"
	"time"

	"github.com/avfs/avfs"
	natsclient "github.com/osapi-io/nats-client/pkg/client"
)

// Vault is a store of secrets, by key. A type, such as nats, makes it from
// the config of the vault. The controller resolves each reference through
// one. A value is []byte until secret.Value replaces it (osapi #1240).
type Vault interface {
	// Name is the name of the vault in the config, such as default.
	Name() string
	// Get is the value of key and its metadata. A key that is not there
	// is ErrNotFound.
	Get(ctx context.Context, key string) ([]byte, Meta, error)
	// Put writes value as the next version of key. Actor(ctx) is who
	// changed it.
	Put(ctx context.Context, key string, value []byte) (Meta, error)
	// Delete deletes key. A key that is not there is ErrNotFound.
	Delete(ctx context.Context, key string) error
	// List is the metadata of each key, sorted by key. It has no value.
	List(ctx context.Context) ([]Meta, error)
}

// Meta is what a vault tells about a key, without its value.
type Meta struct {
	// Key is the name of the key in its vault, such as registry-token.
	Key string `json:"key"`
	// Version is 1 for the first value of the key, and one more for each
	// Put after it.
	Version uint64 `json:"version"`
	// Changed is when the last Put occurred.
	Changed time.Time `json:"changed"`
	// By is who did the last Put, from Actor. Empty when it is not known.
	By string `json:"by,omitempty"`
}

// Spec is one vault of controller.secrets.vaults: its name, its type, the
// targets that its secrets can go to, the key of a type that encrypts, and
// the config block of its type.
type Spec struct {
	Name    string         `mapstructure:"name"     validate:"required,resource_name,max=64" doc:"The name that a reference gives, such as default in secret://default/registry-token."`
	Type    string         `mapstructure:"type"     validate:"required"                      doc:"The type of the vault, such as nats."`
	Targets []string       `mapstructure:"targets"  validate:"dive,required"                 doc:"The targets that the secrets of the vault can go to, such as group:prod. Empty is each host."`
	KeyFile string         `mapstructure:"key_file"                                          doc:"For a type that encrypts, such as nats: the file of the vault key. osapi pki key create --name vault:<name> writes it."`
	KeyEnv  string         `mapstructure:"key_env"                                           doc:"For a type that encrypts, such as nats: the environment variable of the vault key, in place of key_file."`
	Config  map[string]any `mapstructure:"config"                                            doc:"The settings of the type. The struct of the type gives its keys."`
}

// Env is what the controller gives each type when it opens a vault: the
// NATS connection, the secrets bucket with its namespace, the file system and
// the environment.
type Env struct {
	NATS *natsclient.Client
	// Bucket is the KV bucket of nats.secrets.bucket, with the namespace,
	// such as osapi-secrets. Empty when the config names none.
	Bucket string
	FS     avfs.VFS
	// LookupEnv reads an environment variable, as os.LookupEnv does.
	LookupEnv func(key string) (string, bool)
}

// Type is a kind of vault, such as nats: its name, what it is, the Go struct
// of its config block, and how to make a vault of it. Define makes one, and
// Register adds it to the types that the controller and the config check
// know.
type Type struct {
	// Name is what the type of a vault names, such as nats.
	Name string
	// Doc says what the type is, in one line.
	Doc string
	// Config is the struct of the config block. Each field has a
	// mapstructure name, a doc tag and its validate rules.
	Config reflect.Type

	check func(spec Spec) error
	build func(ctx context.Context, env Env, spec Spec, cfg any) (Vault, error)
}
