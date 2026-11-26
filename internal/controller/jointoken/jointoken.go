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

// Package jointoken is the store of join tokens: a token that admits an
// agent at its join (osapi-io/osapi#1313). The controller keeps a hash of
// each token in its enrollment bucket, never the token, with its expiry, the
// number of uses that it has left and the labels that an agent that uses it
// gets. Consume, which the join calls, takes one use.
package jointoken

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/osapi-io/osapi/internal/telemetry/opspan"

	"github.com/nats-io/nats.go/jetstream"
)

// prefix is the prefix of a token's key in the enrollment bucket.
const prefix = "jointoken."

// tokenPrefix starts each token, so a person and a scanner of secrets know
// one when they see it.
const tokenPrefix = "osapi_jt_"

// The causes of a token that does not admit.
var (
	// ErrInvalid is a token that the controller does not have: unknown,
	// revoked, or malformed.
	ErrInvalid = errors.New("join token is not valid")
	// ErrExpired is a token past its expiry.
	ErrExpired = errors.New("join token has expired")
	// ErrUsed is a token with no use left.
	ErrUsed = errors.New("join token has no use left")
)

// KV is what the store needs of the enrollment bucket.
type KV interface {
	KVGet(ctx context.Context, bucket string, key string) (jetstream.KeyValueEntry, error)
	KVCreate(ctx context.Context, bucket string, key string, value []byte) (uint64, error)
	KVUpdate(
		ctx context.Context,
		bucket string,
		key string,
		value []byte,
		revision uint64,
	) (uint64, error)
	KVDelete(ctx context.Context, bucket string, key string) error
	KVKeys(ctx context.Context, bucket string, filters ...string) ([]string, error)
}

// Token is what the controller keeps of a join token.
type Token struct {
	// ID names the token in list and revoke. It is not secret.
	ID string `json:"id"`
	// Hash is the SHA-256 of the secret part of the token, in hex.
	Hash string `json:"hash"`
	// CreatedAt is when it was made.
	CreatedAt time.Time `json:"created_at"`
	// ExpiresAt is when it stops admitting.
	ExpiresAt time.Time `json:"expires_at"`
	// UsesLeft is how many more agents it admits.
	UsesLeft int `json:"uses_left"`
	// Labels are the labels that an agent that joins with it gets.
	Labels map[string]string `json:"labels,omitempty"`
}

// Store keeps the join tokens in bucket.
type Store struct {
	KV     KV
	Bucket string
	// Now is the time, for a test to replace.
	Now func() time.Time
}

// randRead is rand.Read, for a test to replace.
var randRead = rand.Read

// Create makes a token that admits uses agents until expires has passed, and
// gives each the labels. It returns the token, which it does not keep, and
// what it keeps.
func (s Store) create(
	ctx context.Context,
	expires time.Duration,
	uses int,
	labels map[string]string,
) (string, Token, error) {
	id := make([]byte, 4)
	secret := make([]byte, 32)
	if _, err := randRead(id); err != nil {
		return "", Token{}, fmt.Errorf("make a join token: %w", err)
	}
	if _, err := randRead(secret); err != nil {
		return "", Token{}, fmt.Errorf("make a join token: %w", err)
	}
	now := s.Now()
	t := Token{
		ID:        hex.EncodeToString(id),
		Hash:      hash(secret),
		CreatedAt: now,
		ExpiresAt: now.Add(expires),
		UsesLeft:  uses,
		Labels:    labels,
	}
	data, _ := json.Marshal(t)
	if _, err := s.KV.KVCreate(ctx, s.Bucket, prefix+t.ID, data); err != nil {
		return "", Token{}, fmt.Errorf("store the join token: %w", err)
	}

	return tokenPrefix + t.ID + "." + base64.RawURLEncoding.EncodeToString(secret), t, nil
}

// List is each token that the controller keeps, by when it was made.
func (s Store) list(
	ctx context.Context,
) ([]Token, error) {
	keys, err := s.KV.KVKeys(ctx, s.Bucket, prefix+">")
	if err != nil && !errors.Is(err, jetstream.ErrNoKeysFound) {
		return nil, fmt.Errorf("list the join tokens: %w", err)
	}
	out := []Token{}
	for _, k := range keys {
		t, _, err := s.get(ctx, strings.TrimPrefix(k, prefix))
		if err != nil {
			continue
		}
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })

	return out, nil
}

// Revoke removes the token id, so it admits no agent.
func (s Store) revoke(
	ctx context.Context,
	id string,
) error {
	if _, _, err := s.get(ctx, id); err != nil {
		return err
	}
	if err := s.KV.KVDelete(ctx, s.Bucket, prefix+id); err != nil {
		return fmt.Errorf("revoke the join token: %w", err)
	}

	return nil
}

