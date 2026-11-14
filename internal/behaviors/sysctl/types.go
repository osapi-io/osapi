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

package sysctl

// The input and output types. Their tags are the behavior's contract:
//
//   - json names the field everywhere: the request body, the CLI flag, the
//     table column and the --json output.
//   - validate is checked by the API before a job is queued, and becomes the
//     OpenAPI schema's constraints.
//   - doc is the field's description in the spec, the flag's help and the
//     docs pages.
//   - render says how an output field shows in a table: "-" to hide it,
//     bytes, age or status.
//   - telemetry:"attr" records an input field on the method's span. Only
//     tagged fields are recorded.

// ListInput optionally narrows the list to one part of the tree.
type ListInput struct {
	Prefix string `json:"prefix" validate:"omitempty,sysctl_key" doc:"Only this parameter and the ones under it, such as net.ipv4"`
}

// GetInput names one parameter.
type GetInput struct {
	Key string `json:"key" validate:"required,min=1,sysctl_key" doc:"Parameter name, such as net.ipv4.ip_forward" telemetry:"attr"`
}

// CreateInput is a parameter and the value to set.
type CreateInput struct {
	Key   string `json:"key"   validate:"required,min=1,sysctl_key" doc:"Parameter name, such as net.ipv4.ip_forward" telemetry:"attr"`
	Value string `json:"value" validate:"required,no_linebreak"     doc:"Value to set"`
}

// UpdateInput is a parameter and its new value.
type UpdateInput struct {
	Key   string `json:"key"   validate:"required,min=1,sysctl_key" doc:"Parameter name, such as net.ipv4.ip_forward" telemetry:"attr"`
	Value string `json:"value" validate:"required,no_linebreak"     doc:"Value to set"`
}

// DeleteInput names the parameter to stop managing.
type DeleteInput struct {
	Key string `json:"key" validate:"required,min=1,sysctl_key" doc:"Parameter name, such as net.ipv4.ip_forward" telemetry:"attr"`
}

// Entry is a parameter, its current runtime value, and whether osapi
// manages it.
type Entry struct {
	Key     string `json:"key"     doc:"Parameter name"`
	Value   string `json:"value"   doc:"Current runtime value"`
	Managed bool   `json:"managed" doc:"Whether osapi manages it: a file of its own sets it at boot"`
}

// MutationResult says whether a write changed the host.
type MutationResult struct {
	Key     string `json:"key"     doc:"Parameter name"`
	Changed bool   `json:"changed" doc:"Whether the host changed"`
}

// CreateResult, UpdateResult and DeleteResult are what the provider returns,
// and what the hand-written sysctl domain's agent puts on the wire until its
// cutover.

// CreateResult represents the outcome of a sysctl create operation.
type CreateResult struct {
	Key     string `json:"key"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// UpdateResult represents the outcome of a sysctl update operation.
type UpdateResult struct {
	Key     string `json:"key"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// DeleteResult represents the outcome of a sysctl delete operation.
type DeleteResult struct {
	Key     string `json:"key"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}
