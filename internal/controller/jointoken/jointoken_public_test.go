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

package jointoken_test

import (
	"context"
	"crypto/rand"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/controller/jointoken"
)

// entry is a value of the bucket and its revision.
type entry struct {
	jetstream.KeyValueEntry
	value []byte
	rev   uint64
}

func (e entry) Value() []byte    { return e.value }
func (e entry) Revision() uint64 { return e.rev }

// bucket is the enrollment bucket in memory, with an error to return for
// each call that a test sets.
type bucket struct {
	data                                          map[string]entry
	getErr, createErr, updateErr, delErr, keysErr error
}

func (b *bucket) KVGet(_ context.Context, _ string, key string) (jetstream.KeyValueEntry, error) {
	if b.getErr != nil {
		return nil, b.getErr
	}
	e, ok := b.data[key]
	if !ok {
		return nil, jetstream.ErrKeyNotFound
	}

	return e, nil
}

func (b *bucket) KVCreate(_ context.Context, _ string, key string, value []byte) (uint64, error) {
	if b.createErr != nil {
		return 0, b.createErr
	}
	b.data[key] = entry{value: value, rev: 1}

	return 1, nil
}

func (b *bucket) KVUpdate(
	_ context.Context,
	_ string,
	key string,
	value []byte,
	rev uint64,
) (uint64, error) {
	if b.updateErr != nil {
		return 0, b.updateErr
	}
	if b.data[key].rev != rev {
		return 0, errors.New("wrong last sequence")
	}
	b.data[key] = entry{value: value, rev: rev + 1}

	return rev + 1, nil
}

func (b *bucket) KVDelete(_ context.Context, _ string, key string) error {
	if b.delErr != nil {
		return b.delErr
	}
	delete(b.data, key)

	return nil
}

func (b *bucket) KVKeys(_ context.Context, _ string, _ ...string) ([]string, error) {
	if b.keysErr != nil {
		return nil, b.keysErr
	}
	if len(b.data) == 0 {
		return nil, jetstream.ErrNoKeysFound
	}
	out := []string{}
	for k := range b.data {
		out = append(out, k)
	}

	return out, nil
}

type JoinTokenPublicTestSuite struct {
	suite.Suite

	kv    *bucket
	now   time.Time
	store jointoken.Store
}

func (s *JoinTokenPublicTestSuite) SetupTest() {
	s.kv = &bucket{data: map[string]entry{}}
	s.now = time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)
	s.store = jointoken.Store{
		KV:     s.kv,
		Bucket: "enrollment",
		Now:    func() time.Time { return s.now },
	}
}

func (s *JoinTokenPublicTestSuite) TearDownTest() {
	jointoken.SetRandRead(rand.Read)
}

func (s *JoinTokenPublicTestSuite) TestCreate() {
	token, t, err := s.store.Create(
		context.Background(),
		time.Hour,
		2,
		map[string]string{"group": "web"},
	)
	s.Require().NoError(err)
	s.True(strings.HasPrefix(token, "osapi_jt_"+t.ID+"."))
	s.NotContains(string(s.kv.data["jointoken."+t.ID].value), strings.Split(token, ".")[1],
		"the bucket has the hash, not the token")
	s.Equal(s.now.Add(time.Hour), t.ExpiresAt)
	s.Equal(2, t.UsesLeft)

	s.kv.createErr = errors.New("kv down")
	_, _, err = s.store.Create(context.Background(), time.Hour, 1, nil)
	s.ErrorContains(err, "store the join token")

	for _, fail := range []int{1, 2} {
		n := 0
		jointoken.SetRandRead(func(b []byte) (int, error) {
			n++
			if n == fail {
				return 0, errors.New("no entropy")
			}

			return rand.Read(b)
		})
		_, _, err = s.store.Create(context.Background(), time.Hour, 1, nil)
		s.ErrorContains(err, "make a join token")
	}
}

