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

package command_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/behaviortest"
	"github.com/osapi-io/osapi/internal/behaviors/command"
)

type BehaviorPublicTestSuite struct {
	suite.Suite
}

func (s *BehaviorPublicTestSuite) TestBehavior() {
	s.Same(command.Behavior, behavior.Lookup("command"))
	s.Empty(behavior.ValidateBehavior(command.Behavior))
}

// TestPermissions: exec is command:execute and admin only, and shell is
// command:shell, which the shell verb makes admin only.
func (s *BehaviorPublicTestSuite) TestPermissions() {
	exec := command.Behavior.Method("exec")
	shell := command.Behavior.Method("shell")
	s.Equal("command:execute", string(command.Behavior.Permission(exec.PermissionVerb())))
	s.True(exec.Admin)
	s.Equal("command:shell", string(command.Behavior.Permission(shell.PermissionVerb())))
	s.Equal("/api/node/{hostname}/command/exec", command.Behavior.Path(exec))
	s.Equal("/api/node/{hostname}/command/shell", command.Behavior.Path(shell))
}

// TestExamples runs every example and refusal on a fake host with each
// platform's provider, through the methods.
func (s *BehaviorPublicTestSuite) TestExamples() {
	behaviortest.Run(s.T(), command.Behavior)
}

func TestBehaviorPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(BehaviorPublicTestSuite))
}
