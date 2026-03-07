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

package file

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DeployTestSuite struct {
	suite.Suite

	logger *slog.Logger
	ctx    context.Context
}

func (suite *DeployTestSuite) SetupTest() {
	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	suite.ctx = context.Background()
}

func (suite *DeployTestSuite) TearDownTest() {
	marshalJSON = json.Marshal
}

func (suite *DeployTestSuite) TestDeploy() {
	fileContent := []byte("server { listen 80; }")

	tests := []struct {
		name       string
		setupFunc  func()
		setupStubs func() (jetstream.ObjectStore, jetstream.KeyValue)
		req        DeployRequest
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "when marshal state fails returns error",
			setupFunc: func() {
				marshalJSON = func(_ interface{}) ([]byte, error) {
					return nil, fmt.Errorf("marshal failure")
				}
			},
			setupStubs: func() (jetstream.ObjectStore, jetstream.KeyValue) {
				obj := &stubObjStoreInternal{getBytesData: fileContent}
				kv := &stubKVInternal{getErr: assert.AnError}
				return obj, kv
			},
			req: DeployRequest{
				ObjectName:  "nginx.conf",
				Path:        "/etc/nginx/nginx.conf",
				ContentType: "raw",
			},
			wantErr:    true,
			wantErrMsg: "failed to marshal file state",
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			if tc.setupFunc != nil {
				tc.setupFunc()
			}

			objStore, stateKV := tc.setupStubs()

			provider := New(
				suite.logger,
				afero.NewMemMapFs(),
				objStore,
				stateKV,
				"test-host",
			)

			got, err := provider.Deploy(suite.ctx, tc.req)

			if tc.wantErr {
				suite.Error(err)
				suite.ErrorContains(err, tc.wantErrMsg)
				suite.Nil(got)
			} else {
				suite.NoError(err)
				suite.Require().NotNil(got)
			}
		})
	}
}

func TestDeployTestSuite(t *testing.T) {
	suite.Run(t, new(DeployTestSuite))
}

// stubObjStoreInternal embeds jetstream.ObjectStore to satisfy the interface.
// Only GetBytes is implemented; other methods panic if called.
type stubObjStoreInternal struct {
	jetstream.ObjectStore
	getBytesData []byte
}

func (s *stubObjStoreInternal) GetBytes(
	_ context.Context,
	_ string,
	_ ...jetstream.GetObjectOpt,
) ([]byte, error) {
	return s.getBytesData, nil
}

// stubKVInternal embeds jetstream.KeyValue to satisfy the interface.
// Only Get is implemented; other methods panic if called.
type stubKVInternal struct {
	jetstream.KeyValue
	getErr error
}

func (s *stubKVInternal) Get(
	_ context.Context,
	_ string,
) (jetstream.KeyValueEntry, error) {
	return nil, s.getErr
}
