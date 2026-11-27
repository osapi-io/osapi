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

import (
	"crypto/fips140"
	"crypto/tls"
	"fmt"
	"runtime"
	"slices"

	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// TLS is a listener of a process and the TLS config that it serves: nil
// when it serves no TLS.
type TLS struct {
	Name   string
	Config *tls.Config
}

// mode is the FIPS 140-3 mode of this process. Tests swap it.
var mode = realMode

// realMode is the FIPS 140-3 mode that Go runs this process in.
func realMode() types.FIPSMode {
	return ModeOf(fips140.Enabled(), fips140.Enforced())
}

// ModeOf is the FIPS 140-3 mode of a process whose module is enabled, and
// enforced, as crypto/fips140 reports them.
func ModeOf(
	enabled bool,
	enforced bool,
) types.FIPSMode {
	switch {
	case enforced:
		return types.FIPSOnly
	case enabled:
		return types.FIPSOn
	default:
		return types.FIPSOff
	}
}

// New is the report of this process, with its listeners and its uses of
// crypto.
func New(
	listeners []TLS,
	uses []Algorithm,
) Report {
	m := mode()
	r := Report{
		Go: runtime.Version(), Module: fips140.Version(), Mode: m,
		Algorithms: uses, Enforceable: true,
	}
	if m == types.FIPSOff {
		r.Gaps = append(r.Gaps, "FIPS 140-3 mode is off: set GODEBUG=fips140=on")
	}
	for _, l := range listeners {
		rl := listener(l, m)
		r.Listeners = append(r.Listeners, rl)
		if rl.TLS && !rl.Approved {
			r.Gaps = append(r.Gaps, l.Name+": crypto/tls may negotiate a suite or group that is not approved")
		}
	}
	for _, a := range uses {
		if !a.Approved {
			r.Gaps = append(r.Gaps, a.Name+": "+a.Algorithm+" is not an approved algorithm")
		}
		r.Enforceable = r.Enforceable && a.OnlyMode
	}

	return r
}

// listener is l as m negotiates it. With no configured suites or groups,
// crypto/tls picks them: in FIPS 140-3 mode only approved ones, else ones
// such as X25519 and ChaCha20-Poly1305 too.
func listener(
	l TLS,
	m types.FIPSMode,
) Listener {
	if l.Config == nil {
		return Listener{Name: l.Name}
	}
	out := Listener{Name: l.Name, TLS: true, MinVersion: tls.VersionName(minVersion(l.Config))}
	suites := l.Config.CipherSuites
	if len(suites) == 0 {
		for _, s := range tls.CipherSuites() {
			suites = append(suites, s.ID)
		}
	}
	approved := true
	for _, id := range suites {
		out.Suites = append(out.Suites, tls.CipherSuiteName(id))
		approved = approved && slices.Contains(approvedSuites, id)
	}
	// crypto/tls does not export the groups that it offers by default, and
	// outside FIPS 140-3 mode they include X25519.
	if len(l.Config.CurvePreferences) == 0 {
		out.Groups = []string{defaultGroups}
		approved = false
	}
	for _, g := range l.Config.CurvePreferences {
		out.Groups = append(out.Groups, g.String())
		approved = approved && slices.Contains(approvedGroups, g)
	}
	out.Approved = m != types.FIPSOff || approved

	return out
}

// minVersion is the lowest version that c accepts: crypto/tls's TLS 1.2
// when c sets none.
func minVersion(
	c *tls.Config,
) uint16 {
	if c.MinVersion == 0 {
		return tls.VersionTLS12
	}

	return c.MinVersion
}

// approvedSuites are the suites of crypto/tls that SP 800-52 approves: the
// AES-GCM ones.
var approvedSuites = []uint16{
	tls.TLS_AES_128_GCM_SHA256, tls.TLS_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256, tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256, tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
}

// approvedGroups are the key exchange groups that SP 800-56A approves.
var approvedGroups = []tls.CurveID{
	tls.CurveP256, tls.CurveP384, tls.CurveP521, tls.SecP256r1MLKEM768, tls.SecP384r1MLKEM1024,
}

// defaultGroups is the name of the groups that crypto/tls offers when a
// config sets none.
const defaultGroups = "crypto/tls default"

// Verdict is r in one line, for a table: its mode, then approved when it has
// no gap, or the number of its gaps. FIPS 140-3 mode off, with no other gap,
// is off.
func (r Report) Verdict() string {
	gaps := len(r.Gaps)
	if r.Mode == types.FIPSOff {
		gaps--
	}
	switch {
	case gaps == 0 && r.Mode == types.FIPSOff:
		return string(r.Mode)
	case gaps == 0:
		return string(r.Mode) + ", approved"
	case gaps == 1:
		return string(r.Mode) + ", 1 gap"
	default:
		return fmt.Sprintf("%s, %d gaps", r.Mode, gaps)
	}
}
