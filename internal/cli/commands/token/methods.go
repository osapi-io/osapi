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

package token

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"

	"github.com/osapi-io/osapi/internal/authtoken"
)

// Generate signs a new token with the config's signing key.
func Generate(
	_ context.Context,
	d Deps,
	in GenerateInput,
) (GenerateOutput, error) {
	// A role is built in, or a custom role of the config.
	custom, err := d.Config.CustomRoles()
	if err != nil {
		return GenerateOutput{}, err
	}
	for _, r := range in.Roles {
		if _, ok := custom[authtoken.Role(r)]; !ok &&
			!slices.Contains(authtoken.Roles(), authtoken.Role(r)) {
			return GenerateOutput{}, fmt.Errorf("role %q: %w", r, ErrNoRole)
		}
	}

	token, err := d.Tokens.Generate(d.Config.SigningKey(), in.Roles, in.Subject, in.Permissions)
	if err != nil {
		return GenerateOutput{}, fmt.Errorf("failed to generate token: %w", err)
	}

	return GenerateOutput{
		Token:       token,
		Subject:     in.Subject,
		Roles:       in.Roles,
		Permissions: in.Permissions,
	}, nil
}

// Roles is each built-in role, then each custom role of the config, by
// name, with what each grants.
func Roles(
	_ context.Context,
	d Deps,
	_ RolesInput,
) (RolesOutput, error) {
	custom, err := d.Config.CustomRoles()
	if err != nil {
		return RolesOutput{}, err
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

	return RolesOutput{Roles: roles}, nil
}

// RoleGet is one role, built-in or of the config, with what it grants.
func RoleGet(
	ctx context.Context,
	d Deps,
	in RoleInput,
) (Role, error) {
	all, err := Roles(ctx, d, RolesInput{})
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

// Validate checks a token against the config's signing key, and is its
// claims.
func Validate(
	_ context.Context,
	d Deps,
	in ValidateInput,
) (Claims, error) {
	claims, err := d.Tokens.Validate(in.Token, d.Config.SigningKey())
	if err != nil {
		return Claims{}, fmt.Errorf("failed to validate token: %w", err)
	}

	resolved := authtoken.ResolvePermissions(
		authtoken.RolesOf(claims.Roles),
		authtoken.PermissionsOf(claims.Permissions),
		nil,
	)
	effective := make([]string, 0, len(resolved))
	for p := range resolved {
		effective = append(effective, string(p))
	}
	sort.Strings(effective)

	out := Claims{
		Subject:              claims.Subject,
		Roles:                claims.Roles,
		Permissions:          claims.Permissions,
		EffectivePermissions: effective,
		Audience:             claims.Audience,
	}
	if claims.IssuedAt != nil {
		out.IssuedAt = claims.IssuedAt.Time
	}
	if claims.ExpiresAt != nil {
		out.ExpiresAt = claims.ExpiresAt.Time
	}

	return out, nil
}
