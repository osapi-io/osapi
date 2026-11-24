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

package client_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/job/client"
	"github.com/osapi-io/osapi/internal/job/mocks"
	"go.uber.org/mock/gomock"
)

// newSigner returns a generated PKISigner mock backed by a real ed25519 key
// pair. Sign delegates to the real implementation, so a signature it produces
// verifies against the public key returned alongside it.
func newSigner(
	ctrl *gomock.Controller,
) (*mocks.MockPKISigner, ed25519.PublicKey) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)

	m := mocks.NewMockPKISigner(ctrl)
	m.EXPECT().
		Sign(gomock.Any()).
		DoAndReturn(func(data []byte) []byte {
			return ed25519.Sign(priv, data)
		}).
		AnyTimes()
	m.EXPECT().Fingerprint().Return("SHA256:test-fingerprint").AnyTimes()
	m.EXPECT().ControllerPublicKey().Return(nil).AnyTimes()

	return m, pub
}

// newSignerWithControllerKey is newSigner with the signer's own public key
// reported as the controller's, for paths that verify controller-signed
// messages against it.
func newSignerWithControllerKey(
	ctrl *gomock.Controller,
) (*mocks.MockPKISigner, ed25519.PublicKey) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)

	m := mocks.NewMockPKISigner(ctrl)
	m.EXPECT().
		Sign(gomock.Any()).
		DoAndReturn(func(data []byte) []byte {
			return ed25519.Sign(priv, data)
		}).
		AnyTimes()
	m.EXPECT().Fingerprint().Return("SHA256:test-fingerprint").AnyTimes()
	m.EXPECT().ControllerPublicKey().Return(pub).AnyTimes()

	return m, pub
}

type SigningPublicTestSuite struct {
	suite.Suite
}

func (s *SigningPublicTestSuite) TearDownSubTest() {
	client.ResetSigningMarshalFn()
}

func (s *SigningPublicTestSuite) TestWrapInSignedEnvelope() {
	tests := []struct {
		name         string
		payload      []byte
		setupFn      func()
		validateFunc func([]byte, error, ed25519.PublicKey)
	}{
		{
			name:    "when wrapping valid payload",
			payload: []byte(`{"id":"test-job","operation":{"type":"node.hostname.get"}}`),
			validateFunc: func(result []byte, err error, pubKey ed25519.PublicKey) {
				s.NoError(err)

				payload := []byte(`{"id":"test-job","operation":{"type":"node.hostname.get"}}`)

				var envelope job.SignedEnvelope
				s.NoError(json.Unmarshal(result, &envelope))
				s.Equal(payload, envelope.Payload)
				s.NotEmpty(envelope.Signature)
				s.Equal("SHA256:test-fingerprint", envelope.Fingerprint)

				s.True(ed25519.Verify(pubKey, payload, envelope.Signature))
			},
		},
		{
			name:    "when wrapping empty payload",
			payload: []byte{},
			validateFunc: func(result []byte, err error, pubKey ed25519.PublicKey) {
				s.NoError(err)

				payload := []byte{}

				var envelope job.SignedEnvelope
				s.NoError(json.Unmarshal(result, &envelope))
				s.Equal(payload, envelope.Payload)
				s.NotEmpty(envelope.Signature)
				s.Equal("SHA256:test-fingerprint", envelope.Fingerprint)

				s.True(ed25519.Verify(pubKey, payload, envelope.Signature))
			},
		},
		{
			name:    "when marshal fails returns error",
			payload: []byte(`{"id":"test"}`),
			setupFn: func() {
				client.SetSigningMarshalFn(func(_ any) ([]byte, error) {
					return nil, errors.New("marshal error")
				})
			},
			validateFunc: func(_ []byte, err error, _ ed25519.PublicKey) {
				s.Error(err)
				s.Contains(err.Error(), "marshal signed envelope")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.setupFn != nil {
				tt.setupFn()
			}

			signer, pubKey := newSigner(gomock.NewController(s.T()))

			result, err := client.ExportWrapInSignedEnvelope(signer, "", tt.payload)

			tt.validateFunc(result, err, pubKey)
		})
	}
}

func (s *SigningPublicTestSuite) TestUnwrapSignedEnvelope() {
	signer, pubKey := newSigner(gomock.NewController(s.T()))
	wrapped, _ := client.ExportWrapInSignedEnvelope(signer, "", []byte(`{"id":"test"}`))
	otherPub, _, _ := ed25519.GenerateKey(rand.Reader)

	tests := []struct {
		name    string
		data    []byte
		pubKey  ed25519.PublicKey
		want    []byte
		wantErr error
		errText string
	}{
		{
			name:   "when the envelope verifies with the key, it is the payload",
			data:   wrapped,
			pubKey: pubKey,
			want:   []byte(`{"id":"test"}`),
		},
		{
			name:    "when there is no key, it is refused",
			data:    wrapped,
			wantErr: client.ErrNoSigner,
		},
		{
			name:    "when the key is a different one, it is refused",
			data:    wrapped,
			pubKey:  otherPub,
			errText: "invalid signature",
		},
		{
			name:    "when the value is plain JSON, it is refused",
			data:    []byte(`{"id":"test","operation":{"type":"node.hostname.get"}}`),
			pubKey:  pubKey,
			wantErr: client.ErrNotSigned,
		},
		{
			name:    "when the value is not JSON, it is refused",
			data:    []byte(`not json at all`),
			pubKey:  pubKey,
			wantErr: client.ErrNotSigned,
		},
		{
			name:    "when the envelope has empty fields, it is refused",
			data:    []byte(`{"payload":"","signature":"","fingerprint":""}`),
			pubKey:  pubKey,
			wantErr: client.ErrNotSigned,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			payload, err := client.ExportUnwrapSignedEnvelope(tt.data, tt.pubKey)

			switch {
			case tt.wantErr != nil:
				s.ErrorIs(err, tt.wantErr)
				s.Nil(payload)
			case tt.errText != "":
				s.ErrorContains(err, tt.errText)
				s.Nil(payload)
			default:
				s.Require().NoError(err)
				s.Equal(tt.want, payload)
			}
		})
	}
}

func (s *SigningPublicTestSuite) TestRoundTrip() {
	signer, pubKey := newSigner(gomock.NewController(s.T()))
	originalPayload := []byte(`{"id":"round-trip-test","status":"unprocessed"}`)

	wrapped, err := client.ExportWrapInSignedEnvelope(signer, "", originalPayload)
	s.NoError(err)

	unwrapped, err := client.ExportUnwrapSignedEnvelope(wrapped, pubKey)
	s.NoError(err)
	s.Equal(originalPayload, unwrapped)
}

func TestSigningPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(SigningPublicTestSuite))
}
