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
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/osapi-io/osapi/internal/audit"
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

// unwrapStoredEnvelope is unwrapSignedEnvelope for a record that the
// controller stored: any of keys, the controller's current key and each one
// that a rotation replaced, may have signed it.
func unwrapStoredEnvelope(
	data []byte,
	keys []ed25519.PublicKey,
) ([]byte, error) {
	if len(keys) == 0 {
		return unwrapSignedEnvelope(data, nil)
	}
	var err error
	for _, key := range keys {
		var payload []byte
		if payload, err = unwrapSignedEnvelope(data, key); err == nil {
			return payload, nil
		}
		if errors.Is(err, ErrNotSigned) {
			return nil, err
		}
	}

	return nil, err
}

// RefusedError is a response whose signature is not the key that the
// controller admitted for its machine. Hostname is the name recorded at the
// admission, not the one the response claims.
type RefusedError struct {
	MachineID string
	Hostname  string
}

func (e *RefusedError) Error() string {
	return fmt.Sprintf("%s: machine %s", ErrResponseSignatureInvalid, e.MachineID)
}

func (e *RefusedError) Unwrap() error {
	return ErrResponseSignatureInvalid
}

// refused is the row of the host whose response verify refused, and true, or
// false for a response that is not a refusal, such as one that is not signed.
// The controller records the refusal as the host's state of the job, which
// outranks what the agent reported (osapi-io/osapi#1484).
func (c *Client) refused(
	ctx context.Context,
	jobID string,
	verifyErr error,
) (*job.Response, bool) {
	var r *RefusedError
	if !errors.As(verifyErr, &r) || r.Hostname == "" {
		return nil, false
	}
	msg := "the controller refused the result: its signature is not the key that the controller admitted for " + r.Hostname
	if err := c.WriteStatusEvent(ctx, jobID, types.AgentJobStatusRefused, r.Hostname,
		map[string]interface{}{"error": msg}); err != nil {
		c.logger.WarnContext(ctx, "failed to record a refused result",
			slog.String("job_id", jobID), slog.String("error", err.Error()))
	}
	if _, seen := c.refusedSeen.LoadOrStore(r.MachineID, struct{}{}); !seen {
		c.firstRefusal(ctx, r, msg)
	}

	return &job.Response{
		JobID: jobID, Status: types.AgentJobStatusRefused, Hostname: r.Hostname, Error: msg,
	}, true
}

// firstRefusal records the first result of a machine that this client
// refused: an event on the agent's timeline, and an entry of the controller
// in the audit log. No request made it, and the job shows it only as refused.
func (c *Client) firstRefusal(
	ctx context.Context,
	r *RefusedError,
	msg string,
) {
	if err := c.WriteAgentTimelineEvent(ctx, r.MachineID, r.Hostname, "refused", msg); err != nil {
		c.logger.WarnContext(ctx, "failed to write a timeline event",
			slog.String("event", "refused"), slog.String("error", err.Error()))
	}
	if c.auditLog == nil {
		return
	}
	if err := c.auditLog.Write(context.WithoutCancel(ctx), audit.Change(
		ctx, audit.ActorSystem, "controller", "agent.refuse", r.MachineID,
		http.StatusForbidden, msg,
	)); err != nil {
		c.logger.WarnContext(ctx, "failed to write an audit entry",
			slog.String("action", "agent.refuse"), slog.String("error", err.Error()))
	}
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
		return nil, &RefusedError{MachineID: envelope.MachineID, Hostname: record.Hostname}
	}

	var response job.Response
	if err := json.Unmarshal(envelope.Payload, &response); err != nil {
		return nil, fmt.Errorf(
			"%w: machine %s payload is not a response",
			ErrResponseNotSigned,
			envelope.MachineID,
		)
	}

	// The machine ID and its key are the agent's identity: the job went to
	// that machine, and its key signed the answer. The hostname is a name
	// for a person, which a rename changes before the record follows it.
	// Thus the answer shows the name of the record, the controller's own,
	// and an agent cannot put another host's name on its answer.
	response.Hostname = record.Hostname

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

		return
	}
	_ = c.WriteAgentTimelineEvent(ctx, r.MachineID, out.Name, "hostname_changed",
		fmt.Sprintf("hostname changed from %s to %s", r.Hostname, out.Name))
}
