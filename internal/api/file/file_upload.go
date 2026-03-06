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
	"crypto/sha256"
	"fmt"
	"log/slog"

	"github.com/retr0h/osapi/internal/api/file/gen"
	"github.com/retr0h/osapi/internal/validation"
)

// PostFile upload a file to the Object Store.
func (f *File) PostFile(
	ctx context.Context,
	request gen.PostFileRequestObject,
) (gen.PostFileResponseObject, error) {
	if errMsg, ok := validation.Struct(request.Body); !ok {
		return gen.PostFile400JSONResponse{Error: &errMsg}, nil
	}

	name := request.Body.Name
	content := request.Body.Content

	f.logger.Debug("file upload",
		slog.String("name", name),
		slog.Int("size", len(content)),
	)

	hash := sha256.Sum256(content)
	sha256Hex := fmt.Sprintf("%x", hash)

	_, err := f.objStore.PutBytes(ctx, name, content)
	if err != nil {
		errMsg := fmt.Sprintf("failed to store file: %s", err.Error())
		return gen.PostFile500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	return gen.PostFile201JSONResponse{
		Name:   name,
		Sha256: sha256Hex,
		Size:   len(content),
	}, nil
}
