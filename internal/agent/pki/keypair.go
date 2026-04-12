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
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"path/filepath"

	"github.com/avfs/avfs"
)

const (
	pemPrivateKeyType = "ED25519 PRIVATE KEY"
	pemPublicKeyType  = "ED25519 PUBLIC KEY"

	keyDirMode     = 0o700
	privateKeyMode = 0o600
	publicKeyMode  = 0o644
)

// generateKeyPairFn is the function used to generate Ed25519 keypairs.
// Overridable in tests via export_test.go.
var generateKeyPairFn = defaultGenerateKeyPair

// defaultGenerateKeyPair generates a new Ed25519 keypair using crypto/rand.
func defaultGenerateKeyPair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	return ed25519.GenerateKey(rand.Reader)
}

// New creates a PKI Manager with the given filesystem, key directory path,
// and key prefix (e.g., "agent" or "controller"). The prefix determines
// the key filenames: {prefix}.key and {prefix}.pub.
func New(
	fs avfs.VFS,
	keyDir string,
	keyPrefix string,
) *Manager {
	return &Manager{
		fs:        fs,
		keyDir:    keyDir,
		keyPrefix: keyPrefix,
	}
}

// LoadOrGenerate loads existing PEM-encoded Ed25519 keys from disk,
// or generates and saves a new keypair if none exists.
func (m *Manager) LoadOrGenerate() error {
	privPath := filepath.Join(m.keyDir, m.keyPrefix+".key")
	pubPath := filepath.Join(m.keyDir, m.keyPrefix+".pub")

	privData, privErr := m.fs.ReadFile(privPath)
	pubData, pubErr := m.fs.ReadFile(pubPath)

	if privErr == nil && pubErr == nil {
		return m.loadKeys(privData, pubData)
	}

	pub, priv, err := generateKeyPairFn()
	if err != nil {
		return fmt.Errorf("generate keypair: %w", err)
	}

	m.publicKey = pub
	m.privateKey = priv

	return m.saveKeys(privPath, pubPath)
}

// PublicKey returns the Ed25519 public key.
func (m *Manager) PublicKey() ed25519.PublicKey {
	return m.publicKey
}

// PrivateKey returns the Ed25519 private key.
func (m *Manager) PrivateKey() ed25519.PrivateKey {
	return m.privateKey
}

// Fingerprint returns the SHA256 fingerprint of the public key in the
// format "SHA256:<hex>". Returns an empty string if no public key is set.
func (m *Manager) Fingerprint() string {
	if len(m.publicKey) == 0 {
		return ""
	}

	hash := sha256.Sum256(m.publicKey)

	return "SHA256:" + hex.EncodeToString(hash[:])
}

// Sign signs the given data with the manager's private key and returns
// the signature.
func (m *Manager) Sign(
	data []byte,
) []byte {
	return ed25519.Sign(m.privateKey, data)
}

// Verify verifies the signature of the given data using the provided
// public key.
func Verify(
	pubKey ed25519.PublicKey,
	data []byte,
	signature []byte,
) bool {
	return ed25519.Verify(pubKey, data, signature)
}

// SetControllerPublicKey stores the controller's public key for
// verifying controller-signed messages.
func (m *Manager) SetControllerPublicKey(
	key ed25519.PublicKey,
) {
	m.controllerPubKey = key
}

// ControllerPublicKey returns the stored controller public key.
func (m *Manager) ControllerPublicKey() ed25519.PublicKey {
	return m.controllerPubKey
}

// ParsePublicKeyPEM parses a PEM-encoded Ed25519 public key.
func ParsePublicKeyPEM(
	pemData []byte,
) (ed25519.PublicKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	if block.Type != pemPublicKeyType {
		return nil, fmt.Errorf(
			"unexpected block type %q, expected %q",
			block.Type, pemPublicKeyType,
		)
	}

	return ed25519.PublicKey(block.Bytes), nil
}

// loadKeys parses PEM-encoded private and public key data into the
// manager's key fields.
func (m *Manager) loadKeys(
	privData []byte,
	pubData []byte,
) error {
	privBlock, _ := pem.Decode(privData)
	if privBlock == nil {
		return fmt.Errorf("decode private key PEM: no valid block found")
	}

	if privBlock.Type != pemPrivateKeyType {
		return fmt.Errorf(
			"decode private key PEM: unexpected block type %q",
			privBlock.Type,
		)
	}

	pubBlock, _ := pem.Decode(pubData)
	if pubBlock == nil {
		return fmt.Errorf("decode public key PEM: no valid block found")
	}

	if pubBlock.Type != pemPublicKeyType {
		return fmt.Errorf(
			"decode public key PEM: unexpected block type %q",
			pubBlock.Type,
		)
	}

	m.privateKey = ed25519.NewKeyFromSeed(privBlock.Bytes)
	m.publicKey = ed25519.PublicKey(pubBlock.Bytes)

	return nil
}

// saveKeys writes PEM-encoded private and public keys to disk,
// creating the key directory if needed.
func (m *Manager) saveKeys(
	privPath string,
	pubPath string,
) error {
	if err := m.fs.MkdirAll(m.keyDir, keyDirMode); err != nil {
		return fmt.Errorf("create key directory: %w", err)
	}

	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  pemPrivateKeyType,
		Bytes: m.privateKey.Seed(),
	})

	if err := m.fs.WriteFile(privPath, privPEM, privateKeyMode); err != nil {
		return fmt.Errorf("write private key: %w", err)
	}

	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  pemPublicKeyType,
		Bytes: m.publicKey,
	})

	if err := m.fs.WriteFile(pubPath, pubPEM, publicKeyMode); err != nil {
		return fmt.Errorf("write public key: %w", err)
	}

	return nil
}
