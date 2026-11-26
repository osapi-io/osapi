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

import "time"

// The sample answers of the examples, which the docs page of each command
// shows as the CLI prints them. Their times are relative to the time that
// the docs generator fixes, 2026-01-15 10:00 UTC.
var (
	sampleRequested = time.Date(2026, time.January, 15, 9, 40, 0, 0, time.UTC)
	sampleSeen      = time.Date(2026, time.January, 15, 9, 59, 52, 0, time.UTC)

	samplePending = Admission{
		MachineID:     "d530664edc1c431eb5ef1e9dfd6617a5",
		Hostname:      "web-03",
		State:         "pending",
		Online:        true,
		LastSeen:      &sampleSeen,
		Fingerprint:   "SHA256:1f53dea6b0c3e2f9a8d4c71e5b6a0f2d9c8e7b6a5f4e3d2c1b0a9f8e7d6c5b4a",
		EncryptionKey: "BH0Fq2u2_BBIGuYw6KRZYg4I1pT5DACIZUCrhDK5Y7Phpl1MRex_y4Y60InurvbYIiZDQ2MHIcmOXF8SfLlA5fs",
		RequestedAt:   &sampleRequested,
	}

	sampleList = ListAdmissionResponse{
		TotalItems: 2,
		Items: []Admission{
			samplePending,
			{
				MachineID:     "8f14e45fceea167a5a36dedd4bea2543",
				Hostname:      "web-01",
				State:         "admitted",
				Online:        true,
				LastSeen:      &sampleSeen,
				Fingerprint:   "SHA256:9a7c3e1b5d2f4a6c8e0b1d3f5a7c9e2b4d6f8a0c2e4b6d8f0a1c3e5b7d9f1a3c",
				EncryptionKey: "BMZNeoFQM2Jn2FyzAJvsHXrFT1T8vNHfsPS1c09_uxIyhom9_kn_vJIRmuDL3MYmZjkWWwSbXSjzJEZOKzQrots",
				AdmittedAt:    &sampleRequested,
			},
		},
	}
)
