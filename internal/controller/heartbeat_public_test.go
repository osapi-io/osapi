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

package controller_test

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/bus"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller"
)

type HeartbeatPublicTestSuite struct {
	suite.Suite
}

func TestHeartbeatPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(HeartbeatPublicTestSuite))
}

func (s *HeartbeatPublicTestSuite) TestStartHeartbeat() {
	s.Run("when the hostname fails registers the controller as unknown", func() {
		controller.SetHostnameFn(func() (string, error) { return "", errors.New("no hostname") })
		defer controller.ResetHostnameFn()
		cfg := startServer(s.T())

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		c, err := controller.New(ctx, slog.Default(), cfg)
		s.Require().NoError(err)
		defer c.Stop(context.Background())

		nc, err := bus.Dial(slog.Default(), natsOptions(cfg))
		s.Require().NoError(err)
		defer nc.Close()
		bucket := bus.RegistryKV("test", cfg.NATS.Registry).Bucket
		s.Eventually(func() bool {
			keys, _ := nc.KVKeys(bucket)
			return slices.Contains(keys, "controller.unknown")
		}, 5*time.Second, 50*time.Millisecond)
	})
}

func (s *HeartbeatPublicTestSuite) TestSubComponents() {
	s.Run("when each part is on reports each one as ok", func() {
		var cfg config.Config
		cfg.Controller.API.Port = 8080
		cfg.Controller.Metrics = config.MetricsServer{Enabled: true, Host: "127.0.0.1", Port: 9090}
		cfg.Controller.PKI.Enabled = true
		cfg.Controller.Notifications.Enabled = true
		cfg.Telemetry.Tracing.Enabled = true

		got := controller.ExportSubComponents(cfg)

		s.Equal("http://0.0.0.0:8080", got["controller.api"].Address)
		s.Equal("http://127.0.0.1:9090", got["controller.metrics"].Address)
		for _, name := range []string{"controller.metrics", "controller.pki", "controller.notifier", "controller.tracing", "controller.ui"} {
			s.Equal("ok", got[name].Status, name)
		}
	})

	s.Run("when the parts are off reports them as disabled", func() {
		var cfg config.Config
		off := false
		cfg.Controller.UI.Enabled = &off

		got := controller.ExportSubComponents(cfg)

		for _, name := range []string{"controller.metrics", "controller.pki", "controller.notifier", "controller.tracing", "controller.ui"} {
			s.Equal("disabled", got[name].Status, name)
		}
	})
}
