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

// Package behavior is the runtime the generated API handlers call. Each
// behavior's gen/api.go lists its routes and registers them here; each
// gen/api_<method>.go decodes the method's typed input and hands it to Serve,
// which queues the job and shapes the per-host results.
package behavior

import (
	"context"
	"log/slog"
	"strings"

	"github.com/labstack/echo/v5"

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/controller/api"
	"github.com/osapi-io/osapi/internal/job"
)

// scopesKey is the context key ScopeMiddleware reads the required
// permissions from. It matches the key oapi-codegen's wrappers set.
const scopesKey = "BearerAuth.Scopes"

// Response is the envelope every behavior method returns.
type Response struct {
	JobID   string           `json:"job_id"`
	Results []map[string]any `json:"results"`
}

// ErrorResponse is the body of every non-200 response.
type ErrorResponse struct {
	Error string `json:"error"`
}

// Route is one generated endpoint: its HTTP method, its echo path, the
// permission a token needs, and the handler.
type Route struct {
	Method     string
	Path       string
	Permission string
	Handler    echo.HandlerFunc
}

// JobClient is the part of the job client a behavior's routes use: a read
// or a write, on one host or broadcast.
type JobClient interface {
	Query(
		ctx context.Context,
		target string,
		category string,
		operation job.OperationType,
		data any,
	) (string, *job.Response, error)
	QueryBroadcast(
		ctx context.Context,
		target string,
		category string,
		operation job.OperationType,
		data any,
	) (string, map[string]*job.Response, error)
	Modify(
		ctx context.Context,
		target string,
		category string,
		operation job.OperationType,
		data any,
	) (string, *job.Response, error)
	ModifyBroadcast(
		ctx context.Context,
		target string,
		category string,
		operation job.OperationType,
		data any,
	) (string, map[string]*job.Response, error)
}

// Server is what every generated handler shares: the job client and what it
// takes to check a token.
type Server struct {
	logger      *slog.Logger
	jobClient   JobClient
	tokens      api.TokenValidator
	signingKey  string
	customRoles map[string][]string
}

// NewServer returns the server the generated handlers run on.
func NewServer(
	logger *slog.Logger,
	jobClient JobClient,
	signingKey string,
	customRoles map[string][]string,
) *Server {
	return &Server{
		logger:      logger.With(slog.String("subsystem", "controller.behavior")),
		jobClient:   jobClient,
		tokens:      authtoken.New(logger),
		signingKey:  signingKey,
		customRoles: customRoles,
	}
}

// routers are the generated route lists, one per behavior with an API, in
// the order their gen packages registered.
var routers []func(s *Server) []Route

// Register adds one behavior's generated routes. Each gen/api.go calls it
// from init.
func Register(
	routes func(s *Server) []Route,
) {
	routers = append(routers, routes)
}

// Handler returns the route registrations for every registered behavior.
func Handler(
	logger *slog.Logger,
	jobClient JobClient,
	signingKey string,
	customRoles map[string][]string,
) []func(e *echo.Echo) {
	s := NewServer(logger, jobClient, signingKey, customRoles)

	return []func(e *echo.Echo){
		func(e *echo.Echo) {
			for _, routes := range routers {
				s.Mount(e, routes(s))
			}
		},
	}
}

// Mount adds routes to e, each behind the check that the request's token
// grants its permission.
func (s *Server) Mount(
	e *echo.Echo,
	routes []Route,
) {
	for _, r := range routes {
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

// Path is the echo path for a method: its route under behavior.RoutePrefix,
// with {param} written as :param.
func Path(
	b *behavior.Behavior,
	m *behavior.Method,
) string {
	p := strings.ReplaceAll(b.Route(m), "{", ":")

	return strings.ReplaceAll(p, "}", "")
}
