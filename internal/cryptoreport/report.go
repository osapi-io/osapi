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
	"strings"

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

// Verdict is r in one line, for a table: its mode, then what is not
// approved, by kind. With nothing that is not approved it is the mode
// alone, off, or on or only with approved.
//
//	off
//	on · approved
//	off · 1 TLS listener not approved
//	only · 2 TLS listeners and 1 algorithm not approved
func (r Report) Verdict() string {
	var parts []string
	if n := r.notApprovedListeners(); n > 0 {
		parts = append(parts, count(n, "TLS listener"))
	}
	if n := r.notApprovedAlgorithms(); n > 0 {
		parts = append(parts, count(n, "algorithm"))
	}
	switch {
	case len(parts) > 0:
		return string(r.Mode) + " · " + strings.Join(parts, " and ") + " not approved"
	case r.Mode == types.FIPSOff:
		return string(r.Mode)
	default:
		return string(r.Mode) + " · approved"
	}
}

// Reasons are, as sentences, each reason that r is not in FIPS 140-3 mode
// with only approved algorithms, with what closes it: the gaps, then each
// approved algorithm that FIPS 140-3 only mode still refuses, with the
// issue that tracks it.
func (r Report) Reasons() []string {
	out := slices.Clone(r.Gaps)
	for _, a := range r.Algorithms {
		if a.Approved && !a.OnlyMode && a.Reason != "" {
			out = append(out, "FIPS 140-3 only mode refuses "+a.Name+": "+a.Reason)
		}
	}

	return out
}

// notApprovedListeners are the TLS listeners of r that negotiate a suite or
// a group that FIPS 140-3 does not approve.
func (r Report) notApprovedListeners() int {
	n := 0
	for _, l := range r.Listeners {
		if l.TLS && !l.Approved {
			n++
		}
	}

	return n
}

// notApprovedAlgorithms are the algorithms of r that FIPS 140-3 does not
// approve.
func (r Report) notApprovedAlgorithms() int {
	n := 0
	for _, a := range r.Algorithms {
		if !a.Approved {
			n++
		}
	}

	return n
}

// count is n of what, as a person reads it: 1 TLS listener, 2 TLS
// listeners.
func count(
	n int,
	what string,
) string {
	if n == 1 {
		return "1 " + what
	}

	return fmt.Sprintf("%d %ss", n, what)
}
