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
	"time"

	"github.com/osapi-io/osapi/pkg/sdk/client/gen"
)

// AdmissionList is the admission records.
type AdmissionList struct {
	Items      []Admission `json:"items"`
	TotalItems int         `json:"total_items"`
}

// Admission is one machine's admission record.
type Admission struct {
	MachineID           string     `json:"machine_id"`
	Hostname            string     `json:"hostname"`
	Fingerprint         string     `json:"fingerprint"`
	State               string     `json:"state"`
	Online              bool       `json:"online"`
	LastSeen            *time.Time `json:"last_seen,omitempty"`
	RequestedAt         *time.Time `json:"requested_at,omitempty"`
	AdmittedAt          *time.Time `json:"admitted_at,omitempty"`
	DecidedAt           *time.Time `json:"decided_at,omitempty"`
	Reason              string     `json:"reason,omitempty"`
	Previous            string     `json:"previous,omitempty"`
	PreviousFingerprint string     `json:"previous_fingerprint,omitempty"`
}

func admissionListFromGen(
	g *gen.ListAdmissionResponse,
) AdmissionList {
	items := make([]Admission, 0, len(g.Items))
	for _, a := range g.Items {
		items = append(items, admissionFromGen(a))
	}

	return AdmissionList{Items: items, TotalItems: g.TotalItems}
}

func admissionFromGen(
	g gen.Admission,
) Admission {
	return Admission{
		MachineID:           g.MachineId,
		Hostname:            g.Hostname,
		Fingerprint:         g.Fingerprint,
		State:               g.State,
		Online:              g.Online,
		LastSeen:            g.LastSeen,
		RequestedAt:         g.RequestedAt,
		AdmittedAt:          g.AdmittedAt,
		DecidedAt:           g.DecidedAt,
		Reason:              derefString(g.Reason),
		Previous:            derefString(g.Previous),
		PreviousFingerprint: derefString(g.PreviousFingerprint),
	}
}
