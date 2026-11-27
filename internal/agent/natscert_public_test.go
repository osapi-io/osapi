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

package agent_test

import (
	"context"
	"crypto/ed25519"
	"encoding/pem"
	"errors"
	"path/filepath"

	"github.com/avfs/avfs"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/agent"
	"github.com/osapi-io/osapi/internal/pki"
)

// certPEM is a PEM certificate block with der as its bytes.
func certPEM(
	der string,
) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte(der)})
}

// natsCAFile is the CA file of the agent's join in the tests' key directory.
var natsCAFile = filepath.Join(keyDir, "nats-join-ca.pem")

func (s *AgentPublicTestSuite) TestFollowNATSCert() {
	current, next := certPEM("current"), certPEM("next")
	signed := func(rec pki.CertRecord, priv ed25519.PrivateKey) pki.CertRecord {
		rec.Sign(priv)
		return rec
	}

	tests := []struct {
		name     string
		caFile   string
		setup    func(a *agent.Agent, priv ed25519.PrivateKey)
		want     string
		wantFile []byte
	}{
		{
			name:   "a rotation that started: the agent trusts both certificates, and reports next",
			caFile: natsCAFile,
			setup: func(_ *agent.Agent, priv ed25519.PrivateKey) {
				s.Require().NoError(avfs.WriteFile(s.appFs, natsCAFile, current, 0o644))
				s.mockJobClient.EXPECT().NATSCert(gomock.Any()).Return(
					signed(
						pki.CertRecord{Serial: 1, Name: "nats", Current: current, Next: next},
						priv,
					),
					true, nil)
			},
			want:     pki.CertFingerprint(next),
			wantFile: append(append([]byte{}, current...), next...),
		},
		{
			name:   "the switch: the agent trusts the new certificate alone",
			caFile: natsCAFile,
			setup: func(_ *agent.Agent, priv ed25519.PrivateKey) {
				s.Require().
					NoError(avfs.WriteFile(s.appFs, natsCAFile, append(current, next...), 0o644))
				s.mockJobClient.EXPECT().NATSCert(gomock.Any()).Return(
					signed(pki.CertRecord{Serial: 2, Name: "nats", Current: next}, priv), true, nil)
			},
			wantFile: next,
		},
		{
			name:   "a certificate that the record does not name is replaced",
			caFile: natsCAFile,
			setup: func(_ *agent.Agent, priv ed25519.PrivateKey) {
				s.Require().NoError(avfs.WriteFile(s.appFs, natsCAFile, certPEM("old"), 0o644))
				s.mockJobClient.EXPECT().NATSCert(gomock.Any()).Return(
					signed(
						pki.CertRecord{Serial: 3, Name: "nats", Current: current},
						priv,
					), true, nil)
			},
			wantFile: current,
		},
		{
			name:   "certificates that it trusts already change nothing",
			caFile: natsCAFile,
			setup: func(a *agent.Agent, priv ed25519.PrivateKey) {
				s.Require().NoError(avfs.WriteFile(s.appFs, natsCAFile, current, 0o644))
				agent.SetAppFs(a, failingOn(s.appFs, avfs.FnRename))
				s.mockJobClient.EXPECT().NATSCert(gomock.Any()).Return(
					signed(
						pki.CertRecord{Serial: 1, Name: "nats", Current: current},
						priv,
					), true, nil)
			},
			wantFile: current,
		},
		{
			name:   "a file that does not write keeps what it trusts, and reports nothing",
			caFile: natsCAFile,
			setup: func(a *agent.Agent, priv ed25519.PrivateKey) {
				s.Require().NoError(avfs.WriteFile(s.appFs, natsCAFile, current, 0o644))
				agent.SetAppFs(a, failingOn(s.appFs, avfs.FnRename))
				s.mockJobClient.EXPECT().NATSCert(gomock.Any()).Return(
					signed(
						pki.CertRecord{Serial: 1, Name: "nats", Current: current, Next: next},
						priv,
					),
					true, nil)
			},
			wantFile: current,
		},
		{
			name:   "a record that no controller key signed changes nothing",
			caFile: natsCAFile,
			setup: func(_ *agent.Agent, _ ed25519.PrivateKey) {
				s.Require().NoError(avfs.WriteFile(s.appFs, natsCAFile, current, 0o644))
				_, other := s.controllerKeyPair()
				s.mockJobClient.EXPECT().NATSCert(gomock.Any()).Return(
					signed(
						pki.CertRecord{Serial: 1, Name: "nats", Current: next, Next: next},
						other,
					),
					true, nil)
			},
			wantFile: current,
		},
		{
			name:   "no record changes nothing",
			caFile: natsCAFile,
			setup: func(_ *agent.Agent, _ ed25519.PrivateKey) {
				s.Require().NoError(avfs.WriteFile(s.appFs, natsCAFile, current, 0o644))
				s.mockJobClient.EXPECT().NATSCert(gomock.Any()).Return(pki.CertRecord{}, false, nil)
			},
			wantFile: current,
		},
		{
			name:   "a record that does not read changes nothing",
			caFile: natsCAFile,
			setup: func(_ *agent.Agent, _ ed25519.PrivateKey) {
				s.Require().NoError(avfs.WriteFile(s.appFs, natsCAFile, current, 0o644))
				s.mockJobClient.EXPECT().NATSCert(gomock.Any()).
					Return(pki.CertRecord{}, false, errors.New("timeout"))
			},
			wantFile: current,
		},
		{
			name:   "an agent that trusts the server's own certificate reports next, and writes nothing",
			caFile: "/nats/nats.crt",
			setup: func(_ *agent.Agent, priv ed25519.PrivateKey) {
				s.mockJobClient.EXPECT().NATSCert(gomock.Any()).Return(
					signed(
						pki.CertRecord{Serial: 1, Name: "nats", Current: current, Next: next},
						priv,
					),
					true, nil)
			},
			want: pki.CertFingerprint(next),
		},
		{
			name:   "an agent that trusts the server's own certificate reports nothing after the switch",
			caFile: "/nats/nats.crt",
			setup: func(_ *agent.Agent, priv ed25519.PrivateKey) {
				s.mockJobClient.EXPECT().NATSCert(gomock.Any()).Return(
					signed(pki.CertRecord{Serial: 2, Name: "nats", Current: next}, priv), true, nil)
			},
		},
		{
			name:   "a CA file of the agent's own config is the operator's, and is not read",
			caFile: "/etc/ssl/private-ca.pem",
			setup:  func(*agent.Agent, ed25519.PrivateKey) {},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()
			cfg := s.appConfig
			cfg.Agent.PKI.KeyDir = keyDir
			cfg.Agent.NATS.TLS.CAFile = tc.caFile
			cfg.NATS.Server.KeyDir = "/nats"
			s.appConfig = cfg
			a := s.buildAgent()
			m := pki.New(s.appFs, keyDir, "agent")
			pub, priv := s.controllerKeyPair()
			m.SetControllerPublicKey(pub)
			agent.SetAgentPKIManager(a, m)
			tc.setup(a, priv)
			s.Equal(tc.want, agent.ExportFollowNATSCert(context.Background(), a))
			if tc.wantFile != nil {
				data, err := avfs.ReadFile(s.appFs, natsCAFile)
				s.Require().NoError(err)
				s.Equal(tc.wantFile, data)
			}
		})
	}
}
