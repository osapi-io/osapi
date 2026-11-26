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

// Package seal keeps the secret values of a job from each reader but its
// target agents. The controller makes a new random 256-bit key for each job,
// encrypts each value with it (AES-256-GCM), and seals the job key to the
// P-256 ECDH key of each target agent. Each algorithm is FIPS 140-3
// approved, from the Go standard library.
//
// A job has one ephemeral P-256 key, From, that seals its key to each agent.
// For an agent with the public key R, the key-encryption key is
// HKDF-SHA256 of the ECDH secret of From and R, with no salt and the info
// "osapi-seal-v1" || From || R (each the 65 bytes of its uncompressed
// point). The sealed key is base64url, with no padding, of the 12-byte nonce,
// the job key encrypted with AES-256-GCM, and its 16-byte tag; the job ID is
// the additional data. A sealed value is Prefix, then base64url of the nonce,
// the value encrypted with the job key, and the tag, with the job ID as the
// additional data. Thus a value or a key that somebody copies to a different
// job does not open.
package seal

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/osapi-io/osapi/internal/pki"
	"github.com/osapi-io/osapi/internal/secret"
)

// Prefix starts each sealed value in the data of a job.
const Prefix = "sealed:"

// kdfInfo starts the HKDF info of each key-encryption key.
const kdfInfo = "osapi-seal-v1"

// keySize is the size of a job key and of a key-encryption key: AES-256.
const keySize = 32

// ErrNotSealedToAgent is the error of a job that has sealed values, but no
// job key for this agent, such as an agent that joined after the
// controller queued the job.
var ErrNotSealedToAgent = errors.New("the job was not sealed to this agent")

// encoding is the text form of a sealed value and of a sealed key.
var encoding = base64.RawURLEncoding

// Sealer seals the values of one job.
type Sealer struct {
	jobID string
	key   []byte
	aead  cipher.AEAD
	from  *ecdh.PrivateKey
}

// New is a Sealer for the job jobID, with a new random job key, and a new
// P-256 key that seals the job key to each agent.
func New(
	jobID string,
) (*Sealer, error) {
	key := make([]byte, keySize)
	_, _ = rand.Read(key)
	from, err := createKey()
	if err != nil {
		return nil, fmt.Errorf("make the key that seals the job key: %w", err)
	}

	return &Sealer{jobID: jobID, key: key, aead: gcm(key), from: from}, nil
}

// createKey makes a P-256 ECDH key. A test replaces it.
var createKey = func() (*ecdh.PrivateKey, error) {
	return ecdh.P256().GenerateKey(rand.Reader)
}

// gcm is AES-256-GCM with key, which makes a random 12-byte nonce for each
// Seal and puts it in front of the ciphertext: the form of GCM that FIPS
// 140-3 mode approves for a caller. A key of keySize bytes is always a
// valid AES key.
func gcm(
	key []byte,
) cipher.AEAD {
	block, _ := aes.NewCipher(key)
	aead, _ := cipher.NewGCMWithRandomNonce(block)

	return aead
}

// From is the public P-256 key that seals the job key, in
// pki.EncryptionKeyEncoding. The agent needs it to open the job key.
func (s *Sealer) From() string {
	return pki.EncodeEncryptionKey(s.from.PublicKey())
}

// Value is v sealed with the job key: Prefix, then the nonce, the
// ciphertext and the tag.
func (s *Sealer) Value(
	v secret.Value,
) string {
	return Prefix + encoding.EncodeToString(
		s.aead.Seal(nil, nil, []byte(v.Reveal()), []byte(s.jobID)))
}

// KeyFor is the job key sealed to recipient, the public P-256 key of an
// agent in pki.EncryptionKeyEncoding.
func (s *Sealer) KeyFor(
	recipient string,
) (string, error) {
	pub, err := pki.ParseEncryptionKey(recipient)
	if err != nil {
		return "", fmt.Errorf("seal the job key: %w", err)
	}
	kek := wrapKey(s.from, s.from.PublicKey(), pub, pub)

	return encoding.EncodeToString(gcm(kek).Seal(nil, nil, s.key, []byte(s.jobID))), nil
}

// wrapKey is the key-encryption key of the ECDH secret of own and peer, for
// the key that sender sealed to recipient. ECDH of two valid P-256 keys,
// and HKDF-SHA256 of keySize bytes, do not fail.
func wrapKey(
	own *ecdh.PrivateKey,
	sender *ecdh.PublicKey,
	recipient *ecdh.PublicKey,
	peer *ecdh.PublicKey,
) []byte {
	shared, _ := own.ECDH(peer)
	info := kdfInfo + string(sender.Bytes()) + string(recipient.Bytes())
	kek, _ := hkdf.Key(sha256.New, shared, nil, info, keySize)

	return kek
}

// Opener opens the sealed values of one job.
type Opener struct {
	jobID string
	aead  cipher.AEAD
}

// Open is the Opener of the job jobID: own, the P-256 key of the agent,
// opens sealedKey, the job key that from sealed to it.
func Open(
	jobID string,
	sealedKey string,
	from string,
	own *ecdh.PrivateKey,
) (*Opener, error) {
	raw, err := encoding.DecodeString(sealedKey)
	if err != nil {
		return nil, fmt.Errorf("open the job key: %w", err)
	}
	sender, err := pki.ParseEncryptionKey(from)
	if err != nil {
		return nil, fmt.Errorf("open the job key: %w", err)
	}
	kek := wrapKey(own, sender, own.PublicKey(), sender)
	key, err := gcm(kek).Open(nil, nil, raw, []byte(jobID))
	if err != nil || len(key) != keySize {
		return nil, errors.New("open the job key: it is not sealed to this agent for this job")
	}

	return &Opener{jobID: jobID, aead: gcm(key)}, nil
}

// Value is the value that v, a sealed value, holds.
func (o *Opener) Value(
	v secret.Value,
) (secret.Value, error) {
	text, ok := strings.CutPrefix(v.Reveal(), Prefix)
	if !ok {
		return "", errors.New("a secret of the job is not sealed")
	}
	raw, err := encoding.DecodeString(text)
	if err != nil {
		return "", errors.New("a sealed secret of the job is not valid")
	}
	plain, err := o.aead.Open(nil, nil, raw, []byte(o.jobID))
	if err != nil {
		return "", errors.New("a sealed secret of the job does not open")
	}

	return secret.Value(plain), nil
}

// openerKey is the context key of the Opener of a job.
type openerKey struct{}

// WithOpener is ctx with o, the Opener of the job that the agent runs.
func WithOpener(
	ctx context.Context,
	o *Opener,
) context.Context {
	return context.WithValue(ctx, openerKey{}, o)
}

// OpenerFrom is the Opener that WithOpener put in ctx, or nil when the job
// has no sealed value.
func OpenerFrom(
	ctx context.Context,
) *Opener {
	o, _ := ctx.Value(openerKey{}).(*Opener)

	return o
}
