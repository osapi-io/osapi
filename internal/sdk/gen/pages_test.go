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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type PagesTestSuite struct {
	genSuite
}

// handWritten is a file of the SDK that a person wrote: two methods of the
// File service, a method of a generic type, and a function.
const handWritten = `package client

// Upload sends a file. It reads all of it first.
func (s *FileService) Upload(
	ctx context.Context,
	name string,
) (*Response[FileUploadResponse], error) {
	return nil, nil
}

// GetChanged compares the content with the stored file.
func (s *FileService) GetChanged(ctx context.Context) error { return nil }

// Wait waits.
func (s *FileService) Wait() {}

func (s *FileService) send() {}

// RawJSON is the body.
func (r *Response[T]) RawJSON() []byte { return nil }

func helper() {}
`

func (s *PagesTestSuite) SetupTest() {
	s.genSuite.SetupTest()
	s.Require().NoError(generate(s.root, false))
	s.write(dir+"/file_upload.go", handWritten)
	s.write(dir+"/file_upload_test.go", "package client\n")
	s.write(dir+"/notes.txt", "")
	s.write(pageDir+"/"+indexPage, "# Client\n\n<!-- osapi:sdk-services -->\nold\n"+
		"<!-- /osapi:sdk-services -->\n\nThe end.\n")
	s.write(pageDir+"/gone.md", pageHeader+"\na page of a service that is gone\n")
	s.write(pageDir+"/notes.md", "a page that a person wrote\n")
	s.write(pageDir+"/kitchen/"+categoryFile, "{}\n")
	s.write(pageDir+"/kitchen/sink.md", pageHeader+"\na page of an area that is gone\n")
	s.write(pageDir+"/kitchen/notes.md", "a page that a person wrote\n")
	s.write(pageDir+"/"+libraryDir+"/"+categoryFile, "{}\n")
	s.write(pageDir+"/pantry/"+categoryFile, "{}\n")
	s.write(pageDir+"/pantry/jar.md", pageHeader+"\na page of an area that is gone\n")
	s.write(
		exampleDir+"/file_upload.go",
		"package main\n\nc.File.Upload(ctx)\nc.File.GetChanged(ctx)\n",
	)
	s.write(exampleDir+"/go.mod", "module examples\n")
	// A directory with the name of a program is not a program.
	s.Require().NoError(os.MkdirAll(filepath.Join(s.root, exampleDir, "dir.go"), 0o755))
}

func (s *PagesTestSuite) page(
	rel string,
) string {
	data, err := os.ReadFile(filepath.Join(s.root, pageDir, rel))
	s.Require().NoError(err)

	return string(data)
}

