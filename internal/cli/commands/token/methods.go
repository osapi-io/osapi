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
	"fmt"
	"sort"

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/pkg/sdk/client"
)

// Generate signs a new token with the config's signing key.
func Generate(
	_ context.Context,
	d Deps,
	in GenerateInput,
) (GenerateOutput, error) {
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
	for _, name := range []string{client.RoleAdmin, client.RoleWrite, client.RoleRead} {
		perms, _ := authtoken.RolePermissions(name)
		roles = append(roles, Role{Name: name, Source: "built-in", Permissions: perms})
	}
	names := make([]string, 0, len(custom))
	for name := range custom {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		roles = append(roles, Role{Name: name, Source: "config", Permissions: custom[name]})
	}

	return RolesOutput{Roles: roles}, nil
}

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

	resolved := authtoken.ResolvePermissions(claims.Roles, claims.Permissions, nil)
	effective := make([]string, 0, len(resolved))
	for p := range resolved {
		effective = append(effective, p)
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
