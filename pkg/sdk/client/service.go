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
	"net/http"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/osapi-io/osapi/pkg/sdk/client/gen"
)

// ServiceService reads the systemd services of a host, and deploys, removes,
// starts, stops, restarts, enables and disables the services that osapi
// deploys. The unit file of each is an object of the object store, which
// File.Upload stores.
type ServiceService struct {
	client *gen.ClientWithResponses
}

// List lists the services of the target, one row per service per host.
func (s *ServiceService) List(
	ctx context.Context,
	hostname string,
) (*Response[Collection[ServiceInfo]], error) {
	resp, err := s.client.GetNodeServiceWithResponse(ctx, hostname)
	if err != nil {
		return nil, fmt.Errorf("service list: %w", err)
	}
	if resp.JSON200 == nil {
		return nil, failed(
			resp.StatusCode(),
			resp.JSON400,
			resp.JSON401,
			resp.JSON403,
			resp.JSON500,
		)
	}

	return newResponse(
		serviceInfosFromGen(resp.JSON200.JobId, resp.JSON200.Results),
		resp.Body,
		resp.HTTPResponse,
	), nil
}

// Get gets one service of the target, by its name with .service or without
// it.
func (s *ServiceService) Get(
	ctx context.Context,
	hostname string,
	name string,
) (*Response[Collection[ServiceInfo]], error) {
	resp, err := s.client.GetNodeServiceByKeyWithResponse(ctx, hostname, name)
	if err != nil {
		return nil, fmt.Errorf("service get: %w", err)
	}
	if resp.JSON200 == nil {
		return nil, failed(resp.StatusCode(),
			resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON404, resp.JSON500)
	}

	return newResponse(
		serviceInfosFromGen(resp.JSON200.JobId, resp.JSON200.Results),
		resp.Body,
		resp.HTTPResponse,
	), nil
}

// Create deploys the object as the unit file of the service name on the
// target, and reloads systemd. Idempotent.
func (s *ServiceService) Create(
	ctx context.Context,
	hostname string,
	opts ServiceCreateOpts,
) (*Response[Collection[ServiceMutationResult]], error) {
	resp, err := s.client.PostNodeServiceWithResponse(ctx, hostname,
		gen.ServiceCreateRequest{Name: opts.Name, Object: opts.Object})
	if err != nil {
		return nil, fmt.Errorf("service create: %w", err)
	}
	if resp.JSON200 == nil {
		return nil, failed(
			resp.StatusCode(),
			resp.JSON400,
			resp.JSON401,
			resp.JSON403,
			resp.JSON500,
		)
	}

	return serviceMutation(
		resp.JSON200.JobId,
		resp.JSON200.Results,
		resp.Body,
		resp.HTTPResponse,
	), nil
}

// Update deploys the object as the unit file of a service that osapi
// deployed on the target. Idempotent.
func (s *ServiceService) Update(
	ctx context.Context,
	hostname string,
	name string,
	opts ServiceUpdateOpts,
) (*Response[Collection[ServiceMutationResult]], error) {
	resp, err := s.client.PutNodeServiceWithResponse(ctx, hostname, name,
		gen.ServiceUpdateRequest{Object: opts.Object})
	if err != nil {
		return nil, fmt.Errorf("service update: %w", err)
	}
	if resp.JSON200 == nil {
		return nil, failed(resp.StatusCode(),
			resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON404, resp.JSON500)
	}

	return serviceMutation(
		resp.JSON200.JobId,
		resp.JSON200.Results,
		resp.Body,
		resp.HTTPResponse,
	), nil
}

// Delete stops, disables and removes a service that osapi deployed on the
// target. Idempotent.
func (s *ServiceService) Delete(
	ctx context.Context,
	hostname string,
	name string,
) (*Response[Collection[ServiceMutationResult]], error) {
	resp, err := s.client.DeleteNodeServiceWithResponse(ctx, hostname, name)
	if err != nil {
		return nil, fmt.Errorf("service delete: %w", err)
	}
	if resp.JSON200 == nil {
		return nil, failed(resp.StatusCode(),
			resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON404, resp.JSON500)
	}

	return serviceMutation(
		resp.JSON200.JobId,
		resp.JSON200.Results,
		resp.Body,
		resp.HTTPResponse,
	), nil
}

