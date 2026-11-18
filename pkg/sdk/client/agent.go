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

// AgentService provides agent discovery and details operations.
type AgentService struct {
	client *gen.ClientWithResponses
}

// List retrieves all active agents.
func (s *AgentService) List(
	ctx context.Context,
) (*Response[AgentList], error) {
	resp, err := s.client.GetAgentsWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("list agents: %w", err)
	}

	if err := checkError(resp.StatusCode(), resp.JSON401, resp.JSON403, resp.JSON500); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return newResponse(agentListFromGen(resp.JSON200), resp.Body, resp.HTTPResponse), nil
}

// Get retrieves detailed information about a specific agent by hostname.
func (s *AgentService) Get(
	ctx context.Context,
	hostname string,
) (*Response[Agent], error) {
	resp, err := s.client.GetAgentDetailsWithResponse(ctx, hostname)
	if err != nil {
		return nil, fmt.Errorf("get agent %s: %w", hostname, err)
	}

	if err := checkError(resp.StatusCode(), resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON404, resp.JSON500); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return newResponse(agentFromGen(resp.JSON200), resp.Body, resp.HTTPResponse), nil
}

// Drain initiates draining of an agent, stopping it from accepting
// new jobs while allowing in-flight jobs to complete.
func (s *AgentService) Drain(
	ctx context.Context,
	hostname string,
) error {
	resp, err := s.client.DrainAgentWithResponse(ctx, hostname)
	if err != nil {
		return fmt.Errorf("drain agent %s: %w", hostname, err)
	}

	return checkError(
		resp.StatusCode(),
		resp.JSON401,
		resp.JSON403,
		resp.JSON404,
		resp.JSON409,
		resp.JSON500,
	)
}

// Undrain resumes job acceptance on a drained agent.
func (s *AgentService) Undrain(
	ctx context.Context,
	hostname string,
) error {
	resp, err := s.client.UndrainAgentWithResponse(ctx, hostname)
	if err != nil {
		return fmt.Errorf("undrain agent %s: %w", hostname, err)
	}

	return checkError(
		resp.StatusCode(),
		resp.JSON401,
		resp.JSON403,
		resp.JSON404,
		resp.JSON409,
		resp.JSON500,
	)
}

// ListPending retrieves all agents awaiting enrollment acceptance.
func (s *AgentService) ListPending(
	ctx context.Context,
) (*Response[PendingAgentList], error) {
	resp, err := s.client.GetAgentsPendingWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("list pending agents: %w", err)
	}

	if err := checkError(resp.StatusCode(), resp.JSON401, resp.JSON403, resp.JSON500); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return newResponse(pendingAgentListFromGen(resp.JSON200), resp.Body, resp.HTTPResponse), nil
}

// Accept accepts a pending agent enrollment. If fingerprint is provided,
// the agent is accepted by fingerprint; otherwise by hostname.
func (s *AgentService) Accept(
	ctx context.Context,
	hostname string,
	fingerprint string,
) error {
	var params *gen.AcceptAgentParams
	if fingerprint != "" {
		params = &gen.AcceptAgentParams{
			Fingerprint: &fingerprint,
		}
	}

	resp, err := s.client.AcceptAgentWithResponse(ctx, hostname, params)
	if err != nil {
		return fmt.Errorf("accept agent %s: %w", hostname, err)
	}

	return checkError(resp.StatusCode(), resp.JSON401, resp.JSON403, resp.JSON404, resp.JSON500)
}

// Reject rejects a pending agent enrollment.
func (s *AgentService) Reject(
	ctx context.Context,
	hostname string,
) error {
	resp, err := s.client.RejectAgentWithResponse(ctx, hostname)
	if err != nil {
		return fmt.Errorf("reject agent %s: %w", hostname, err)
	}

	return checkError(resp.StatusCode(), resp.JSON401, resp.JSON403, resp.JSON404, resp.JSON500)
}

// Remove removes an accepted agent's stored key, so nothing it signs verifies
// afterwards. machineID chooses one agent when more than one accepted agent
// has the hostname; pass "" otherwise. Without it, a shared hostname returns a
// ConflictError naming the machine IDs.
func (s *AgentService) Delete(
	ctx context.Context,
	hostname string,
	machineID string,
) error {
	params := &gen.DeleteAgentParams{}
	if machineID != "" {
		params.MachineId = &machineID
	}

	resp, err := s.client.DeleteAgentWithResponse(ctx, hostname, params)
	if err != nil {
		return fmt.Errorf("delete agent %s: %w", hostname, err)
	}

	return checkError(
		resp.StatusCode(),
		resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON404, resp.JSON409, resp.JSON500,
	)
}
