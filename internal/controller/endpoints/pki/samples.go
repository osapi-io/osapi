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

import (
	"time"

	"github.com/osapi-io/osapi/internal/cryptoreport"
)

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
	sampleCertRotationStarted = CertRotation{
		Current:   "SHA256:4b7d20e1c9f8a7b6d5e4c3b2a1f0e9d8c7b6a5f4e3d2c1b0a9f8e7d6c5b4a3f2",
		Next:      "SHA256:a83f61c2e5d4c3b2a1f0e9d8c7b6a5f4e3d2c1b0a9f8e7d6c5b4a3f2e1d0c9b8",
		StartedAt: sampleStarted,
		SwitchBy:  sampleStarted.Add(24 * time.Hour),
		Agents: []CertRotationAgent{
			{Hostname: "web-01", MachineID: "d530664edc1c431eb5ef1e9dfd6617a5"},
		},
	}
	sampleCertRotation = CertRotation{
		Current:   sampleCertRotationStarted.Current,
		Next:      sampleCertRotationStarted.Next,
		StartedAt: sampleStarted,
		SwitchBy:  sampleCertRotationStarted.SwitchBy,
		Agents: []CertRotationAgent{
			{Hostname: "web-01", MachineID: "d530664edc1c431eb5ef1e9dfd6617a5", HasNext: true},
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

// sampleTLS is what pki tls list prints for the controller's host.
var sampleTLS = TLSList{
	Items: []TLS{
		{
			Component: ComponentController, Hostname: "osapi-controller",
			Summary: "FIPS off · cert 78d", FIPS: "off", Expiry: ExpiryOK,
			Reasons: []string{"FIPS 140-3 mode is off: set GODEBUG=fips140=on"},
			Certificates: []TLSCert{{
				Name: "api", State: "ok", Subject: "CN=osapi-controller",
				Names: []string{"osapi-controller", "127.0.0.1"}, Issuer: "CN=osapi-controller",
				Source: "self-signed", NotBefore: sampleCert.Expires.AddDate(-1, 0, 0),
				Expires: sampleCert.Expires, DaysLeft: &sampleDays, Expiry: ExpiryOK,
				Key: "ECDSA P-256", FIPS: &sampleApproved,
			}},
			Listeners: []cryptoreport.Listener{
				{
					Name:       "api",
					TLS:        true,
					MinVersion: "TLS 1.2",
					Approved:   true,
					Suites: []string{
						"TLS_AES_128_GCM_SHA256",
						"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
					},
					Groups: []string{"SecP256r1MLKEM768", "CurveP256"},
				},
			},
			Go: "go1.27.3", Module: "v1.0.0",
			Algorithms: []cryptoreport.Algorithm{cryptoreport.Seal, cryptoreport.Vault},
		},
	},
	Total: 1,
}
