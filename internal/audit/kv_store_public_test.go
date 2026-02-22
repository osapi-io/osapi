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
	"encoding/json"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/audit"
	"github.com/retr0h/osapi/internal/job/mocks"
)

type KVStorePublicTestSuite struct {
	suite.Suite

	ctrl   *gomock.Controller
	mockKV *mocks.MockKeyValue
	store  *audit.KVStore
}

func (s *KVStorePublicTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockKV = mocks.NewMockKeyValue(s.ctrl)
	s.store = audit.NewKVStore(slog.Default(), s.mockKV)
}

func (s *KVStorePublicTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *KVStorePublicTestSuite) newEntry(
	id string,
) audit.Entry {
	return audit.Entry{
		ID:           id,
		Timestamp:    time.Now(),
		User:         "user@example.com",
		Roles:        []string{"admin"},
		Method:       "GET",
		Path:         "/system/hostname",
		SourceIP:     "127.0.0.1",
		ResponseCode: 200,
		DurationMs:   42,
	}
}

func (s *KVStorePublicTestSuite) TestWrite() {
	tests := []struct {
		name      string
		entry     audit.Entry
		setupMock func()
		wantErr   bool
	}{
		{
			name:  "successfully writes entry",
			entry: s.newEntry("entry-1"),
			setupMock: func() {
				s.mockKV.EXPECT().
					Put("entry-1", gomock.Any()).
					Return(uint64(1), nil)
			},
			wantErr: false,
		},
		{
			name:  "returns error when put fails",
			entry: s.newEntry("entry-2"),
			setupMock: func() {
				s.mockKV.EXPECT().
					Put("entry-2", gomock.Any()).
					Return(uint64(0), fmt.Errorf("kv error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()
			err := s.store.Write(context.Background(), tt.entry)
			if tt.wantErr {
				s.Error(err)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *KVStorePublicTestSuite) TestGet() {
	entry := s.newEntry("entry-1")
	data, _ := json.Marshal(entry)

	tests := []struct {
		name      string
		id        string
		setupMock func()
		validate  func(*audit.Entry, error)
	}{
		{
			name: "successfully gets entry",
			id:   "entry-1",
			setupMock: func() {
				mockEntry := mocks.NewMockKeyValueEntry(s.ctrl)
				mockEntry.EXPECT().Value().Return(data)
				s.mockKV.EXPECT().Get("entry-1").Return(mockEntry, nil)
			},
			validate: func(e *audit.Entry, err error) {
				s.NoError(err)
				s.Require().NotNil(e)
				s.Equal("entry-1", e.ID)
				s.Equal("user@example.com", e.User)
			},
		},
		{
			name: "returns error when key not found",
			id:   "missing",
			setupMock: func() {
				s.mockKV.EXPECT().Get("missing").Return(nil, nats.ErrKeyNotFound)
			},
			validate: func(e *audit.Entry, err error) {
				s.Error(err)
				s.Nil(e)
			},
		},
		{
			name: "returns error when unmarshal fails",
			id:   "bad-json",
			setupMock: func() {
				mockEntry := mocks.NewMockKeyValueEntry(s.ctrl)
				mockEntry.EXPECT().Value().Return([]byte("not-json"))
				s.mockKV.EXPECT().Get("bad-json").Return(mockEntry, nil)
			},
			validate: func(e *audit.Entry, err error) {
				s.Error(err)
				s.Nil(e)
				s.Contains(err.Error(), "unmarshal audit entry")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()
			result, err := s.store.Get(context.Background(), tt.id)
			tt.validate(result, err)
		})
	}
}

func (s *KVStorePublicTestSuite) TestList() {
	entry1 := s.newEntry("aaa")
	entry2 := s.newEntry("bbb")
	entry3 := s.newEntry("ccc")
	data1, _ := json.Marshal(entry1)
	data2, _ := json.Marshal(entry2)
	data3, _ := json.Marshal(entry3)

	tests := []struct {
		name      string
		limit     int
		offset    int
		setupMock func()
		validate  func([]audit.Entry, int, error)
	}{
		{
			name:   "returns all entries when within limit",
			limit:  10,
			offset: 0,
			setupMock: func() {
				s.mockKV.EXPECT().Keys().Return([]string{"aaa", "bbb", "ccc"}, nil)
				me1 := mocks.NewMockKeyValueEntry(s.ctrl)
				me1.EXPECT().Value().Return(data3)
				me2 := mocks.NewMockKeyValueEntry(s.ctrl)
				me2.EXPECT().Value().Return(data2)
				me3 := mocks.NewMockKeyValueEntry(s.ctrl)
				me3.EXPECT().Value().Return(data1)
				s.mockKV.EXPECT().Get("ccc").Return(me1, nil)
				s.mockKV.EXPECT().Get("bbb").Return(me2, nil)
				s.mockKV.EXPECT().Get("aaa").Return(me3, nil)
			},
			validate: func(entries []audit.Entry, total int, err error) {
				s.NoError(err)
				s.Equal(3, total)
				s.Len(entries, 3)
				// Sorted descending
				s.Equal("ccc", entries[0].ID)
				s.Equal("bbb", entries[1].ID)
				s.Equal("aaa", entries[2].ID)
			},
		},
		{
			name:   "applies pagination correctly",
			limit:  1,
			offset: 1,
			setupMock: func() {
				s.mockKV.EXPECT().Keys().Return([]string{"aaa", "bbb", "ccc"}, nil)
				me := mocks.NewMockKeyValueEntry(s.ctrl)
				me.EXPECT().Value().Return(data2)
				s.mockKV.EXPECT().Get("bbb").Return(me, nil)
			},
			validate: func(entries []audit.Entry, total int, err error) {
				s.NoError(err)
				s.Equal(3, total)
				s.Len(entries, 1)
				s.Equal("bbb", entries[0].ID)
			},
		},
		{
			name:   "returns empty when offset exceeds total",
			limit:  10,
			offset: 100,
			setupMock: func() {
				s.mockKV.EXPECT().Keys().Return([]string{"aaa"}, nil)
			},
			validate: func(entries []audit.Entry, total int, err error) {
				s.NoError(err)
				s.Equal(1, total)
				s.Empty(entries)
			},
		},
		{
			name:   "returns empty for empty bucket",
			limit:  10,
			offset: 0,
			setupMock: func() {
				s.mockKV.EXPECT().Keys().Return(nil, nats.ErrNoKeysFound)
			},
			validate: func(entries []audit.Entry, total int, err error) {
				s.NoError(err)
				s.Equal(0, total)
				s.Empty(entries)
			},
		},
		{
			name:   "returns error when keys fails",
			limit:  10,
			offset: 0,
			setupMock: func() {
				s.mockKV.EXPECT().Keys().Return(nil, fmt.Errorf("connection error"))
			},
			validate: func(entries []audit.Entry, total int, err error) {
				s.Error(err)
				s.Nil(entries)
				s.Equal(0, total)
			},
		},
		{
			name:   "skips entry when individual get fails",
			limit:  10,
			offset: 0,
			setupMock: func() {
				s.mockKV.EXPECT().Keys().Return([]string{"aaa", "bbb"}, nil)
				me1 := mocks.NewMockKeyValueEntry(s.ctrl)
				me1.EXPECT().Value().Return(data1)
				s.mockKV.EXPECT().Get("bbb").Return(nil, fmt.Errorf("get error"))
				s.mockKV.EXPECT().Get("aaa").Return(me1, nil)
			},
			validate: func(entries []audit.Entry, total int, err error) {
				s.NoError(err)
				s.Equal(2, total)
				s.Len(entries, 1)
				s.Equal("aaa", entries[0].ID)
			},
		},
		{
			name:   "skips entry when unmarshal fails",
			limit:  10,
			offset: 0,
			setupMock: func() {
				s.mockKV.EXPECT().Keys().Return([]string{"aaa", "bbb"}, nil)
				badEntry := mocks.NewMockKeyValueEntry(s.ctrl)
				badEntry.EXPECT().Value().Return([]byte("not-json"))
				goodEntry := mocks.NewMockKeyValueEntry(s.ctrl)
				goodEntry.EXPECT().Value().Return(data1)
				s.mockKV.EXPECT().Get("bbb").Return(badEntry, nil)
				s.mockKV.EXPECT().Get("aaa").Return(goodEntry, nil)
			},
			validate: func(entries []audit.Entry, total int, err error) {
				s.NoError(err)
				s.Equal(2, total)
				s.Len(entries, 1)
				s.Equal("aaa", entries[0].ID)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()
			entries, total, err := s.store.List(context.Background(), tt.limit, tt.offset)
			tt.validate(entries, total, err)
		})
	}
}

func TestKVStorePublicTestSuite(t *testing.T) {
	suite.Run(t, new(KVStorePublicTestSuite))
}
