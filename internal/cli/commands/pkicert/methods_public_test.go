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
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/audit"
	"github.com/osapi-io/osapi/internal/cli/commands/pkicert"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller/certrotation"
	"github.com/osapi-io/osapi/internal/controller/certrotation/mocks"
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

// natsDeps are deps with the controller's key in its key directory, which
// signs the record of a rotation of the NATS certificate.
func (s *MethodsPublicTestSuite) natsDeps() pkicert.Deps {
	d := s.deps()
	keyDir := filepath.Join(s.dir, "controller")
	s.Require().NoError(pki.New(d.FS, keyDir, "controller").LoadOrGenerate())
	d.Config.Controller.PKI.KeyDir = keyDir
	d.Config.NATS.State.Bucket = "agent-state"
	d.Config.Controller.NATS.Namespace = "osapi"

	return d
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
			pkicert.RotateInput{Name: "api"},
		)
		s.Require().NoError(err)
		s.NotEqual(before.Fingerprint, got.Cert.Fingerprint)
		s.Contains(got.Next, "next handshake")
		s.Contains(got.Next, "join tokens made from now")
		s.Require().Len(s.entries, 1)
		s.Equal("cert.rotate", s.entries[0].Action)
	})
	s.Run("a metrics certificate says only when it is served", func() {
		d := s.deps()
		d.Certs = func() []pki.CertSource {
			cert, key := config.SelfSignedFiles(s.dir, "api")
			return []pki.CertSource{{
				Name: "controller-metrics", CertFile: cert, KeyFile: key, SelfSigned: true,
			}}
		}
		got, err := pkicert.Rotate(
			context.Background(),
			d,
			pkicert.RotateInput{Name: "controller-metrics"},
		)
		s.Require().NoError(err)
		s.Equal("the server serves it at its next handshake", got.Next)
	})
	s.Run("the NATS certificate starts a rotation that the agents follow", func() {
		d := s.natsDeps()
		before, _ := pkicert.Get(context.Background(), d, pkicert.GetInput{Name: "nats"})
		ctrl := gomock.NewController(s.T())
		kv := mocks.NewMockKV(ctrl)
		kv.EXPECT().KVGet(gomock.Any(), "osapi-agent-state", pki.NATSCertStateKey).
			Return(nil, jetstream.ErrKeyNotFound)
		kv.EXPECT().KVCreate(gomock.Any(), "osapi-agent-state", pki.NATSCertStateKey, gomock.Any()).
			Return(uint64(1), nil)
		closed := false
		d.Dial = func(context.Context) (certrotation.KV, func(), error) {
			return kv, func() { closed = true }, nil
		}
		got, err := pkicert.Rotate(context.Background(), d, pkicert.RotateInput{Name: "nats"})
		s.Require().NoError(err)
		s.True(closed)
		s.Equal(
			before.Fingerprint,
			got.Cert.Fingerprint,
			"the server serves the old one until the switch",
		)
		s.Contains(got.NextCert, "SHA256:")
		s.NotEqual(got.Cert.Fingerprint, got.NextCert)
		s.Contains(got.Next, "pki cert get")
		next, _ := certrotation.NextFiles(d.Certs()[1])
		s.FileExists(next)
		s.Require().Len(s.entries, 1)
		s.Contains(s.entries[0].RequestSummary, "next="+got.NextCert)
	})
	s.Run("the NATS certificate needs the controller's key", func() {
		d := s.natsDeps()
		d.Config.Controller.PKI.KeyDir = filepath.Join(s.dir, "none")
		_, err := pkicert.Rotate(context.Background(), d, pkicert.RotateInput{Name: "nats"})
		s.ErrorContains(err, "run this on the controller's host")
	})
	s.Run("the NATS certificate that is not on this host", func() {
		d := s.natsDeps()
		s.Require().NoError(os.Remove(filepath.Join(s.dir, "nats.crt")))
		_, err := pkicert.Rotate(context.Background(), d, pkicert.RotateInput{Name: "nats"})
		s.ErrorIs(err, os.ErrNotExist)
	})
	s.Run("the NATS certificate when NATS does not connect", func() {
		d := s.natsDeps()
		d.Dial = func(context.Context) (certrotation.KV, func(), error) {
			return nil, nil, os.ErrDeadlineExceeded
		}
		_, err := pkicert.Rotate(context.Background(), d, pkicert.RotateInput{Name: "nats"})
		s.ErrorIs(err, os.ErrDeadlineExceeded)
	})
	s.Run("the NATS certificate when the record does not write", func() {
		d := s.natsDeps()
		ctrl := gomock.NewController(s.T())
		kv := mocks.NewMockKV(ctrl)
		kv.EXPECT().KVGet(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, os.ErrPermission)
		d.Dial = func(context.Context) (certrotation.KV, func(), error) {
			return kv, func() {}, nil
		}
		_, err := pkicert.Rotate(context.Background(), d, pkicert.RotateInput{Name: "nats"})
		s.ErrorIs(err, os.ErrPermission)
	})
	s.Run("a cert_file of the operator is refused", func() {
		_, err := pkicert.Rotate(
			context.Background(),
			s.deps(),
			pkicert.RotateInput{Name: "mine"},
		)
		s.ErrorIs(err, pkicert.ErrOperatorCert)
	})
	s.Run("a name that is not a certificate", func() {
		_, err := pkicert.Rotate(
			context.Background(),
			s.deps(),
			pkicert.RotateInput{Name: "none"},
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
			pkicert.RotateInput{Name: "x"},
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
			pkicert.RotateInput{Name: "api"},
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
			pkicert.RotateInput{Name: "api"},
		)
		s.ErrorContains(err, "no PEM certificate")
	})
	s.Run("a directory that cannot be removed from", func() {
		s.Require().NoError(os.Chmod(s.dir, 0o500))
		defer func() { _ = os.Chmod(s.dir, 0o700) }()
		_, err := pkicert.Rotate(
			context.Background(),
			s.deps(),
			pkicert.RotateInput{Name: "api"},
		)
		s.ErrorContains(err, "remove")
	})
}
