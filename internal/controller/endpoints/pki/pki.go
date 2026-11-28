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

// Package pki is the pki endpoint group: the certificates of the fleet,
// the controller's and each agent's, through the API. osapi pki cert reads
// the certificates of the host that it runs on.
package pki

import (
	"strings"

	"github.com/osapi-io/osapi/internal/area"
	"github.com/osapi-io/osapi/internal/controller/certrotation"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
)

// Group is the pki group's definition. The routes, the spec, the CLI
// commands and their pages come from it.
var Group = endpoint.Register(endpoint.Define(
	"pki",
	"The certificates of the fleet, and the rotation of the controller's key and of the NATS server's certificate",
	endpoint.Area(area.Fleet),
	endpoint.SubDoc("cert", "Read the certificates of the controller and of each agent"),
	endpoint.SubDoc("key", "Rotate the controller's key with no new admission of the agents"),
	endpoint.SubDoc("tls",
		"Show the TLS of each component: FIPS, certificates, listeners and crypto"),
	endpoint.SubDoc("nats",
		"Rotate the NATS server's certificate with no restart and no new join of the agents"),
	endpoint.Title("PKI API"),
	endpoint.Tag(
		"pki_operations",
		"PKI",
		"The certificates of the fleet, and the rotation of the controller's key and of "+
			"the NATS server's certificate.",
	),
	endpoint.Uses[Deps](),
	endpoint.Get[ListInput, CertList]("list").
		Handle(List).
		Sub("cert").
		Use("list").
		ID("ListPKICerts").
		Describe("List the certificates of the fleet").
		Explain("Lists the certificates of the controller's host, the API's, the "+
			"controller's metrics and the NATS server's, and the certificates that each "+
			"agent reports in its registration, with their expiry. A certificate that a "+
			"host has not made is not made.").
		Returns("The certificates of each host.").
		Command("List the certificates of the fleet",
			"Lists the certificates of the controller and of each agent, with the\n"+
				"days that are left until each expires. osapi pki cert list reads only\n"+
				"the host that it runs on.").
		Example("").
		Example("--hostname web-01").
		Prints(sampleList).
		Fails("a hostname over 255 characters", ListInput{Hostname: strings.Repeat("a", 256)}),
	endpoint.Get[TLSListInput, TLSList]("tls").
		Handle(TLSSections).
		Sub("tls").
		Use("list").
		ID("ListPKITLS").
		Describe("List the TLS of each component").
		Explain("Lists the TLS section of the controller, of the NATS server and of "+
			"each agent: the FIPS 140-3 mode, with each reason that the component is "+
			"not approved as a sentence; each certificate that it serves, from pki cert "+
			"list, with its subject, its issuer, its validity and the days that are "+
			"left; each TLS listener, with its lowest version, its suites and its "+
			"groups, and whether FIPS 140-3 approves them; and the Go version, the "+
			"module and the algorithms of its crypto. A certificate within 30 days of "+
			"its end is soon, and one past its end is expired.").
		Returns("The TLS section of each component.").
		Command("List the TLS of each component",
			"Shows for the controller, the NATS server and each agent its FIPS 140-3\n"+
				"mode and the reasons it is not approved, its certificates with the days\n"+
				"left, its TLS listeners with their suites and groups, and its crypto.\n"+
				"The summary is the line that the tables of health status and agent\n"+
				"list link to.").
		Example("--component nats").
		Example("--hostname web-01").
		Example("").
		Prints(sampleTLS).
		Fails("a hostname over 255 characters", TLSListInput{Hostname: strings.Repeat("a", 256)}).
		Fails("a component that is not one", TLSListInput{Component: "ui"}),
	endpoint.Get[TLSGetInput, TLS]("tlsget").
		Handle(TLSSection).
		Sub("tls").
		Use("get").
		Key("component").
		ID("GetPKITLS").
		Describe("Get the TLS of one component").
		Explain("Reads the TLS section of one component of one host in full: the "+
			"controller or the NATS server of the controller's host when the hostname "+
			"is empty, or the agent of the hostname. It has the reasons, each "+
			"certificate, each listener and each algorithm.").
		Returns("The TLS section of the component.").
		Command("Get the TLS of one component",
			"Shows one component's TLS in full: its FIPS 140-3 mode and why it is\n"+
				"not approved, each certificate with its names, issuer and validity,\n"+
				"each TLS listener with its suites and groups, and its algorithms.").
		Example("--component agent --hostname web-01").
		Example("--component controller").
		Prints(sampleTLS.Items[0]).
		Errors(endpoint.NotFound(ErrNoComponent, "The host has no such component.")).
		Fails("a component that is not one", TLSGetInput{Component: "ui"}).
		Fails("a hostname over 255 characters",
			TLSGetInput{Component: ComponentAgent, Hostname: strings.Repeat("a", 256)}),
	endpoint.Get[GetInput, Cert]("get").
		Handle(Get).
		Sub("cert").
		Use("get").
		Key("name").
		ID("GetPKICert").
		Describe("Get one certificate of a host").
		Explain("Reads one certificate of one host: the controller's host when the "+
			"hostname is empty, or the agent of the hostname.").
		Returns("The certificate.").
		Command("Get one certificate of a host",
			"Shows one certificate in full: its names, its fingerprint and its\n"+
				"algorithms. With no hostname, it is a certificate of the controller's\n"+
				"host, such as api.").
		Example("--name api").
		Example("--name agent-metrics --hostname web-01").
		Prints(sampleCert).
		Errors(endpoint.NotFound(ErrNotFound, "The host has no certificate with that name.")).
		Fails("a name over 64 characters", GetInput{Name: strings.Repeat("a", 65)}).
		Fails("a hostname over 255 characters",
			GetInput{Name: "api", Hostname: strings.Repeat("a", 256)}),
	endpoint.Get[RotationInput, Rotation]("rotation").
		Handle(KeyRotation).
		Sub("key").
		Use("get").
		ID("GetPKIKeyRotation").
		Describe("Show the rotation of the controller's key").
		Explain("Shows the key that the controller signs jobs with, and during a "+
			"rotation the key that it switches to, when it switches at the latest, and "+
			"each admitted agent that the switch waits for, with whether it has the "+
			"next key.").
		Returns("The rotation of the controller's key.").
		Command("Show the rotation of the controller's key",
			"Shows the controller's key, and during a rotation, each agent that the\n"+
				"switch waits for. The controller switches when each one has the next key,\n"+
				"or when controller.pki.rotation_grace_period ends.").
		Example("").
		Prints(sampleRotation),
	endpoint.Post[RotateInput, Rotation]("rotate").
		AdminOnly().
		Handle(RotateKey).
		Sub("key").
		Use("rotate").
		ID("RotatePKIKey").
		Describe("Rotate the controller's key").
		Explain("Starts the rotation of the controller's key, with no new admission "+
			"of the agents. The controller makes the next key, and signs a record "+
			"that gives it to the agents with its current key. Each agent takes it at "+
			"its next heartbeat. The controller switches to the next key when each "+
			"admitted agent that reported within controller.pki.rotation_grace_period "+
			"has it, or when the grace period ends. An agent that does not have it "+
			"then needs a new admission. A rotation in progress is not started again.").
		Returns("The rotation that started, or the one in progress.").
		Command("Rotate the controller's key",
			"Starts the rotation of the controller's key. The agents take the next key\n"+
				"at their next heartbeat, and the controller switches when each one has it.\n"+
				"osapi client pki key get shows the progress. An agent that is away\n"+
				"for the whole grace period needs a new admission.").
		Example("").
		Check("client pki key get", "", "next", "{out.next}").
		Prints(sampleRotationStarted),
	endpoint.Get[CertRotationInput, CertRotation]("natscertrotation").
		Handle(NATSCertRotation).
		Sub("nats").
		Use("get").
		ID("GetPKINATSCertRotation").
		Describe("Show the rotation of the NATS server's certificate").
		Explain("Shows the certificate that the NATS server serves, and during a "+
			"rotation the certificate that it switches to, when it switches at the "+
			"latest, and each admitted agent that the switch waits for, with whether "+
			"it trusts the next certificate.").
		Returns("The rotation of the NATS server's certificate.").
		Command("Show the rotation of the NATS server's certificate",
			"Shows the NATS server's certificate, and during a rotation, each agent that\n"+
				"the switch waits for. The controller switches when each one trusts the next\n"+
				"certificate, or when controller.pki.rotation_grace_period ends.").
		Example("").
		Prints(sampleCertRotation),
	endpoint.Post[CertRotateInput, CertRotation]("natscertrotate").
		AdminOnly().
		Handle(RotateNATSCert).
		Sub("nats").
		Use("rotate").
		ID("RotatePKINATSCert").
		Describe("Rotate the NATS server's certificate").
		Explain("Starts the rotation of the NATS server's self-signed certificate, "+
			"which each joined agent pins, with no restart and no new join. The "+
			"controller makes the next certificate beside the current one, and signs a "+
			"record that gives both to the agents. Each agent trusts the next one at its "+
			"next heartbeat. The controller renames it into place when each admitted "+
			"agent that reported within controller.pki.rotation_grace_period trusts it, "+
			"or when the grace period ends, and NATS serves it at its next handshake. A "+
			"rotation in progress is not started again. It runs on the controller's host, "+
			"which must hold the NATS server's certificate, and refuses a certificate of "+
			"the operator's cert_file.").
		Returns("The rotation that started, or the one in progress.").
		Command("Rotate the NATS server's certificate",
			"Starts the rotation of the NATS server's certificate. The agents trust the\n"+
				"next certificate at their next heartbeat, and the controller switches when\n"+
				"each one does. osapi client pki nats get shows the progress.").
		Example("").
		Check("client pki nats get", "", "next", "{out.next}").
		Prints(sampleCertRotationStarted).
		Errors(endpoint.Conflict(
			certrotation.ErrOperatorCert,
			"The NATS server serves a certificate of the operator's cert_file: replace that file.",
		)),
))
