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

	"github.com/osapi-io/osapi/internal/controller/certrotation"
	"github.com/osapi-io/osapi/internal/controller/keyrotation"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/pki"
)

// Deps is what the group's handlers use: the agents, whose registrations
// have their certificates, and the controller's own certificates.
type Deps struct {
	Agents Agents
	// Local are the certificates of the controller's host, the controller's
	// and the NATS server's, as the controller reads them now. Nil is none.
	Local func() []pki.Cert
	// Hostname is the hostname of the controller's host.
	Hostname string
	// Processes are the controller and the NATS server, each with the
	// crypto of its registration, for the TLS sections. Nil is none.
	Processes func(ctx context.Context) []Process
	// Rotation starts and shows the rotation of the controller's key.
	Rotation Rotator
	// CertRotation starts and shows the rotation of the NATS server's
	// certificate.
	CertRotation CertRotator
}

// CertRotator starts and shows the rotation of the NATS server's
// certificate. A *certrotation.Rotation is one.
type CertRotator interface {
	Start(ctx context.Context) (certrotation.Status, error)
	Status(ctx context.Context) (certrotation.Status, error)
}

// Rotator starts and shows the rotation of the controller's key. A
// *keyrotation.Rotation is one.
type Rotator interface {
	Start(ctx context.Context) (keyrotation.Status, error)
	Status(ctx context.Context) (keyrotation.Status, error)
}

// Agents lists the agents, each with the certificates of its registration.
type Agents interface {
	ListAgents(ctx context.Context) ([]job.AgentInfo, error)
}
