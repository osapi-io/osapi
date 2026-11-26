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
// Package natscallout is osapi's NATS auth callout service. A NATS server
// with an auth_callout block sends each new connection to it, and it decides
// who the connection is: a host with a join token, or an admitted agent with
// its NATS user key. It answers with a user JWT that has the permissions of
// that role, signed with the issuer key that the server's block names.
package natscallout

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"

	"github.com/osapi-io/osapi/internal/bus"
)

// Subject is where a NATS server sends each authorization request.
const Subject = "$SYS.REQ.USER.AUTH"

// ErrRefused is a connection that the callout does not admit.
var ErrRefused = errors.New("refused")

// Decision is who a connection is: its name, and what it may do until it
// expires.
type Decision struct {
	// Name is the user's name in the server, such as agent:<machine ID>.
	Name string
	// Permissions are the subjects of its role.
	Permissions bus.Permissions
	// Expires is when the server closes the connection, or zero for never.
	Expires time.Time
}

// Decide is who the connection of req is, or an error that refuses it. Its
// error message goes back to the server, so it must hold no secret.
type Decide func(ctx context.Context, req *jwt.AuthorizationRequest) (Decision, error)

// Service answers the authorization requests of a NATS server.
type Service struct {
	// Issuer is the account key that the server's auth_callout names as its
	// issuer. It signs each answer and each user JWT.
	Issuer nkeys.KeyPair
	// Account is the account that a user joins, such as OSAPI.
	Account string
	// Decide decides each connection.
	Decide Decide
	Logger *slog.Logger
}

// Reply is the answer to the authorization request in data: a signed
// response with a user JWT, or with the reason of a refusal.
func (s Service) Reply(
	ctx context.Context,
	data []byte,
) ([]byte, error) {
	rc, err := jwt.DecodeAuthorizationRequestClaims(string(data))
	if err != nil {
		return nil, fmt.Errorf("the authorization request: %w", err)
	}
	req := &rc.AuthorizationRequest
	resp := jwt.NewAuthorizationResponseClaims(req.UserNkey)
	resp.Audience = req.Server.ID
	d, err := s.Decide(ctx, req)
	if err != nil {
		resp.Error = err.Error()
	} else {
		uc := jwt.NewUserClaims(req.UserNkey)
		uc.Name = d.Name
		uc.Audience = s.Account
		uc.Pub.Allow = d.Permissions.Publish
		uc.Sub.Allow = d.Permissions.Subscribe
		if d.Permissions.Responses {
			uc.Resp = &jwt.ResponsePermission{MaxMsgs: 1, Expires: time.Minute}
		}
		if !d.Expires.IsZero() {
			uc.Expires = d.Expires.Unix()
		}
		token, err := uc.Encode(s.Issuer)
		if err != nil {
			return nil, fmt.Errorf("sign the user: %w", err)
		}
		resp.Jwt = token
	}
	out, err := resp.Encode(s.Issuer)
	if err != nil {
		return nil, fmt.Errorf("sign the answer: %w", err)
	}

	return []byte(out), nil
}

// Subscribe subscribes a handler to a subject: the Subscribe of a NATS
// connection.
type Subscribe func(subject string, handler nats.MsgHandler) (*nats.Subscription, error)

// Serve answers each authorization request until ctx ends. subscribe must
// be of a connection of one of the server's auth_users, which the callout
// does not decide.
func (s Service) Serve(
	ctx context.Context,
	subscribe Subscribe,
) error {
	sub, err := subscribe(Subject, func(msg *nats.Msg) {
		out, err := s.Reply(ctx, msg.Data)
		if err != nil {
			s.Logger.Warn("could not answer an authorization request",
				slog.String("error", err.Error()))

			return
		}
		_ = msg.Respond(out)
	})
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", Subject, err)
	}
	<-ctx.Done()
	_ = sub.Unsubscribe()

	return nil
}

// VerifyNkey is nil when the connection of req holds the seed of its nkey:
// the server's nonce, signed with that key. The server does not check it,
// because the callout decides the connection.
func VerifyNkey(
	req *jwt.AuthorizationRequest,
) error {
	o := req.ConnectOptions
	if o.Nkey == "" || o.SignedNonce == "" {
		return fmt.Errorf("%w: no nkey", ErrRefused)
	}
	key, err := nkeys.FromPublicKey(o.Nkey)
	if err != nil {
		return fmt.Errorf("%w: the nkey is not a public key", ErrRefused)
	}
	sig, err := decodeSig(o.SignedNonce)
	if err != nil || key.Verify([]byte(req.ClientInformation.Nonce), sig) != nil {
		return fmt.Errorf("%w: the nonce is not signed by the nkey", ErrRefused)
	}

	return nil
}

// decodeSig is a signature as a client sends it, in raw URL base64, or in
// standard base64 from an older client.
func decodeSig(
	s string,
) ([]byte, error) {
	if b, err := base64.RawURLEncoding.DecodeString(s); err == nil {
		return b, nil
	}

	return base64.StdEncoding.DecodeString(s)
}
