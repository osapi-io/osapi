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

package all_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli/command"
	_ "github.com/osapi-io/osapi/internal/cli/commands/all"
	_ "github.com/osapi-io/osapi/internal/cli/commands/all/gen"
)

type AllPublicTestSuite struct {
	suite.Suite
}

// TestValidate holds each registered local group to the rules of a
// definition.
func (s *AllPublicTestSuite) TestValidate() {
	s.NotEmpty(command.All())
	for _, g := range command.All() {
		s.Run(strings.Join(g.CommandPath(), " "), func() {
			s.Empty(command.Validate(g))
		})
	}
}

// TestRefusals runs each input that a method's Fails declares, as the
// generated command does, and expects a usage error: the command exits 2,
// and the handler does not run. Validate makes each rule of an input have
// one, so this covers each rule of each local command.
func (s *AllPublicTestSuite) TestRefusals() {
	count := 0
	for _, g := range command.All() {
		for _, m := range g.Methods {
			for _, r := range m.Refusals() {
				count++
				s.Run(g.Name+" "+m.MethodName()+": "+r.Doc, func() {
					err := command.Run(&cobra.Command{}, g, m.MethodName(), r.Input)

					var ue *command.UsageError
					s.Require().True(errors.As(err, &ue), "a usage error, not %v", err)
					s.NotEmpty(ue.Msg)
					s.True(strings.HasPrefix(ue.Msg, "--"), "it names the flag: %s", ue.Msg)
				})
			}
		}
	}
	s.NotZero(count)
}

func TestAllPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(AllPublicTestSuite))
}
