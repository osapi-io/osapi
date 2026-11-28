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
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strconv"

	"github.com/avfs/avfs"

	"github.com/osapi-io/osapi/internal/tlspolicy"
)

// errNoCertificate is the error of a ca_file with no PEM certificate.
var errNoCertificate = errors.New("no PEM certificate")

// errNoServerCertificate is the error of a metrics server with TLS on and
// no certificate.
var errNoServerCertificate = errors.New("tls.enabled needs cert_file and key_file")

// Config is the TLS config of a connection with c: the system's roots and
// ca_file, the client certificate, the server name, and InsecureSkipVerify.
func (c ClientTLS) Config() (*tls.Config, error) {
	out := tlspolicy.Client()
	out.ServerName = c.ServerName
	out.InsecureSkipVerify = c.InsecureSkipVerify //nolint:gosec // the setting says so, and Warn logs it
	if c.CAFile != "" {
		pem, err := avfs.ReadFile(hostFS, c.CAFile)
		if err != nil {
			return nil, fmt.Errorf("read ca_file: %w", err)
		}
		pool := rootPool()
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

// Dial is the TLS config of a connection to host that has no scheme to
// decide, such as one to NATS: nil when c is not enabled, else Config. With a
// ca_file, it reads the file again at each handshake, and checks the server's
// chain against it and host or server_name: a connection that reconnects
// trusts a certificate that the file gained since it connected, such as the
// next one of a rotation of the NATS server's certificate.
func (c ClientTLS) Dial(
	host string,
) (*tls.Config, error) {
	if !c.Enabled {
		return nil, nil
	}
	out, err := c.Config()
	if err != nil || c.CAFile == "" || c.InsecureSkipVerify {
		return out, err
	}
	name := c.ServerName
	if name == "" {
		name = host
	}
	file := c.CAFile
	// The check is VerifyConnection's, against the file as it is now, in
	// place of the one of crypto/tls, against the file as it was.
	out.InsecureSkipVerify = true
	out.VerifyConnection = func(cs tls.ConnectionState) error {
		return verifyChain(cs, file, name)
	}

	return out, nil
}

// verifyChain checks the server's chain in cs against the system's roots and
// the certificates in file, and that its leaf is name's.
func verifyChain(
	cs tls.ConnectionState,
	file string,
	name string,
) error {
	if len(cs.PeerCertificates) == 0 {
		return errors.New("the server showed no certificate")
	}
	pem, err := avfs.ReadFile(hostFS, file)
	if err != nil {
		return fmt.Errorf("read ca_file: %w", err)
	}
	pool := rootPool()
	if !pool.AppendCertsFromPEM(pem) {
		return fmt.Errorf("%s: %w", file, errNoCertificate)
	}
	intermediates := x509.NewCertPool()
	for _, c := range cs.PeerCertificates[1:] {
		intermediates.AddCert(c)
	}
	_, err = cs.PeerCertificates[0].Verify(x509.VerifyOptions{
		DNSName: name, Roots: pool, Intermediates: intermediates,
	})

	return err
}

// SelfClient is the TLS config with which a server of osapi connects to
// itself, such as the NATS server for its setup and heartbeat, from the
// config that it serves: it trusts the certificate that the server serves
// at the time of each handshake, so a certificate that the server loads
// again changes nothing for it. Nil when the server serves no TLS.
func SelfClient(
	server *tls.Config,
) *tls.Config {
	if server == nil {
		return nil
	}

	out := tlspolicy.Client()
	// The check is VerifyConnection's: the server's own certificate.
	out.InsecureSkipVerify = true
	out.VerifyConnection = func(cs tls.ConnectionState) error {
		own, err := server.GetCertificate(&tls.ClientHelloInfo{})
		if err != nil {
			return err
		}
		if len(cs.PeerCertificates) == 0 ||
			!bytes.Equal(cs.PeerCertificates[0].Raw, own.Certificate[0]) {
			return errors.New("the server showed a certificate that is not its own")
		}

		return nil
	}

	return out
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
// a self-signed fallback, so TLS on with no certificate is an error. The
// server reads the certificate again when its files change, at the next
// handshake: a certificate that a rotation or the operator replaces takes
// effect with no restart.
func (s ServerTLS) Config() (*tls.Config, error) {
	if !s.Enabled {
		return nil, nil
	}
	if s.CertFile == "" {
		return nil, errNoServerCertificate
	}
	r, err := newCertReloader(s.CertFile, s.KeyFile)
	if err != nil {
		return nil, err
	}

	out := s.Negotiation()
	out.GetCertificate = r.certificate

	return out, nil
}

// Negotiation is what a server with s negotiates, with no certificate:
// tlspolicy's TLS 1.2 or later, and only the suites and groups that FIPS
// 140-3 approves. Nil when s is not enabled. Config serves it, and the
// crypto report of the process reads it.
func (s ServerTLS) Negotiation() *tls.Config {
	if !s.Enabled {
		return nil
	}

	return tlspolicy.Server()
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

// systemCertPoolFn reads the system's roots. A test replaces it.
var systemCertPoolFn = x509.SystemCertPool

// rootPool returns the system's roots, or an empty pool when the system has
// none it can read, so a ca_file still makes a pool on such a host.
func rootPool() *x509.CertPool {
	pool, err := systemCertPoolFn()
	if err != nil || pool == nil {
		return x509.NewCertPool()
	}

	return pool
}
