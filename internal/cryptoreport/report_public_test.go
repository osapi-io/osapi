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

package cryptoreport_test

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/fips140"
	"crypto/tls"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/nats-io/nkeys"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cryptoreport"
	"github.com/osapi-io/osapi/internal/pki"
	"github.com/osapi-io/osapi/internal/tlspolicy"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

type ReportPublicTestSuite struct {
	suite.Suite
}

func (s *ReportPublicTestSuite) TearDownTest() {
	cryptoreport.ResetMode()
}

// approvedTLS sets only approved suites and groups, with no TLS 1.3 guard.
var approvedTLS = &tls.Config{
	MinVersion:       tls.VersionTLS13,
	CipherSuites:     []uint16{tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256},
	CurvePreferences: []tls.CurveID{tls.CurveP256},
}

func (s *ReportPublicTestSuite) TestNew() {
	tests := []struct {
		name         string
		mode         types.FIPSMode
		listeners    []cryptoreport.TLS
		uses         []cryptoreport.Algorithm
		validateFunc func(r cryptoreport.Report)
	}{
		{
			name: "when the mode is off, the mode is a gap",
			mode: types.FIPSOff,
			uses: []cryptoreport.Algorithm{cryptoreport.Seal},
			validateFunc: func(r cryptoreport.Report) {
				s.Equal(types.FIPSOff, r.Mode)
				s.Equal(runtime.Version(), r.Go)
				s.Equal([]string{"FIPS 140-3 mode is off: set GODEBUG=fips140=on"}, r.Gaps)
				s.True(r.Enforceable)
			},
		},
		{
			name: "when the mode is on and each use is approved, there is no gap",
			mode: types.FIPSOn,
			uses: []cryptoreport.Algorithm{cryptoreport.Seal, cryptoreport.Vault},
			validateFunc: func(r cryptoreport.Report) {
				s.Empty(r.Gaps)
				s.Len(r.Algorithms, 2)
			},
		},
		{
			name: "when a use fails in only mode, the process is not enforceable",
			mode: types.FIPSOn,
			uses: []cryptoreport.Algorithm{cryptoreport.Seal, cryptoreport.NATSAuth},
			validateFunc: func(r cryptoreport.Report) {
				s.Empty(r.Gaps)
				s.False(r.Enforceable)
			},
		},
		{
			name: "when a use is not approved, it is a gap",
			mode: types.FIPSOnly,
			uses: []cryptoreport.Algorithm{{Name: "old", Algorithm: "MD5"}},
			validateFunc: func(r cryptoreport.Report) {
				s.Equal(types.FIPSOnly, r.Mode)
				s.Equal([]string{"old: MD5 is not an approved algorithm"}, r.Gaps)
			},
		},
		{
			name:      "when a listener serves no TLS, it has only its name",
			mode:      types.FIPSOn,
			listeners: []cryptoreport.TLS{{Name: "metrics"}},
			validateFunc: func(r cryptoreport.Report) {
				s.Equal([]cryptoreport.Listener{{Name: "metrics"}}, r.Listeners)
				s.Empty(r.Gaps)
			},
		},
		{
			name: "when a listener sets nothing outside FIPS mode, it is a gap",
			mode: types.FIPSOff,
			listeners: []cryptoreport.TLS{{Name: "api", Config: &tls.Config{
				MinVersion: tls.VersionTLS12,
			}}},
			validateFunc: func(r cryptoreport.Report) {
				l := r.Listeners[0]
				s.True(l.TLS)
				s.Equal("TLS 1.2", l.MinVersion)
				s.Equal([]string{"crypto/tls default"}, l.Groups)
				s.NotEmpty(l.Suites)
				s.False(l.Approved)
				s.Contains(r.Gaps,
					"api: crypto/tls may negotiate a suite or group that is not approved")
			},
		},
		{
			name:      "when a listener sets nothing in FIPS mode, crypto/tls restricts it",
			mode:      types.FIPSOn,
			listeners: []cryptoreport.TLS{{Name: "nats", Config: &tls.Config{}}},
			validateFunc: func(r cryptoreport.Report) {
				s.Equal("TLS 1.2", r.Listeners[0].MinVersion)
				s.True(r.Listeners[0].Approved)
				s.Empty(r.Gaps)
			},
		},
		{
			name:      "when a listener has tlspolicy's server config, it is approved",
			mode:      types.FIPSOff,
			listeners: []cryptoreport.TLS{{Name: "api", Config: tlspolicy.Server()}},
			validateFunc: func(r cryptoreport.Report) {
				l := r.Listeners[0]
				s.Equal("TLS 1.2", l.MinVersion)
				s.Equal([]string{
					"TLS_AES_128_GCM_SHA256",
					"TLS_AES_256_GCM_SHA384",
					"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
					"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
					"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
					"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
				}, l.Suites)
				s.Equal(
					[]string{"SecP256r1MLKEM768", "CurveP256", "CurveP384", "CurveP521"},
					l.Groups,
				)
				s.True(l.Approved)
				s.Len(r.Gaps, 1)
			},
		},
		{
			name:      "when a listener sets approved suites and groups with no TLS 1.3 guard, ChaCha20 is a gap",
			mode:      types.FIPSOff,
			listeners: []cryptoreport.TLS{{Name: "api", Config: approvedTLS}},
			validateFunc: func(r cryptoreport.Report) {
				l := r.Listeners[0]
				s.Equal("TLS 1.3", l.MinVersion)
				s.Contains(l.Suites, "TLS_CHACHA20_POLY1305_SHA256")
				s.False(l.Approved)
				s.Len(r.Gaps, 2)
			},
		},
		{
			name: "when a listener sets a group that is not approved, it is a gap",
			mode: types.FIPSOff,
			listeners: []cryptoreport.TLS{{Name: "api", Config: &tls.Config{
				CipherSuites:     approvedTLS.CipherSuites,
				CurvePreferences: []tls.CurveID{tls.X25519},
			}}},
			validateFunc: func(r cryptoreport.Report) {
				s.False(r.Listeners[0].Approved)
			},
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			cryptoreport.SetMode(tc.mode)
			tc.validateFunc(cryptoreport.New(tc.listeners, tc.uses))
		})
	}
}

