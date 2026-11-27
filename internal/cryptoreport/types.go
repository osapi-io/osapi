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

import "github.com/osapi-io/osapi/pkg/sdk/types"

// Report is the crypto of one process.
type Report struct {
	_          struct{}       `schema:"CryptoReport"`
	Go         string         `json:"go"                   doc:"The Go version that built the process."                                                                                                        example:"go1.27.3"`
	Module     string         `json:"module"               doc:"The version of the Go Cryptographic Module in the process."                                                                                   example:"v1.0.0"`
	Mode       types.FIPSMode `json:"mode"                 doc:"The FIPS 140-3 mode of the process: off, on, or only, which also refuses each algorithm that is not approved. GODEBUG=fips140 sets it at the start."  example:"off"`
	Enforceable bool          `json:"enforceable"          doc:"Whether each algorithm of the process works in FIPS 140-3 only mode, GODEBUG=fips140=only." label:"Only mode"`
	Listeners  []Listener     `json:"listeners,omitempty"  doc:"Each TLS listener of the process."`
	Algorithms []Algorithm    `json:"algorithms,omitempty" doc:"Each algorithm that the process uses."`
	Gaps       []string       `json:"gaps,omitempty"       doc:"What keeps the process from FIPS 140-3 mode with only approved algorithms. Empty when the mode is on and each algorithm and listener is approved."`
}

// Listener is a TLS listener of a process.
type Listener struct {
	_          struct{} `schema:"CryptoListener"`
	Name       string   `json:"name"                doc:"The listener, such as api, nats or metrics."                                                                       example:"api"`
	TLS        bool     `json:"tls"                 doc:"Whether the listener serves TLS. A listener with no TLS has no other fields."`
	MinVersion string   `json:"min_version,omitempty" doc:"The lowest TLS version that the listener accepts."                                                              example:"TLS 1.2"`
	Suites     []string `json:"suites,omitempty"    doc:"The cipher suites that the listener can negotiate: the configured ones, or those of crypto/tls for the process's mode."`
	Groups     []string `json:"groups,omitempty"    doc:"The key exchange groups that the listener can negotiate: the configured ones, or those of crypto/tls for the process's mode."`
	Approved   bool     `json:"approved"            doc:"Whether the listener negotiates only approved versions, suites and groups: true in FIPS 140-3 mode, where crypto/tls restricts them."`
}

// Algorithm is a use of crypto in a process.
type Algorithm struct {
	_         struct{} `schema:"CryptoAlgorithm"`
	Name      string   `json:"name"             doc:"What the process uses it for."                                                                     example:"secrets sealed to agents"`
	Algorithm string   `json:"algorithm"        doc:"The algorithm, with its key size or curve."                                                       example:"ECDH P-256, HKDF-SHA256, AES-256-GCM"`
	Approved  bool     `json:"approved"         doc:"Whether FIPS 140-3 approves the algorithm."`
	OnlyMode  bool     `json:"only_mode"        doc:"Whether it works in FIPS 140-3 only mode. An approved algorithm can still fail there, when its code calls the module in a way that the mode refuses."`
	Reason    string   `json:"reason,omitempty" doc:"Why it is not approved, or why it fails in FIPS 140-3 only mode."`
}
