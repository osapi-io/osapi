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
	"crypto/ed25519"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/avfs/avfs"
	"github.com/nats-io/nats.go"

	agentpki "github.com/retr0h/osapi/internal/agent/pki"
	"github.com/retr0h/osapi/internal/job"
)

// marshalJSONEnrollment is a package-level variable for testing the marshal error path.
var marshalJSONEnrollment = json.Marshal

// handlePKIEnrollment manages the PKI enrollment lifecycle.
// It loads or generates a keypair, checks for existing controller key,
// and if not enrolled, publishes an enrollment request, starts a
// background listener for acceptance, and sets state to Pending.
func (a *Agent) handlePKIEnrollment(
	ctx context.Context,
) error {
	keyDir := a.appConfig.Agent.PKI.KeyDir

	m := agentpki.New(a.appFs, keyDir, "agent")

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
		pubKey, parseErr := agentpki.ParsePublicKeyPEM(pubPEM)
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

	// Start background listener for acceptance response.
	a.startEnrollmentListener(ctx)

	a.state = job.AgentStatePending
	a.logger.Info(
		"agent not enrolled, entering pending state",
		slog.String("fingerprint", m.Fingerprint()),
	)

	return nil
}

// publishEnrollmentRequest sends the agent's enrollment request to the
// controller via core NATS (not JetStream).
func (a *Agent) publishEnrollmentRequest() error {
	if a.natsClient == nil {
		return fmt.Errorf("NATS client not available")
	}

	req := agentpki.EnrollmentRequest{
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

	if err := a.natsClient.PublishCore(subject, data); err != nil {
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

// startEnrollmentListener subscribes to the enrollment response subject
// for this agent's machine ID. When the controller accepts the agent,
// the listener saves the controller's public key, transitions to Ready,
// and starts job consumers.
func (a *Agent) startEnrollmentListener(
	ctx context.Context,
) {
	if a.natsClient == nil {
		return
	}

	namespace := a.appConfig.Agent.NATS.Namespace
	subject := "enroll.response." + a.machineID
	if namespace != "" {
		subject = namespace + "." + subject
	}

	sub, err := a.natsClient.Subscribe(subject, func(msg *nats.Msg) {
		a.handleEnrollmentResponse(msg)
	})
	if err != nil {
		a.logger.Warn(
			"failed to subscribe to enrollment response",
			slog.String("subject", subject),
			slog.String("error", err.Error()),
		)

		return
	}

	a.logger.Info(
		"listening for enrollment acceptance",
		slog.String("subject", subject),
	)

	// Unsubscribe when context is cancelled.
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		<-ctx.Done()
		_ = sub.Unsubscribe()
	}()
}

// handleEnrollmentResponse processes the controller's enrollment response.
// On acceptance, it saves the controller public key and transitions to Ready.
func (a *Agent) handleEnrollmentResponse(
	msg *nats.Msg,
) {
	var resp agentpki.EnrollmentResponse
	if err := json.Unmarshal(msg.Data, &resp); err != nil {
		a.logger.Warn(
			"failed to parse enrollment response",
			slog.String("error", err.Error()),
		)

		return
	}

	if !resp.Accepted {
		a.logger.Warn(
			"enrollment rejected",
			slog.String("reason", resp.Reason),
		)

		return
	}

	// Save controller public key to disk.
	controllerPubKey := ed25519.PublicKey(resp.ControllerPublicKey)
	a.pkiManager.SetControllerPublicKey(controllerPubKey)

	keyDir := a.appConfig.Agent.PKI.KeyDir
	controllerPubPath := filepath.Join(keyDir, "controller.pub")

	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "ED25519 PUBLIC KEY",
		Bytes: controllerPubKey,
	})

	if err := avfs.WriteFile(a.appFs, controllerPubPath, pubPEM, 0o644); err != nil {
		a.logger.Error(
			"failed to save controller public key",
			slog.String("path", controllerPubPath),
			slog.String("error", err.Error()),
		)

		return
	}

	// Transition to Ready and start consumers.
	a.state = job.AgentStateReady
	a.startConsumers()

	a.logger.Info(
		"enrollment accepted, agent is now ready",
		slog.String("controller_fingerprint", a.pkiManager.Fingerprint()),
	)
}
