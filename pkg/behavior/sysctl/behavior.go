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

// Package sysctl is the sysctl behavior: kernel parameters osapi writes to
// /etc/sysctl.d and applies.
package sysctl

import (
	"context"

	prov "github.com/osapi-io/osapi/internal/provider/node/sysctl"
	"github.com/osapi-io/osapi/pkg/behavior"
)

// Behavior is the sysctl definition. Everything else, the routes, the agent's
// dispatch, the CLI commands, the spec and the docs, is built from it.
var Behavior = behavior.Register(behavior.Define(
	"sysctl", "Kernel parameters managed under /etc/sysctl.d",
	behavior.Node,
	behavior.Read("list", List).
		Describe("List the kernel parameters osapi manages").
		Example("Every managed parameter", ListInput{}, []Entry{
			{Key: "net.ipv4.ip_forward", Value: "1"},
			{Key: "vm.swappiness", Value: "10"},
		}),
	behavior.Read("get", Get).Key("key").
		Describe("Get a kernel parameter's current value").
		Example("One parameter", GetInput{Key: "vm.swappiness"},
			Entry{Key: "vm.swappiness", Value: "10"}),
	behavior.Write("create", Create).
		Describe("Manage a kernel parameter and apply it").
		Example("Turn on IP forwarding",
			CreateInput{Key: "net.ipv4.ip_forward", Value: "1"},
			Result{Key: "net.ipv4.ip_forward", Changed: true}),
	behavior.Write("update", Update).Key("key").
		Describe("Change a managed kernel parameter and apply it").
		Example("Lower swappiness",
			UpdateInput{Key: "vm.swappiness", Value: "10"},
			Result{Key: "vm.swappiness", Changed: true}),
	behavior.Write("delete", Delete).Key("key").
		Describe("Stop managing a kernel parameter and restore its default").
		Example("Remove a parameter", DeleteInput{Key: "vm.swappiness"},
			Result{Key: "vm.swappiness", Changed: true}),
))

// Provider is what each platform implements.
type Provider = prov.Provider

// ListInput takes nothing beyond the target.
type ListInput struct{}

// GetInput names one parameter.
type GetInput struct {
	Key string `json:"key" validate:"required,min=1,sysctl_key" doc:"Parameter name, such as net.ipv4.ip_forward"`
}

// CreateInput is a parameter and the value to set.
type CreateInput struct {
	Key   string `json:"key"   validate:"required,min=1,sysctl_key" doc:"Parameter name, such as net.ipv4.ip_forward"`
	Value string `json:"value" validate:"required,no_linebreak"     doc:"Value to set"`
}

// UpdateInput is a managed parameter and its new value.
type UpdateInput struct {
	Key   string `json:"key"   validate:"required,min=1,sysctl_key" doc:"Parameter name, such as net.ipv4.ip_forward"`
	Value string `json:"value" validate:"required,no_linebreak"     doc:"Value to set"`
}

// DeleteInput names the parameter to stop managing.
type DeleteInput struct {
	Key string `json:"key" validate:"required,min=1,sysctl_key" doc:"Parameter name, such as net.ipv4.ip_forward"`
}

// Entry is a parameter and its current runtime value.
type Entry struct {
	Key   string `json:"key"   doc:"Parameter name"`
	Value string `json:"value" doc:"Current runtime value"`
}

// Result says whether a write changed the host.
type Result struct {
	Key     string `json:"key"     doc:"Parameter name"`
	Changed bool   `json:"changed" doc:"Whether the host changed"`
}

// List returns every managed parameter.
func List(
	ctx context.Context,
	p Provider,
	_ ListInput,
) ([]Entry, error) {
	entries, err := p.List(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]Entry, 0, len(entries))
	for _, e := range entries {
		out = append(out, Entry{Key: e.Key, Value: e.Value})
	}

	return out, nil
}

// Get returns one parameter.
func Get(
	ctx context.Context,
	p Provider,
	in GetInput,
) (Entry, error) {
	e, err := p.Get(ctx, in.Key)
	if err != nil {
		return Entry{}, err
	}

	return Entry{Key: e.Key, Value: e.Value}, nil
}

// Create manages a parameter and applies it.
func Create(
	ctx context.Context,
	p Provider,
	in CreateInput,
) (Result, error) {
	r, err := p.Create(ctx, prov.Entry{Key: in.Key, Value: in.Value})
	if err != nil {
		return Result{}, err
	}

	return Result{Key: r.Key, Changed: r.Changed}, nil
}

// Update changes a managed parameter and applies it.
func Update(
	ctx context.Context,
	p Provider,
	in UpdateInput,
) (Result, error) {
	r, err := p.Update(ctx, prov.Entry{Key: in.Key, Value: in.Value})
	if err != nil {
		return Result{}, err
	}

	return Result{Key: r.Key, Changed: r.Changed}, nil
}

// Delete stops managing a parameter.
func Delete(
	ctx context.Context,
	p Provider,
	in DeleteInput,
) (Result, error) {
	r, err := p.Delete(ctx, in.Key)
	if err != nil {
		return Result{}, err
	}

	return Result{Key: r.Key, Changed: r.Changed}, nil
}
