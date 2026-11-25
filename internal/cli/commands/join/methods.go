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

package join

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/osapi-io/osapi/internal/controller/api"
	"github.com/osapi-io/osapi/internal/controller/jointoken"
	"github.com/osapi-io/osapi/internal/job"
)

// errNoTLS is the error of create on a controller that does not serve
// HTTPS: a join pins the controller's certificate.
var errNoTLS = errors.New(
	"a join needs controller.api.tls.enabled: the join token pins the controller's certificate",
)

// errNoURL is the error of create with no controller.api.url, which the
// join string carries.
var errNoURL = errors.New("a join needs controller.api.url: the address that an agent joins at")

// store is the join token store in the enrollment bucket, and the function
// that closes its connection.
func store(
	ctx context.Context,
	d Deps,
) (jointoken.Store, func(), error) {
	kv, closeKV, err := d.Dial(ctx)
	if err != nil {
		return jointoken.Store{}, nil, err
	}
	bucket := job.ApplyNamespaceToInfraName(
		d.Config.Controller.NATS.Namespace, d.Config.NATS.Enrollment.Bucket,
	)

	return jointoken.Store{KV: kv, Bucket: bucket, Now: d.Now}, closeKV, nil
}

// Create makes a token and prints it one time, with its join string.
func Create(
	ctx context.Context,
	d Deps,
	in CreateInput,
) (CreateOutput, error) {
	a := d.Config.Controller.API
	if !a.TLS.Enabled {
		return CreateOutput{}, errNoTLS
	}
	u, err := url.Parse(a.URL)
	if a.URL == "" || err != nil || u.Host == "" {
		return CreateOutput{}, errNoURL
	}
	fp, err := api.CertFingerprint(a.TLS, d.Config.Controller.PKI.KeyDir)
	if err != nil {
		return CreateOutput{}, err
	}
	expires, _ := time.ParseDuration(in.Expires)
	s, closeKV, err := store(ctx, d)
	if err != nil {
		return CreateOutput{}, err
	}
	defer closeKV()
	token, t, err := s.Create(ctx, expires, in.Uses, in.Labels)
	if err != nil {
		return CreateOutput{}, err
	}
	q := url.Values{"token": {token}, "fp": {fp}}

	return CreateOutput{
		ID:        t.ID,
		Token:     token,
		Join:      fmt.Sprintf("osapi://%s?%s", u.Host, q.Encode()),
		ExpiresAt: t.ExpiresAt,
		Uses:      t.UsesLeft,
		Labels:    t.Labels,
	}, nil
}

// List is each join token, with no secret.
func List(
	ctx context.Context,
	d Deps,
	_ ListInput,
) (ListOutput, error) {
	s, closeKV, err := store(ctx, d)
	if err != nil {
		return ListOutput{}, err
	}
	defer closeKV()
	tokens, err := s.List(ctx)
	if err != nil {
		return ListOutput{}, err
	}
	out := ListOutput{Items: []Row{}}
	for _, t := range tokens {
		out.Items = append(out.Items, Row{
			ID: t.ID, CreatedAt: t.CreatedAt, ExpiresAt: t.ExpiresAt,
			UsesLeft: t.UsesLeft, Labels: t.Labels,
		})
	}

	return out, nil
}

// Revoke removes the token, so it admits no agent.
func Revoke(
	ctx context.Context,
	d Deps,
	in RevokeInput,
) (RevokeOutput, error) {
	s, closeKV, err := store(ctx, d)
	if err != nil {
		return RevokeOutput{}, err
	}
	defer closeKV()
	if err := s.Revoke(ctx, in.ID); err != nil {
		return RevokeOutput{}, err
	}

	return RevokeOutput{ID: in.ID}, nil
}
