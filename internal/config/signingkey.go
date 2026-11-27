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

package config

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/avfs/avfs"
	"github.com/osapi-io/osapi/internal/validation"
)

// SigningKeyFile is the file in controller.pki.key_dir of the signing key
// that the controller makes when the config has none.
const SigningKeyFile = "signing.key"

// SigningKeyID is the id of that key, which each token that it signs names.
const SigningKeyID = "local"

// randRead is rand.Read, for a test to replace.
var randRead = rand.Read

// ErrShortSigningKey is a signing key with fewer characters than
// validation.MinSigningKeyLen. An empty key would let anybody sign a token.
var ErrShortSigningKey = errors.New("the signing key is too short")

// EnsureSigningKey gives c the signing key in controller.pki.key_dir when
// its config has none, and makes the key first when it is not there. Thus a
// controller with an empty config verifies tokens, and controller token
// generate on its host signs them with the same key. A key in the config
// wins: the file is used only when there is none.
//
// Each key, from the config, from a secret reference or from the file, has
// at least validation.MinSigningKeyLen characters. The config check runs
// before the references are resolved and before the file is read, so this is
// the check that sees the key that signs.
func EnsureSigningKey(
	c *Config,
) error {
	sec := &c.Controller.API.Security
	if len(sec.SigningKeys) == 0 {
		key, err := readOrMakeSigningKey(
			filepath.Join(c.Controller.PKI.KeyDir, SigningKeyFile),
		)
		if err != nil {
			return err
		}
		sec.SigningKeys = []SigningKey{{ID: SigningKeyID, Key: key}}
	}
	for _, k := range sec.SigningKeys {
		if len(k.Key) < validation.MinSigningKeyLen {
			return fmt.Errorf(
				"%w: key %q has %d characters, and needs %d",
				ErrShortSigningKey, k.ID, len(k.Key), validation.MinSigningKeyLen,
			)
		}
	}

	return nil
}

// readFile, writeFile and link are the os calls of readOrMakeSigningKey,
// for a test to replace.
var (
	readFile  = func(name string) ([]byte, error) { return avfs.ReadFile(hostFS, name) }
	writeFile = func(name string, data []byte, perm os.FileMode) error {
		return avfs.WriteFile(hostFS, name, data, perm)
	}
	link = func(oldname, newname string) error { return hostFS.Link(oldname, newname) }
)

// readOrMakeSigningKey is the key in path, which it makes when path is not
// there. The new key goes to a temporary file, which a link puts at path. A
// link does not replace a file, so of two processes that make a key at the
// same time, one key wins and both read it. A reader never sees a part of a
// key.
func readOrMakeSigningKey(
	path string,
) (string, error) {
	data, err := readFile(path)
	if err == nil {
		return strings.TrimSpace(string(data)), nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("read the signing key: %w", err)
	}
	// 32 bytes of key, then 8 bytes that name the temporary file. The name
	// shows no byte of the key.
	b := make([]byte, 40)
	if _, err := randRead(b); err != nil {
		return "", fmt.Errorf("make the signing key: %w", err)
	}
	key, suffix := b[:32], b[32:]
	if err := hostFS.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", fmt.Errorf("make the signing key: %w", err)
	}
	tmp := path + "." + hex.EncodeToString(suffix) + ".tmp"
	defer func() { _ = hostFS.Remove(tmp) }()
	if err := writeFile(tmp, []byte(hex.EncodeToString(key)+"\n"), 0o600); err != nil {
		return "", fmt.Errorf("write the signing key: %w", err)
	}
	if err := link(tmp, path); err != nil && !errors.Is(err, os.ErrExist) {
		return "", fmt.Errorf("write the signing key: %w", err)
	}
	// The key at path: this one, or the one of a process that was first.
	data, err = readFile(path)
	if err != nil {
		return "", fmt.Errorf("read the signing key: %w", err)
	}

	return strings.TrimSpace(string(data)), nil
}
