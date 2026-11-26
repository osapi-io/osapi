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
package file

import "github.com/osapi-io/osapi/pkg/sdk/types"

// ListInput is the input of list, which has none.
type ListInput struct{}

// StaleInput is the input of stale, which has none.
type StaleInput struct{}

// GetInput names the file to read.
type GetInput struct {
	Name string `json:"name" in:"path" validate:"required,min=1,max=255" doc:"The name of the file in the Object Store."`
}

// DeleteInput names the file to delete.
type DeleteInput struct {
	Name string `json:"name" in:"path" validate:"required,min=1,max=255" doc:"The name of the file in the Object Store."`
}

// UploadInput is a new file: its name, how a deploy treats it, and its
// content, as a multipart form.
type UploadInput struct {
	Name        string            `json:"name"         in:"form" validate:"required,min=1,max=255" doc:"The name of the file in the Object Store."                                                                                     example:"nginx.conf"`
	ContentType types.ContentType `json:"content_type" in:"form" validate:"omitempty,enum"         doc:"How the file should be treated during deploy. raw writes bytes as-is; template renders with Go text/template and agent facts."                      default:"raw"`
	File        []byte            `json:"file"         in:"form" validate:"required,min=1"         doc:"The file content."`
}

// UpdateInput is the new content of a file that the path names.
type UpdateInput struct {
	Name        string            `json:"name"         in:"path" validate:"required,min=1,max=255" doc:"The name of the file in the Object Store."`
	ContentType types.ContentType `json:"content_type" in:"form" validate:"omitempty,enum"         doc:"How the file should be treated during deploy. raw writes bytes as-is; template renders with Go text/template and agent facts." default:"raw"`
	File        []byte            `json:"file"         in:"form" validate:"required,min=1"         doc:"The file content."`
}

// Info is a stored file.
type Info struct {
	_ struct{} `schema:"FileInfo"`

	Name        string            `json:"name"         doc:"The name of the file."                                           example:"nginx.conf"`
	Sha256      string            `json:"sha256"       doc:"SHA-256 hash of the file content."                               example:"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"`
	Size        int               `json:"size"         doc:"File size in bytes."                                             example:"1024"                                                             render:"bytes"`
	ContentType types.ContentType `json:"content_type" doc:"How the file should be treated during deploy (raw or template)." example:"raw"                                                                             table:"-"`
	Source      string            `json:"source"       doc:"Source of the file: 'system' or 'user'."                         example:"user"`
}

// InfoResponse is the metadata of a file.
type InfoResponse struct {
	_ struct{} `schema:"FileInfoResponse"`

	Name        string            `json:"name"         doc:"The name of the file."                                           example:"nginx.conf"`
	Sha256      string            `json:"sha256"       doc:"SHA-256 hash of the file content."                               example:"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"`
	Size        int               `json:"size"         doc:"File size in bytes."                                             example:"1024"                                                             render:"bytes"`
	ContentType types.ContentType `json:"content_type" doc:"How the file should be treated during deploy (raw or template)." example:"raw"`
}

// ListResponse is each stored file.
type ListResponse struct {
	_ struct{} `schema:"FileListResponse"`

	Items      []Info `json:"items"       doc:"List of stored files."  empty:"Upload one with osapi client file upload."`
	TotalItems int    `json:"total_items" doc:"Total number of files."                                                   example:"5"`
}

// UploadResponse is the file that an upload or an update wrote.
type UploadResponse struct {
	_ struct{} `schema:"FileUploadResponse"`

	Name        string            `json:"name"         doc:"The name of the uploaded file."                                                                                     example:"nginx.conf"`
	Sha256      string            `json:"sha256"       doc:"SHA-256 hash of the file content."                                                                                  example:"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"`
	Size        int               `json:"size"         doc:"File size in bytes."                                                                                                example:"1024"                                                             render:"bytes"`
	Changed     bool              `json:"changed"      doc:"Whether the file content changed. False when the Object Store already held an object with the same SHA-256 digest." example:"true"`
	ContentType types.ContentType `json:"content_type" doc:"How the file should be treated during deploy (raw or template)."                                                    example:"raw"`
}

// StaleDeployment is a deploy whose object changed since.
type StaleDeployment struct {
	ObjectName  string `json:"object_name"  doc:"Name of the object in the Object Store."           example:"hello-echo"                                   label:"Object"`
	Hostname    string `json:"hostname"     doc:"Host where the file is deployed."                  example:"web-01"`
	Provider    string `json:"provider"     doc:"Provider that manages this deployment."            example:"service"`
	DeployedAt  string `json:"deployed_at"  doc:"When the file was last deployed."                  example:"2026-04-01T18:00:00Z"                         label:"Deployed" render:"time"`
	Path        string `json:"path"         doc:"Deployment path on the target host."               example:"/etc/systemd/system/osapi-hello-echo.service"                                    table:"-"`
	DeployedSha string `json:"deployed_sha" doc:"SHA-256 of the object when it was deployed."       example:"abc123def456"                                                  render:"short=12"`
	CurrentSha  string `json:"current_sha"  doc:"SHA-256 hash of the current object store content." example:"789abc012def"                                                  render:"short=12"           none:"(deleted)"`
}

// StaleDeploymentsResponse is each stale deploy.
type StaleDeploymentsResponse struct {
	_          struct{}          `title:"Stale deployments"`
	Items      []StaleDeployment `                          json:"items"       doc:"The stale deployments."             empty:"Every deployment is in sync."`
	TotalItems int               `                          json:"total_items" doc:"Total number of stale deployments."`
}
