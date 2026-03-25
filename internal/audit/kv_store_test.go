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

package audit_test

import (
	"context"
	"fmt"
	"log/slog"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/audit"
	"github.com/retr0h/osapi/internal/job/mocks"
)

type KVStoreMarshalTestSuite struct {
	suite.Suite

	ctrl   *gomock.Controller
	mockKV *mocks.MockKeyValue
	store  *audit.KVStore
}

func (s *KVStoreMarshalTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockKV = mocks.NewMockKeyValue(s.ctrl)
	s.store = audit.NewKVStore(slog.Default(), s.mockKV)
}

func (s *KVStoreMarshalTestSuite) TearDownTest() {
	s.ctrl.Finish()
	audit.ResetMarshalJSON()
}

func (s *KVStoreMarshalTestSuite) TestWriteMarshalError() {
	tests := []struct {
		name         string
		setupMock    func()
		validateFunc func(err error)
	}{
		{
			name: "when marshal fails returns wrapped error",
			setupMock: func() {
				audit.SetMarshalJSON(func(_ interface{}) ([]byte, error) {
					return nil, fmt.Errorf("marshal failure")
				})
			},
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "marshal audit entry")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()
			err := s.store.Write(context.Background(), audit.Entry{ID: "test-id"})
			tt.validateFunc(err)
		})
	}
}

func TestKVStoreMarshalTestSuite(t *testing.T) {
	suite.Run(t, new(KVStoreMarshalTestSuite))
}
