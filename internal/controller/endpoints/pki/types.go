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

package pki

import "time"

// ListInput filters the certificates: a hostname gives only that host's.
type ListInput struct {
	Hostname string `json:"hostname,omitempty" in:"query" validate:"omitempty,min=1,max=255" doc:"Only the certificates of this host: the controller's hostname, or an agent's."`
}

// GetInput names one certificate of one host.
type GetInput struct {
	Name     string `json:"name"               in:"path"  validate:"required,min=1,max=64"   doc:"The name of the certificate, such as api or agent-metrics."`
	Hostname string `json:"hostname,omitempty" in:"query" validate:"omitempty,min=1,max=255" doc:"The host of the certificate. Empty is the controller's host, which has the certificates of the controller."`
}

// CertList is the certificate of each host, with the total.
type CertList struct {
	_     struct{} `schema:"PKICertList" title:"Certs"`
	Items []Cert   `                                   json:"items"`
	Total int      `                                   json:"total_items" doc:"The number of certificates."`
}

// Cert is one certificate of one host: the controller's own, or one that an
// agent reports in its registration.
type Cert struct {
	_           struct{}  `schema:"PKICert" title:"Cert"`
	Hostname    string    `                              json:"hostname"                doc:"The host that serves the certificate."                                                         example:"web-01"`
	Name        string    `                              json:"name"                    doc:"The name of the certificate, such as api or agent-metrics."                                    example:"agent-metrics"`
	State       string    `                              json:"state"                   doc:"ok; not made, when the component has not started on the host; or unreadable, with the reason." example:"ok"`
	Component   string    `                              json:"component"               doc:"The component that serves it: controller, nats or agent."                                      example:"agent"`
	Source      string    `                              json:"source"                  doc:"self-signed, which osapi made, or cert_file, the operator's."                                  example:"self-signed"`
	File        string    `                              json:"file"                    doc:"The certificate's file on the host."                                                                                   table:"-"`
	Names       []string  `                              json:"names,omitempty"         doc:"The DNS names and the IP addresses that the certificate is for."                                                       table:"-"`
	Expires     time.Time `                              json:"expires,omitzero"        doc:"When the certificate expires."`
	DaysLeft    *int      `                              json:"days_left,omitempty"     doc:"The days that are left until it expires."                                                                                        label:"Days left"`
	Fingerprint string    `                              json:"fingerprint,omitempty"   doc:"The SHA-256 fingerprint of the certificate."                                                                           table:"-"`
	Key         string    `                              json:"key,omitempty"           doc:"The algorithm and size of the public key, such as ECDSA P-256."`
	Signature   string    `                              json:"signature,omitempty"     doc:"The algorithm that signed the certificate, such as ECDSA-SHA256."                                                      table:"-"`
	FIPS        *bool     `                              json:"fips_approved,omitempty" doc:"Whether FIPS 186-5 and SP 800-131A approve the key and the signature algorithms."                                                label:"FIPS approved"`
}
