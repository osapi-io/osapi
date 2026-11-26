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
	"github.com/osapi-io/osapi/internal/controller/endpoint"
)

// Group is the pki group's definition. The routes, the spec, the CLI
// commands and their pages come from it.
var Group = endpoint.Register(endpoint.Define(
	"pki", "The certificates of the fleet",
	endpoint.Area(area.Fleet),
	endpoint.SubDoc("cert", "Read the certificates of the controller and of each agent"),
	endpoint.Title("PKI API"),
	endpoint.Tag("pki_operations", "PKI", "The certificates of the fleet."),
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
))
