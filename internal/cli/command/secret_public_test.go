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
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/secret"
)

type SecretPublicTestSuite struct {
	suite.Suite

	reported error
}

func (s *SecretPublicTestSuite) SetupTest() {
	s.reported = nil
	s.Require().NoError(command.Attach(&cobra.Command{Use: "client"}, command.Deps{
		JSON:  func() bool { return false },
		Error: func(err error) { s.reported = err },
	}))
}

// failing is standard input that cannot be read.
type failing struct{}

func (failing) Read([]byte) (int, error) { return 0, errors.New("closed") }

func (s *SecretPublicTestSuite) TestSecret() {
	tests := []struct {
		name    string
		flag    string
		stdin   string
		failing bool
		want    secret.Value
		wantErr string
	}{
		{
			name: "when the flag is empty, the secret is empty",
		},
		{
			name: "when the flag is a reference, it is sent as it is",
			flag: "secret://default/pw",
			want: "secret://default/pw",
		},
		{
			name:  "when the flag is -, the value is standard input without its line break",
			flag:  "-",
			stdin: "hunter2\r\n",
			want:  "hunter2",
		},
		{
			name:    "when the flag is a value, it is refused: the shell history keeps it",
			flag:    "hunter2",
			wantErr: "--password takes a reference, secret://<vault>/<key>, or - to read the value from standard input; a value in a flag stays in the shell history",
		},
		{
			name:    "when standard input is empty, it is refused",
			flag:    "-",
			stdin:   "\n",
			wantErr: "--password -: standard input is empty",
		},
		{
			name:    "when standard input cannot be read, it is refused",
			flag:    "-",
			failing: true,
			wantErr: "--password: read standard input: closed",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()
			cmd := &cobra.Command{}
			cmd.SetIn(strings.NewReader(tc.stdin))
			if tc.failing {
				cmd.SetIn(failing{})
			}

			got, ok := command.Secret(cmd, "password", tc.flag)

			if tc.wantErr != "" {
				s.False(ok)
				s.EqualError(s.reported, tc.wantErr)

				return
			}
			s.True(ok)
			s.NoError(s.reported)
			s.Equal(tc.want, got)
		})
	}
}

func (s *SecretPublicTestSuite) TestSecrets() {
	tests := []struct {
		name    string
		flags   map[string]string
		want    map[string]secret.Value
		wantErr string
	}{
		{
			name: "when the flag is not given, there is no map",
		},
		{
			name:  "when each value is a reference, each is sent as it is",
			flags: map[string]string{"TOKEN": "secret://default/token"},
			want:  map[string]secret.Value{"TOKEN": "secret://default/token"},
		},
		{
			name:    "when a value is not a reference, it is refused",
			flags:   map[string]string{"A": "secret://default/a", "B": "hunter2"},
			wantErr: "--env B takes a reference, secret://<vault>/<key>; a value in a flag stays in the shell history",
		},
		{
			name:    "when a value is -, it is refused: standard input gives one value",
			flags:   map[string]string{"A": "-"},
			wantErr: "--env A takes a reference, secret://<vault>/<key>; a value in a flag stays in the shell history",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()

			got, ok := command.Secrets(&cobra.Command{}, "env", tc.flags)

			if tc.wantErr != "" {
				s.False(ok)
				s.EqualError(s.reported, tc.wantErr)

				return
			}
			s.True(ok)
			s.Equal(tc.want, got)
		})
	}
}

func TestSecretPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(SecretPublicTestSuite))
}
