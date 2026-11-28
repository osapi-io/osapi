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
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	_ "github.com/osapi-io/osapi/internal/behaviors/all"
	"github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/cli/commands/config"
	osapiconfig "github.com/osapi-io/osapi/internal/config"

	// Each command group, which registers the owner of each section of the
	// config and its defaults (command.Config), as the binary does.
	_ "github.com/osapi-io/osapi/internal/cli/commands/all"
)

type MethodsPublicTestSuite struct {
	suite.Suite
}

// file writes body to a config file, and is the deps that read it.
func (s *MethodsPublicTestSuite) file(
	body string,
) config.Deps {
	path := filepath.Join(s.T().TempDir(), "osapi.yaml")
	s.Require().NoError(os.WriteFile(path, []byte(body), 0o600))

	return config.Deps{File: func() string { return path }}
}

// valid is a config that each rule accepts.
const valid = `controller:
  api:
    security:
      signing_keys:
        - id: test
          key: 0123456789abcdef0123456789abcdef
  client:
    security:
      bearer_token: token
`

func (s *MethodsPublicTestSuite) TestList() {
	tests := []struct {
		name         string
		deps         config.Deps
		validateFunc func(config.ListOutput, error)
	}{
		{
			name: "when the file sets a key, its value comes from the file",
			deps: s.file("controller:\n  api:\n    port: 9090\n"),
			validateFunc: func(got config.ListOutput, err error) {
				s.Require().NoError(err)
				s.NotEmpty(got.Items)
				for _, it := range got.Items {
					if it.Name == "controller.api.port" {
						s.EqualValues(9090, it.Value)
						s.Equal("file", it.Source)
					}
				}
			},
		},
		{
			name: "when the file is not there, the values are the defaults",
			deps: config.Deps{
				File: func() string { return filepath.Join(s.T().TempDir(), "none.yaml") },
			},
			validateFunc: func(got config.ListOutput, err error) {
				s.Require().NoError(err)
				s.Contains(got.File, "not found: defaults and environment only")
			},
		},
		{
			name: "when the file does not parse, it is the error",
			deps: s.file("controller: [\n"),
			validateFunc: func(_ config.ListOutput, err error) {
				s.Error(err)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.validateFunc(config.List(context.Background(), tc.deps, config.ListInput{}))
		})
	}
}

func (s *MethodsPublicTestSuite) TestGet() {
	tests := []struct {
		name         string
		deps         config.Deps
		key          string
		validateFunc func(config.GetOutput, error)
	}{
		{
			name: "when the key is a setting, it is the setting",
			deps: s.file(valid),
			key:  "controller.api.security.signing_keys",
			validateFunc: func(got config.GetOutput, err error) {
				s.Require().NoError(err)
				s.Equal("controller.api.security.signing_keys", got.Setting.Name)
				s.Equal("********", got.Setting.Value)
			},
		},
		{
			name: "when the key is not a setting, it says so",
			deps: s.file(valid),
			key:  "no.such.key",
			validateFunc: func(_ config.GetOutput, err error) {
				s.EqualError(err, `no setting "no.such.key": osapi config list shows them all`)
			},
		},
		{
			name: "when the file does not parse, it is the error",
			deps: s.file("controller: [\n"),
			key:  "controller.api.port",
			validateFunc: func(_ config.GetOutput, err error) {
				s.Error(err)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.validateFunc(config.Get(context.Background(), tc.deps, config.GetInput{Key: tc.key}))
		})
	}
}

// TestFlag names the command of a setting that has a flag: a behavior's
// setting, which osapi agent start takes as a flag.
func (s *MethodsPublicTestSuite) TestFlag() {
	got, err := config.List(context.Background(), s.file(valid), config.ListInput{})
	s.Require().NoError(err)
	flags := 0
	for _, it := range got.Items {
		if it.Flag != "" {
			flags++
			s.Contains(it.Flag, "osapi agent start --")
		}
	}
	s.Positive(flags)
}

func (s *MethodsPublicTestSuite) TestSpec() {
	got, err := config.Spec(context.Background(), config.Deps{}, config.SpecInput{})
	s.Require().NoError(err)
	s.Equal(osapiconfig.Spec().Sections(), got.Kinds)
}

func (s *MethodsPublicTestSuite) TestDefault() {
	got, err := config.Default(context.Background(), config.Deps{}, config.DefaultInput{})
	s.Require().NoError(err)
	s.Equal(osapiconfig.DefaultYAML(), got.Content)
}

func (s *MethodsPublicTestSuite) TestValidate() {
	tests := []struct {
		name         string
		deps         config.Deps
		validateFunc func(config.ValidateOutput, error)
	}{
		{
			name: "when the config is valid, it says so and the command succeeds",
			deps: s.file(valid),
			validateFunc: func(got config.ValidateOutput, err error) {
				s.Require().NoError(err)
				s.True(got.Valid)
				s.Empty(got.Problems)
				s.NotNil(got.Problems)
				s.Nil(got.Fixes)
			},
		},
		{
			name: "when the config breaks a rule, the output has each problem and its fix, and the command fails",
			deps: s.file("controller:\n  api:\n    port: 99999\n"),
			validateFunc: func(got config.ValidateOutput, err error) {
				s.True(errors.Is(err, command.FailedError{}))
				s.False(got.Valid)
				s.NotEmpty(got.Problems)
				s.NotEmpty(got.Fixes["value"])
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.validateFunc(config.Validate(context.Background(), tc.deps, config.ValidateInput{}))
		})
	}
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
