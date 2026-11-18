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
	"fmt"
	"strings"

	"github.com/osapi-io/osapi/internal/controller/api/file/gen"
)

// PutFileByName replaces a file in the Object Store, or creates it when
// the name is new. The same content answers 200 and writes nothing.
func (f *File) PutFileByName(
	ctx context.Context,
	request gen.PutFileByNameRequestObject,
) (gen.PutFileByNameResponseObject, error) {
	form, errMsg := readForm(request.Body, request.Name)
	if errMsg != "" {
		return gen.PutFileByName400JSONResponse{Error: &errMsg}, nil
	}
	if strings.HasPrefix(form.Name, "osapi/") {
		msg := fmt.Sprintf("cannot overwrite protected osapi file: %s", form.Name)
		return gen.PutFileByName403JSONResponse{Error: &msg}, nil
	}

	existing, err := f.objStore.ObjectInfo(ctx, f.objBucket, form.Name)
	existed := err == nil && existing != nil
	if existed && existing.Digest == form.digest() {
		return gen.PutFileByName200JSONResponse(form.result(false)), nil
	}

	if msg := f.store(ctx, form); msg != "" {
		return gen.PutFileByName500JSONResponse{Error: &msg}, nil
	}
	if existed {
		return gen.PutFileByName200JSONResponse(form.result(true)), nil
	}
	location := "/api/file/" + form.Name

	return gen.PutFileByName201JSONResponse{
		Body:    form.result(true),
		Headers: gen.PutFileByName201ResponseHeaders{Location: &location},
	}, nil
}
