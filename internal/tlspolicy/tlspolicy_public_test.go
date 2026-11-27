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

package tlspolicy_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/tlspolicy"
)

type TLSPolicyPublicTestSuite struct {
	suite.Suite
}

func TestTLSPolicyPublicTestSuite(t *testing.T) {
	suite.Run(t, new(TLSPolicyPublicTestSuite))
}

// cert is a self-signed P-256 certificate for localhost.
func (s *TLSPolicyPublicTestSuite) cert() tls.Certificate {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	s.Require().NoError(err)
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "localhost"},
		DNSNames:     []string{"localhost"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	s.Require().NoError(err)

	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}
}

// handshake connects client to a server with server, over loopback TCP,
// and returns what the client negotiated.
func (s *TLSPolicyPublicTestSuite) handshake(
	server *tls.Config,
	client *tls.Config,
) (tls.ConnectionState, error) {
	ln, err := tls.Listen("tcp", "127.0.0.1:0", server)
	s.Require().NoError(err)
	defer func() { _ = ln.Close() }()
	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		_ = c.(*tls.Conn).Handshake()
		_ = c.Close()
	}()
	c, err := tls.Dial("tcp", ln.Addr().String(), client)
	if err != nil {
		return tls.ConnectionState{}, err
	}
	defer func() { _ = c.Close() }()

	return c.ConnectionState(), nil
}

func (s *TLSPolicyPublicTestSuite) TestHandshake() {
	server := tlspolicy.Server()
	server.Certificates = []tls.Certificate{s.cert()}

	tests := []struct {
		name         string
		hardware     bool
		client       func() *tls.Config
		validateFunc func(tls.ConnectionState, error)
	}{
		{
			name:     "when a client of crypto/tls connects, it gets TLS 1.3, AES-GCM and an approved group",
			hardware: true,
			client: func() *tls.Config {
				return &tls.Config{InsecureSkipVerify: true} //nolint:gosec // test
			},
			validateFunc: func(cs tls.ConnectionState, err error) {
				s.Require().NoError(err)
				s.Equal(uint16(tls.VersionTLS13), cs.Version)
				s.NotEqual(tls.TLS_CHACHA20_POLY1305_SHA256, cs.CipherSuite)
				s.Contains(tlspolicy.Groups, cs.CurveID)
			},
		},
		{
			name:     "when the CPU has no AES in hardware, the client gets TLS 1.2 with AES-GCM, not ChaCha20",
			hardware: false,
			client: func() *tls.Config {
				return &tls.Config{InsecureSkipVerify: true} //nolint:gosec // test
			},
			validateFunc: func(cs tls.ConnectionState, err error) {
				s.Require().NoError(err)
				s.Equal(uint16(tls.VersionTLS12), cs.Version)
				s.Contains(tlspolicy.Suites, cs.CipherSuite)
			},
		},
		{
			name:     "when a client offers only X25519, it is refused",
			hardware: true,
			client: func() *tls.Config {
				return &tls.Config{ //nolint:gosec // test
					InsecureSkipVerify: true,
					CurvePreferences:   []tls.CurveID{tls.X25519},
				}
			},
			validateFunc: func(_ tls.ConnectionState, err error) {
				s.Error(err)
			},
		},
		{
			name:     "when a client of tlspolicy connects, it gets an approved suite and group",
			hardware: true,
			client: func() *tls.Config {
				c := tlspolicy.Client()
				c.InsecureSkipVerify = true

				return c
			},
			validateFunc: func(cs tls.ConnectionState, err error) {
				s.Require().NoError(err)
				s.NotEqual(tls.TLS_CHACHA20_POLY1305_SHA256, cs.CipherSuite)
				s.Contains(tlspolicy.Groups, cs.CurveID)
			},
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			defer tlspolicy.SetAESHardware(tc.hardware)()
			tc.validateFunc(s.handshake(server, tc.client()))
		})
	}
}

