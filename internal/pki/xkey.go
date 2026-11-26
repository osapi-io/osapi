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

package pki

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"

	"github.com/nats-io/nkeys"

	"github.com/osapi-io/osapi/internal/atomicfile"
)

// curveRand is the entropy that a new X25519 key comes from. Overridable in
// tests via export_test.go.
var curveRand = io.Reader(rand.Reader)

// fromCurveSeedFn parses an nkeys curve seed. Overridable in tests via
// export_test.go.
var fromCurveSeedFn = nkeys.FromCurveSeed

// generateCurveSeed makes a new X25519 key and returns its nkeys seed.
func generateCurveSeed() ([]byte, error) {
	kp, err := nkeys.CreateCurveKeysWithRand(curveRand)
	if err != nil {
		return nil, err
	}

	return kp.Seed()
}

// generateCurveSeedFn is the function used to generate X25519 keys.
// Overridable in tests via export_test.go.
var generateCurveSeedFn = generateCurveSeed

// EncryptionPublicKey returns the public X25519 key in the nkeys form that
// starts with "X", or an empty string when the manager has none.
func (m *Manager) EncryptionPublicKey() string {
	return m.encryptionPublicKey
}

// EncryptionKey returns the X25519 key pair, or nil when the manager has none.
func (m *Manager) EncryptionKey() nkeys.KeyPair {
	return m.encryptionKey
}

// ValidEncryptionKey reports whether key is a public X25519 key in the nkeys
// form.
func ValidEncryptionKey(
	key string,
) bool {
	return nkeys.IsValidPublicCurveKey(key)
}

// FingerprintOfKeys returns the fingerprint that covers an agent's two public
// keys: the SHA-256 of the 32 bytes of the Ed25519 key, then the 32 bytes of
// the X25519 key, as "SHA256:<hex>". An operator who admits it admits both
// keys. It is empty when the signing key is empty or the encryption key is
// not a public X25519 key, so a record without both keys never reports a
// fingerprint that admits it.
func FingerprintOfKeys(
	signingKey ed25519.PublicKey,
	encryptionKey string,
) string {
	if len(signingKey) == 0 {
		return ""
	}

	raw, err := nkeys.Decode(nkeys.PrefixByteCurve, []byte(encryptionKey))
	if err != nil {
		return ""
	}

	h := sha256.New()
	_, _ = h.Write(signingKey)
	_, _ = h.Write(raw)

	return "SHA256:" + hex.EncodeToString(h.Sum(nil))
}

// loadOrGenerateEncryptionKey loads the X25519 key from {prefix}.xkey, or
// generates and saves one when the file does not exist. The seed holds the
// public key too, so there is no half pair. A file it cannot read or parse is
// an error, as for the signing key: a new key is not the admitted one.
func (m *Manager) loadOrGenerateEncryptionKey() error {
	path := filepath.Join(m.keyDir, m.keyPrefix+".xkey")

	data, err := m.fs.ReadFile(path)
	switch {
	case err == nil:
		return m.setEncryptionKey(path, bytes.TrimSpace(data))
	case !errors.Is(err, fs.ErrNotExist):
		return fmt.Errorf(
			"read key %s: %w; the agent does not replace a key it cannot read",
			path, err,
		)
	}

	seed, err := generateCurveSeedFn()
	if err != nil {
		return fmt.Errorf("generate encryption key: %w", err)
	}

	if err := m.setEncryptionKey(path, seed); err != nil {
		return err
	}

	if err := m.fs.MkdirAll(m.keyDir, keyDirMode); err != nil {
		return fmt.Errorf("create key directory: %w", err)
	}

	if err := atomicfile.Write(
		context.Background(),
		m.fs,
		path,
		append(seed, '\n'),
		privateKeyMode,
	); err != nil {
		return fmt.Errorf("write encryption key: %w", err)
	}

	return nil
}

// setEncryptionKey parses an nkeys curve seed into the manager.
func (m *Manager) setEncryptionKey(
	path string,
	seed []byte,
) error {
	kp, err := fromCurveSeedFn(seed)
	if err != nil {
		return fmt.Errorf("decode encryption key %s: %w", path, err)
	}

	pub, err := kp.PublicKey()
	if err != nil {
		return fmt.Errorf("decode encryption key %s: %w", path, err)
	}

	m.encryptionKey = kp
	m.encryptionPublicKey = pub

	return nil
}
