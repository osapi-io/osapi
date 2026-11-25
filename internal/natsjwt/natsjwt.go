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

// Package natsjwt mints the NATS user JWTs of osapi in operator mode: the
// bootstrap user of a join token, and the user of an agent that joined. It
// signs them with an account signing key that the NATS administrator gives
// the controller.
package natsjwt

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"

	"github.com/osapi-io/osapi/internal/bus"
)

// ErrSigningKey is a signing key file that holds no account seed.
var ErrSigningKey = errors.New("the signing key file holds no account seed (SA...)")

// Signer mints user JWTs for one account with one of its signing keys.
type Signer struct {
	key     nkeys.KeyPair
	account string
}

// Load is the signer of the account signing key seed in file, for the
// account with public key account. The key can be the account's own key or
// one of its signing keys.
func Load(
	file string,
	account string,
) (*Signer, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("read the signing key: %w", err)
	}
	key, err := nkeys.FromSeed([]byte(strings.TrimSpace(string(data))))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSigningKey, err)
	}
	if pub, _ := key.PublicKey(); !nkeys.IsValidPublicAccountKey(pub) {
		return nil, ErrSigningKey
	}
	if !nkeys.IsValidPublicAccountKey(account) {
		return nil, fmt.Errorf("the account %q is not an account public key (A...)", account)
	}

	return &Signer{key: key, account: account}, nil
}

// Mint is a user JWT for the user nkey userKey, named name, with the
// permissions p, that expires at expires, or never when it is zero.
func (s *Signer) Mint(
	userKey string,
	name string,
	p bus.Permissions,
	expires time.Time,
) (string, error) {
	if !nkeys.IsValidPublicUserKey(userKey) {
		return "", fmt.Errorf("%q is not a user public key (U...)", userKey)
	}
	uc := jwt.NewUserClaims(userKey)
	uc.Name = name
	uc.IssuerAccount = s.account
	uc.Pub.Allow = p.Publish
	uc.Sub.Allow = p.Subscribe
	if p.Responses {
		uc.Resp = &jwt.ResponsePermission{MaxMsgs: 1, Expires: time.Minute}
	}
	if !expires.IsZero() {
		uc.Expires = expires.Unix()
	}

	return uc.Encode(s.key)
}

// createUser makes a user nkey. Tests swap it.
var createUser = nkeys.CreateUser

// NewUser is a new user nkey: its seed, which stays with its owner, and its
// public key, which a JWT names.
func NewUser() (seed []byte, public string, err error) {
	kp, err := createUser()
	if err != nil {
		return nil, "", err
	}
	seed, _ = kp.Seed()
	public, _ = kp.PublicKey()

	return seed, public, nil
}

// Creds is the credentials file of a user: its JWT and its seed.
func Creds(
	userJWT string,
	seed []byte,
) ([]byte, error) {
	return jwt.FormatUserConfig(userJWT, seed)
}

// NewCreds are the credentials of a new user, named name, with the
// permissions p, that expire at expires: its JWT and its seed, for a user
// whose seed osapi hands over, such as the bootstrap user of a join token.
func (s *Signer) NewCreds(
	name string,
	p bus.Permissions,
	expires time.Time,
) ([]byte, error) {
	seed, pub, err := NewUser()
	if err != nil {
		return nil, err
	}
	userJWT, err := s.Mint(pub, name, p, expires)
	if err != nil {
		return nil, err
	}

	return Creds(userJWT, seed)
}
