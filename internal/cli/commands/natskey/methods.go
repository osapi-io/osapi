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

package natskey

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/osapi-io/osapi/internal/audit"
)

// keyBytes is the length of a key before it is hex encoded.
const keyBytes = 32

// randRead fills a key with random bytes (injectable for testing).
var randRead = rand.Read

// Create writes a new random key to the key file of the config. It refuses
// a file that exists, because a different key makes the store unreadable.
func Create(
	ctx context.Context,
	d Deps,
	_ CreateInput,
) (CreateOutput, error) {
	if d.KeyFile == "" {
		return CreateOutput{}, errors.New("nats.server.jetstream.key_file is not set")
	}
	if _, err := d.FS.Stat(d.KeyFile); err == nil {
		return CreateOutput{}, fmt.Errorf(
			"%s exists: a new key in its place makes each stream unreadable", d.KeyFile,
		)
	} else if !errors.Is(err, os.ErrNotExist) {
		return CreateOutput{}, fmt.Errorf("check %s: %w", d.KeyFile, err)
	}

	raw := make([]byte, keyBytes)
	if _, err := randRead(raw); err != nil {
		return CreateOutput{}, fmt.Errorf("make the key: %w", err)
	}
	if err := d.FS.MkdirAll(filepath.Dir(d.KeyFile), 0o700); err != nil {
		return CreateOutput{}, fmt.Errorf("make the directory of %s: %w", d.KeyFile, err)
	}
	f, err := d.FS.OpenFile(d.KeyFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return CreateOutput{}, fmt.Errorf("write %s: %w", d.KeyFile, err)
	}
	if _, err := f.Write([]byte(hex.EncodeToString(raw) + "\n")); err != nil {
		_ = f.Close()

		return CreateOutput{}, fmt.Errorf("write %s: %w", d.KeyFile, err)
	}
	if err := f.Close(); err != nil {
		return CreateOutput{}, fmt.Errorf("write %s: %w", d.KeyFile, err)
	}

	audit.RecordLocal(ctx, d.Audit, "natskey.create", d.KeyFile, "")

	return CreateOutput{KeyFile: d.KeyFile}, nil
}
