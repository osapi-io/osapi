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

package osapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/retr0h/osapi/pkg/sdk/osapi/gen"
)

// UploadOption configures Upload behavior.
type UploadOption func(*uploadOptions)

type uploadOptions struct {
	force bool
}

// WithForce bypasses both SDK-side pre-check and server-side digest
// check. The file is always uploaded and changed is always true.
func WithForce() UploadOption {
	return func(o *uploadOptions) { o.force = true }
}

// FileService provides file management operations for the Object Store.
type FileService struct {
	client *gen.ClientWithResponses
}

// Upload uploads a file to the Object Store via multipart/form-data.
// By default, it computes SHA-256 locally and compares against the
// stored hash to skip the upload when content is unchanged. Use
// WithForce to bypass this check.
func (s *FileService) Upload(
	ctx context.Context,
	name string,
	contentType string,
	file io.Reader,
	opts ...UploadOption,
) (*Response[FileUpload], error) {
	var options uploadOptions
	for _, o := range opts {
		o(&options)
	}

	// Buffer file content for hashing and multipart construction.
	fileData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	// Compute SHA-256 locally.
	hash := sha256.Sum256(fileData)
	sha256Hex := fmt.Sprintf("%x", hash)

	// SDK-side pre-check: skip upload if content unchanged.
	// Skipped when force is set.
	if !options.force {
		existing, err := s.Get(ctx, name)
		if err == nil && existing.Data.SHA256 == sha256Hex {
			return NewResponse(FileUpload{
				Name:        name,
				SHA256:      sha256Hex,
				Size:        len(fileData),
				Changed:     false,
				ContentType: contentType,
			}, nil), nil
		}
		// On error (404, network, etc.) fall through to upload.
	}

	// Build multipart body from buffered content.
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("name", name)
	_ = writer.WriteField("content_type", contentType)

	part, _ := writer.CreateFormFile("file", name)
	_, _ = part.Write(fileData)
	_ = writer.Close()

	// Pass force as query param.
	params := &gen.PostFileParams{}
	if options.force {
		params.Force = &options.force
	}

	resp, err := s.client.PostFileWithBodyWithResponse(
		ctx,
		params,
		writer.FormDataContentType(),
		body,
	)
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

	if resp.JSON201 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(fileUploadFromGen(resp.JSON201), resp.Body), nil
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

	return NewResponse(fileListFromGen(resp.JSON200), resp.Body), nil
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

	return NewResponse(fileMetadataFromGen(resp.JSON200), resp.Body), nil
}

// Delete removes a file from the Object Store.
func (s *FileService) Delete(
	ctx context.Context,
	name string,
) (*Response[FileDelete], error) {
	resp, err := s.client.DeleteFileByNameWithResponse(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("delete file %s: %w", name, err)
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

	return NewResponse(fileDeleteFromGen(resp.JSON200), resp.Body), nil
}

// Changed computes the SHA-256 of the provided content and compares
// it against the stored hash in the Object Store. Returns true if
// the content differs or the file does not exist yet.
func (s *FileService) Changed(
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
