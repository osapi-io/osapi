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

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/osapi-io/osapi/internal/controller/api/file/gen"
	"github.com/osapi-io/osapi/internal/validation"
)

// PostFile creates a file in the Object Store from a multipart form. A name
// that exists with the same content answers 200 and changes nothing; with
// other content it is a 409, and PUT /api/file/{name} replaces it.
func (f *File) PostFile(
	ctx context.Context,
	request gen.PostFileRequestObject,
) (gen.PostFileResponseObject, error) {
	form, errMsg := readForm(request.Body, "")
	if errMsg != "" {
		return gen.PostFile400JSONResponse{Error: &errMsg}, nil
	}
	if strings.HasPrefix(form.Name, "osapi/") {
		msg := fmt.Sprintf("cannot overwrite protected osapi file: %s", form.Name)
		return gen.PostFile403JSONResponse{Error: &msg}, nil
	}

	existing, err := f.objStore.ObjectInfo(ctx, f.objBucket, form.Name)
	if err == nil && existing != nil {
		if existing.Digest == form.digest() {
			return gen.PostFile200JSONResponse(form.result(false)), nil
		}
		msg := fmt.Sprintf(
			"file %s already exists with different content; replace it with PUT /api/file/%s",
			form.Name, form.Name,
		)
		return gen.PostFile409JSONResponse{Error: &msg}, nil
	}

	if msg := f.store(ctx, form); msg != "" {
		return gen.PostFile500JSONResponse{Error: &msg}, nil
	}
	location := "/api/file/" + form.Name

	return gen.PostFile201JSONResponse{
		Body:    form.result(true),
		Headers: gen.PostFile201ResponseHeaders{Location: &location},
	}, nil
}

// store writes the form's file to the Object Store, and says why it could
// not.
func (f *File) store(
	ctx context.Context,
	form uploadForm,
) string {
	f.logger.Debug(
		"file upload",
		slog.String("name", form.Name),
		slog.String("content_type", form.ContentType),
		slog.Int("size", len(form.File)),
	)

	meta := jetstream.ObjectMeta{
		Name: form.Name,
		Headers: nats.Header{
			"Osapi-Content-Type": []string{form.ContentType},
		},
	}
	if _, err := f.objStore.ObjectPut(ctx, f.objBucket, meta, bytes.NewReader(form.File)); err != nil {
		return fmt.Sprintf("failed to store file: %s", err.Error())
	}

	return ""
}

// readForm reads the parts of a multipart form, and validates them. name
// is the name from the path of a PUT; a POST gives "" and reads the name
// part.
func readForm(
	r *multipart.Reader,
	name string,
) (uploadForm, string) {
	form := uploadForm{Name: name}
	for {
		part, err := r.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return form, fmt.Sprintf("failed to read multipart: %s", err.Error())
		}

		switch part.FormName() {
		case "name":
			if name == "" {
				b, _ := io.ReadAll(part)
				form.Name = string(b)
			}
		case "content_type":
			b, _ := io.ReadAll(part)
			form.ContentType = string(b)
		case "file":
			form.File, _ = io.ReadAll(part)
		}
		_ = part.Close()
	}

	if form.ContentType == "" {
		form.ContentType = "raw"
	}
	if errMsg, ok := validation.Struct(form); !ok {
		return form, errMsg
	}

	return form, ""
}

// digest is the Object Store digest of the form's file.
func (u uploadForm) digest() string {
	hash := sha256.Sum256(u.File)

	return "SHA-256=" + base64.URLEncoding.EncodeToString(hash[:])
}

// result is the answer of a write of the form's file.
func (u uploadForm) result(
	changed bool,
) gen.FileUploadResponse {
	return gen.FileUploadResponse{
		Name:        u.Name,
		Sha256:      fmt.Sprintf("%x", sha256.Sum256(u.File)),
		Size:        len(u.File),
		Changed:     changed,
		ContentType: u.ContentType,
	}
}

// uploadForm carries the multipart form fields so they can be validated
// against the rules the spec declares for them.
//
// oapi-codegen does not bind multipart/form-data into a struct, so the parts
// are read by hand. Without this the rules had to be rewritten as if
// statements, which left the validate tags in gen/api.yaml as documentation
// nothing enforced, and produced 400s phrased unlike every other 400 in the
// API. The tags here are the ones that spec declares, so changing one place
// without the other now fails a test.
type uploadForm struct {
	Name        string `validate:"required,min=1,max=255"`
	ContentType string `validate:"omitempty,oneof=raw template"`
	File        []byte `validate:"required,min=1"`
}
