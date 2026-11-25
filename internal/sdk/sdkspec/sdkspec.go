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

// Package sdkspec is the Go SDK as data: its spec, which osapi sdk spec get
// prints and the SDK skill is made from, and its checks, which osapi sdk
// validate runs. It reads the SDK's own description of itself, types.All
// and client.Services, as osapi agent collector reads gohai.Collectors.
package sdkspec

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/osapi-io/osapi/internal/guide"
	"github.com/osapi-io/osapi/internal/guide/sdktypes"
	"github.com/osapi-io/osapi/pkg/sdk/client"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// The kinds of the SDK's spec.
const (
	// KindPackage is a package of pkg/sdk.
	KindPackage = "package"
	// KindCall is how a call is made and how it fails.
	KindCall = "call"
	// KindService is a service of the client.
	KindService = "service"
)

// Spec is the SDK: its packages, how a call is made, its types for values
// and the services of its client.
func Spec(
	ts []types.TypeSpec,
	ss []client.ServiceSpec,
) guide.Spec {
	s := guide.Spec{
		Kinds: []guide.Kind{
			{Kind: KindPackage, Title: "Packages"},
			{Kind: KindCall, Title: "Calls"},
		},
		Terms: []guide.Term{
			{
				Kind: KindPackage,
				Name: "github.com/osapi-io/osapi/pkg/sdk/client",
				Does: "Calls the API of the controller: client.New(url, token) is a Client, and each service of the API is a field of it, such as c.Hostname.",
				When: "In a program that calls osapi. The services below are its fields.",
			},
			{
				Kind: KindPackage,
				Name: "github.com/osapi-io/osapi/pkg/sdk/types",
				Does: "The vocabulary of the API: the permissions, the roles, the targets and each closed set of values, as types with constants.",
				When: "For each value of a closed set: a constant such as types.ContentRaw, never a string.",
			},
			{
				Kind: KindCall,
				Name: "c.<Service>.<Method>(ctx, ...)",
				Does: "Sends one request and answers with a *client.Response[T] of the SDK's types. A method of a node takes the target after ctx: a hostname, a label such as group:web, string(types.TargetAll) or string(types.TargetAny), and its T is a Collection of a row for each host.",
				When: "For each call. osapi sdk get --name <service> lists the methods, with the permission each needs.",
			},
			{
				Kind: KindCall,
				Name: "error",
				Does: "A call that the API refuses fails with a typed error: *client.APIError has the StatusCode and the Reason (types.Reason), and NotFoundError, AuthError, ValidationError, ConflictError and ServerError wrap it. A row of a Collection has its own status (types.RowStatus), so one host can fail while the others succeed.",
				When: "Branch with errors.As on the error type and on its Reason, never on the message; then check the status of each row.",
			},
		},
	}
	s = sdktypes.AppendOf(s, ts)
	s.Kinds = append(s.Kinds, guide.Kind{Kind: KindService, Title: "Services"})
	for _, svc := range ss {
		methods := make([]string, 0, len(svc.Methods))
		for _, m := range svc.Methods {
			methods = append(methods, m.Name)
		}
		s.Terms = append(s.Terms, guide.Term{
			Kind: KindService,
			Name: "c." + svc.Name,
			Does: svc.Doc + " Methods: " + strings.Join(methods, ", ") + ".",
			When: "osapi sdk get --name " + svc.Name + " shows each method with its permission, route, input and result.",
		})
	}

	return s
}

// The rules of Checks.
const (
	RuleDoc     = "doc"
	RuleValues  = "values"
	RuleExample = "example"
	RuleClient  = "client"
)

// Checks are the rules that osapi sdk validate applies.
var Checks = guide.Checks{
	{
		Rule:     RuleDoc,
		Requires: "Each type, each service and each method of the SDK says what it is.",
		Fix:      "Give the type a doc comment, or the schema or the operation of the spec a description, and run just generate.",
	},
	{
		Rule:     RuleValues,
		Requires: "Each type for values has at least one value.",
		Fix:      "Give the type constants and list them in its Values.",
	},
	{
		Rule:     RuleExample,
		Requires: "Each method of the client has an example, which examples/sdk/client runs.",
		Fix:      "Add an Example to the method's definition, and run just generate.",
	},
	{
		Rule:     RuleClient,
		Requires: "Each service that client.Services lists is a field of client.Client, and each of its methods is a method of that field.",
		Fix:      "Run just generate: client.Services and the client come from one run.",
	},
}

// Problem is one thing of the SDK that breaks a rule of Checks.
type Problem struct {
	// Name is the type, the service or the method, such as Hostname.Update.
	Name string `json:"name"`
	// Rule is the rule, from Checks.
	Rule string `json:"rule"`
	// Message says what is wrong.
	Message string `json:"message"`
	// Fix is how to fix it, from the check of Rule.
	Fix string `json:"fix" table:"-"`
}

// Validate is each problem of ts and ss against Checks.
func Validate(
	ts []types.TypeSpec,
	ss []client.ServiceSpec,
) []Problem {
	var out []Problem
	add := func(name, rule, msg string) {
		out = append(out, Problem{Name: name, Rule: rule, Message: msg, Fix: Checks.Fix(rule)})
	}
	for _, t := range ts {
		if t.Doc == "" {
			add(t.Name, RuleDoc, "the type has no doc")
		}
		if len(t.Values) == 0 {
			add(t.Name, RuleValues, "the type has no values")
		}
	}
	c := reflect.TypeFor[client.Client]()
	for _, svc := range ss {
		if svc.Doc == "" {
			add(svc.Name, RuleDoc, "the service has no doc")
		}
		f, ok := c.FieldByName(svc.Name)
		if !ok {
			add(svc.Name, RuleClient, "client.Client has no field "+svc.Name)
		}
		for _, m := range svc.Methods {
			name := svc.Name + "." + m.Name
			if m.Doc == "" {
				add(name, RuleDoc, "the method has no doc")
			}
			if m.Examples == 0 {
				add(name, RuleExample, "the method has no example")
			}
			if ok {
				if _, has := f.Type.MethodByName(m.Name); !has {
					add(
						name,
						RuleClient,
						fmt.Sprintf("client.%s has no method %s", svc.Name, m.Name),
					)
				}
			}
		}
	}

	return out
}
