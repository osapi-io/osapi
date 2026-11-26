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

package pkikey

import (
	"context"

	"github.com/avfs/avfs"

	"github.com/osapi-io/osapi/internal/agent/identity"
	"github.com/osapi-io/osapi/internal/audit"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/secret/vault/natsvault"
)

// Deps is what the key commands use: the file system, the config that names
// each key, the identity of this machine for the admission of its agent,
// what connects to NATS for a rotation of a vault key, what rotates the
// agent's keys with the controller, how an environment variable is read for
// a vault with key_env, and the record of a change in the audit log, or nil.
type Deps struct {
	FS          avfs.VFS
	Config      config.Config
	Identity    func() (*identity.Identity, error)
	Dial        func(ctx context.Context) (natsvault.KV, func(), error)
	RotateAgent func(ctx context.Context) (string, error)
	LookupEnv   func(key string) (string, bool)
	Audit       func(ctx context.Context, e audit.Entry) error
}
