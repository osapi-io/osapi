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

package job

import (
	"crypto/ed25519"
	"sort"
	"strings"
)

// RegistrationSigningBytes returns the canonical bytes an agent signs when it
// registers, and the controller verifies before letting a registration decide
// where work goes.
//
// Only the fields that decide routing are covered, and every one of them is
// covered:
//
//   - machine ID, because it selects the stored key the signature is checked
//     against
//   - hostname, because a target names it
//   - fingerprint, because the fleet view reports it as this agent's key
//   - state, because a Pending agent is refused work and a forged Ready would
//     lift that
//   - labels, because a label target routes on them, so an unsigned label is a
//     way to attract another host's jobs
//
// The signature itself is excluded: it is carried beside these bytes, never
// inside them. Everything else in a registration — uptime, load, memory,
// conditions — is reporting rather than routing, and moves on every heartbeat.
//
// The encoding is built field by field in a fixed order with the labels sorted,
// rather than marshalled, so that both sides derive the same bytes no matter
// how a struct field is ordered or a map is walked.
func RegistrationSigningBytes(
	reg *AgentRegistration,
) []byte {
	if reg == nil {
		return nil
	}

	var b strings.Builder

	b.WriteString("machine_id=")
	b.WriteString(reg.MachineID)
	b.WriteString("\nhostname=")
	b.WriteString(reg.Hostname)
	b.WriteString("\nfingerprint=")
	b.WriteString(reg.Fingerprint)
	b.WriteString("\nstate=")
	b.WriteString(reg.State)
	b.WriteString("\n")

	keys := make([]string, 0, len(reg.Labels))
	for k := range reg.Labels {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		b.WriteString("label:")
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(reg.Labels[k])
		b.WriteString("\n")
	}

	return []byte(b.String())
}

// VerifyRegistration reports whether the registration's signature was produced
// by pubKey over its canonical bytes.
//
// It answers one question only. Whether pubKey is the right key for this agent
// is the caller's to establish from the store, and an unsigned registration is
// not verified.
func VerifyRegistration(
	reg *AgentRegistration,
	pubKey ed25519.PublicKey,
) bool {
	if reg == nil || len(reg.Signature) == 0 || len(pubKey) == 0 {
		return false
	}

	return ed25519.Verify(pubKey, RegistrationSigningBytes(reg), reg.Signature)
}
