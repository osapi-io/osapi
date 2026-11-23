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

package docker

// ListInput is list's input.
type ListInput struct {
	State string `json:"state,omitempty" validate:"omitempty,oneof=running stopped all" doc:"Only the containers in this state: running, stopped or all, the default"`
}

// GetInput is get's input: the container, by its name or its ID.
type GetInput struct {
	Name string `json:"name" validate:"required,container_id" doc:"The container's name or ID"`
}

// CreateInput is create's input.
type CreateInput struct {
	Name      string            `json:"name"                         validate:"required,container_id"         doc:"The container's name"                                                                  telemetry:"attr"`
	Image     string            `json:"image"                        validate:"required,image_ref"            doc:"The image to run, such as alpine:3.20; docker pulls it when the host does not have it"`
	HostName  string            `json:"container_hostname,omitempty" validate:"omitempty,dns_name,max=253"    doc:"The hostname inside the container"`
	DNS       []string          `json:"dns,omitempty"                validate:"omitempty,dive,ip"             doc:"The DNS servers of the container"`
	Command   []string          `json:"command,omitempty"            validate:"omitempty"                     doc:"The command, in place of the image's own"`
	Env       map[string]string `json:"env,omitempty"                validate:"omitempty"                     doc:"The environment variables of the container"                                                             secret:"true"`
	Ports     []string          `json:"ports,omitempty"              validate:"omitempty,dive,port_mapping"   doc:"The published ports, each host:container, such as 8080:80"`
	Volumes   []string          `json:"volumes,omitempty"            validate:"omitempty,dive,volume_mapping" doc:"The bind mounts, each host:container, such as /srv/data:/data"`
	AutoStart bool              `json:"auto_start,omitempty"         validate:"omitempty"                     doc:"Start the container once it is created"`
}

// DeleteInput is delete's input.
type DeleteInput struct {
	Name string `json:"name" validate:"required,container_id" doc:"The container's name or ID"`
}

// StartInput is start's input.
type StartInput struct {
	Name string `json:"name" validate:"required,container_id" doc:"The container's name or ID"`
}

// StopInput is stop's input.
type StopInput struct {
	Name    string `json:"name"              validate:"required,container_id"   doc:"The container's name or ID"`
	Timeout int    `json:"timeout,omitempty" validate:"omitempty,min=0,max=300" doc:"Seconds to wait before docker kills it; docker's default when 0"`
}

// ExecInput is exec's input.
type ExecInput struct {
	Name       string            `json:"name"                  validate:"required,container_id"  doc:"The container's name or ID"`
	Command    []string          `json:"command"               validate:"required,min=1"         doc:"The command and its arguments"`
	Env        map[string]string `json:"env,omitempty"         validate:"omitempty"              doc:"Environment variables for the command" secret:"true"`
	WorkingDir string            `json:"working_dir,omitempty" validate:"omitempty,startswith=/" doc:"The directory the command runs in"`
}

// Entry is one container, as list and get answer it.
type Entry struct {
	ID       string   `json:"id"                  doc:"The container's full ID"`
	Name     string   `json:"name"                doc:"The container's name"`
	Image    string   `json:"image"               doc:"The image it runs"`
	State    string   `json:"state"               doc:"running, exited, created, paused or restarting"                                   render:"status"`
	Summary  string   `json:"summary,omitempty"   doc:"What docker says of its state, such as Up 2 minutes"                                              table:"-"`
	Created  string   `json:"created,omitempty"   doc:"When it was created, RFC 3339"                                                    render:"time"`
	Health   string   `json:"health,omitempty"    doc:"The state of its health check: healthy, unhealthy or starting; empty without one"                 table:"-"`
	Ports    []string `json:"ports,omitempty"     doc:"The published ports, host:port->container port"                                                   table:"-"`
	Mounts   []string `json:"mounts,omitempty"    doc:"The mounts, source:destination, with :ro when read only"                                          table:"-"`
	Networks []string `json:"networks,omitempty"  doc:"The networks, with the container's address and gateway on each"                                   table:"-"`
	EnvNames []string `json:"env_names,omitempty" doc:"The names of its environment variables; osapi never reads their values"                           table:"-"`
}

// Result is the output of create, delete, start and stop.
type Result struct {
	Name    string `json:"name"         doc:"The container"`
	ID      string `json:"id,omitempty" doc:"The container's full ID, for create"`
	Changed bool   `json:"changed"      doc:"Whether the host changed"`
}

// ExecResult is exec's output.
type ExecResult struct {
	Stdout   string `json:"stdout"    doc:"What the command wrote to standard output"`
	Stderr   string `json:"stderr"    doc:"What the command wrote to standard error"`
	ExitCode int    `json:"exit_code" doc:"The command's exit code"`
	Changed  bool   `json:"changed"   doc:"Always true: osapi cannot tell what the command changed"`
}
