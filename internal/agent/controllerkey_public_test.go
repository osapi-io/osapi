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
	"crypto/rand"
	"encoding/pem"
	"errors"
	"path/filepath"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/failfs"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/agent"
	"github.com/osapi-io/osapi/internal/pki"
)

// failingOn is fs, where the call fn fails.
func failingOn(
	fs avfs.VFS,
	fn avfs.FnVFS,
) avfs.VFS {
	vfs := failfs.New(fs)
	_ = vfs.SetFailFunc(func(_ avfs.VFSBase, got avfs.FnVFS, _ *failfs.FailParam) error {
		if got == fn {
			return errors.New("injected")
		}

		return nil
	})

	return vfs
}

// controllerKeyPair is a controller key for the tests.
func (s *AgentPublicTestSuite) controllerKeyPair() (ed25519.PublicKey, ed25519.PrivateKey) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	s.Require().NoError(err)

	return pub, priv
}

// keyFile is the PEM file of key.
func keyFile(
	key ed25519.PublicKey,
) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "ED25519 PUBLIC KEY", Bytes: key})
}

func (s *AgentPublicTestSuite) TestFollowControllerKey() {
	nextPath := filepath.Join(keyDir, agent.EnrolledControllerNextKey)
	currentPath := filepath.Join(keyDir, agent.EnrolledControllerKey)

	tests := []struct {
		name         string
		setup        func(a *agent.Agent, m *pki.Manager, current ed25519.PublicKey, currentPriv ed25519.PrivateKey)
		validateFunc func(m *pki.Manager, current ed25519.PublicKey)
	}{
		{
			name: "a rotation that started: the agent keeps the next key",
			setup: func(_ *agent.Agent, _ *pki.Manager, current ed25519.PublicKey, priv ed25519.PrivateKey) {
				next, _ := s.controllerKeyPair()
				rec := pki.ControllerKeyRecord{Serial: 1, Current: current, Next: next}
				rec.Sign(priv)
				s.mockJobClient.EXPECT().ControllerKey(gomock.Any()).Return(rec, true, nil)
			},
			validateFunc: func(m *pki.Manager, current ed25519.PublicKey) {
				next := m.NextControllerPublicKey()
				s.NotNil(next)
				s.Equal(current, m.ControllerPublicKey())
				data, err := avfs.ReadFile(s.appFs, nextPath)
				s.Require().NoError(err)
				s.Equal(keyFile(next), data)
			},
		},
		{
			name: "the next key that it has already changes nothing",
			setup: func(_ *agent.Agent, m *pki.Manager, current ed25519.PublicKey, priv ed25519.PrivateKey) {
				next, _ := s.controllerKeyPair()
				m.SetNextControllerPublicKey(next)
				rec := pki.ControllerKeyRecord{Serial: 1, Current: current, Next: next}
				rec.Sign(priv)
				s.mockJobClient.EXPECT().ControllerKey(gomock.Any()).Return(rec, true, nil)
			},
			validateFunc: func(m *pki.Manager, _ ed25519.PublicKey) {
				s.NotNil(m.NextControllerPublicKey())
				_, err := s.appFs.Stat(nextPath)
				s.Error(err)
			},
		},
		{
			name: "a next key that does not write is not kept",
			setup: func(a *agent.Agent, _ *pki.Manager, current ed25519.PublicKey, priv ed25519.PrivateKey) {
				agent.SetAppFs(a, failingOn(s.appFs, avfs.FnRename))
				next, _ := s.controllerKeyPair()
				rec := pki.ControllerKeyRecord{Serial: 1, Current: current, Next: next}
				rec.Sign(priv)
				s.mockJobClient.EXPECT().ControllerKey(gomock.Any()).Return(rec, true, nil)
			},
			validateFunc: func(m *pki.Manager, _ ed25519.PublicKey) {
				s.Nil(m.NextControllerPublicKey())
			},
		},
		{
			name: "a record with no rotation drops a next key that it replaced",
			setup: func(_ *agent.Agent, m *pki.Manager, current ed25519.PublicKey, priv ed25519.PrivateKey) {
				next, _ := s.controllerKeyPair()
				m.SetNextControllerPublicKey(next)
				s.Require().NoError(avfs.WriteFile(s.appFs, nextPath, keyFile(next), 0o644))
				rec := pki.ControllerKeyRecord{Serial: 2, Current: current}
				rec.Sign(priv)
				s.mockJobClient.EXPECT().ControllerKey(gomock.Any()).Return(rec, true, nil)
			},
			validateFunc: func(m *pki.Manager, _ ed25519.PublicKey) {
				s.Nil(m.NextControllerPublicKey())
				_, err := s.appFs.Stat(nextPath)
				s.Error(err)
			},
		},
		{
			name: "a next key that does not remove stays",
			setup: func(a *agent.Agent, m *pki.Manager, current ed25519.PublicKey, priv ed25519.PrivateKey) {
				next, _ := s.controllerKeyPair()
				m.SetNextControllerPublicKey(next)
				s.Require().NoError(avfs.WriteFile(s.appFs, nextPath, keyFile(next), 0o644))
				agent.SetAppFs(a, failingOn(s.appFs, avfs.FnRemove))
				rec := pki.ControllerKeyRecord{Serial: 2, Current: current}
				rec.Sign(priv)
				s.mockJobClient.EXPECT().ControllerKey(gomock.Any()).Return(rec, true, nil)
			},
			validateFunc: func(m *pki.Manager, _ ed25519.PublicKey) {
				s.NotNil(m.NextControllerPublicKey())
			},
		},
		{
			name: "a record with no rotation and no next key changes nothing",
			setup: func(_ *agent.Agent, _ *pki.Manager, current ed25519.PublicKey, priv ed25519.PrivateKey) {
				rec := pki.ControllerKeyRecord{Serial: 2, Current: current}
				rec.Sign(priv)
				s.mockJobClient.EXPECT().ControllerKey(gomock.Any()).Return(rec, true, nil)
			},
			validateFunc: func(m *pki.Manager, current ed25519.PublicKey) {
				s.Equal(current, m.ControllerPublicKey())
				s.Nil(m.NextControllerPublicKey())
			},
		},
		{
			name: "the switch: the next key becomes the controller key, and the old one the previous",
			setup: func(_ *agent.Agent, m *pki.Manager, current ed25519.PublicKey, _ ed25519.PrivateKey) {
				next, nextPriv := s.controllerKeyPair()
				m.SetNextControllerPublicKey(next)
				s.Require().NoError(avfs.WriteFile(s.appFs, nextPath, keyFile(next), 0o644))
				rec := pki.ControllerKeyRecord{Serial: 2, Current: next, Previous: current}
				rec.Sign(nextPriv)
				s.mockJobClient.EXPECT().ControllerKey(gomock.Any()).Return(rec, true, nil)
			},
			validateFunc: func(m *pki.Manager, current ed25519.PublicKey) {
				s.NotEqual(current, m.ControllerPublicKey())
				s.Equal(current, m.PreviousControllerPublicKey())
				s.Nil(m.NextControllerPublicKey())
				data, err := avfs.ReadFile(s.appFs, currentPath)
				s.Require().NoError(err)
				s.Equal(keyFile(m.ControllerPublicKey()), data)
				_, err = s.appFs.Stat(nextPath)
				s.Error(err)
			},
		},
		{
			name: "a switch whose key does not write keeps the old key",
			setup: func(a *agent.Agent, m *pki.Manager, current ed25519.PublicKey, _ ed25519.PrivateKey) {
				next, nextPriv := s.controllerKeyPair()
				m.SetNextControllerPublicKey(next)
				agent.SetAppFs(a, failingOn(s.appFs, avfs.FnRename))
				rec := pki.ControllerKeyRecord{Serial: 2, Current: next, Previous: current}
				rec.Sign(nextPriv)
				s.mockJobClient.EXPECT().ControllerKey(gomock.Any()).Return(rec, true, nil)
			},
			validateFunc: func(m *pki.Manager, current ed25519.PublicKey) {
				s.Equal(current, m.ControllerPublicKey())
			},
		},
		{
			name: "a switch whose next file does not remove still switches",
			setup: func(a *agent.Agent, m *pki.Manager, current ed25519.PublicKey, _ ed25519.PrivateKey) {
				next, nextPriv := s.controllerKeyPair()
				m.SetNextControllerPublicKey(next)
				agent.SetAppFs(a, failingOn(s.appFs, avfs.FnRemove))
				rec := pki.ControllerKeyRecord{Serial: 2, Current: next, Previous: current}
				rec.Sign(nextPriv)
				s.mockJobClient.EXPECT().ControllerKey(gomock.Any()).Return(rec, true, nil)
			},
			validateFunc: func(m *pki.Manager, current ed25519.PublicKey) {
				s.NotEqual(current, m.ControllerPublicKey())
			},
		},
		{
			name: "a key that the agent does not have changes nothing",
			setup: func(_ *agent.Agent, _ *pki.Manager, _ ed25519.PublicKey, _ ed25519.PrivateKey) {
				other, otherPriv := s.controllerKeyPair()
				rec := pki.ControllerKeyRecord{Serial: 5, Current: other}
				rec.Sign(otherPriv)
				s.mockJobClient.EXPECT().ControllerKey(gomock.Any()).Return(rec, true, nil)
			},
			validateFunc: func(m *pki.Manager, current ed25519.PublicKey) {
				s.Equal(current, m.ControllerPublicKey())
			},
		},
		{
			name: "a record that the agent's key did not sign changes nothing",
			setup: func(_ *agent.Agent, _ *pki.Manager, current ed25519.PublicKey, _ ed25519.PrivateKey) {
				next, _ := s.controllerKeyPair()
				_, otherPriv := s.controllerKeyPair()
				rec := pki.ControllerKeyRecord{Serial: 5, Current: current, Next: next}
				rec.Sign(otherPriv)
				s.mockJobClient.EXPECT().ControllerKey(gomock.Any()).Return(rec, true, nil)
			},
			validateFunc: func(m *pki.Manager, _ ed25519.PublicKey) {
				s.Nil(m.NextControllerPublicKey())
			},
		},
		{
			name: "no record changes nothing",
			setup: func(_ *agent.Agent, _ *pki.Manager, _ ed25519.PublicKey, _ ed25519.PrivateKey) {
				s.mockJobClient.EXPECT().ControllerKey(gomock.Any()).
					Return(pki.ControllerKeyRecord{}, false, nil)
			},
			validateFunc: func(m *pki.Manager, current ed25519.PublicKey) {
				s.Equal(current, m.ControllerPublicKey())
			},
		},
		{
			name: "a record that does not read changes nothing",
			setup: func(_ *agent.Agent, _ *pki.Manager, _ ed25519.PublicKey, _ ed25519.PrivateKey) {
				s.mockJobClient.EXPECT().ControllerKey(gomock.Any()).
					Return(pki.ControllerKeyRecord{}, false, errors.New("timeout"))
			},
			validateFunc: func(m *pki.Manager, current ed25519.PublicKey) {
				s.Equal(current, m.ControllerPublicKey())
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()
			a := s.buildAgent()
			m := pki.New(s.appFs, keyDir, "agent")
			current, priv := s.controllerKeyPair()
			m.SetControllerPublicKey(current)
			agent.SetAgentPKIManager(a, m)
			tc.setup(a, m, current, priv)
			agent.ExportFollowControllerKey(context.Background(), a)
			tc.validateFunc(m, current)
		})
	}

	s.Run("an agent that is not admitted reads no record", func() {
		s.SetupTest()
		a := s.buildAgent()
		agent.SetAgentPKIManager(a, pki.New(s.appFs, keyDir, "agent"))
		agent.ExportFollowControllerKey(context.Background(), a)
	})
}

