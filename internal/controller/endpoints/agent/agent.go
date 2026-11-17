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
// Package agent is the endpoint group for the agents of the fleet: the
// agents that report, the ones that wait for enrollment, and the actions on
// one agent. Its spec part comes from this definition; until the group
// moves, the controller serves the hand-written spec in
// internal/controller/api/agent, and TestEndpointParity holds the two
// together.
package agent

import (
	"errors"

	"github.com/osapi-io/osapi/internal/controller/endpoint"
)

var (
	// ErrNotFound is an agent that is not there: not registered, or not
	// pending.
	ErrNotFound = errors.New("agent not found")
	// ErrAmbiguous is a hostname that more than one accepted agent has.
	ErrAmbiguous = errors.New("more than one accepted agent has the hostname")
	// ErrState is an agent that is already in the state that was asked
	// for, or cannot go to it.
	ErrState = errors.New("agent is not in a state that allows it")
)

// Group is the agent group's definition.
var Group = endpoint.Register(endpoint.Define(
	"agent", "The agents of the fleet",
	endpoint.Title("Agent Management API"),
	endpoint.Tag("agent_operations", "Agent",
		"Operations related to agent discovery and details."),
	endpoint.Get[ListInput, ListAgentsResponse]("list").
		ID("getAgents").
		Describe("List active agents").
		Explain("Discover all active agents in the fleet.").
		Returns("List of active agents."),
	endpoint.Get[ListInput, ListPendingAgentsResponse]("pending").
		Sub("pending").
		ID("getAgentsPending").
		Describe("List agents awaiting enrollment").
		Explain("List all agents that have submitted enrollment requests but have "+
			"not yet been accepted or rejected.").
		Returns("List of pending agents."),
	endpoint.Get[GetInput, AgentInfo]("get").
		ID("getAgentDetails").
		Describe("Get agent details").
		Explain("Get detailed information about a specific agent by hostname.").
		Returns("Agent details.").
		Errors(endpoint.NotFound(ErrNotFound, "Agent not found")),
	endpoint.Delete[RemoveInput, Message]("remove").
		ID("removeAgent").
		Describe("Remove an enrolled agent").
		Explain("Remove an accepted agent's stored public key. Its responses and "+
			"registration stop verifying at once, so targeting no longer resolves it. "+
			"The agent does not enroll again by itself: remove its controller key on the "+
			"host and restart it. Two machines may share a hostname; when they do, name "+
			"one with machine_id.").
		Returns("Agent removed.").
		Errors(
			endpoint.NotFound(ErrNotFound, "No accepted agent with that hostname."),
			endpoint.Conflict(ErrAmbiguous,
				"More than one accepted agent has the hostname; name one with machine_id."),
		),
	endpoint.Post[AcceptInput, Message]("accept").
		Action("accept").
		ID("acceptAgent").
		Describe("Accept a pending agent").
		Explain("Accept a pending agent enrollment request. The agent can be "+
			"identified by hostname (path parameter) or by fingerprint (query parameter).").
		Returns("Agent accepted.").
		Errors(endpoint.NotFound(ErrNotFound, "Pending agent not found.")),
	endpoint.Post[RejectInput, Message]("reject").
		Action("reject").
		ID("rejectAgent").
		Describe("Reject a pending agent").
		Explain("Reject a pending agent enrollment request.").
		Returns("Agent rejected.").
		Errors(endpoint.NotFound(ErrNotFound, "Pending agent not found.")),
	endpoint.Post[DrainInput, Message]("drain").
		Action("drain").
		ID("drainAgent").
		Describe("Drain an agent").
		Explain("Stop the agent from accepting new jobs. In-flight jobs continue to "+
			"completion.").
		Returns("Agent drain initiated.").
		Errors(
			endpoint.NotFound(ErrNotFound, "Agent not found."),
			endpoint.Conflict(ErrState, "Agent already in requested state."),
		),
	endpoint.Post[UndrainInput, Message]("undrain").
		Action("undrain").
		ID("undrainAgent").
		Describe("Undrain an agent").
		Explain("Resume accepting jobs on a drained agent.").
		Returns("Agent undrain initiated.").
		Errors(
			endpoint.NotFound(ErrNotFound, "Agent not found."),
			endpoint.Conflict(ErrState, "Agent not in draining or cordoned state."),
		),
))
