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
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/retr0h/osapi/internal/controller/api/file/gen"
)

// GetFiles list all files stored in the Object Store.
func (f *File) GetFiles(
	ctx context.Context,
	_ gen.GetFilesRequestObject,
) (gen.GetFilesResponseObject, error) {
	f.logger.Debug("file list")

	objects, err := f.objStore.List(ctx)
	if err != nil {
		if errors.Is(err, jetstream.ErrNoObjectsFound) {
			return gen.GetFiles200JSONResponse{
				Files: []gen.FileInfo{},
				Total: 0,
			}, nil
		}

		errMsg := fmt.Sprintf("failed to list files: %s", err.Error())
		return gen.GetFiles500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	files := make([]gen.FileInfo, 0, len(objects))
	for _, obj := range objects {
		if obj.Deleted {
			continue
		}

		digestB64 := strings.TrimPrefix(obj.Digest, "SHA-256=")
		sha256Hex := digestB64
		if digestBytes, err := base64.URLEncoding.DecodeString(digestB64); err == nil {
			sha256Hex = fmt.Sprintf("%x", digestBytes)
		}

		contentType := ""
		if obj.Headers != nil {
			contentType = obj.Headers.Get("Osapi-Content-Type")
		}

		source := "user"
		if strings.HasPrefix(obj.Name, "system/") {
			source = "system"
		}

		files = append(files, gen.FileInfo{
			Name:        obj.Name,
			Sha256:      sha256Hex,
			Size:        int(obj.Size),
			ContentType: contentType,
			Source:      source,
		})
	}

	return gen.GetFiles200JSONResponse{
		Files: files,
		Total: len(files),
	}, nil
}
