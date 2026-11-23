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

package certificate

// ListInput is list's input: it takes none.
type ListInput struct{}

// GetInput names the certificate authority to read.
type GetInput struct {
	Name string `json:"name" validate:"required,max=255,ca_name" doc:"The CA, as list names it, such as internal-ca, or ISRG Root X1 on macOS; a name with a slash, such as mozilla/ISRG_Root_X1, is in list only, because a path segment cannot have one"`
}

// CreateInput is the CA to add to the trust store, from an object in the
// Object Store.
type CreateInput struct {
	Name   string `json:"name"   validate:"required,resource_name" doc:"The name of the CA, such as internal-ca, which osapi writes as osapi-internal-ca.crt"             telemetry:"attr"`
	Object string `json:"object" validate:"required,min=1"         doc:"The object of the Object Store that has the PEM of the CA, as osapi client file upload stored it"`
}

// UpdateInput writes a CA that osapi deployed again, from its object or from
// a different one.
type UpdateInput struct {
	Name   string `json:"name"             validate:"required,resource_name" doc:"The CA that osapi deployed, such as internal-ca"`
	Object string `json:"object,omitempty" validate:"omitempty,min=1"        doc:"The object that has the PEM; the one it was deployed from when empty"`
}

// DeleteInput names the CA that osapi deployed to remove.
type DeleteInput struct {
	Name string `json:"name" validate:"required,resource_name" doc:"The CA that osapi deployed, such as internal-ca"`
}

// Entry is a certificate authority that the host trusts.
type Entry struct {
	Name        string `json:"name"                  doc:"The CA: the name it was deployed with for a CA that osapi deployed, else its file's name"`
	Path        string `json:"path"                  doc:"The file, or on macOS the keychain, that has it"`
	Source      string `json:"source"                doc:"system for a CA of the distribution, local for one an administrator added"`
	Enabled     bool   `json:"enabled"               doc:"Whether the host trusts it: false for a system CA that the trust store turns off"`
	Subject     string `json:"subject,omitempty"     doc:"Its subject, such as CN=ISRG Root X1,O=Internet Security Research Group,C=US"             none:"-"`
	Fingerprint string `json:"fingerprint,omitempty" doc:"The SHA-256 of the certificate, in hex"                                                            table:"-"`
	NotAfter    string `json:"not_after,omitempty"   doc:"When it expires"                                                                          none:"-"`
	Managed     bool   `json:"managed"               doc:"Whether osapi deployed it, and its writes change it"`
	Object      string `json:"object,omitempty"      doc:"The object that osapi deployed it from"                                                   none:"-"`
}

// Result is what a write did.
type Result struct {
	Name    string `json:"name"    doc:"The CA"`
	Changed bool   `json:"changed" doc:"Whether the host changed"`
}
