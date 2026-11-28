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
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"
	"time"

	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/internal/cryptoreport"
	"github.com/osapi-io/osapi/internal/pki"
)

// ErrNoComponent is a component that the host does not have, or that has
// not registered.
var ErrNoComponent = errors.New("component not found")

// expirySoonDays is the days left at which a certificate's expiry is soon.
const expirySoonDays = 30

// now is the time that an expiry is judged at. Tests swap it.
var now = time.Now

// TLSSections is the TLS section of the controller, of the NATS server and
// of each agent, or of those that the input names. Each section's
// certificates are those of pki cert list, of its host and its component.
func TLSSections(
	ctx context.Context,
	d Deps,
	in TLSListInput,
) (TLSList, error) {
	agents, err := d.Agents.ListAgents(ctx)
	if err != nil {
		return TLSList{}, fmt.Errorf("list the agents: %w", err)
	}
	all := certsOf(d, agents)
	var sections []TLS
	if d.Processes != nil {
		for _, p := range d.Processes(ctx) {
			sections = append(sections, section(p.Component, p.Hostname, p.Crypto, all))
		}
	}
	for _, a := range agents {
		sections = append(sections, section(ComponentAgent, a.Hostname, a.Crypto, all))
	}
	sort.SliceStable(sections, func(i, j int) bool {
		a, b := sections[i], sections[j]
		if a.Component != b.Component {
			return order(a.Component) < order(b.Component)
		}

		return a.Hostname < b.Hostname
	})
	out := TLSList{Items: []TLS{}}
	for _, s := range sections {
		if (in.Hostname == "" || s.Hostname == in.Hostname) &&
			(in.Component == "" || s.Component == in.Component) {
			out.Items = append(out.Items, s)
		}
	}
	out.Total = len(out.Items)

	return out, nil
}

// order is where c goes in the list: the controller, NATS, then the agents.
func order(
	c Component,
) int {
	return slices.Index(c.Values(), c)
}

// Values are each component, in the order that list gives them.
func (Component) Values() []Component {
	return []Component{ComponentController, ComponentNATS, ComponentAgent}
}

// Valid reports whether c is a component.
func (c Component) Valid() bool {
	return slices.Contains(c.Values(), c)
}

// Values are each expiry, from the best to the worst.
func (Expiry) Values() []Expiry {
	return []Expiry{ExpiryOK, ExpirySoon, ExpiryExpired, ExpiryUnknown}
}

// Valid reports whether e is an expiry.
func (e Expiry) Valid() bool {
	return slices.Contains(e.Values(), e)
}

// section is the TLS section of component c of host, with its crypto r,
// nil when it reports none, and its certificates of all.
func section(
	c Component,
	host string,
	r *cryptoreport.Report,
	all []Cert,
) TLS {
	out := TLS{Component: c, Hostname: host}
	var mine []Cert
	for _, cert := range all {
		if cert.Hostname == host && cert.Component == string(c) {
			mine = append(mine, cert)
			out.Certificates = append(out.Certificates, tlsCert(cert))
		}
	}
	if r != nil {
		out.FIPS = r.Verdict()
		out.Reasons = r.Reasons()
		out.Listeners = r.Listeners
		out.Go = r.Go
		out.Module = r.Module
		out.Algorithms = r.Algorithms
	}
	days, readable := firstExpiry(mine)
	out.Expiry = expiry(days, readable)
	out.Summary = summary(out.FIPS, out.Expiry, days)

	return out
}

// tlsCert is c as a TLS section shows it, with its expiry now.
func tlsCert(
	c Cert,
) TLSCert {
	days, readable := firstExpiry([]Cert{c})

	return TLSCert{
		Name: c.Name, State: c.State, Subject: c.Subject, Names: c.Names, Issuer: c.Issuer,
		Source: c.Source, NotBefore: c.NotBefore, Expires: c.Expires, DaysLeft: c.DaysLeft,
		Expiry: expiry(days, readable), Key: c.Key, FIPS: c.FIPS,
	}
}

// TLSSection is the TLS section of one component of one host: the
// controller's host when the input names none.
func TLSSection(
	ctx context.Context,
	d Deps,
	in TLSGetInput,
) (TLS, error) {
	host := in.Hostname
	if host == "" {
		host = d.Hostname
	}
	all, err := TLSSections(ctx, d, TLSListInput{Hostname: host, Component: in.Component})
	if err != nil {
		return TLS{}, fmt.Errorf("list the TLS sections: %w", err)
	}
	if len(all.Items) == 0 {
		return TLS{}, endpoint.Errorf(ErrNoComponent,
			"host %q has no %s: osapi client pki tls list shows each component", host, in.Component)
	}

	return all.Items[0], nil
}

// firstExpiry is the days left of the certificate of certs that expires
// first, and whether certs has one that its component read.
func firstExpiry(
	certs []Cert,
) (int, bool) {
	days, readable := 0, false
	for _, c := range certs {
		if c.State != pki.CertOK || c.Expires.IsZero() {
			continue
		}
		d := int(c.Expires.Sub(now()).Hours() / 24)
		if c.Expires.Before(now()) {
			d = -1
		}
		if !readable || d < days {
			days, readable = d, true
		}
	}

	return days, readable
}

// expiry is the expiry of a certificate with days left, unknown when no
// certificate is readable.
func expiry(
	days int,
	readable bool,
) Expiry {
	switch {
	case !readable:
		return ExpiryUnknown
	case days < 0:
		return ExpiryExpired
	case days <= expirySoonDays:
		return ExpirySoon
	default:
		return ExpiryOK
	}
}

// summary is the section in one line: the FIPS 140-3 verdict, then the
// certificate that expires first.
//
//	FIPS off · cert 62d
//	FIPS on · approved · cert expired
//	FIPS off
func summary(
	fips string,
	e Expiry,
	days int,
) string {
	out := "FIPS " + fips
	if fips == "" {
		out = "FIPS unknown"
	}
	switch e {
	case ExpiryExpired:
		return out + " · cert expired"
	case ExpiryOK, ExpirySoon:
		return out + fmt.Sprintf(" · cert %dd", days)
	default:
		return out
	}
}
