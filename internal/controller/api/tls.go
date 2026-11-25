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

package api

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/osapi-io/osapi/internal/config"
)

// The files of the self-signed certificate in controller.pki.key_dir.
const (
	SelfSignedCert = "api.crt"
	SelfSignedKey  = "api.key"
)

// selfSignedFor is how long a self-signed certificate is valid.
const selfSignedFor = 5 * 365 * 24 * time.Hour

// hostnameFn and writeFileFn are os.Hostname and os.WriteFile, for a test to
// replace.
var (
	hostnameFn  = os.Hostname
	writeFileFn = os.WriteFile
)

// TLSConfig is the server's TLS config: the certificate of c, or the
// self-signed certificate in keyDir, which it makes when it does not exist.
// A certificate that c names later replaces the self-signed one with no
// other change.
func TLSConfig(
	c config.ServerTLS,
	keyDir string,
	urls ...string,
) (*tls.Config, error) {
	certFile, keyFile := c.CertFile, c.KeyFile
	if certFile == "" {
		certFile = filepath.Join(keyDir, SelfSignedCert)
		keyFile = filepath.Join(keyDir, SelfSignedKey)
		if _, err := os.Stat(certFile); err != nil {
			if err := writeSelfSigned(certFile, keyFile, urls); err != nil {
				return nil, err
			}
		}
	}
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("load the API's certificate: %w", err)
	}

	return &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12}, nil
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
	host, _ := hostnameFn()
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
	if err := writeFileFn(
		keyFile,
		pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}),
		0o600,
	); err != nil {
		return fmt.Errorf("write the API's key: %w", err)
	}
	if err := writeFileFn(
		certFile,
		pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
		0o644,
	); err != nil {
		return fmt.Errorf("write the API's certificate: %w", err)
	}

	return nil
}

// CertFingerprint is the SHA-256 of the certificate that the API serves, as
// SHA256:<hex>, which a join token pins: the certificate of c, or the
// self-signed one in keyDir.
func CertFingerprint(
	c config.ServerTLS,
	keyDir string,
) (string, error) {
	certFile := c.CertFile
	if certFile == "" {
		certFile = filepath.Join(keyDir, SelfSignedCert)
	}
	data, err := os.ReadFile(certFile)
	if err != nil {
		return "", fmt.Errorf("read the API's certificate: %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return "", fmt.Errorf("%s: no PEM certificate", certFile)
	}
	sum := sha256.Sum256(block.Bytes)

	return "SHA256:" + hex.EncodeToString(sum[:]), nil
}
