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
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/avfs/avfs"

	"github.com/retr0h/osapi/internal/agent/pki"
	"github.com/retr0h/osapi/internal/job"
)

// marshalJSONEnrollment is a package-level variable for testing the marshal error path.
var marshalJSONEnrollment = json.Marshal

// handlePKIEnrollment manages the PKI enrollment lifecycle.
// It loads or generates a keypair, checks for existing controller key,
// and if not enrolled, publishes an enrollment request and sets state
// to Pending (no consumers started).
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

	// Not enrolled — publish enrollment request and enter pending state.
	if err := a.publishEnrollmentRequest(); err != nil {
		a.logger.Warn(
			"failed to publish enrollment request",
			slog.String("error", err.Error()),
		)
	}

	a.state = job.AgentStatePending
	a.logger.Info(
		"agent not enrolled, entering pending state",
		slog.String("fingerprint", m.Fingerprint()),
	)

	return nil
}

// publishEnrollmentRequest sends the agent's enrollment request to the
// controller via NATS. The controller's enrollment watcher picks it up
// and stores it in the enrollment KV bucket.
func (a *Agent) publishEnrollmentRequest() error {
	if a.natsClient == nil {
		return fmt.Errorf("NATS client not available")
	}

	req := pki.EnrollmentRequest{
		MachineID:   a.machineID,
		Hostname:    a.hostname,
		PublicKey:   a.pkiManager.PublicKey(),
		Fingerprint: a.pkiManager.Fingerprint(),
	}

	data, err := marshalJSONEnrollment(req)
	if err != nil {
		return fmt.Errorf("marshal enrollment request: %w", err)
	}

	namespace := a.appConfig.Agent.NATS.Namespace
	subject := "enroll.request"
	if namespace != "" {
		subject = namespace + "." + subject
	}

	if err := a.natsClient.Publish(context.Background(), subject, data); err != nil {
		return fmt.Errorf("publish enrollment request: %w", err)
	}

	a.logger.Info(
		"enrollment request published",
		slog.String("subject", subject),
		slog.String("machine_id", a.machineID),
		slog.String("fingerprint", a.pkiManager.Fingerprint()),
	)

	return nil
}
