// Copyright (c) 2024 John Dewey

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

package authtoken

import "context"

// permissionsKey is the context key of the permissions of a token.
type permissionsKey struct{}

// WithPermissions is ctx with the permissions that the token of the
// request resolves to, from its roles or its own permissions.
func WithPermissions(
	ctx context.Context,
	resolved map[Permission]bool,
) context.Context {
	return context.WithValue(ctx, permissionsKey{}, resolved)
}

// Granted reports whether the token of ctx has the permission p. A ctx
// with no token, as on a public route, has no permission.
func Granted(
	ctx context.Context,
	p Permission,
) bool {
	resolved, _ := ctx.Value(permissionsKey{}).(map[Permission]bool)

	return HasPermission(resolved, p)
}
