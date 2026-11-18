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
// Package facts is the endpoint group for the fact keys: the names a
// behavior's input can refer to as @fact.<key>. Until the group moves, the
// controller serves the hand-written spec in internal/controller/api/facts,
// and TestEndpointParity holds the two together.
package facts

import "github.com/osapi-io/osapi/internal/controller/endpoint"

// Group is the facts group's definition.
var Group = endpoint.Register(endpoint.Define(
	"facts", "The fact keys",
	endpoint.Title("Facts API"),
	endpoint.Tag("facts", "Facts", "Fact key discovery for @fact. references."),
	endpoint.Get[KeysInput, FactKeysResponse]("keys").
		Sub("keys").
		ID("GetFactKeys").
		Permission("agent:read").
		Describe("List available fact keys").
		Explain("Returns the list of fact keys that can be used in @fact. references. "+
			"Includes built-in keys and any custom keys registered by agents.").
		Returns("List of available fact keys."),
))
