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
	"crypto/ed25519"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/controller/admission"
	admissionMocks "github.com/osapi-io/osapi/internal/controller/admission/mocks"
)

type RotationPublicTestSuite struct {
	suite.Suite

	mockCtrl *gomock.Controller
	mockNC   *admissionMocks.MockNATSSubscriber
	mockKV   *admissionMocks.MockKVStore
	mockPKI  *admissionMocks.MockPKIProvider
	watcher  *admission.Watcher
	pubKey   ed25519.PublicKey
}

func (s *RotationPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockNC = admissionMocks.NewMockNATSSubscriber(s.mockCtrl)
	s.mockKV = admissionMocks.NewMockKVStore(s.mockCtrl)
	s.mockPKI = admissionMocks.NewMockPKIProvider(s.mockCtrl)
	s.pubKey = make(ed25519.PublicKey, ed25519.PublicKeySize)

	s.watcher = admission.NewWatcher(
		slog.Default(),
		s.mockNC,
		s.mockKV, "enrollment",
		s.mockPKI,
		false,
		"osapi",
	)
}

func (s *RotationPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *RotationPublicTestSuite) TearDownSubTest() {
	admission.ResetMarshalFn()
}

func (s *RotationPublicTestSuite) TestRotateControllerKey() {
	tests := []struct {
		name         string
		setupMock    func()
		validateFunc func(error)
	}{
		{
			name: "when rotation succeeds publishes new key",
			setupMock: func() {
				s.mockPKI.EXPECT().PublicKey().Return(s.pubKey)
				s.mockNC.EXPECT().
					PublishCore(gomock.Any(), "osapi.pki.rotate", gomock.Any()).
					Return(nil)
			},
			validateFunc: func(err error) {
				require.NoError(s.T(), err)
			},
		},
		{
			name: "when marshal fails returns error",
			setupMock: func() {
				s.mockPKI.EXPECT().PublicKey().Return(s.pubKey)
				admission.SetMarshalFn(func(_ any) ([]byte, error) {
					return nil, errors.New("marshal error")
				})
			},
			validateFunc: func(err error) {
				require.Error(s.T(), err)
				assert.Contains(s.T(), err.Error(), "marshal rotation message")
			},
		},
		{
			name: "when publish fails returns error",
			setupMock: func() {
				s.mockPKI.EXPECT().PublicKey().Return(s.pubKey)
				s.mockNC.EXPECT().
					PublishCore(gomock.Any(), "osapi.pki.rotate", gomock.Any()).
					Return(errors.New("publish error"))
			},
			validateFunc: func(err error) {
				require.Error(s.T(), err)
				assert.Contains(s.T(), err.Error(), "publish key rotation")
			},
		},
		{
			name: "when namespace is empty uses bare subject",
			setupMock: func() {
				s.watcher = admission.NewWatcher(
					slog.Default(),
					s.mockNC,
					s.mockKV, "enrollment",
					s.mockPKI,
					false,
					"",
				)
				s.mockPKI.EXPECT().PublicKey().Return(s.pubKey)
				s.mockNC.EXPECT().
					PublishCore(gomock.Any(), "pki.rotate", gomock.Any()).
					Return(nil)
			},
			validateFunc: func(err error) {
				require.NoError(s.T(), err)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setupMock()

			tc.validateFunc(s.watcher.RotateControllerKey())
		})
	}
}

func TestRotationPublicTestSuite(
	t *testing.T,
) {
	t.Parallel()
	suite.Run(t, new(RotationPublicTestSuite))
}