func (s *AgentPublicTestSuite) TestLoadNextControllerKey() {
	nextPath := filepath.Join(keyDir, agent.EnrolledControllerNextKey)

	tests := []struct {
		name         string
		setup        func(a *agent.Agent) ed25519.PublicKey
		wantErr      string
		validateFunc func(m *pki.Manager, want ed25519.PublicKey)
	}{
		{
			name: "the next key that an earlier run kept",
			setup: func(_ *agent.Agent) ed25519.PublicKey {
				next, _ := s.controllerKeyPair()
				s.Require().NoError(avfs.WriteFile(s.appFs, nextPath, keyFile(next), 0o644))

				return next
			},
			validateFunc: func(m *pki.Manager, want ed25519.PublicKey) {
				s.Equal(want, m.NextControllerPublicKey())
			},
		},
		{
			name: "no next key is none",
			setup: func(_ *agent.Agent) ed25519.PublicKey {
				return nil
			},
			validateFunc: func(m *pki.Manager, _ ed25519.PublicKey) {
				s.Nil(m.NextControllerPublicKey())
			},
		},
		{
			name: "a next key that does not parse",
			setup: func(_ *agent.Agent) ed25519.PublicKey {
				s.Require().NoError(avfs.WriteFile(s.appFs, nextPath, []byte("x"), 0o644))

				return nil
			},
			wantErr: "parse " + agent.EnrolledControllerNextKey,
		},
		{
			name: "a next key that does not read",
			setup: func(a *agent.Agent) ed25519.PublicKey {
				agent.SetAppFs(a, failingOn(s.appFs, avfs.FnOpenFile))

				return nil
			},
			wantErr: "read " + agent.EnrolledControllerNextKey,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()
			s.Require().NoError(s.appFs.MkdirAll(keyDir, 0o700))
			a := s.buildAgent()
			m := pki.New(s.appFs, keyDir, "agent")
			agent.SetAgentPKIManager(a, m)
			want := tc.setup(a)
			err := agent.ExportLoadNextControllerKey(a, keyDir)
			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			tc.validateFunc(m, want)
		})
	}
}

