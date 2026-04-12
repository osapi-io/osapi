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

	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/job/client"
)

// mockPKISigner implements client.PKISigner for testing.
type mockPKISigner struct {
	pubKey  ed25519.PublicKey
	privKey ed25519.PrivateKey
	ctrlKey ed25519.PublicKey
}

func newMockPKISigner() *mockPKISigner {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	return &mockPKISigner{
		pubKey:  pub,
		privKey: priv,
	}
}

func (m *mockPKISigner) Sign(
	data []byte,
) []byte {
	return ed25519.Sign(m.privKey, data)
}

func (m *mockPKISigner) Fingerprint() string {
	return "SHA256:test-fingerprint"
}

func (m *mockPKISigner) ControllerPublicKey() ed25519.PublicKey {
	return m.ctrlKey
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
		expectError  bool
		wantContains string
	}{
		{
			name:        "when wrapping valid payload",
			payload:     []byte(`{"id":"test-job","operation":{"type":"node.hostname.get"}}`),
			expectError: false,
		},
		{
			name:        "when wrapping empty payload",
			payload:     []byte{},
			expectError: false,
		},
		{
			name:    "when marshal fails returns error",
			payload: []byte(`{"id":"test"}`),
			setupFn: func() {
				client.SetSigningMarshalFn(func(_ any) ([]byte, error) {
					return nil, errors.New("marshal error")
				})
			},
			expectError:  true,
			wantContains: "marshal signed envelope",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.setupFn != nil {
				tt.setupFn()
			}

			signer := newMockPKISigner()

			result, err := client.ExportWrapInSignedEnvelope(signer, tt.payload)

			if tt.expectError {
				s.Error(err)
				if tt.wantContains != "" {
					s.Contains(err.Error(), tt.wantContains)
				}
				return
			}

			s.NoError(err)

			// Verify the result is a valid SignedEnvelope.
			var envelope job.SignedEnvelope
			s.NoError(json.Unmarshal(result, &envelope))
			s.Equal(tt.payload, envelope.Payload)
			s.NotEmpty(envelope.Signature)
			s.Equal("SHA256:test-fingerprint", envelope.Fingerprint)

			// Verify signature is valid.
			s.True(ed25519.Verify(signer.pubKey, tt.payload, envelope.Signature))
		})
	}
}

func (s *SigningPublicTestSuite) TestUnwrapSignedEnvelope() {
	signer := newMockPKISigner()

	tests := []struct {
		name        string
		setupData   func() []byte
		pubKey      ed25519.PublicKey
		wantPayload []byte
		wantEnv     bool
		expectError bool
		errorMsg    string
	}{
		{
			name: "when valid signed envelope with correct key",
			setupData: func() []byte {
				payload := []byte(`{"id":"test"}`)
				wrapped, _ := client.ExportWrapInSignedEnvelope(signer, payload)
				return wrapped
			},
			pubKey:      signer.pubKey,
			wantPayload: []byte(`{"id":"test"}`),
			wantEnv:     true,
			expectError: false,
		},
		{
			name: "when valid signed envelope with nil key skips verification",
			setupData: func() []byte {
				payload := []byte(`{"id":"test"}`)
				wrapped, _ := client.ExportWrapInSignedEnvelope(signer, payload)
				return wrapped
			},
			pubKey:      nil,
			wantPayload: []byte(`{"id":"test"}`),
			wantEnv:     true,
			expectError: false,
		},
		{
			name: "when valid signed envelope with wrong key fails verification",
			setupData: func() []byte {
				payload := []byte(`{"id":"test"}`)
				wrapped, _ := client.ExportWrapInSignedEnvelope(signer, payload)
				return wrapped
			},
			pubKey: func() ed25519.PublicKey {
				otherPub, _, _ := ed25519.GenerateKey(rand.Reader)
				return otherPub
			}(),
			wantPayload: nil,
			wantEnv:     true,
			expectError: true,
			errorMsg:    "invalid signature",
		},
		{
			name: "when raw JSON passes through as non-envelope",
			setupData: func() []byte {
				return []byte(`{"id":"test","operation":{"type":"node.hostname.get"}}`)
			},
			pubKey:      signer.pubKey,
			wantPayload: []byte(`{"id":"test","operation":{"type":"node.hostname.get"}}`),
			wantEnv:     false,
			expectError: false,
		},
		{
			name: "when invalid JSON passes through as non-envelope",
			setupData: func() []byte {
				return []byte(`not json at all`)
			},
			pubKey:      signer.pubKey,
			wantPayload: []byte(`not json at all`),
			wantEnv:     false,
			expectError: false,
		},
		{
			name: "when envelope-like JSON with empty payload passes through",
			setupData: func() []byte {
				return []byte(`{"payload":"","signature":"","fingerprint":""}`)
			},
			pubKey:      signer.pubKey,
			wantPayload: []byte(`{"payload":"","signature":"","fingerprint":""}`),
			wantEnv:     false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			data := tt.setupData()

			payload, isEnvelope, err := client.ExportUnwrapSignedEnvelope(data, tt.pubKey)

			if tt.expectError {
				s.Error(err)
				if tt.errorMsg != "" {
					s.Contains(err.Error(), tt.errorMsg)
				}
				return
			}

			s.NoError(err)
			s.Equal(tt.wantEnv, isEnvelope)
			s.Equal(tt.wantPayload, payload)
		})
	}
}

func (s *SigningPublicTestSuite) TestRoundTrip() {
	signer := newMockPKISigner()
	originalPayload := []byte(`{"id":"round-trip-test","status":"unprocessed"}`)

	// Wrap
	wrapped, err := client.ExportWrapInSignedEnvelope(signer, originalPayload)
	s.NoError(err)

	// Unwrap with correct key
	unwrapped, isEnvelope, err := client.ExportUnwrapSignedEnvelope(wrapped, signer.pubKey)
	s.NoError(err)
	s.True(isEnvelope)
	s.Equal(originalPayload, unwrapped)
}

func TestSigningPublicTestSuite(t *testing.T) {
	suite.Run(t, new(SigningPublicTestSuite))
}
