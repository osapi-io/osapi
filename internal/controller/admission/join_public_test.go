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

	"github.com/osapi-io/osapi/internal/controller/admission"
	"github.com/osapi-io/osapi/internal/controller/jointoken"
	"github.com/osapi-io/osapi/internal/pki"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// fakeTokens is a store of join tokens that admits one token.
type fakeTokens struct {
	err error
}

func (f fakeTokens) Consume(
	_ context.Context,
	token string,
) (jointoken.Token, error) {
	if f.err != nil {
		return jointoken.Token{}, f.err
	}
	if token != "osapi_jt_t1.secret" {
		return jointoken.Token{}, jointoken.ErrInvalid
	}

	return jointoken.Token{ID: "t1"}, nil
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
		EncryptionKey: "XA7NTLWJ2ECILFDLHORNVYROXOK65ONQ2SJYHRNAWRJZ3UXMVXIGUVJM",
	}
	r.Sign(priv)

	return r
}

func (s *StorePublicTestSuite) TestJoin() {
	s.Run("when the token admits, the machine is admitted at once, with the token named", func() {
		s.watcher.SetJoinTokens(fakeTokens{})
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
		s.Equal(pki.FingerprintOfKeys(req.PublicKey, req.EncryptionKey), got.Fingerprint)
		s.NotContains(s.kv.data, admission.KVPrefix()+"m-1")
		s.NotContains(s.kv.data, "rejected.m-1")
		s.NotContains(s.kv.data, "revoked.m-1")
	})

	s.Run("when the request is not signed by its key, nothing is admitted", func() {
		s.watcher.SetJoinTokens(fakeTokens{})
		req := joinRequest("osapi_jt_t1.secret")
		req.Hostname = "other"

		_, err := s.watcher.Join(context.Background(), req)

		s.ErrorIs(err, admission.ErrJoinSignature)
		s.NotContains(s.kv.data, admission.AcceptedKVPrefix()+"m-1")
	})

	s.Run("when the token does not admit, its error comes back and nothing is admitted", func() {
		s.watcher.SetJoinTokens(fakeTokens{err: jointoken.ErrExpired})

		_, err := s.watcher.Join(context.Background(), joinRequest("osapi_jt_t1.secret"))

		s.ErrorIs(err, jointoken.ErrExpired)
		s.NotContains(s.kv.data, admission.AcceptedKVPrefix()+"m-1")
	})

	s.Run("when the controller has no store of tokens, it takes no join", func() {
		_, err := s.watcher.Join(context.Background(), joinRequest("osapi_jt_t1.secret"))

		s.ErrorIs(err, admission.ErrNoJoinTokens)
	})

	s.Run("when the record cannot be written, it is the error", func() {
		s.watcher.SetJoinTokens(fakeTokens{})
		admission.SetMarshalFn(func(any) ([]byte, error) { return nil, errors.New("no marshal") })

		_, err := s.watcher.Join(context.Background(), joinRequest("osapi_jt_t1.secret"))

		s.ErrorContains(err, "admit m-1")
	})
}
