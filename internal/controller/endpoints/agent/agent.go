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

	"github.com/osapi-io/osapi/internal/secret"

	"github.com/osapi-io/osapi/internal/controller/admission"
	"github.com/osapi-io/osapi/internal/controller/jointoken"

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

// Group is the agent group's definition.
var Group = endpoint.Register(endpoint.Define(
	"agent",
	"Manage the agents in the fleet: discovery, drain and PKI enrollment",
	endpoint.Area(area.Fleet),
	endpoint.CommandLong(
		"Most commands read the NATS KV heartbeat registry directly, and make no job.",
	),
	endpoint.SubDoc("facts", "Read the keys of @fact references"),
	endpoint.SubDoc("config", "Read and set the config that the controller gives an agent"),
	endpoint.Title("Agent Management API"),
	endpoint.Tag("agent_operations", "Agent",
		"Operations related to agent discovery and details."),
	endpoint.Uses[Deps](),
	endpoint.Doc[types.AgentState](
		"The scheduling state of an agent: Ready, Draining, Cordoned or Pending. A drain that the agent has not applied yet is Draining.",
	),
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
	endpoint.Get[GetInput, Info]("get").
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
	endpoint.Post[JoinInput, JoinResponse]("join").
		Handle(Join).
		Sub("join").
		ID("JoinAgent").
		Public().
		NoCommand().
		Describe("Join an agent with a join token").
		Explain("An agent calls it one time, at its first start, over HTTPS, after it checks "+
			"the controller's certificate against the fingerprint of its join string. A token "+
			"that admits admits the machine at once, as admit does, and the admission record "+
			"names the token. The request is signed with the agent's signing key, which shows "+
			"that it holds the key that it asks the controller to admit. It needs no bearer "+
			"token: the join token authorizes it.").
		Returns("The machine is admitted.").
		Fails("no token", joinWith(func(r *JoinRequest) { r.Token = "" })).
		Fails("a token over 256 characters", joinWith(func(r *JoinRequest) { r.Token = secret.Value(strings.Repeat("a", 257)) })).
		Fails("no machine ID", joinWith(func(r *JoinRequest) { r.MachineID = "" })).
		Fails("a machine ID over 255 characters", joinWith(func(r *JoinRequest) { r.MachineID = long })).
		Fails("no hostname", joinWith(func(r *JoinRequest) { r.Hostname = "" })).
		Fails("a hostname over 255 characters", joinWith(func(r *JoinRequest) { r.Hostname = long })).
		Fails("no public key", joinWith(func(r *JoinRequest) { r.PublicKey = "" })).
		Fails("a public key over 128 characters", joinWith(func(r *JoinRequest) { r.PublicKey = strings.Repeat("a", 132) })).
		Fails("a public key that is not base64", joinWith(func(r *JoinRequest) { r.PublicKey = "not base64!" })).
		Fails("no encryption key", joinWith(func(r *JoinRequest) { r.EncryptionKey = "" })).
		Fails("an encryption key over 128 characters", joinWith(func(r *JoinRequest) { r.EncryptionKey = strings.Repeat("a", 129) })).
		Fails("no signature", joinWith(func(r *JoinRequest) { r.Signature = "" })).
		Fails("a signature over 256 characters", joinWith(func(r *JoinRequest) { r.Signature = strings.Repeat("a", 260) })).
		Fails("a signature that is not base64", joinWith(func(r *JoinRequest) { r.Signature = "not base64!" })).
		Errors(
			endpoint.Forbidden(jointoken.ErrInvalid, "The token is not one the controller has, or it was revoked."),
			endpoint.Forbidden(jointoken.ErrExpired, "The token has expired."),
			endpoint.Forbidden(jointoken.ErrUsed, "The token has no use left."),
			endpoint.Forbidden(admission.ErrJoinSignature, "The request is not signed by the key that it asks to admit."),
			endpoint.Unavailable(admission.ErrNoJoinTokens, "This controller takes no join."),
		),
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
	endpoint.Get[ConfigGetInput, ConfigResponse]("getconfig").
		Handle(GetConfig).
		Sub("config").
		Use("get").
		ID("GetAgentConfig").
		Permission(types.PermAgentRead).
		Describe("Read the config of an agent").
		Explain("Returns the config that the controller holds for the agent, with its version.").
		Returns("The config of the agent.").
		Command("Read the config of an agent",
			"Shows the labels that the controller gives the agent, and the version of\n"+
				"the config. The agent applies each version at once.").
		Example("--hostname web-01").
		Prints(sampleConfig).
		Fails("a hostname over 255 characters", ConfigGetInput{Hostname: long}).
		Errors(endpoint.NotFound(ErrNotFound, "Agent not found.")),
	endpoint.Put[ConfigSetInput, ConfigResponse]("setconfig").
		Handle(SetConfig).
		Sub("config").
		Use("update").
		ID("PutAgentConfig").
		AdminOnly().
		Describe("Set the config of an agent").
		Explain("Sets the labels, the max jobs, the log level and the facts interval of the agent "+
			"as the next version of its config. The agent applies them at once: it takes the jobs of the "+
			"new labels, runs at most max jobs at the same time, logs at the level, collects "+
			"its facts at the interval, and no restart is necessary.").
		Returns("The config of the agent after the change.").
		Command("Set the labels, the max jobs, the log level and the facts interval of an agent",
			"The labels replace the labels of the agent's config file with the same\n"+
				"keys. Max jobs replaces its max_jobs, the log level the level of its debug\n"+
				"setting, and the facts interval its agent.facts.interval; with none, the\n"+
				"agent keeps the file's. The agent\n"+
				"applies them at once, with no restart. A host that joins with a join\n"+
				"token gets the labels of the token.").
		Example("--hostname web-01 --labels group=web --labels dc=east --max-jobs 4 --log-level debug --facts-interval 5m").
		Prints(sampleConfig).
		Fails("a hostname over 255 characters", ConfigSetInput{Hostname: long}).
		Fails("max jobs below 1", ConfigSetInput{
			Hostname: "web-01", Body: ConfigRequest{MaxJobs: -1},
		}).
		Fails("max jobs over 1000", ConfigSetInput{
			Hostname: "web-01", Body: ConfigRequest{MaxJobs: 1001},
		}).
		Fails("a facts interval that is not a Go duration", ConfigSetInput{
			Hostname: "web-01", Body: ConfigRequest{FactsInterval: "7d"},
		}).
		Fails("a log level that is not one", ConfigSetInput{
			Hostname: "web-01", Body: ConfigRequest{LogLevel: types.LogLevel("verbose")},
		}).
		Errors(endpoint.NotFound(ErrNotFound, "Agent not found.")),
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

// joinWith is a valid join, with one change, for the Fails of join.
func joinWith(
	change func(*JoinRequest),
) JoinInput {
	r := JoinRequest{
		Token:         "osapi_jt_1a2b3c4d.secret",
		MachineID:     "d530664edc1c431eb5ef1e9dfd6617a5",
		Hostname:      "web-01",
		PublicKey:     "MCowBQYDK2VwAyEA",
		EncryptionKey: "XA7NTLWJ2ECILFDLHORNVYROXOK65ONQ2SJYHRNAWRJZ3UXMVXIGUVJM",
		Signature:     "c2lnbmF0dXJl",
	}
	change(&r)

	return JoinInput{Body: r}
}