func (s *TLSPolicyPublicTestSuite) TestChaCha() {
	tls13 := []uint16{tls.VersionTLS13, tls.VersionTLS12}
	tests := []struct {
		name     string
		hardware bool
		hello    tls.ClientHelloInfo
		want     bool
	}{
		{
			name:     "when the client prefers AES-GCM, crypto/tls picks it",
			hardware: true,
			hello: tls.ClientHelloInfo{SupportedVersions: tls13, CipherSuites: []uint16{
				tls.TLS_AES_128_GCM_SHA256, tls.TLS_CHACHA20_POLY1305_SHA256,
			}},
		},
		{
			name:     "when the client prefers ChaCha20, crypto/tls picks it",
			hardware: true,
			hello: tls.ClientHelloInfo{SupportedVersions: tls13, CipherSuites: []uint16{
				tls.TLS_CHACHA20_POLY1305_SHA256, tls.TLS_AES_128_GCM_SHA256,
			}},
			want: true,
		},
		{
			name:     "when the client's first suite is one crypto/tls does not know, the next decides",
			hardware: true,
			hello: tls.ClientHelloInfo{SupportedVersions: tls13, CipherSuites: []uint16{
				0x0a0a, tls.TLS_AES_256_GCM_SHA384, tls.TLS_CHACHA20_POLY1305_SHA256,
			}},
		},
		{
			name:     "when the client's first known suite is an insecure one, crypto/tls picks ChaCha20",
			hardware: true,
			hello: tls.ClientHelloInfo{SupportedVersions: tls13, CipherSuites: []uint16{
				tls.TLS_RSA_WITH_RC4_128_SHA,
				tls.TLS_AES_128_GCM_SHA256,
				tls.TLS_CHACHA20_POLY1305_SHA256,
			}},
			want: true,
		},
		{
			name:     "when the client offers no AES-GCM for TLS 1.3, crypto/tls picks ChaCha20",
			hardware: true,
			hello: tls.ClientHelloInfo{SupportedVersions: tls13, CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256, tls.TLS_CHACHA20_POLY1305_SHA256,
			}},
			want: true,
		},
		{
			name:     "when only unknown suites come before ChaCha20, crypto/tls picks ChaCha20",
			hardware: true,
			hello: tls.ClientHelloInfo{SupportedVersions: tls13, CipherSuites: []uint16{
				0x0a0a,
				tls.TLS_CHACHA20_POLY1305_SHA256,
				tls.TLS_AES_128_GCM_SHA256,
			}},
			want: true,
		},
		{
			name:     "when the CPU has no AES in hardware, crypto/tls picks ChaCha20",
			hardware: false,
			hello: tls.ClientHelloInfo{SupportedVersions: tls13, CipherSuites: []uint16{
				tls.TLS_AES_128_GCM_SHA256, tls.TLS_CHACHA20_POLY1305_SHA256,
			}},
			want: true,
		},
		{
			name:     "when the client does not offer TLS 1.3, there is no TLS 1.3 suite to pick",
			hardware: true,
			hello: tls.ClientHelloInfo{
				SupportedVersions: []uint16{tls.VersionTLS12},
				CipherSuites:      []uint16{tls.TLS_CHACHA20_POLY1305_SHA256},
			},
		},
		{
			name:     "when the client does not offer ChaCha20, it cannot be picked",
			hardware: false,
			hello: tls.ClientHelloInfo{SupportedVersions: tls13, CipherSuites: []uint16{
				tls.TLS_AES_128_GCM_SHA256,
			}},
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			defer tlspolicy.SetAESHardware(tc.hardware)()
			s.Equal(tc.want, tlspolicy.ChaCha(&tc.hello))
		})
	}
}

func (s *TLSPolicyPublicTestSuite) TestApproved() {
	tests := []struct {
		name   string
		config *tls.Config
		server bool
		want   bool
	}{
		{name: "when there is no config, there is no TLS to judge", want: true},
		{
			name:   "when it is Server's, a server is approved",
			config: tlspolicy.Server(),
			server: true,
			want:   true,
		},
		{name: "when it is Client's, a client is approved", config: tlspolicy.Client(), want: true},
		{
			name:   "when it is Client's, a server lacks the TLS 1.3 guard",
			config: tlspolicy.Client(),
			server: true,
		},
		{name: "when it sets nothing, it has crypto/tls's defaults", config: &tls.Config{}},
		{
			name: "when it accepts TLS 1.1, it is not approved",
			config: func() *tls.Config {
				c := tlspolicy.Client()
				c.MinVersion = tls.VersionTLS11

				return c
			}(),
		},
		{
			name: "when it sets a suite that is not approved, it is not approved",
			config: func() *tls.Config {
				c := tlspolicy.Client()
				c.CipherSuites = append(c.CipherSuites, tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305)

				return c
			}(),
		},
		{
			name: "when it sets a group that is not approved, it is not approved",
			config: func() *tls.Config {
				c := tlspolicy.Client()
				c.CurvePreferences = []tls.CurveID{tls.X25519}

				return c
			}(),
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, tlspolicy.Approved(tc.config, tc.server))
		})
	}
}
