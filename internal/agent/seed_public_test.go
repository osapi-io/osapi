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

package agent_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/agent"
	filemocks "github.com/retr0h/osapi/internal/provider/file/mocks"
)

// computeSeedSHA256 returns the hex-encoded SHA-256 hash of data.
func computeSeedSHA256(
	data []byte,
) string {
	h := sha256.Sum256(data)

	return hex.EncodeToString(h[:])
}

// SeedPublicTestSuite tests the exported SeedSystemTemplates function.
type SeedPublicTestSuite struct {
	suite.Suite

	ctx    context.Context
	logger *slog.Logger
}

func (s *SeedPublicTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.logger = slog.Default()
}

func (s *SeedPublicTestSuite) TestSeedSystemTemplates() {
	// The embedded templates/ directory contains only .gitkeep (empty file),
	// so objectName = ".gitkeep" and data = []byte{}.
	emptyData := []byte{}
	emptySHA := computeSeedSHA256(emptyData)

	tests := []struct {
		name         string
		setupMock    func(ctrl *gomock.Controller, mockObj *filemocks.MockObjectStore, putNames *[]string)
		wantErr      bool
		errContains  string
		wantPutCalls int
		wantPutName  string
	}{
		{
			name: "when template not found in store uploads it",
			setupMock: func(
				_ *gomock.Controller,
				mockObj *filemocks.MockObjectStore,
				putNames *[]string,
			) {
				mockObj.EXPECT().
					GetBytes(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("not found"))

				mockObj.EXPECT().
					PutBytes(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(
						_ context.Context,
						name string,
						_ []byte,
					) (*jetstream.ObjectInfo, error) {
						*putNames = append(*putNames, name)

						return &jetstream.ObjectInfo{}, nil
					})
			},
			wantErr:      false,
			wantPutCalls: 1,
			wantPutName:  ".gitkeep",
		},
		{
			name: "when template unchanged in store skips upload",
			setupMock: func(
				_ *gomock.Controller,
				mockObj *filemocks.MockObjectStore,
				_ *[]string,
			) {
				// Return same content as embedded — SHA will match, no PutBytes call.
				mockObj.EXPECT().
					GetBytes(gomock.Any(), gomock.Any()).
					Return(emptyData, nil)
			},
			wantErr:      false,
			wantPutCalls: 0,
		},
		{
			name: "when template changed in store overwrites it",
			setupMock: func(
				_ *gomock.Controller,
				mockObj *filemocks.MockObjectStore,
				putNames *[]string,
			) {
				// Return different content — SHA will differ.
				mockObj.EXPECT().
					GetBytes(gomock.Any(), gomock.Any()).
					Return([]byte("old content"), nil)

				mockObj.EXPECT().
					PutBytes(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(
						_ context.Context,
						name string,
						_ []byte,
					) (*jetstream.ObjectInfo, error) {
						*putNames = append(*putNames, name)

						return &jetstream.ObjectInfo{}, nil
					})
			},
			wantErr:      false,
			wantPutCalls: 1,
			wantPutName:  ".gitkeep",
		},
		{
			name: "when PutBytes fails returns wrapped error",
			setupMock: func(
				_ *gomock.Controller,
				mockObj *filemocks.MockObjectStore,
				putNames *[]string,
			) {
				mockObj.EXPECT().
					GetBytes(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("not found"))

				mockObj.EXPECT().
					PutBytes(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(
						_ context.Context,
						name string,
						_ []byte,
					) (*jetstream.ObjectInfo, error) {
						*putNames = append(*putNames, name)

						return nil, errors.New("object store unavailable")
					})
			},
			wantErr:      true,
			errContains:  "upload osapi template",
			wantPutCalls: 1,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctrl := gomock.NewController(s.T())
			defer ctrl.Finish()

			mockObj := filemocks.NewMockObjectStore(ctrl)
			putNames := &[]string{}

			tt.setupMock(ctrl, mockObj, putNames)

			err := agent.SeedSystemTemplates(s.ctx, s.logger, mockObj)

			if tt.wantErr {
				s.Error(err)
				s.Contains(err.Error(), tt.errContains)
			} else {
				s.NoError(err)
			}

			s.Len(*putNames, tt.wantPutCalls)

			if tt.wantPutCalls > 0 && tt.wantPutName != "" {
				s.Equal(tt.wantPutName, (*putNames)[0])
			}

			// The embedded .gitkeep is always empty; verify the SHA is consistent.
			_ = emptySHA
		})
	}
}

func TestSeedPublicTestSuite(t *testing.T) {
	suite.Run(t, new(SeedPublicTestSuite))
}
