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

package sdkmodel_test

import (
	"os"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/sdk/gen/sdkmodel"
)

type ModelPublicTestSuite struct {
	suite.Suite
}

// load is the spec of testdata/spec.yaml, with edit applied to its text.
func (s *ModelPublicTestSuite) load(
	edit func(string) string,
) *openapi3.T {
	data, err := os.ReadFile("testdata/spec.yaml")
	s.Require().NoError(err)
	doc, err := openapi3.NewLoader().LoadFromData([]byte(edit(string(data))))
	s.Require().NoError(err)

	return doc
}

func same(
	text string,
) string {
	return text
}

func (s *ModelPublicTestSuite) TestNew() {
	m, err := sdkmodel.New(s.load(same))
	s.Require().NoError(err)

	services := map[string]*sdkmodel.Service{}
	names := make([]string, 0, len(m.Services))
	for _, svc := range m.Services {
		services[svc.Name] = svc
		names = append(names, svc.Name)
	}

	tests := []struct {
		name         string
		validateFunc func()
	}{
		{
			name: "when the services are named from the paths of their groups",
			validateFunc: func() {
				s.Equal([]string{"File", "Job", "NetworkDNS", "NodeFile"}, names)
			},
		},
		{
			name: "when a method has the name of x-osapi-method, with its parameters",
			validateFunc: func() {
				get := services["NetworkDNS"].Methods[0]
				s.Equal("Get", get.Name)
				s.Equal("GetNodeNetworkDnsByKey", get.OperationID)
				s.Equal("dns:read", get.Permission)
				s.Equal([]int{200, 404}, get.Statuses)
				s.Equal("Collection[DnsEntry]", get.Result.String())
				s.Require().Len(get.PathParams, 2)
				s.Equal("hostname", get.PathParams[0].ArgName)
				s.Equal("interfaceName", get.PathParams[1].ArgName)
				s.Equal("InterfaceName", get.PathParams[1].GoName)
			},
		},
		{
			name: "when a write has a JSON body",
			validateFunc: func() {
				update := services["NetworkDNS"].Methods[1]
				s.Equal("Update", update.Name)
				s.Empty(update.Permission)
				s.Equal("hostname", update.Target().Name)
				s.Require().NotNil(update.Input)
				s.Equal("NetworkDNSUpdateInput", update.Input.Name)
				s.Require().Len(update.Input.Fields, 2)
				s.True(update.Input.Fields[0].Path)
				s.Equal("InterfaceName", update.Input.Fields[0].GoName)
				s.Equal("OverrideDHCP", update.Input.Fields[1].GoName)
				s.False(update.Input.Fields[1].Path)
			},
		},
		{
			name: "when a body is a multipart form, the method is marked to write by hand",
			validateFunc: func() {
				upload := services["File"].Methods[0]
				s.True(upload.Multipart)
				s.Equal("FileUploadResponse", upload.Result.String())
			},
		},
		{
			name: "when a query has required and optional parameters",
			validateFunc: func() {
				status := services["NodeFile"].Methods[0]
				s.Len(status.RequiredQuery(), 1)
				s.Empty(status.OptionalQuery())
				list := services["Job"].Methods[0]
				s.Empty(list.RequiredQuery())
				s.Len(list.OptionalQuery(), 2)
				s.Equal("int", list.Query[0].Type.String())
				s.Equal("bool", list.Query[1].Type.String())
			},
		},
		{
			name: "when an answer has no body, the method has no result",
			validateFunc: func() {
				del := services["Job"].Methods[1]
				s.Equal("Delete", del.Name)
				s.Nil(del.Result)
				s.Nil(del.Target())
				s.Len(del.Keys(), 1)
				s.Equal("string", del.PathParams[0].Type.String())
			},
		},
		{
			name: "when the types are each schema a method uses, and no collection",
			validateFunc: func() {
				types := make([]string, 0, len(m.Types))
				for _, t := range m.Types {
					types = append(types, t.Name)
				}
				s.Equal([]string{
					"DnsEntry", "ErrorResponse",
					"FileUploadResponse", "JobList", "JobListTimeline",
				}, types)
			},
		},
		{
			name: "when a field has its Go type, its enum and whether it is required",
			validateFunc: func() {
				fields := map[string]*sdkmodel.Field{}
				for _, t := range m.Types {
					for _, f := range t.Fields {
						fields[t.Name+"."+f.Name] = f
					}
				}
				s.True(fields["DnsEntry.hostname"].Required)
				s.False(fields["DnsEntry.servers"].Required)
				s.Equal([]string{"ok", "failed"}, fields["DnsEntry.status"].Enum)
				s.Equal("[]string", fields["DnsEntry.servers"].Type.String())
				s.Equal("int64", fields["FileUploadResponse.size"].Type.String())
				s.Equal("time.Time", fields["JobList.created"].Type.String())
				s.Equal("float64", fields["JobList.ratio"].Type.String())
				s.Equal("map[string]string", fields["JobList.labels"].Type.String())
				s.Equal("map[string]any", fields["JobList.facts"].Type.String())
				s.Equal("any", fields["JobList.result"].Type.String())
				s.Equal("[]JobListTimeline", fields["JobList.timeline"].Type.String())
				s.Equal("Min1", fields["JobList.1min"].GoName)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.validateFunc()
		})
	}
}

