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
package cmd

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/audit"
	"github.com/osapi-io/osapi/internal/config"
)

// LocalAuditTestSuite holds that a local command on an agent host, whose
// config file sets no controller, does not try the controller's NATS
// settings for its audit entry. It warned "the change is not in the audit
// log" on each such host (osapi-io/osapi#1552).
type LocalAuditTestSuite struct {
	suite.Suite

	logs      bytes.Buffer
	oldLogger *slog.Logger
	oldConfig config.Config
	oldInFile func() bool
}

func (s *LocalAuditTestSuite) SetupTest() {
	s.logs.Reset()
	s.oldLogger, s.oldConfig, s.oldInFile = logger, appConfig, controllerInFile
	logger = slog.New(slog.NewTextHandler(&s.logs, nil))
	// The controller's NATS settings name a CA file that does not exist, as
	// on an agent host.
	appConfig = config.Config{}
	appConfig.NATS.Audit.Stream = "AUDIT"
	appConfig.NATS.Audit.Subject = "audit"
	appConfig.Controller.NATS.Host = "127.0.0.1"
	appConfig.Controller.NATS.Port = 1
	appConfig.Controller.NATS.TLS.Enabled = true
	appConfig.Controller.NATS.TLS.CAFile = "/no/such/nats.crt"
}

func (s *LocalAuditTestSuite) TearDownTest() {
	logger, appConfig, controllerInFile = s.oldLogger, s.oldConfig, s.oldInFile
}

func (s *LocalAuditTestSuite) TestLocalAudit() {
	tests := []struct {
		name         string
		inFile       bool
		validateFunc func(error, string)
	}{
		{
			name:   "on an agent host it writes nothing, and says why once",
			inFile: false,
			validateFunc: func(err error, logs string) {
				s.NoError(err)
				s.Contains(logs, "an agent host keeps no local audit log")
				s.NotContains(logs, "the change is not in the audit log")
			},
		},
		{
			name:   "with the controller in the file it tries NATS, and warns when it cannot",
			inFile: true,
			validateFunc: func(err error, logs string) {
				s.Error(err)
				s.Contains(logs, "the change is not in the audit log")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.logs.Reset()
			controllerInFile = func() bool { return tc.inFile }
			err := localAudit(context.Background(), audit.Entry{Action: "key.rotate"})
			tc.validateFunc(err, s.logs.String())
		})
	}
}

func TestLocalAuditTestSuite(t *testing.T) {
	suite.Run(t, new(LocalAuditTestSuite))
}
