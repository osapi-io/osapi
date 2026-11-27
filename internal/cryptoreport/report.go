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

	"github.com/osapi-io/osapi/internal/tlspolicy"
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
			r.Gaps = append(
				r.Gaps,
				l.Name+": crypto/tls may negotiate a suite or group that is not approved",
			)
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

// listener is what the server l negotiates, in FIPS 140-3 mode m. With
// tlspolicy's suites, groups and TLS 1.3 guard it is approved in each mode.
// With crypto/tls's defaults it is approved only in FIPS 140-3 mode, which
// restricts them: outside it they include X25519 and ChaCha20-Poly1305.
func listener(
	l TLS,
	m types.FIPSMode,
) Listener {
	if l.Config == nil {
		return Listener{Name: l.Name}
	}
	pinned := tlspolicy.Approved(l.Config, true)
	out := Listener{
		Name: l.Name, TLS: true, MinVersion: tls.VersionName(minVersion(l.Config)),
		Approved: m != types.FIPSOff || pinned,
	}
	for _, id := range suites(l.Config, pinned) {
		out.Suites = append(out.Suites, tls.CipherSuiteName(id))
	}
	if len(l.Config.CurvePreferences) == 0 {
		// crypto/tls does not export the groups that it offers by default.
		out.Groups = []string{defaultGroups}
	}
	for _, g := range l.Config.CurvePreferences {
		out.Groups = append(out.Groups, g.String())
	}

	return out
}

// suites are the suites that a server with c can pick, for TLS 1.3 and TLS
// 1.2. crypto/tls takes no list for TLS 1.3: with tlspolicy's guard, a client
// gets only its AES-GCM suites, and with no guard, ChaCha20-Poly1305 too.
func suites(
	c *tls.Config,
	pinned bool,
) []uint16 {
	if len(c.CipherSuites) == 0 {
		all := tls.CipherSuites()
		out := make([]uint16, 0, len(all))
		for _, s := range all {
			out = append(out, s.ID)
		}

		return out
	}
	out := []uint16{tls.TLS_AES_128_GCM_SHA256, tls.TLS_AES_256_GCM_SHA384}
	if !pinned {
		out = append(out, tls.TLS_CHACHA20_POLY1305_SHA256)
	}

	return append(out, c.CipherSuites...)
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
