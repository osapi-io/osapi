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

package pki_test

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"net"
	"testing"
	"time"

	"github.com/avfs/avfs/vfs/memfs"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/pki"
)

type CertPublicTestSuite struct {
	suite.Suite
}

func TestCertPublicTestSuite(t *testing.T) {
	suite.Run(t, new(CertPublicTestSuite))
}

// certPEM is a self-signed certificate of key, signed with alg, valid for
// a day from now, for web-01 and 10.0.0.1.
func (s *CertPublicTestSuite) certPEM(
	key crypto.Signer,
	alg x509.SignatureAlgorithm,
) []byte {
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1), NotAfter: time.Now().Add(49 * time.Hour),
		SignatureAlgorithm: alg, DNSNames: []string{"web-01"},
		IPAddresses: []net.IP{net.ParseIP("10.0.0.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, key.Public(), key)
	s.Require().NoError(err)

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

func (s *CertPublicTestSuite) TestReadCert() {
	p256, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	p224, _ := ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	rsa1024, _ := rsa.GenerateKey(rand.Reader, 1024)
	rsa2048, _ := rsa.GenerateKey(rand.Reader, 2048)
	_, ed, _ := ed25519.GenerateKey(rand.Reader)
	tests := []struct {
		name string
		key  crypto.Signer
		alg  x509.SignatureAlgorithm
		want string
		fips bool
	}{
		{"ECDSA P-256", p256, x509.ECDSAWithSHA256, "ECDSA P-256", true},
		{"ECDSA P-224", p224, x509.ECDSAWithSHA256, "ECDSA P-224", false},
		{"RSA 2048", rsa2048, x509.SHA256WithRSA, "RSA 2048", true},
		{"RSA of less than 2048 bits", rsa1024, x509.SHA256WithRSA, "RSA 1024", false},
		{"a signature with SHA-1", rsa2048, x509.SHA1WithRSA, "RSA 2048", false},
		{"Ed25519", ed, x509.PureEd25519, "Ed25519", true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			fsys := memfs.New()
			s.Require().NoError(fsys.WriteFile("/api.crt", s.certPEM(tt.key, tt.alg), 0o644))
			got, err := pki.ReadCert(fsys, pki.CertSource{
				Name: "api", Component: "controller", CertFile: "/api.crt", SelfSigned: true,
			}, time.Now())
			s.Require().NoError(err)
			s.Equal(pki.CertOK, got.State)
			s.Equal(tt.want, got.Key)
			s.Equal(tt.fips, *got.FIPS)
			s.Equal("self-signed", got.Source)
			s.Equal([]string{"web-01", "10.0.0.1"}, got.Names)
			s.Equal(2, *got.DaysLeft)
			s.Contains(got.Fingerprint, "SHA256:")
		})
	}
}

func (s *CertPublicTestSuite) TestReadCerts() {
	p256, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	fsys := memfs.New()
	s.Require().NoError(fsys.WriteFile("/api.crt", s.certPEM(p256, x509.ECDSAWithSHA256), 0o644))
	s.Require().NoError(fsys.WriteFile("/bad.crt", []byte("not a certificate"), 0o644))
	s.Require().NoError(fsys.WriteFile("/junk.crt",
		pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("junk")}), 0o644))

	got := pki.ReadCerts(fsys, []pki.CertSource{
		{Name: "api", Component: "controller", CertFile: "/api.crt"},
		{Name: "nats", Component: "nats", CertFile: "/nats.crt", SelfSigned: true},
		{Name: "bad", Component: "agent", CertFile: "/bad.crt"},
		{Name: "junk", Component: "agent", CertFile: "/junk.crt"},
	}, time.Now())

	s.Require().Len(got, 4)
	s.Equal(pki.CertOK, got[0].State)
	s.Equal("cert_file", got[0].Source)
	s.Equal(pki.CertNotMade, got[1].State)
	s.Equal("self-signed", got[1].Source)
	s.Equal("unreadable: /bad.crt: no PEM certificate", got[2].State)
	s.Contains(got[3].State, "unreadable: /junk.crt: ")
}
