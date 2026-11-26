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

package pkilist

import "time"

var sampleExpires = time.Date(2031, 11, 25, 12, 0, 0, 0, time.UTC)

var sampleList = ListOutput{Items: []Item{
	{
		Kind: KindKey, Name: "controller", Component: "controller", State: "ok",
		File: "/etc/osapi/pki/controller.key", Algorithm: "Ed25519",
		Fingerprint: "SHA256:9c1f2e3d4b5a69788796a5b4c3d2e1f0a9b8c7d6e5f4a3b2c1d0e9f8a7b6c5d4",
	},
	{
		Kind: KindKey, Name: "jetstream", Component: "nats", State: "ok",
		File: "/etc/osapi/nats/jetstream.key", Algorithm: "ChaCha20-Poly1305",
		Fingerprint: "7f3a9c1e5b2d8f40",
	},
	{
		Kind: KindCert, Name: "api", Component: "controller", State: "ok",
		File: "/etc/osapi/pki/api.crt", Algorithm: "ECDSA P-256",
		Fingerprint: "SHA256:1f53dea6b0c3e2f9a8d4c71e5b6a0f2d9c8e7b6a5f4e3d2c1b0a9f8e7d6c5b4a",
		Expires:     &sampleExpires,
	},
}}
