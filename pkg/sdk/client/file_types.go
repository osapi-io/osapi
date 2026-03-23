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

package client

import "github.com/retr0h/osapi/pkg/sdk/client/gen"

// FileUpload represents a successfully uploaded file.
type FileUpload struct {
	Name        string `json:"name"`
	SHA256      string `json:"sha256"`
	Size        int    `json:"size"`
	Changed     bool   `json:"changed"`
	ContentType string `json:"content_type"`
}

// FileItem represents file metadata in a list.
type FileItem struct {
	Name        string `json:"name"`
	SHA256      string `json:"sha256"`
	Size        int    `json:"size"`
	ContentType string `json:"content_type"`
	Source      string `json:"source"`
}

// FileList is a collection of files with total count.
type FileList struct {
	Files []FileItem `json:"files"`
	Total int        `json:"total"`
}

// FileMetadata represents metadata for a single file.
type FileMetadata struct {
	Name        string `json:"name"`
	SHA256      string `json:"sha256"`
	Size        int    `json:"size"`
	ContentType string `json:"content_type"`
}

// FileDelete represents the result of a file deletion.
type FileDelete struct {
	Name    string `json:"name"`
	Deleted bool   `json:"deleted"`
}

// FileChanged represents the result of a change detection check.
type FileChanged struct {
	Name    string `json:"name"`
	Changed bool   `json:"changed"`
	SHA256  string `json:"sha256"`
}

// FileDeployResult represents the result of a file deploy operation.
type FileDeployResult struct {
	JobID    string `json:"job_id"`
	Hostname string `json:"hostname"`
	Changed  bool   `json:"changed"`
}

// FileUndeployResult represents the result of a file undeploy operation.
type FileUndeployResult struct {
	JobID    string `json:"job_id"`
	Hostname string `json:"hostname"`
	Changed  bool   `json:"changed"`
}

// FileStatusResult represents the result of a file status check.
type FileStatusResult struct {
	JobID    string `json:"job_id"`
	Hostname string `json:"hostname"`
	Path     string `json:"path"`
	Status   string `json:"status"`
	SHA256   string `json:"sha256,omitempty"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

// fileUploadFromGen converts a gen.FileUploadResponse to a FileUpload.
func fileUploadFromGen(
	g *gen.FileUploadResponse,
) FileUpload {
	return FileUpload{
		Name:        g.Name,
		SHA256:      g.Sha256,
		Size:        g.Size,
		Changed:     g.Changed,
		ContentType: g.ContentType,
	}
}

// fileListFromGen converts a gen.FileListResponse to a FileList.
func fileListFromGen(
	g *gen.FileListResponse,
) FileList {
	files := make([]FileItem, 0, len(g.Files))
	for _, f := range g.Files {
		files = append(files, FileItem{
			Name:        f.Name,
			SHA256:      f.Sha256,
			Size:        f.Size,
			ContentType: f.ContentType,
			Source:      f.Source,
		})
	}

	return FileList{
		Files: files,
		Total: g.Total,
	}
}

// fileMetadataFromGen converts a gen.FileInfoResponse to a FileMetadata.
func fileMetadataFromGen(
	g *gen.FileInfoResponse,
) FileMetadata {
	return FileMetadata{
		Name:        g.Name,
		SHA256:      g.Sha256,
		Size:        g.Size,
		ContentType: g.ContentType,
	}
}

// fileDeleteFromGen converts a gen.FileDeleteResponse to a FileDelete.
func fileDeleteFromGen(
	g *gen.FileDeleteResponse,
) FileDelete {
	return FileDelete{
		Name:    g.Name,
		Deleted: g.Deleted,
	}
}

// fileDeployResultFromGen converts a gen.FileDeployResponse to a FileDeployResult.
func fileDeployResultFromGen(
	g *gen.FileDeployResponse,
) FileDeployResult {
	return FileDeployResult{
		JobID:    g.JobId,
		Hostname: g.Hostname,
		Changed:  g.Changed,
	}
}

// fileUndeployResultFromGen converts a gen.FileUndeployResponse to a FileUndeployResult.
func fileUndeployResultFromGen(
	g *gen.FileUndeployResponse,
) FileUndeployResult {
	return FileUndeployResult{
		JobID:    g.JobId,
		Hostname: g.Hostname,
		Changed:  g.Changed,
	}
}

// fileStatusResultFromGen converts a gen.FileStatusResponse to a FileStatusResult.
func fileStatusResultFromGen(
	g *gen.FileStatusResponse,
) FileStatusResult {
	r := FileStatusResult{
		JobID:    g.JobId,
		Hostname: g.Hostname,
		Path:     g.Path,
		Status:   g.Status,
		Changed:  derefBool(g.Changed),
		Error:    derefString(g.Error),
	}

	if g.Sha256 != nil {
		r.SHA256 = *g.Sha256
	}

	return r
}
