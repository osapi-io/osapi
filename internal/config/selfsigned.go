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
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

// selfSignedFor is how long a self-signed certificate is valid.
const selfSignedFor = 5 * 365 * 24 * time.Hour

// SelfSignedHostname and SelfSignedWrite are os.Hostname and os.WriteFile,
// for a test to replace.
var (
	SelfSignedHostname = os.Hostname
	SelfSignedWrite    = os.WriteFile
)

// SelfSignedFiles are the certificate and key files of the self-signed
// certificate name in dir, such as api.crt and api.key.
func SelfSignedFiles(
	dir string,
	name string,
) (string, string) {
	return filepath.Join(dir, name+".crt"), filepath.Join(dir, name+".key")
}

// WithSelfSigned is s with a certificate: its own cert_file, or, when TLS is
// on and it names none, the self-signed certificate name in dir, which it
// makes when it does not exist, for this host, localhost and the hosts of
// urls. A certificate that s names later replaces it with no other change.
func (s ServerTLS) WithSelfSigned(
	dir string,
	name string,
	urls ...string,
) (ServerTLS, error) {
	if !s.Enabled || s.CertFile != "" {
		return s, nil
	}
	certFile, keyFile := SelfSignedFiles(dir, name)
	if _, err := os.Stat(certFile); err != nil {
		if err := writeSelfSigned(certFile, keyFile, urls); err != nil {
			return s, err
		}
	}
	s.CertFile, s.KeyFile = certFile, keyFile

	return s, nil
}

// WithCurrentSelfSigned is WithSelfSigned, except that a self-signed
// certificate that does not name this host's hostname now is made again: a
// renamed host, whose metrics certificate named its old name, gets one for
// its new name at its next start (osapi-io/osapi#1493). Only a certificate
// that nothing pins may change this way: not NATS's, which each join string
// pins, nor the API's, which clients on other hosts trust.
func (s ServerTLS) WithCurrentSelfSigned(
	dir string,
	name string,
) (ServerTLS, error) {
	if !s.Enabled || s.CertFile != "" {
		return s, nil
	}
	certFile, keyFile := SelfSignedFiles(dir, name)
	if host, _ := SelfSignedHostname(); host != "" && !certNames(certFile, host) {
		_ = os.Remove(certFile)
		_ = os.Remove(keyFile)
	}

	return s.WithSelfSigned(dir, name)
}

// certNames reports whether the certificate in file names host, or false
// when there is no certificate to read.
func certNames(
	file string,
	host string,
) bool {
	data, err := os.ReadFile(file)
	if err != nil {
		return false
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return false
	}
	c, err := x509.ParseCertificate(block.Bytes)

	return err == nil && c.VerifyHostname(host) == nil
}

// writeSelfSigned makes an ECDSA P-256 certificate for this host, localhost
// and the hosts of urls, and writes it and its key.
func writeSelfSigned(
	certFile string,
	keyFile string,
	urls []string,
) error {
	// With crypto/rand, a P-256 key, a self-signed template and PKCS #8 of
	// an ECDSA key do not fail.
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	host, _ := SelfSignedHostname()
	tmpl := x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().UnixNano()),
		Subject:               pkix.Name{CommonName: host, Organization: []string{"osapi"}},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(selfSignedFor),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
	}
	names := []string{host}
	for _, u := range urls {
		if p, err := url.Parse(u); err == nil && p.Hostname() != "" {
			names = append(names, p.Hostname())
		}
	}
	for _, n := range names {
		if n == "" {
			continue
		}
		if ip := net.ParseIP(n); ip != nil {
			tmpl.IPAddresses = append(tmpl.IPAddresses, ip)
		} else {
			tmpl.DNSNames = append(tmpl.DNSNames, n)
		}
	}
	der, _ := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &key.PublicKey, key)
	keyDER, _ := x509.MarshalPKCS8PrivateKey(key)
	if err := os.MkdirAll(filepath.Dir(certFile), 0o700); err != nil {
		return fmt.Errorf("make the key directory: %w", err)
	}
	if err := SelfSignedWrite(
		keyFile,
		pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}),
		0o600,
	); err != nil {
		return fmt.Errorf("write the key of %s: %w", certFile, err)
	}
	if err := SelfSignedWrite(
		certFile,
		pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
		0o644,
	); err != nil {
		return fmt.Errorf("write %s: %w", certFile, err)
	}

	return nil
}
