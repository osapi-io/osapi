// Copyright (c) 2024 John Dewey

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
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
)

// errNoCertificate is the error of a ca_file with no PEM certificate.
var errNoCertificate = errors.New("no PEM certificate")

// errNoServerCertificate is the error of a metrics server with TLS on and
// no certificate.
var errNoServerCertificate = errors.New("tls.enabled needs cert_file and key_file")

// Config is the TLS config of a connection with c: the system's roots and
// ca_file, the client certificate, the server name, and InsecureSkipVerify.
func (c ClientTLS) Config() (*tls.Config, error) {
	out := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		ServerName:         c.ServerName,
		InsecureSkipVerify: c.InsecureSkipVerify, //nolint:gosec // the setting says so, and Warn logs it
	}
	if c.CAFile != "" {
		pem, err := os.ReadFile(c.CAFile)
		if err != nil {
			return nil, fmt.Errorf("read ca_file: %w", err)
		}
		// The system pool is never an error on the platforms osapi runs on.
		pool, _ := x509.SystemCertPool()
		if !pool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("%s: %w", c.CAFile, errNoCertificate)
		}
		out.RootCAs = pool
	}
	if c.CertFile != "" {
		cert, err := tls.LoadX509KeyPair(c.CertFile, c.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("load the client certificate: %w", err)
		}
		out.Certificates = []tls.Certificate{cert}
	}

	return out, nil
}

// Dial is the TLS config of a connection that has no scheme to decide, such
// as one to NATS: nil when c is not enabled, else Config.
func (c ClientTLS) Dial() (*tls.Config, error) {
	if !c.Enabled {
		return nil, nil
	}

	return c.Config()
}

// SelfClient is the TLS config with which a server of osapi connects to
// itself, such as the NATS server for its setup and heartbeat, from the
// config that it serves: it trusts its own certificate, and checks the first
// name in it, since it dials an address that the certificate may not name.
// Nil when the server serves no TLS.
func SelfClient(
	server *tls.Config,
) *tls.Config {
	if server == nil {
		return nil
	}
	// LoadX509KeyPair fills Leaf.
	leaf := server.Certificates[0].Leaf
	pool := x509.NewCertPool()
	pool.AddCert(leaf)
	names := append([]string{}, leaf.DNSNames...)
	for _, ip := range leaf.IPAddresses {
		names = append(names, ip.String())
	}
	name := "localhost"
	if len(names) > 0 {
		name = names[0]
	}

	return &tls.Config{RootCAs: pool, ServerName: name, MinVersion: tls.VersionTLS12}
}

// Warn logs a warning when c accepts any certificate, with the key of its
// setting, such as telemetry.tracing.tls, at each start that has it on.
func (c ClientTLS) Warn(
	log *slog.Logger,
	key string,
) {
	if c.InsecureSkipVerify {
		log.Warn("TLS accepts any certificate: anyone in the path can read and change the traffic",
			slog.String("setting", key+".insecure_skip_verify"))
	}
}

// Config is the TLS config that a server with s serves, a metrics server or
// the NATS server: its certificate, or nil when s is not enabled. Neither has
// a self-signed fallback, so TLS on with no certificate is an error.
func (s ServerTLS) Config() (*tls.Config, error) {
	if !s.Enabled {
		return nil, nil
	}
	if s.CertFile == "" {
		return nil, errNoServerCertificate
	}
	cert, err := tls.LoadX509KeyPair(s.CertFile, s.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("load the server certificate: %w", err)
	}

	return &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12}, nil
}

// Address is where m listens, with https:// when it serves TLS, as health
// reports it.
func (m MetricsServer) Address() string {
	scheme := "http"
	if m.TLS.Enabled {
		scheme = "https"
	}

	return scheme + "://" + net.JoinHostPort(m.Host, strconv.Itoa(m.Port))
}
