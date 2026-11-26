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

// VerifyWithGrace checks the signature against each controller key that the
// manager accepts: the current one, the previous one, which signed the jobs
// that a rotation of the controller's key left in the queue, and the next
// one, which the controller signs with once it switches.
func (m *Manager) VerifyWithGrace(
	data []byte,
	signature []byte,
) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, key := range []ed25519.PublicKey{
		m.controllerPubKey, m.previousControllerPubKey, m.nextControllerPubKey,
	} {
		if len(key) == ed25519.PublicKeySize && ed25519.Verify(key, data, signature) {
			return true
		}
	}

	return false
}

// RotateControllerKey makes newKey the controller key: the current one
// becomes the previous one, which still verifies, and the next one is
// cleared.
func (m *Manager) RotateControllerKey(
	newKey ed25519.PublicKey,
) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.previousControllerPubKey = m.controllerPubKey
	m.controllerPubKey = newKey
	m.nextControllerPubKey = nil
}

// PreviousControllerPublicKey returns the previous controller public
// key, if any. This is the key that was current before the last
// rotation.
func (m *Manager) PreviousControllerPublicKey() ed25519.PublicKey {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.previousControllerPubKey
}

// SetPreviousControllerPublicKey sets the previous controller key, such as
// the one that a controller read from its rotation record at its start.
func (m *Manager) SetPreviousControllerPublicKey(
	key ed25519.PublicKey,
) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.previousControllerPubKey = key
}

// NextControllerPublicKey is the key that the controller switches to, as
// its rotation record gave it, or nil.
func (m *Manager) NextControllerPublicKey() ed25519.PublicKey {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.nextControllerPubKey
}

// SetNextControllerPublicKey sets the key that the controller switches to.
// Nil clears it.
func (m *Manager) SetNextControllerPublicKey(
	key ed25519.PublicKey,
) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextControllerPubKey = key
}
