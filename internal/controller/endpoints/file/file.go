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
// Package file is the endpoint group for the files in the Object Store:
// what a deploy writes to a host. Until the group moves, the controller
// serves the hand-written spec in internal/controller/api/file, and
// TestEndpointParity holds the two together.
package file

import (
	"errors"
	"net/http"

	"github.com/osapi-io/osapi/internal/controller/endpoint"
)

var (
	// ErrNotFound is a file that is not there.
	ErrNotFound = errors.New("file not found")
	// ErrExists is a file that exists with other content.
	ErrExists = errors.New("file exists with other content")
	// ErrProtected is a file that osapi owns.
	ErrProtected = errors.New("file is protected")
)

// Group is the file group's definition.
var Group = endpoint.Register(endpoint.Define(
	"file", "The files in the Object Store",
	endpoint.Title("File Management API"),
	endpoint.Tag("file_operations", "File", "Object Store file management operations."),
	endpoint.Get[ListInput, FileListResponse]("list").
		ID("GetFiles").
		Describe("List stored files").
		Explain("List all files stored in the Object Store.").
		Returns("List of stored files."),
	endpoint.Post[UploadInput, FileUploadResponse]("upload").
		ID("PostFile").
		Describe("Upload a file").
		Explain("Upload a file to the Object Store.").
		Returns("The file is new.").
		Status(http.StatusCreated).
		Location("name").
		Answers(http.StatusOK, "The file exists with the same content. Nothing changed.").
		Errors(
			endpoint.Conflict(ErrExists,
				"The file exists with different content. Replace it with PUT /api/file/{name}."),
			endpoint.Forbidden(ErrProtected, "The file is an osapi file, which a caller cannot write."),
		),
	endpoint.Get[GetInput, FileInfoResponse]("get").
		ID("GetFileByName").
		Describe("Get file metadata").
		Explain("Get metadata for a specific file in the Object Store.").
		Returns("The metadata of the file.").
		Errors(endpoint.NotFound(ErrNotFound, "File not found.")),
	endpoint.Put[UpdateInput, FileUploadResponse]("update").
		ID("PutFileByName").
		Describe("Replace a file").
		Explain("Replace a file in the Object Store, or create it when the name is new.").
		Returns("The file is new.").
		Status(http.StatusCreated).
		Location("name").
		Answers(http.StatusOK, "The file existed. changed says whether its content changed.").
		Errors(endpoint.Forbidden(ErrProtected, "The file is an osapi file, which a caller cannot write.")),
	endpoint.Delete[DeleteInput, struct{}]("delete").
		ID("DeleteFileByName").
		Describe("Delete a file").
		Explain("Delete a file from the Object Store.").
		Returns("File deleted.").
		Status(http.StatusNoContent).
		Errors(
			endpoint.NotFound(ErrNotFound, "File not found."),
			endpoint.Forbidden(ErrProtected, "The file is an osapi file, which a caller cannot delete."),
		),
	endpoint.Get[StaleInput, StaleDeploymentsResponse]("stale").
		Sub("stale").
		ID("GetFileStale").
		Describe("List stale file deployments").
		Explain("Returns deployments where the source object in the Object Store has been "+
			"updated since the file was last deployed. Controller-only check that compares "+
			"state KV SHA-256 hashes against current object content.").
		Returns("The stale deployments."),
))
