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
)

// SigningKeyFile is the file in controller.pki.key_dir of the signing key
// that the controller makes when the config has none.
const SigningKeyFile = "signing.key"

// SigningKeyID is the id of that key, which each token that it signs names.
const SigningKeyID = "local"

// randRead is rand.Read, for a test to replace.
var randRead = rand.Read

// EnsureSigningKey gives c the signing key in controller.pki.key_dir when
// its config has none, and makes the key first when it is not there. Thus a
// controller with an empty config verifies tokens, and controller token
// generate on its host signs them with the same key. A key in the config
// wins: the file is used only when there is none.
func EnsureSigningKey(
	c *Config,
) error {
	sec := &c.Controller.API.Security
	if len(sec.SigningKeys) > 0 {
		return nil
	}
	path := filepath.Join(c.Controller.PKI.KeyDir, SigningKeyFile)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		key := make([]byte, 32)
		if _, err := randRead(key); err != nil {
			return fmt.Errorf("make the signing key: %w", err)
		}
		data = []byte(hex.EncodeToString(key) + "\n")
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return fmt.Errorf("make the signing key: %w", err)
		}
		if err := os.WriteFile(path, data, 0o600); err != nil {
			return fmt.Errorf("write the signing key: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("read the signing key: %w", err)
	}
	sec.SigningKeys = []SigningKey{{ID: SigningKeyID, Key: strings.TrimSpace(string(data))}}

	return nil
}
