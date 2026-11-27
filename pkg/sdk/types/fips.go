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

package types

import "slices"

// FIPSMode is the FIPS 140-3 mode of a process, one of its Values. Go sets
// it once, at the start of the process, from GODEBUG=fips140.
type FIPSMode string

// The FIPS 140-3 modes of a process.
const (
	// FIPSOff uses the Go Cryptographic Module with no FIPS 140-3 mode.
	FIPSOff FIPSMode = "off"
	// FIPSOn runs the module in FIPS 140-3 mode: it does its self-tests,
	// and crypto/tls negotiates only approved versions, suites and groups.
	FIPSOn FIPSMode = "on"
	// FIPSOnly is FIPS 140-3 mode that also refuses each algorithm that is
	// not approved.
	FIPSOnly FIPSMode = "only"
)

// Values are each FIPS 140-3 mode of a process.
func (FIPSMode) Values() []FIPSMode {
	return []FIPSMode{FIPSOff, FIPSOn, FIPSOnly}
}

// Valid is whether m is one of Values.
func (m FIPSMode) Valid() bool {
	return slices.Contains(m.Values(), m)
}
