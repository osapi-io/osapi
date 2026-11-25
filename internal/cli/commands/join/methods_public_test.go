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

package join_test

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli/commands/join"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller/api"
	"github.com/osapi-io/osapi/internal/controller/jointoken"
)

// entry is a value of the bucket.
type entry struct {
	jetstream.KeyValueEntry
	value []byte
}

func (e entry) Value() []byte    { return e.value }
func (e entry) Revision() uint64 { return 1 }

// bucket is the enrollment bucket in memory.
type bucket struct {
	data map[string][]byte
	err  error
}

func (b *bucket) KVGet(_ context.Context, _ string, key string) (jetstream.KeyValueEntry, error) {
	v, ok := b.data[key]
	if !ok {
		return nil, jetstream.ErrKeyNotFound
	}

	return entry{value: v}, nil
}

func (b *bucket) KVCreate(_ context.Context, _ string, key string, value []byte) (uint64, error) {
	b.data[key] = value

	return 1, b.err
}

func (b *bucket) KVUpdate(context.Context, string, string, []byte, uint64) (uint64, error) {
	return 2, nil
}

func (b *bucket) KVDelete(_ context.Context, _ string, key string) error {
	delete(b.data, key)

	return nil
}

func (b *bucket) KVKeys(context.Context, string, ...string) ([]string, error) {
	if b.err != nil {
		return nil, b.err
	}
	out := []string{}
	for k := range b.data {
		out = append(out, k)
	}

	return out, nil
}

type MethodsPublicTestSuite struct {
	suite.Suite

	kv      *bucket
	dialErr error
}

func (s *MethodsPublicTestSuite) SetupTest() {
	s.kv = &bucket{data: map[string][]byte{}}
	s.dialErr = nil
}

// deps is a controller that serves HTTPS with its self-signed certificate.
func (s *MethodsPublicTestSuite) deps() join.Deps {
	var cfg config.Config
	cfg.Controller.API.URL = "https://osapi.example.com:8080/"
	cfg.Controller.API.TLS.Enabled = true
	cfg.Controller.PKI.KeyDir = s.T().TempDir()
	_, err := api.TLSConfig(cfg.Controller.API.TLS, cfg.Controller.PKI.KeyDir)
	s.Require().NoError(err)
	cfg.NATS.Enrollment.Bucket = "enrollment"

	return join.Deps{
		Config: cfg,
		Now:    func() time.Time { return time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC) },
		Dial: func(context.Context) (jointoken.KV, func(), error) {
			if s.dialErr != nil {
				return nil, nil, s.dialErr
			}

			return s.kv, func() {}, nil
		},
	}
}

func (s *MethodsPublicTestSuite) TestCreate() {
	ctx := context.Background()
	in := join.CreateInput{Expires: "1h", Uses: 2, Labels: map[string]string{"group": "web"}}

	d := s.deps()
	got, err := join.Create(ctx, d, in)
	s.Require().NoError(err)
	s.True(strings.HasPrefix(got.Token, "osapi_jt_"+got.ID+"."))
	u, err := url.Parse(got.Join)
	s.Require().NoError(err)
	s.Equal("osapi", u.Scheme)
	s.Equal("osapi.example.com:8080", u.Host)
	s.Equal(got.Token, u.Query().Get("token"))
	fp, err := api.CertFingerprint(d.Config.Controller.API.TLS, d.Config.Controller.PKI.KeyDir)
	s.Require().NoError(err)
	s.Equal(fp, u.Query().Get("fp"))
	s.Equal(2, got.Uses)

	noTLS := s.deps()
	noTLS.Config.Controller.API.TLS.Enabled = false
	_, err = join.Create(ctx, noTLS, in)
	s.ErrorContains(err, "controller.api.tls.enabled")

	noURL := s.deps()
	noURL.Config.Controller.API.URL = ""
	_, err = join.Create(ctx, noURL, in)
	s.ErrorContains(err, "controller.api.url")

	noCert := s.deps()
	noCert.Config.Controller.PKI.KeyDir = s.T().TempDir()
	_, err = join.Create(ctx, noCert, in)
	s.ErrorContains(err, "read the API's certificate")

	s.dialErr = errors.New("no nats")
	_, err = join.Create(ctx, s.deps(), in)
	s.ErrorContains(err, "no nats")
	s.dialErr = nil

	s.kv.err = errors.New("kv down")
	_, err = join.Create(ctx, s.deps(), in)
	s.ErrorContains(err, "store the join token")
}

func (s *MethodsPublicTestSuite) TestListAndRevoke() {
	ctx := context.Background()
	d := s.deps()
	got, err := join.List(ctx, d, join.ListInput{})
	s.Require().NoError(err)
	s.Empty(got.Items)

	made, err := join.Create(ctx, d, join.CreateInput{Expires: "1h", Uses: 1})
	s.Require().NoError(err)
	got, err = join.List(ctx, d, join.ListInput{})
	s.Require().NoError(err)
	s.Require().Len(got.Items, 1)
	s.Equal(made.ID, got.Items[0].ID)

	r, err := join.Revoke(ctx, d, join.RevokeInput{ID: made.ID})
	s.Require().NoError(err)
	s.Equal(made.ID, r.ID)
	_, err = join.Revoke(ctx, d, join.RevokeInput{ID: made.ID})
	s.ErrorIs(err, jointoken.ErrInvalid)

	s.dialErr = errors.New("no nats")
	_, err = join.List(ctx, d, join.ListInput{})
	s.ErrorContains(err, "no nats")
	_, err = join.Revoke(ctx, d, join.RevokeInput{ID: "x"})
	s.ErrorContains(err, "no nats")
	s.dialErr = nil

	s.kv.err = errors.New("kv down")
	_, err = join.List(ctx, d, join.ListInput{})
	s.ErrorContains(err, "list the join tokens")
}

func TestMethodsPublicTestSuite(t *testing.T) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
