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

import "crypto/ed25519"

// VerifyWithGrace checks the signature against both the current and
// previous controller public keys. Returns true if either key
// validates the signature, supporting graceful key rotation.
func (m *Manager) VerifyWithGrace(
	data []byte,
	signature []byte,
) bool {
	if m.controllerPubKey != nil {
		if ed25519.Verify(m.controllerPubKey, data, signature) {
			return true
		}
	}

	if m.previousControllerPubKey != nil {
		if ed25519.Verify(m.previousControllerPubKey, data, signature) {
			return true
		}
	}

	return false
}

// RotateControllerKey stores the new controller key and moves the
// current one to previous, enabling a grace period where both keys
// are accepted.
func (m *Manager) RotateControllerKey(
	newKey ed25519.PublicKey,
) {
	m.previousControllerPubKey = m.controllerPubKey
	m.controllerPubKey = newKey
}

// PreviousControllerPublicKey returns the previous controller public
// key, if any. This is the key that was current before the last
// rotation.
func (m *Manager) PreviousControllerPublicKey() ed25519.PublicKey {
	return m.previousControllerPubKey
}
