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

// The kinds of item.
const (
	// KindKey is a key, as osapi pki key list shows it.
	KindKey = "key"
	// KindCert is a certificate, as osapi pki cert list shows it.
	KindCert = "cert"
)

// ListInput is the input of list, which has none.
type ListInput struct{}

// ListOutput is each key and each certificate.
type ListOutput struct {
	_     struct{} `title:"Keys and certificates"`
	Items []Item   `                              json:"items"`
}

// Item is a key or a certificate of osapi on this host.
type Item struct {
	Kind        string     `json:"kind"                  doc:"key or cert."`
	Name        string     `json:"name"`
	Component   string     `json:"component"             doc:"The component that uses it: controller, agent, nats or vault."`
	State       string     `json:"state"                 doc:"ok; not made; off, for a JetStream store with no key; or unreadable, with the reason."`
	File        string     `json:"file"                                                                                                                     table:"-"`
	Algorithm   string     `json:"algorithm"             doc:"The algorithm of the key, such as Ed25519, or of the certificate's key, such as ECDSA P-256."`
	Fingerprint string     `json:"fingerprint,omitempty" doc:"The fingerprint of a public key or a certificate, or the ID of a secret key."                 table:"-"`
	Expires     *time.Time `json:"expires,omitempty"     doc:"When a certificate expires."`
}
