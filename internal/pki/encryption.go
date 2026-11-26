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
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/osapi-io/osapi/internal/atomicfile"
)

// The encryption key of an agent is an ECDH key on P-256, a FIPS 140-3
// approved curve. Its private key is in {prefix}.ekey as a PKCS #8 PEM
// block. Its public key, in the join request and the admission record, is
// EncryptionKeyEncoding of the 65 bytes of the uncompressed point.

// EncryptionKeyEncoding is the text form of a public encryption key.
var EncryptionKeyEncoding = base64.RawURLEncoding

// encryptionKeySuffix is the file of the encryption key, after the prefix.
const encryptionKeySuffix = ".ekey"

// oldEncryptionKeySuffix is the file of the X25519 key that an agent had
// before. Its admission does not hold the new key, so the agent joins again.
const oldEncryptionKeySuffix = ".xkey"

// generateEncryptionKey makes a new P-256 ECDH key.
func generateEncryptionKey() (*ecdh.PrivateKey, error) {
	return ecdh.P256().GenerateKey(rand.Reader)
}

// generateEncryptionKeyFn makes the encryption key. Overridable in tests via
// export_test.go.
var generateEncryptionKeyFn = generateEncryptionKey

// marshalEncryptionKeyFn is the PKCS #8 form of a key. Overridable in tests
// via export_test.go.
var marshalEncryptionKeyFn = x509.MarshalPKCS8PrivateKey

// EncryptionPublicKey returns the public P-256 key in EncryptionKeyEncoding,
// or an empty string when the manager has none.
func (m *Manager) EncryptionPublicKey() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.encryptionPublicKey
}

// EncryptionKey returns the P-256 ECDH key, or nil when the manager has none.
func (m *Manager) EncryptionKey() *ecdh.PrivateKey {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.encryptionKey
}

// PreviousEncryptionKey returns the encryption key that the last Reload
// replaced, or nil when no Reload replaced one.
func (m *Manager) PreviousEncryptionKey() *ecdh.PrivateKey {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.previousEncryptionKey
}

// ParseEncryptionKey is the public P-256 key that key holds in
// EncryptionKeyEncoding.
func ParseEncryptionKey(
	key string,
) (*ecdh.PublicKey, error) {
	raw, err := EncryptionKeyEncoding.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("encryption key: %w", err)
	}
	pub, err := ecdh.P256().NewPublicKey(raw)
	if err != nil {
		return nil, fmt.Errorf("encryption key: %w", err)
	}

	return pub, nil
}

// EncodeEncryptionKey is pub in EncryptionKeyEncoding.
func EncodeEncryptionKey(
	pub *ecdh.PublicKey,
) string {
	return EncryptionKeyEncoding.EncodeToString(pub.Bytes())
}

// ValidEncryptionKey reports whether key is a public P-256 key in
// EncryptionKeyEncoding.
func ValidEncryptionKey(
	key string,
) bool {
	_, err := ParseEncryptionKey(key)

	return err == nil
}

// FingerprintOfKeys returns the fingerprint that covers an agent's two public
// keys: the SHA-256 of the 32 bytes of the Ed25519 key, then the 65 bytes of
// the P-256 key, as "SHA256:<hex>". An operator who admits it admits both
// keys. It is empty when the signing key is empty or the encryption key is
// not a public P-256 key, so a record without both keys never reports a
// fingerprint that admits it.
func FingerprintOfKeys(
	signingKey ed25519.PublicKey,
	encryptionKey string,
) string {
	if len(signingKey) == 0 {
		return ""
	}
	pub, err := ParseEncryptionKey(encryptionKey)
	if err != nil {
		return ""
	}

	h := sha256.New()
	_, _ = h.Write(signingKey)
	_, _ = h.Write(pub.Bytes())

	return "SHA256:" + hex.EncodeToString(h.Sum(nil))
}

// loadOrGenerateEncryptionKey loads the P-256 key from {prefix}.ekey, or
// makes and saves one when the file does not exist, and then removes the
// X25519 key of an older agent, {prefix}.xkey. A file it cannot read or
// parse is an error, as for the signing key: a new key is not the admitted
// one.
func (m *Manager) loadOrGenerateEncryptionKey() error {
	path := filepath.Join(m.keyDir, m.keyPrefix+encryptionKeySuffix)

	data, err := m.fs.ReadFile(path)
	switch {
	case err == nil:
		return m.setEncryptionKey(path, data)
	case !errors.Is(err, fs.ErrNotExist):
		return fmt.Errorf(
			"read key %s: %w; the agent does not replace a key it cannot read",
			path, err,
		)
	}

	key, err := generateEncryptionKeyFn()
	if err != nil {
		return fmt.Errorf("generate encryption key: %w", err)
	}
	der, err := marshalEncryptionKeyFn(key)
	if err != nil {
		return fmt.Errorf("encode encryption key: %w", err)
	}
	if err := m.fs.MkdirAll(m.keyDir, keyDirMode); err != nil {
		return fmt.Errorf("create key directory: %w", err)
	}
	if err := atomicfile.Write(
		context.Background(),
		m.fs,
		path,
		pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}),
		privateKeyMode,
	); err != nil {
		return fmt.Errorf("write encryption key: %w", err)
	}
	old := filepath.Join(m.keyDir, m.keyPrefix+oldEncryptionKeySuffix)
	if err := m.fs.Remove(old); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("remove the old encryption key %s: %w", old, err)
	}
	m.encryptionKey = key
	m.encryptionPublicKey = EncodeEncryptionKey(key.PublicKey())

	return nil
}

// ParseEncryptionKeyFile is the P-256 key in data, the PKCS #8 PEM block of
// an {prefix}.ekey file.
func ParseEncryptionKeyFile(
	data []byte,
) (*ecdh.PrivateKey, error) {
	block, _ := pem.Decode(bytes.TrimSpace(data))
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, errors.New("no PRIVATE KEY PEM block")
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	ec, ok := parsed.(*ecdsa.PrivateKey)
	if !ok || ec.Curve != elliptic.P256() {
		return nil, errors.New("not a P-256 key")
	}
	// A P-256 ecdsa key is always a valid ecdh key.
	key, _ := ec.ECDH()

	return key, nil
}

// setEncryptionKey parses a PKCS #8 PEM block of a P-256 key into the
// manager.
func (m *Manager) setEncryptionKey(
	path string,
	data []byte,
) error {
	key, err := ParseEncryptionKeyFile(data)
	if err != nil {
		return fmt.Errorf("decode encryption key %s: %w", path, err)
	}
	m.encryptionKey = key
	m.encryptionPublicKey = EncodeEncryptionKey(key.PublicKey())

	return nil
}
