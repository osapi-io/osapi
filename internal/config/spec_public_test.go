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

package config_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/config"
)

type SpecPublicTestSuite struct {
	suite.Suite
}

// TestSpecOf groups the keys by their top word, and says of each key what it
// does, its type, its default, how to set it and its rules.
func (s *SpecPublicTestSuite) TestSpecOf() {
	spec := config.SpecOf([]config.Key{
		{
			Name:     "controller.api.port",
			Env:      "OSAPI_CONTROLLER_API_PORT",
			Type:     "int",
			Default:  8080,
			Validate: "min=1",
			Doc:      "The port.",
		},
		{
			Name:   "controller.api.token",
			Env:    "OSAPI_CONTROLLER_API_TOKEN",
			Type:   "string",
			Secret: true,
		},
		{Name: "agent.labels", Env: "OSAPI_AGENT_LABELS", Type: "map", Default: ""},
	})

	s.Equal([]string{"controller", "agent"}, []string{spec.Kinds[0].Kind, spec.Kinds[1].Kind})
	port := spec.Of("controller")[0]
	s.Equal("The port. Type: int. Default: 8080.", port.Does)
	s.Equal(
		"In osapi.yaml, or with OSAPI_CONTROLLER_API_PORT, which wins over the file. Rules: min=1.",
		port.When,
	)
	token := spec.Of("controller")[1]
	s.Equal("No description. Type: string.", token.Does)
	s.Contains(token.When, "A secret")
	s.Equal("No description. Type: map.", spec.Of("agent")[0].Does, "an empty default is not shown")
}

// TestSpec has a term for each key.
func (s *SpecPublicTestSuite) TestSpec() {
	s.Len(config.Spec().Terms, len(config.Keys()))
}

func TestSpecPublicTestSuite(t *testing.T) {
	suite.Run(t, new(SpecPublicTestSuite))
}
