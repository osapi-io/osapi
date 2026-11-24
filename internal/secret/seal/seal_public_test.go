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

package seal_test

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"testing"

	"github.com/nats-io/nkeys"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/secret"
	"github.com/osapi-io/osapi/internal/secret/seal"
)

type SealPublicTestSuite struct {
	suite.Suite

	agent nkeys.KeyPair
	other nkeys.KeyPair
}

func (s *SealPublicTestSuite) SetupTest() {
	var err error
	s.agent, err = nkeys.CreateCurveKeys()
	s.Require().NoError(err)
	s.other, err = nkeys.CreateCurveKeys()
	s.Require().NoError(err)
}

func (s *SealPublicTestSuite) TearDownTest() {
	seal.ResetCreateCurveKeys()
}

// sealed is a value sealed for the job jobID, and the job key sealed to
// s.agent.
func (s *SealPublicTestSuite) sealed(
	jobID string,
	value secret.Value,
) (string, string, string) {
	sealer, err := seal.New(jobID)
	s.Require().NoError(err)
	pub, _ := s.agent.PublicKey()
	key, err := sealer.KeyFor(pub)
	s.Require().NoError(err)

	return sealer.Value(value), key, sealer.From()
}

func (s *SealPublicTestSuite) TestRoundTrip() {
	value, key, from := s.sealed("job-1", "hunter2")

	s.True(strings.HasPrefix(value, seal.Prefix))
	s.NotContains(value, "hunter2")
	o, err := seal.Open("job-1", key, from, s.agent)
	s.Require().NoError(err)
	got, err := o.Value(secret.Value(value))
	s.Require().NoError(err)
	s.Equal(secret.Value("hunter2"), got)

	again, _, _ := s.sealed("job-1", "hunter2")
	s.NotEqual(value, again, "each seal has its own nonce")
}

func (s *SealPublicTestSuite) TestOpenRefuses() {
	value, key, from := s.sealed("job-1", "hunter2")
	tests := []struct {
		name    string
		open    func() (*seal.Opener, error)
		value   string
		wantErr string
	}{
		{
			name:    "when the job key is not base64, it does not open",
			open:    func() (*seal.Opener, error) { return seal.Open("job-1", "!", from, s.agent) },
			wantErr: "open the job key",
		},
		{
			name:    "when the key of a different agent opens it, it does not open",
			open:    func() (*seal.Opener, error) { return seal.Open("job-1", key, from, s.other) },
			wantErr: "open the job key",
		},
		{
			name: "when the job key has the wrong size, it does not open",
			open: func() (*seal.Opener, error) {
				fromKP, _ := nkeys.CreateCurveKeys()
				fromPub, _ := fromKP.PublicKey()
				pub, _ := s.agent.PublicKey()
				short, _ := fromKP.Seal([]byte("short"), pub)

				return seal.Open("job-1", encode(short), fromPub, s.agent)
			},
			wantErr: "open the job key",
		},
		{
			name:    "when the value was sealed for a different job, it does not open",
			open:    func() (*seal.Opener, error) { return seal.Open("job-2", key, from, s.agent) },
			value:   value,
			wantErr: "a sealed secret of the job does not open",
		},
		{
			name:    "when the value is not sealed, it is refused",
			open:    func() (*seal.Opener, error) { return seal.Open("job-1", key, from, s.agent) },
			value:   "hunter2",
			wantErr: "a secret of the job is not sealed",
		},
		{
			name:    "when the sealed value is not base64, it is refused",
			open:    func() (*seal.Opener, error) { return seal.Open("job-1", key, from, s.agent) },
			value:   seal.Prefix + "!",
			wantErr: "a sealed secret of the job is not valid",
		},
		{
			name:    "when the sealed value is shorter than a nonce, it is refused",
			open:    func() (*seal.Opener, error) { return seal.Open("job-1", key, from, s.agent) },
			value:   seal.Prefix + "AAAA",
			wantErr: "a sealed secret of the job is not valid",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			o, err := tc.open()
			if tc.value == "" {
				s.ErrorContains(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			got, err := o.Value(secret.Value(tc.value))
			s.EqualError(err, tc.wantErr)
			s.Empty(got)
		})
	}
}

func (s *SealPublicTestSuite) TestNewAndKeyForFail() {
	sealer, err := seal.New("job-1")
	s.Require().NoError(err)
	_, err = sealer.KeyFor("not-a-key")
	s.ErrorContains(err, "seal the job key")

	seal.SetCreateCurveKeys(func() (nkeys.KeyPair, error) { return nil, errors.New("no entropy") })
	_, err = seal.New("job-1")
	s.ErrorContains(err, "no entropy")
}

func (s *SealPublicTestSuite) TestOpenerContext() {
	s.Nil(seal.OpenerFrom(context.Background()))
	_, key, from := s.sealed("job-1", "x")
	o, err := seal.Open("job-1", key, from, s.agent)
	s.Require().NoError(err)
	s.Same(o, seal.OpenerFrom(seal.WithOpener(context.Background(), o)))
}

func TestSealPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(SealPublicTestSuite))
}

// encode is raw in the text form of a sealed key.
func encode(
	raw []byte,
) string {
	return base64.RawURLEncoding.EncodeToString(raw)
}