func (s *ReportPublicTestSuite) TestCertificates() {
	yes, no := true, false
	tests := []struct {
		name  string
		certs []pki.Cert
		want  []cryptoreport.Algorithm
	}{
		{
			name: "when a certificate is approved",
			certs: []pki.Cert{
				{
					Name:      "api",
					State:     pki.CertOK,
					Key:       "ECDSA P-256",
					Signature: "ECDSA-SHA256",
					FIPS:      &yes,
				},
			},
			want: []cryptoreport.Algorithm{{
				Name: "certificate api", Algorithm: "ECDSA P-256, ECDSA-SHA256",
				Approved: true, OnlyMode: true,
			}},
		},
		{
			name: "when a certificate is not approved",
			certs: []pki.Cert{{
				Name: "old", State: pki.CertOK, Key: "RSA 1024", Signature: "SHA1-RSA", FIPS: &no,
			}},
			want: []cryptoreport.Algorithm{{
				Name: "certificate old", Algorithm: "RSA 1024, SHA1-RSA",
				Reason: "the key or the signature algorithm is not approved",
			}},
		},
		{
			name:  "when a certificate is not made, it is not a use",
			certs: []pki.Cert{{Name: "nats", State: pki.CertNotMade}},
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, cryptoreport.Certificates(tc.certs))
		})
	}
}

func (s *ReportPublicTestSuite) TestVerdict() {
	tests := []struct {
		name   string
		report cryptoreport.Report
		want   string
	}{
		{
			name:   "when the mode is off and it is the one gap",
			report: cryptoreport.Report{Mode: types.FIPSOff, Gaps: []string{"off"}},
			want:   "off",
		},
		{
			name:   "when the mode is off with one other gap",
			report: cryptoreport.Report{Mode: types.FIPSOff, Gaps: []string{"off", "md5"}},
			want:   "off, 1 gap",
		},
		{
			name:   "when the mode is on with no gap",
			report: cryptoreport.Report{Mode: types.FIPSOn},
			want:   "on, approved",
		},
		{
			name:   "when the mode is only with two gaps",
			report: cryptoreport.Report{Mode: types.FIPSOnly, Gaps: []string{"a", "b"}},
			want:   "only, 2 gaps",
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, tc.report.Verdict())
		})
	}
}

func TestReportPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ReportPublicTestSuite))
}

// probe runs f, and is its error or the panic that it raised: FIPS 140-3
// only mode refuses some calls with a panic.
func probe(
	f func() error,
) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	return f()
}

// nkeysSign is what osapi does with nkeys: make a user and sign with it.
func nkeysSign() error {
	kp, err := nkeys.CreateUser()
	if err != nil {
		return err
	}
	_, err = kp.Sign([]byte("osapi"))

	return err
}

// storeSeal is the call that nats-server makes to encrypt a block of the
// JetStream store: AES-GCM with a nonce that it picks.
func storeSeal() error {
	block, err := aes.NewCipher(make([]byte, 32))
	if err != nil {
		return err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	aead.Seal(nil, make([]byte, aead.NonceSize()), []byte("osapi"), nil)

	return nil
}

func (s *ReportPublicTestSuite) TestProbes() {
	tests := []struct {
		name  string
		use   cryptoreport.Algorithm
		probe func() error
	}{
		{name: "NATS JWTs sign with nkeys", use: cryptoreport.NATSAuth, probe: nkeysSign},
		{
			name:  "the JetStream store seals with its own nonce",
			use:   cryptoreport.Store,
			probe: storeSeal,
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			err := probe(tc.probe)
			if !fips140.Enforced() {
				s.NoError(err)
				return
			}
			s.Equal(
				tc.use.OnlyMode,
				err == nil,
			)
		})
	}
}

// TestOnlyModeIsTested: each use that claims to work in only mode is in a
// package that the fips140=only CI job tests.
func (s *ReportPublicTestSuite) TestOnlyModeIsTested() {
	data, err := os.ReadFile("../../justfile")
	s.Require().NoError(err)
	m := regexp.MustCompile(`(?m)^go_fips_packages := "([^"]+)"`).FindSubmatch(data)
	s.Require().NotNil(m)
	tested := strings.Fields(string(m[1]))
	tests := []struct {
		use     cryptoreport.Algorithm
		pkg     string
		reasons bool
	}{
		{use: cryptoreport.APITokens, pkg: "./internal/authtoken/"},
		{use: cryptoreport.Signatures, pkg: "./internal/pki/"},
		{use: cryptoreport.Seal, pkg: "./internal/secret/seal/"},
		{use: cryptoreport.Vault, pkg: "./internal/secret/vault/natsvault/"},
		{use: cryptoreport.NATSAuth, pkg: "./internal/cryptoreport/"},
		{use: cryptoreport.Store, pkg: "./internal/cryptoreport/"},
	}
	for _, tc := range tests {
		s.Run(tc.use.Name, func() {
			s.Contains(tested, tc.pkg)
			s.True(tc.use.Approved)
			if !tc.use.OnlyMode {
				s.NotEmpty(tc.use.Reason)
			}
		})
	}
}

// TestMode: the report reads the mode that Go runs the process in.
func (s *ReportPublicTestSuite) TestMode() {
	cryptoreport.ResetMode()
	s.Equal(cryptoreport.ModeOf(fips140.Enabled(), fips140.Enforced()),
		cryptoreport.New(nil, nil).Mode)
	tests := []struct {
		name     string
		enabled  bool
		enforced bool
		want     types.FIPSMode
	}{
		{name: "when the module is not enabled", want: types.FIPSOff},
		{name: "when it is enabled", enabled: true, want: types.FIPSOn},
		{name: "when it is enforced", enabled: true, enforced: true, want: types.FIPSOnly},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, cryptoreport.ModeOf(tc.enabled, tc.enforced))
		})
	}
}
