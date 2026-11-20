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

package natssetup_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	natsclient "github.com/osapi-io/nats-client/pkg/client"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli/commands/natssetup"
	"github.com/osapi-io/osapi/internal/config"
)

type MethodsPublicTestSuite struct {
	suite.Suite
}

func (s *MethodsPublicTestSuite) TestSetup() {
	tests := []struct {
		name         string
		prepareErr   error
		validateFunc func(natssetup.Output, *natsclient.Options, error)
	}{
		{
			name: "when the server is prepared, the output is the server",
			validateFunc: func(got natssetup.Output, opts *natsclient.Options, err error) {
				s.Require().NoError(err)
				s.Equal(
					natssetup.Output{
						Host:      "nats.example",
						Port:      4222,
						Namespace: "osapi",
						Auth:      "none",
					},
					got,
				)
				s.Equal("osapi-nats-setup", opts.Name)
			},
		},
		{
			name:       "when the server cannot be prepared, it is the error",
			prepareErr: errors.New("no route"),
			validateFunc: func(_ natssetup.Output, _ *natsclient.Options, err error) {
				s.EqualError(err, "set up JetStream on nats.example:4222: no route")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cfg := config.Config{}
			cfg.Controller.NATS.Host = "nats.example"
			cfg.Controller.NATS.Port = 4222
			cfg.Controller.NATS.Namespace = "osapi"
			cfg.Controller.NATS.Auth.Type = "none"
			var opts *natsclient.Options
			d := natssetup.Deps{
				Config: cfg,
				Logger: slog.New(slog.DiscardHandler),
				Prepare: func(
					_ context.Context, _ *slog.Logger, o *natsclient.Options, _ string, _ config.NATS,
				) error {
					opts = o

					return tc.prepareErr
				},
			}

			got, err := natssetup.Setup(context.Background(), d, natssetup.Input{})
			tc.validateFunc(got, opts, err)
		})
	}
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
