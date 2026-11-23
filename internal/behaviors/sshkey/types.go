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

package sshkey

// ListInput names the user whose keys to list.
type ListInput struct {
	Name string `json:"name" validate:"required,max=32,account_name" doc:"The user, such as alice" telemetry:"attr"`
}

// CreateInput is a key to add to a user's authorized_keys.
type CreateInput struct {
	Name string `json:"name" validate:"required,max=32,account_name" doc:"The user, such as alice"                                                                      telemetry:"attr"`
	Key  string `json:"key"  validate:"required,no_linebreak"        doc:"The public key line, such as ssh-ed25519 AAAA... alice@laptop, with options before it if any"`
}

// DeleteInput names a key of a user by its fingerprint. The fingerprint
// stays in the body: its base64 can have a slash, which cannot go in a path
// segment.
type DeleteInput struct {
	Name        string `json:"name"        validate:"required,max=32,account_name"             doc:"The user, such as alice"                             telemetry:"attr"`
	Fingerprint string `json:"fingerprint" validate:"required,startswith=SHA256:,no_linebreak" doc:"The SHA256 fingerprint of the key, as list shows it" telemetry:"attr"`
}

// Key is one key of a user's authorized_keys.
type Key struct {
	Type        string `json:"type"              doc:"The key type, such as ssh-ed25519"`
	Fingerprint string `json:"fingerprint"       doc:"The SHA256 fingerprint of the key"`
	Comment     string `json:"comment,omitempty" doc:"The comment after the key, such as alice@laptop"`
	Options     string `json:"options,omitempty" doc:"The options before the key, such as from=\"10.0.0.0/8\""`
}

// Result is what a write did.
type Result struct {
	Fingerprint string `json:"fingerprint" doc:"The SHA256 fingerprint of the key"`
	Changed     bool   `json:"changed"     doc:"Whether the host changed"`
}
