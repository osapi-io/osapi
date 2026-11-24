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
// agents that report, and the actions on one agent. Admitting an agent is
// the admission group. Its spec part comes from this definition; until the group
// moves, the controller serves the hand-written spec in
// internal/controller/api/agent, and TestEndpointParity holds the two
// together.
package agent

import (
	"errors"
	"net/http"
	"strings"

	"github.com/osapi-io/osapi/internal/area"

	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

var (
	// ErrNotFound is an agent that is not there: not registered, or not
	// pending.
	ErrNotFound = errors.New("agent not found")
	// ErrState is an agent that is already in the state that was asked
	// for, or cannot go to it.
	ErrState = errors.New("agent is not in a state that allows it")
)

// agentStates are the values of types.AgentState.
func agentStates() []string {
	values := make([]string, 0, len(types.AgentStates))
	for _, s := range types.AgentStates {
		values = append(values, string(s))
	}

	return values
}

// conditionTypes are the values of types.ConditionType.
func conditionTypes() []string {
	values := make([]string, 0, len(types.ConditionTypes))
	for _, c := range types.ConditionTypes {
		values = append(values, string(c))
	}

	return values
}

// Group is the agent group's definition.
var Group = endpoint.Register(endpoint.Define(
	"agent",
	"Manage the agents in the fleet: discovery, drain and PKI enrollment",
	endpoint.Area(area.Fleet),
	endpoint.CommandLong(
		"Most commands read the NATS KV heartbeat registry directly, and make no job.",
	),
	endpoint.SubDoc("facts", "Read the keys of @fact references"),
	endpoint.Title("Agent Management API"),
	endpoint.Tag("agent_operations", "Agent",
		"Operations related to agent discovery and details."),
	endpoint.Uses[Deps](),
	endpoint.Enum[types.AgentState](agentStates),
	endpoint.Doc[types.AgentState](
		"The scheduling state of an agent: Ready, Draining, Cordoned or Pending. A drain that the agent has not applied yet is Draining.",
	),
	endpoint.Enum[types.ConditionType](conditionTypes),
	endpoint.Doc[types.ConditionType]("A condition that an agent or a component evaluates."),
	endpoint.Doc[LoadAverageResponse]("The system load averages for 1, 5, and 15 minutes."),
	endpoint.Doc[MemoryResponse]("Memory usage information."),
	endpoint.Doc[OSInfoResponse]("Operating system information."),
	endpoint.Doc[RouteResponse]("A network routing table entry."),
	endpoint.Get[FactsInput, FactKeysResponse]("facts").
		Handle(Facts).
		Sub("facts").
		Use("list").
		ID("GetAgentFacts").
		Permission(types.PermAgentRead).
		Describe("List the fact keys").
		Explain("Returns the fact keys that an input can refer to as @fact.<key>: the "+
			"built-in keys and the custom keys that agents register.").
		Returns("The fact keys.").
		Command("List the fact keys",
			"Lists each fact key that an input can refer to as @fact.<key>, with\n"+
				"its description.").
		Example("").
		Prints(sampleFacts).
		Example("--json"),
	endpoint.Get[ListInput, ListAgentsResponse]("list").
		Notes("The `--json` output has more system facts that the agent collects:\n"+
			"architecture, kernel version, FQDN, CPU count, network interfaces, service\n"+
			"manager and package manager. The table view does not show these fields.").
		Command("List active agents",
			"This command reads the agent heartbeat registry directly. It makes no job.\nEach agent writes a heartbeat every 10 seconds, with a 30-second TTL. When\nan agent stops its heartbeat, it goes from the list.").
		Example("").
		Prints(sampleList).
		Handle(List).
		ID("GetAgents").
		Describe("List active agents").
		Explain("Discover all active agents in the fleet.").
		Returns("List of active agents."),
	endpoint.Get[GetInput, AgentInfo]("get").
		Fails("a hostname over 255 characters", GetInput{Hostname: long}).
		Command("Get agent details",
			"This command reads the agent heartbeat registry directly. It makes no job.\nThe data comes from the last heartbeat of the agent. `node status` runs a\nfull inspection through the job system, with disk usage and more metrics.").
		Example("--hostname web-01").
		Prints(sampleWeb01).
		Handle(Get).
		ID("GetAgentDetails").
		Describe("Get agent details").
		Explain("Get detailed information about a specific agent by hostname.").
		Returns("Agent details.").
		Errors(endpoint.NotFound(ErrNotFound, "Agent not found")),
	endpoint.Post[DrainInput, struct{}]("drain").
		AdminOnly().
		Fails("a hostname over 255 characters", DrainInput{Hostname: long}).
		Status(http.StatusNoContent).
		Command("Drain an agent",
			"The agent changes from `Ready` to `Draining`. When all its running jobs are\ncomplete, the state is `Cordoned`. The agent continues to run and to send\nheartbeats. It only stops the intake of new work from the job queue. Use\n`agent undrain` to accept jobs again. The jobs that it runs now continue\nuntil they are complete.").
		Example("--hostname web-01").
		Handle(Drain).
		Action("drain").
		Says("draining").
		ID("DrainAgent").
		Describe("Drain an agent").
		Explain("Stop the agent from accepting new jobs. In-flight jobs continue to "+
			"completion.").
		Returns("Agent drain initiated.").
		Errors(
			endpoint.NotFound(ErrNotFound, "Agent not found."),
			endpoint.Conflict(ErrState, "Agent already in requested state."),
		),
	endpoint.Post[UndrainInput, struct{}]("undrain").
		AdminOnly().
		Fails("a hostname over 255 characters", UndrainInput{Hostname: long}).
		Status(http.StatusNoContent).
		Command("Undrain an agent",
			"The agent subscribes to the NATS JetStream consumers again and changes back\nto `Ready`.").
		Example("--hostname web-01").
		Handle(Undrain).
		Action("undrain").
		Says("ready").
		ID("UndrainAgent").
		Describe("Undrain an agent").
		Explain("Resume accepting jobs on a drained agent.").
		Returns("Agent undrain initiated.").
		Errors(
			endpoint.NotFound(ErrNotFound, "Agent not found."),
			endpoint.Conflict(ErrState, "Agent not in draining or cordoned state."),
		),
))

// long is a name over the 255 characters each name may have, for the inputs
// that Fails declares.
var long = strings.Repeat("a", 256)
