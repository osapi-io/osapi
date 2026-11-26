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
package pkicert_test

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/fips140"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/avfs/avfs/vfs/osfs"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/audit"
	"github.com/osapi-io/osapi/internal/cli/commands/pkicert"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/pki"
)

type MethodsPublicTestSuite struct {
	suite.Suite
	dir     string
	entries []audit.Entry
}

func TestMethodsPublicTestSuite(t *testing.T) {
	suite.Run(t, new(MethodsPublicTestSuite))
}

func (s *MethodsPublicTestSuite) SetupSubTest() {
	s.dir = s.T().TempDir()
	s.entries = nil
	for _, name := range []string{"api", "nats"} {
		_, err := config.ServerTLS{Enabled: true}.WithSelfSigned(s.dir, name)
		s.Require().NoError(err)
	}
}

func (s *MethodsPublicTestSuite) deps() pkicert.Deps {
	src := func(name string, self bool) pki.CertSource {
		cert, key := config.SelfSignedFiles(s.dir, name)
		return pki.CertSource{Name: name, CertFile: cert, KeyFile: key, SelfSigned: self}
	}

	return pkicert.Deps{
		FS: osfs.NewWithNoIdm(),
		Certs: func() []pki.CertSource {
			return []pki.CertSource{
				src("api", true), src("nats", true), src("mine", false), src("missing", true),
			}
		},
		Now: func() time.Time { return time.Now() },
		Audit: func(_ context.Context, e audit.Entry) error {
			s.entries = append(s.entries, e)
			return nil
		},
	}
}

func (s *MethodsPublicTestSuite) TestList() {
	s.Run("each certificate that can be read, with its days left", func() {
		got, err := pkicert.List(context.Background(), s.deps(), pkicert.ListInput{})
		s.Require().NoError(err)
		s.Len(got.Items, 4, "each certificate is in the list")
		s.Equal("not made", got.Items[3].State, "a certificate that is not there")
		s.Equal("ok", got.Items[0].State)
		s.Equal("api", got.Items[0].Name)
		s.Equal("self-signed", got.Items[0].Source)
		s.Greater(*got.Items[0].DaysLeft, 1800)
		s.Nil(got.Items[3].DaysLeft, "no days left without a certificate")
		s.Nil(got.Items[3].FIPS)
		s.Contains(got.Items[0].Names, "localhost")
	})
}

func (s *MethodsPublicTestSuite) TestGet() {
	s.Run("one certificate", func() {
		got, err := pkicert.Get(context.Background(), s.deps(), pkicert.GetInput{Name: "nats"})
		s.Require().NoError(err)
		s.Contains(got.Fingerprint, "SHA256:")
	})
	s.Run("a certificate that this host does not have", func() {
		_, err := pkicert.Get(context.Background(), s.deps(), pkicert.GetInput{Name: "missing"})
		s.ErrorIs(err, pkicert.ErrNotMade)
		s.ErrorContains(err, "missing is not on this host")
	})
	s.Run("a name that is not a certificate", func() {
		_, err := pkicert.Get(context.Background(), s.deps(), pkicert.GetInput{Name: "none"})
		s.ErrorIs(err, pkicert.ErrNoCert)
	})
	s.Run("a list with a certificate that cannot be read", func() {
		s.Require().NoError(os.WriteFile(filepath.Join(s.dir, "nats.crt"), []byte("nope"), 0o644))
		got, err := pkicert.List(context.Background(), s.deps(), pkicert.ListInput{})
		s.Require().NoError(err)
		s.Contains(got.Items[1].State, "unreadable: ")
	})
	s.Run("a file that is not a certificate", func() {
		s.Require().NoError(os.WriteFile(filepath.Join(s.dir, "nats.crt"), []byte("nope"), 0o644))
		_, err := pkicert.Get(context.Background(), s.deps(), pkicert.GetInput{Name: "nats"})
		s.ErrorContains(err, "no PEM certificate")
	})
	s.Run("PEM that is not a certificate", func() {
		s.Require().NoError(os.WriteFile(filepath.Join(s.dir, "nats.crt"),
			[]byte("-----BEGIN CERTIFICATE-----\nAAAA\n-----END CERTIFICATE-----\n"), 0o644))
		_, err := pkicert.Get(context.Background(), s.deps(), pkicert.GetInput{Name: "nats"})
		s.Error(err)
	})
}

// writeCert writes a self-signed certificate of key, signed with alg, as
// name.crt in the test's directory.
func (s *MethodsPublicTestSuite) writeCert(
	name string,
	key crypto.Signer,
	alg x509.SignatureAlgorithm,
) {
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1), NotAfter: time.Now().Add(24 * time.Hour),
		SignatureAlgorithm: alg,
	}
	// A fixture of a weak certificate, such as one signed with SHA-1, is
	// made with no FIPS 140-3 enforcement, so the test of its report runs
	// in only mode too.
	var der []byte
	var err error
	fips140.WithoutEnforcement(func() {
		der, err = x509.CreateCertificate(rand.Reader, tmpl, tmpl, key.Public(), key)
	})
	s.Require().NoError(err)
	cert, _ := config.SelfSignedFiles(s.dir, name)
	s.Require().
		NoError(os.WriteFile(cert, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o644))
}

