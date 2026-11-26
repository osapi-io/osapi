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

package pkikey

var (
	sampleController = Key{
		Name: "controller", Kind: KindSigning, Component: "controller", State: StateOK,
		File: "/etc/osapi/pki/controller.key", Algorithm: "Ed25519",
		Fingerprint: "SHA256:9c1f2e3d4b5a69788796a5b4c3d2e1f0a9b8c7d6e5f4a3b2c1d0e9f8a7b6c5d4",
	}
	sampleAgent = Key{
		Name: "agent", Kind: KindSigning, Component: "agent", State: StateOK,
		File: "/etc/osapi/pki/agent/agent.key", Algorithm: "Ed25519",
		Fingerprint: "SHA256:4b2a7c6d5e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b",
		Admit: "osapi client admission admit --machine-id 4c4c4544-0035-4b10-8031-b4c04f4d3332 " +
			"--fingerprint SHA256:4b2a7c6d5e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b",
	}
	sampleJetStream = Key{
		Name: "jetstream", Kind: KindEncryption, Component: "nats", State: StateOK,
		File: "/etc/osapi/nats/jetstream.key", Algorithm: "ChaCha20-Poly1305",
		KeyID: "7f3a9c1e5b2d8f40",
	}
	sampleVault = Key{
		Name:      "vault:default",
		Kind:      KindEncryption,
		Component: "vault",
		State:     StateOK,
		File:      "/etc/osapi/secrets.key",
		Algorithm: "AES-256-GCM",
		KeyID:     "b8e14a2c9d3f6071",
		Keys:      1,
	}
	sampleList = ListOutput{Items: []Key{
		sampleController,
		sampleAgent,
		{
			Name:      "agent-encryption",
			Kind:      KindEncryption,
			Component: "agent",
			State:     StateOK,
			File:      "/etc/osapi/pki/agent/agent.ekey",
			Algorithm: "ECDH P-256",
			PublicKey: "BH0Fq2u2_BBIGuYw6KRZYg4I1pT5DACIZUCrhDK5Y7Phpl1MRex_y4Y60InurvbYIiZDQ2MHIcmOXF8SfLlA5fs",
		},
		sampleJetStream,
		sampleVault,
	}}
	sampleRotate = RotateOutput{Key: sampleVault, Values: 12}
)
