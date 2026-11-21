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

import (
	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/behaviortest"
	"github.com/osapi-io/osapi/internal/guide"
)

// ListInput is the input of list, which has none.
type ListInput struct{}

// ListOutput is each chain.
type ListOutput struct {
	_     struct{}  `title:"Chains"`
	Items []Summary `               json:"items"`
}

// Summary is one behavior in a chain: the tree of the behaviors that it
// calls, and the behaviors that call it.
type Summary struct {
	Behavior string   `json:"behavior"`
	Internal bool     `json:"internal,omitempty"`
	Chain    string   `json:"chain"`
	Calls    []string `json:"calls"              none:"none"`
	UsedBy   []string `json:"used_by"            none:"none" label:"Used by"`
}

// GetInput names the behavior at the top of the chain.
type GetInput struct {
	Name string `json:"name" validate:"required" doc:"The behavior, from osapi behavior chain list"`
}

// GetOutput is one chain: its tree, and the calls that each example of the
// behavior makes, with what each sends and gets.
type GetOutput struct {
	_        struct{}                `title:"Chain"`
	Behavior string                  `              json:"behavior"`
	Chain    string                  `              json:"chain"`
	Calls    []string                `              json:"calls"    none:"none"`
	UsedBy   []string                `              json:"used_by"  none:"none" label:"Used by"`
	Examples []ExampleCalls          `              json:"examples"                             empty:"No example calls another behavior." table:"sections"`
	Runs     []behaviortest.ChainRun `              json:"runs"                                                                                             render:"-"`
}

// ExampleCalls is one example of a method, and the calls that it makes, each
// under the call that made it.
type ExampleCalls struct {
	Example string `json:"example"`
	Calls   []Call `json:"calls"`
}

// Call is one call of an example: what it sends and what it gets, as the
// fields that have a value. --json has each in full, under runs.
type Call struct {
	Call  string `json:"call"`
	Sends string `json:"sends"`
	Gets  string `json:"gets"`
}

// ValidateInput is the input of validate, which has none.
type ValidateInput struct{}

// ValidateOutput is each problem of each chain, and the fix of each rule
// that one breaks.
type ValidateOutput struct {
	_         struct{}           `title:"Validation"`
	Behaviors int                `                   json:"behaviors"`
	Valid     bool               `                   json:"valid"`
	Problems  []behavior.Problem `                   json:"problems,omitempty"`
	Fixes     map[string]any     `                   json:"fixes,omitempty"`
}

// SpecInput is the input of spec get, which has none.
type SpecInput struct{}

// SpecOutput is each kind of term of a chain, with its terms.
type SpecOutput struct {
	Kinds []guide.Section `json:"kinds" table:"sections"`
}
