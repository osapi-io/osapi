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

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/memfs"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/job"
)

type UndeployTestSuite struct {
	suite.Suite

	logger *slog.Logger
	ctx    context.Context
}

func (suite *UndeployTestSuite) SetupTest() {
	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	suite.ctx = context.Background()
}

func (suite *UndeployTestSuite) TearDownTest() {
	marshalJSON = json.Marshal
}

func (suite *UndeployTestSuite) TestUndeploy() {
	tests := []struct {
		name       string
		setupFunc  func()
		setupStubs func() (avfs.VFS, jetstream.KeyValue)
		req        UndeployRequest
		want       *UndeployResult
	}{
		{
			name: "when marshal state fails still returns changed",
			setupFunc: func() {
				marshalJSON = func(_ interface{}) ([]byte, error) {
					return nil, fmt.Errorf("marshal failure")
				}
			},
			setupStubs: func() (avfs.VFS, jetstream.KeyValue) {
				appFs := memfs.New()
				_ = appFs.MkdirAll("/etc/cron.d", 0o755)
				_ = appFs.WriteFile("/etc/cron.d/backup", []byte("content"), 0o644)

				stateJSON, _ := json.Marshal(job.FileState{
					ObjectName: "backup-script",
					Path:       "/etc/cron.d/backup",
					SHA256:     "abc123",
					DeployedAt: "2026-03-22T00:00:00Z",
				})

				kv := &stubKVWithEntryInternal{
					entry: &stubKVEntryInternal{value: stateJSON},
				}

				return appFs, kv
			},
			req: UndeployRequest{Path: "/etc/cron.d/backup"},
			want: &UndeployResult{
				Changed: true,
				Path:    "/etc/cron.d/backup",
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			if tc.setupFunc != nil {
				tc.setupFunc()
			}

			appFs, kv := tc.setupStubs()

			provider := New(
				suite.logger,
				appFs,
				&stubObjStoreInternal{},
				kv,
				"test-host",
			)

			got, err := provider.Undeploy(suite.ctx, tc.req)
			suite.NoError(err)
			suite.Require().NotNil(got)
			suite.Equal(tc.want.Changed, got.Changed)
			suite.Equal(tc.want.Path, got.Path)
		})
	}
}

func TestUndeployTestSuite(t *testing.T) {
	suite.Run(t, new(UndeployTestSuite))
}

// stubKVWithEntryInternal is a KV stub that returns a fixed entry.
// Used for internal tests where import cycles prevent using gomock.
type stubKVWithEntryInternal struct {
	jetstream.KeyValue
	entry jetstream.KeyValueEntry
}

func (s *stubKVWithEntryInternal) Get(
	_ context.Context,
	_ string,
) (jetstream.KeyValueEntry, error) {
	return s.entry, nil
}

// stubKVEntryInternal satisfies jetstream.KeyValueEntry.
type stubKVEntryInternal struct {
	jetstream.KeyValueEntry
	value []byte
}

func (e *stubKVEntryInternal) Value() []byte { return e.value }
