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

// PackageService provides package management operations.
type PackageService struct {
	client *gen.ClientWithResponses
}

// List lists all installed packages on the target host.
func (s *PackageService) List(
	ctx context.Context,
	hostname string,
) (*Response[Collection[PackageInfoResult]], error) {
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

	return NewResponse(packageInfoCollectionFromList(resp.JSON200), resp.Body), nil
}

// Get retrieves a single package by name on the target host.
func (s *PackageService) Get(
	ctx context.Context,
	hostname string,
	name string,
) (*Response[Collection[PackageInfoResult]], error) {
	resp, err := s.client.GetNodePackageByNameWithResponse(ctx, hostname, name)
	if err != nil {
		return nil, fmt.Errorf("package get: %w", err)
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

	return NewResponse(packageInfoCollectionFromGet(resp.JSON200), resp.Body), nil
}

// Install installs a package on the target host.
func (s *PackageService) Install(
	ctx context.Context,
	hostname string,
	name string,
) (*Response[Collection[PackageMutationResult]], error) {
	body := gen.PackageInstallRequest{
		Name: name,
	}

	resp, err := s.client.PostNodePackageWithResponse(ctx, hostname, body)
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

	return NewResponse(packageMutationCollectionFromInstall(resp.JSON200), resp.Body), nil
}

// Remove removes a package from the target host.
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

	return NewResponse(packageMutationCollectionFromRemove(resp.JSON200), resp.Body), nil
}

// Update updates all packages on the target host.
func (s *PackageService) Update(
	ctx context.Context,
	hostname string,
) (*Response[Collection[PackageMutationResult]], error) {
	resp, err := s.client.PostNodePackageUpdateWithResponse(ctx, hostname)
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

	return NewResponse(packageMutationCollectionFromUpdate(resp.JSON200), resp.Body), nil
}

// ListUpdates lists available package updates on the target host.
func (s *PackageService) ListUpdates(
	ctx context.Context,
	hostname string,
) (*Response[Collection[PackageUpdateResult]], error) {
	resp, err := s.client.GetNodePackageUpdateWithResponse(ctx, hostname)
	if err != nil {
		return nil, fmt.Errorf("package list updates: %w", err)
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

	return NewResponse(packageUpdateCollectionFromGen(resp.JSON200), resp.Body), nil
}
