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

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	gohaidocker "github.com/osapi-io/gohai/pkg/gohai/collectors/docker"
	"github.com/osapi-io/gohai/pkg/gohai/source"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/secret"
)

// cli reads the containers through gohai's docker collector, and changes
// them with the docker CLI. It is the same on each platform: docker answers
// the same way on each.
type cli struct {
	src     source.Sources
	exec    host.Exec
	enabled bool
}

// newCLI is the provider on each platform.
func newCLI(
	h host.Host,
) (Provider, error) {
	return &cli{src: h.Sources(), exec: h.Exec, enabled: h.CommandEnabled}, nil
}

// containers are the containers that gohai reads. A host whose docker does
// not answer has no containers to read, which is skipped, not failed.
func (c *cli) containers(
	ctx context.Context,
) ([]gohaidocker.Container, error) {
	out, err := gohaidocker.NewWith(c.src).Collect(ctx, nil)
	info, ok := out.(*gohaidocker.Info)
	if err != nil || !ok || info == nil {
		return nil, fmt.Errorf("docker does not answer on this host: %w",
			errors.Join(behavior.ErrUnsupported, err))
	}

	return info.Containers, nil
}

// List reads the containers, of state: running, stopped, or all when empty.
func (c *cli) List(
	ctx context.Context,
	state types.ContainerState,
) ([]Entry, error) {
	all, err := c.containers(ctx)
	if err != nil {
		return nil, err
	}
	list := []Entry{}
	for _, g := range all {
		running := g.State == "running"
		if (state == "running" && !running) || (state == "stopped" && running) {
			continue
		}
		list = append(list, entry(g))
	}
	slices.SortFunc(list, func(a, b Entry) int { return strings.Compare(a.Name, b.Name) })

	return list, nil
}

// Get reads one container, by its name, its ID, or the start of its ID.
func (c *cli) Get(
	ctx context.Context,
	name string,
) (Entry, error) {
	all, err := c.containers(ctx)
	if err != nil {
		return Entry{}, err
	}
	// An exact name or ID wins over the start of an ID: a container named
	// abc is abc, also when another container's ID starts with abc.
	for _, g := range all {
		if g.Name == name || g.ID == name {
			return entry(g), nil
		}
	}
	var found []gohaidocker.Container
	for _, g := range all {
		if strings.HasPrefix(g.ID, name) {
			found = append(found, g)
		}
	}
	switch len(found) {
	case 0:
		return Entry{}, fmt.Errorf("container %q: %w", name, behavior.ErrNotFound)
	case 1:
		return entry(found[0]), nil
	default:
		return Entry{}, fmt.Errorf("container %q is the start of the ID of %d containers: "+
			"give more of the ID, or the name", name, len(found))
	}
}

// byName is the container of the name, and only of the name: create makes a
// container of its name, so another container's ID that starts with it is
// not that container.
func (c *cli) byName(
	ctx context.Context,
	name string,
) (Entry, error) {
	all, err := c.containers(ctx)
	if err != nil {
		return Entry{}, err
	}
	for _, g := range all {
		if g.Name == name {
			return entry(g), nil
		}
	}

	return Entry{}, fmt.Errorf("container %q: %w", name, behavior.ErrNotFound)
}

// Create makes the container, and starts it when in says so. A container of
// the name that the host has already is not made again.
func (c *cli) Create(
	ctx context.Context,
	in CreateInput,
) (Result, error) {
	// A command or a bind mount makes the container a way to run a command
	// on the host, so it is refused where commands are.
	if !c.enabled && (len(in.Command) > 0 || len(in.Volumes) > 0) {
		return Result{}, host.ErrCommandDisabled
	}
	existing, err := c.byName(ctx, in.Name)
	switch {
	case err == nil:
		started := false
		if in.AutoStart && existing.State != "running" {
			if started, err = c.Start(ctx, in.Name); err != nil {
				return Result{}, err
			}
		}

		return Result{Name: in.Name, ID: existing.ID, Changed: started}, nil
	case !errors.Is(err, behavior.ErrNotFound):
		return Result{}, err
	}

	args := []string{"create", "--name", in.Name}
	if in.HostName != "" {
		args = append(args, "--hostname", in.HostName)
	}
	for _, d := range in.DNS {
		args = append(args, "--dns", d)
	}
	for _, p := range in.Ports {
		args = append(args, "-p", p)
	}
	for _, v := range in.Volumes {
		args = append(args, "-v", v)
	}
	var out string
	if len(in.Env) > 0 {
		// The values go on standard input, not in the arguments, which each
		// user of the host can read.
		file, fileErr := envFile(in.Env)
		if fileErr != nil {
			return Result{}, fileErr
		}
		args = append(args, "--env-file", "/dev/stdin", in.Image)
		args = append(args, in.Command...)
		out, err = c.exec.RunPrivilegedCmdWithStdin(ctx, "docker", args, file)
	} else {
		args = append(args, in.Image)
		args = append(args, in.Command...)
		out, err = c.exec.RunPrivilegedCmd(ctx, "docker", args)
	}
	if err != nil {
		return Result{}, fmt.Errorf(
			"docker create %s: %w: %s",
			in.Name,
			err,
			strings.TrimSpace(out),
		)
	}
	id := strings.TrimSpace(out)
	if in.AutoStart {
		if _, err := c.exec.RunPrivilegedCmd(
			ctx,
			"docker",
			[]string{"start", in.Name},
		); err != nil {
			return Result{}, fmt.Errorf("docker start %s: %w", in.Name, err)
		}
	}

	return Result{Name: in.Name, ID: id, Changed: true}, nil
}

