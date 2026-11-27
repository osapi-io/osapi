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

package pkilist

import (
	"context"

	"github.com/osapi-io/osapi/internal/cli/commands/pkikey"
	"github.com/osapi-io/osapi/internal/pki"
)

// List is each key, then each certificate, of osapi on this host.
func List(
	_ context.Context,
	d Deps,
	_ ListInput,
) (ListOutput, error) {
	out := ListOutput{Items: []Item{}}
	for _, k := range d.Keys() {
		out.Items = append(out.Items, ofKey(k))
	}
	for _, c := range d.Certs() {
		out.Items = append(out.Items, ofCert(c))
	}

	return out, nil
}

// ofKey is k as an item: its fingerprint, or the ID of a secret key.
func ofKey(
	k pkikey.Key,
) Item {
	fp := k.Fingerprint
	if fp == "" {
		fp = k.KeyID
	}
	if fp == "" {
		fp = k.PublicKey
	}

	return Item{
		Kind: KindKey, Name: k.Name, Component: k.Component, State: k.State,
		File: k.File, Algorithm: k.Algorithm, Approved: k.Approved, Fingerprint: fp,
	}
}

// ofCert is c as an item, with its expiry when osapi read it.
func ofCert(
	c pki.Cert,
) Item {
	it := Item{
		Kind: KindCert, Name: c.Name, Component: c.Component, State: c.State,
		File: c.File, Algorithm: c.Key, Approved: c.FIPS != nil && *c.FIPS,
		Fingerprint: c.Fingerprint,
	}
	if !c.Expires.IsZero() {
		t := c.Expires
		it.Expires = &t
	}

	return it
}
