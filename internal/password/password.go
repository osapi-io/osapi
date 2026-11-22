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

// Package password hashes a plain password into the form that /etc/shadow
// and chpasswd -e take, on the controller. Thus the plain password never
// goes in the data of a job: only its hash travels to the agent. The user
// domain calls it, and a behavior's .Before step can.
package password

import (
	"fmt"

	sha512crypt "github.com/GehirnInc/crypt/sha512_crypt"
)

// generate is the crypt library call behind Hash, a variable so a test can
// make it fail.
var generate = sha512crypt.New().Generate

// Hash is the SHA-512 crypt hash of plain ("$6$<salt>$<hash>"), with a new
// random salt from crypto/rand each time.
func Hash(
	plain string,
) (string, error) {
	// A nil salt makes generate create one that is valid by construction,
	// so no password reaches this error. The check stays because the
	// signature can fail, and a test makes it fail.
	hash, err := generate([]byte(plain), nil)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	return hash, nil
}
