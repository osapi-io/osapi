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

package job_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/job"
)

type RegistrationPublicTestSuite struct {
	suite.Suite

	pub  ed25519.PublicKey
	priv ed25519.PrivateKey
}

func (s *RegistrationPublicTestSuite) SetupTest() {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	s.Require().NoError(err)
	s.pub, s.priv = pub, priv
}

// signed returns reg with a valid signature over its canonical bytes.
func (s *RegistrationPublicTestSuite) signed(
	reg job.AgentRegistration,
) job.AgentRegistration {
	s.T().Helper()

	reg.Signature = ed25519.Sign(s.priv, job.RegistrationSigningBytes(&reg))

	return reg
}

func (s *RegistrationPublicTestSuite) base() job.AgentRegistration {
	return job.AgentRegistration{
		MachineID:   "machine-001",
		Hostname:    "web-01",
		Fingerprint: "SHA256:abc",
		State:       "Ready",
		Labels:      map[string]string{"group": "web", "env": "prod"},
	}
}

func (s *RegistrationPublicTestSuite) TestVerifyRegistration() {
	tests := []struct {
		name         string
		build        func() (job.AgentRegistration, ed25519.PublicKey)
		validateFunc func(bool)
	}{
		{
			name: "accepts a registration signed by the stored key",
			build: func() (job.AgentRegistration, ed25519.PublicKey) {
				return s.signed(s.base()), s.pub
			},
			validateFunc: func(ok bool) { s.True(ok) },
		},
		{
			name: "rejects a signature from another key",
			build: func() (job.AgentRegistration, ed25519.PublicKey) {
				other, _, _ := ed25519.GenerateKey(rand.Reader)

				return s.signed(s.base()), other
			},
			validateFunc: func(ok bool) { s.False(ok) },
		},
		{
			name: "rejects an unsigned registration",
			build: func() (job.AgentRegistration, ed25519.PublicKey) {
				return s.base(), s.pub
			},
			validateFunc: func(ok bool) { s.False(ok) },
		},
		{
			name: "rejects when no key is supplied",
			build: func() (job.AgentRegistration, ed25519.PublicKey) {
				return s.signed(s.base()), nil
			},
			validateFunc: func(ok bool) { s.False(ok) },
		},
		{
			name: "rejects a changed hostname",
			build: func() (job.AgentRegistration, ed25519.PublicKey) {
				reg := s.signed(s.base())
				reg.Hostname = "web-02"

				return reg, s.pub
			},
			validateFunc: func(ok bool) { s.False(ok) },
		},
		{
			name: "rejects a changed machine ID",
			build: func() (job.AgentRegistration, ed25519.PublicKey) {
				reg := s.signed(s.base())
				reg.MachineID = "machine-evil"

				return reg, s.pub
			},
			validateFunc: func(ok bool) { s.False(ok) },
		},
		{
			name: "rejects a state lifted out of Pending",
			build: func() (job.AgentRegistration, ed25519.PublicKey) {
				reg := s.base()
				reg.State = "Pending"
				reg = s.signed(reg)
				reg.State = "Ready"

				return reg, s.pub
			},
			validateFunc: func(ok bool) { s.False(ok) },
		},
		{
			name: "rejects an altered label, which is a routing input",
			build: func() (job.AgentRegistration, ed25519.PublicKey) {
				reg := s.signed(s.base())
				reg.Labels = map[string]string{"group": "db", "env": "prod"}

				return reg, s.pub
			},
			validateFunc: func(ok bool) { s.False(ok) },
		},
		{
			name: "rejects an added label",
			build: func() (job.AgentRegistration, ed25519.PublicKey) {
				reg := s.signed(s.base())
				reg.Labels["tier"] = "edge"

				return reg, s.pub
			},
			validateFunc: func(ok bool) { s.False(ok) },
		},
		{
			name: "rejects a changed fingerprint",
			build: func() (job.AgentRegistration, ed25519.PublicKey) {
				reg := s.signed(s.base())
				reg.Fingerprint = "SHA256:someone-else"

				return reg, s.pub
			},
			validateFunc: func(ok bool) { s.False(ok) },
		},
		{
			name: "ignores fields that only report, so a heartbeat stays valid",
			build: func() (job.AgentRegistration, ed25519.PublicKey) {
				reg := s.signed(s.base())
				reg.Uptime = 1234
				reg.AgentVersion = "9.9.9"

				return reg, s.pub
			},
			validateFunc: func(ok bool) { s.True(ok) },
		},
		{
			name: "rejects a nil registration",
			build: func() (job.AgentRegistration, ed25519.PublicKey) {
				return job.AgentRegistration{}, s.pub
			},
			validateFunc: func(ok bool) { s.False(ok) },
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			reg, key := tt.build()

			tt.validateFunc(job.VerifyRegistration(&reg, key))
		})
	}
}

// TestSigningBytesIndependentOfMapOrder is the property both sides depend on:
// the same labels must produce the same bytes however the map is walked.
func (s *RegistrationPublicTestSuite) TestSigningBytesIndependentOfMapOrder() {
	a := s.base()
	b := s.base()
	b.Labels = map[string]string{"env": "prod", "group": "web"}

	s.Equal(
		string(job.RegistrationSigningBytes(&a)),
		string(job.RegistrationSigningBytes(&b)),
	)
}

// TestSigningBytesExcludeSignature proves the signature is carried beside the
// signed bytes rather than inside them, which is what makes verification
// possible at all.
func (s *RegistrationPublicTestSuite) TestSigningBytesExcludeSignature() {
	reg := s.base()
	before := string(job.RegistrationSigningBytes(&reg))

	reg.Signature = []byte("anything at all")

	s.Equal(before, string(job.RegistrationSigningBytes(&reg)))
}

func (s *RegistrationPublicTestSuite) TestSigningBytesNilRegistration() {
	s.Nil(job.RegistrationSigningBytes(nil))
	s.False(job.VerifyRegistration(nil, s.pub))
}

func TestRegistrationPublicTestSuite(
	t *testing.T,
) {
	t.Parallel()
	suite.Run(t, new(RegistrationPublicTestSuite))
}
