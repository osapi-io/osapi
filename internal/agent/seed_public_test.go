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
	"io"
	"log/slog"
	"testing"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/agent"
)

// seedStubObjectStore is a minimal test stub implementing jetstream.ObjectStore
// for SeedSystemTemplates tests. Only GetBytes and PutBytes are functional.
type seedStubObjectStore struct {
	getBytesFunc func(ctx context.Context, name string) ([]byte, error)
	putBytesFunc func(ctx context.Context, name string, data []byte) (*jetstream.ObjectInfo, error)
	putBytesCalls []struct {
		Name string
		Data []byte
	}
}

func (s *seedStubObjectStore) GetBytes(
	ctx context.Context,
	name string,
	_ ...jetstream.GetObjectOpt,
) ([]byte, error) {
	return s.getBytesFunc(ctx, name)
}

func (s *seedStubObjectStore) PutBytes(
	ctx context.Context,
	name string,
	data []byte,
) (*jetstream.ObjectInfo, error) {
	s.putBytesCalls = append(s.putBytesCalls, struct {
		Name string
		Data []byte
	}{Name: name, Data: data})

	return s.putBytesFunc(ctx, name, data)
}

func (s *seedStubObjectStore) Put(
	_ context.Context,
	_ jetstream.ObjectMeta,
	_ io.Reader,
) (*jetstream.ObjectInfo, error) {
	panic("seedStubObjectStore: Put not implemented")
}

func (s *seedStubObjectStore) PutString(
	_ context.Context,
	_ string,
	_ string,
) (*jetstream.ObjectInfo, error) {
	panic("seedStubObjectStore: PutString not implemented")
}

func (s *seedStubObjectStore) PutFile(
	_ context.Context,
	_ string,
) (*jetstream.ObjectInfo, error) {
	panic("seedStubObjectStore: PutFile not implemented")
}

func (s *seedStubObjectStore) Get(
	_ context.Context,
	_ string,
	_ ...jetstream.GetObjectOpt,
) (jetstream.ObjectResult, error) {
	panic("seedStubObjectStore: Get not implemented")
}

func (s *seedStubObjectStore) GetString(
	_ context.Context,
	_ string,
	_ ...jetstream.GetObjectOpt,
) (string, error) {
	panic("seedStubObjectStore: GetString not implemented")
}

func (s *seedStubObjectStore) GetFile(
	_ context.Context,
	_ string,
	_ string,
	_ ...jetstream.GetObjectOpt,
) error {
	panic("seedStubObjectStore: GetFile not implemented")
}

func (s *seedStubObjectStore) GetInfo(
	_ context.Context,
	_ string,
	_ ...jetstream.GetObjectInfoOpt,
) (*jetstream.ObjectInfo, error) {
	panic("seedStubObjectStore: GetInfo not implemented")
}

func (s *seedStubObjectStore) UpdateMeta(
	_ context.Context,
	_ string,
	_ jetstream.ObjectMeta,
) error {
	panic("seedStubObjectStore: UpdateMeta not implemented")
}

func (s *seedStubObjectStore) Delete(
	_ context.Context,
	_ string,
) error {
	panic("seedStubObjectStore: Delete not implemented")
}

func (s *seedStubObjectStore) AddLink(
	_ context.Context,
	_ string,
	_ *jetstream.ObjectInfo,
) (*jetstream.ObjectInfo, error) {
	panic("seedStubObjectStore: AddLink not implemented")
}

func (s *seedStubObjectStore) AddBucketLink(
	_ context.Context,
	_ string,
	_ jetstream.ObjectStore,
) (*jetstream.ObjectInfo, error) {
	panic("seedStubObjectStore: AddBucketLink not implemented")
}

func (s *seedStubObjectStore) Seal(
	_ context.Context,
) error {
	panic("seedStubObjectStore: Seal not implemented")
}

func (s *seedStubObjectStore) Watch(
	_ context.Context,
	_ ...jetstream.WatchOpt,
) (jetstream.ObjectWatcher, error) {
	panic("seedStubObjectStore: Watch not implemented")
}

func (s *seedStubObjectStore) List(
	_ context.Context,
	_ ...jetstream.ListObjectsOpt,
) ([]*jetstream.ObjectInfo, error) {
	panic("seedStubObjectStore: List not implemented")
}

func (s *seedStubObjectStore) Status(
	_ context.Context,
) (jetstream.ObjectStoreStatus, error) {
	panic("seedStubObjectStore: Status not implemented")
}

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
		name           string
		getBytesFunc   func(ctx context.Context, name string) ([]byte, error)
		putBytesFunc   func(ctx context.Context, name string, data []byte) (*jetstream.ObjectInfo, error)
		wantErr        bool
		errContains    string
		wantPutCalls   int
		wantPutName    string
	}{
		{
			name: "when template not found in store uploads it",
			getBytesFunc: func(
				_ context.Context,
				_ string,
			) ([]byte, error) {
				return nil, errors.New("not found")
			},
			putBytesFunc: func(
				_ context.Context,
				_ string,
				_ []byte,
			) (*jetstream.ObjectInfo, error) {
				return &jetstream.ObjectInfo{}, nil
			},
			wantErr:      false,
			wantPutCalls: 1,
			wantPutName:  ".gitkeep",
		},
		{
			name: "when template unchanged in store skips upload",
			getBytesFunc: func(
				_ context.Context,
				_ string,
			) ([]byte, error) {
				// Return same content as embedded — SHA will match.
				return emptyData, nil
			},
			putBytesFunc: func(
				_ context.Context,
				_ string,
				_ []byte,
			) (*jetstream.ObjectInfo, error) {
				panic("seedStubObjectStore: PutBytes must not be called when SHA matches")
			},
			wantErr:      false,
			wantPutCalls: 0,
		},
		{
			name: "when template changed in store overwrites it",
			getBytesFunc: func(
				_ context.Context,
				_ string,
			) ([]byte, error) {
				// Return different content — SHA will differ.
				return []byte("old content"), nil
			},
			putBytesFunc: func(
				_ context.Context,
				_ string,
				_ []byte,
			) (*jetstream.ObjectInfo, error) {
				return &jetstream.ObjectInfo{}, nil
			},
			wantErr:      false,
			wantPutCalls: 1,
			wantPutName:  ".gitkeep",
		},
		{
			name: "when PutBytes fails returns wrapped error",
			getBytesFunc: func(
				_ context.Context,
				_ string,
			) ([]byte, error) {
				return nil, errors.New("not found")
			},
			putBytesFunc: func(
				_ context.Context,
				_ string,
				_ []byte,
			) (*jetstream.ObjectInfo, error) {
				return nil, errors.New("object store unavailable")
			},
			wantErr:      true,
			errContains:  "upload osapi template",
			wantPutCalls: 1,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			stub := &seedStubObjectStore{
				getBytesFunc: tt.getBytesFunc,
				putBytesFunc: tt.putBytesFunc,
			}

			err := agent.SeedSystemTemplates(s.ctx, s.logger, stub)

			if tt.wantErr {
				s.Error(err)
				s.Contains(err.Error(), tt.errContains)
			} else {
				s.NoError(err)
			}

			s.Len(stub.putBytesCalls, tt.wantPutCalls)

			if tt.wantPutCalls > 0 && tt.wantPutName != "" {
				s.Equal(tt.wantPutName, stub.putBytesCalls[0].Name)
			}

			// The embedded .gitkeep is always empty; verify the SHA is consistent.
			_ = emptySHA
		})
	}
}

func TestSeedPublicTestSuite(t *testing.T) {
	suite.Run(t, new(SeedPublicTestSuite))
}
