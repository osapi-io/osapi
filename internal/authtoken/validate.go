// Copyright (c) 2024 John Dewey

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

package authtoken

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v4"

	"github.com/osapi-io/osapi/internal/validation"
)

// ErrNoKeyID is a token without a kid header.
var ErrNoKeyID = errors.New("token has no kid")

// ErrUnknownKeyID is a token whose kid names no key of the config.
var ErrUnknownKeyID = errors.New("unknown kid")

// Validate parses and validates the JWT with the key that its kid header
// names. A token without a kid, or with a kid that keys does not have, is
// refused: Validate does not try each key. The claims carry the kid as
// KeyID.
func (t *Token) Validate(
	tokenString string,
	keys Keys,
) (*CustomClaims, error) {
	claims := &CustomClaims{}

	_, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			kid, _ := token.Header["kid"].(string)
			if kid == "" {
				return nil, ErrNoKeyID
			}
			key, ok := keys.Lookup(kid)
			if !ok {
				return nil, fmt.Errorf("%w %q", ErrUnknownKeyID, kid)
			}
			claims.KeyID = kid

			return []byte(key.Secret), nil
		},
	)
	if err != nil {
		return nil, err
	}

	if err := validation.Instance().Struct(claims); err != nil {
		return nil, err
	}

	return claims, nil
}
