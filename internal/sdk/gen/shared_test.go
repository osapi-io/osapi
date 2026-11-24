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

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dave/jennifer/jen"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/sdk/gen/sdkmodel"
)

// SharedTestSuite covers the types of pkg/sdk/types, which the client refers
// to and does not write, and the constants of an enum that it writes.
type SharedTestSuite struct {
	genSuite
}

func (s *SharedTestSuite) TearDownTest() {
	shared = map[string]bool{}
}

func (s *SharedTestSuite) TestDeclaredTypes() {
	tests := []struct {
		name    string
		setup   func() string
		want    map[string]*goType
		wantErr string
	}{
		{
			name:  "when the directory does not exist, it declares none",
			setup: func() string { return filepath.Join(s.root, "none") },
			want:  map[string]*goType{},
		},
		{
			name: "when the directory has Go files, a test and a directory",
			setup: func() string {
				s.write(typesDir+"/a.go", `package types

type Step struct {
	Behavior string `+"`json:\"behavior\"`"+`
	Changed  *bool  `+"`json:\"changed,omitempty\"`"+`
	Hidden   string `+"`json:\"-\"`"+`
	Plain, Other int
}

type (
	Mode string
	Kind = string
)

const (
	ModeRaw      Mode = "raw"
	ModeTemplate Mode = "template"
	KindOne      Kind = "one"
	Untyped           = "untyped"
	Number       Mode = 1
	Count        int  = 2
	Steps        Step = Step{}
)

var Other = "other"

func f() {}
`)
				s.write(typesDir+"/a_test.go", "package types\n\ntype Hidden struct{}\n")
				s.write(typesDir+"/notes.txt", "")
				s.Require().NoError(os.MkdirAll(filepath.Join(s.root, typesDir, "dir.go"), 0o755))

				return filepath.Join(s.root, typesDir)
			},
			want: map[string]*goType{
				"Step": {Struct: true, Names: []string{"Other", "Plain", "behavior", "changed"}},
				"Mode": {Names: []string{"raw", "template"}},
				"Kind": {Names: []string{"one"}},
			},
		},
		{
			name: "when a Go file does not parse",
			setup: func() string {
				s.write(typesDir+"/bad.go", "package types\n\ntype {\n")

				return filepath.Join(s.root, typesDir)
			},
			wantErr: "expected",
		},
		{
			name: "when the path is a file, not a directory",
			setup: func() string {
				s.write("file.go", "package types\n")

				return filepath.Join(s.root, "file.go")
			},
			wantErr: "not a directory",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()
			got, err := declaredTypes(tc.setup())

			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			s.Equal(tc.want, got)
		})
	}
}

func (s *SharedTestSuite) TestRunSharedType() {
	s.write(typesDir+"/types.go", "package types\n\ntype FileUploadResponse struct {\n"+
		"\tSHA256 string `json:\"sha256\"`\n\tSize int64 `json:\"size\"`\n}\n")

	s.Require().NoError(generate(s.root, false))

	var all strings.Builder
	entries, err := os.ReadDir(filepath.Join(s.root, dir))
	s.Require().NoError(err)
	for _, e := range entries {
		data, err := os.ReadFile(filepath.Join(s.root, dir, e.Name()))
		s.Require().NoError(err)
		all.Write(data)
	}
	s.NotContains(all.String(), "type FileUploadResponse struct",
		"a type of pkg/sdk/types is not written again")
	s.Contains(all.String(), "types.FileUploadResponse",
		"the client refers to the type of pkg/sdk/types")
}

func (s *SharedTestSuite) TestRunSharedTypeDiffers() {
	tests := []struct {
		name    string
		source  string
		wantErr string
	}{
		{
			name: "when a struct has other fields than the properties of the schema",
			source: "type FileUploadResponse struct {\n" +
				"\tSHA256 string `json:\"sha256\"`\n\tLength int64 `json:\"length\"`\n}\n",
			wantErr: "the schema FileUploadResponse binds to types.FileUploadResponse, " +
				"but the properties of the schema are [sha256 size], " +
				"and of types.FileUploadResponse are [length sha256]",
		},
		{
			name: "when a type has other values than the enum of the schema",
			source: "type DnsKind string\n\nconst (\n\tDnsKindStatic DnsKind = \"static\"\n" +
				"\tDnsKindAuto DnsKind = \"auto\"\n)\n",
			wantErr: "the schema DnsKind binds to types.DnsKind, " +
				"but the values of the schema are [dhcp static], and of types.DnsKind are [auto static]",
		},
		{
			name:   "when a schema is an object and the type is not a struct",
			source: "type FileUploadResponse string\n",
			wantErr: "the schema FileUploadResponse binds to types.FileUploadResponse, " +
				"but one is a struct and the other is not",
		},
		{
			name:    "when a schema is an enum and the type is a struct",
			source:  "type DnsKind struct{}\n",
			wantErr: "one is a struct and the other is not",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()
			s.write(typesDir+"/types.go", "package types\n\n"+tc.source)

			s.ErrorContains(generate(s.root, false), tc.wantErr)
			s.FileExists(filepath.Join(s.root, dir, "gone.gen.go"),
				"a run that fails writes nothing")
		})
	}
}

