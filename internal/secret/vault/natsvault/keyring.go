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

package natsvault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/osapi-io/osapi/internal/secret/vault"
)

// keySize is the size of a vault key: AES-256.
const keySize = 32

// ParseKeyring is the keyring in data: one key on each line, as 64 hex
// digits, the current key first. A blank line, and a line that starts with
// #, is not a key.
func ParseKeyring(
	data []byte,
) (Keyring, error) {
	var k Keyring
	for i, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if vault.IsReference(line) {
			return Keyring{}, fmt.Errorf(
				"line %d is a reference: the vault key cannot come from a vault, "+
					"because the controller needs the key to read the vault: "+
					"put the key itself here", i+1,
			)
		}
		key, err := hex.DecodeString(line)
		if err != nil || len(key) != keySize {
			return Keyring{}, fmt.Errorf(
				"line %d: a vault key is %d hex digits", i+1, keySize*2,
			)
		}
		k.keys = append(k.keys, key)
	}
	if len(k.keys) == 0 {
		return Keyring{}, errors.New("has no key")
	}

	return k, nil
}

// NewKeyring is a keyring with a new random key, then the keys of k. The
// new key is the current key.
func NewKeyring(
	k Keyring,
) Keyring {
	key := make([]byte, keySize)
	_, _ = rand.Read(key)

	return Keyring{keys: append([][]byte{key}, k.keys...)}
}

// Current is the keyring with only its current key.
func (k Keyring) Current() Keyring {
	return Keyring{keys: k.keys[:1]}
}

// ID is the ID of the current key: the first 16 hex digits of its SHA-256.
// A value records the ID of its key, and the key never leaves the host.
func (k Keyring) ID() string {
	return keyID(k.keys[0])
}

// Len is the number of keys.
func (k Keyring) Len() int {
	return len(k.keys)
}

// Marshal is the keyring as ParseKeyring reads it.
func (k Keyring) Marshal() []byte {
	var b strings.Builder
	for _, key := range k.keys {
		b.WriteString(hex.EncodeToString(key))
		b.WriteString("\n")
	}

	return []byte(b.String())
}

func keyID(
	key []byte,
) string {
	sum := sha256.Sum256(key)

	return hex.EncodeToString(sum[:8])
}

// seal encrypts value with the current key. The GCM makes a new random
// nonce and puts it in front of the data, as FIPS 140-3 only mode
// requires. aad binds the result to its vault and key, so a value copied
// to a different key does not open.
func (k Keyring) seal(
	value []byte,
	aad string,
) (keyID string, data []byte) {
	return k.ID(), newGCM(k.keys[0]).Seal(nil, nil, value, []byte(aad))
}

// open decrypts data with the key id.
func (k Keyring) open(
	id string,
	data []byte,
	aad string,
) ([]byte, error) {
	for _, key := range k.keys {
		if keyID(key) != id {
			continue
		}
		value, err := newGCM(key).Open(nil, nil, data, []byte(aad))
		if err != nil {
			return nil, fmt.Errorf("the value does not decrypt: %w", err)
		}

		return value, nil
	}

	return nil, fmt.Errorf("the value is under key %s, which the vault key does not have", id)
}

// newGCM is AES-256-GCM with key, which is keySize bytes, and a random
// nonce that it makes for each Seal and puts in front of the data.
func newGCM(
	key []byte,
) cipher.AEAD {
	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCMWithRandomNonce(block)

	return gcm
}
