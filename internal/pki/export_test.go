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
	"io"

	"github.com/nats-io/nkeys"
)

// SetGenerateKeyPairFn overrides the generateKeyPairFn function for testing.
func SetGenerateKeyPairFn(
	fn func() (ed25519.PublicKey, ed25519.PrivateKey, error),
) {
	generateKeyPairFn = fn
}

// ResetGenerateKeyPairFn restores the default generateKeyPairFn function.
func ResetGenerateKeyPairFn() {
	generateKeyPairFn = defaultGenerateKeyPair
}

// SetCurveRand overrides the entropy of a new X25519 key for testing.
func SetCurveRand(
	r io.Reader,
) {
	curveRand = r
}

// SetGenerateCurveSeedFn overrides the generateCurveSeedFn function for testing.
func SetGenerateCurveSeedFn(
	fn func() ([]byte, error),
) {
	generateCurveSeedFn = fn
}

// SetFromCurveSeedFn overrides the fromCurveSeedFn function for testing.
func SetFromCurveSeedFn(
	fn func([]byte) (nkeys.KeyPair, error),
) {
	fromCurveSeedFn = fn
}

// ResetCurveFns restores the defaults of the X25519 test hooks.
func ResetCurveFns() {
	curveRand = rand.Reader
	generateCurveSeedFn = generateCurveSeed
	fromCurveSeedFn = nkeys.FromCurveSeed
}
