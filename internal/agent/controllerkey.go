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

package agent

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/pem"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"

	"github.com/avfs/avfs"

	"github.com/osapi-io/osapi/internal/atomicfile"
	"github.com/osapi-io/osapi/internal/pki"
)

// EnrolledControllerNextKey is the file in agent.pki.key_dir of the key that
// the controller switches to, during a rotation of its key. The agent keeps
// it beside EnrolledControllerKey, so that it still has it after a restart.
const EnrolledControllerNextKey = "enrolled-controller.next.pub"

// loadNextControllerKey loads the next controller key that an earlier run
// saved, when there is one.
func (a *Agent) loadNextControllerKey(
	keyDir string,
) error {
	pubPEM, err := avfs.ReadFile(a.appFs, filepath.Join(keyDir, EnrolledControllerNextKey))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read %s: %w", EnrolledControllerNextKey, err)
	}
	key, err := pki.ParsePublicKeyPEM(pubPEM)
	if err != nil {
		return fmt.Errorf("parse %s: %w", EnrolledControllerNextKey, err)
	}
	a.pkiManager.SetNextControllerPublicKey(key)

	return nil
}

// followControllerKey follows the rotation of the controller's key, from
// the record in the state bucket, at each heartbeat:
//
//   - a record that the current key signed, with a next key, is a rotation
//     that started: the agent keeps the next key, and accepts jobs that it
//     signs, and the registration reports it, which the controller waits for
//   - a record that the next key signed, with it as the current key, is the
//     switch: the next key becomes the current one, and the old one stays
//     as the previous one, which verifies the jobs that it signed before
//
// A record that no key of the agent signed changes nothing: an agent that
// missed the whole rotation needs a new admission, and the log says so.
func (a *Agent) followControllerKey(
	ctx context.Context,
) {
	current := a.pkiManager.ControllerPublicKey()
	if len(current) == 0 {
		return
	}
	rec, ok, err := a.jobClient.ControllerKey(ctx)
	if err != nil {
		a.pkiLogger.Warn(
			"failed to read the controller key record",
			slog.String("error", err.Error()),
		)

		return
	}
	if !ok {
		return
	}
	next := a.pkiManager.NextControllerPublicKey()
	switch {
	case bytes.Equal(rec.Current, current) && rec.SignedBy(current):
		a.takeNextControllerKey(ctx, rec, next)
	case len(next) > 0 && bytes.Equal(rec.Current, next) && rec.SignedBy(next):
		a.switchControllerKey(ctx, next)
	case !bytes.Equal(rec.Current, current):
		a.pkiLogger.Warn(
			"the controller signs with a key that this agent does not have: it "+
				"rotated its key while the agent was away; admit the machine again",
			slog.String("controller_key", pki.FingerprintOf(rec.Current)),
			slog.String("agent_has", pki.FingerprintOf(current)),
		)
	}
}

// takeNextControllerKey keeps the next key of rec, or drops a next key that
// the record no longer has: a rotation that a later record replaced.
func (a *Agent) takeNextControllerKey(
	ctx context.Context,
	rec pki.ControllerKeyRecord,
	next ed25519.PublicKey,
) {
	keyDir := a.appConfig.Agent.PKI.KeyDir
	path := filepath.Join(keyDir, EnrolledControllerNextKey)
	if !rec.Rotating() {
		if len(next) == 0 {
			return
		}
		if err := a.appFs.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
			a.pkiLogger.Warn(
				"failed to remove the next controller key",
				slog.String("error", err.Error()),
			)

			return
		}
		a.pkiManager.SetNextControllerPublicKey(nil)

		return
	}
	if bytes.Equal(rec.Next, next) {
		return
	}
	if err := writeControllerKey(ctx, a.appFs, path, rec.Next); err != nil {
		a.pkiLogger.Warn(
			"failed to keep the next controller key",
			slog.String("error", err.Error()),
		)

		return
	}
	a.pkiManager.SetNextControllerPublicKey(rec.Next)
	a.pkiLogger.Info(
		"the controller rotates its key: keeping its next key",
		slog.String("next", pki.FingerprintOf(rec.Next)),
	)
}

// switchControllerKey makes next the controller key, after the controller
// switched to it.
func (a *Agent) switchControllerKey(
	ctx context.Context,
	next ed25519.PublicKey,
) {
	keyDir := a.appConfig.Agent.PKI.KeyDir
	if err := writeControllerKey(
		ctx,
		a.appFs,
		filepath.Join(keyDir, EnrolledControllerKey),
		next,
	); err != nil {
		a.pkiLogger.Warn(
			"failed to keep the controller's new key",
			slog.String("error", err.Error()),
		)

		return
	}
	a.pkiManager.RotateControllerKey(next)
	if err := a.appFs.Remove(filepath.Join(keyDir, EnrolledControllerNextKey)); err != nil &&
		!errors.Is(err, fs.ErrNotExist) {
		a.pkiLogger.Warn(
			"failed to remove the next controller key",
			slog.String("error", err.Error()),
		)
	}
	a.pkiLogger.Info(
		"the controller switched to its new key",
		slog.String("controller_key", pki.FingerprintOf(next)),
	)
}

// controllerKeys are the fingerprints of the controller keys that the agent
// has, as its registration reports them: the current one, and the next one
// during a rotation.
func (a *Agent) controllerKeys() (string, string) {
	if a.pkiManager == nil {
		return "", ""
	}
	current := a.pkiManager.ControllerPublicKey()
	if len(current) == 0 {
		return "", ""
	}
	next := ""
	if n := a.pkiManager.NextControllerPublicKey(); len(n) > 0 {
		next = pki.FingerprintOf(n)
	}

	return pki.FingerprintOf(current), next
}

// writeControllerKey writes key to path as a PEM public key.
func writeControllerKey(
	ctx context.Context,
	fsys avfs.VFS,
	path string,
	key ed25519.PublicKey,
) error {
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "ED25519 PUBLIC KEY", Bytes: key})
	if err := atomicfile.Write(ctx, fsys, path, pubPEM, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	return nil
}
