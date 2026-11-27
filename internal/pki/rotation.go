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
	"time"
)

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

// SetRetiredControllerKeys sets the keys that switches before the last one
// replaced, from the controller's key record, and how long after it
// retired each one still verifies a stored record.
func (m *Manager) SetRetiredControllerKeys(
	keys []RetiredKey,
	lifetime time.Duration,
) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.retiredControllerKeys = keys
	m.retiredKeyLifetime = lifetime
}

// StoredRecordKeys is each controller key that verifies a record that the
// controller stored, such as a job in the queue: the current, the next, the
// previous and each retired key that has not expired, in that order. A
// message from the network is verified with VerifyWithGrace, which takes no
// retired key.
func (m *Manager) StoredRecordKeys() []ed25519.PublicKey {
	m.mu.RLock()
	defer m.mu.RUnlock()
	now := time.Now()
	keys := make([]ed25519.PublicKey, 0, 3+len(m.retiredControllerKeys))
	for _, k := range []ed25519.PublicKey{
		m.controllerPubKey, m.nextControllerPubKey, m.previousControllerPubKey,
	} {
		if len(k) == ed25519.PublicKeySize {
			keys = append(keys, k)
		}
	}
	for _, k := range m.retiredControllerKeys {
		if len(k.Key) == ed25519.PublicKeySize && !k.Expired(now, m.retiredKeyLifetime) {
			keys = append(keys, k.Key)
		}
	}

	return keys
}
