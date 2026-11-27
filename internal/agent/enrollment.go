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
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/avfs/avfs"
	"github.com/nats-io/nats.go"

	"github.com/osapi-io/osapi/internal/pki"
)

// EnrolledControllerKey is the file in agent.pki.key_dir of the controller's
// public key, which the agent saves when the controller admits it. It is not
// controller.pub: that is the name of the controller's own key pair, and the
// two share a directory when both run on one host with the default key_dir.
const EnrolledControllerKey = "enrolled-controller.pub"

// legacyControllerKey is the file an agent saved before EnrolledControllerKey.
// It is read only when no controller.key is beside it, since with one it is
// the controller's own key, not proof that the agent was admitted.
const legacyControllerKey = "controller.pub"

// enrolledKey is the controller's public key that the agent saved when it
// was admitted, from keyDir, or an error when it saved none.
func (a *Agent) enrolledKey(
	keyDir string,
) ([]byte, error) {
	pubPEM, err := avfs.ReadFile(a.appFs, filepath.Join(keyDir, EnrolledControllerKey))
	if err == nil {
		return pubPEM, nil
	}
	if _, keyErr := a.appFs.Stat(filepath.Join(keyDir, "controller.key")); keyErr == nil {
		return nil, err
	}

	return avfs.ReadFile(a.appFs, filepath.Join(keyDir, legacyControllerKey))
}

// marshalJSONEnrollment is a package-level variable for testing the marshal error path.
var marshalJSONEnrollment = json.Marshal

// handlePKIEnrollment manages the PKI enrollment lifecycle.
// It loads or generates the signing and encryption keys, checks for existing controller key,
// and if not enrolled, publishes an enrollment request, starts a
// background listener for acceptance, and sets state to Pending.
func (a *Agent) handlePKIEnrollment(
	ctx context.Context,
) error {
	keyDir := a.appConfig.Agent.PKI.KeyDir

	m := pki.New(a.appFs, keyDir, "agent", pki.WithEncryptionKey())

	if err := m.LoadOrGenerate(); err != nil {
		return err
	}

	a.pkiManager = m

	// Wire the signer into the job client now that the agent keypair is
	// loaded: job responses are signed going forward. jobClient is nil
	// only in tests that exercise enrollment in isolation.
	if a.jobClient != nil {
		a.jobClient.SetPKISigner(m)
		// Stamp this agent's identity on what it signs, so the controller
		// knows which stored key to verify its responses against. Start
		// resolves the machine ID before this runs.
		a.jobClient.SetMachineID(a.machineID)
	}

	a.pkiLogger.Info(
		"PKI keypair loaded",
		slog.String("fingerprint", m.Fingerprint()),
		slog.String("key_dir", keyDir),
	)

	// Check if the agent saved the controller's key when it was admitted.
	pubPEM, err := a.enrolledKey(keyDir)
	if err == nil {
		pubKey, parseErr := pki.ParsePublicKeyPEM(pubPEM)
		if parseErr != nil {
			return fmt.Errorf("failed to parse controller public key: %w", parseErr)
		}

		m.SetControllerPublicKey(pubKey)
		a.pkiLogger.Info("controller public key loaded, agent enrolled")

		// A rotation of the controller's key that started before this run
		// left its next key.
		return a.loadNextControllerKey(keyDir)
	}

	// Not enrolled. With a join string, join over HTTPS, which admits at
	// once; without one, or when the join fails, ask over NATS.
	if a.appConfig.Agent.Join != "" {
		err := a.join(ctx)
		if err == nil {
			a.pkiLogger.Info("joined the controller with the join token, agent enrolled")

			return nil
		}
		a.pkiLogger.Warn(
			"failed to join with the join string; asking to be admitted over NATS",
			slog.String("error", err.Error()),
		)
	}

	// Not enrolled — publish enrollment request and enter pending state.
	if err := a.publishEnrollmentRequest(); err != nil {
		a.pkiLogger.Warn(
			"failed to publish enrollment request",
			slog.String("error", err.Error()),
		)
	}

	// Start background listener for acceptance response.
	a.startEnrollmentListener(ctx)

	a.setState(types.AgentPending)
	a.pkiLogger.Info(
		"agent not enrolled, entering pending state",
		slog.String("fingerprint", m.Fingerprint()),
	)

	return nil
}

