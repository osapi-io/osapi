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
	"log/slog"
	"time"

	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// signingMarshalFn is the JSON marshal function used by signing helpers.
// Overridable in tests via export_test.go.
var signingMarshalFn = json.Marshal

// nowFn returns the current time, so a rotation grace window can be tested
// without waiting for one. Overridable in tests via export_test.go.
var nowFn = time.Now

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

// unwrapSignedEnvelope is the payload of data, a SignedEnvelope, after its
// signature verifies with pubKey. Data that is not a signed envelope, and
// a missing key, are errors: no path returns a payload unverified.
func unwrapSignedEnvelope(
	data []byte,
	pubKey ed25519.PublicKey,
) ([]byte, error) {
	var envelope job.SignedEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil ||
		len(envelope.Payload) == 0 || len(envelope.Signature) == 0 || envelope.Fingerprint == "" {
		return nil, ErrNotSigned
	}
	if len(pubKey) == 0 {
		return nil, ErrNoSigner
	}
	if !ed25519.Verify(pubKey, envelope.Payload, envelope.Signature) {
		return nil, fmt.Errorf("invalid signature on signed envelope")
	}

	return envelope.Payload, nil
}

// verifyAgentResponse checks a response against the key stored for the agent
// that claims to have sent it, and returns the response in its payload.
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
) (*job.Response, error) {
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

	verified := false

	for _, key := range record.AcceptableKeys(nowFn()) {
		if ed25519.Verify(key, envelope.Payload, envelope.Signature) {
			verified = true

			break
		}
	}

	if !verified {
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

	response.MachineID = envelope.MachineID

	return &response, nil
}

// hostnameUpdate is the operation of the hostname behavior's update.
const hostnameUpdate = job.OperationType(job.CategoryBehavior + ".hostname.update")

// followRename records the new name of a host that an operator renamed
// through osapi, so the agent's later answers and registrations, which carry
// the new name, verify. The answer of the update itself carries the old
// name, which the stored record still has.
func (c *Client) followRename(
	ctx context.Context,
	req *job.Request,
	r *job.Response,
) {
	if req.Operation != hostnameUpdate || r.Status != types.AgentJobStatusCompleted {
		return
	}
	var out struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(r.Data, &out); err != nil || out.Name == "" ||
		out.Name == r.Hostname {
		return
	}
	if err := c.agentKeyStore.RenameAgentKey(ctx, r.MachineID, out.Name); err != nil {
		c.logger.WarnContext(ctx, "record the new hostname of a renamed host",
			slog.String("machine_id", r.MachineID),
			slog.String("hostname", out.Name),
			slog.String("error", err.Error()),
		)
	}
}
