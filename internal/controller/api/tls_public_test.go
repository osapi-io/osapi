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

package api_test

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller/api"
)

type TLSPublicTestSuite struct {
	suite.Suite
}

// leaf is the certificate of the first PEM block of file.
func (s *TLSPublicTestSuite) leaf(
	file string,
) *x509.Certificate {
	data, err := os.ReadFile(file)
	s.Require().NoError(err)
	block, _ := pem.Decode(data)
	s.Require().NotNil(block)
	cert, err := x509.ParseCertificate(block.Bytes)
	s.Require().NoError(err)

	return cert
}

func (s *TLSPublicTestSuite) TestTLSConfig() {
	tests := []struct {
		name         string
		setup        func(dir string) config.ServerTLS
		validateFunc func(dir string, err error)
	}{
		{
			name:  "when no certificate is set, it makes a self-signed one for this host and the URLs",
			setup: func(string) config.ServerTLS { return config.ServerTLS{Enabled: true} },
			validateFunc: func(dir string, err error) {
				s.Require().NoError(err)
				cert := s.leaf(filepath.Join(dir, api.SelfSignedCert))
				s.Contains(cert.DNSNames, "localhost")
				s.Contains(cert.DNSNames, "osapi.example.com")
				s.Len(cert.IPAddresses, 3, "127.0.0.1, ::1 and 10.0.0.1")
				info, err := os.Stat(filepath.Join(dir, api.SelfSignedKey))
				s.Require().NoError(err)
				s.Equal(os.FileMode(0o600), info.Mode().Perm())
			},
		},
		{
			name: "when the self-signed certificate exists, it is used again",
			setup: func(dir string) config.ServerTLS {
				_, err := api.TLSConfig(config.ServerTLS{}, dir)
				s.Require().NoError(err)
				s.Require().NoError(os.WriteFile(filepath.Join(dir, "marker"), nil, 0o600))

				return config.ServerTLS{}
			},
			validateFunc: func(dir string, err error) {
				s.Require().NoError(err)
				s.NotContains(
					s.leaf(filepath.Join(dir, api.SelfSignedCert)).DNSNames,
					"osapi.example.com",
					"made the first time, with no URL",
				)
			},
		},
		{
			name: "when a certificate is set, it is served",
			setup: func(dir string) config.ServerTLS {
				_, err := api.TLSConfig(config.ServerTLS{}, filepath.Join(dir, "real"))
				s.Require().NoError(err)

				return config.ServerTLS{
					CertFile: filepath.Join(dir, "real", api.SelfSignedCert),
					KeyFile:  filepath.Join(dir, "real", api.SelfSignedKey),
				}
			},
			validateFunc: func(dir string, err error) {
				s.Require().NoError(err)
				_, statErr := os.Stat(filepath.Join(dir, api.SelfSignedCert))
				s.True(os.IsNotExist(statErr), "no self-signed certificate is made")
			},
		},
		{
			name: "when the certificate does not load, it is an error",
			setup: func(dir string) config.ServerTLS {
				return config.ServerTLS{
					CertFile: filepath.Join(dir, "none.crt"),
					KeyFile:  filepath.Join(dir, "none.key"),
				}
			},
			validateFunc: func(_ string, err error) {
				s.ErrorContains(err, "load the API's certificate")
			},
		},
		{
			name: "when the host has no name, the certificate has localhost and the URLs",
			setup: func(string) config.ServerTLS {
				api.SetHostnameFn(func() (string, error) { return "", errors.New("no name") })

				return config.ServerTLS{}
			},
			validateFunc: func(dir string, err error) {
				s.Require().NoError(err)
				s.Equal([]string{"localhost", "osapi.example.com"},
					s.leaf(filepath.Join(dir, api.SelfSignedCert)).DNSNames)
			},
		},
		{
			name: "when the key cannot be written, it is an error",
			setup: func(string) config.ServerTLS {
				api.SetWriteFileFn(
					func(string, []byte, os.FileMode) error { return errors.New("disk full") },
				)

				return config.ServerTLS{}
			},
			validateFunc: func(_ string, err error) {
				s.ErrorContains(err, "write the key of")
			},
		},
		{
			name: "when the certificate cannot be written, it is an error",
			setup: func(string) config.ServerTLS {
				api.SetWriteFileFn(func(name string, data []byte, perm os.FileMode) error {
					if filepath.Base(name) == api.SelfSignedCert {
						return errors.New("disk full")
					}

					return os.WriteFile(name, data, perm)
				})

				return config.ServerTLS{}
			},
			validateFunc: func(_ string, err error) {
				s.ErrorContains(err, "api.crt")
			},
		},
		{
			name: "when the key directory cannot be made, it is an error",
			setup: func(dir string) config.ServerTLS {
				s.Require().NoError(os.WriteFile(filepath.Join(dir, "file"), nil, 0o600))

				return config.ServerTLS{}
			},
			validateFunc: func(_ string, err error) {
				s.ErrorContains(err, "make the key directory")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			defer api.ResetTLSFns()
			dir := s.T().TempDir()
			c := tc.setup(dir)
			keyDir := dir
			if tc.name == "when the key directory cannot be made, it is an error" {
				keyDir = filepath.Join(dir, "file", "pki")
			}
			_, err := api.TLSConfig(
				c,
				keyDir,
				"https://osapi.example.com:8443",
				"http://10.0.0.1:8080",
				"::bad",
			)
			tc.validateFunc(dir, err)
		})
	}
}

func TestTLSPublicTestSuite(t *testing.T) {
	suite.Run(t, new(TLSPublicTestSuite))
}

// TestWithTLS: a server with a TLS config serves HTTPS.
func (s *TLSPublicTestSuite) TestWithTLS() {
	c, err := api.TLSConfig(config.ServerTLS{}, s.T().TempDir())
	s.Require().NoError(err)
	s.NotNil(api.New(config.Config{}, slog.Default(), api.WithTLS(c)))
}

// TestCertFingerprint is the SHA-256 of the served certificate, of a set
// certificate or of the self-signed one.
func (s *TLSPublicTestSuite) TestCertFingerprint() {
	dir := s.T().TempDir()
	_, err := api.TLSConfig(config.ServerTLS{}, dir)
	s.Require().NoError(err)

	self, err := api.CertFingerprint(config.ServerTLS{}, dir)
	s.Require().NoError(err)
	s.Regexp(`^SHA256:[0-9a-f]{64}$`, self)

	set, err := api.CertFingerprint(
		config.ServerTLS{CertFile: filepath.Join(dir, api.SelfSignedCert)},
		"/elsewhere",
	)
	s.Require().NoError(err)
	s.Equal(self, set, "the same certificate has the same fingerprint")

	_, err = api.CertFingerprint(config.ServerTLS{}, s.T().TempDir())
	s.ErrorContains(err, "read the API's certificate")

	bad := filepath.Join(dir, "bad.crt")
	s.Require().NoError(os.WriteFile(bad, []byte("not pem"), 0o600))
	_, err = api.CertFingerprint(config.ServerTLS{CertFile: bad}, dir)
	s.ErrorContains(err, "no PEM certificate")
}
