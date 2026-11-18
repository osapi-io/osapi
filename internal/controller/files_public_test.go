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

package controller_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/controller"
	"github.com/osapi-io/osapi/internal/controller/mocks"
)

type FilesPublicTestSuite struct {
	suite.Suite
}

func TestFilesPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(FilesPublicTestSuite))
}

func (s *FilesPublicTestSuite) TestOpenFileBuckets() {
	tests := []struct {
		name       string
		objErr     error
		kvErr      error
		wantObject bool
		wantState  bool
	}{
		{
			name:       "when both buckets open returns both",
			wantObject: true,
			wantState:  true,
		},
		{
			name:      "when the Object Store fails runs without it",
			objErr:    errBoom,
			wantState: true,
		},
		{
			name:       "when the file-state bucket fails runs without it",
			kvErr:      errBoom,
			wantObject: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			ctrl := gomock.NewController(s.T())
			nc := mocks.NewMockFileBuckets(ctrl)
			nc.EXPECT().ObjectStoreCreateOrUpdate(gomock.Any(), gomock.Any()).Return(tc.objErr)
			nc.EXPECT().KVStatus(gomock.Any(), "ns-file-state").Return(nil, tc.kvErr)

			objBucket, stateBucket := controller.ExportOpenFileBuckets(
				context.Background(),
				slog.Default(),
				nc,
				"ns",
				natsConfig(),
			)

			s.Equal(tc.wantObject, objBucket != "")
			s.Equal(tc.wantState, stateBucket != "")
		})
	}
}

// TestFileDeps holds the file group's buckets to the names the controller
// opened: each call goes to its bucket, and a bucket that is not open is nil.
func (s *FilesPublicTestSuite) TestFileDeps() {
	ctrl := gomock.NewController(s.T())
	nc := mocks.NewMockFileBuckets(ctrl)
	ctx := context.Background()

	none := controller.ExportFileDeps(nc, "", "")
	s.Nil(none.Objects)
	s.Nil(none.States)

	d := controller.ExportFileDeps(nc, "objects", "state")
	nc.EXPECT().ObjectPut(ctx, "objects", jetstream.ObjectMeta{Name: "a"}, nil).Return(nil, nil)
	nc.EXPECT().ObjectGetBytes(ctx, "objects", "a").Return([]byte("x"), nil)
	nc.EXPECT().ObjectInfo(ctx, "objects", "a").Return(nil, nil)
	nc.EXPECT().ObjectDelete(ctx, "objects", "a").Return(nil)
	nc.EXPECT().ObjectList(ctx, "objects").Return(nil, nil)
	nc.EXPECT().KVKeys(ctx, "state").Return([]string{"k"}, nil)
	nc.EXPECT().KVGet(ctx, "state", "k").Return(nil, nil)

	_, _ = d.Objects.Put(ctx, jetstream.ObjectMeta{Name: "a"}, nil)
	data, _ := d.Objects.GetBytes(ctx, "a")
	s.Equal("x", string(data))
	_, _ = d.Objects.Info(ctx, "a")
	s.NoError(d.Objects.Delete(ctx, "a"))
	_, _ = d.Objects.List(ctx)
	keys, _ := d.States.Keys(ctx)
	s.Equal([]string{"k"}, keys)
	_, _ = d.States.Get(ctx, "k")
}
