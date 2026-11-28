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

import (
	"time"

	"github.com/osapi-io/osapi/internal/cryptoreport"
)

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
	Hostname    string    `                              json:"hostname"                doc:"The host that serves the certificate."                                                            example:"web-01"`
	Name        string    `                              json:"name"                    doc:"The name of the certificate, such as api or agent-metrics."                                       example:"agent-metrics"`
	State       string    `                              json:"state"                   doc:"ok; not made, when the component has not started on the host; or unreadable, with the reason."    example:"ok"`
	Component   string    `                              json:"component"               doc:"The component that serves it: controller, nats or agent."                                         example:"agent"`
	Source      string    `                              json:"source"                  doc:"self-signed, which osapi made, or cert_file, the operator's."                                     example:"self-signed"`
	File        string    `                              json:"file"                    doc:"The certificate's file on the host."                                                                                      table:"-"`
	Names       []string  `                              json:"names,omitempty"         doc:"The DNS names and the IP addresses that the certificate is for."                                                          table:"-"`
	Subject     string    `                              json:"subject,omitempty"       doc:"The subject of the certificate, as a distinguished name."                                                                 table:"-"`
	Issuer      string    `                              json:"issuer,omitempty"        doc:"The issuer of the certificate, as a distinguished name. The subject again for a self-signed one."                         table:"-"`
	NotBefore   time.Time `                              json:"not_before,omitzero"     doc:"When the certificate starts to be valid."                                                                                 table:"-"`
	Expires     time.Time `                              json:"expires,omitzero"        doc:"When the certificate expires."`
	DaysLeft    *int      `                              json:"days_left,omitempty"     doc:"The days that are left until it expires."                                                                                           label:"Days left"`
	Fingerprint string    `                              json:"fingerprint,omitempty"   doc:"The SHA-256 fingerprint of the certificate."                                                                              table:"-"`
	Key         string    `                              json:"key,omitempty"           doc:"The algorithm and size of the public key, such as ECDSA P-256."`
	Signature   string    `                              json:"signature,omitempty"     doc:"The algorithm that signed the certificate, such as ECDSA-SHA256."                                                         table:"-"`
	FIPS        *bool     `                              json:"fips_approved,omitempty" doc:"Whether FIPS 186-5 and SP 800-131A approve the key and the signature algorithms."                                                   label:"FIPS approved"`
}

// RotationInput is the input of key get: it has no field.
type RotationInput struct{}

// RotateInput is the input of key rotate: it has no field.
type RotateInput struct{}

// Rotation is the controller's key and its rotation.
type Rotation struct {
	_         struct{}        `schema:"PKIKeyRotation" title:"Controller key"`
	Current   string          `                                               json:"current"             doc:"The fingerprint of the key that the controller signs jobs with."                                                                                       example:"SHA256:1f53dea6"`
	Next      string          `                                               json:"next,omitempty"      doc:"The fingerprint of the key that the controller switches to. Empty when no rotation is in progress."                                                    example:"SHA256:9c2e41b0"`
	StartedAt time.Time       `                                               json:"started_at,omitzero" doc:"When the rotation started."`
	SwitchBy  time.Time       `                                               json:"switch_by,omitzero"  doc:"When the controller switches at the latest: the start and controller.pki.rotation_grace_period. It switches before, once each agent has the next key."                           label:"Switch by"`
	Agents    []RotationAgent `                                               json:"agents,omitempty"    doc:"The admitted agents that the switch waits for: each one that reported within the grace period."`
}

// CertRotationInput is the input of nats get: it has no field.
type CertRotationInput struct{}

// CertRotateInput is the input of nats rotate: it has no field.
type CertRotateInput struct{}

