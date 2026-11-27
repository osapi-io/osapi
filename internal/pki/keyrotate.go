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
	"crypto/ecdh"
	"crypto/ed25519"
	"encoding/pem"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sync"
)

// nextSuffix follows the prefix of the keys that a rotation makes before
// the controller takes them: agent.next.key, agent.next.pub and
// agent.next.ekey.
const nextSuffix = ".next"

// ErrNoKeys is a Load of keys that are not on the host.
var ErrNoKeys = errors.New("the keys are not on this host")

// paths are the files of the manager's keys: the private and public signing
// key, and the encryption key.
func (m *Manager) paths() (string, string, string) {
	return m.pathsOf(m.keyPrefix)
}

// pathsOf are the files of the keys of prefix in the manager's directory.
func (m *Manager) pathsOf(
	prefix string,
) (string, string, string) {
	base := filepath.Join(m.keyDir, prefix)

	return base + ".key", base + ".pub", base + encryptionKeySuffix
}

// moves are the renames of a Promote of the keys in from, in their order:
// the encryption key first and the public key last.
func (m *Manager) moves(
	fromPriv, fromPub, fromEkey string,
) [][2]string {
	toPriv, toPub, toEkey := m.paths()
	moves := [][2]string{{fromPriv, toPriv}, {fromPub, toPub}}
	if m.withEncryption {
		moves = append([][2]string{{fromEkey, toEkey}}, moves...)
	}

	return moves
}

// Load loads the manager's keys, which must be on the host: it makes none.
// A key that is missing is ErrNoKeys.
func (m *Manager) Load() error {
	pub, priv, ekey, err := m.read()
	if err != nil {
		return err
	}
	m.set(pub, priv, ekey)

	return nil
}

// read is the manager's keys on the host. The private key must be the pair
// of the public one: a rotation that is halfway through its renames is not
// a pair, and is an error.
func (m *Manager) read() (ed25519.PublicKey, ed25519.PrivateKey, *ecdh.PrivateKey, error) {
	privPath, pubPath, ekeyPath := m.paths()
	want := []string{privPath, pubPath}
	if m.withEncryption {
		want = append(want, ekeyPath)
	}
	data := make([][]byte, 0, len(want))
	for _, p := range want {
		b, err := m.fs.ReadFile(p)
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil, nil, fmt.Errorf("%w: %s", ErrNoKeys, p)
		}
		if err != nil {
			return nil, nil, nil, fmt.Errorf("read key %s: %w", p, err)
		}
		data = append(data, b)
	}
	keys := &Manager{}
	if err := keys.loadKeys(data[0], data[1]); err != nil {
		return nil, nil, nil, err
	}
	if !bytes.Equal(keys.privateKey.Public().(ed25519.PublicKey), keys.publicKey) {
		return nil, nil, nil, fmt.Errorf("%s is not the pair of %s", privPath, pubPath)
	}
	var ekey *ecdh.PrivateKey
	if m.withEncryption {
		k, err := ParseEncryptionKeyFile(data[2])
		if err != nil {
			return nil, nil, nil, fmt.Errorf("decode encryption key %s: %w", ekeyPath, err)
		}
		ekey = k
	}

	return keys.publicKey, keys.privateKey, ekey, nil
}

// set makes pub, priv and ekey the manager's keys.
func (m *Manager) set(
	pub ed25519.PublicKey,
	priv ed25519.PrivateKey,
	ekey *ecdh.PrivateKey,
) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if ekey != nil && m.encryptionKey != nil && !m.encryptionKey.Equal(ekey) {
		m.previousEncryptionKey = m.encryptionKey
	}
	m.publicKey, m.privateKey = pub, priv
	if ekey != nil {
		m.encryptionKey = ekey
		m.encryptionPublicKey = EncodeEncryptionKey(ekey.PublicKey())
	}
}

// Reload reads the manager's keys again, and reports whether they changed:
// a rotation replaced them while the process ran. The keys stay as they
// were when the files are not a whole set, such as during the renames of a
// rotation; a later Reload takes them.
func (m *Manager) Reload() (bool, error) {
	pub, priv, ekey, err := m.read()
	if err != nil {
		return false, err
	}
	m.mu.RLock()
	same := bytes.Equal(pub, m.publicKey) &&
		(ekey == nil || (m.encryptionKey != nil && m.encryptionKey.Equal(ekey)))
	m.mu.RUnlock()
	if same {
		return false, nil
	}
	m.set(pub, priv, ekey)

	return true, nil
}

