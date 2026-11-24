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

package chain

import "github.com/osapi-io/osapi/internal/guide"

// The sample outputs of the examples, which the docs page of each command
// shows as the CLI prints it.
var (
	sampleList = ListOutput{Items: []Summary{
		{
			Behavior: "dns",
			Chain:    "dns → netplan → managedfile",
			Calls:    []string{"netplan"},
			UsedBy:   []string{},
		},
		{
			Behavior: "netplan",
			Internal: true,
			Chain:    "netplan → managedfile",
			Calls:    []string{"managedfile"},
			UsedBy:   []string{"route", "dns", "interface"},
		},
		{
			Behavior: "managedfile",
			Internal: true,
			Chain:    "managedfile",
			Calls:    []string{},
			UsedBy:   []string{"certificate", "file", "netplan"},
		},
	}}

	sampleGet = GetOutput{
		Behavior: "dns",
		Chain:    "dns → netplan → managedfile",
		Calls:    []string{"netplan"},
		UsedBy:   []string{},
		Examples: []ExampleCalls{{
			Example: "Set the servers of an interface (update)",
			Calls: []Call{
				{
					Call:  "dns update",
					Sends: "interface_name=osapi0 servers=[9.9.9.9]",
					Gets:  "changed=true interface_name=osapi0",
				},
				{
					Call:  "└─ netplan get",
					Sends: "name=osapi0",
					Gets:  "id=osapi0 name=osapi0 section=ethernets type=ethernet",
				},
				{
					Call:  "└─ netplan update",
					Sends: "content=<9 lines> path=/etc/netplan/osapi-dns-osapi0.yaml",
					Gets:  "changed=true path=/etc/netplan/osapi-dns-osapi0.yaml",
				},
				{
					Call:  "· └─ managedfile update",
					Sends: "content=<9 lines> path=/etc/netplan/osapi-dns-osapi0.yaml",
					Gets:  "changed=true path=/etc/netplan/osapi-dns-osapi0.yaml",
				},
			},
		}},
	}

	sampleValidate = ValidateOutput{Behaviors: 33, Valid: true}

	// sampleSpec is the first term of the spec that spec get prints.
	sampleSpec = SpecOutput{Kinds: []guide.Section{{
		Kind:  "chain",
		Title: "What a chain is",
		Terms: []guide.Term{{
			Kind: "chain",
			Name: "A chain",
			Does: "The behaviors that a behavior calls, and the ones that they call: dns " +
				"calls netplan, and netplan calls managedfile. The chain has the name of " +
				"the behavior at its top. It is not declared on its own: Requires on each " +
				"behavior makes it.",
			When: "osapi dsl chain list shows each chain; osapi dsl chain " +
				"get --name <behavior> shows its tree.",
		}},
	}}}
)
