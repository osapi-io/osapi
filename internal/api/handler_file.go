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
	"github.com/labstack/echo/v4"
	strictecho "github.com/oapi-codegen/runtime/strictmiddleware/echo"

	"github.com/retr0h/osapi/internal/api/file"
	fileGen "github.com/retr0h/osapi/internal/api/file/gen"
	"github.com/retr0h/osapi/internal/authtoken"
)

// GetFileHandler returns file handler for registration.
func (s *Server) GetFileHandler(
	objStore file.ObjectStoreManager,
) []func(e *echo.Echo) {
	var tokenManager TokenValidator = authtoken.New(s.logger)

	fileHandler := file.New(s.logger, objStore)

	strictHandler := fileGen.NewStrictHandler(
		fileHandler,
		[]fileGen.StrictMiddlewareFunc{
			func(handler strictecho.StrictEchoHandlerFunc, _ string) strictecho.StrictEchoHandlerFunc {
				return scopeMiddleware(
					handler,
					tokenManager,
					s.appConfig.API.Server.Security.SigningKey,
					fileGen.BearerAuthScopes,
					s.customRoles,
				)
			},
		},
	)

	return []func(e *echo.Echo){
		func(e *echo.Echo) {
			fileGen.RegisterHandlers(e, strictHandler)
		},
	}
}