// Start starts a service that osapi deployed on the target. Idempotent.
func (s *ServiceService) Start(
	ctx context.Context,
	hostname string,
	name string,
) (*Response[Collection[ServiceMutationResult]], error) {
	resp, err := s.client.PostNodeServiceStartWithResponse(ctx, hostname, name)
	if err != nil {
		return nil, fmt.Errorf("service start: %w", err)
	}
	if resp.JSON200 == nil {
		return nil, failed(resp.StatusCode(),
			resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON404, resp.JSON500)
	}

	return serviceMutation(
		resp.JSON200.JobId,
		resp.JSON200.Results,
		resp.Body,
		resp.HTTPResponse,
	), nil
}

// Stop stops a service that osapi deployed on the target. Idempotent.
func (s *ServiceService) Stop(
	ctx context.Context,
	hostname string,
	name string,
) (*Response[Collection[ServiceMutationResult]], error) {
	resp, err := s.client.PostNodeServiceStopWithResponse(ctx, hostname, name)
	if err != nil {
		return nil, fmt.Errorf("service stop: %w", err)
	}
	if resp.JSON200 == nil {
		return nil, failed(resp.StatusCode(),
			resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON404, resp.JSON500)
	}

	return serviceMutation(
		resp.JSON200.JobId,
		resp.JSON200.Results,
		resp.Body,
		resp.HTTPResponse,
	), nil
}

// Restart restarts a service that osapi deployed on the target. It changes
// the host each time.
func (s *ServiceService) Restart(
	ctx context.Context,
	hostname string,
	name string,
) (*Response[Collection[ServiceMutationResult]], error) {
	resp, err := s.client.PostNodeServiceRestartWithResponse(ctx, hostname, name)
	if err != nil {
		return nil, fmt.Errorf("service restart: %w", err)
	}
	if resp.JSON200 == nil {
		return nil, failed(resp.StatusCode(),
			resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON404, resp.JSON500)
	}

	return serviceMutation(
		resp.JSON200.JobId,
		resp.JSON200.Results,
		resp.Body,
		resp.HTTPResponse,
	), nil
}

// Enable makes a service that osapi deployed on the target start at boot.
// Idempotent.
func (s *ServiceService) Enable(
	ctx context.Context,
	hostname string,
	name string,
) (*Response[Collection[ServiceMutationResult]], error) {
	resp, err := s.client.PostNodeServiceEnableWithResponse(ctx, hostname, name)
	if err != nil {
		return nil, fmt.Errorf("service enable: %w", err)
	}
	if resp.JSON200 == nil {
		return nil, failed(resp.StatusCode(),
			resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON404, resp.JSON500)
	}

	return serviceMutation(
		resp.JSON200.JobId,
		resp.JSON200.Results,
		resp.Body,
		resp.HTTPResponse,
	), nil
}

// Disable makes a service that osapi deployed on the target not start at
// boot. Idempotent.
func (s *ServiceService) Disable(
	ctx context.Context,
	hostname string,
	name string,
) (*Response[Collection[ServiceMutationResult]], error) {
	resp, err := s.client.PostNodeServiceDisableWithResponse(ctx, hostname, name)
	if err != nil {
		return nil, fmt.Errorf("service disable: %w", err)
	}
	if resp.JSON200 == nil {
		return nil, failed(resp.StatusCode(),
			resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON404, resp.JSON500)
	}

	return serviceMutation(
		resp.JSON200.JobId,
		resp.JSON200.Results,
		resp.Body,
		resp.HTTPResponse,
	), nil
}

// failed is the error of a response with no body for its success: the API
// error of its status, or a nil body.
func failed(
	statusCode int,
	responses ...*gen.ErrorResponse,
) error {
	if err := checkError(statusCode, responses...); err != nil {
		return err
	}

	return &UnexpectedStatusError{APIError{StatusCode: statusCode, Message: "nil response body"}}
}

// serviceMutation is the response of a write or an action.
func serviceMutation(
	jobID *openapi_types.UUID,
	rows []gen.ServiceResult,
	body []byte,
	httpResp *http.Response,
) *Response[Collection[ServiceMutationResult]] {
	results := make([]ServiceMutationResult, 0, len(rows))
	for _, r := range rows {
		results = append(results, ServiceMutationResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Error:    derefString(r.Error),
			Name:     derefString(r.Name),
			Changed:  derefBool(r.Changed),
		})
	}

	return newResponse(
		Collection[ServiceMutationResult]{Results: results, JobID: jobIDFromGen(jobID)},
		body,
		httpResp,
	)
}
