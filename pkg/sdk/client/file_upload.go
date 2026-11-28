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

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// The two methods of the file group that send a multipart form, PostFile and
// PutFileByName, are written here by hand: the generator writes a method
// only for a JSON body. It writes their examples, which call each one with
// the arguments in the order of FormArgs in internal/sdk/gen/sdkmodel: the
// keys of the path, the required fields of the form, the optional ones, and
// then the file.

// Upload creates a file in the Object Store. It computes the SHA-256 of
// the content first, and skips the upload when the stored file has the
// same content. A file with other content is a ConflictError: Update
// replaces it.
func (s *FileService) Upload(
	ctx context.Context,
	name string,
	contentType types.ContentType,
	file io.Reader,
) (*Response[FileUploadResponse], error) {
	fileData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	sha256Hex := fmt.Sprintf("%x", sha256.Sum256(fileData))
	// The read only finds an upload that it can skip. Any error, such as a
	// 404 for a new file or a 403 for a token that can write but not read,
	// goes on to the upload, which reports its own error.
	if existing, err := s.Get(ctx, name); err == nil && existing.Data.SHA256 == sha256Hex {
		return NewResponse(FileUploadResponse{
			Name:        name,
			SHA256:      sha256Hex,
			Size:        len(fileData),
			Changed:     false,
			ContentType: contentType,
		}, nil), nil
	}

	body, formType := fileForm(name, contentType, fileData)

	return s.sendForm(ctx, http.MethodPost, "/api/file", formType, body)
}

// Update replaces a file that exists in the Object Store. A name that is
// not there is a NotFoundError: Upload creates it.
func (s *FileService) Update(
	ctx context.Context,
	name string,
	contentType types.ContentType,
	file io.Reader,
) (*Response[FileUploadResponse], error) {
	fileData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	body, formType := fileForm("", contentType, fileData)

	return s.sendForm(ctx, http.MethodPut, "/api/file/"+url.PathEscape(name), formType, body)
}

// sendForm sends a multipart form, and decodes the answer of the write.
func (s *FileService) sendForm(
	ctx context.Context,
	method string,
	path string,
	formType string,
	body []byte,
) (*Response[FileUploadResponse], error) {
	resp, err := s.client.Send(ctx, method, path, formType, body)
	if err != nil {
		return nil, err
	}

	return decode[FileUploadResponse](method, path, resp)
}

// fileForm is the multipart form of a file: its name when the body names
// it, its content type and its content.
func fileForm(
	name string,
	contentType types.ContentType,
	data []byte,
) ([]byte, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if name != "" {
		_ = writer.WriteField("name", name)
	}
	_ = writer.WriteField("content_type", string(contentType))
	part, _ := writer.CreateFormFile("file", "upload")
	_, _ = part.Write(data)
	_ = writer.Close()

	return body.Bytes(), writer.FormDataContentType()
}

// FileChanged is whether content differs from the file of the same name in
// the Object Store.
type FileChanged struct {
	// Name is the name of the file.
	Name string `json:"name"`
	// SHA256 is the SHA-256 of the content.
	SHA256 string `json:"sha256"`
	// Changed is true when the content differs from the stored file, or
	// no file has the name.
	Changed bool `json:"changed"`
}

// GetChanged computes the SHA-256 of the content and compares it with the
// SHA-256 of the stored file. Changed is true when the content differs, or
// the file does not exist yet.
func (s *FileService) GetChanged(
	ctx context.Context,
	name string,
	file io.Reader,
) (*Response[FileChanged], error) {
	fileData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	sha256Hex := fmt.Sprintf("%x", sha256.Sum256(fileData))

	existing, err := s.Get(ctx, name)
	if err != nil {
		var notFound *NotFoundError
		if errors.As(err, &notFound) {
			return NewResponse(FileChanged{
				Name:    name,
				Changed: true,
				SHA256:  sha256Hex,
			}, nil), nil
		}

		return nil, fmt.Errorf("check file %s: %w", name, err)
	}

	return NewResponse(FileChanged{
		Name:    name,
		Changed: existing.Data.SHA256 != sha256Hex,
		SHA256:  sha256Hex,
	}, nil), nil
}
