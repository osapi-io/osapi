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

// SysctlService provides sysctl management operations.
type SysctlService struct {
	client *gen.ClientWithResponses
}

// SysctlList lists all sysctl parameters on the target host.
func (s *SysctlService) SysctlList(
	ctx context.Context,
	hostname string,
) (*Response[Collection[SysctlEntryResult]], error) {
	resp, err := s.client.GetNodeSysctlWithResponse(ctx, hostname)
	if err != nil {
		return nil, fmt.Errorf("sysctl list: %w", err)
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

	return NewResponse(sysctlEntryCollectionFromGen(resp.JSON200), resp.Body), nil
}

// SysctlGet retrieves a single sysctl parameter by key on the target host.
func (s *SysctlService) SysctlGet(
	ctx context.Context,
	hostname string,
	key string,
) (*Response[Collection[SysctlEntryResult]], error) {
	resp, err := s.client.GetNodeSysctlByKeyWithResponse(ctx, hostname, key)
	if err != nil {
		return nil, fmt.Errorf("sysctl get: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
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

	return NewResponse(sysctlEntryCollectionFromGet(resp.JSON200), resp.Body), nil
}

// SysctlCreate creates a sysctl parameter on the target host.
func (s *SysctlService) SysctlCreate(
	ctx context.Context,
	hostname string,
	opts SysctlCreateOpts,
) (*Response[Collection[SysctlMutationResult]], error) {
	body := gen.SysctlCreateRequest{
		Key:   opts.Key,
		Value: opts.Value,
	}

	resp, err := s.client.PostNodeSysctlWithResponse(ctx, hostname, body)
	if err != nil {
		return nil, fmt.Errorf("sysctl create: %w", err)
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

	return NewResponse(sysctlMutationCollectionFromCreate(resp.JSON200), resp.Body), nil
}

// SysctlUpdate updates an existing sysctl parameter on the target host.
func (s *SysctlService) SysctlUpdate(
	ctx context.Context,
	hostname string,
	key string,
	opts SysctlUpdateOpts,
) (*Response[Collection[SysctlMutationResult]], error) {
	body := gen.SysctlUpdateRequest{
		Value: opts.Value,
	}

	resp, err := s.client.PutNodeSysctlWithResponse(ctx, hostname, key, body)
	if err != nil {
		return nil, fmt.Errorf("sysctl update: %w", err)
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

	return NewResponse(sysctlMutationCollectionFromUpdate(resp.JSON200), resp.Body), nil
}

// SysctlDelete removes a sysctl parameter on the target host.
func (s *SysctlService) SysctlDelete(
	ctx context.Context,
	hostname string,
	key string,
) (*Response[Collection[SysctlMutationResult]], error) {
	resp, err := s.client.DeleteNodeSysctlWithResponse(ctx, hostname, key)
	if err != nil {
		return nil, fmt.Errorf("sysctl delete: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
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

	return NewResponse(sysctlMutationCollectionFromDelete(resp.JSON200), resp.Body), nil
}
