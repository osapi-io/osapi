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

// MessageResponse represents a simple message response from the API.
type MessageResponse struct {
	Message string
}

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

	return NewResponse(agentListFromGen(resp.JSON200), resp.Body), nil
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

	if err := checkError(resp.StatusCode(), resp.JSON401, resp.JSON403, resp.JSON404, resp.JSON500); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(agentFromGen(resp.JSON200), resp.Body), nil
}

// Drain initiates draining of an agent, stopping it from accepting
// new jobs while allowing in-flight jobs to complete.
func (s *AgentService) Drain(
	ctx context.Context,
	hostname string,
) (*Response[MessageResponse], error) {
	resp, err := s.client.DrainAgentWithResponse(ctx, hostname)
	if err != nil {
		return nil, fmt.Errorf("drain agent %s: %w", hostname, err)
	}

	if err := checkError(resp.StatusCode(), resp.JSON401, resp.JSON403, resp.JSON404, resp.JSON409); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	msg := MessageResponse{
		Message: resp.JSON200.Message,
	}

	return NewResponse(msg, resp.Body), nil
}

// Undrain resumes job acceptance on a drained agent.
func (s *AgentService) Undrain(
	ctx context.Context,
	hostname string,
) (*Response[MessageResponse], error) {
	resp, err := s.client.UndrainAgentWithResponse(ctx, hostname)
	if err != nil {
		return nil, fmt.Errorf("undrain agent %s: %w", hostname, err)
	}

	if err := checkError(resp.StatusCode(), resp.JSON401, resp.JSON403, resp.JSON404, resp.JSON409); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	msg := MessageResponse{
		Message: resp.JSON200.Message,
	}

	return NewResponse(msg, resp.Body), nil
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

	return NewResponse(pendingAgentListFromGen(resp.JSON200), resp.Body), nil
}

// Accept accepts a pending agent enrollment. If fingerprint is provided,
// the agent is accepted by fingerprint; otherwise by hostname.
func (s *AgentService) Accept(
	ctx context.Context,
	hostname string,
	fingerprint string,
) (*Response[MessageResponse], error) {
	var params *gen.AcceptAgentParams
	if fingerprint != "" {
		params = &gen.AcceptAgentParams{
			Fingerprint: &fingerprint,
		}
	}

	resp, err := s.client.AcceptAgentWithResponse(ctx, hostname, params)
	if err != nil {
		return nil, fmt.Errorf("accept agent %s: %w", hostname, err)
	}

	if err := checkError(resp.StatusCode(), resp.JSON401, resp.JSON403, resp.JSON404, resp.JSON500); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	msg := MessageResponse{
		Message: resp.JSON200.Message,
	}

	return NewResponse(msg, resp.Body), nil
}

// Reject rejects a pending agent enrollment.
func (s *AgentService) Reject(
	ctx context.Context,
	hostname string,
) (*Response[MessageResponse], error) {
	resp, err := s.client.RejectAgentWithResponse(ctx, hostname)
	if err != nil {
		return nil, fmt.Errorf("reject agent %s: %w", hostname, err)
	}

	if err := checkError(resp.StatusCode(), resp.JSON401, resp.JSON403, resp.JSON404, resp.JSON500); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	msg := MessageResponse{
		Message: resp.JSON200.Message,
	}

	return NewResponse(msg, resp.Body), nil
}