func (s *ModelPublicTestSuite) TestNewErrors() {
	tests := []struct {
		name string
		edit func(string) string
		want string
	}{
		{
			name: "when an operation has no x-osapi-method",
			edit: func(text string) string {
				return strings.Replace(text, "      x-osapi-method: list\n", "", 1)
			},
			want: "GetJobs: the operation has no x-osapi-method",
		},
		{
			name: "when an operation has no tag of the spec",
			edit: func(text string) string {
				return strings.Replace(text, "tags: [Job_API]", "tags: [Other]", 1)
			},
			want: "GetJobs: the operation has no tag of the spec",
		},
		{
			name: "when a tag has no x-displayName",
			edit: func(text string) string {
				return strings.Replace(text, "    x-displayName: Job\n", "", 1)
			},
			want: "tag Job_API has no x-displayName",
		},
		{
			name: "when a path has no segment of its group",
			edit: func(text string) string {
				return strings.Replace(text, "x-displayName: Job\n", "x-displayName: Work\n", 1)
			},
			want: "has no segment work",
		},
		{
			name: "when two methods of a service have one name",
			edit: func(text string) string {
				return strings.Replace(text, "x-osapi-method: delete", "x-osapi-method: list", 1)
			},
			want: "Job has two methods with the name List",
		},
		{
			name: "when the successes of an operation have different bodies",
			edit: func(text string) string {
				return strings.Replace(text, `        "201":
          description: created
          content:
            application/json:
              schema: {$ref: '#/components/schemas/FileUploadResponse'}`, `        "201":
          description: created`, 1)
			},
			want: "PostFile: its successes have different bodies",
		},
		{
			name: "when a parameter is not a scalar",
			edit: func(text string) string {
				return strings.Replace(
					text,
					"schema: {type: boolean}",
					"schema: {type: array, items: {type: string}}",
					1,
				)
			},
			want: "parameter all is a []string",
		},
		{
			name: "when the body is not a schema of the components",
			edit: func(text string) string {
				return strings.Replace(text,
					"schema: {$ref: '#/components/schemas/DnsUpdateRequest'}",
					"schema: {type: object}", 1)
			},
			want: "the body is not a schema of the components",
		},
		{
			name: "when the body and the path have one field",
			edit: func(text string) string {
				return strings.Replace(text, "override_dhcp: {type: boolean}",
					"interface_name: {type: string}", 1)
			},
			want: "the body and the path both have interface_name",
		},
		{
			name: "when a body is not JSON and not a form",
			edit: func(text string) string {
				return strings.Replace(text, "multipart/form-data:", "text/plain:", 1)
			},
			want: "the body is not JSON and not a multipart form",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			_, err := sdkmodel.New(s.load(tc.edit))
			s.ErrorContains(err, tc.want)
		})
	}
}

func (s *ModelPublicTestSuite) TestGoName() {
	tests := []struct {
		in   string
		want string
	}{
		{in: "search_domains", want: "SearchDomains"},
		{in: "job_id", want: "JobID"},
		{in: "15min", want: "Min15"},
		{in: "ipv4", want: "IPv4"},
		{in: "network_dns", want: "NetworkDNS"},
		{in: "user_sshkey", want: "UserSSHKey"},
		{in: "content-type", want: "ContentType"},
	}

	for _, tc := range tests {
		s.Run(tc.in, func() {
			s.Equal(tc.want, sdkmodel.GoName(tc.in))
		})
	}
}

func TestModelPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ModelPublicTestSuite))
}
