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

// CertificateService provides CA certificate management operations.
type CertificateService struct {
	client *gen.ClientWithResponses
}

// List lists all CA certificates on the target host.
func (s *CertificateService) List(
	ctx context.Context,
	hostname string,
) (*Response[Collection[CertificateCAResult]], error) {
	resp, err := s.client.GetNodeCertificateCaWithResponse(ctx, hostname)
	if err != nil {
		return nil, fmt.Errorf("certificate list: %w", err)
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

	return NewResponse(certificateCACollectionFromGen(resp.JSON200), resp.Body), nil
}

// Create creates a new CA certificate on the target host.
func (s *CertificateService) Create(
	ctx context.Context,
	hostname string,
	opts CertificateCreateOpts,
) (*Response[Collection[CertificateCAMutationResult]], error) {
	body := gen.CertificateCACreateRequest{
		Name:   opts.Name,
		Object: opts.Object,
	}

	resp, err := s.client.PostNodeCertificateCaWithResponse(ctx, hostname, body)
	if err != nil {
		return nil, fmt.Errorf("certificate create: %w", err)
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

	return NewResponse(certificateCAMutationCollectionFromGen(resp.JSON200), resp.Body), nil
}

// Update updates an existing CA certificate on the target host.
func (s *CertificateService) Update(
	ctx context.Context,
	hostname string,
	name string,
	opts CertificateUpdateOpts,
) (*Response[Collection[CertificateCAMutationResult]], error) {
	body := gen.CertificateCAUpdateRequest{
		Object: opts.Object,
	}

	resp, err := s.client.PutNodeCertificateCaWithResponse(ctx, hostname, name, body)
	if err != nil {
		return nil, fmt.Errorf("certificate update: %w", err)
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

	return NewResponse(certificateCAMutationCollectionFromGen(resp.JSON200), resp.Body), nil
}

// Delete deletes a CA certificate on the target host.
func (s *CertificateService) Delete(
	ctx context.Context,
	hostname string,
	name string,
) (*Response[Collection[CertificateCAMutationResult]], error) {
	resp, err := s.client.DeleteNodeCertificateCaWithResponse(ctx, hostname, name)
	if err != nil {
		return nil, fmt.Errorf("certificate delete: %w", err)
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

	return NewResponse(certificateCAMutationCollectionFromGen(resp.JSON200), resp.Body), nil
}
