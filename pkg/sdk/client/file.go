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

	"github.com/osapi-io/osapi/pkg/sdk/client/gen"
)

// FileService provides file management operations for the Object Store.
type FileService struct {
	client *gen.ClientWithResponses
}

// Upload creates a file in the Object Store. It computes the SHA-256 of
// the content first, and skips the upload when the stored file has the
// same content. A file with other content is a ConflictError: Update
// replaces it.
func (s *FileService) Upload(
	ctx context.Context,
	name string,
	contentType string,
	file io.Reader,
) (*Response[FileUpload], error) {
	fileData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	sha256Hex := fmt.Sprintf("%x", sha256.Sum256(fileData))
	if existing, err := s.Get(ctx, name); err == nil && existing.Data.SHA256 == sha256Hex {
		return NewResponse(FileUpload{
			Name:        name,
			SHA256:      sha256Hex,
			Size:        len(fileData),
			Changed:     false,
			ContentType: contentType,
		}, nil), nil
	}

	body, formType := fileForm(name, contentType, fileData)
	resp, err := s.client.PostFileWithBodyWithResponse(ctx, formType, body)
	if err != nil {
		return nil, fmt.Errorf("upload file: %w", err)
	}
	if err := checkError(
		resp.StatusCode(),
		resp.JSON400,
		resp.JSON401,
		resp.JSON403,
		resp.JSON409,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	return written(resp.JSON201, resp.JSON200, resp.StatusCode(), resp.Body, resp.HTTPResponse)
}

// Update replaces a file in the Object Store, or creates it when the name
// is new.
func (s *FileService) Update(
	ctx context.Context,
	name string,
	contentType string,
	file io.Reader,
) (*Response[FileUpload], error) {
	fileData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	body, formType := fileForm("", contentType, fileData)
	resp, err := s.client.PutFileByNameWithBodyWithResponse(ctx, name, formType, body)
	if err != nil {
		return nil, fmt.Errorf("update file: %w", err)
	}
	if err := checkError(
		resp.StatusCode(),
		resp.JSON400,
		resp.JSON401,
		resp.JSON403,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	return written(resp.JSON201, resp.JSON200, resp.StatusCode(), resp.Body, resp.HTTPResponse)
}

// fileForm is the multipart form of a file: its name when the body names
// it, its content type and its content.
func fileForm(
	name string,
	contentType string,
	data []byte,
) (*bytes.Buffer, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if name != "" {
		_ = writer.WriteField("name", name)
	}
	_ = writer.WriteField("content_type", contentType)
	part, _ := writer.CreateFormFile("file", "upload")
	_, _ = part.Write(data)
	_ = writer.Close()

	return body, writer.FormDataContentType()
}

// written is the answer of a write: a new file answers 201, and one that
// existed 200.
func written(
	created *gen.FileUploadResponse,
	existed *gen.FileUploadResponse,
	code int,
	body []byte,
	httpResp *http.Response,
) (*Response[FileUpload], error) {
	result := created
	if result == nil {
		result = existed
	}
	if result == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: code,
			Message:    "nil response body",
		}}
	}

	return newResponse(fileUploadFromGen(result), body, httpResp), nil
}

// List retrieves all files stored in the Object Store.
func (s *FileService) List(
	ctx context.Context,
) (*Response[FileList], error) {
	resp, err := s.client.GetFilesWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("list files: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON401,
		resp.JSON403,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return newResponse(fileListFromGen(resp.JSON200), resp.Body, resp.HTTPResponse), nil
}

// Get retrieves metadata for a specific file in the Object Store.
func (s *FileService) Get(
	ctx context.Context,
	name string,
) (*Response[FileMetadata], error) {
	resp, err := s.client.GetFileByNameWithResponse(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("get file %s: %w", name, err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON400,
		resp.JSON401,
		resp.JSON403,
		resp.JSON404,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return newResponse(fileMetadataFromGen(resp.JSON200), resp.Body, resp.HTTPResponse), nil
}

// Delete removes a file from the Object Store.
func (s *FileService) Delete(
	ctx context.Context,
	name string,
) error {
	resp, err := s.client.DeleteFileByNameWithResponse(ctx, name)
	if err != nil {
		return fmt.Errorf("delete file %s: %w", name, err)
	}

	return checkError(
		resp.StatusCode(),
		resp.JSON400,
		resp.JSON401,
		resp.JSON403,
		resp.JSON404,
		resp.JSON500,
	)
}

// Stale lists deployments where the deployed file content no longer
// matches the current object store content.
func (s *FileService) ListStale(
	ctx context.Context,
) (*Response[StaleList], error) {
	resp, err := s.client.GetFileStaleWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("list stale deployments: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON401,
		resp.JSON403,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return newResponse(staleListFromGen(resp.JSON200), resp.Body, resp.HTTPResponse), nil
}

// Changed computes the SHA-256 of the provided content and compares
// it against the stored hash in the Object Store. Returns true if
// the content differs or the file does not exist yet.
func (s *FileService) GetChanged(
	ctx context.Context,
	name string,
	file io.Reader,
) (*Response[FileChanged], error) {
	fileData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	hash := sha256.Sum256(fileData)
	sha256Hex := fmt.Sprintf("%x", hash)

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

	changed := existing.Data.SHA256 != sha256Hex

	return NewResponse(FileChanged{
		Name:    name,
		Changed: changed,
		SHA256:  sha256Hex,
	}, nil), nil
}