// Delete removes the container, also one that runs, as docker rm -f does.
// One that the host does not have does not change it.
func (c *cli) Delete(
	ctx context.Context,
	name string,
) (bool, error) {
	if _, err := c.Get(ctx, name); errors.Is(err, behavior.ErrNotFound) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	if out, err := c.exec.RunPrivilegedCmd(ctx, "docker", []string{"rm", "-f", name}); err != nil {
		return false, fmt.Errorf("docker rm %s: %w: %s", name, err, strings.TrimSpace(out))
	}

	return true, nil
}

// Start starts the container. One that runs already does not change.
func (c *cli) Start(
	ctx context.Context,
	name string,
) (bool, error) {
	e, err := c.Get(ctx, name)
	if err != nil {
		return false, err
	}
	if e.State == "running" {
		return false, nil
	}
	if out, err := c.exec.RunPrivilegedCmd(ctx, "docker", []string{"start", name}); err != nil {
		return false, fmt.Errorf("docker start %s: %w: %s", name, err, strings.TrimSpace(out))
	}

	return true, nil
}

// Stop stops the container, waiting timeout seconds before docker kills
// it, or docker's default when 0. One that does not run does not change.
func (c *cli) Stop(
	ctx context.Context,
	name string,
	timeout int,
) (bool, error) {
	e, err := c.Get(ctx, name)
	if err != nil {
		return false, err
	}
	if e.State != "running" {
		return false, nil
	}
	args := []string{"stop"}
	if timeout > 0 {
		args = append(args, "-t", strconv.Itoa(timeout))
	}
	if out, err := c.exec.RunPrivilegedCmd(ctx, "docker", append(args, name)); err != nil {
		return false, fmt.Errorf("docker stop %s: %w: %s", name, err, strings.TrimSpace(out))
	}

	return true, nil
}

// Exec runs the command in the container, and gives what it wrote and its
// exit code. A command that fails is an answer, not an error.
func (c *cli) Exec(
	ctx context.Context,
	in ExecInput,
) (ExecResult, error) {
	if !c.enabled {
		return ExecResult{}, host.ErrCommandDisabled
	}
	args := []string{"exec"}
	stdin := ""
	if len(in.Env) > 0 {
		// The values go on standard input, not in the arguments, which each
		// user of the host can read. Without -i, docker exec gives the
		// container no standard input, so docker reads all of it.
		args = append(args, "--env-file", "/dev/stdin")
		file, err := envFile(in.Env)
		if err != nil {
			return ExecResult{}, err
		}
		stdin = file
	}
	if in.WorkingDir != "" {
		args = append(args, "-w", in.WorkingDir)
	}
	args = append(append(args, in.Name), in.Command...)
	stdout, stderr, code, err := c.exec.RunPrivilegedCmdFull(ctx, "docker", args, stdin)
	if err != nil {
		return ExecResult{}, fmt.Errorf("docker exec %s: %w", in.Name, err)
	}

	return ExecResult{Stdout: stdout, Stderr: stderr, ExitCode: code, Changed: true}, nil
}

// envFile is env as docker's --env-file reads it: one KEY=value a line,
// by name.
func envFile(
	env map[string]secret.Value,
) (string, error) {
	var b strings.Builder
	for _, k := range slices.Sorted(maps.Keys(env)) {
		// docker reads one KEY=value on each line: a key with = or a line
		// end, or a value with a line end, would make lines of its own.
		if !envKey.MatchString(k) {
			return "", fmt.Errorf("environment variable %q: a name is letters, digits "+
				"and _, and does not start with a digit", k)
		}
		v := env[k].Reveal()
		if strings.ContainsAny(v, "\r\n") {
			return "", fmt.Errorf("environment variable %q: a value has no line end", k)
		}
		b.WriteString(k + "=" + v + "\n")
	}

	return b.String(), nil
}

// envKey is the name of an environment variable that docker's env file
// takes as one.
var envKey = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// entry is a container as list and get answer it.
func entry(
	g gohaidocker.Container,
) Entry {
	e := Entry{
		ID: g.ID, Name: g.Name, Image: g.Image, State: g.State, Summary: g.Status,
		Created: g.Created, Health: g.Health, EnvNames: g.EnvNames,
	}
	for _, p := range g.Ports {
		if p.HostPort == "" {
			e.Ports = append(e.Ports, p.ContainerPort)

			continue
		}
		e.Ports = append(e.Ports, p.HostIP+":"+p.HostPort+"->"+p.ContainerPort)
	}
	for _, m := range g.Mounts {
		s := m.Source + ":" + m.Destination
		if m.ReadOnly {
			s += ":ro"
		}
		e.Mounts = append(e.Mounts, s)
	}
	for _, n := range g.Networks {
		e.Networks = append(e.Networks, n.Name+" "+n.IPAddress+" via "+n.Gateway)
	}

	return e
}
