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
	"github.com/osapi-io/osapi/internal/sdk/sdkspec"
	"github.com/osapi-io/osapi/pkg/sdk/client"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// The outputs that the docs pages show, one for each method that prints.
var (
	sampleList = ListOutput{Items: []Item{
		{
			Kind:  KindType,
			Name:  "ContentType",
			Count: 2,
			Doc:   "ContentType is how a deploy treats a file.",
		},
		{
			Kind:  KindService,
			Name:  "Hostname",
			Count: 2,
			Doc:   "Read and set the hostname of each node.",
		},
	}}

	sampleGet = GetOutput{Service: &client.ServiceSpec{
		Name: "Hostname",
		Doc:  "Read and set the hostname of each node.",
		Area: "system",
		Node: true,
		Methods: []client.MethodSpec{{
			Name: "Update", Doc: "Set the hostname of the host", Permission: "hostname:write",
			HTTP: "PUT /api/node/{hostname}/hostname", Input: "HostnameUpdateInput",
			Result: "Collection[HostnameResult]", Examples: 2,
		}},
	}}

	sampleSpec = SpecOutput{Kinds: sdkspec.Spec(nil, nil).Sections()[:1]}

	sampleValidate = ValidateOutput{Types: len(types.All()), Services: 1, Valid: true}
)
