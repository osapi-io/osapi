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

package client

// ServiceSpec is one service of the client, such as Hostname: what it does
// and each of its methods. Services lists them, so a program or an agent
// reads what the client can call from the client itself.
type ServiceSpec struct {
	// Name is the field of the service on Client, such as Hostname.
	Name string `json:"name"`
	// Doc is what the service does.
	Doc string `json:"doc"`
	// Area is the area of the API it is in, such as node.
	Area string `json:"area,omitempty"`
	// Node is true for a service of a node, whose methods take a target.
	Node bool `json:"node"`
	// Methods are its methods.
	Methods []MethodSpec `json:"methods"`
}

// MethodSpec is one method of a service.
type MethodSpec struct {
	// Name is the Go name of the method, such as Update.
	Name string `json:"name"`
	// Doc is what the method does.
	Doc string `json:"doc"`
	// Permission is the permission that the call needs, such as
	// hostname:write.
	Permission string `json:"permission"`
	// HTTP is the method and the path that the call sends, such as
	// PUT /node/{hostname}/hostname.
	HTTP string `json:"http"`
	// Input is the Go type of its request body, empty when it has none.
	Input string `json:"input,omitempty"`
	// Result is the Go type of its answer, such as Collection[HostnameResult].
	Result string `json:"result,omitempty"`
	// Examples is the number of examples of the call in the SDK's examples.
	Examples int `json:"examples"`
	// Program is a method that a program calls, not a person, such as the
	// agent's join. It has no example, which would run it against a
	// controller.
	Program bool `json:"program,omitempty"`
}
