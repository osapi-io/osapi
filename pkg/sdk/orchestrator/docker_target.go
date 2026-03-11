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

package orchestrator

import (
	"context"
	"fmt"
)

// ExecFn executes a command inside a container and returns stdout/stderr/exit code.
type ExecFn func(
	ctx context.Context,
	containerID string,
	command []string,
) (stdout, stderr string, exitCode int, err error)

// DockerTarget implements RuntimeTarget for Docker containers.
type DockerTarget struct {
	name   string
	image  string
	execFn ExecFn
}

// NewDockerTarget creates a new Docker runtime target.
func NewDockerTarget(
	name string,
	image string,
	execFn ExecFn,
) *DockerTarget {
	return &DockerTarget{
		name:   name,
		image:  image,
		execFn: execFn,
	}
}

// Name returns the container name.
func (t *DockerTarget) Name() string {
	return t.name
}

// Runtime returns "docker".
func (t *DockerTarget) Runtime() string {
	return "docker"
}

// Image returns the container image.
func (t *DockerTarget) Image() string {
	return t.image
}

// ExecProvider runs a provider operation inside this container via docker exec.
func (t *DockerTarget) ExecProvider(
	ctx context.Context,
	provider string,
	operation string,
	data []byte,
) ([]byte, error) {
	cmd := []string{"/osapi", "provider", "run", provider, operation}
	if len(data) > 0 {
		cmd = append(cmd, "--data", string(data))
	}

	stdout, stderr, exitCode, err := t.execFn(ctx, t.name, cmd)
	if err != nil {
		return nil, fmt.Errorf("exec provider in container %s: %w", t.name, err)
	}

	if exitCode != 0 {
		return nil, fmt.Errorf(
			"provider %s/%s failed (exit %d): %s",
			provider,
			operation,
			exitCode,
			stderr,
		)
	}

	return []byte(stdout), nil
}
