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

// Package dsl is the list of osapi's DSLs: the behavior DSL, the chain rules,
// the endpoint DSL and the command DSL. Each DSL registers once, with what it
// shares with the others: the commands that teach and check it, its skill,
// its docs page, and, for a DSL whose definitions serve the API, its
// permissions and operations. The role rule, the SDK's permission constants
// and the UI's operation permissions read this list, and a test fails when a
// DSL lacks a part that the others have. Thus a change to what the DSLs
// share reaches each of them.
package dsl

import (
	"slices"
	"strings"
	"sync"

	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// The commands that every DSL has, under the command of its group. A DSL
// where one does not apply says why in Without.
const (
	CommandList     = "list"
	CommandGet      = "get"
	CommandMethod   = "method get"
	CommandSpec     = "spec get"
	CommandValidate = "validate"
	CommandCreate   = "create"
	CommandOpenAPI  = "openapi get"
)

// Commands are the shared commands, in the order that a page lists them.
func Commands() []string {
	return []string{
		CommandList, CommandGet, CommandMethod, CommandSpec,
		CommandValidate, CommandCreate, CommandOpenAPI,
	}
}

// Declared is a permission that a method of a definition needs.
type Declared struct {
	// Permission is the permission.
	Permission types.Permission
	// AdminOnly is true when the method is marked AdminOnly.
	AdminOnly bool
	// Named is true when the method names a permission that something else
	// declares, as an endpoint method's .Permission does.
	Named bool
	// Where is the definition and its method, for an error.
	Where string
}

// Operation is a method that the API serves: its operationId and the
// permission that it needs.
type Operation struct {
	ID         string
	Permission types.Permission
}

// DSL is one DSL and the parts that it shares with the others.
type DSL struct {
	// Name is the DSL, such as behavior.
	Name string
	// Command is the words of the command of its group, such as dsl
	// endpoint.
	Command []string
	// Skill is the dev skill that teaches it, such as build-a-behavior.
	Skill string
	// Page is its page under docs/docs/sidebar/development/dsl, which its
	// spec writes.
	Page string
	// Permissions are the permissions of each registered definition. Nil for
	// a DSL whose definitions serve no API.
	Permissions func() []Declared
	// Operations are the served methods of each registered definition. Nil
	// for a DSL whose definitions serve no API.
	Operations func() []Operation
	// Without are the shared commands that do not apply to the DSL, each
	// with why.
	Without map[string]string
}

// ServesAPI reports whether the DSL's definitions serve the API.
func (d DSL) ServesAPI() bool {
	return d.Permissions != nil
}

var (
	mu   sync.RWMutex
	dsls []DSL
)

// Register adds d to the DSLs. Each DSL's package calls it in its init.
func Register(
	d DSL,
) DSL {
	mu.Lock()
	defer mu.Unlock()
	dsls = append(dsls, d)

	return d
}

// All are the registered DSLs, by name.
func All() []DSL {
	mu.RLock()
	defer mu.RUnlock()
	out := slices.Clone(dsls)
	slices.SortStableFunc(out, func(a, b DSL) int { return strings.Compare(a.Name, b.Name) })

	return out
}
