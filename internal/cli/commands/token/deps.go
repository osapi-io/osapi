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

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/osapi-io/osapi/internal/authtoken"
)

// Deps is what the token group's methods use. The CLI gives it with
// command.Provide in cmd/commands.go; a test gives fakes.
type Deps struct {
	// Tokens signs and checks tokens.
	Tokens Tokens
	// Config is what the methods read from the osapi config.
	Config Config
}

// Tokens signs and checks tokens. *authtoken.Token is one.
type Tokens interface {
	Generate(
		key authtoken.Key,
		roles []types.Role,
		subject string,
		permissions []string,
		targets []string,
	) (string, error)
	Validate(
		tokenString string,
		keys authtoken.Keys,
	) (*authtoken.CustomClaims, error)
}

// Config is what the methods read from the osapi config.
type Config interface {
	// SigningKeys are controller.api.security.signing_keys. The first
	// signs. A key that is a reference, secret://<vault>/<key>, is
	// resolved from the vaults of the config.
	SigningKeys(ctx context.Context) (authtoken.Keys, error)
	// CustomRoles are the roles of controller.api.security.roles, each with
	// its permissions. Without a config file there are none.
	CustomRoles() (authtoken.CustomRoles, error)
}
