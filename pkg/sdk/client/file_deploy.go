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
	"context"
	"fmt"

	"github.com/retr0h/osapi/pkg/sdk/client/gen"
)

// FileDeployService provides file deployment operations on target hosts.
type FileDeployService struct {
	client *gen.ClientWithResponses
}

// Deploy deploys a file from the Object Store to the target host.
func (s *FileDeployService) Deploy(
	ctx context.Context,
	req FileDeployOpts,
) (*Response[Collection[FileDeployResult]], error) {
	body := gen.FileDeployRequest{
		ObjectName:  req.ObjectName,
		Path:        req.Path,
		ContentType: gen.FileDeployRequestContentType(req.ContentType),
	}

	if req.Mode != "" {
		body.Mode = &req.Mode
	}

	if req.Owner != "" {
		body.Owner = &req.Owner
	}

	if req.Group != "" {
		body.Group = &req.Group
	}

	if len(req.Vars) > 0 {
		body.Vars = &req.Vars
	}

	resp, err := s.client.PostNodeFileDeployWithResponse(ctx, req.Target, body)
	if err != nil {
		return nil, fmt.Errorf("file deploy: %w", err)
	}

	if err := checkError(resp.StatusCode(), resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON500); err != nil {
		return nil, err
	}

	if resp.JSON202 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(fileDeployCollectionFromGen(resp.JSON202), resp.Body), nil
}

// Undeploy removes a deployed file from the target host filesystem.
func (s *FileDeployService) Undeploy(
	ctx context.Context,
	req FileUndeployOpts,
) (*Response[Collection[FileUndeployResult]], error) {
	body := gen.FileUndeployRequest{
		Path: req.Path,
	}

	resp, err := s.client.PostNodeFileUndeployWithResponse(ctx, req.Target, body)
	if err != nil {
		return nil, fmt.Errorf("file undeploy: %w", err)
	}

	if err := checkError(resp.StatusCode(), resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON500); err != nil {
		return nil, err
	}

	if resp.JSON202 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(fileUndeployCollectionFromGen(resp.JSON202), resp.Body), nil
}

// Status checks the deployment status of a file on the target host.
func (s *FileDeployService) Status(
	ctx context.Context,
	target string,
	path string,
) (*Response[Collection[FileStatusResult]], error) {
	body := gen.FileStatusRequest{
		Path: path,
	}

	resp, err := s.client.PostNodeFileStatusWithResponse(ctx, target, body)
	if err != nil {
		return nil, fmt.Errorf("file status: %w", err)
	}

	if err := checkError(resp.StatusCode(), resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON500); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(fileStatusCollectionFromGen(resp.JSON200), resp.Body), nil
}
