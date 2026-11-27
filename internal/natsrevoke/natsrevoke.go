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

// Package natsrevoke cuts the NATS user of a machine from a NATS server in
// operator mode at once. It adds the user's key to the revocations of
// osapi's account, signs the account again with an operator signing key,
// and sends it to the server as a user of the system account. The server
// closes each connection of the user, and refuses each JWT of the user that
// was issued before the revocation.
//
// The server keeps the account that it gets in memory only: at its next
// start it has the account of its resolver file again. Thus the revocations
// are records in KV, and Run sends the account again while any revocation
// is left.
package natsrevoke

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nats-io/nkeys"
)

// kvPrefix keys the time of each revocation by the user's public key.
const kvPrefix = "natsrevoked."

// KV is the store of the revocations.
type KV interface {
	KVPut(ctx context.Context, bucket string, key string, value []byte) (uint64, error)
	KVGet(ctx context.Context, bucket string, key string) (jetstream.KeyValueEntry, error)
	KVDelete(ctx context.Context, bucket string, key string) error
	KVKeys(ctx context.Context, bucket string, filters ...string) ([]string, error)
}

// Conn sends a request to the NATS server as a user of the system account.
type Conn interface {
	RequestWithContext(ctx context.Context, subj string, data []byte) (*nats.Msg, error)
}

// Config is what a Revoker needs.
type Config struct {
	KV     KV
	Bucket string
	Conn   Conn
	// Signer is a signing key of the operator: the server takes an
	// account that the operator or one of its signing keys signed.
	Signer nkeys.KeyPair
	// AccountJWTFile is the JWT of osapi's account, as the server has it.
	// Each send reads it again, so an account that an operator update
	// signed again keeps its new permissions.
	AccountJWTFile string
	// TTL is how long a JWT of an agent lasts. A revocation older than TTL
	// is removed: each JWT that it refuses expired. Zero keeps each one.
	TTL time.Duration
	// Interval is how often Run sends the account again. Zero is a minute.
	Interval time.Duration
	Logger   *slog.Logger
}

// Revoker revokes NATS users.
type Revoker struct {
	cfg Config
	// mu keeps the sends in order: a send with fewer revocations must not
	// replace one with more.
	mu sync.Mutex
	// sent is the account that the last send had, and the JWT that it
	// sent: the same account is sent with the same JWT, which the server
	// takes as no change.
	sentKey  string
	sentJWT  string
	sentSubj string
}

// nowFn is the time. Tests swap it.
var nowFn = time.Now

// New is a Revoker of cfg.
func New(
	cfg Config,
) *Revoker {
	if cfg.Interval <= 0 {
		cfg.Interval = time.Minute
	}

	return &Revoker{cfg: cfg}
}

// Revoke records that each JWT of the user key that was issued until now is
// revoked, and sends the account to the server. Only a record that is not
// written is an error: a send that fails, Run sends again.
func (r *Revoker) Revoke(
	ctx context.Context,
	userKey string,
) error {
	if !nkeys.IsValidPublicUserKey(userKey) {
		return fmt.Errorf("%q is not a user public key (U...)", userKey)
	}
	at := strconv.FormatInt(nowFn().Unix(), 10)
	if _, err := r.cfg.KV.KVPut(ctx, r.cfg.Bucket, kvPrefix+userKey, []byte(at)); err != nil {
		return fmt.Errorf("record the revocation of %s: %w", userKey, err)
	}
	if err := r.Send(ctx); err != nil {
		r.cfg.Logger.Warn("failed to send the revocation to the NATS server; it is sent again",
			slog.String("user", userKey),
			slog.String("error", err.Error()),
		)
	}

	return nil
}

