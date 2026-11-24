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

package admission

import (
	"context"
	"errors"
	"time"

	"github.com/osapi-io/osapi/internal/controller/admission"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
)

// List returns the admission records, in the state that in names when it
// names one, each with whether its node is online and when it last
// reported.
func List(
	ctx context.Context,
	d Deps,
	in ListInput,
) (ListAdmissionResponse, error) {
	if d.Store == nil {
		return ListAdmissionResponse{}, notConfigured()
	}
	records, err := d.Store.List(ctx)
	if err != nil {
		return ListAdmissionResponse{}, endpoint.Internal("failed to list admission records", err)
	}
	seen := lastSeen(ctx, d)

	items := make([]Admission, 0, len(records))
	for _, r := range records {
		if in.State != "" && r.State != in.State {
			continue
		}
		items = append(items, admissionOf(r, seen))
	}

	return ListAdmissionResponse{TotalItems: len(items), Items: items}, nil
}

// Get returns one machine's admission record.
func Get(
	ctx context.Context,
	d Deps,
	in MachineInput,
) (Admission, error) {
	if d.Store == nil {
		return Admission{}, notConfigured()
	}
	r, err := d.Store.Get(ctx, in.MachineID)
	if err != nil {
		if errors.Is(err, admission.ErrNotFound) {
			return Admission{}, endpoint.Errorf(
				ErrNotFound,
				"no admission record for machine %s",
				in.MachineID,
			)
		}

		return Admission{}, endpoint.Internal("failed to get the admission record", err)
	}

	return admissionOf(r, lastSeen(ctx, d)), nil
}

// Admit admits a pending machine when the fingerprint is the one of its
// request.
func Admit(
	ctx context.Context,
	d Deps,
	in AdmitInput,
) (struct{}, error) {
	if d.Store == nil {
		return struct{}{}, notConfigured()
	}
	err := d.Store.Admit(ctx, in.MachineID, in.Body.Fingerprint)
	switch {
	case err == nil:
		return struct{}{}, nil
	case errors.Is(err, admission.ErrNoPendingAgent):
		return struct{}{}, endpoint.Errorf(
			ErrNotFound,
			"no pending request from machine %s",
			in.MachineID,
		)
	case errors.Is(err, admission.ErrFingerprintMismatch):
		return struct{}{}, endpoint.Errorf(
			ErrMismatch,
			"fingerprint %s is not the one of machine %s's request; run osapi agent key fingerprint on the machine",
			in.Body.Fingerprint,
			in.MachineID,
		)
	default:
		return struct{}{}, endpoint.Internal("failed to admit the machine", err)
	}
}

// Reject rejects a pending machine.
func Reject(
	ctx context.Context,
	d Deps,
	in RejectInput,
) (struct{}, error) {
	if d.Store == nil {
		return struct{}{}, notConfigured()
	}
	err := d.Store.Reject(ctx, in.MachineID, in.Body.Reason)
	switch {
	case err == nil:
		return struct{}{}, nil
	case errors.Is(err, admission.ErrNoPendingAgent):
		return struct{}{}, endpoint.Errorf(
			ErrNotFound,
			"no pending request from machine %s",
			in.MachineID,
		)
	default:
		return struct{}{}, endpoint.Internal("failed to reject the machine", err)
	}
}

// Revoke removes an admitted machine's key.
func Revoke(
	ctx context.Context,
	d Deps,
	in MachineInput,
) (struct{}, error) {
	if d.Store == nil {
		return struct{}{}, notConfigured()
	}
	err := d.Store.Revoke(ctx, in.MachineID)
	switch {
	case err == nil:
		return struct{}{}, nil
	case errors.Is(err, admission.ErrNoAcceptedAgent):
		return struct{}{}, endpoint.Errorf(ErrNotFound, "no admitted machine %s", in.MachineID)
	default:
		return struct{}{}, endpoint.Internal("failed to revoke the machine", err)
	}
}

// seenNow is when each node last reported, and whether it is online now.
type seenNow struct {
	online map[string]time.Time
	stored map[string]time.Time
}

// lastSeen reads which nodes are online, from the registry, and when each
// last reported. A registry or a store that cannot be read leaves the
// records without it: the record itself is what the operator decides on.
func lastSeen(
	ctx context.Context,
	d Deps,
) seenNow {
	s := seenNow{online: map[string]time.Time{}, stored: map[string]time.Time{}}
	if d.Jobs == nil {
		return s
	}
	if agents, err := d.Jobs.ListAgents(ctx); err == nil {
		for _, a := range agents {
			s.online[a.MachineID] = a.RegisteredAt
		}
	}
	if stored, err := d.Jobs.AgentsLastSeen(ctx); err == nil {
		s.stored = stored
	}

	return s
}

// admissionOf is a record as the API answers it.
func admissionOf(
	r admission.Record,
	seen seenNow,
) Admission {
	a := Admission{
		MachineID:           r.MachineID,
		Hostname:            r.Hostname,
		Fingerprint:         r.Fingerprint,
		EncryptionKey:       r.EncryptionKey,
		State:               r.State,
		RequestedAt:         timeOf(r.RequestedAt),
		AdmittedAt:          timeOf(r.AdmittedAt),
		DecidedAt:           timeOf(r.DecidedAt),
		Reason:              r.Reason,
		Previous:            r.Previous,
		PreviousFingerprint: r.PreviousFingerprint,
	}
	if t, ok := seen.online[r.MachineID]; ok {
		a.Online = true
		a.LastSeen = timeOf(t)
	} else {
		a.LastSeen = timeOf(seen.stored[r.MachineID])
	}

	return a
}

// timeOf is t, or nil for the zero time, so an unset time is absent from
// the answer.
func timeOf(
	t time.Time,
) *time.Time {
	if t.IsZero() {
		return nil
	}

	return &t
}

// notConfigured is the 503 of each method when PKI is off.
func notConfigured() error {
	return endpoint.Errorf(ErrNotConfigured,
		"admission is not configured: set controller.pki.enabled on the controller")
}

// Delete removes a machine's key and each record of it.
func Delete(
	ctx context.Context,
	d Deps,
	in MachineInput,
) (struct{}, error) {
	if d.Store == nil {
		return struct{}{}, notConfigured()
	}
	err := d.Store.Delete(ctx, in.MachineID)
	switch {
	case err == nil:
		return struct{}{}, nil
	case errors.Is(err, admission.ErrNotFound):
		return struct{}{}, endpoint.Errorf(ErrNotFound, "no record of machine %s", in.MachineID)
	default:
		return struct{}{}, endpoint.Internal("failed to delete the machine", err)
	}
}
