// Copyright (c) 2024 John Dewey

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

// Package tlspolicy is the TLS that every connection of osapi negotiates:
// only the suites and the groups that FIPS 140-3 approves, also when the
// process does not run in FIPS 140-3 mode.
package tlspolicy

import (
	"crypto/tls"
	"slices"

	"golang.org/x/sys/cpu"
)

// Suites are the TLS 1.2 suites that SP 800-52 approves: the ECDHE ones with
// AES-GCM. crypto/tls takes no list for TLS 1.3, whose suites Server guards.
var Suites = []uint16{
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
}

// Groups are the key exchange groups that SP 800-56A approves, with the
// hybrid of P-256 and ML-KEM-768 first.
var Groups = []tls.CurveID{
	tls.SecP256r1MLKEM768, tls.CurveP256, tls.CurveP384, tls.CurveP521,
}

// tls13AESGCM are the TLS 1.3 suites with AES-GCM, which SP 800-52 approves.
var tls13AESGCM = []uint16{tls.TLS_AES_128_GCM_SHA256, tls.TLS_AES_256_GCM_SHA384}

// aesHardware is whether crypto/tls takes this CPU to have AES-GCM in
// hardware. With none, a server of crypto/tls prefers ChaCha20-Poly1305 for
// TLS 1.3. A var, so a test can take the other branch.
var aesHardware = cpu.X86.HasAES && cpu.X86.HasPCLMULQDQ ||
	cpu.ARM64.HasAES && cpu.ARM64.HasPMULL

// Client is the TLS config of a connection that osapi makes: TLS 1.2 or
// later, the approved suites and the approved groups. crypto/tls offers its
// TLS 1.3 suites in a fixed list, with AES-GCM first on a CPU with AES in
// hardware. A server of osapi, which Server makes, picks AES-GCM from them.
func Client() *tls.Config {
	return &tls.Config{
		MinVersion:       tls.VersionTLS12,
		CipherSuites:     slices.Clone(Suites),
		CurvePreferences: slices.Clone(Groups),
	}
}

// Server is the TLS config of a server of osapi: Client's, and a guard for
// TLS 1.3. crypto/tls picks ChaCha20-Poly1305 for TLS 1.3 when the client
// prefers it or offers no AES-GCM, or when this CPU has no AES in hardware,
// and no setting changes that outside FIPS 140-3 mode. The guard offers such
// a client TLS 1.2 only, with the approved suites. A client with none of them
// is refused.
func Server() *tls.Config {
	c := Client()
	c.GetConfigForClient = func(hello *tls.ClientHelloInfo) (*tls.Config, error) {
		if !chacha(hello) {
			return nil, nil
		}
		out := c.Clone()
		out.GetConfigForClient = nil
		out.MaxVersion = tls.VersionTLS12

		return out, nil
	}

	return c
}

// chacha is whether a server of crypto/tls with TLS 1.3 picks
// ChaCha20-Poly1305 for hello: the client offers TLS 1.3 and ChaCha20, and
// it offers no AES-GCM, or its first suite that crypto/tls knows is not one
// with AES-GCM, or this CPU has no AES in hardware.
func chacha(
	hello *tls.ClientHelloInfo,
) bool {
	if !slices.Contains(hello.SupportedVersions, tls.VersionTLS13) ||
		!slices.Contains(hello.CipherSuites, tls.TLS_CHACHA20_POLY1305_SHA256) {
		return false
	}
	if !aesHardware || !slices.ContainsFunc(hello.CipherSuites, isTLS13AESGCM) {
		return true
	}
	// A TLS 1.3 suite with AES-GCM is in the list, so a suite that
	// crypto/tls knows is.
	first := hello.CipherSuites[slices.IndexFunc(hello.CipherSuites, known)]

	return !isAESGCM(first)
}

// known is whether crypto/tls has the suite id.
func known(
	id uint16,
) bool {
	for _, s := range tls.CipherSuites() {
		if s.ID == id {
			return true
		}
	}
	for _, s := range tls.InsecureCipherSuites() {
		if s.ID == id {
			return true
		}
	}

	return false
}

// isAESGCM is whether the suite id uses AES-GCM, for TLS 1.2 or TLS 1.3.
func isAESGCM(
	id uint16,
) bool {
	return isTLS13AESGCM(id) || slices.Contains(Suites, id) ||
		id == tls.TLS_RSA_WITH_AES_128_GCM_SHA256 || id == tls.TLS_RSA_WITH_AES_256_GCM_SHA384
}

// isTLS13AESGCM is whether id is a TLS 1.3 suite with AES-GCM.
func isTLS13AESGCM(
	id uint16,
) bool {
	return slices.Contains(tls13AESGCM, id)
}

// Approved is whether c negotiates only what FIPS 140-3 approves: TLS 1.2 or
// later, only the approved suites and groups, and, for a server, the guard
// that Server sets. A nil c serves no TLS.
func Approved(
	c *tls.Config,
	server bool,
) bool {
	if c == nil {
		return true
	}
	if c.MinVersion < tls.VersionTLS12 || len(c.CipherSuites) == 0 || len(c.CurvePreferences) == 0 {
		return false
	}
	for _, s := range c.CipherSuites {
		if !slices.Contains(Suites, s) {
			return false
		}
	}
	for _, g := range c.CurvePreferences {
		if !slices.Contains(Groups, g) {
			return false
		}
	}

	return !server || c.GetConfigForClient != nil
}
