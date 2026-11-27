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

package pki

import "time"

var (
	sampleDays     = 78
	sampleApproved = true
	sampleCert     = Cert{
		Hostname: "web-01", Name: "agent-metrics", State: "ok", Component: "agent",
		Source: "self-signed", File: "/etc/osapi/pki/agent/agent-metrics.crt",
		Names:       []string{"web-01", "10.0.0.11"},
		Expires:     time.Date(2027, 2, 12, 19, 40, 1, 0, time.UTC),
		DaysLeft:    &sampleDays,
		Fingerprint: "SHA256:3f1c0d9e4b2a7c6d5e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d",
		Key:         "ECDSA P-256", Signature: "ECDSA-SHA256", FIPS: &sampleApproved,
	}
	sampleList = CertList{
		Items: []Cert{
			{
				Hostname: "osapi-controller", Name: "api", State: "ok", Component: "controller",
				Source: "self-signed", Expires: sampleCert.Expires, DaysLeft: &sampleDays,
				Key: "ECDSA P-256", FIPS: &sampleApproved,
			},
			sampleCert,
		},
		Total: 2,
	}
	sampleStarted         = time.Date(2026, 11, 26, 12, 0, 0, 0, time.UTC)
	sampleRotationStarted = Rotation{
		Current:   "SHA256:1f53dea6b0c3e2f9a8d4c71e5b6a0f2d9c8e7b6a5f4e3d2c1b0a9f8e7d6c5b4a",
		Next:      "SHA256:9c2e41b07d3a8f6e5c4b3a2918f7e6d5c4b3a29180f7e6d5c4b3a2918f7e6d5c",
		StartedAt: sampleStarted,
		SwitchBy:  sampleStarted.Add(24 * time.Hour),
		Agents: []RotationAgent{
			{Hostname: "web-01", MachineID: "d530664edc1c431eb5ef1e9dfd6617a5"},
		},
	}
	sampleRotation = Rotation{
		Current:   sampleRotationStarted.Current,
		Next:      sampleRotationStarted.Next,
		StartedAt: sampleStarted,
		SwitchBy:  sampleRotationStarted.SwitchBy,
		Agents: []RotationAgent{
			{Hostname: "web-01", MachineID: "d530664edc1c431eb5ef1e9dfd6617a5", HasNext: true},
		},
	}
)
