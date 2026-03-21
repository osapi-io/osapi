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
	"log/slog"
	"strings"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/retr0h/osapi/internal/controller/api/file/gen"
)

// GetFileByName get metadata for a specific file in the Object Store.
func (f *File) GetFileByName(
	ctx context.Context,
	request gen.GetFileByNameRequestObject,
) (gen.GetFileByNameResponseObject, error) {
	if errMsg, ok := validateFileName(request.Name); !ok {
		return gen.GetFileByName400JSONResponse{Error: &errMsg}, nil
	}

	f.logger.Debug("file get",
		slog.String("name", request.Name),
	)

	info, err := f.objStore.GetInfo(ctx, request.Name)
	if err != nil {
		if errors.Is(err, jetstream.ErrObjectNotFound) {
			errMsg := fmt.Sprintf("file not found: %s", request.Name)
			return gen.GetFileByName404JSONResponse{
				Error: &errMsg,
			}, nil
		}

		errMsg := fmt.Sprintf("failed to get file info: %s", err.Error())
		return gen.GetFileByName500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	digestB64 := strings.TrimPrefix(info.Digest, "SHA-256=")
	sha256Hex := digestB64
	if digestBytes, err := base64.URLEncoding.DecodeString(digestB64); err == nil {
		sha256Hex = fmt.Sprintf("%x", digestBytes)
	}

	contentType := ""
	if info.Headers != nil {
		contentType = info.Headers.Get("Osapi-Content-Type")
	}

	return gen.GetFileByName200JSONResponse{
		Name:        info.Name,
		Sha256:      sha256Hex,
		Size:        int(info.Size),
		ContentType: contentType,
	}, nil
}
