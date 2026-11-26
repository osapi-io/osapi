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

// Package pkilist is osapi pki list: each key and each certificate of osapi
// on this host, in one list.
package pkilist

import "github.com/osapi-io/osapi/internal/cli/command"

// Group is the list group's definition. It adds no word: its method is
// osapi pki list.
var Group = command.Register(command.Define(
	"pki", "List each key and each certificate of osapi on this host",
	command.Under(command.ParentPKI),
	command.Flat(),
	command.Uses[Deps](),
	command.Method[ListInput, ListOutput]("list").
		Short("List each key and each certificate of osapi on this host").
		Long("Run this on the host. It reads what osapi pki key list and osapi pki cert\n"+
			"list read, in one list: the kind, the name, the component, the state, the\n"+
			"algorithm, the fingerprint, or the ID of a secret key, and the expiry of a\n"+
			"certificate.").
		Example("").
		Prints(sampleList).
		Handle(List),
))
