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

import "time"

// GenerateInput is what a new token carries.
type GenerateInput struct {
	Roles       []string `json:"roles"       short:"r" validate:"required,dive,oneof=admin read write" doc:"Roles for the token"`
	Subject     string   `json:"subject"     short:"u" validate:"required"                             doc:"Subject for the token (e.g., user ID or unique identifier)"`
	Permissions []string `json:"permissions" short:"p" validate:"omitempty,dive,permission"            doc:"Direct permissions, which override the roles; osapi token roles list lists them"`
}

// GenerateOutput is a new token, with what it carries.
type GenerateOutput struct {
	Token       string   `json:"token"`
	Subject     string   `json:"subject"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions,omitempty"`
}

// RolesInput is the input of roles list, which has none.
type RolesInput struct{}

// Role is one role, with what it grants.
type Role struct {
	// Name is the role's name, such as admin.
	Name string `json:"name"`
	// Source is built-in, or config for a custom role of the config.
	Source string `json:"source"`
	// Permissions are what the role grants.
	Permissions []string `json:"permissions"`
}

// RolesOutput is each role: the built-in ones, then those of the config.
type RolesOutput struct {
	Roles []Role `json:"roles"`
}

// ValidateInput is the token to check.
type ValidateInput struct {
	Token string `json:"token" short:"t" validate:"required" doc:"The Token string"`
}

// Claims are what a valid token carries.
type Claims struct {
	Subject     string   `json:"subject"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions,omitempty"`
	// EffectivePermissions are what the roles and the permissions grant,
	// sorted.
	EffectivePermissions []string  `json:"effective_permissions"`
	Audience             []string  `json:"audience"`
	IssuedAt             time.Time `json:"issued_at,omitzero"`
	ExpiresAt            time.Time `json:"expires_at,omitzero"`
}
