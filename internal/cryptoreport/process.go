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

package cryptoreport

import (
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/pki"
)

// natsVault is the type of the vault whose values the controller encrypts.
const natsVault = "nats"

// Controller is the report of the controller that cfg starts, with the
// certificates that it serves.
func Controller(
	cfg config.Config,
	certs []pki.Cert,
) Report {
	c := cfg.Controller
	listeners := append(
		[]TLS{{Name: "api", Config: c.API.TLS.Negotiation()}},
		metrics(c.Metrics)...)
	uses := []Algorithm{APITokens, Signatures, Seal}
	for _, v := range c.Secrets.Vaults {
		if v.Type == natsVault {
			uses = append(uses, Vault)
			break
		}
	}
	uses = append(uses, NATSAuth)

	return New(listeners, append(uses, Certificates(certs)...))
}

// NATS is the report of the embedded NATS server of sc, with the
// certificates that it serves. Its store is a use when a key encrypts it.
func NATS(
	sc config.NATSServer,
	certs []pki.Cert,
) Report {
	listeners := append([]TLS{{Name: "nats", Config: sc.TLS.Negotiation()}}, metrics(sc.Metrics)...)
	uses := []Algorithm{NATSAuth}
	if sc.JetStream.KeyFile != "" {
		uses = append(uses, Store)
	}

	return New(listeners, append(uses, Certificates(certs)...))
}

// Agent is the report of the agent that cfg starts, with the certificates
// that it serves.
func Agent(
	cfg config.Config,
	certs []pki.Cert,
) Report {
	uses := append([]Algorithm{Signatures, Seal, NATSAuth}, Certificates(certs)...)

	return New(metrics(cfg.Agent.Metrics), uses)
}

// metrics is the listener of m, when its server runs.
func metrics(
	m config.MetricsServer,
) []TLS {
	if !m.Enabled {
		return nil
	}

	return []TLS{{Name: "metrics", Config: m.TLS.Negotiation()}}
}
