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

package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/controller/endpoint/command"
	"github.com/osapi-io/osapi/internal/controller/endpoints/admission"
)

// The admission group's commands are generated from its definition. These
// are how each answer prints without --json.
func init() {
	command.Render(admission.Group, "list", renderAdmissionList)
	command.Render(
		admission.Group,
		"get",
		func(_ *cobra.Command, _ admission.MachineInput, out admission.Admission) {
			renderAdmission(out)
		},
	)
	// admit, reject, revoke and delete answer 204, with no body: each prints what
	// it did, with --json too.
	command.Acts(
		admission.Group,
		"admit",
		func(_ *cobra.Command, in admission.AdmitInput, _ struct{}) { admissionDone(in.MachineID, "admitted") },
	)
	command.Acts(
		admission.Group,
		"reject",
		func(_ *cobra.Command, in admission.RejectInput, _ struct{}) { admissionDone(in.MachineID, "rejected") },
	)
	command.Acts(
		admission.Group,
		"delete",
		func(_ *cobra.Command, in admission.MachineInput, _ struct{}) { admissionDone(in.MachineID, "deleted") },
	)
	command.Acts(
		admission.Group,
		"revoke",
		func(_ *cobra.Command, in admission.MachineInput, _ struct{}) { admissionDone(in.MachineID, "revoked") },
	)
}

func renderAdmissionList(
	_ *cobra.Command,
	_ admission.ListInput,
	out admission.ListAdmissionResponse,
) {
	if len(out.Items) == 0 {
		cli.PrintHeading("Admission (0)")
		cli.PrintNote("None.")

		return
	}

	rows := make([][]string, 0, len(out.Items))
	for _, a := range out.Items {
		rows = append(rows, []string{
			a.MachineID,
			a.Hostname,
			a.State,
			online(a.Online),
			since(a.LastSeen),
			shortFingerprint(a.Fingerprint),
		})
	}

	cli.PrintCompactTable([]cli.Section{{
		Title:   fmt.Sprintf("Admission (%d)", out.TotalItems),
		Headers: []string{"MACHINE ID", "HOSTNAME", "STATE", "ONLINE", "LAST SEEN", "FINGERPRINT"},
		Rows:    rows,
	}})
}

// renderAdmission is one machine's record, as admission get prints it.
func renderAdmission(
	a admission.Admission,
) {
	cli.PrintHeading("Admission")

	kv := []string{
		"Machine ID", a.MachineID,
		"Hostname", a.Hostname,
		"State", a.State,
		"Fingerprint", a.Fingerprint,
		"Online", online(a.Online),
		"Last Seen", since(a.LastSeen),
		"Requested", since(a.RequestedAt),
	}
	if a.AdmittedAt != nil {
		kv = append(kv, "Admitted", since(a.AdmittedAt))
	}
	if a.DecidedAt != nil {
		kv = append(kv, "Decided", since(a.DecidedAt))
	}
	if a.Reason != "" {
		kv = append(kv, "Reason", a.Reason)
	}
	if a.Previous != "" {
		kv = append(kv, "Previous", a.Previous, "Previous Fingerprint", a.PreviousFingerprint)
	}
	cli.PrintKV(kv...)
}

func admissionDone(
	machineID string,
	state string,
) {
	if jsonOutput {
		result, _ := json.Marshal(map[string]string{"machine_id": machineID, "state": state})
		cli.PrintJSON(result)

		return
	}

	cli.PrintHeading("Admission")
	cli.PrintKV("Machine ID", machineID, "State", state)
}

func online(
	on bool,
) string {
	if on {
		return "yes"
	}

	return "no"
}

// since is the age of t, or "never" when there is none.
func since(
	t *time.Time,
) string {
	if t == nil {
		return "never"
	}

	return cli.FormatAge(time.Since(*t))
}

func shortFingerprint(
	fp string,
) string {
	if len(fp) > 20 {
		return fp[:20] + "..."
	}

	return fp
}
