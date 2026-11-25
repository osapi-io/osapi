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
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/osapi-io/osapi/internal/bus"
	"github.com/osapi-io/osapi/internal/natsjwt"

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

// errNoNATSURL is the error of a NATS join with no address of NATS.
var errNoNATSURL = errors.New(
	"a join over NATS needs controller.join.url or nats.server.url: the address that an agent joins at",
)

// errNATSFingerprint is a tls:// server whose certificate create could not read.
var errNATSFingerprint = errors.New("read the certificate of the NATS server")

// natsFingerprint is the SHA-256 fingerprint of the certificate of the NATS
// server at url, which it connects to with auth, the bootstrap user, to read
// it. Tests swap it.
var natsFingerprint = func(
	url string,
	auth []nats.Option,
) (string, error) {
	var fp string
	opts := []nats.Option{
		nats.Name("osapi-join-token"),
		nats.Timeout(10 * time.Second),
		nats.Secure(&tls.Config{
			MinVersion: tls.VersionTLS12,
			// It only reads the certificate, to put its fingerprint in the
			// join string; it sends nothing.
			InsecureSkipVerify: true, //nolint:gosec // reads the certificate only
			VerifyConnection: func(cs tls.ConnectionState) error {
				if len(cs.PeerCertificates) > 0 {
					sum := sha256.Sum256(cs.PeerCertificates[0].Raw)
					fp = "SHA256:" + hex.EncodeToString(sum[:])
				}

				return nil
			},
		}),
	}
	nc, err := nats.Connect(url, append(opts, auth...)...)
	if err != nil {
		return "", fmt.Errorf("%w at %s: %w", errNATSFingerprint, url, err)
	}
	nc.Close()

	return fp, nil
}

// natsJoin is what makes the join string of a token for a join over NATS:
// nats|tls://host:port?token=&user=&pass=&fp=&ns=. It checks the settings,
// and reads the server's certificate, before a token is made.
func natsJoin(
	d Deps,
	expires time.Time,
) (func(token string) string, error) {
	raw := d.Config.JoinURL()
	u, err := url.Parse(raw)
	if raw == "" || err != nil || u.Host == "" {
		return nil, errNoNATSURL
	}
	q := url.Values{}
	var auth []nats.Option
	if d.Config.Controller.Join.SigningKeyFile != "" {
		creds, err := bootstrapCreds(d, expires)
		if err != nil {
			return nil, err
		}
		q.Set("creds", base64.RawURLEncoding.EncodeToString(creds))
		auth = append(auth, nats.UserCredentialBytes(creds))
	} else if user, _ := d.Config.JoinUsers(); user.Username != "" {
		q.Set("user", user.Username)
		q.Set("pass", user.Password)
		auth = append(auth, nats.UserInfo(user.Username, user.Password))
	}
	if ns := d.Config.Controller.NATS.Namespace; ns != "" {
		q.Set("ns", ns)
	}
	if u.Scheme == "tls" {
		fp, err := natsFingerprint(raw, auth)
		if err != nil {
			return nil, err
		}
		q.Set("fp", fp)
	}

	return func(token string) string {
		q.Set("token", token)

		return fmt.Sprintf("%s://%s?%s", u.Scheme, u.Host, q.Encode())
	}, nil
}

// Create makes a token and prints it one time, with its join string.
func Create(
	ctx context.Context,
	d Deps,
	in CreateInput,
) (CreateOutput, error) {
	if in.Transport == "nats" {
		expires, _ := time.ParseDuration(in.Expires)
		join, err := natsJoin(d, d.Now().Add(expires))
		if err != nil {
			return CreateOutput{}, err
		}

		return create(ctx, d, in, join)
	}
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

	return create(ctx, d, in, func(token string) string {
		q := url.Values{"token": {token}, "fp": {fp}}

		return fmt.Sprintf("osapi://%s?%s", u.Host, q.Encode())
	})
}

// create makes the token, and its join string with join.
func create(
	ctx context.Context,
	d Deps,
	in CreateInput,
	join func(token string) string,
) (CreateOutput, error) {
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
	return CreateOutput{
		ID:        t.ID,
		Token:     token,
		Join:      join(token),
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

// bootstrapCreds are, in operator mode, the credentials of a new bootstrap
// user: a user JWT with the join role's permissions, which expires with the
// token, signed with controller.join.signing_key_file.
func bootstrapCreds(
	d Deps,
	expires time.Time,
) ([]byte, error) {
	j := d.Config.Controller.Join
	signer, err := natsjwt.Load(j.SigningKeyFile, j.Account)
	if err != nil {
		return nil, err
	}
	perms, _ := bus.RolePermissions(bus.RoleJoin, d.Config.Controller.NATS.Namespace, d.Config.NATS)

	return signer.NewCreds("osapi-join", perms, expires)
}
