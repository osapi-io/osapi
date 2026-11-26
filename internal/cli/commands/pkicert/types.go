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
package pkicert

import "time"

// Source is a certificate that a server serves: its name, its files, and
// whether osapi made it.
type Source struct {
	Name       string
	Component  string
	CertFile   string
	KeyFile    string
	SelfSigned bool
}

// ListInput is the input of list, which has none.
type ListInput struct{}

// GetInput names a certificate.
type GetInput struct {
	Name string `json:"name" validate:"required,max=64" doc:"The name of the certificate, as list shows it, such as api or nats."`
}

// RotateInput names a certificate, and whether to rotate one that agents pin.
type RotateInput struct {
	Name  string `json:"name"  validate:"required,max=64" doc:"The name of the certificate, as list shows it, such as api or nats."`
	Force bool   `json:"force" validate:"omitempty"       doc:"Rotate the NATS server's certificate, which each join string pins and each joined agent trusts."`
}

// ListOutput is each certificate.
type ListOutput struct {
	Items []CertInfo `json:"items"`
}

// RotateOutput is the certificate that rotate made.
type RotateOutput struct {
	_    struct{} `title:"Rotated"`
	Cert CertInfo `                json:"cert"`
	Next string   `                json:"next"`
}

// CertInfo is a certificate as list and get show it.
type CertInfo struct {
	Name        string    `json:"name"`
	State       string    `json:"state"                   doc:"ok; not made, when the component has not started on this host; or unreadable, with the reason."`
	Component   string    `json:"component"`
	Source      string    `json:"source"                  doc:"self-signed, which osapi made, or cert_file, the operator's."`
	File        string    `json:"file"                                                                                                                         table:"-"`
	Names       []string  `json:"names"                                                                                                                        table:"-"`
	Expires     time.Time `json:"expires"`
	DaysLeft    *int      `json:"days_left,omitempty"                                                                                                                    label:"Days left"`
	Fingerprint string    `json:"fingerprint"                                                                                                                  table:"-"`
	Key         string    `json:"key"                     doc:"The algorithm and size of the public key, such as ECDSA P-256."`
	Signature   string    `json:"signature"               doc:"The algorithm that signed the certificate, such as ECDSA-SHA256."                               table:"-"`
	FIPS        *bool     `json:"fips_approved,omitempty" doc:"Whether FIPS 186-5 and SP 800-131A approve the key and the signature algorithms."                         label:"FIPS approved"`
}
