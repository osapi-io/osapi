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

package cli_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/osapi-io/osapi/internal/config"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli"
)

type TLSPublicTestSuite struct {
	suite.Suite
}

// caFile writes a self-signed PEM certificate and returns its path.
func (s *TLSPublicTestSuite) caFile() string {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	s.Require().NoError(err)
	tmpl := x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "osapi"},
		NotBefore: time.Now(), NotAfter: time.Now().Add(time.Hour), IsCA: true,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &key.PublicKey, key)
	s.Require().NoError(err)
	path := filepath.Join(s.T().TempDir(), "api.crt")
	s.Require().
		NoError(os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o600))

	return path
}

func (s *TLSPublicTestSuite) TestTLSTransport() {
	tests := []struct {
		name         string
		caFile       func() string
		validateFunc func(http.RoundTripper, error)
	}{
		{
			name:   "when there is no ca_file, it is the default transport",
			caFile: func() string { return "" },
			validateFunc: func(rt http.RoundTripper, err error) {
				s.Require().NoError(err)
				s.Equal(http.DefaultTransport, rt)
			},
		},
		{
			name:   "when there is a ca_file, the transport trusts it",
			caFile: s.caFile,
			validateFunc: func(rt http.RoundTripper, err error) {
				s.Require().NoError(err)
				t, ok := rt.(*http.Transport)
				s.Require().True(ok)
				s.NotNil(t.TLSClientConfig.RootCAs)
			},
		},
		{
			name:   "when the ca_file does not exist, it is an error",
			caFile: func() string { return filepath.Join(s.T().TempDir(), "none.crt") },
			validateFunc: func(_ http.RoundTripper, err error) {
				s.Error(err)
			},
		},
		{
			name: "when the ca_file has no certificate, it is an error",
			caFile: func() string {
				path := filepath.Join(s.T().TempDir(), "empty.crt")
				s.Require().NoError(os.WriteFile(path, []byte("not a certificate"), 0o600))

				return path
			},
			validateFunc: func(_ http.RoundTripper, err error) {
				s.ErrorContains(err, "no PEM certificate")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.validateFunc(cli.TLSTransport(config.ClientTLS{CAFile: tc.caFile()}))
		})
	}
}

func TestTLSPublicTestSuite(t *testing.T) {
	suite.Run(t, new(TLSPublicTestSuite))
}
