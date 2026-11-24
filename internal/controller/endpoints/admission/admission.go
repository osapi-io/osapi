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

// Package admission is the endpoint group for admission: the nodes that ask
// to join, the nodes osapi has admitted, and the decisions against nodes.
// It is separate from the agent group, which is the running fleet: an
// admitted node keeps its record while it is offline, and a node waits for
// admission before it is part of the fleet.
package admission

import (
	"errors"
	"net/http"
	"strings"

	"github.com/osapi-io/osapi/internal/area"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
)

var (
	// ErrNotFound is a machine that osapi has no record of, or that has no
	// record in the state the method needs.
	ErrNotFound = errors.New("admission record not found")
	// ErrMismatch is an admit whose fingerprint is not the one of the key
	// the machine requested with.
	ErrMismatch = errors.New("fingerprint does not match the request")
	// ErrNotConfigured is a controller with PKI off, which admits no one.
	ErrNotConfigured = errors.New("admission is not configured")
)

// notConfigured503 is the 503 of each method when PKI is off.
var notConfigured503 = endpoint.Unavailable(ErrNotConfigured,
	"Admission is not configured: set controller.pki.enabled.")

// Group is the admission group's definition.
var Group = endpoint.Register(endpoint.Define(
	"admission", "The admission of nodes",
	endpoint.Area(area.Fleet),
	endpoint.Title("Admission API"),
	endpoint.Tag("admission", "Admission",
		"Admit, reject and revoke the nodes that ask to join."),
	endpoint.Uses[Deps](),
	endpoint.Get[ListInput, ListAdmissionResponse]("list").
		Fails("a state that is not one of the four", ListInput{State: "accepted"}).
		Handle(List).
		ID("GetAdmissions").
		Describe("List admission records").
		Explain("Returns each machine's admission record: pending, admitted, rejected or "+
			"revoked, with whether it is online and when it was last seen.").
		Returns("The admission records.").
		Command("List admission records",
			"Each node that asked to join, each node osapi admitted, and each decision\n"+
				"against a node. An admitted node that is not running is Offline, with when\n"+
				"it was last seen.").
		Example("").
		Prints(sampleList).
		Example("--state pending").
		Errors(notConfigured503),
	endpoint.Get[MachineInput, Admission]("get").
		Fails("a machine ID over 255 characters", MachineInput{MachineID: long}).
		Handle(Get).
		ID("GetAdmission").
		Describe("Get a machine's admission record").
		Explain("Returns one machine's admission record by machine ID.").
		Returns("The admission record.").
		Command("Get a machine's admission record", "").
		Example("--machine-id d530664edc1c431eb5ef1e9dfd6617a5").
		Prints(samplePending).
		Errors(endpoint.NotFound(ErrNotFound, "No admission record for the machine."), notConfigured503),
	endpoint.Post[AdmitInput, struct{}]("admit").
		AdminOnly().
		Fails("a machine ID over 255 characters", AdmitInput{MachineID: long, Body: AdmitRequest{Fingerprint: "SHA256:x"}}).
		Fails("no fingerprint", AdmitInput{MachineID: "m1", Body: AdmitRequest{}}).
		Fails("a fingerprint over 255 characters", AdmitInput{MachineID: "m1", Body: AdmitRequest{Fingerprint: long}}).
		Status(http.StatusNoContent).
		Handle(Admit).
		Action("admit").
		ID("AdmitMachine").
		Describe("Admit a pending machine").
		Explain("Admits a pending machine. The fingerprint must be the one that "+
			"`osapi agent key fingerprint` prints on the machine: it proves that the request "+
			"came from the machine the operator is looking at.").
		Returns("The machine is admitted.").
		Command("Admit a pending machine",
			"Run `osapi agent key fingerprint` on the machine, and give the fingerprint that it\n"+
				"prints. The controller admits the machine only when the fingerprint is the\n"+
				"one of the key in its request.").
		Example("--machine-id d530664edc1c431eb5ef1e9dfd6617a5 --fingerprint SHA256:1f53dea6...").
		Errors(
			endpoint.NotFound(ErrNotFound, "No pending request from the machine."),
			endpoint.Conflict(ErrMismatch, "The fingerprint does not match the request."),
			notConfigured503,
		),
	endpoint.Post[RejectInput, struct{}]("reject").
		AdminOnly().
		Fails("a machine ID over 255 characters", RejectInput{MachineID: long}).
		Fails("a reason over 1024 characters", RejectInput{MachineID: "m1", Body: RejectRequest{Reason: strings.Repeat("a", 1025)}}).
		Status(http.StatusNoContent).
		Handle(Reject).
		Action("reject").
		ID("RejectMachine").
		Describe("Reject a pending machine").
		Explain("Rejects a pending machine, and keeps a record of the decision, so a "+
			"machine that asks again is flagged.").
		Returns("The machine is rejected.").
		Command("Reject a pending machine", "").
		Example("--machine-id d530664edc1c431eb5ef1e9dfd6617a5 --reason \"unknown host\"").
		Errors(endpoint.NotFound(ErrNotFound, "No pending request from the machine."), notConfigured503),
	endpoint.Post[MachineInput, struct{}]("revoke").
		AdminOnly().
		Fails("a machine ID over 255 characters", MachineInput{MachineID: long}).
		Status(http.StatusNoContent).
		Handle(Revoke).
		Action("revoke").
		ID("RevokeMachine").
		Describe("Revoke an admitted machine").
		Explain("Removes an admitted machine's key, and keeps a record of the decision. "+
			"Nothing the machine signs verifies afterwards; to return, it asks again.").
		Returns("The machine is revoked.").
		Command("Revoke an admitted machine",
			"The controller removes the key of the machine. Its responses and its\n"+
				"registration stop verifying at once, so a target does not find it.").
		Example("--machine-id d530664edc1c431eb5ef1e9dfd6617a5").
		Errors(endpoint.NotFound(ErrNotFound, "No admitted machine with the ID."), notConfigured503),
	endpoint.Delete[MachineInput, struct{}]("delete").
		AdminOnly().
		Fails("a machine ID over 255 characters", MachineInput{MachineID: long}).
		Status(http.StatusNoContent).
		Handle(Delete).
		ID("DeleteMachine").
		Describe("Delete a machine").
		Explain("Removes a machine's key and each record of it, and keeps none. A machine "+
			"that asks again is new. Use it for a machine that goes away on purpose, such "+
			"as a VM that is rebuilt; revoke a machine that is not trusted.").
		Returns("The machine is deleted.").
		Command("Delete a machine",
			"The controller removes the key of the machine and each record of it. Unlike\n"+
				"revoke, it keeps no record, so a machine that asks again is new. Use it\n"+
				"before a machine is rebuilt or removed.").
		Example("--machine-id d530664edc1c431eb5ef1e9dfd6617a5").
		Errors(endpoint.NotFound(ErrNotFound, "No record of the machine."), notConfigured503),
))

// long is a name over the 255 characters a name may have, for the inputs
// that Fails declares.
var long = strings.Repeat("a", 256)
