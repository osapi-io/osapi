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

package client

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"fmt"

	"github.com/osapi-io/osapi/internal/job"
)

// signingMarshalFn is the JSON marshal function used by signing helpers.
// Overridable in tests via export_test.go.
var signingMarshalFn = json.Marshal

// wrapInSignedEnvelope signs the payload and wraps it in a SignedEnvelope.
// machineID identifies the signer so the verifier can find the key to check
// the signature against; it is empty for controller-signed payloads.
// Returns the JSON-encoded envelope.
func wrapInSignedEnvelope(
	signer PKISigner,
	machineID string,
	payload []byte,
) ([]byte, error) {
	signature := signer.Sign(payload)
	envelope := job.SignedEnvelope{
		Payload:     payload,
		Signature:   signature,
		Fingerprint: signer.Fingerprint(),
		MachineID:   machineID,
	}

	envelopeJSON, err := signingMarshalFn(envelope)
	if err != nil {
		return nil, fmt.Errorf("marshal signed envelope: %w", err)
	}

	return envelopeJSON, nil
}

// unwrapSignedEnvelope attempts to unwrap a SignedEnvelope. If the data
// is not a signed envelope (missing required fields), it returns the
// original data unchanged with isEnvelope=false. If it is a signed
// envelope, it verifies the signature against the provided public key
// and returns the inner payload.
func unwrapSignedEnvelope(
	data []byte,
	pubKey ed25519.PublicKey,
) (payload []byte, isEnvelope bool, err error) {
	var envelope job.SignedEnvelope
	if jsonErr := json.Unmarshal(data, &envelope); jsonErr != nil {
		return data, false, nil
	}

	// Check if this looks like a signed envelope by verifying required
	// fields are present.
	if len(envelope.Payload) == 0 || len(envelope.Signature) == 0 || envelope.Fingerprint == "" {
		return data, false, nil
	}

	// If no public key is available, return the payload without verification.
	if len(pubKey) == 0 {
		return envelope.Payload, true, nil
	}

	// Verify the signature.
	if !ed25519.Verify(pubKey, envelope.Payload, envelope.Signature) {
		return nil, true, fmt.Errorf("invalid signature on signed envelope")
	}

	return envelope.Payload, true, nil
}

// verifyAgentResponse checks a response against the key stored for the agent
// that claims to have sent it, and returns the inner payload.
//
// The envelope's machine ID is self-reported, so it is used only to choose
// which stored record to verify against: a response naming another agent is
// checked against that agent's key and fails unless it was actually signed by
// it. The hostname in the payload must also match the one recorded at
// acceptance, so an accepted agent cannot answer for a host it did not enrol
// as.
//
// Every rejection carries a distinct cause. There is no path that returns the
// payload unverified.
func verifyAgentResponse(
	ctx context.Context,
	store AgentKeyStore,
	data []byte,
) ([]byte, error) {
	var envelope job.SignedEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, ErrResponseNotSigned
	}

	if len(envelope.Payload) == 0 || len(envelope.Signature) == 0 ||
		envelope.MachineID == "" {
		return nil, ErrResponseNotSigned
	}

	record, err := store.LookupAgentKey(ctx, envelope.MachineID)
	if err != nil {
		return nil, err
	}

	if !ed25519.Verify(record.PublicKey, envelope.Payload, envelope.Signature) {
		return nil, fmt.Errorf(
			"%w: machine %s",
			ErrResponseSignatureInvalid,
			envelope.MachineID,
		)
	}

	var response job.Response
	if err := json.Unmarshal(envelope.Payload, &response); err != nil {
		return nil, fmt.Errorf(
			"%w: machine %s payload is not a response",
			ErrResponseNotSigned,
			envelope.MachineID,
		)
	}

	if response.Hostname != record.Hostname {
		return nil, fmt.Errorf(
			"%w: machine %s enrolled as %q but answered as %q",
			ErrResponseHostnameMismatch,
			envelope.MachineID,
			record.Hostname,
			response.Hostname,
		)
	}

	return envelope.Payload, nil
}