func (s *MethodsPublicTestSuite) TestAlgorithms() {
	var rsa1024 *rsa.PrivateKey
	fips140.WithoutEnforcement(func() { rsa1024, _ = rsa.GenerateKey(rand.Reader, 1024) })
	rsa2048, _ := rsa.GenerateKey(rand.Reader, 2048)
	p224, _ := ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	p384, _ := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	_, ed, _ := ed25519.GenerateKey(rand.Reader)
	tests := []struct {
		name string
		key  crypto.Signer
		alg  x509.SignatureAlgorithm
		want string
		fips bool
	}{
		{"self-signed default", nil, 0, "ECDSA P-256", true},
		{"RSA 2048 with SHA-256", rsa2048, x509.SHA256WithRSA, "RSA 2048", true},
		{"RSA of less than 2048 bits", rsa1024, x509.SHA256WithRSA, "RSA 1024", false},
		{"a signature with SHA-1", rsa2048, x509.SHA1WithRSA, "RSA 2048", false},
		{"ECDSA P-384", p384, x509.ECDSAWithSHA384, "ECDSA P-384", true},
		{"ECDSA P-224", p224, x509.ECDSAWithSHA256, "ECDSA P-224", false},
		{"Ed25519", ed, x509.PureEd25519, "Ed25519", true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.key != nil {
				s.writeCert("api", tt.key, tt.alg)
			}
			got, err := pkicert.Get(context.Background(), s.deps(), pkicert.GetInput{Name: "api"})
			s.Require().NoError(err)
			s.Equal(tt.want, got.Key)
			s.Equal(tt.fips, *got.FIPS)
			s.NotEmpty(got.Signature)
		})
	}
}

func (s *MethodsPublicTestSuite) TestRotate() {
	s.Run("a self-signed certificate is made again, and audited", func() {
		before, _ := pkicert.Get(context.Background(), s.deps(), pkicert.GetInput{Name: "api"})
		got, err := pkicert.Rotate(
			context.Background(),
			s.deps(),
			pkicert.RotateInput{Name: "api", Force: false},
		)
		s.Require().NoError(err)
		s.NotEqual(before.Fingerprint, got.Cert.Fingerprint)
		s.Equal("restart the server to serve it", got.Next)
		s.Require().Len(s.entries, 1)
		s.Equal("cert.rotate", s.entries[0].Action)
	})
	s.Run("the NATS certificate needs --force, and says to join again", func() {
		_, err := pkicert.Rotate(
			context.Background(),
			s.deps(),
			pkicert.RotateInput{Name: "nats", Force: false},
		)
		s.ErrorIs(err, pkicert.ErrPinned)
		got, err := pkicert.Rotate(
			context.Background(),
			s.deps(),
			pkicert.RotateInput{Name: "nats", Force: true},
		)
		s.Require().NoError(err)
		s.Contains(got.Next, "join each agent again")
	})
	s.Run("a cert_file of the operator is refused", func() {
		_, err := pkicert.Rotate(
			context.Background(),
			s.deps(),
			pkicert.RotateInput{Name: "mine", Force: false},
		)
		s.ErrorIs(err, pkicert.ErrOperatorCert)
	})
	s.Run("a name that is not a certificate", func() {
		_, err := pkicert.Rotate(
			context.Background(),
			s.deps(),
			pkicert.RotateInput{Name: "none", Force: false},
		)
		s.ErrorIs(err, pkicert.ErrNoCert)
	})
	s.Run("a certificate whose directory cannot hold a new one", func() {
		d := s.deps()
		blocked := filepath.Join(s.dir, "file")
		s.Require().NoError(os.WriteFile(blocked, nil, 0o600))
		d.Certs = func() []pki.CertSource {
			return []pki.CertSource{{
				Name: "x", CertFile: filepath.Join(blocked, "x.crt"),
				KeyFile: filepath.Join(blocked, "x.key"), SelfSigned: true,
			}}
		}
		_, err := pkicert.Rotate(
			context.Background(),
			d,
			pkicert.RotateInput{Name: "x", Force: false},
		)
		s.Error(err)
	})
	s.Run("a certificate that cannot be made again", func() {
		defer func(w func(string, []byte, os.FileMode) error) { config.SelfSignedWrite = w }(
			config.SelfSignedWrite,
		)
		config.SelfSignedWrite = func(string, []byte, os.FileMode) error { return os.ErrPermission }
		_, err := pkicert.Rotate(
			context.Background(),
			s.deps(),
			pkicert.RotateInput{Name: "api", Force: false},
		)
		s.ErrorIs(err, os.ErrPermission)
	})
	s.Run("a new certificate that cannot be read", func() {
		defer func(w func(string, []byte, os.FileMode) error) { config.SelfSignedWrite = w }(
			config.SelfSignedWrite,
		)
		config.SelfSignedWrite = func(name string, _ []byte, perm os.FileMode) error {
			return os.WriteFile(name, []byte("nope"), perm)
		}
		_, err := pkicert.Rotate(
			context.Background(),
			s.deps(),
			pkicert.RotateInput{Name: "api", Force: false},
		)
		s.ErrorContains(err, "no PEM certificate")
	})
	s.Run("a directory that cannot be removed from", func() {
		s.Require().NoError(os.Chmod(s.dir, 0o500))
		defer func() { _ = os.Chmod(s.dir, 0o700) }()
		_, err := pkicert.Rotate(
			context.Background(),
			s.deps(),
			pkicert.RotateInput{Name: "api", Force: false},
		)
		s.ErrorContains(err, "remove")
	})
}
