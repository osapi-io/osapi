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
	"time"

	"github.com/labstack/echo/v4"
	strictecho "github.com/oapi-codegen/runtime/strictmiddleware/echo"

	"github.com/retr0h/osapi/internal/api/health"
	healthGen "github.com/retr0h/osapi/internal/api/health/gen"
	"github.com/retr0h/osapi/internal/authtoken"
)

// unauthenticatedOperations lists operation IDs that skip auth.
var unauthenticatedOperations = map[string]bool{
	"GetHealth":      true,
	"GetHealthReady": true,
}

// GetHealthHandler returns health handler for registration.
func (s *Server) GetHealthHandler(
	checker health.Checker,
	startTime time.Time,
	version string,
	metrics health.MetricsProvider,
) []func(e *echo.Echo) {
	var tokenManager TokenValidator = authtoken.New(s.logger)

	healthHandler := health.New(s.logger, checker, startTime, version, metrics)

	strictHandler := healthGen.NewStrictHandler(
		healthHandler,
		[]healthGen.StrictMiddlewareFunc{
			func(handler strictecho.StrictEchoHandlerFunc, operationID string) strictecho.StrictEchoHandlerFunc {
				if unauthenticatedOperations[operationID] {
					return handler
				}

				return scopeMiddleware(
					handler,
					tokenManager,
					s.appConfig.API.Server.Security.SigningKey,
					healthGen.BearerAuthScopes,
					s.customRoles,
				)
			},
		},
	)

	return []func(e *echo.Echo){
		func(e *echo.Echo) {
			healthGen.RegisterHandlers(e, strictHandler)
		},
	}
}
