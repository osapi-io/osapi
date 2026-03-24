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

package agent

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/suite"
)

type SeedTestSuite struct {
	suite.Suite

	logger *slog.Logger
	ctx    context.Context
}

func (suite *SeedTestSuite) SetupTest() {
	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	suite.ctx = context.Background()
}

func (suite *SeedTestSuite) TearDownTest() {
	readEmbeddedFile = func(path string) ([]byte, error) {
		return systemTemplates.ReadFile(path)
	}
}

func (suite *SeedTestSuite) TestSeedSystemTemplates() {
	tests := []struct {
		name       string
		setupFunc  func()
		objStore   jetstream.ObjectStore
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "when ReadFile fails returns error",
			setupFunc: func() {
				readEmbeddedFile = func(_ string) ([]byte, error) {
					return nil, fmt.Errorf("read failure")
				}
			},
			objStore:   &seedStubObjStore{},
			wantErr:    true,
			wantErrMsg: "read embedded template",
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			if tc.setupFunc != nil {
				tc.setupFunc()
			}

			err := SeedSystemTemplates(suite.ctx, suite.logger, tc.objStore)

			if tc.wantErr {
				suite.Error(err)
				suite.ErrorContains(err, tc.wantErrMsg)
			} else {
				suite.NoError(err)
			}
		})
	}
}

func TestSeedTestSuite(t *testing.T) {
	suite.Run(t, new(SeedTestSuite))
}

// seedStubObjStore is a minimal stub for internal seed tests.
type seedStubObjStore struct {
	jetstream.ObjectStore
}

func (s *seedStubObjStore) GetBytes(
	_ context.Context,
	_ string,
	_ ...jetstream.GetObjectOpt,
) ([]byte, error) {
	return nil, fmt.Errorf("not found")
}

func (s *seedStubObjStore) PutBytes(
	_ context.Context,
	_ string,
	_ []byte,
) (*jetstream.ObjectInfo, error) {
	return nil, nil
}

func (s *seedStubObjStore) GetInfo(
	_ context.Context,
	_ string,
	_ ...jetstream.GetObjectInfoOpt,
) (*jetstream.ObjectInfo, error) {
	return nil, fmt.Errorf("not found")
}