// Next is the manager of the keys that a rotation makes, beside the
// manager's own: the ones that an earlier rotation made and the controller
// did not take, or new ones.
func (m *Manager) Next() (*Manager, error) {
	next := &Manager{
		mu:             &sync.RWMutex{},
		fs:             m.fs,
		keyDir:         m.keyDir,
		keyPrefix:      m.keyPrefix + nextSuffix,
		withEncryption: m.withEncryption,
	}
	if err := next.LoadOrGenerate(); err != nil {
		return nil, err
	}

	return next, nil
}

// Promote moves the keys of next, from Next, in place of the manager's own,
// and loads them. The encryption key moves first and the public key last,
// so a Reload during the renames sees the old pair or the new one.
func (m *Manager) Promote(
	next *Manager,
) error {
	for _, mv := range m.moves(next.paths()) {
		if err := m.fs.Rename(mv[0], mv[1]); err != nil {
			return fmt.Errorf("move %s to %s: %w", mv[0], mv[1], err)
		}
	}
	m.set(next.PublicKey(), next.PrivateKey(), next.EncryptionKey())

	return nil
}

// Resume finishes a Promote that stopped between its renames, as at a
// crash, and loads the keys: want is the public key of the next keys. It
// reports whether it moved a key. When no rename of a Promote is done, or
// each is, it does nothing.
//
// A Promote that stopped leaves the private key of one pair beside the
// public key of the other, and the manager signs with a key that its public
// key does not verify, so each signature fails until the renames finish.
func (m *Manager) Resume(
	want ed25519.PublicKey,
) (bool, error) {
	fromPriv, fromPub, fromEkey := m.pathsOf(m.keyPrefix + nextSuffix)
	moves := m.moves(fromPriv, fromPub, fromEkey)
	var left [][2]string
	for _, mv := range moves {
		_, err := m.fs.Stat(mv[0])
		if err == nil {
			left = append(left, mv)

			continue
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return false, fmt.Errorf("read key %s: %w", mv[0], err)
		}
		if len(left) > 0 {
			// A key that moved after one that did not: not a Promote.
			return false, nil
		}
	}
	if len(left) == 0 || len(left) == len(moves) {
		return false, nil
	}
	toPriv, toPub, _ := m.paths()
	privPath, pubPath := toPriv, toPub
	for _, mv := range left {
		switch mv[0] {
		case fromPriv:
			privPath = fromPriv
		case fromPub:
			pubPath = fromPub
		}
	}
	if !m.pairOf(privPath, pubPath, want) {
		return false, fmt.Errorf(
			"%s and %s are halfway through a rotation, and are not the pair of the next key",
			privPath, pubPath,
		)
	}
	for _, mv := range left {
		if err := m.fs.Rename(mv[0], mv[1]); err != nil {
			return false, fmt.Errorf("move %s to %s: %w", mv[0], mv[1], err)
		}
	}

	return true, m.Load()
}

// ResumeNext is Resume for a host that keeps the next public key itself, as
// an agent does: the public key of the next keys is in its next public key
// file until the last rename of a Promote moves it. With no next public key
// file there is nothing to finish.
func (m *Manager) ResumeNext() (bool, error) {
	_, fromPub, _ := m.pathsOf(m.keyPrefix + nextSuffix)
	data, err := m.fs.ReadFile(fromPub)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read key %s: %w", fromPub, err)
	}
	block, _ := pem.Decode(data)
	if block == nil || block.Type != pemPublicKeyType {
		return false, fmt.Errorf("decode key %s: no public key PEM block", fromPub)
	}

	return m.Resume(ed25519.PublicKey(block.Bytes))
}

// pairOf is whether the keys in privPath and pubPath are a pair, of want.
func (m *Manager) pairOf(
	privPath string,
	pubPath string,
	want ed25519.PublicKey,
) bool {
	priv, err := m.fs.ReadFile(privPath)
	if err != nil {
		return false
	}
	pub, err := m.fs.ReadFile(pubPath)
	if err != nil {
		return false
	}
	keys := &Manager{}
	if err := keys.loadKeys(priv, pub); err != nil {
		return false
	}

	return bytes.Equal(keys.publicKey, want) &&
		bytes.Equal(keys.privateKey.Public().(ed25519.PublicKey), want)
}

// Drop removes the manager's keys from the host: the keys of a rotation
// that the controller refused.
func (m *Manager) Drop() error {
	privPath, pubPath, ekeyPath := m.paths()
	for _, p := range []string{privPath, pubPath, ekeyPath} {
		if err := m.fs.Remove(p); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("remove %s: %w", p, err)
		}
	}

	return nil
}
