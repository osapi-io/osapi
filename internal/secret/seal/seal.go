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
// target agents. The controller makes a new random key for each job,
// encrypts each value with it (XChaCha20-Poly1305), and seals the key to
// the X25519 key of each target agent (NaCl box, as nkeys xkeys gives it).
// The agent opens the job key with its own X25519 key, then each value.
package seal

import (
	"context"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/nats-io/nkeys"
	"golang.org/x/crypto/chacha20poly1305"

	"github.com/osapi-io/osapi/internal/secret"
)

// Prefix starts each sealed value in the data of a job.
const Prefix = "sealed:"

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
	from  nkeys.KeyPair
}

// New is a Sealer for the job jobID, with a new random job key, and a new
// X25519 key that seals the job key to each agent. The job ID is the
// additional data of each value, so a value that somebody copies to a
// different job does not open.
func New(
	jobID string,
) (*Sealer, error) {
	key := make([]byte, chacha20poly1305.KeySize)
	_, _ = rand.Read(key)
	from, err := createCurveKeys()
	if err != nil {
		return nil, fmt.Errorf("make the key that seals the job key: %w", err)
	}
	// A key of KeySize bytes is a key that NewX takes.
	aead, _ := chacha20poly1305.NewX(key)

	return &Sealer{jobID: jobID, key: key, aead: aead, from: from}, nil
}

// createCurveKeys makes an X25519 key. A test replaces it.
var createCurveKeys = nkeys.CreateCurveKeys

// From is the public X25519 key that seals the job key. The agent needs it
// to open the job key.
func (s *Sealer) From() string {
	// A curve key pair always has a public key.
	pub, _ := s.from.PublicKey()

	return pub
}

// Value is v sealed with the job key: Prefix, then the nonce and the
// ciphertext.
func (s *Sealer) Value(
	v secret.Value,
) string {
	nonce := make([]byte, s.aead.NonceSize(), s.aead.NonceSize()+len(v)+s.aead.Overhead())
	_, _ = rand.Read(nonce)

	return Prefix + encoding.EncodeToString(
		s.aead.Seal(nonce, nonce, []byte(v.Reveal()), []byte(s.jobID)))
}

// KeyFor is the job key sealed to recipient, the public X25519 key of an
// agent in the nkeys form.
func (s *Sealer) KeyFor(
	recipient string,
) (string, error) {
	sealed, err := s.from.Seal(s.key, recipient)
	if err != nil {
		return "", fmt.Errorf("seal the job key: %w", err)
	}

	return encoding.EncodeToString(sealed), nil
}

// Opener opens the sealed values of one job.
type Opener struct {
	jobID string
	aead  cipher.AEAD
}

// Open is the Opener of the job jobID: own, the X25519 key of the agent,
// opens sealedKey, the job key that from sealed to it.
func Open(
	jobID string,
	sealedKey string,
	from string,
	own nkeys.KeyPair,
) (*Opener, error) {
	raw, err := encoding.DecodeString(sealedKey)
	if err != nil {
		return nil, fmt.Errorf("open the job key: %w", err)
	}
	key, err := own.Open(raw, from)
	if err != nil {
		return nil, fmt.Errorf("open the job key: %w", err)
	}
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, fmt.Errorf("open the job key: %w", err)
	}

	return &Opener{jobID: jobID, aead: aead}, nil
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
	if err != nil || len(raw) < o.aead.NonceSize() {
		return "", errors.New("a sealed secret of the job is not valid")
	}
	nonce, ciphertext := raw[:o.aead.NonceSize()], raw[o.aead.NonceSize():]
	plain, err := o.aead.Open(nil, nonce, ciphertext, []byte(o.jobID))
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
