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

// NodeService provides node management operations.
type NodeService struct {
	client *gen.ClientWithResponses
}

// ExecRequest contains parameters for direct command execution.
type ExecRequest struct {
	// Command is the binary to execute (required).
	Command string

	// Args is the argument list passed to the command.
	Args []string

	// Cwd is the working directory. Empty uses the agent default.
	Cwd string

	// Timeout in seconds. Zero uses the server default (30s).
	Timeout int

	// Target specifies the host: "_any", "_all", hostname, or
	// label ("group:web").
	Target string
}

// FileDeployOpts contains parameters for file deployment.
type FileDeployOpts struct {
	// ObjectName is the name of the file in the Object Store (required).
	ObjectName string

	// Path is the destination path on the target filesystem (required).
	Path string

	// ContentType is "raw" or "template" (required).
	ContentType string

	// Mode is the file permission mode (e.g., "0644"). Optional.
	Mode string

	// Owner is the file owner user. Optional.
	Owner string

	// Group is the file owner group. Optional.
	Group string

	// Vars are template variables when ContentType is "template". Optional.
	Vars map[string]any

	// Target specifies the host: "_any", "_all", hostname, or
	// label ("group:web").
	Target string
}

// FileUndeployOpts contains parameters for file undeployment.
type FileUndeployOpts struct {
	// Path is the filesystem path to remove from the target host (required).
	Path string

	// Target specifies the host: "_any", "_all", hostname, or
	// label ("group:web").
	Target string
}

// ShellRequest contains parameters for shell command execution.
type ShellRequest struct {
	// Command is the shell command string passed to /bin/sh -c (required).
	Command string

	// Cwd is the working directory. Empty uses the agent default.
	Cwd string

	// Timeout in seconds. Zero uses the server default (30s).
	Timeout int

	// Target specifies the host: "_any", "_all", hostname, or
	// label ("group:web").
	Target string
}

// Status retrieves node status (OS info, disk, memory, load) from the
// target host.
func (s *NodeService) Status(
	ctx context.Context,
	target string,
) (*Response[Collection[NodeStatus]], error) {
	resp, err := s.client.GetNodeStatusWithResponse(ctx, target)
	if err != nil {
		return nil, fmt.Errorf("get status: %w", err)
	}

	if err := checkError(resp.StatusCode(), resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON500); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(nodeStatusCollectionFromGen(resp.JSON200), resp.Body), nil
}

// Hostname retrieves the hostname from the target host.
func (s *NodeService) Hostname(
	ctx context.Context,
	target string,
) (*Response[Collection[HostnameResult]], error) {
	resp, err := s.client.GetNodeHostnameWithResponse(ctx, target)
	if err != nil {
		return nil, fmt.Errorf("get hostname: %w", err)
	}

	if err := checkError(resp.StatusCode(), resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON500); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(hostnameCollectionFromGen(resp.JSON200), resp.Body), nil
}

// Disk retrieves disk usage information from the target host.
func (s *NodeService) Disk(
	ctx context.Context,
	target string,
) (*Response[Collection[DiskResult]], error) {
	resp, err := s.client.GetNodeDiskWithResponse(ctx, target)
	if err != nil {
		return nil, fmt.Errorf("get disk: %w", err)
	}

	if err := checkError(resp.StatusCode(), resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON500); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(diskCollectionFromGen(resp.JSON200), resp.Body), nil
}

// Memory retrieves memory usage information from the target host.
func (s *NodeService) Memory(
	ctx context.Context,
	target string,
) (*Response[Collection[MemoryResult]], error) {
	resp, err := s.client.GetNodeMemoryWithResponse(ctx, target)
	if err != nil {
		return nil, fmt.Errorf("get memory: %w", err)
	}

	if err := checkError(resp.StatusCode(), resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON500); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(memoryCollectionFromGen(resp.JSON200), resp.Body), nil
}

// Load retrieves load average information from the target host.
func (s *NodeService) Load(
	ctx context.Context,
	target string,
) (*Response[Collection[LoadResult]], error) {
	resp, err := s.client.GetNodeLoadWithResponse(ctx, target)
	if err != nil {
		return nil, fmt.Errorf("get load: %w", err)
	}

	if err := checkError(resp.StatusCode(), resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON500); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(loadCollectionFromGen(resp.JSON200), resp.Body), nil
}

// OS retrieves operating system information from the target host.
func (s *NodeService) OS(
	ctx context.Context,
	target string,
) (*Response[Collection[OSInfoResult]], error) {
	resp, err := s.client.GetNodeOSWithResponse(ctx, target)
	if err != nil {
		return nil, fmt.Errorf("get os: %w", err)
	}

	if err := checkError(resp.StatusCode(), resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON500); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(osInfoCollectionFromGen(resp.JSON200), resp.Body), nil
}

// Uptime retrieves uptime information from the target host.
func (s *NodeService) Uptime(
	ctx context.Context,
	target string,
) (*Response[Collection[UptimeResult]], error) {
	resp, err := s.client.GetNodeUptimeWithResponse(ctx, target)
	if err != nil {
		return nil, fmt.Errorf("get uptime: %w", err)
	}

	if err := checkError(resp.StatusCode(), resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON500); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(uptimeCollectionFromGen(resp.JSON200), resp.Body), nil
}

// GetDNS retrieves DNS configuration for a network interface on the
// target host.
func (s *NodeService) GetDNS(
	ctx context.Context,
	target string,
	interfaceName string,
) (*Response[Collection[DNSConfig]], error) {
	resp, err := s.client.GetNodeNetworkDNSByInterfaceWithResponse(ctx, target, interfaceName)
	if err != nil {
		return nil, fmt.Errorf("get dns: %w", err)
	}

	if err := checkError(resp.StatusCode(), resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON500); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(dnsConfigCollectionFromGen(resp.JSON200), resp.Body), nil
}

