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
	"errors"
	"testing"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/agent"
	agentMocks "github.com/osapi-io/osapi/internal/agent/mocks"
	"github.com/osapi-io/osapi/internal/behavior"
)

type ObjectsPublicTestSuite struct {
	suite.Suite
}

func (s *ObjectsPublicTestSuite) TestObject() {
	tests := []struct {
		name      string
		get       error
		info      *jetstream.ObjectInfo
		infoErr   error
		wantType  types.ContentType
		wantErr   error
		wantError string
	}{
		{
			name: "when the upload gave a content type returns it",
			info: &jetstream.ObjectInfo{ObjectMeta: jetstream.ObjectMeta{
				Headers: nats.Header{"Osapi-Content-Type": []string{"template"}},
			}},
			wantType: types.ContentTemplate,
		},
		{
			name:     "when the object has no headers is raw",
			info:     &jetstream.ObjectInfo{},
			wantType: types.ContentRaw,
		},
		{
			name:     "when the info does not read is raw",
			infoErr:  errors.New("boom"),
			wantType: types.ContentRaw,
		},
		{
			name:    "when the store does not have the object is not found",
			get:     jetstream.ErrObjectNotFound,
			wantErr: behavior.ErrNotFound,
		},
		{
			name:      "when the store fails returns its error",
			get:       errors.New("store down"),
			wantError: "store down",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			store := agentMocks.NewMockObjectReader(gomock.NewController(s.T()))
			store.EXPECT().ObjectGetBytes(gomock.Any(), "b", "a").Return([]byte("x"), tc.get)
			if tc.get == nil {
				store.EXPECT().ObjectInfo(gomock.Any(), "b", "a").Return(tc.info, tc.infoErr)
			}

			content, contentType, err := agent.ExportStoreObjects(store, "b").
				Object(context.Background(), "a")

			switch {
			case tc.wantErr != nil:
				s.ErrorIs(err, tc.wantErr)
			case tc.wantError != "":
				s.ErrorContains(err, tc.wantError)
			default:
				s.Require().NoError(err)
				s.Equal("x", string(content))
				s.Equal(tc.wantType, contentType)
			}
		})
	}
}

func TestObjectsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ObjectsPublicTestSuite))
}