func (s *JoinTokenPublicTestSuite) TestConsume() {
	ctx := context.Background()
	token, _, err := s.store.Create(ctx, time.Hour, 1, nil)
	s.Require().NoError(err)

	tests := []struct {
		name    string
		token   string
		setup   func()
		wantErr error
		errText string
	}{
		{
			name:    "when the token is not a join token, it is invalid",
			token:   "nope",
			wantErr: jointoken.ErrInvalid,
		},
		{
			name:    "when the token has no secret, it is invalid",
			token:   "osapi_jt_abc",
			wantErr: jointoken.ErrInvalid,
		},
		{
			name:    "when the secret is not base64, it is invalid",
			token:   "osapi_jt_abc.***",
			wantErr: jointoken.ErrInvalid,
		},
		{
			name:    "when the controller does not have it, it is invalid",
			token:   "osapi_jt_00000000.c2VjcmV0",
			wantErr: jointoken.ErrInvalid,
		},
		{
			name:    "when the secret is wrong, it is invalid",
			token:   strings.Split(token, ".")[0] + ".d3Jvbmc",
			wantErr: jointoken.ErrInvalid,
		},
		{
			name: "when the bucket cannot be read, it is an error", token: token,
			setup:   func() { s.kv.getErr = errors.New("kv down") },
			errText: "read the join token",
		},
		{
			name: "when the use cannot be taken, it is an error", token: token,
			setup:   func() { s.kv.updateErr = errors.New("kv down") },
			errText: "take a use of the join token",
		},
		{
			name: "when it has expired, it does not admit", token: token,
			setup:   func() { s.now = s.now.Add(2 * time.Hour) },
			wantErr: jointoken.ErrExpired,
		},
		{name: "when it is valid, it admits and takes a use", token: token},
		{
			name:    "when its uses are gone, it does not admit",
			token:   token,
			wantErr: jointoken.ErrUsed,
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			now := s.now
			defer func() { s.kv.getErr, s.kv.updateErr, s.now = nil, nil, now }()
			if tc.setup != nil {
				tc.setup()
			}
			t, err := s.store.Consume(ctx, tc.token)
			switch {
			case tc.wantErr != nil:
				s.ErrorIs(err, tc.wantErr)
			case tc.errText != "":
				s.ErrorContains(err, tc.errText)
			default:
				s.Require().NoError(err)
				s.Equal(0, t.UsesLeft)
			}
		})
	}

	s.kv.data["jointoken.bad"] = entry{value: []byte("not json"), rev: 1}
	_, err = s.store.Consume(ctx, "osapi_jt_bad.c2VjcmV0")
	s.ErrorContains(err, "read the join token")
}

func (s *JoinTokenPublicTestSuite) TestListAndRevoke() {
	ctx := context.Background()
	got, err := s.store.List(ctx)
	s.Require().NoError(err)
	s.Empty(got, "an empty bucket lists none")

	_, first, err := s.store.Create(ctx, time.Hour, 1, nil)
	s.Require().NoError(err)
	s.now = s.now.Add(time.Minute)
	_, second, err := s.store.Create(ctx, time.Hour, 1, nil)
	s.Require().NoError(err)
	s.kv.data["jointoken.bad"] = entry{value: []byte("not json"), rev: 1}

	got, err = s.store.List(ctx)
	s.Require().NoError(err)
	s.Equal(
		[]string{first.ID, second.ID},
		[]string{got[0].ID, got[1].ID},
		"by when it was made; a bad one is left out",
	)

	s.Require().NoError(s.store.Revoke(ctx, first.ID))
	s.ErrorIs(s.store.Revoke(ctx, first.ID), jointoken.ErrInvalid, "a revoked token is gone")

	s.kv.delErr = errors.New("kv down")
	s.ErrorContains(s.store.Revoke(ctx, second.ID), "revoke the join token")

	s.kv.keysErr = errors.New("kv down")
	_, err = s.store.List(ctx)
	s.ErrorContains(err, "list the join tokens")
}

func (s *JoinTokenPublicTestSuite) TestID() {
	id, ok := jointoken.ID("osapi_jt_t1.c2VjcmV0")
	s.True(ok)
	s.Equal("t1", id)
	_, ok = jointoken.ID("not-a-token")
	s.False(ok)
}

func TestJoinTokenPublicTestSuite(t *testing.T) {
	suite.Run(t, new(JoinTokenPublicTestSuite))
}
