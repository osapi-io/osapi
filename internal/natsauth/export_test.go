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

package natsauth

import (
	"time"

	"github.com/nats-io/nkeys"

	"github.com/osapi-io/osapi/internal/bus"
	"github.com/osapi-io/osapi/internal/natsjwt"
)

// SetFns swaps how a create reads random bytes, makes keys and mints
// credentials; a nil keeps the real one. It returns a func that puts them
// back.
func SetFns(
	random func([]byte) (int, error),
	keys func(int) ([]nkeys.KeyPair, error),
	creds func(*natsjwt.Signer, string, bus.Permissions, time.Time) ([]byte, error),
) func() {
	oldR, oldK, oldC := randomRead, newKeys, newCreds
	if random != nil {
		randomRead = random
	}
	if keys != nil {
		newKeys = keys
	}
	if creds != nil {
		newCreds = creds
	}

	return func() { randomRead, newKeys, newCreds = oldR, oldK, oldC }
}

// SetKeyMakers swaps how the operator key and the account keys are made,
// and returns a func that puts them back.
func SetKeyMakers(
	op func() (nkeys.KeyPair, error),
	acct func() (nkeys.KeyPair, error),
) func() {
	oldO, oldA := createOperator, createAccount
	createOperator, createAccount = op, acct

	return func() { createOperator, createAccount = oldO, oldA }
}
