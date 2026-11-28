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

	"github.com/osapi-io/osapi/internal/controller/certrotation"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/internal/controller/keyrotation"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/pki"
)

// ErrNotFound is a certificate that no host has.
var ErrNotFound = errors.New("certificate not found")

// List is each certificate of the controller's host and of each agent, or
// of the host that the input names.
func List(
	ctx context.Context,
	d Deps,
	in ListInput,
) (CertList, error) {
	all, err := certs(ctx, d)
	if err != nil {
		return CertList{}, err
	}
	out := CertList{Items: []Cert{}}
	for _, c := range all {
		if in.Hostname == "" || c.Hostname == in.Hostname {
			out.Items = append(out.Items, c)
		}
	}
	out.Total = len(out.Items)

	return out, nil
}

// Get is the certificate with the name on the host, the controller's host
// when the input names none.
func Get(
	ctx context.Context,
	d Deps,
	in GetInput,
) (Cert, error) {
	host := in.Hostname
	if host == "" {
		host = d.Hostname
	}
	all, err := certs(ctx, d)
	if err != nil {
		return Cert{}, err
	}
	for _, c := range all {
		if c.Name == in.Name && c.Hostname == host {
			return c, nil
		}
	}

	return Cert{}, endpoint.Errorf(ErrNotFound,
		"host %q has no certificate %q: osapi client pki cert list shows each", host, in.Name)
}

// certs are the controller's certificates, then each agent's, in the order
// of the agents.
func certs(
	ctx context.Context,
	d Deps,
) ([]Cert, error) {
	agents, err := d.Agents.ListAgents(ctx)
	if err != nil {
		return nil, err
	}

	return certsOf(d, agents), nil
}

// certsOf are the controller's certificates, then those of each of agents.
func certsOf(
	d Deps,
	agents []job.AgentInfo,
) []Cert {
	var out []Cert
	if d.Local != nil {
		for _, c := range d.Local() {
			out = append(out, of(d.Hostname, c))
		}
	}
	for _, a := range agents {
		for _, c := range a.Certificates {
			out = append(out, of(a.Hostname, c))
		}
	}

	return out
}

// of is c, a certificate of host, in the response.
func of(
	host string,
	c pki.Cert,
) Cert {
	return Cert{
		Hostname: host, Name: c.Name, State: c.State, Component: c.Component,
		Source: c.Source, File: c.File, Names: c.Names, Subject: c.Subject,
		Issuer: c.Issuer, NotBefore: c.NotBefore, Expires: c.Expires,
		DaysLeft: c.DaysLeft, Fingerprint: c.Fingerprint, Key: c.Key,
		Signature: c.Signature, FIPS: c.FIPS,
	}
}

// RotateKey starts the rotation of the controller's key, or shows the one
// in progress.
func RotateKey(
	ctx context.Context,
	d Deps,
	_ RotateInput,
) (Rotation, error) {
	s, err := d.Rotation.Start(ctx)
	if err != nil {
		return Rotation{}, err
	}

	return rotationOf(s), nil
}

// KeyRotation is the rotation of the controller's key now.
func KeyRotation(
	ctx context.Context,
	d Deps,
	_ RotationInput,
) (Rotation, error) {
	s, err := d.Rotation.Status(ctx)
	if err != nil {
		return Rotation{}, err
	}

	return rotationOf(s), nil
}

// RotateNATSCert starts the rotation of the NATS server's certificate, or
// shows the one in progress.
func RotateNATSCert(
	ctx context.Context,
	d Deps,
	_ CertRotateInput,
) (CertRotation, error) {
	s, err := d.CertRotation.Start(ctx)
	if err != nil {
		return CertRotation{}, err
	}

	return certRotationOf(s), nil
}

// NATSCertRotation is the rotation of the NATS server's certificate now.
func NATSCertRotation(
	ctx context.Context,
	d Deps,
	_ CertRotationInput,
) (CertRotation, error) {
	s, err := d.CertRotation.Status(ctx)
	if err != nil {
		return CertRotation{}, err
	}

	return certRotationOf(s), nil
}

// certRotationOf is s in the API's shape.
func certRotationOf(
	s certrotation.Status,
) CertRotation {
	out := CertRotation{
		Current: s.Current, Next: s.Next, StartedAt: s.StartedAt, SwitchBy: s.SwitchBy,
	}
	for _, a := range s.Agents {
		out.Agents = append(out.Agents, CertRotationAgent{
			Hostname: a.Hostname, MachineID: a.MachineID, HasNext: a.HasNext,
		})
	}

	return out
}

// rotationOf is s in the API's shape.
func rotationOf(
	s keyrotation.Status,
) Rotation {
	out := Rotation{Current: s.Current, Next: s.Next, StartedAt: s.StartedAt, SwitchBy: s.SwitchBy}
	for _, a := range s.Agents {
		out.Agents = append(out.Agents, RotationAgent{
			Hostname: a.Hostname, MachineID: a.MachineID, HasNext: a.HasNext,
		})
	}

	return out
}