// retryEnrollmentIfPending asks again for an agent still waiting to be accepted.
//
// The first request is published once, over core NATS, with no delivery guarantee
// and nothing to retry it: a controller that was down, a broker that dropped it,
// or a publish that simply failed left the agent Pending with the controller
// having never heard of it. Since Start only consumes jobs when the agent is not
// Pending, that state was permanent until somebody restarted the process.
//
// Republishing is safe because acceptance is keyed by machine ID: the controller
// overwrites the pending entry rather than accumulating them, and an operator who
// has already accepted this agent sees nothing new. It also recovers the case
// where the controller lost its pending entries and the agent is the only thing
// that still knows it wants in.
//
// Called from the heartbeat, which runs while Pending precisely so the agent stays
// visible.
func (a *Agent) retryEnrollmentIfPending() {
	if a.State() != types.AgentPending {
		return
	}

	if err := a.publishEnrollmentRequest(); err != nil {
		a.pkiLogger.Warn(
			"failed to republish enrollment request",
			slog.String("error", err.Error()),
		)

		return
	}

	a.pkiLogger.Info(
		"enrollment request republished, still waiting for admission",
		slog.String("machine_id", a.machineID),
	)
}

// publishEnrollmentRequest sends the agent's enrollment request to the
// controller via core NATS (not JetStream).
func (a *Agent) publishEnrollmentRequest() error {
	if a.natsClient == nil {
		return fmt.Errorf("NATS client not available")
	}

	req := pki.EnrollmentRequest{
		MachineID:     a.machineID,
		Hostname:      a.currentHostname(),
		PublicKey:     a.pkiManager.PublicKey(),
		EncryptionKey: a.pkiManager.EncryptionPublicKey(),
		Fingerprint:   a.pkiManager.Fingerprint(),
	}

	data, err := marshalJSONEnrollment(req)
	if err != nil {
		return fmt.Errorf("marshal enrollment request: %w", err)
	}

	namespace := a.appConfig.Agent.NATS.Namespace
	subject := pki.EnrollRequestSuffix
	if namespace != "" {
		subject = namespace + "." + subject
	}

	if err := a.natsClient.PublishCore(context.Background(), subject, data); err != nil {
		return fmt.Errorf("publish enrollment request: %w", err)
	}

	a.pkiLogger.Info(
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
	subject := pki.EnrollResponsePrefix + "." + a.machineID
	if namespace != "" {
		subject = namespace + "." + subject
	}

	sub, err := a.natsClient.Subscribe(ctx, subject, func(_ context.Context, msg *nats.Msg) {
		a.handleEnrollmentResponse(msg)
	})
	if err != nil {
		a.pkiLogger.Warn(
			"failed to subscribe to enrollment response",
			slog.String("subject", subject),
			slog.String("error", err.Error()),
		)

		return
	}

	a.pkiLogger.Info(
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
	var resp pki.EnrollmentResponse
	if err := json.Unmarshal(msg.Data, &resp); err != nil {
		a.pkiLogger.Warn(
			"failed to parse enrollment response",
			slog.String("error", err.Error()),
		)

		return
	}

	if !resp.Accepted {
		a.pkiLogger.Warn(
			"enrollment rejected",
			slog.String("reason", resp.Reason),
		)

		return
	}

	if err := a.keepControllerKey(ed25519.PublicKey(resp.ControllerPublicKey)); err != nil {
		a.pkiLogger.Error("failed to save controller public key", slog.String("error", err.Error()))

		return
	}

	// Transition to Ready and start consumers, as one step for the reason given
	// on lifecycleMu.
	a.lifecycleMu.Lock()
	a.state = types.AgentReady
	a.startConsumersLocked()
	a.lifecycleMu.Unlock()

	a.pkiLogger.Info(
		"admitted, agent is now ready",
		slog.String("controller_fingerprint", a.pkiManager.Fingerprint()),
	)
}

// keepControllerKey uses key, the controller's public key that the agent got
// when it was admitted, and saves it as EnrolledControllerKey, so that the
// agent is admitted at its next start too.
func (a *Agent) keepControllerKey(
	key ed25519.PublicKey,
) error {
	a.pkiManager.SetControllerPublicKey(key)
	path := filepath.Join(a.appConfig.Agent.PKI.KeyDir, EnrolledControllerKey)

	return writeControllerKey(context.Background(), a.appFs, path, key)
}
