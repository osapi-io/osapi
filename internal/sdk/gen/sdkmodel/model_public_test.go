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
			name: "when the areas are those of x-osapi-areas, in its order",
			validateFunc: func() {
				s.Equal([]*sdkmodel.Area{
					{Name: "config", Label: "Config"},
					{Name: "networking", Label: "Networking"},
					{Name: "files", Label: "Files"},
					{Name: "fleet", Label: "Fleet"},
					{Name: "audit", Label: "Audit"},
				}, m.Areas)
			},
		},
		{
			name: "when a service has the area of its operations",
			validateFunc: func() {
				s.Equal("networking", services["NetworkDNS"].Area)
				s.Equal("files", services["File"].Area)
				s.Equal("config", services["NodeFile"].Area)
				s.Equal("fleet", services["Job"].Area)
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
			name: "when a method has the examples of its definition",
			validateFunc: func() {
				list := services["Job"].Methods[0]
				s.Equal([]*sdkmodel.Example{
					{Doc: "Every job", Input: map[string]any{}},
					{Doc: "Three jobs", Input: map[string]any{"limit": float64(3), "all": true}},
				}, list.Examples)
				s.Empty(services["File"].Methods[0].Examples)
			},
		},
		{
			name: "when a service of a node is marked as one",
			validateFunc: func() {
				s.True(services["NetworkDNS"].Node)
				s.True(services["NodeFile"].Node)
				s.False(services["File"].Node)
				s.False(services["Job"].Node)
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
			name: "when the spec has no x-osapi-areas",
			edit: func(text string) string {
				i := strings.Index(text, "x-osapi-areas:")
				j := strings.Index(text, "paths:\n")

				return text[:i] + text[j:]
			},
			want: "the spec has no x-osapi-areas",
		},
		{
			name: "when x-osapi-areas is not a list",
			edit: func(text string) string {
				i := strings.Index(text, "x-osapi-areas:")
				j := strings.Index(text, "paths:\n")

				return text[:i] + "x-osapi-areas: many\n" + text[j:]
			},
			want: "x-osapi-areas: json",
		},
		{
			name: "when an operation has no x-osapi-area",
			edit: func(text string) string {
				return strings.Replace(text, "      x-osapi-area: fleet\n", "", 1)
			},
			want: `GetJobs: the x-osapi-area "" is not one of the spec's x-osapi-areas`,
		},
		{
			name: "when the operations of a tag are in two areas",
			edit: func(text string) string {
				i := strings.LastIndex(text, "      x-osapi-area: fleet\n")

				return text[:i] + "      x-osapi-area: files\n" + text[i+len("      x-osapi-area: fleet\n"):]
			},
			want: "the operations of tag Job_API are in the areas fleet and files",
		},
		{
			name: "when an operation has no x-osapi-method",
			edit: func(text string) string {
				return strings.Replace(text, "      x-osapi-method: list\n", "", 1)
			},
			want: "GetJobs: the operation has no x-osapi-method",
		},
		{
			name: "when the examples of an operation are not a list",
			edit: func(text string) string {
				return strings.Replace(text, `      x-osapi-examples:
        - doc: A job
          input: {id: 550e8400-e29b-41d4-a716-446655440000}
`, "      x-osapi-examples: many\n", 1)
			},
			want: "DeleteJobByID: x-osapi-examples",
		},
		{
			name: "when an operation has no tag of the spec",
			edit: func(text string) string {
				return strings.Replace(text, "tags: [Job_API]", "tags: [Other]", 1)
			},
			want: "GetJobs: the operation has no tag of the spec",
		},
		{
			name: "when a tag has no x-osapi-group",
			edit: func(text string) string {
				return strings.Replace(text, "    x-osapi-group: job\n", "", 1)
			},
			want: "tag Job_API has no x-osapi-group",
		},
		{
			name: "when a path has no segment of its group",
			edit: func(text string) string {
				return strings.Replace(text, "x-osapi-group: job\n", "x-osapi-group: work\n", 1)
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
