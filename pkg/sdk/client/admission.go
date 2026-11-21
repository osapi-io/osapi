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

// AdmissionService admits, rejects and revokes the nodes that ask to join.
type AdmissionService struct {
	client *gen.ClientWithResponses
}

// List returns each machine's admission record. state, when not "", keeps
// only the records in that state: pending, admitted, rejected or revoked.
func (s *AdmissionService) List(
	ctx context.Context,
	state string,
) (*Response[AdmissionList], error) {
	params := &gen.GetAdmissionsParams{}
	if state != "" {
		st := gen.GetAdmissionsParamsState(state)
		params.State = &st
	}
	resp, err := s.client.GetAdmissionsWithResponse(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list admission records: %w", err)
	}
	if err := checkError(resp.StatusCode(), resp.JSON400, resp.JSON401, resp.JSON403,
		resp.JSON500, resp.JSON503); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return newResponse(admissionListFromGen(resp.JSON200), resp.Body, resp.HTTPResponse), nil
}

// Get returns one machine's admission record.
func (s *AdmissionService) Get(
	ctx context.Context,
	machineID string,
) (*Response[Admission], error) {
	resp, err := s.client.GetAdmissionWithResponse(ctx, machineID)
	if err != nil {
		return nil, fmt.Errorf("get admission record %s: %w", machineID, err)
	}
	if err := checkError(resp.StatusCode(), resp.JSON400, resp.JSON401, resp.JSON403,
		resp.JSON404, resp.JSON500, resp.JSON503); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return newResponse(admissionFromGen(*resp.JSON200), resp.Body, resp.HTTPResponse), nil
}

// Admit admits a pending machine. fingerprint is the one that osapi agent
// admission prints on the machine; a different one is a ConflictError, and
// nothing is admitted.
func (s *AdmissionService) Admit(
	ctx context.Context,
	machineID string,
	fingerprint string,
) error {
	resp, err := s.client.AdmitMachineWithResponse(ctx, machineID,
		gen.AdmitMachineJSONRequestBody{Fingerprint: fingerprint})
	if err != nil {
		return fmt.Errorf("admit machine %s: %w", machineID, err)
	}

	return checkError(resp.StatusCode(), resp.JSON400, resp.JSON401, resp.JSON403,
		resp.JSON404, resp.JSON409, resp.JSON500, resp.JSON503)
}

// Reject rejects a pending machine, with why.
func (s *AdmissionService) Reject(
	ctx context.Context,
	machineID string,
	reason string,
) error {
	body := gen.RejectMachineJSONRequestBody{}
	if reason != "" {
		body.Reason = &reason
	}
	resp, err := s.client.RejectMachineWithResponse(ctx, machineID, body)
	if err != nil {
		return fmt.Errorf("reject machine %s: %w", machineID, err)
	}

	return checkError(resp.StatusCode(), resp.JSON400, resp.JSON401, resp.JSON403,
		resp.JSON404, resp.JSON500, resp.JSON503)
}

// Delete removes a machine's key and each record of it, and keeps none: a
// machine that asks again is new. Use it for a machine that goes away on
// purpose, such as a VM that is rebuilt; Revoke a machine that is not
// trusted.
func (s *AdmissionService) Delete(
	ctx context.Context,
	machineID string,
) error {
	resp, err := s.client.DeleteMachineWithResponse(ctx, machineID)
	if err != nil {
		return fmt.Errorf("delete machine %s: %w", machineID, err)
	}

	return checkError(resp.StatusCode(), resp.JSON400, resp.JSON401, resp.JSON403,
		resp.JSON404, resp.JSON500, resp.JSON503)
}

// Revoke removes an admitted machine's key. Nothing it signs verifies
// afterwards; to return, it asks again.
func (s *AdmissionService) Revoke(
	ctx context.Context,
	machineID string,
) error {
	resp, err := s.client.RevokeMachineWithResponse(ctx, machineID)
	if err != nil {
		return fmt.Errorf("revoke machine %s: %w", machineID, err)
	}

	return checkError(resp.StatusCode(), resp.JSON400, resp.JSON401, resp.JSON403,
		resp.JSON404, resp.JSON500, resp.JSON503)
}
