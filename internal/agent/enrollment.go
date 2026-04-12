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
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/avfs/avfs"

	"github.com/retr0h/osapi/internal/agent/pki"
)

// handlePKIEnrollment manages the PKI enrollment lifecycle.
// It loads or generates a keypair, checks for existing controller key,
// and if not enrolled, the agent enters pending state.
func (a *Agent) handlePKIEnrollment(
	_ context.Context,
) error {
	keyDir := a.appConfig.Agent.PKI.KeyDir

	m := pki.New(a.appFs, keyDir, "agent")

	if err := m.LoadOrGenerate(); err != nil {
		return err
	}

	a.pkiManager = m

	a.logger.Info(
		"PKI keypair loaded",
		slog.String("fingerprint", m.Fingerprint()),
		slog.String("key_dir", keyDir),
	)

	// Check if controller public key already exists (previously enrolled).
	controllerPubPath := filepath.Join(keyDir, "controller.pub")

	pubPEM, err := avfs.ReadFile(a.appFs, controllerPubPath)
	if err == nil {
		pubKey, parseErr := pki.ParsePublicKeyPEM(pubPEM)
		if parseErr != nil {
			return fmt.Errorf("failed to parse controller public key: %w", parseErr)
		}

		m.SetControllerPublicKey(pubKey)
		a.logger.Info("controller public key loaded, agent enrolled")

		return nil
	}

	// Not enrolled yet — agent enters pending state.
	// The enrollment request will be handled by a background goroutine.
	a.logger.Info(
		"agent not enrolled, entering pending state",
		slog.String("fingerprint", m.Fingerprint()),
	)

	return nil
}
