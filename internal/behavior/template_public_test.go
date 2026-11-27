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

package behavior_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
)

// deployIn holds a template and its vars.
type deployIn struct {
	Content string            `json:"content" template:"vars"`
	Vars    map[string]string `json:"vars"`
}

// noVarsIn names a vars field that it does not have.
type noVarsIn struct {
	Content string `json:"content" template:"vars"`
}

// notTextIn tags a field that is not a string.
type notTextIn struct {
	Content int               `json:"content" template:"vars"`
	Vars    map[string]string `json:"vars"`
}

// wrongVarsIn has vars of the wrong type.
type wrongVarsIn struct {
	Content string         `json:"content" template:"vars"`
	Vars    map[string]int `json:"vars"`
}

func deploy(context.Context, fakeProvider, deployIn) (result, error) {
	return result{}, nil
}

func deployNoVars(context.Context, fakeProvider, noVarsIn) (result, error) {
	return result{}, nil
}

func deployNotText(context.Context, fakeProvider, notTextIn) (result, error) {
	return result{}, nil
}

func deployWrongVars(context.Context, fakeProvider, wrongVarsIn) (result, error) {
	return result{}, nil
}

type TemplatePublicTestSuite struct {
	suite.Suite
}

func (s *TemplatePublicTestSuite) TestRender() {
	data := behavior.TemplateData{
		Facts:    map[string]any{"architecture": "arm64"},
		Vars:     map[string]string{"region": "west"},
		Hostname: "web-01",
	}

	tests := []struct {
		name    string
		text    string
		want    string
		wantErr string
	}{
		{
			name: "when the template reads facts, vars and the hostname",
			text: "{{.Hostname}} {{.Facts.architecture}} {{.Vars.region}}",
			want: "web-01 arm64 west",
		},
		{name: "when the template does not parse", text: "{{.Vars", wantErr: "parse template"},
		{name: "when a var is missing", text: "{{.Vars.zone}}", wantErr: "render template"},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			got, err := behavior.Render(tc.text, data)
			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			s.Equal(tc.want, got)
		})
	}
}

func (s *TemplatePublicTestSuite) TestTemplateFields() {
	tests := []struct {
		name string
		t    reflect.Type
		want []string
	}{
		{
			name: "when a field is tagged",
			t:    reflect.TypeFor[deployIn](),
			want: []string{"content=vars"},
		},
		{
			name: "when given a pointer",
			t:    reflect.TypeFor[*deployIn](),
			want: []string{"content=vars"},
		},
		{name: "when no field is tagged", t: reflect.TypeFor[getIn](), want: nil},
		{name: "when not a struct", t: reflect.TypeFor[string](), want: nil},
		{name: "when there is no type", t: nil, want: nil},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var got []string
			for _, f := range behavior.TemplateFields(tc.t) {
				got = append(got, f.Field+"="+f.Vars)
			}
			s.Equal(tc.want, got)
		})
	}
}

// TestCheckRefusesTemplates holds a method's input check, which the agent and
// behavior.Call run, to the template rule the API holds a request to.
func (s *TemplatePublicTestSuite) TestCheckRefusesTemplates() {
	m := behavior.Write("deploy", deploy)

	s.ErrorContains(
		m.Check(context.Background(), []byte(`{"content":"{{.Vars.missing}}"}`)),
		"missing",
	)
	s.NoError(
		m.Check(context.Background(), []byte(`{"content":"{{.Vars.port}}","vars":{"port":"80"}}`)),
	)
}

