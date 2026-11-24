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
package password

import (
	"context"

	pwhash "github.com/osapi-io/osapi/internal/password"
)

// Update sets the password of a user. A new hash differs each time, so it
// always changes the host.
func Update(
	ctx context.Context,
	p Provider,
	in UpdateInput,
) (Result, error) {
	if err := p.Set(ctx, in.Name, in.PasswordHash); err != nil {
		return Result{Name: in.Name}, err
	}

	return Result{Name: in.Name, Changed: true}, nil
}

var hashFn = pwhash.Hash

// hashPassword is update's controller step: the password becomes its hash,
// so the job has no plain password.
func hashPassword(
	_ context.Context,
	in UpdateInput,
) (UpdateInput, error) {
	if in.Password == "" {
		return in, nil
	}
	hash, err := hashFn(in.Password)
	if err != nil {
		return UpdateInput{}, err
	}
	in.PasswordHash, in.Password = hash, ""

	return in, nil
}
