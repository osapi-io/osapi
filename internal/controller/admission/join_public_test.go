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

package admission_test

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"errors"

	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/controller/admission"
	admissionMocks "github.com/osapi-io/osapi/internal/controller/admission/mocks"
	"github.com/osapi-io/osapi/internal/controller/jointoken"
	"github.com/osapi-io/osapi/internal/pki"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// newTokens is a store of join tokens that admits the token
// osapi_jt_t1.secret, or that fails each use with err.
func newTokens(
	ctrl *gomock.Controller,
	err error,
) *admissionMocks.MockJoinTokens {
	m := admissionMocks.NewMockJoinTokens(ctrl)
	m.EXPECT().Consume(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, token string) (jointoken.Token, error) {
			if err != nil {
				return jointoken.Token{}, err
			}
			if token != "osapi_jt_t1.secret" {
				return jointoken.Token{}, jointoken.ErrInvalid
			}

			return jointoken.Token{ID: "t1", Labels: map[string]string{"group": "web"}}, nil
		},
	).AnyTimes()

	return m
}

// newCountingTokens admits each token whose ID is t1, and counts each use in
// calls.
func newCountingTokens(
	ctrl *gomock.Controller,
	calls *int,
) *admissionMocks.MockJoinTokens {
	m := admissionMocks.NewMockJoinTokens(ctrl)
	m.EXPECT().Consume(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, token string) (jointoken.Token, error) {
			*calls++
			if id, ok := jointoken.ID(token); !ok || id != "t1" {
				return jointoken.Token{}, jointoken.ErrInvalid
			}

			return jointoken.Token{ID: "t1", Labels: map[string]string{"group": "web"}}, nil
		},
	).AnyTimes()

	return m
}

// joinRequest is a join of m-1 with token, signed by its own key.
func joinRequest(
	token string,
) pki.JoinRequest {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	r := pki.JoinRequest{
		Token:         token,
		MachineID:     "m-1",
		Hostname:      "web-01",
		PublicKey:     pub,
		EncryptionKey: "BH0Fq2u2_BBIGuYw6KRZYg4I1pT5DACIZUCrhDK5Y7Phpl1MRex_y4Y60InurvbYIiZDQ2MHIcmOXF8SfLlA5fs",
	}
	r.Sign(priv)

	return r
}

func (s *StorePublicTestSuite) TestJoin() {
	s.Run("when the token admits, the machine is admitted at once, with the token named", func() {
		s.watcher.SetJoinTokens(newTokens(gomock.NewController(s.T()), nil))
		s.pending("m-1", "web-01")
		s.put("rejected.m-1", admission.Record{MachineID: "m-1"})
		s.put("revoked.m-1", admission.Record{MachineID: "m-1"})
		req := joinRequest("osapi_jt_t1.secret")

		key, err := s.watcher.Join(context.Background(), req)

		s.Require().NoError(err)
		s.Len(key, ed25519.PublicKeySize, "the controller's key, for the agent to keep")
		got, err := s.watcher.Get(context.Background(), "m-1")
		s.Require().NoError(err)
		s.Equal(types.AdmissionAdmitted, got.State)
		s.Equal("t1", got.JoinToken)
		s.Equal(map[string]string{"group": "web"}, got.JoinLabels, "the token's labels are kept")
		s.Equal(pki.FingerprintOfKeys(req.PublicKey, req.EncryptionKey), got.Fingerprint)
		s.NotContains(s.kv.data, admission.KVPrefix()+"m-1")
		s.NotContains(s.kv.data, "rejected.m-1")
		s.NotContains(s.kv.data, "revoked.m-1")
	})

	s.Run(
		"when the reply was lost, the same machine, key and token join again with no use",
		func() {
			calls := 0
			s.watcher.SetJoinTokens(newCountingTokens(gomock.NewController(s.T()), &calls))
			req := joinRequest("osapi_jt_t1.c2VjcmV0")

			_, err := s.watcher.Join(context.Background(), req)
			s.Require().NoError(err)
			got, err := s.watcher.Get(context.Background(), "m-1")
			s.Require().NoError(err)
			s.Equal(map[string]string{"group": "web"}, got.JoinLabels)
			key, err := s.watcher.Join(context.Background(), req)

			s.Require().NoError(err)
			s.Len(key, ed25519.PublicKeySize)
			s.Equal(1, calls, "one use of the token")

			other := joinRequest("osapi_jt_t1.c2VjcmV0")
			_, err = s.watcher.Join(context.Background(), other)
			s.Require().NoError(err)
			s.Equal(2, calls, "a new key takes a use")

			bad := joinRequest("not-a-token")
			_, _ = s.watcher.Join(context.Background(), bad)
			s.Equal(3, calls)
		},
	)

	s.Run("when the request is not signed by its key, nothing is admitted", func() {
		s.watcher.SetJoinTokens(newTokens(gomock.NewController(s.T()), nil))
		req := joinRequest("osapi_jt_t1.secret")
		req.Hostname = "other"

		_, err := s.watcher.Join(context.Background(), req)

		s.ErrorIs(err, admission.ErrJoinSignature)
		s.NotContains(s.kv.data, admission.AcceptedKVPrefix()+"m-1")
	})

	s.Run("when the encryption key is not a P-256 key, nothing is admitted", func() {
		s.watcher.SetJoinTokens(newTokens(gomock.NewController(s.T()), nil))
		pub, priv, _ := ed25519.GenerateKey(rand.Reader)
		req := pki.JoinRequest{
			Token:         "osapi_jt_t1.secret",
			MachineID:     "m-1",
			Hostname:      "web-01",
			PublicKey:     pub,
			EncryptionKey: "XA7NTLWJ2ECILFDLHORNVYROXOK65ONQ2SJYHRNAWRJZ3UXMVXIGUVJM",
		}
		req.Sign(priv)

		_, err := s.watcher.Join(context.Background(), req)

		s.ErrorIs(err, admission.ErrJoinEncryptionKey)
		s.NotContains(s.kv.data, admission.AcceptedKVPrefix()+"m-1")
	})

	s.Run("when the token does not admit, its error comes back and nothing is admitted", func() {
		s.watcher.SetJoinTokens(newTokens(gomock.NewController(s.T()), jointoken.ErrExpired))

		_, err := s.watcher.Join(context.Background(), joinRequest("osapi_jt_t1.secret"))

		s.ErrorIs(err, jointoken.ErrExpired)
		s.NotContains(s.kv.data, admission.AcceptedKVPrefix()+"m-1")
	})

	s.Run("when the controller has no store of tokens, it takes no join", func() {
		_, err := s.watcher.Join(context.Background(), joinRequest("osapi_jt_t1.secret"))

		s.ErrorIs(err, admission.ErrNoJoinTokens)
	})

	s.Run("when the record cannot be written, it is the error", func() {
		s.watcher.SetJoinTokens(newTokens(gomock.NewController(s.T()), nil))
		admission.SetMarshalFn(func(any) ([]byte, error) { return nil, errors.New("no marshal") })

		_, err := s.watcher.Join(context.Background(), joinRequest("osapi_jt_t1.secret"))

		s.ErrorContains(err, "admit m-1")
	})
}
