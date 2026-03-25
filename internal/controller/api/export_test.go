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

package api

import (
	"log/slog"

	"github.com/labstack/echo/v4"
	strictecho "github.com/oapi-codegen/runtime/strictmiddleware/echo"

	"github.com/retr0h/osapi/internal/audit"
	"github.com/retr0h/osapi/internal/authtoken"
)

// ExportAuditMiddleware exposes the private auditMiddleware for testing.
func ExportAuditMiddleware(
	store audit.Store,
	logger *slog.Logger,
) echo.MiddlewareFunc {
	return auditMiddleware(store, logger)
}

// ExportScopeMiddleware exposes the private scopeMiddleware for testing.
func ExportScopeMiddleware(
	next strictecho.StrictEchoHandlerFunc,
	tokenManager *authtoken.Token,
	signingKey string,
	contextKey string,
	customRoles map[string][]string,
) strictecho.StrictEchoHandlerFunc {
	return scopeMiddleware(next, tokenManager, signingKey, contextKey, customRoles)
}
