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

package cryptoreport

import "github.com/osapi-io/osapi/internal/pki"

// The uses of crypto in osapi. OnlyMode is a claim that a check holds: a
// use that works in FIPS 140-3 only mode is in a package that the
// fips140=only CI job tests (go_fips_packages in the justfile), and a use
// that does not has a test in that job that shows the failure, so a fix
// upstream fails the test until the use says so.
var (
	// APITokens sign and verify the API's bearer tokens.
	APITokens = Algorithm{
		Name: "API tokens", Algorithm: "HMAC-SHA256 (HS256)", Approved: true, OnlyMode: true,
	}
	// Signatures sign and verify jobs, joins, admissions and registrations.
	Signatures = Algorithm{
		Name: "job, join and admission signatures", Algorithm: "Ed25519",
		Approved: true, OnlyMode: true,
	}
	// Seal seals each secret of a job to the agents that run it.
	Seal = Algorithm{
		Name: "secrets sealed to agents", Algorithm: "ECDH P-256, HKDF-SHA256, AES-256-GCM",
		Approved: true, OnlyMode: true,
	}
	// Vault encrypts the values of the controller's vault.
	Vault = Algorithm{
		Name: "controller vault", Algorithm: "AES-256-GCM", Approved: true, OnlyMode: true,
	}
	// NATSAuth signs and verifies the NATS operator, account and user JWTs.
	NATSAuth = Algorithm{
		Name: "NATS operator, account and user JWTs", Algorithm: "Ed25519 (nkeys)",
		Approved: true,
		Reason: "nkeys v0.4.17 makes a key with ed25519.GenerateKey from a seed, " +
			"which FIPS 140-3 only mode refuses (osapi#1527)",
	}
	// Store encrypts the JetStream store of the NATS server.
	Store = Algorithm{
		Name: "JetStream store", Algorithm: "AES-256-GCM (nats-server)", Approved: true,
		Reason: "nats-server picks its own GCM nonce, which FIPS 140-3 only mode refuses (osapi#1549)",
	}
)

// Certificates are the uses of the certificates that a process serves: each
// with its key and signature algorithms, approved when pki approves both.
func Certificates(
	certs []pki.Cert,
) []Algorithm {
	var out []Algorithm
	for _, c := range certs {
		if c.State != pki.CertOK {
			continue
		}
		ok := c.FIPS != nil && *c.FIPS
		a := Algorithm{
			Name:      "certificate " + c.Name,
			Algorithm: c.Key + ", " + c.Signature,
			Approved:  ok, OnlyMode: ok,
		}
		if !ok {
			a.Reason = "the key or the signature algorithm is not approved"
		}
		out = append(out, a)
	}

	return out
}
