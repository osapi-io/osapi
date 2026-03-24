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
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/retr0h/osapi/internal/controller/api/file/gen"
)

// DeleteFileByName delete a file from the Object Store.
func (f *File) DeleteFileByName(
	ctx context.Context,
	request gen.DeleteFileByNameRequestObject,
) (gen.DeleteFileByNameResponseObject, error) {
	if errMsg, ok := validateFileName(request.Name); !ok {
		return gen.DeleteFileByName400JSONResponse{Error: &errMsg}, nil
	}

	if strings.HasPrefix(request.Name, "osapi/") {
		errMsg := fmt.Sprintf("cannot delete protected osapi file: %s", request.Name)
		return gen.DeleteFileByName403JSONResponse{Error: &errMsg}, nil
	}

	f.logger.Debug("file delete",
		slog.String("name", request.Name),
	)

	// Check if the file exists before attempting deletion.
	_, err := f.objStore.GetInfo(ctx, request.Name)
	if err != nil {
		if errors.Is(err, jetstream.ErrObjectNotFound) {
			errMsg := fmt.Sprintf("file not found: %s", request.Name)
			return gen.DeleteFileByName404JSONResponse{
				Error: &errMsg,
			}, nil
		}

		errMsg := fmt.Sprintf("failed to get file info: %s", err.Error())
		return gen.DeleteFileByName500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	if err := f.objStore.Delete(ctx, request.Name); err != nil {
		errMsg := fmt.Sprintf("failed to delete file: %s", err.Error())
		return gen.DeleteFileByName500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	return gen.DeleteFileByName200JSONResponse{
		Name:    request.Name,
		Deleted: true,
	}, nil
}