// UpdateDNS updates DNS configuration for a network interface on the
// target host.
func (s *NodeService) UpdateDNS(
	ctx context.Context,
	target string,
	interfaceName string,
	servers []string,
	searchDomains []string,
) (*Response[Collection[DNSUpdateResult]], error) {
	body := gen.DNSConfigUpdateRequest{
		InterfaceName: interfaceName,
	}

	if len(servers) > 0 {
		body.Servers = &servers
	}

	if len(searchDomains) > 0 {
		body.SearchDomains = &searchDomains
	}

	resp, err := s.client.PutNodeNetworkDNSWithResponse(ctx, target, body)
	if err != nil {
		return nil, fmt.Errorf("update dns: %w", err)
	}

	if err := checkError(resp.StatusCode(), resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON500); err != nil {
		return nil, err
	}

	if resp.JSON202 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(dnsUpdateCollectionFromGen(resp.JSON202), resp.Body), nil
}

// Ping sends an ICMP ping to the specified address from the target host.
func (s *NodeService) Ping(
	ctx context.Context,
	target string,
	address string,
) (*Response[Collection[PingResult]], error) {
	body := gen.PostNodeNetworkPingJSONRequestBody{
		Address: address,
	}

	resp, err := s.client.PostNodeNetworkPingWithResponse(ctx, target, body)
	if err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	if err := checkError(resp.StatusCode(), resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON500); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(pingCollectionFromGen(resp.JSON200), resp.Body), nil
}

// Exec executes a command directly without a shell interpreter.
func (s *NodeService) Exec(
	ctx context.Context,
	req ExecRequest,
) (*Response[Collection[CommandResult]], error) {
	body := gen.CommandExecRequest{
		Command: req.Command,
	}

	if len(req.Args) > 0 {
		body.Args = &req.Args
	}

	if req.Cwd != "" {
		body.Cwd = &req.Cwd
	}

	if req.Timeout > 0 {
		body.Timeout = &req.Timeout
	}

	resp, err := s.client.PostNodeCommandExecWithResponse(ctx, req.Target, body)
	if err != nil {
		return nil, fmt.Errorf("exec command: %w", err)
	}

	if err := checkError(resp.StatusCode(), resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON500); err != nil {
		return nil, err
	}

	if resp.JSON202 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(commandCollectionFromGen(resp.JSON202), resp.Body), nil
}

// Shell executes a command through /bin/sh -c with shell features
// (pipes, redirects, variable expansion).
func (s *NodeService) Shell(
	ctx context.Context,
	req ShellRequest,
) (*Response[Collection[CommandResult]], error) {
	body := gen.CommandShellRequest{
		Command: req.Command,
	}

	if req.Cwd != "" {
		body.Cwd = &req.Cwd
	}

	if req.Timeout > 0 {
		body.Timeout = &req.Timeout
	}

	resp, err := s.client.PostNodeCommandShellWithResponse(ctx, req.Target, body)
	if err != nil {
		return nil, fmt.Errorf("shell command: %w", err)
	}

	if err := checkError(resp.StatusCode(), resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON500); err != nil {
		return nil, err
	}

	if resp.JSON202 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(commandCollectionFromGen(resp.JSON202), resp.Body), nil
}

// FileDeploy deploys a file from the Object Store to the target host.
func (s *NodeService) FileDeploy(
	ctx context.Context,
	req FileDeployOpts,
) (*Response[Collection[FileDeployResult]], error) {
	body := gen.FileDeployRequest{
		ObjectName:  req.ObjectName,
		Path:        req.Path,
		ContentType: gen.FileDeployRequestContentType(req.ContentType),
	}

	if req.Mode != "" {
		body.Mode = &req.Mode
	}

	if req.Owner != "" {
		body.Owner = &req.Owner
	}

	if req.Group != "" {
		body.Group = &req.Group
	}

	if len(req.Vars) > 0 {
		body.Vars = &req.Vars
	}

	resp, err := s.client.PostNodeFileDeployWithResponse(ctx, req.Target, body)
	if err != nil {
		return nil, fmt.Errorf("file deploy: %w", err)
	}

	if err := checkError(resp.StatusCode(), resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON500); err != nil {
		return nil, err
	}

	if resp.JSON202 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(fileDeployCollectionFromGen(resp.JSON202), resp.Body), nil
}

// FileUndeploy removes a deployed file from the target host filesystem.
func (s *NodeService) FileUndeploy(
	ctx context.Context,
	req FileUndeployOpts,
) (*Response[Collection[FileUndeployResult]], error) {
	body := gen.FileUndeployRequest{
		Path: req.Path,
	}

	resp, err := s.client.PostNodeFileUndeployWithResponse(ctx, req.Target, body)
	if err != nil {
		return nil, fmt.Errorf("file undeploy: %w", err)
	}

	if err := checkError(resp.StatusCode(), resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON500); err != nil {
		return nil, err
	}

	if resp.JSON202 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(fileUndeployCollectionFromGen(resp.JSON202), resp.Body), nil
}

// FileStatus checks the deployment status of a file on the target host.
func (s *NodeService) FileStatus(
	ctx context.Context,
	target string,
	path string,
) (*Response[Collection[FileStatusResult]], error) {
	body := gen.FileStatusRequest{
		Path: path,
	}

	resp, err := s.client.PostNodeFileStatusWithResponse(ctx, target, body)
	if err != nil {
		return nil, fmt.Errorf("file status: %w", err)
	}

	if err := checkError(resp.StatusCode(), resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON500); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(fileStatusCollectionFromGen(resp.JSON200), resp.Body), nil
}
