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

import "github.com/osapi-io/osapi/pkg/sdk/client/gen"

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
	Items      []FileItem `json:"items"`
	TotalItems int        `json:"total_items"`
}

// FileMetadata represents metadata for a single file.
type FileMetadata struct {
	Name        string `json:"name"`
	SHA256      string `json:"sha256"`
	Size        int    `json:"size"`
	ContentType string `json:"content_type"`
}

// FileChanged represents the result of a change detection check.
type FileChanged struct {
	Name    string `json:"name"`
	Changed bool   `json:"changed"`
	SHA256  string `json:"sha256"`
}

// FileDeployResult is the deploy of one file on one host. A host that failed
// or skipped has a row with its status and no file.
type FileDeployResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Path     string `json:"path,omitempty"`
	SHA256   string `json:"sha256,omitempty"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

// FileUndeployResult is the undeploy of one file on one host.
type FileUndeployResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Path     string `json:"path,omitempty"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

// FileStatusResult is one deployed file on one host against what osapi
// wrote: State is in-sync, drifted, missing or unmanaged.
type FileStatusResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Path     string `json:"path,omitempty"`
	State    string `json:"state,omitempty"`
	SHA256   string `json:"sha256,omitempty"`
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
	Items      []StaleDeployment `json:"items"`
	TotalItems int               `json:"total_items"`
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
	files := make([]FileItem, 0, len(g.Items))
	for _, f := range g.Items {
		files = append(files, FileItem{
			Name:        f.Name,
			SHA256:      f.Sha256,
			Size:        f.Size,
			ContentType: f.ContentType,
			Source:      f.Source,
		})
	}

	return FileList{
		Items:      files,
		TotalItems: g.TotalItems,
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

// fileDeployCollectionFromGen converts the rows of a deploy.
func fileDeployCollectionFromGen(
	g *gen.FileUpdateResponse,
) Collection[FileDeployResult] {
	results := make([]FileDeployResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, FileDeployResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Path:     derefString(r.Path),
			SHA256:   derefString(r.Sha256),
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

// fileUndeployCollectionFromGen converts the rows of an undeploy.
func fileUndeployCollectionFromGen(
	g *gen.FileDeleteResponse,
) Collection[FileUndeployResult] {
	results := make([]FileUndeployResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, FileUndeployResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Path:     derefString(r.Path),
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

// fileStatusCollectionFromGen converts the rows of a status.
func fileStatusCollectionFromGen(
	g *gen.FileStatusResponse,
) Collection[FileStatusResult] {
	results := make([]FileStatusResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, FileStatusResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Path:     derefString(r.Path),
			State:    derefString(r.State),
			SHA256:   derefString(r.Sha256),
			Error:    derefString(r.Error),
		})
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
	items := make([]StaleDeployment, 0, len(g.Items))
	for _, s := range g.Items {
		items = append(items, staleDeploymentFromGen(s))
	}

	return StaleList{Items: items, TotalItems: g.TotalItems}
}