func (s *AgentPublicTestSuite) TestControllerKeys() {
	tests := []struct {
		name     string
		setup    func(a *agent.Agent)
		wantKey  bool
		wantNext bool
	}{
		{name: "with no keys, none", setup: func(*agent.Agent) {}},
		{
			name: "an agent that is not admitted has none",
			setup: func(a *agent.Agent) {
				agent.SetAgentPKIManager(a, pki.New(s.appFs, keyDir, "agent"))
			},
		},
		{
			name: "the controller key",
			setup: func(a *agent.Agent) {
				m := pki.New(s.appFs, keyDir, "agent")
				current, _ := s.controllerKeyPair()
				m.SetControllerPublicKey(current)
				agent.SetAgentPKIManager(a, m)
			},
			wantKey: true,
		},
		{
			name: "the controller key and the next one",
			setup: func(a *agent.Agent) {
				m := pki.New(s.appFs, keyDir, "agent")
				current, _ := s.controllerKeyPair()
				next, _ := s.controllerKeyPair()
				m.SetControllerPublicKey(current)
				m.SetNextControllerPublicKey(next)
				agent.SetAgentPKIManager(a, m)
			},
			wantKey:  true,
			wantNext: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()
			a := s.buildAgent()
			agent.SetAgentPKIManager(a, nil)
			tc.setup(a)
			key, next := agent.ExportControllerKeys(a)
			s.Equal(tc.wantKey, key != "")
			s.Equal(tc.wantNext, next != "")
		})
	}
}
