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

// Package server is the runtime the generated routes of an endpoint group
// call. Each group's gen/api.go lists its routes and registers them here;
// each gen/api_<method>.go decodes the method's typed input and hands it to
// Serve, which runs the method's handler and maps its error to a status.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"strings"

	"github.com/labstack/echo/v5"

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/controller/api"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/internal/telemetry/respspan"
)

// scopesKey is the context key ScopeMiddleware reads the required
// permissions from. It matches the key oapi-codegen's wrappers set.
const scopesKey = "BearerAuth.Scopes"

// ErrorResponse is the body of every answer that is not a success.
type ErrorResponse struct {
	Error string `json:"error"`
	// Reason is why, as one of endpoint.Reasons.
	Reason endpoint.Reason `json:"reason"`
}

// Route is one generated endpoint: its HTTP method, its echo path, the
// permission a token needs ("" for a public one), and the handler.
type Route struct {
	Method     string
	Path       string
	Permission string
	Handler    echo.HandlerFunc
}

// Server is what every generated route shares: what it takes to check a
// token, and the dependencies of each group, by type.
type Server struct {
	logger      *slog.Logger
	tokens      api.TokenValidator
	signingKey  string
	customRoles authtoken.CustomRoles
	deps        map[reflect.Type]any
}

// New returns the server the generated routes run on. deps are the values
// each group Uses, one of each type.
func New(
	logger *slog.Logger,
	signingKey string,
	customRoles authtoken.CustomRoles,
	deps ...any,
) *Server {
	s := &Server{
		logger:      logger.With(slog.String("subsystem", "controller.endpoint")),
		tokens:      authtoken.New(logger),
		signingKey:  signingKey,
		customRoles: customRoles,
		deps:        make(map[reflect.Type]any, len(deps)),
	}
	for _, d := range deps {
		s.deps[reflect.TypeOf(d)] = d
	}

	return s
}

// Deps is the value of type D that the server was given. A group whose
// dependencies the controller did not give is a mistake in the wiring, so
// it panics at startup, when the routes are made. An empty struct needs no
// wiring: a new group starts with one, and its first field makes the
// controller give it.
func Deps[D any](
	s *Server,
) D {
	d, ok := s.deps[reflect.TypeFor[D]()].(D)
	if t := reflect.TypeFor[D](); !ok && t.Kind() == reflect.Struct && t.NumField() == 0 {
		return d
	}
	if !ok {
		panic(fmt.Sprintf("endpoint: no %s given to the server", reflect.TypeFor[D]()))
	}

	return d
}

// routers are the generated route lists, one per group, in the order their
// gen packages registered.
var routers []func(s *Server) []Route

// Register adds one group's generated routes. Each gen/api.go calls it from
// init.
func Register(
	routes func(s *Server) []Route,
) {
	routers = append(routers, routes)
}

// Handler returns the route registrations for every registered group.
func Handler(
	logger *slog.Logger,
	signingKey string,
	customRoles authtoken.CustomRoles,
	deps ...any,
) []func(e *echo.Echo) {
	s := New(logger, signingKey, customRoles, deps...)

	return []func(e *echo.Echo){
		func(e *echo.Echo) {
			for _, routes := range routers {
				s.Mount(e, routes(s))
			}
		},
	}
}

// Mount adds routes to e, each behind the check that the request's token
// grants its permission. A route with no permission is public.
func (s *Server) Mount(
	e *echo.Echo,
	routes []Route,
) {
	for _, r := range routes {
		if r.Permission == "" {
			e.Add(r.Method, r.Path, r.Handler)

			continue
		}
		e.Add(r.Method, r.Path, s.authorize(r))
	}
}

// authorize wraps a route's handler in the token check.
func (s *Server) authorize(
	r Route,
) echo.HandlerFunc {
	run := api.ScopeMiddleware(
		func(c *echo.Context, _ any) (any, error) {
			return nil, r.Handler(c)
		},
		s.tokens,
		s.signingKey,
		scopesKey,
		s.customRoles,
	)

	return func(c *echo.Context) error {
		c.Set(scopesKey, []string{r.Permission})
		_, err := run(c, nil)

		return err
	}
}

// Decoder reads a method's typed input from the request, and says what is
// wrong when it cannot.
type Decoder[In any] func(c *echo.Context) (In, string, bool)

// Serve is the handler of one method: it decodes the input, runs the
// method's handler with the group's dependencies, and writes the output
// with the method's success status, or the error with the status that
// Errors maps it to, or 500.
func Serve[D, In, Out any](
	s *Server,
	g *endpoint.Group,
	method string,
	run func(context.Context, D, In) (Out, error),
	decode Decoder[In],
) echo.HandlerFunc {
	m := g.Method(method)
	deps := Deps[D](s)

	return func(c *echo.Context) error {
		in, msg, ok := decode(c)
		if !ok {
			return c.JSON(http.StatusBadRequest,
				ErrorResponse{Error: msg, Reason: endpoint.ReasonInvalidRequest})
		}

		out, err := run(c.Request().Context(), deps, in)
		var answer *endpoint.AnswerError
		if errors.As(err, &answer) {
			respspan.Handled(c)

			return c.JSON(answer.Code, out)
		}
		if err != nil {
			code, reason := status(s, g, m, err)

			return c.JSON(code, ErrorResponse{Error: err.Error(), Reason: reason})
		}

		respspan.Handled(c)
		if f := m.LocationField(); f != "" {
			c.Response().Header().Set("Location", "/api/"+g.Name+"/"+field(out, f))
		}
		if m.SuccessStatus() == http.StatusNoContent {
			return c.NoContent(http.StatusNoContent)
		}

		return c.JSON(m.SuccessStatus(), out)
	}
}

// status is the code and the reason that an error maps to, by errors.Is,
// or 500, which is logged: it is a failure of the controller.
func status(
	s *Server,
	g *endpoint.Group,
	m *endpoint.Method,
	err error,
) (int, endpoint.Reason) {
	for _, e := range m.Mapped() {
		if errors.Is(err, e.Err) {
			return e.Code, e.Reason
		}
	}
	cause := err.Error()
	var internal *endpoint.InternalError
	if errors.As(err, &internal) && internal.Cause != nil {
		cause = internal.Cause.Error()
	}
	s.logger.Error("endpoint failed",
		slog.String("group", g.Name),
		slog.String("method", m.Name),
		slog.String("error", cause))

	return http.StatusInternalServerError, endpoint.ReasonServerError
}

// Path is the echo path of a method: its route, with {param} written as
// :param.
func Path(
	g *endpoint.Group,
	m *endpoint.Method,
) string {
	p := strings.ReplaceAll(g.Path(m), "{", ":")

	return strings.ReplaceAll(p, "}", "")
}

// field is the value of the field of out whose json name is name, as text.
func field(
	out any,
	name string,
) string {
	raw, _ := json.Marshal(out)
	var fields map[string]any
	_ = json.Unmarshal(raw, &fields)

	return fmt.Sprint(fields[name])
}
