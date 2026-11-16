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
