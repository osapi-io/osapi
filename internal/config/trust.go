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
package config

import (
	"net"
	"net/url"
	"path/filepath"
)

// TrustLocal is c where each client that connects to a server of osapi on
// this host, with TLS and no ca_file, trusts that server's self-signed
// certificate in its key directory: the CLI the API's api.crt, and the
// controller and the agent the NATS server's nats.crt. A server with its own
// cert_file, or one on another host, needs ca_file, or the pin of a join
// string.
func TrustLocal(
	c Config,
) Config {
	if u, err := url.Parse(c.Controller.Client.URL); err == nil &&
		u.Scheme == "https" && isLocal(u.Hostname()) &&
		c.Controller.Client.TLS.CAFile == "" && c.Controller.API.TLS.CertFile == "" {
		c.Controller.Client.TLS.CAFile, _ = SelfSignedFiles(c.Controller.PKI.KeyDir, "api")
	}
	// osapi's NATS server in operator mode with no resolver_file makes its
	// credentials at its first start in NATSAuthDir: a client on this host
	// with no auth of its own connects with its role's.
	auth := c.NATS.Server.Auth
	if auth.Type == "operator" && auth.ResolverFile == "" {
		dir := NATSAuthDir(c.NATS.Server.KeyDir)
		creds := map[*NATSConnection]string{
			&c.Controller.NATS: "controller.creds", &c.Agent.NATS: "agent.creds",
		}
		for n, file := range creds {
			if isLocal(n.Host) && (n.Auth.Type == "" || n.Auth.Type == "none") {
				n.Auth = NATSAuth{Type: "creds", CredsFile: filepath.Join(dir, file)}
			}
		}
		if j := &c.Controller.Join; j.SigningKeyFile == "" && j.Callout.IssuerKeyFile == "" {
			j.Scoped = true
			j.SigningKeyFile = filepath.Join(dir, "agent-signing.nk")
			j.JoinSigningKeyFile = filepath.Join(dir, "join-signing.nk")
		}
	}
	natsCA, _ := SelfSignedFiles(c.NATS.Server.KeyDir, "nats")
	for _, n := range []*NATSConnection{&c.Controller.NATS, &c.Agent.NATS} {
		if n.TLS.Enabled && n.TLS.CAFile == "" && isLocal(n.Host) &&
			c.NATS.Server.TLS.CertFile == "" {
			n.TLS.CAFile = natsCA
		}
	}

	return c
}

// NATSAuthDir is where osapi's NATS server keeps the credentials that it
// makes in operator mode with no resolver_file: nats-auth in its key
// directory.
func NATSAuthDir(
	keyDir string,
) string {
	return filepath.Join(keyDir, "nats-auth")
}

// isLocal is whether host is this host by its loopback name or address.
func isLocal(
	host string,
) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)

	return ip != nil && ip.IsLoopback()
}
