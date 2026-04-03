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

// FileDeployResult represents the result of a file deploy operation for a
// single host in a collection response.
type FileDeployResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

// FileUndeployResult represents the result of a file undeploy operation for a
// single host in a collection response.
type FileUndeployResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

// FileStatusResult represents the result of a file status check for a single
// host in a collection response.
type FileStatusResult struct {
	Hostname string `json:"hostname"`
	Path     string `json:"path,omitempty"`
	Status   string `json:"status,omitempty"`
	SHA256   string `json:"sha256,omitempty"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

// StaleDeployment represents a deployment that is out of sync
// with the current object store content.
type StaleDeployment struct {
	ObjectName  string `json:"object_name"`
	Hostname    string `json:"hostname"`
	Provider    string `json:"provider"`
	Path        string `json:"path"`
	DeployedSHA string `json:"deployed_sha"`
	CurrentSHA  string `json:"current_sha"`
	DeployedAt  string `json:"deployed_at"`
}

// StaleList is a list of stale deployments.
type StaleList struct {
	Stale []StaleDeployment `json:"stale"`
	Total int               `json:"total"`
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

// fileDeployCollectionFromGen converts a gen.FileDeployCollectionResponse to
// a Collection[FileDeployResult].
func fileDeployCollectionFromGen(
	g *gen.FileDeployCollectionResponse,
) Collection[FileDeployResult] {
	results := make([]FileDeployResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, FileDeployResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Changed:  derefBool(r.Changed),
			Error:    derefString(r.Error),
		})
	}

	c := Collection[FileDeployResult]{Results: results}
	if g.JobId != nil {
		c.JobID = g.JobId.String()
	}

	return c
}

// fileUndeployCollectionFromGen converts a gen.FileUndeployCollectionResponse
// to a Collection[FileUndeployResult].
func fileUndeployCollectionFromGen(
	g *gen.FileUndeployCollectionResponse,
) Collection[FileUndeployResult] {
	results := make([]FileUndeployResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, FileUndeployResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Changed:  derefBool(r.Changed),
			Error:    derefString(r.Error),
		})
	}

	c := Collection[FileUndeployResult]{Results: results}
	if g.JobId != nil {
		c.JobID = g.JobId.String()
	}

	return c
}

// fileStatusCollectionFromGen converts a gen.FileStatusCollectionResponse to a
// Collection[FileStatusResult].
func fileStatusCollectionFromGen(
	g *gen.FileStatusCollectionResponse,
) Collection[FileStatusResult] {
	results := make([]FileStatusResult, 0, len(g.Results))
	for _, r := range g.Results {
		item := FileStatusResult{
			Hostname: r.Hostname,
			Path:     derefString(r.Path),
			Status:   derefString(r.Status),
			Changed:  derefBool(r.Changed),
			Error:    derefString(r.Error),
		}

		if r.Sha256 != nil {
			item.SHA256 = *r.Sha256
		}

		results = append(results, item)
	}

	c := Collection[FileStatusResult]{Results: results}
	if g.JobId != nil {
		c.JobID = g.JobId.String()
	}

	return c
}

// staleDeploymentFromGen converts a gen.StaleDeployment to a StaleDeployment.
func staleDeploymentFromGen(
	g gen.StaleDeployment,
) StaleDeployment {
	return StaleDeployment{
		ObjectName:  g.ObjectName,
		Hostname:    g.Hostname,
		Provider:    g.Provider,
		Path:        g.Path,
		DeployedSHA: g.DeployedSha,
		CurrentSHA:  g.CurrentSha,
		DeployedAt:  g.DeployedAt,
	}
}

// staleListFromGen converts a gen.StaleDeploymentsResponse to a StaleList.
func staleListFromGen(
	g *gen.StaleDeploymentsResponse,
) StaleList {
	items := make([]StaleDeployment, 0, len(g.Stale))
	for _, s := range g.Stale {
		items = append(items, staleDeploymentFromGen(s))
	}

	return StaleList{Stale: items, Total: g.Total}
}
