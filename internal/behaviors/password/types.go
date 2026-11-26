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

import "github.com/osapi-io/osapi/internal/secret"

// UpdateInput sets the password of a user. The caller sends password; the
// controller's step hashes it into password_hash.
type UpdateInput struct {
	Name         string       `json:"name"                    validate:"required,max=32,account_name" doc:"The user, such as alice"`
	Password     secret.Value `json:"password,omitempty"      validate:"omitempty,no_linebreak"       doc:"The new password, which osapi hashes on the controller"`
	PasswordHash string       `json:"password_hash,omitempty" validate:"omitempty"                    doc:"The sha512-crypt hash of the password, which the controller sets" before:"set"`
}

// Result is what update reports for the user.
type Result struct {
	Name    string `json:"name"    doc:"The user"`
	Changed bool   `json:"changed" doc:"Whether the host changed"`
}
