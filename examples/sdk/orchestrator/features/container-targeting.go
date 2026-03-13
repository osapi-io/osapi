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

// Package main demonstrates orchestrating container lifecycle operations
// through the standard OSAPI SDK client. The plan composes pull, create,
// exec, inspect, and cleanup as a DAG of TaskFunc steps.
//
// Expected output statuses:
//
//	changed   — pull, create, exec, cleanup
//	unchanged — inspect (read-only)
//	failed    — deliberately-fails (returns an error)
//
// Prerequisites:
//   - A running OSAPI stack (API server + agent + NATS)
//   - Docker available on the agent host
//
// Run with: OSAPI_TOKEN="<jwt>" go run container-targeting.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/retr0h/osapi/pkg/sdk/client"
	"github.com/retr0h/osapi/pkg/sdk/client/gen"
	"github.com/retr0h/osapi/pkg/sdk/orchestrator"
)

const (
	containerName  = "example-orchestrator-container"
	containerImage = "ubuntu:24.04"
)

func ptr(s string) *string { return &s }

func main() {
	url := os.Getenv("OSAPI_URL")
	if url == "" {
		url = "http://localhost:8080"
	}

	token := os.Getenv("OSAPI_TOKEN")
	if token == "" {
		log.Fatal("OSAPI_TOKEN is required")
	}

	target := os.Getenv("OSAPI_TARGET")
	if target == "" {
		target = "_any"
	}

	apiClient := client.New(url, token)

	// ── Plan setup ────────────────────────────────────────────────
	//
	// OnError(Continue) keeps independent tasks running after a
	// failure so the report shows all statuses: changed, unchanged,
	// failed. The AfterTask hook prints each result as it completes.

	hooks := orchestrator.Hooks{
		AfterTask: func(_ *orchestrator.Task, result orchestrator.TaskResult) {
			status := fmt.Sprintf("[%s]", result.Status)
			if result.Error != nil {
				status += " " + result.Error.Error()
			}
			fmt.Printf("  %-12s %-25s changed=%v\n",
				status, result.Name, result.Changed)
		},
	}

	plan := orchestrator.NewPlan(apiClient,
		orchestrator.WithHooks(hooks),
		orchestrator.OnError(orchestrator.Continue),
	)

	// ── Pull image ───────────────────────────────────────────────

	pull := plan.TaskFunc(
		"pull-image",
		func(
			ctx context.Context,
			c *client.Client,
		) (*orchestrator.Result, error) {
			resp, err := c.Container.Pull(ctx, target, gen.ContainerPullRequest{
				Image: containerImage,
			})
			if err != nil {
				return nil, fmt.Errorf("pull: %w", err)
			}

			r := resp.Data.Results[0]

			return &orchestrator.Result{
				Changed: true,
				Data:    map[string]any{"image_id": r.ImageID},
			}, nil
		},
	)

	// ── Create container ─────────────────────────────────────────

	autoStart := true
	create := plan.TaskFunc(
		"create-container",
		func(
			ctx context.Context,
			c *client.Client,
		) (*orchestrator.Result, error) {
			resp, err := c.Container.Create(ctx, target, gen.ContainerCreateRequest{
				Image:     containerImage,
				Name:      ptr(containerName),
				AutoStart: &autoStart,
				Command:   &[]string{"sleep", "600"},
			})
			if err != nil {
				return nil, fmt.Errorf("create: %w", err)
			}

			r := resp.Data.Results[0]

			return &orchestrator.Result{
				Changed: true,
				Data:    map[string]any{"id": r.ID, "name": r.Name},
			}, nil
		},
	)
	create.DependsOn(pull)

	// ── Exec: run commands inside the container ──────────────────

	execHostname := plan.TaskFunc(
		"exec-hostname",
		func(
			ctx context.Context,
			c *client.Client,
		) (*orchestrator.Result, error) {
			resp, err := c.Container.Exec(
				ctx,
				target,
				containerName,
				gen.ContainerExecRequest{
					Command: []string{"hostname"},
				},
			)
			if err != nil {
				return nil, fmt.Errorf("exec: %w", err)
			}

			r := resp.Data.Results[0]
			fmt.Printf("    hostname = %s", r.Stdout)

			return &orchestrator.Result{
				Changed: true,
				Data: map[string]any{
					"stdout":    r.Stdout,
					"exit_code": r.ExitCode,
				},
			}, nil
		},
	)
	execHostname.DependsOn(create)

	execUname := plan.TaskFunc(
		"exec-uname",
		func(
			ctx context.Context,
			c *client.Client,
		) (*orchestrator.Result, error) {
			resp, err := c.Container.Exec(
				ctx,
				target,
				containerName,
				gen.ContainerExecRequest{
					Command: []string{"uname", "-a"},
				},
			)
			if err != nil {
				return nil, fmt.Errorf("exec: %w", err)
			}

			r := resp.Data.Results[0]
			fmt.Printf("    uname -a = %s", r.Stdout)

			return &orchestrator.Result{
				Changed: true,
				Data:    map[string]any{"exit_code": r.ExitCode},
			}, nil
		},
	)
	execUname.DependsOn(create)

	execOSRelease := plan.TaskFunc(
		"exec-os-release",
		func(
			ctx context.Context,
			c *client.Client,
		) (*orchestrator.Result, error) {
			resp, err := c.Container.Exec(
				ctx,
				target,
				containerName,
				gen.ContainerExecRequest{
					Command: []string{"sh", "-c", "head -2 /etc/os-release"},
				},
			)
			if err != nil {
				return nil, fmt.Errorf("exec: %w", err)
			}

			r := resp.Data.Results[0]
			fmt.Printf("    os-release =\n%s", r.Stdout)

			return &orchestrator.Result{
				Changed: true,
				Data:    map[string]any{"exit_code": r.ExitCode},
			}, nil
		},
	)
	execOSRelease.DependsOn(create)

	// ── Inspect: read-only, reports unchanged ────────────────────

	inspect := plan.TaskFunc(
		"inspect-container",
		func(
			ctx context.Context,
			c *client.Client,
		) (*orchestrator.Result, error) {
			resp, err := c.Container.Inspect(ctx, target, containerName)
			if err != nil {
				return nil, fmt.Errorf("inspect: %w", err)
			}

			r := resp.Data.Results[0]
			fmt.Printf("    state = %s  image = %s\n", r.State, r.Image)

			return &orchestrator.Result{
				Data: map[string]any{
					"state": r.State,
					"image": r.Image,
				},
			}, nil
		},
	)
	inspect.DependsOn(create)

	// ── Deliberately failing task: shows StatusFailed ─────────────
	//
	// Returning an error from the task function marks it as failed.
	// With OnError(Continue), independent tasks keep running but
	// any task that DependsOn this one would be skipped.

	deliberatelyFails := plan.TaskFunc(
		"deliberately-fails",
		func(
			_ context.Context,
			_ *client.Client,
		) (*orchestrator.Result, error) {
			return nil, fmt.Errorf("this task always fails to demonstrate error reporting")
		},
	)
	deliberatelyFails.DependsOn(create)

	// ── Cleanup ──────────────────────────────────────────────────
	//
	// Depends on all operational tasks EXCEPT deliberately-fails so
	// cleanup is not skipped when OnError(Continue) is active.

	cleanup := plan.TaskFunc(
		"cleanup",
		func(
			ctx context.Context,
			c *client.Client,
		) (*orchestrator.Result, error) {
			force := true
			_, err := c.Container.Remove(
				ctx,
				target,
				containerName,
				&gen.DeleteNodeContainerByIDParams{Force: &force},
			)
			if err != nil {
				return nil, fmt.Errorf("remove: %w", err)
			}

			return &orchestrator.Result{Changed: true}, nil
		},
	)
	cleanup.DependsOn(execHostname, execUname, execOSRelease, inspect)

	// Suppress unused variable warning — deliberately-fails has no
	// dependents by design.
	_ = deliberatelyFails

	// ── Run ──────────────────────────────────────────────────────

	fmt.Println("=== Container Orchestration Example ===")
	fmt.Println()

	report, err := plan.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf(
		"\n=== Summary: %s in %s ===\n",
		report.Summary(),
		report.Duration.Truncate(time.Millisecond),
	)
}
