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
// what a deploy writes to a host.
package file

import (
	"errors"
	"net/http"
	"strings"

	"github.com/osapi-io/osapi/internal/area"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
)

var (
	// ErrNotFound is a file that is not there.
	ErrNotFound = errors.New("file not found")
	// ErrExists is a file that exists with other content.
	ErrExists = errors.New("file exists with other content")
	// ErrProtected is a file that osapi owns.
	ErrProtected = errors.New("file is protected")
	// ErrNoStore is a controller with no Object Store.
	ErrNoStore = errors.New("the Object Store is not configured")
	// ErrNoState is a controller with no file state KV.
	ErrNoState = errors.New("the file state KV is not configured")
)

// notAContentType is a content type that the API refuses, for the Fails of
// the upload and the update.
const notAContentType = "zip"

// noStore503 is the 503 of each method, when the controller has no Object
// Store.
var noStore503 = endpoint.Unavailable(ErrNoStore,
	"The Object Store is not configured on the controller.")

// Group is the file group's definition.
var Group = endpoint.Register(endpoint.Define(
	"file",
	"Manage the files in the Object Store",
	endpoint.Area(area.ObjectStore),
	endpoint.SubDoc("stale", "Read the deployments whose file changed since"),
	endpoint.Title("File Management API"),
	endpoint.Tag("file_operations", "File", "Object Store file management operations."),
	endpoint.Uses[Deps](),
	endpoint.Get[ListInput, ListResponse]("list").
		Handle(List).
		ID("GetFiles").
		Describe("List stored files").
		Explain("List all files stored in the Object Store.").
		Returns("List of stored files.").
		Command("List stored files", "").
		Example("").
		Prints(sampleList).
		Example("--json").
		Errors(noStore503),
	endpoint.Post[UploadInput, UploadResponse]("upload").
		Fails("no name", UploadInput{File: content}).
		Fails("no file", UploadInput{Name: "app.conf"}).
		Fails("a name over 255 characters", UploadInput{Name: long, File: content}).
		Fails("a content type that is not raw or template", UploadInput{Name: "app.conf", ContentType: notAContentType, File: content}).
		Fails("an empty file", UploadInput{Name: "app.conf", File: []byte{}}).
		Handle(Upload).
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
			noStore503,
		).
		Command("Upload a file to the Object Store",
			"An upload of the same content changes nothing (`Changed: false`). If the\n"+
				"file exists with different content, the command rejects the upload with 409\n"+
				"Conflict. `file update` replaces the file.").
		Example("--name app.conf --file /tmp/app.conf").
		Prints(sampleUpload).
		Example("--name app.conf.tmpl --file /tmp/app.conf.tmpl --content-type template").
		Prints(sampleUploadTemplate),
	endpoint.Get[GetInput, InfoResponse]("get").
		Fails("a name over 255 characters", GetInput{Name: long}).
		Handle(Get).
		ID("GetFileByName").
		Describe("Get file metadata").
		Explain("Get metadata for a specific file in the Object Store.").
		Returns("The metadata of the file.").
		Command("Get file metadata", "").
		Example("--name app.conf").
		Prints(sampleInfo).
		Example("--name app.conf --json").
		Errors(endpoint.NotFound(ErrNotFound, "File not found."), noStore503),
	endpoint.Put[UpdateInput, UploadResponse]("update").
		Fails("no file", UpdateInput{Name: "app.conf"}).
		Fails("a name over 255 characters", UpdateInput{Name: long, File: content}).
		Fails("a content type that is not raw or template", UpdateInput{Name: "app.conf", ContentType: notAContentType, File: content}).
		Fails("an empty file", UpdateInput{Name: "app.conf", File: []byte{}}).
		Handle(Update).
		ID("PutFileByName").
		Describe("Replace a file").
		Explain("Replace a file in the Object Store, or create it when the name is new.").
		Returns("The file is new.").
		Status(http.StatusCreated).
		Location("name").
		Answers(http.StatusOK, "The file existed. changed says whether its content changed.").
		Command("Replace a file in the Object Store",
			"The command replaces the content of a stored file, or stores it when the\n"+
				"name is new. The same content changes nothing (`Changed: false`).").
		Example("--name app.conf --file /tmp/app.conf").
		Prints(sampleUpload).
		Errors(
			endpoint.Forbidden(ErrProtected, "The file is an osapi file, which a caller cannot write."),
			noStore503,
		),
	endpoint.Delete[DeleteInput, struct{}]("delete").
		Fails("a name over 255 characters", DeleteInput{Name: long}).
		Handle(Delete).
		ID("DeleteFileByName").
		Describe("Delete a file").
		Explain("Delete a file from the Object Store.").
		Returns("File deleted.").
		Status(http.StatusNoContent).
		Errors(
			endpoint.NotFound(ErrNotFound, "File not found."),
			endpoint.Forbidden(ErrProtected, "The file is an osapi file, which a caller cannot delete."),
			noStore503,
		).
		Command("Delete a file from the Object Store", "").
		Example("--name app.conf").
		Example("--name app.conf --json"),
	endpoint.Get[StaleInput, StaleDeploymentsResponse]("stale").
		Handle(Stale).
		Sub("stale").
		Use("list").
		ID("GetFileStale").
		Describe("List stale file deployments").
		Explain("Returns deployments where the source object in the Object Store has been "+
			"updated since the file was last deployed. Controller-only check that compares "+
			"state KV SHA-256 hashes against current object content.").
		Returns("The stale deployments.").
		Command("List stale file deployments",
			"The list shows the deployments whose object changed after the last\n"+
				"deploy. Those are the files that you must deploy again.").
		Example("").
		Prints(sampleStale).
		Example("--json").
		Errors(
			noStore503,
			endpoint.Unavailable(ErrNoState, "The Object Store or the file state KV is not configured."),
		),
))

// long is a name over the 255 characters a file name may have, and content
// is a file's content, for the inputs that Fails declares.
var (
	long    = strings.Repeat("a", 256)
	content = []byte("data")
)
