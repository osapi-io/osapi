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

package secret_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/secret"
)

type SecretPublicTestSuite struct {
	suite.Suite
}

func (s *SecretPublicTestSuite) TestValue() {
	tests := []struct {
		name      string
		value     secret.Value
		reference bool
		shows     string
	}{
		{
			name:  "when the value is a literal",
			value: "hunter2",
			shows: secret.Redacted,
		},
		{
			name:      "when the value is a reference",
			value:     "secret://default/registry-token",
			reference: true,
			shows:     "secret://default/registry-token",
		},
		{
			name:  "when the value is empty",
			value: "",
			shows: "",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(string(tc.value), tc.value.Reveal())
			s.Equal(tc.reference, tc.value.IsReference())
			s.Equal(tc.value == "", tc.value.IsZero())
			s.Equal(tc.shows, tc.value.String())
			for _, verb := range []string{"%v", "%s", "%+v", "%#v", "%10s"} {
				s.NotContains(fmt.Sprintf(verb, tc.value), "hunter2", verb)
			}
			s.Equal(fmt.Sprintf("%q", tc.shows), fmt.Sprintf("%q", tc.value))

			data, err := json.Marshal(tc.value)
			s.Require().NoError(err)
			want, _ := json.Marshal(tc.shows)
			s.Equal(string(want), string(data))

			var buf bytes.Buffer
			slog.New(slog.NewJSONHandler(&buf, nil)).Info("x", "password", tc.value)
			s.NotContains(buf.String(), "hunter2")
			s.Contains(buf.String(), `"password":`+string(want))
		})
	}
}

func (s *SecretPublicTestSuite) TestDecode() {
	var in struct {
		Password secret.Value            `json:"password"`
		Env      map[string]secret.Value `json:"env"`
	}
	s.Require().NoError(json.Unmarshal(
		[]byte(`{"password":"hunter2","env":{"TOKEN":"secret://default/token"}}`), &in))
	s.Equal("hunter2", in.Password.Reveal())
	s.True(in.Env["TOKEN"].IsReference())
	s.Error(json.Unmarshal([]byte(`{"password":1}`), &in), "a secret is a string")
}

type inner struct {
	Token secret.Value `json:"token,omitempty"`
	Name  string       `json:"name"`
}

type embedded struct {
	Key secret.Value `json:"key,omitempty"`
}

type input struct {
	embedded
	Password secret.Value            `json:"password,omitempty"`
	Env      map[string]secret.Value `json:"env,omitempty"`
	List     []secret.Value          `json:"list,omitempty"`
	Pair     [2]secret.Value         `json:"pair"`
	Ptr      *secret.Value           `json:"ptr,omitempty"`
	Inner    inner                   `json:"inner"`
	Inners   []*inner                `json:"inners,omitempty"`
	Skipped  secret.Value            `json:"-"`
	Plain    string                  `json:"plain,omitempty"`
	hidden   secret.Value
}

func (s *SecretPublicTestSuite) TestPlain() {
	ptr := secret.Value("p")
	tests := []struct {
		name string
		in   any
		want string
	}{
		{
			name: "when the input has a secret at each depth",
			in: input{
				embedded: embedded{Key: "k"},
				Password: "hunter2",
				Env:      map[string]secret.Value{"A": "a", "B": "secret://default/b"},
				List:     []secret.Value{"l"},
				Pair:     [2]secret.Value{"x", "y"},
				Ptr:      &ptr,
				Inner:    inner{Token: "t", Name: "n"},
				Inners:   []*inner{{Token: "u"}, nil},
				Skipped:  "s",
				Plain:    "plain",
				hidden:   "h",
			},
			want: `{"key":"k","password":"hunter2","env":{"A":"a","B":"secret://default/b"},` +
				`"list":["l"],"pair":["x","y"],"ptr":"p","inner":{"token":"t","name":"n"},` +
				`"inners":[{"token":"u","name":""},null],"plain":"plain"}`,
		},
		{
			name: "when the secrets are empty, omitempty leaves them out",
			in:   &input{},
			want: `{"pair":["",""],"inner":{"name":""}}`,
		},
		{
			name: "when the input has no secret",
			in:   struct{ Name string }{Name: "n"},
			want: `{"Name":"n"}`,
		},
		{
			name: "when the input is nil",
			in:   nil,
			want: `null`,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			data, err := json.Marshal(secret.Plain(tc.in))
			s.Require().NoError(err)
			s.JSONEq(tc.want, string(data))
		})
	}

	redacted, err := json.Marshal(
		input{Password: "hunter2", Env: map[string]secret.Value{"A": "a"}},
	)
	s.Require().NoError(err)
	s.NotContains(string(redacted), "hunter2", "without Plain, a literal is redacted")
}

func (s *SecretPublicTestSuite) TestReferences() {
	ref := secret.Value("secret://v/ptr")
	in := &input{
		embedded: embedded{Key: "secret://v/key"},
		Password: "secret://v/password",
		Env:      map[string]secret.Value{"B": "secret://v/b", "A": "literal"},
		Ptr:      &ref,
		Inner:    inner{Token: "t"},
		hidden:   "secret://v/hidden",
	}
	s.Equal(
		[]string{"secret://v/key", "secret://v/password", "secret://v/b", "secret://v/ptr"},
		secret.References(in),
	)
	s.Empty(secret.References(nil))
	s.Empty(secret.References(any(input{Password: "literal"})))
}

func (s *SecretPublicTestSuite) TestIsAndFields() {
	s.True(secret.Is(reflectTypeOf[secret.Value]()))
	s.True(secret.Is(reflectTypeOf[map[string]secret.Value]()))
	s.True(secret.Is(reflectTypeOf[*[]secret.Value]()))
	s.False(secret.Is(reflectTypeOf[string]()))
	s.False(secret.Is(reflectTypeOf[inner]()))

	type self struct {
		Next *self        `json:"next"`
		Pass secret.Value `json:"pass"`
	}
	s.Equal(
		[]string{"env", "key", "list", "pair", "password", "ptr", "token"},
		secret.Fields(reflectTypeOf[*input]()),
	)
	s.Equal([]string{"pass"}, secret.Fields(reflectTypeOf[self]()))
	s.Empty(secret.Fields(reflectTypeOf[string]()))
	s.Equal(
		[]string{"Untagged"},
		secret.Fields(reflectTypeOf[struct{ Untagged secret.Value }]()),
	)
}

func (s *SecretPublicTestSuite) TestDescribe() {
	s.Equal(secret.Takes, secret.Describe(""))
	s.Equal("The password. "+secret.Takes, secret.Describe("The password."))
	s.Equal("The password. "+secret.Takes, secret.Describe("The password"))
}

func (s *SecretPublicTestSuite) TestUnresolved() {
	s.EqualError(
		secret.Unresolved("secret://default/registry-token"),
		`secret://default/registry-token: the vault "default" does not exist`,
	)
}

func TestSecretPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(SecretPublicTestSuite))
}
