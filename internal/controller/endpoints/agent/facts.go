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
// Package agent is the endpoint group for the agents of the fleet: the
// agents that report, and the actions on one agent. Admitting an agent is
// the admission group. Its spec part comes from this definition; until the group
// moves, the controller serves the hand-written spec in
// internal/controller/api/agent, and TestEndpointParity holds the two
// together.
package agent

import (
	"context"

	factskeys "github.com/osapi-io/osapi/internal/facts"
)

// descriptions are the descriptions of the built-in fact keys.
var descriptions = map[string]string{
	factskeys.KeyInterfacePrimary: "Primary network interface name",
	factskeys.KeyHostname:         "Agent hostname",
	factskeys.KeyArch:             "CPU architecture",
	factskeys.KeyKernel:           "Kernel version",
	factskeys.KeyFQDN:             "Fully qualified domain name",
	factskeys.KeyContainerized:    "Whether the agent is running inside a container",
}

// Facts is each built-in fact key that an input can refer to as @fact.<key>, with its description.
func Facts(
	_ context.Context,
	_ Deps,
	_ FactsInput,
) (FactKeysResponse, error) {
	builtIn := factskeys.BuiltInKeys()
	keys := make([]FactKeyEntry, 0, len(builtIn))
	for _, key := range builtIn {
		keys = append(keys, FactKeyEntry{
			Key:         key,
			Description: descriptions[key],
			Builtin:     true,
		})
	}

	return FactKeysResponse{Items: keys, TotalItems: len(keys)}, nil
}

// FactsInput is the input of facts, which has none.
type FactsInput struct{}

// FactKeysResponse is each fact key.
type FactKeysResponse struct {
	_          struct{}       `title:"Fact keys"`
	Items      []FactKeyEntry `                  json:"items"       doc:"Available fact keys for @fact. references."`
	TotalItems int            `                  json:"total_items" doc:"Total number of fact keys."`
}

// FactKeyEntry is one fact key.
type FactKeyEntry struct {
	Key         string `json:"key"                   doc:"The fact key, such as 'interface.primary'." label:"Reference" prefix:"@fact."`
	Description string `json:"description,omitempty" doc:"Human-readable description of the fact."`
	Builtin     bool   `json:"builtin,omitempty"     doc:"Whether this is a built-in fact key."                                         render:"-"`
}

// sampleFacts is the sample answer of the list example, which the docs page
// of the command shows as the CLI prints it.
var sampleFacts = FactKeysResponse{
	TotalItems: 3,
	Items: []FactKeyEntry{
		{Key: "hostname", Description: "The hostname of the agent.", Builtin: true},
		{
			Key:         "primary_interface",
			Description: "The interface of the default route.",
			Builtin:     true,
		},
		{Key: "datacenter", Description: "A custom fact from the config of the agent."},
	},
}
