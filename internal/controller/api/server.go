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

package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/osapi-io/osapi/internal/controller/endpoint"

	echootel "github.com/labstack/echo-opentelemetry"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/telemetry/httplog"
	"github.com/osapi-io/osapi/internal/telemetry/respspan"
)

// shutdownGracePeriod bounds how long Stop waits for in-flight requests.
const shutdownGracePeriod = 10 * time.Second

// New initialize a new Server and configure an Echo server.
func New(
	appConfig config.Config,
	logger *slog.Logger,
	opts ...Option,
) *Server {
	e := echo.New()
	e.HTTPErrorHandler = errorHandler

	allowOrigins := appConfig.Controller.API.Security.CORS.AllowOrigins

	s := &Server{
		Echo:      e,
		logger:    logger.With(slog.String("subsystem", "controller.api")),
		appConfig: appConfig,
	}

	for _, opt := range opts {
		opt(s)
	}

	// Register middleware after options so MeterProvider is available.
	otelConfig := echootel.Config{
		ServerName:     "osapi-api",
		TracerProvider: otel.GetTracerProvider(),
		Skipper:        untraced,
	}
	if s.meterProvider != nil {
		otelConfig.MeterProvider = s.meterProvider
	}

	e.Use(echootel.NewMiddlewareWithConfig(otelConfig))
	e.Use(traceResponse)
	e.Use(respspan.Middleware)
	e.Use(httplog.New(logger))
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	// CORS is registered only when origins are configured.
	//
	// v4 defaulted an empty AllowOrigins to "*", which sends
	// Access-Control-Allow-Origin: * from an authenticated API without anyone
	// choosing it. v5 panics instead, and this package does not panic, so
	// leaving the middleware off is the remaining option. No header means the
	// browser applies same-origin policy, which is what an operator who has
	// not configured CORS should get.
	if len(allowOrigins) > 0 {
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: allowOrigins,
		}))
		s.logger.Info(
			"CORS enabled",
			slog.Any("allow_origins", allowOrigins),
		)
	} else {
		// Say so rather than leaving the operator to infer it from a browser
		// error with no server-side counterpart.
		s.logger.Info(
			"CORS not configured, cross-origin browser requests will be refused",
			slog.String(
				"configure",
				"controller.api.security.cors.allow_origins",
			),
		)
	}

	// Register audit middleware if an audit store is configured.
	if s.auditStore != nil {
		e.Use(auditMiddleware(s.auditStore, logger))
	}

	return s
}

// Start starts the Echo server on the configured host and port. An empty
// host binds every interface.
func (s *Server) Start() {
	// v5 drives the listener from a context rather than a Shutdown method,
	// so Stop cancels this one.
	ctx, cancel := context.WithCancel(context.Background())
	s.stop = cancel

	go func() {
		s.logger.Info("starting server")

		sc := echo.StartConfig{
			Address: net.JoinHostPort(
				s.appConfig.Controller.API.Host,
				strconv.Itoa(s.appConfig.Controller.API.Port),
			),
			HideBanner:      true,
			GracefulTimeout: shutdownGracePeriod,
		}
		if err := sc.Start(ctx, s.Echo); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			s.logger.Error(
				"failed to start server",
				slog.String("error", err.Error()),
			)
		}
	}()
}

// Stop gracefully shuts down the Echo server.
// Stop asks the server to drain and returns without waiting for it.
//
// Cancelling is what starts the drain; StartConfig.GracefulTimeout bounds how long
// Echo waits, and the goroutine in Start reports the outcome. A caller must not
// treat this returning as "every in-flight request finished" — the start
// command then runs the cleanups within one shutdown budget. That budget bounds
// the process, not this call.
//
// The context is unused for the same reason: there is nothing here to cancel.
func (s *Server) Stop(
	_ context.Context,
) {
	s.logger.Info("stopping server")

	if s.stop == nil {
		s.logger.Info("server stopped gracefully")

		return
	}

	// Cancelling is what asks v5 to drain. StartConfig.GracefulTimeout bounds
	// how long it waits, and the goroutine in Start reports any error.
	s.stop()
	s.stop = nil
	s.logger.Info("server stopped gracefully")
}

// untraced tells which requests get no span and no HTTP metrics: the health
// checks and the files of the UI. A probe and the UI poll the health
// endpoints each few seconds, and their traces hide the traces of the jobs.
func untraced(
	c *echo.Context,
) bool {
	path := c.Request().URL.Path

	return !strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/api/health")
}

// traceResponse puts the trace of the request in the traceresponse header of
// the response, from W3C Trace Context level 2. The CLI prints the trace ID
// from it, so a person can find the trace of a command without access to
// the tracing backend from the machine of the CLI. A request with no span,
// such as a health check, gets no header.
func traceResponse(
	next echo.HandlerFunc,
) echo.HandlerFunc {
	return func(c *echo.Context) error {
		sc := trace.SpanContextFromContext(c.Request().Context())
		if sc.IsValid() {
			c.Response().Header().Set(
				"traceresponse",
				fmt.Sprintf("00-%s-%s-%s", sc.TraceID(), sc.SpanID(), sc.TraceFlags()),
			)
		}

		return next(c)
	}
}

// errorHandler answers each error that a route returns, such as a path or
// query parameter that does not bind, as {"error": "..."}: the body of
// each error of the API. It keeps the error's status, and a 500 tells only
// its status text, so no internal detail leaves the controller.
func errorHandler(
	c *echo.Context,
	err error,
) {
	if r, _ := echo.UnwrapResponse(c.Response()); r != nil && r.Committed {
		return
	}

	code := http.StatusInternalServerError
	var sc echo.HTTPStatusCoder
	if errors.As(err, &sc) && sc.StatusCode() != 0 {
		code = sc.StatusCode()
	}
	msg := http.StatusText(code)
	var he *echo.HTTPError
	if errors.As(err, &he) && he.Message != "" && code != http.StatusInternalServerError {
		msg = he.Message
	}

	if c.Request().Method == http.MethodHead {
		_ = c.NoContent(code)

		return
	}
	_ = c.JSON(code, map[string]string{"error": msg, "reason": endpoint.ReasonFor(code)})
}