func (s *TemplatePublicTestSuite) TestCheckTemplates() {
	vars := map[string]string{"region": "west"}

	tests := []struct {
		name    string
		in      any
		wantMsg string
	}{
		{
			name: "when the template reads its data it passes",
			in: deployIn{
				Content: `{{.Hostname}} {{.Facts.fqdn}} {{.Vars.region}} {{printf "%s" .Vars.region}}`,
				Vars:    vars,
			},
		},
		{
			name: "when an if reads a var it is checked",
			in: deployIn{
				Content: "{{if .Vars.zone}}x{{else}}{{.Vars.region}}{{end}}",
				Vars:    vars,
			},
			wantMsg: "content reads .Vars.zone, which vars does not have",
		},
		{
			name: "when range and with change the dot their bodies are not read",
			in: deployIn{
				Content: "{{range .Facts.interfaces}}{{.Name}}{{end}}" +
					"{{with .Vars.region}}{{.}}{{else}}{{.Hostname}}{{end}}",
				Vars: vars,
			},
		},
		{
			name:    "when index reads a var that is not given",
			in:      deployIn{Content: `{{index .Vars "zone"}}`, Vars: vars},
			wantMsg: "content reads .Vars.zone, which vars does not have",
		},
		{
			name: "when index reads a var that is given, or reads by a variable",
			in: deployIn{
				Content: `{{index .Vars "region"}}{{index .Facts "fqdn"}}{{$k := "a"}}{{index .Vars $k}}`,
				Vars:    vars,
			},
		},
		{
			name:    "when $ reads a var that is not given",
			in:      deployIn{Content: "{{$.Vars.zone}}", Vars: vars},
			wantMsg: "content reads .Vars.zone, which vars does not have",
		},
		{
			name: "when $ reads a var inside range, it is checked",
			in: deployIn{
				Content: "{{range .Facts.interfaces}}{{$.Vars.zone}}{{end}}",
				Vars:    vars,
			},
			wantMsg: "content reads .Vars.zone, which vars does not have",
		},
		{
			name:    "when index reads a var through $ that is not given",
			in:      deployIn{Content: `{{index $.Vars "zone"}}`, Vars: vars},
			wantMsg: "content reads .Vars.zone, which vars does not have",
		},
		{
			name:    "when $ reads another field",
			in:      deployIn{Content: "{{$.Secret}}"},
			wantMsg: "content reads .Secret; a template reads only .Facts, .Vars and .Hostname",
		},
		{
			name: "when $ reads the data it has, or a declared variable is read",
			in: deployIn{
				Content: `{{$.Vars.region}}{{$.Hostname}}{{$x := .Vars.region}}{{$x}}{{range $i, $v := .Facts.interfaces}}{{$v.Name}}{{end}}`,
				Vars:    vars,
			},
		},
		{name: "when the template is empty it passes", in: deployIn{}},
		{
			name: "when the template only defines others it passes",
			in:   deployIn{Content: `{{define "x"}}y{{end}}`},
		},
		{
			name:    "when the template does not parse",
			in:      deployIn{Content: "{{.Vars"},
			wantMsg: "content is not a template",
		},
		{
			name:    "when the template reads another field",
			in:      deployIn{Content: "{{.Secret}}"},
			wantMsg: "content reads .Secret; a template reads only .Facts, .Vars and .Hostname",
		},
		{
			name:    "when the template reads a var that is not given",
			in:      &deployIn{Content: "{{.Vars.zone}}", Vars: vars},
			wantMsg: "content reads .Vars.zone, which vars does not have",
		},
		{
			name:    "when the input has no vars field every var is missing",
			in:      noVarsIn{Content: "{{.Vars.zone}}"},
			wantMsg: "content reads .Vars.zone, which vars does not have",
		},
		{name: "when the input is a nil pointer", in: (*deployIn)(nil)},
		{name: "when the input is not a struct", in: "x"},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			msg, ok := behavior.CheckTemplates(tc.in)
			if tc.wantMsg == "" {
				s.True(ok, msg)

				return
			}
			s.False(ok)
			s.Contains(msg, tc.wantMsg)
		})
	}
}

func (s *TemplatePublicTestSuite) TestTemplateRule() {
	tests := []struct {
		name    string
		m       *behavior.Method
		wantMsg string
		// surface is set when an input is a type that no flag can hold.
		surface bool
	}{
		{name: "when the template and its vars are right", m: behavior.Write("create", deploy)},
		{
			name:    "when the vars field is missing",
			m:       behavior.Write("create", deployNoVars),
			wantMsg: `names "vars" for its vars, which the input does not have`,
		},
		{
			name:    "when the template field is not a string",
			m:       behavior.Write("create", deployNotText),
			wantMsg: "is tagged template but is int, not a string",
		},
		{
			name:    "when the vars are not map[string]string",
			m:       behavior.Write("create", deployWrongVars),
			wantMsg: "is the vars of content but is map[string]int, not map[string]string",
			surface: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			b := behavior.Define(
				"files",
				"Files",
				behavior.Node,
				onDebian,
				behavior.API,
				behavior.Area("config"),
				behavior.CLI,
				tc.m.Describe("Create").Example("One", nil, nil),
			)
			var msgs []string
			for _, p := range behavior.ValidateBehavior(b) {
				if p.Rule == behavior.RuleTemplate {
					msgs = append(msgs, p.Message)
				}
				if !tc.surface {
					s.NotEqual(behavior.RuleSurface, p.Rule, "a map[string]string is a CLI flag")
				}
			}
			if tc.wantMsg == "" {
				s.Empty(msgs)

				return
			}
			s.Require().Len(msgs, 1)
			s.Contains(msgs[0], tc.wantMsg)
		})
	}
}

func TestTemplatePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(TemplatePublicTestSuite))
}