// CertRotation is the NATS server's certificate and its rotation.
type CertRotation struct {
	_         struct{}            `schema:"PKINATSCertRotation" title:"NATS certificate"`
	Current   string              `                                                      json:"current"             doc:"The fingerprint of the certificate that the NATS server serves."                                                                                              example:"SHA256:4b7d20e1"`
	Next      string              `                                                      json:"next,omitempty"      doc:"The fingerprint of the certificate that the server switches to. Empty when no rotation is in progress."                                                       example:"SHA256:a83f61c2"`
	StartedAt time.Time           `                                                      json:"started_at,omitzero" doc:"When the rotation started."`
	SwitchBy  time.Time           `                                                      json:"switch_by,omitzero"  doc:"When the server switches at the latest: the start and controller.pki.rotation_grace_period. It switches before, once each agent trusts the next certificate."                           label:"Switch by"`
	Agents    []CertRotationAgent `                                                      json:"agents,omitempty"    doc:"The admitted agents that the switch waits for: each one that reported within the grace period."`
}

// CertRotationAgent is one agent that the switch of a certificate's
// rotation waits for.
type CertRotationAgent struct {
	_         struct{} `schema:"PKINATSCertRotationAgent" title:"Agent"`
	Hostname  string   `                                                json:"hostname"   doc:"The hostname of the agent."                                     example:"web-01"`
	MachineID string   `                                                json:"machine_id" doc:"The machine ID of the agent."                                   example:"d530664edc1c431eb5ef1e9dfd6617a5"`
	HasNext   bool     `                                                json:"has_next"   doc:"Whether the agent reports that it trusts the next certificate."                                            label:"Has next"`
}

// RotationAgent is one agent that the switch of a rotation waits for.
type RotationAgent struct {
	_         struct{} `schema:"PKIKeyRotationAgent" title:"Agent"`
	Hostname  string   `                                           json:"hostname"   doc:"The hostname of the agent."                          example:"web-01"`
	MachineID string   `                                           json:"machine_id" doc:"The machine ID of the agent."                        example:"d530664edc1c431eb5ef1e9dfd6617a5"`
	HasNext   bool     `                                           json:"has_next"   doc:"Whether the agent reports that it has the next key."                                            label:"Has next"`
}

// TLSListInput filters the TLS sections: by host, by component, or both.
type TLSListInput struct {
	Hostname  string    `json:"hostname,omitempty"  in:"query" validate:"omitempty,min=1,max=255" doc:"Only the components of this host: the controller's hostname, or an agent's."`
	Component Component `json:"component,omitempty" in:"query" validate:"omitempty,enum"          doc:"Only this component: controller, nats or agent."`
}

// TLSGetInput names the TLS section of one component of one host.
type TLSGetInput struct {
	Component Component `json:"component"          in:"path"  validate:"required,enum"           doc:"The component: controller, nats or agent."`
	Hostname  string    `json:"hostname,omitempty" in:"query" validate:"omitempty,min=1,max=255" doc:"The host of the component. Empty is the controller's host, which an agent's section must name."`
}

// Component is a component of osapi that serves TLS.
type Component string

// The components that serve TLS.
const (
	ComponentController Component = "controller"
	ComponentNATS       Component = "nats"
	ComponentAgent      Component = "agent"
)

// Expiry is how near a certificate is to its end: ok; soon, within
// expirySoonDays; or expired. Unknown when its component reports no
// certificate that it can read.
type Expiry string

// The expiries of a certificate.
const (
	ExpiryOK      Expiry = "ok"
	ExpirySoon    Expiry = "soon"
	ExpiryExpired Expiry = "expired"
	ExpiryUnknown Expiry = "unknown"
)

// TLSList is the TLS section of each component, with the total.
type TLSList struct {
	_     struct{} `schema:"PKITLSList" title:"TLS"`
	Items []TLS    `                                json:"items"`
	Total int      `                                json:"total_items" doc:"The number of components."`
}

