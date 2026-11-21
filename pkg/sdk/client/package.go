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

	"github.com/osapi-io/osapi/pkg/sdk/client/gen"
)

// PackageService reads, installs and removes the packages of a host, and
// reads and refreshes its package index.
type PackageService struct {
	client *gen.ClientWithResponses
}

// List lists the packages installed on the target.
func (s *PackageService) List(
	ctx context.Context,
	hostname string,
) (*Response[Collection[PackageInfo]], error) {
	resp, err := s.client.GetNodePackageWithResponse(ctx, hostname)
	if err != nil {
		return nil, fmt.Errorf("package list: %w", err)
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

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return newResponse(
		packageInfosFromGen(resp.JSON200.JobId, resp.JSON200.Results),
		resp.Body,
		resp.HTTPResponse,
	), nil
}

// Get gets one installed package from the target.
func (s *PackageService) Get(
	ctx context.Context,
	hostname string,
	name string,
) (*Response[Collection[PackageInfo]], error) {
	resp, err := s.client.GetNodePackageByKeyWithResponse(ctx, hostname, name)
	if err != nil {
		return nil, fmt.Errorf("package get: %w", err)
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

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return newResponse(
		packageInfosFromGen(resp.JSON200.JobId, resp.JSON200.Results),
		resp.Body,
		resp.HTTPResponse,
	), nil
}

// Install installs a package on the target, at version when it is not
// empty. Idempotent.
func (s *PackageService) Install(
	ctx context.Context,
	hostname string,
	name string,
	version string,
) (*Response[Collection[PackageMutationResult]], error) {
	body := gen.PackageCreateRequest{Name: name, Version: version}
	resp, err := s.client.PostNodePackageWithResponse(
		ctx,
		hostname,
		body,
	)
	if err != nil {
		return nil, fmt.Errorf("package install: %w", err)
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

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return newResponse(
		packageMutationsFromGen(resp.JSON200.JobId, resp.JSON200.Results),
		resp.Body,
		resp.HTTPResponse,
	), nil
}

// Remove removes a package from the target. Idempotent.
func (s *PackageService) Remove(
	ctx context.Context,
	hostname string,
	name string,
) (*Response[Collection[PackageMutationResult]], error) {
	resp, err := s.client.DeleteNodePackageWithResponse(ctx, hostname, name)
	if err != nil {
		return nil, fmt.Errorf("package remove: %w", err)
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

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return newResponse(
		packageMutationsFromGen(resp.JSON200.JobId, resp.JSON200.Results),
		resp.Body,
		resp.HTTPResponse,
	), nil
}

// ListUpdates lists the packages the target's index has newer versions of.
func (s *PackageService) ListUpdates(
	ctx context.Context,
	hostname string,
) (*Response[Collection[PackageUpdate]], error) {
	resp, err := s.client.GetNodeUpdatepackagesWithResponse(ctx, hostname)
	if err != nil {
		return nil, fmt.Errorf("package listupdates: %w", err)
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

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return newResponse(packageUpdatesFromGen(resp.JSON200), resp.Body, resp.HTTPResponse), nil
}

// Update refreshes the target's package index.
func (s *PackageService) Update(
	ctx context.Context,
	hostname string,
) (*Response[Collection[PackageIndexResult]], error) {
	resp, err := s.client.PutNodeUpdatepackagesWithResponse(ctx, hostname)
	if err != nil {
		return nil, fmt.Errorf("package update: %w", err)
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

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return newResponse(packageIndexFromGen(resp.JSON200), resp.Body, resp.HTTPResponse), nil
}
