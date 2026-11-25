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
	"time"

	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// ListInput filters the records by state.
type ListInput struct {
	State types.AdmissionState `json:"state" in:"query" validate:"omitempty,enum" doc:"Only the records in this state: pending, admitted, rejected or revoked."`
}

// MachineInput names a machine.
type MachineInput struct {
	MachineID string `json:"machine_id" in:"path" validate:"required,min=1,max=255" doc:"The machine ID of the node."`
}

// AdmitInput names a pending machine, and the fingerprint the machine
// printed.
type AdmitInput struct {
	MachineID string       `json:"machine_id" in:"path" validate:"required,min=1,max=255" doc:"The machine ID of the node."`
	Body      AdmitRequest `                  in:"body" validate:"required"`
}

// AdmitRequest is the fingerprint that proves the machine.
type AdmitRequest struct {
	Fingerprint string `json:"fingerprint" validate:"required,min=1,max=255" doc:"The fingerprint that osapi agent key fingerprint prints on the machine." example:"SHA256:1f53dea6dba17ae38388b931df5c6370076c8236018d6e8a2adecc8474ae93ef" flag:"fingerprint"`
}

// RejectInput names a pending machine, and why it is rejected.
type RejectInput struct {
	MachineID string        `json:"machine_id" in:"path" validate:"required,min=1,max=255" doc:"The machine ID of the node."`
	Body      RejectRequest `                  in:"body" validate:"required"`
}

// RejectRequest is why the machine is rejected.
type RejectRequest struct {
	Reason string `json:"reason,omitempty" validate:"omitempty,max=1024" doc:"Why the machine is rejected; the machine and the record keep it." example:"unknown host" flag:"reason"`
}

// ListAdmissionResponse is the admission records.
type ListAdmissionResponse struct {
	TotalItems int         `json:"total_items" doc:"The number of records."                    example:"3"`
	Items      []Admission `json:"items"       doc:"The admission records, in hostname order."`
}

// Admission is one machine's admission record.
type Admission struct {
	MachineID           string               `json:"machine_id"                     doc:"The machine ID of the node."                                                               example:"d530664edc1c431eb5ef1e9dfd6617a5"`
	Hostname            string               `json:"hostname"                       doc:"The hostname the node gave."                                                               example:"web-01"`
	State               types.AdmissionState `json:"state"                          doc:"pending, admitted, rejected or revoked."                                                   example:"admitted"`
	Online              bool                 `json:"online"                         doc:"Whether the node is running now: it is in the registry."`
	LastSeen            *time.Time           `json:"last_seen,omitempty"            doc:"When the node last reported."                                                                                                                                                render:"age"      none:"never"`
	Fingerprint         string               `json:"fingerprint"                    doc:"The fingerprint of the node's two public keys: signing and encryption."                    example:"SHA256:1f53dea6dba17ae38388b931df5c6370076c8236018d6e8a2adecc8474ae93ef" render:"short=20"`
	EncryptionKey       string               `json:"encryption_key,omitempty"       doc:"The node's public X25519 encryption key, which the fingerprint covers."                    example:"XA7NTLWJ2ECILFDLHORNVYROXOK65ONQ2SJYHRNAWRJZ3UXMVXIGUVJM"                                               table:"-"`
	RequestedAt         *time.Time           `json:"requested_at,omitempty"         doc:"When the node asked to join."                                                                                                                                                render:"age"                   table:"-"`
	AdmittedAt          *time.Time           `json:"admitted_at,omitempty"          doc:"When the node was admitted."                                                                                                                                                 render:"age"                   table:"-"`
	DecidedAt           *time.Time           `json:"decided_at,omitempty"           doc:"When the node was rejected or revoked."                                                                                                                                      render:"age"                   table:"-"`
	Reason              string               `json:"reason,omitempty"               doc:"Why the node was rejected."                                                                                                                                                                                 table:"-"`
	Previous            types.AdmissionState `json:"previous,omitempty"             doc:"On a pending record, the state of the node's earlier record: a returning node."            example:"revoked"                                                                                                table:"-"`
	PreviousFingerprint string               `json:"previous_fingerprint,omitempty" doc:"On a pending record, the fingerprint of the earlier record; a different one is a new key."                                                                                                                  table:"-"`
	JoinToken           string               `json:"join_token,omitempty"           doc:"On an admitted record that a join made, the ID of the join token that admitted the node."                                                                                                                   table:"-"`
}
