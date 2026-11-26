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
package pkicert

import "time"

// The sample outputs of the examples, which the docs page of each cert
// group shows.
var (
	SampleCert = CertInfo{
		Name:        "api",
		State:       "ok",
		Component:   "controller",
		Source:      "self-signed",
		File:        "/etc/osapi/pki/api.crt",
		Names:       []string{"localhost", "controller.example.com", "127.0.0.1"},
		Expires:     time.Date(2031, 11, 25, 12, 0, 0, 0, time.UTC),
		DaysLeft:    new(1825),
		Fingerprint: "SHA256:1f53dea6b0c3e2f9a8d4c71e5b6a0f2d9c8e7b6a5f4e3d2c1b0a9f8e7d6c5b4a",
		Key:         "ECDSA P-256",
		Signature:   "ECDSA-SHA256",
		FIPS:        new(true),
	}
	SampleList = ListOutput{Items: []CertInfo{SampleCert, {
		Name: "agent-metrics", State: "not made", Component: "agent", Source: "self-signed",
		File: "/etc/osapi/pki/agent/agent-metrics.crt",
	}}}
	SampleRotate = RotateOutput{Cert: SampleCert, Next: "restart the server to serve it"}
)