func (s *SharedTestSuite) TestPagesKnowTheSharedTypes() {
	s.write(typesDir+"/types.go", "package types\n\ntype ContentType string\n")
	s.write(
		pageDir+"/"+indexPage,
		"# Client\n\n<!-- osapi:sdk-services -->\n<!-- /osapi:sdk-services -->\n",
	)

	s.Require().NoError(generate(s.root, true))

	s.True(shared["ContentType"],
		"the pages write a value of a shared type as types.X, as the examples do")
}

func (s *SharedTestSuite) TestRunTypesDirError() {
	s.write(typesDir+"/bad.go", "package types\n\ntype {\n")

	s.Error(generate(s.root, false))
}

func (s *SharedTestSuite) TestTypeCode() {
	shared = map[string]bool{"Step": true}

	s.Equal(
		`types.Step`,
		fmt.Sprintf(
			"%#v",
			jen.Null().Add(typeCode(&sdkmodel.TypeRef{Kind: sdkmodel.Named, Name: "Step"})),
		),
	)
	s.Equal(
		`Mode`,
		fmt.Sprintf(
			"%#v",
			jen.Null().Add(typeCode(&sdkmodel.TypeRef{Kind: sdkmodel.Named, Name: "Mode"})),
		),
	)
}

func (s *SharedTestSuite) TestWriteEnum() {
	tests := []struct {
		name string
		typ  *sdkmodel.Type
		want []string
	}{
		{
			name: "when the type has values, each is a constant",
			typ: &sdkmodel.Type{
				Name:   "Mode",
				Scalar: "string",
				Enum:   []string{"raw", "partial_failure"},
			},
			want: []string{
				"type Mode string",
				`ModeRaw            Mode = "raw"`,
				`ModePartialFailure Mode = "partial_failure"`,
			},
		},
		{
			name: "when the type has no values, it has no constants",
			typ:  &sdkmodel.Type{Name: "Mode", Scalar: "string"},
			want: []string{"type Mode string"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			f := jen.NewFile("client")
			writeType(f, tc.typ)
			got := fmt.Sprintf("%#v", f)

			for _, w := range tc.want {
				s.Contains(got, w)
			}
			if len(tc.typ.Enum) == 0 {
				s.False(strings.Contains(got, "const"), "no constant")
			}
		})
	}
}

func (s *SharedTestSuite) TestNamedValue() {
	named := &sdkmodel.TypeRef{Kind: sdkmodel.Named, Name: "ContentType"}

	shared = map[string]bool{"ContentType": true}
	code, err := value(named, "raw")
	s.Require().NoError(err)
	s.Equal(`types.ContentType("raw")`, fmt.Sprintf("%#v", code))

	shared = map[string]bool{}
	code, err = value(named, "raw")
	s.Require().NoError(err)
	s.Equal(`client.ContentType("raw")`, fmt.Sprintf("%#v", code))

	_, err = value(named, 3)
	s.ErrorContains(err, "has no literal")

	s.Equal("", zeroJSON(named))
}

func (s *SharedTestSuite) TestPathCodeEnum() {
	mode := &sdkmodel.Param{
		Name:    "mode",
		ArgName: "mode",
		Type:    &sdkmodel.TypeRef{Kind: sdkmodel.Named, Name: "Mode"},
	}
	m := &sdkmodel.Method{Path: "/api/x/{mode}", PathParams: []*sdkmodel.Param{mode}}

	s.Equal(
		`"/api/x/" + url.PathEscape(string(mode))`,
		fmt.Sprintf("%#v", jen.Null().Add(pathCode(m))),
		"an enum of the path is its text, escaped",
	)
}

func TestSharedTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(SharedTestSuite))
}
