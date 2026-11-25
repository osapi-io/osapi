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

package sdk

import (
	"slices"

	"github.com/osapi-io/osapi/internal/guide"
	"github.com/osapi-io/osapi/internal/sdk/sdkspec"
	"github.com/osapi-io/osapi/pkg/sdk/client"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// Kind is what a row of list is: a type or a service.
type Kind string

// The kinds of a row of list.
const (
	// KindType is a type for values of pkg/sdk/types.
	KindType Kind = "type"
	// KindService is a service of the client.
	KindService Kind = "service"
)

// Values are the kinds of a row.
func (Kind) Values() []Kind {
	return []Kind{KindType, KindService}
}

// Valid reports whether k is a kind of a row.
func (k Kind) Valid() bool {
	return slices.Contains(k.Values(), k)
}

// ListInput is the input of list, which has none.
type ListInput struct{}

// Item is one row of list.
type Item struct {
	Kind  Kind   `json:"kind"`
	Name  string `json:"name"`
	Count int    `json:"count" doc:"The values of a type, or the methods of a service"`
	Doc   string `json:"doc"                                                           render:"short=60"`
}

// ListOutput is each type for values, then each service.
type ListOutput struct {
	_     struct{} `title:"SDK"`
	Items []Item   `            json:"items"`
}

// GetInput names a type or a service.
type GetInput struct {
	Name string `json:"name" validate:"required" doc:"A type, such as ContentType, or a service, such as Hostname"`
}

// GetOutput is the type or the service of the name.
type GetOutput struct {
	_       struct{}            `title:"SDK"`
	Type    *types.TypeSpec     `            json:"type,omitempty"`
	Service *client.ServiceSpec `            json:"service,omitempty"`
}

// SpecInput is the input of spec get, which has none.
type SpecInput struct{}

// SpecOutput is each kind of term of the SDK, with its terms.
type SpecOutput struct {
	Kinds []guide.Section `json:"kinds" table:"sections"`
}

// ValidateInput is the input of validate, which has none.
type ValidateInput struct{}

// ValidateOutput is each problem of the SDK, and the fix of each rule that
// one breaks.
type ValidateOutput struct {
	_        struct{}          `title:"Validation"`
	Types    int               `                   json:"types"`
	Services int               `                   json:"services"`
	Valid    bool              `                   json:"valid"`
	Problems []sdkspec.Problem `                   json:"problems,omitempty"`
	Fixes    map[string]any    `                   json:"fixes,omitempty"`
}