// Consume checks token, and takes one of its uses: ErrInvalid, ErrExpired
// or ErrUsed when it does not admit. Two joins at once cannot both take the
// last use: the update is a compare and swap on the revision.
func (s Store) consume(
	ctx context.Context,
	token string,
) (Token, error) {
	t, rev, err := s.check(ctx, token)
	if err != nil {
		return Token{}, err
	}
	t.UsesLeft--
	data, _ := json.Marshal(t)
	if _, err := s.KV.KVUpdate(ctx, s.Bucket, prefix+t.ID, data, rev); err != nil {
		return Token{}, fmt.Errorf("take a use of the join token: %w", err)
	}

	return t, nil
}

// check is the token, and its revision, when it admits now: ErrInvalid,
// ErrExpired or ErrUsed when it does not. It takes no use.
func (s Store) check(
	ctx context.Context,
	token string,
) (Token, uint64, error) {
	id, secret, ok := parse(token)
	if !ok {
		return Token{}, 0, ErrInvalid
	}
	t, rev, err := s.get(ctx, id)
	if err != nil {
		return Token{}, 0, err
	}
	if subtle.ConstantTimeCompare([]byte(t.Hash), []byte(hash(secret))) != 1 {
		return Token{}, 0, ErrInvalid
	}
	if !s.Now().Before(t.ExpiresAt) {
		return Token{}, 0, ErrExpired
	}
	if t.UsesLeft < 1 {
		return Token{}, 0, ErrUsed
	}

	return t, rev, nil
}

// get is the token id and its revision; ErrInvalid when there is none.
func (s Store) get(
	ctx context.Context,
	id string,
) (Token, uint64, error) {
	e, err := s.KV.KVGet(ctx, s.Bucket, prefix+id)
	if errors.Is(err, jetstream.ErrKeyNotFound) {
		return Token{}, 0, ErrInvalid
	}
	if err != nil {
		return Token{}, 0, fmt.Errorf("read the join token: %w", err)
	}
	var t Token
	if err := json.Unmarshal(e.Value(), &t); err != nil {
		return Token{}, 0, fmt.Errorf("read the join token: %w", err)
	}

	return t, e.Revision(), nil
}

// ID is the ID of token, which is not secret, and whether token has the
// form of a join token.
func ID(
	token string,
) (string, bool) {
	id, _, ok := parse(token)

	return id, ok
}

// parse splits a token into its ID and its secret.
func parse(
	token string,
) (string, []byte, bool) {
	rest, ok := strings.CutPrefix(token, tokenPrefix)
	if !ok {
		return "", nil, false
	}
	id, enc, ok := strings.Cut(rest, ".")
	if !ok || id == "" {
		return "", nil, false
	}
	secret, err := base64.RawURLEncoding.DecodeString(enc)
	if err != nil || len(secret) == 0 {
		return "", nil, false
	}

	return id, secret, true
}

// hash is the SHA-256 of secret, in hex.
func hash(
	secret []byte,
) string {
	sum := sha256.Sum256(secret)

	return hex.EncodeToString(sum[:])
}

// Create makes a token, in a jointoken.create span that names its ID. The
// span never holds the token.
func (s Store) Create(
	ctx context.Context,
	expires time.Duration,
	uses int,
	labels map[string]string,
) (string, Token, error) {
	ctx, span := opspan.Start(ctx, tracerName, "jointoken.create",
		attribute.Int("osapi.jointoken.uses", uses),
	)
	token, t, err := s.create(ctx, expires, uses, labels)
	span.SetAttributes(attribute.String("osapi.jointoken.id", t.ID))

	return token, t, opspan.End(span, err)
}

// List is each token, in a jointoken.list span.
func (s Store) List(
	ctx context.Context,
) ([]Token, error) {
	ctx, span := opspan.Start(ctx, tracerName, "jointoken.list")
	tokens, err := s.list(ctx)

	return tokens, opspan.End(span, err)
}

// Revoke removes the token with id, in a jointoken.revoke span.
func (s Store) Revoke(
	ctx context.Context,
	id string,
) error {
	ctx, span := opspan.Start(ctx, tracerName, "jointoken.revoke",
		attribute.String("osapi.jointoken.id", id),
	)

	return opspan.End(span, s.revoke(ctx, id))
}

// Consume takes a use of token, in a jointoken.consume span that names the
// token's ID, never the token.
func (s Store) Consume(
	ctx context.Context,
	token string,
) (Token, error) {
	id, _ := ID(token)
	ctx, span := opspan.Start(ctx, tracerName, "jointoken.consume",
		attribute.String("osapi.jointoken.id", id),
	)
	t, err := s.consume(ctx, token)

	return t, opspan.End(span, err)
}

// Check is token when it admits now, in a jointoken.check span that names
// the token's ID, never the token. It takes no use: the auth callout checks
// a token at each connection of its join.
func (s Store) Check(
	ctx context.Context,
	token string,
) (Token, error) {
	id, _ := ID(token)
	ctx, span := opspan.Start(ctx, tracerName, "jointoken.check",
		attribute.String("osapi.jointoken.id", id),
	)
	t, _, err := s.check(ctx, token)

	return t, opspan.End(span, err)
}

// tracerName is the tracer of the join token store.
const tracerName = "osapi-jointoken"