// TLS is the TLS section of one component of one host: its FIPS 140-3 mode
// with the reasons it is not approved, each certificate that it serves, each
// TLS listener, and the versions and the algorithms of its crypto.
type TLS struct {
	_            struct{}                 `schema:"PKITLSSection" title:"TLS"`
	Component    Component                `                                   json:"component"              doc:"The component: controller, nats or agent."                                                                                                                              example:"controller"`
	Hostname     string                   `                                   json:"hostname"               doc:"The host of the component."                                                                                                                                             example:"osapi-controller"`
	Summary      string                   `                                   json:"summary"                doc:"The section in one line, for a table: the FIPS 140-3 verdict, then the days left of the certificate that expires first."                                                example:"FIPS off · cert 62d"`
	FIPS         string                   `                                   json:"fips"                   doc:"The FIPS 140-3 mode, then what is not approved, by kind, or approved."                                                                                                  example:"off"                 label:"FIPS"`
	Expiry       Expiry                   `                                   json:"expiry"                 doc:"How near the certificate that expires first is to its end: ok, soon (30 days or less), expired, or unknown when the component reports no certificate that it can read." example:"ok"                               render:"status"`
	Reasons      []string                 `                                   json:"reasons,omitempty"      doc:"Each reason that the component is not in FIPS 140-3 mode with only approved algorithms, as a sentence with what closes it, or with the issue that tracks it."                                                                      table:"-"`
	Certificates []TLSCert                `                                   json:"certificates,omitempty" doc:"Each certificate that the component serves, from the certificates of pki cert list."                                                                                                                                               table:"-"`
	Listeners    []cryptoreport.Listener  `                                   json:"listeners,omitempty"    doc:"Each TLS listener of the component, with its lowest TLS version, its suites and its groups, and whether FIPS 140-3 approves them."                                                                                                 table:"-"`
	Go           string                   `                                   json:"go,omitempty"           doc:"The Go version that built the component."                                                                                                                               example:"go1.27.3"                                         table:"-"`
	Module       string                   `                                   json:"module,omitempty"       doc:"The version of the Go Cryptographic Module in the component."                                                                                                           example:"v1.0.0"                                           table:"-"`
	Algorithms   []cryptoreport.Algorithm `                                   json:"algorithms,omitempty"   doc:"Each algorithm that the component uses, with whether FIPS 140-3 approves it."                                                                                                                                                      table:"-"`
}

// TLSCert is a certificate that a component serves, as its TLS section shows
// it: who it is for, who issued it, when it is valid, and how near its end
// is. It is a certificate of pki cert list.
type TLSCert struct {
	_         struct{}  `schema:"PKITLSCert" title:"Certificates"`
	Name      string    `                                         json:"name"                    doc:"The name of the certificate, such as api or agent-metrics."                                         example:"api"`
	State     string    `                                         json:"state"                   doc:"ok; not made, when the component has not started on the host; or unreadable, with the reason."      example:"ok"`
	Subject   string    `                                         json:"subject,omitempty"       doc:"The subject of the certificate, as a distinguished name."                                           example:"CN=osapi-controller"`
	Names     []string  `                                         json:"names,omitempty"         doc:"The DNS names and the IP addresses that the certificate is for: its subject alternative names."                                   label:"SAN"`
	Issuer    string    `                                         json:"issuer,omitempty"        doc:"The issuer of the certificate. The subject again for a self-signed one."                            example:"CN=osapi-controller"`
	Source    string    `                                         json:"source"                  doc:"self-signed, which osapi made, or cert_file, the operator's."                                       example:"self-signed"`
	NotBefore time.Time `                                         json:"not_before,omitzero"     doc:"When the certificate starts to be valid."                                                                                         label:"Valid from"`
	Expires   time.Time `                                         json:"expires,omitzero"        doc:"When the certificate expires."                                                                                                    label:"Valid to"`
	DaysLeft  *int      `                                         json:"days_left,omitempty"     doc:"The days that are left until it expires."                                                                                         label:"Days left"`
	Expiry    Expiry    `                                         json:"expiry"                  doc:"ok, soon (30 days or less), expired, or unknown when the component could not read the certificate." example:"ok"                                        render:"status"`
	Key       string    `                                         json:"key,omitempty"           doc:"The algorithm and size of the public key, such as ECDSA P-256."                                     example:"ECDSA P-256"`
	FIPS      *bool     `                                         json:"fips_approved,omitempty" doc:"Whether FIPS 186-5 and SP 800-131A approve the key and the signature algorithms."                                                 label:"FIPS approved"`
}

// Process is a component of osapi that is not an agent, the controller or
// the NATS server, with the crypto that it reports in its registration.
type Process struct {
	Component Component
	Hostname  string
	Crypto    *cryptoreport.Report
}
