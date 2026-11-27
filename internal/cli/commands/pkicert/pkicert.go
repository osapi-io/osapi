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

// Package pkicert is osapi pki cert: it shows and rotates the TLS
// certificates that the servers of osapi serve on this host, the self-signed
// ones that each makes at its first start and a cert_file of the operator.
package pkicert

import (
	"strings"

	"github.com/osapi-io/osapi/internal/cli/command"
)

// Group is osapi pki cert.
var Group = command.Register(command.Define(
	"cert", "Show and rotate the TLS certificates of osapi on this host",
	command.Under(command.ParentPKI),
	command.Uses[Deps](),
	command.Method[ListInput, ListOutput]("list").
		Short("List the certificates, with their state, algorithms and expiry").
		Long("Run this on the host. It reads each certificate that the API, the NATS\n"+
			"server and the metrics servers serve: the self-signed one in the key\n"+
			"directory of the component, or the cert_file of the config. Each of the\n"+
			"five is in the list. A component that has not started on this host made\n"+
			"no certificate: its state is not made. Run it on the host of the server;\n"+
			"a client config reads no certificate.").
		Example("").
		Prints(SampleList).
		Handle(List),
	command.Method[GetInput, CertInfo]("get").
		Short("Show one certificate in full: its names, algorithms and fingerprint").
		Example("--name api").
		Prints(SampleCert).
		Fails("no name", GetInput{}).
		Fails("a name of more than 64 characters", GetInput{Name: strings.Repeat("a", 65)}).
		Handle(Get),
	command.Method[RotateInput, RotateOutput]("rotate").
		Short("Make a new self-signed certificate in place of the old one").
		Long("It replaces a self-signed certificate, and the running server serves the\n"+
			"new one at once. It refuses a cert_file of the operator: replace that file.\n"+
			"Each join string pins the NATS server's certificate, and each joined agent\n"+
			"trusts it: a new one cuts those agents off, so it needs --force, then new\n"+
			"join tokens and a join again of each agent.").
		Example("--name api").
		Check("pki cert get", "--name api", "state", "ok", "source", "self-signed").
		Prints(SampleRotate).
		Fails("no name", RotateInput{}).
		Fails("a name of more than 64 characters", RotateInput{Name: strings.Repeat("a", 65)}).
		Handle(Rotate),
))