// Run sends the account at once, then each Interval, until ctx is done, so
// a server that started again gets the revocations again.
func (r *Revoker) Run(
	ctx context.Context,
) error {
	ticker := time.NewTicker(r.cfg.Interval)
	defer ticker.Stop()

	for {
		if err := r.Send(ctx); err != nil {
			r.cfg.Logger.Warn("failed to send the revocations to the NATS server",
				slog.String("error", err.Error()),
			)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

// Send sends osapi's account, with each revocation that is not older than
// TTL, to the server. With no revocation, it sends nothing, or the account
// with none one time after the last revocation expires: the server then
// has the account of its resolver file.
func (r *Revoker) Send(
	ctx context.Context,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	revoked, err := r.revocations(ctx)
	if err != nil {
		return err
	}
	if len(revoked) == 0 && r.sentKey == "" {
		return nil
	}
	raw, err := os.ReadFile(r.cfg.AccountJWTFile)
	if err != nil {
		return fmt.Errorf("read the account: %w", err)
	}
	account := strings.TrimSpace(string(raw))
	key := account + "\n" + revocationKey(revoked)
	switch {
	case key == r.sentKey && len(revoked) == 0:
		// The server has the account with no revocation, as its resolver
		// file has it: a server that starts again needs nothing.
		return nil
	case key == r.sentKey:
		return r.send(ctx, r.sentSubj, r.sentJWT)
	}
	ac, err := jwt.DecodeAccountClaims(account)
	if err != nil {
		return fmt.Errorf("decode the account: %w", err)
	}
	for user, at := range revoked {
		ac.RevokeAt(user, time.Unix(at, 0))
	}
	signed, err := ac.Encode(r.cfg.Signer)
	if err != nil {
		return fmt.Errorf("sign the account: %w", err)
	}
	if err := r.send(ctx, ac.Subject, signed); err != nil {
		return err
	}
	r.sentKey, r.sentJWT, r.sentSubj = key, signed, ac.Subject

	return nil
}

// revocations is the time of each revocation that is not older than TTL.
// It removes the older ones.
func (r *Revoker) revocations(
	ctx context.Context,
) (map[string]int64, error) {
	keys, err := r.cfg.KV.KVKeys(ctx, r.cfg.Bucket, kvPrefix+">")
	if err != nil && !errors.Is(err, jetstream.ErrNoKeysFound) {
		return nil, fmt.Errorf("list the revocations: %w", err)
	}
	now := nowFn()
	out := map[string]int64{}
	for _, k := range keys {
		entry, err := r.cfg.KV.KVGet(ctx, r.cfg.Bucket, k)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", k, err)
		}
		at, err := strconv.ParseInt(string(bytes.TrimSpace(entry.Value())), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", k, err)
		}
		if r.cfg.TTL > 0 && now.Sub(time.Unix(at, 0)) > r.cfg.TTL {
			if err := r.cfg.KV.KVDelete(ctx, r.cfg.Bucket, k); err != nil {
				return nil, fmt.Errorf("remove %s: %w", k, err)
			}

			continue
		}
		out[strings.TrimPrefix(k, kvPrefix)] = at
	}

	return out, nil
}

// revocationKey is revoked in a fixed order.
func revocationKey(
	revoked map[string]int64,
) string {
	users := make([]string, 0, len(revoked))
	for u := range revoked {
		users = append(users, u)
	}
	slices.Sort(users)
	var b strings.Builder
	for _, u := range users {
		fmt.Fprintf(&b, "%s=%d\n", u, revoked[u])
	}

	return b.String()
}

// reply is the server's answer to an account that it was sent.
type reply struct {
	Data *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"data"`
	Error *struct {
		Code        int    `json:"code"`
		Description string `json:"description"`
	} `json:"error"`
}

// send sends signed, the JWT of the account with the public key account, to
// the server.
func (r *Revoker) send(
	ctx context.Context,
	account string,
	signed string,
) error {
	msg, err := r.cfg.Conn.RequestWithContext(ctx, ClaimsUpdateSubject(account), []byte(signed))
	if err != nil {
		return fmt.Errorf("send the account: %w", err)
	}
	var rep reply
	if err := json.Unmarshal(msg.Data, &rep); err != nil {
		return fmt.Errorf("the server's reply: %w", err)
	}
	switch {
	case rep.Error != nil:
		return fmt.Errorf("the server refused the account: %s", rep.Error.Description)
	case rep.Data == nil || rep.Data.Code != 200:
		return fmt.Errorf("the server did not take the account: %s", msg.Data)
	}

	return nil
}

// Connector is the part of a NATS connection that a request needs.
type Connector interface {
	Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error)
	PublishMsg(msg *nats.Msg) error
}

// requestTimeout bounds a request whose ctx has no deadline.
const requestTimeout = 10 * time.Second

// Requester is a Conn of a Connector: it subscribes to an inbox, publishes
// the request with the inbox as its reply, and waits for one reply.
type Requester struct {
	NC Connector
}

// RequestWithContext sends data to subj and returns the first reply.
func (q Requester) RequestWithContext(
	ctx context.Context,
	subj string,
	data []byte,
) (*nats.Msg, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, requestTimeout)
		defer cancel()
	}
	inbox := nats.NewInbox()
	replies := make(chan *nats.Msg, 1)
	sub, err := q.NC.Subscribe(inbox, func(m *nats.Msg) {
		select {
		case replies <- m:
		default:
		}
	})
	if err != nil {
		return nil, fmt.Errorf("subscribe to the reply: %w", err)
	}
	defer func() { _ = sub.Unsubscribe() }()
	if err := q.NC.PublishMsg(&nats.Msg{Subject: subj, Reply: inbox, Data: data}); err != nil {
		return nil, fmt.Errorf("publish the request: %w", err)
	}
	select {
	case m := <-replies:
		return m, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("wait for the reply: %w", ctx.Err())
	}
}

// ClaimsUpdateSubject is the subject of the system account that a NATS
// server in operator mode takes a new JWT of account on.
func ClaimsUpdateSubject(
	account string,
) string {
	return "$SYS.REQ.ACCOUNT." + account + ".CLAIMS.UPDATE"
}
