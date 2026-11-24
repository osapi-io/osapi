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

package role

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/osapi-io/osapi/internal/authtoken"
)

// List is each built-in role, then each custom role of the config, by
// name, with what each grants.
func List(
	_ context.Context,
	d Deps,
	_ ListInput,
) (ListOutput, error) {
	custom, err := d.Config.CustomRoles()
	if err != nil {
		return ListOutput{}, err
	}

	roles := make([]Role, 0, 3+len(custom))
	for _, name := range authtoken.Roles() {
		perms, _ := authtoken.RolePermissions(name)
		roles = append(
			roles,
			Role{Name: string(name), Source: "built-in", Permissions: authtoken.Strings(perms)},
		)
	}
	names := make([]string, 0, len(custom))
	for name := range custom {
		names = append(names, string(name))
	}
	sort.Strings(names)
	for _, name := range names {
		roles = append(roles, Role{
			Name:        name,
			Source:      "config",
			Permissions: authtoken.Strings(custom[authtoken.Role(name)]),
		})
	}

	return ListOutput{Roles: roles}, nil
}

// Get is one role, built-in or of the config, with what it grants.
func Get(
	ctx context.Context,
	d Deps,
	in GetInput,
) (Role, error) {
	all, err := List(ctx, d, ListInput{})
	if err != nil {
		return Role{}, err
	}
	for _, r := range all.Roles {
		if r.Name == in.Name {
			return r, nil
		}
	}

	return Role{}, fmt.Errorf("role %q: %w", in.Name, ErrNoRole)
}

// ErrNoRole is a role that neither the built-in roles nor the config have.
var ErrNoRole = errors.New("no such role")