func (s *PagesTestSuite) TestRunPages() {
	s.Require().NoError(generate(s.root, true))

	tests := []struct {
		name         string
		validateFunc func()
	}{
		{
			name: "when a page has the methods, their permissions and their examples",
			validateFunc: func() {
				dns := s.page("networking/network_dns.md")
				s.Contains(dns, "title: NetworkDNS\n---\n\n"+pageHeader)
				s.Contains(dns, "`c.NetworkDNS` calls the API's NetworkDNS group. "+
					"The DNS resolvers of each interface of the host.")
				s.Contains(dns, "| [`Get(ctx, hostname, interfaceName)`](#get) | "+
					"`*Response[Collection[DnsEntry]]` | `dns:read` | "+
					"Get the DNS resolvers of an interface. |")
				s.Contains(dns, "func (s *NetworkDNSService) Get(\n\tctx context.Context,")
				s.Contains(
					dns,
					"It sends `GET /api/node/{hostname}/network/dns/{interface_name}`, "+
						"and needs the permission `dns:read`.",
				)
				s.Contains(dns, "and needs no permission.")
				s.Contains(dns, "An interface, `go run network_dns.go get`:\n\n```go\n"+
					"resp, err := c.NetworkDNS.Get(ctx, target, \"eth0\")\n```")
				s.Contains(dns, "### NetworkDNSUpdateInput")
				s.Contains(dns, "| `InterfaceName` | `string` | The interface_name of the path. |")
				s.Contains(dns, "### DnsEntry\n\nThe resolvers of one host.\n")
				s.Contains(dns, "It is one of: ok, failed.")
				s.Contains(dns, "### DnsKind\n\nThe kind of a resolver.\n\nA `string`.\n")
				s.Contains(dns, "[`examples/sdk/client/network_dns.go`]("+exampleURL+
					"network_dns.go) runs each example of this page.")
			},
		},
		{
			name: "when a method has no answer, its example has only an error",
			validateFunc: func() {
				job := s.page("fleet/job.md")
				s.Contains(
					job,
					"err := c.Job.Delete(ctx, \"550e8400-e29b-41d4-a716-446655440000\")",
				)
				s.Contains(job, "| `error` |")
			},
		},
		{
			name: "when a person wrote a method, the page has its doc comment",
			validateFunc: func() {
				file := s.page("files/file.md")
				s.Contains(
					file,
					"| [`Upload(ctx, name)`](#upload) | `*Response[FileUploadResponse]` | "+
						" | Upload sends a file. |",
				)
				s.Contains(file, "Upload sends a file. It reads all of it first.\n\n### GetChanged")
				s.Contains(file, "| [`GetChanged(ctx)`](#getchanged) | `error` |  | "+
					"GetChanged compares the content with the stored file. |")
				s.Contains(file, "It is not an operation of the API: it runs in the client.")
				s.Contains(file, "| [`Wait()`](#wait) |  |")
				s.NotContains(file, "send(")
				s.Contains(file, "[`examples/sdk/client/file_upload.go`]("+exampleURL+
					"file_upload.go) calls the methods that a person wrote.")
				s.NotContains(file, "runs each example of this page")
			},
		},
		{
			name: "when client.md has a table of the services of each area, in the areas' order",
			validateFunc: func() {
				index := s.page(indexPage)
				s.Contains(index, "<!-- osapi:sdk-services -->\n\n### Config\n\n"+
					"| Service | Description |\n| --- | --- |\n"+
					"| [NodeFile](../config/node_file.md) | The files of the host. |\n\n"+
					"### Networking\n\n")
				s.Contains(
					index,
					"| [NetworkDNS](../networking/network_dns.md) | "+
						"The DNS resolvers of each interface of the host. |",
				)
				s.Contains(index, "### Files\n\n")
				s.Contains(index, "### Fleet\n\n| Service | Description |\n| --- | --- |\n"+
					"| [Job](../fleet/job.md) | The job queue. |")
				s.NotContains(index, "### Audit", "an area with no service has no table")
				s.NotContains(index, "old")
				s.Contains(index, "The end.")
			},
		},
		{
			name: "when the folder of each area has its label and its order",
			validateFunc: func() {
				s.Equal("{\n  \"label\": \"Config\",\n  \"position\": 1\n}\n",
					s.page("config/"+categoryFile))
				s.Equal("{\n  \"label\": \"Fleet\",\n  \"position\": 4\n}\n",
					s.page("fleet/"+categoryFile))
				s.NoDirExists(filepath.Join(s.root, pageDir, "audit"),
					"an area with no service has no folder")
			},
		},
		{
			name: "when a stale generated page goes, and a page a person wrote stays",
			validateFunc: func() {
				s.NoFileExists(filepath.Join(s.root, pageDir, "gone.md"))
				s.FileExists(filepath.Join(s.root, pageDir, "notes.md"))
				s.NoDirExists(filepath.Join(s.root, pageDir, "node"))
				s.NoFileExists(filepath.Join(s.root, pageDir, "kitchen", "sink.md"))
				s.NoFileExists(filepath.Join(s.root, pageDir, "kitchen", categoryFile))
				s.FileExists(filepath.Join(s.root, pageDir, "kitchen", "notes.md"))
				s.FileExists(filepath.Join(s.root, pageDir, libraryDir, categoryFile),
					"the library folder is not the generator's")
				s.NoDirExists(filepath.Join(s.root, pageDir, "pantry"),
					"the folder of an area that has no page now is removed")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.validateFunc()
		})
	}
}

func (s *PagesTestSuite) TestRunPagesErrors() {
	tests := []struct {
		name    string
		setup   func()
		wantErr string
	}{
		{
			name:    "when the SDK is not there",
			setup:   func() { s.Require().NoError(os.RemoveAll(filepath.Join(s.root, dir))) },
			wantErr: "no such file or directory",
		},
		{
			name:    "when a file of the SDK is not Go",
			setup:   func() { s.write(dir+"/bad.go", "package client\n\nfunc {") },
			wantErr: "bad.go",
		},
		{
			name: "when an example has no value for a key",
			setup: func() {
				// A root of its own: the loader of the spec keeps each file it
				// read by its path, and the setup read this one.
				data, err := os.ReadFile("sdkmodel/testdata/spec.yaml")
				s.Require().NoError(err)
				s.root = s.T().TempDir()
				s.write(spec, replaceOnce(string(data),
					"input: {id: 550e8400-e29b-41d4-a716-446655440000}", "input: {}"))
				s.write(dir+"/client.go", "package client\n")
			},
			wantErr: "it has no id",
		},
		{
			name: "when client.md is not there",
			setup: func() {
				s.Require().NoError(os.Remove(filepath.Join(s.root, pageDir, indexPage)))
			},
			wantErr: "client.md",
		},
		{
			name: "when the folder of an area is a file",
			setup: func() {
				s.Require().NoError(os.RemoveAll(filepath.Join(s.root, pageDir, "fleet")))
				s.write(pageDir+"/fleet", "a file that a person wrote\n")
			},
			wantErr: "fleet",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.setup()

			s.ErrorContains(generate(s.root, true), tc.wantErr)
		})
	}
}

func (s *PagesTestSuite) TestExamplesSection() {
	s.Empty(examplesSection(nil, false, nil), "a service with no program has no section")
}

func TestPagesTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(PagesTestSuite))
}
