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
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli/commands/agentcmd"
	"github.com/osapi-io/osapi/internal/cli/commands/controllercmd"
	"github.com/osapi-io/osapi/internal/cli/commands/natsserver"
)

// KeyDirsTestSuite holds that the agent keeps its keys in a directory of its
// own. A default of /etc/osapi/pki, the controller's and the NATS server's,
// made an agent that had keys there from an older install sign with them,
// so the controller refused each of its results (osapi-io/osapi#1393).
type KeyDirsTestSuite struct {
	suite.Suite
}

func (s *KeyDirsTestSuite) TestAgentKeyDirIsItsOwn() {
	s.Equal("/etc/osapi/pki/agent", agentcmd.Defaults.PKI.KeyDir,
		"the docs and existing agents use this directory")
	s.NotEqual(controllercmd.Defaults.PKI.KeyDir, agentcmd.Defaults.PKI.KeyDir)
	s.NotEqual(natsserver.Defaults.Server.KeyDir, agentcmd.Defaults.PKI.KeyDir)
}

func TestKeyDirsTestSuite(t *testing.T) {
	suite.Run(t, new(KeyDirsTestSuite))
}
