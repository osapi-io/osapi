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
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/osapi-io/osapi/internal/controller/api/file/gen"
	"github.com/osapi-io/osapi/internal/validation"
)

// PostFile upload a file to the Object Store via multipart/form-data.
func (f *File) PostFile(
	ctx context.Context,
	request gen.PostFileRequestObject,
) (gen.PostFileResponseObject, error) {
	// Defense in depth: current params use omitempty so validation
	// always passes, but guards against future parameter additions.
	if errMsg, ok := validation.Struct(request.Params); !ok {
		return gen.PostFile400JSONResponse{Error: &errMsg}, nil
	}

	name, contentType, fileData, errResp := f.parseMultipart(request)
	if errResp != nil {
		return errResp, nil
	}

	if strings.HasPrefix(name, "osapi/") {
		errMsg := fmt.Sprintf("cannot overwrite protected osapi file: %s", name)
		return gen.PostFile403JSONResponse{Error: &errMsg}, nil
	}

	f.logger.Debug(
		"file upload",
		slog.String("name", name),
		slog.String("content_type", contentType),
		slog.Int("size", len(fileData)),
	)

	hash := sha256.Sum256(fileData)
	sha256Hex := fmt.Sprintf("%x", hash)
	newDigest := "SHA-256=" + base64.URLEncoding.EncodeToString(hash[:])

	force := request.Params.Force != nil && *request.Params.Force

	// Unless forced, check if the Object Store already has this file.
	if !force {
		existing, err := f.objStore.GetInfo(ctx, name)
		if err == nil && existing != nil {
			if existing.Digest == newDigest {
				// Same content — skip the write.
				return gen.PostFile201JSONResponse{
					Name:        name,
					Sha256:      sha256Hex,
					Size:        len(fileData),
					Changed:     false,
					ContentType: contentType,
				}, nil
			}

			// Different content — reject without force.
			errMsg := fmt.Sprintf(
				"file %s already exists with different content; use force to overwrite",
				name,
			)
			return gen.PostFile409JSONResponse{Error: &errMsg}, nil
		}
	}

	meta := jetstream.ObjectMeta{
		Name: name,
		Headers: nats.Header{
			"Osapi-Content-Type": []string{contentType},
		},
	}
	_, err := f.objStore.Put(ctx, meta, bytes.NewReader(fileData))
	if err != nil {
		errMsg := fmt.Sprintf("failed to store file: %s", err.Error())
		return gen.PostFile500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	return gen.PostFile201JSONResponse{
		Name:        name,
		Sha256:      sha256Hex,
		Size:        len(fileData),
		Changed:     true,
		ContentType: contentType,
	}, nil
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

// parseMultipart reads multipart parts and extracts name, content_type,
// and file data. Returns a 400 response on validation failure.
func (f *File) parseMultipart(
	request gen.PostFileRequestObject,
) (string, string, []byte, gen.PostFileResponseObject) {
	var name string
	var contentType string
	var fileData []byte

	for {
		part, err := request.Body.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			errMsg := fmt.Sprintf("failed to read multipart: %s", err.Error())
			return "", "", nil, gen.PostFile400JSONResponse{Error: &errMsg}
		}

		switch part.FormName() {
		case "name":
			b, _ := io.ReadAll(part)
			name = string(b)
		case "content_type":
			b, _ := io.ReadAll(part)
			contentType = string(b)
		case "file":
			fileData, _ = io.ReadAll(part)
		}
		_ = part.Close()
	}

	if contentType == "" {
		contentType = "raw"
	}

	if errMsg, ok := validation.Struct(uploadForm{
		Name:        name,
		ContentType: contentType,
		File:        fileData,
	}); !ok {
		return "", "", nil, gen.PostFile400JSONResponse{Error: &errMsg}
	}

	return name, contentType, fileData, nil
}
