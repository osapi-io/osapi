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

// Package main demonstrates orchestrating Docker container lifecycle
// operations using the DSL helpers. The plan composes pull, create,
// exec, inspect, and cleanup as a DAG.
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
	"github.com/retr0h/osapi/pkg/sdk/orchestrator"
)

const (
	containerName  = "example-orchestrator-container"
	containerImage = "ubuntu:24.04"
)

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

	// ── Pre-cleanup: remove leftover container from previous run ─
	// Swallow errors — the container may not exist.

	preCleanup := plan.TaskFunc("pre-cleanup",
		func(
			ctx context.Context,
			c *client.Client,
		) (*orchestrator.Result, error) {
			_, _ = c.Docker.Remove(ctx, target, containerName,
				&client.DockerRemoveParams{Force: true},
			)

			return &orchestrator.Result{Changed: false}, nil
		},
	)

	// ── Pull image ───────────────────────────────────────────────

	pull := plan.TaskFunc("pull-image",
		func(
			ctx context.Context,
			c *client.Client,
		) (*orchestrator.Result, error) {
			resp, err := c.Docker.Pull(ctx, target, client.DockerPullOpts{
				Image: containerImage,
			})
			if err != nil {
				return nil, err
			}

			return orchestrator.CollectionResult(resp.Data,
				func(r client.DockerPullResult) orchestrator.HostResult {
					return orchestrator.HostResult{
						Hostname: r.Hostname,
						Changed:  r.Changed,
						Error:    r.Error,
					}
				},
			), nil
		},
	)
	pull.DependsOn(preCleanup)

	// ── Create container ─────────────────────────────────────────

	autoStart := true
	create := plan.TaskFunc("create-container",
		func(
			ctx context.Context,
			c *client.Client,
		) (*orchestrator.Result, error) {
			resp, err := c.Docker.Create(ctx, target, client.DockerCreateOpts{
				Image:     containerImage,
				Name:      containerName,
				AutoStart: &autoStart,
				Command:   []string{"sleep", "600"},
			})
			if err != nil {
				return nil, err
			}

			return orchestrator.CollectionResult(resp.Data,
				func(r client.DockerResult) orchestrator.HostResult {
					return orchestrator.HostResult{
						Hostname: r.Hostname,
						Changed:  r.Changed,
						Error:    r.Error,
					}
				},
			), nil
		},
	)
	create.DependsOn(pull)

	// ── Exec: run commands inside the container ──────────────────

	execHostname := plan.TaskFunc("exec-hostname",
		func(
			ctx context.Context,
			c *client.Client,
		) (*orchestrator.Result, error) {
			resp, err := c.Docker.Exec(ctx, target, containerName,
				client.DockerExecOpts{Command: []string{"hostname"}},
			)
			if err != nil {
				return nil, err
			}

			return orchestrator.CollectionResult(resp.Data,
				func(r client.DockerExecResult) orchestrator.HostResult {
					return orchestrator.HostResult{
						Hostname: r.Hostname,
						Changed:  r.Changed,
						Error:    r.Error,
					}
				},
			), nil
		},
	)
	execHostname.DependsOn(create)

	execUname := plan.TaskFunc("exec-uname",
		func(
			ctx context.Context,
			c *client.Client,
		) (*orchestrator.Result, error) {
			resp, err := c.Docker.Exec(ctx, target, containerName,
				client.DockerExecOpts{Command: []string{"uname", "-a"}},
			)
			if err != nil {
				return nil, err
			}

			return orchestrator.CollectionResult(resp.Data,
				func(r client.DockerExecResult) orchestrator.HostResult {
					return orchestrator.HostResult{
						Hostname: r.Hostname,
						Changed:  r.Changed,
						Error:    r.Error,
					}
				},
			), nil
		},
	)
	execUname.DependsOn(create)

	execOS := plan.TaskFunc("exec-os-release",
		func(
			ctx context.Context,
			c *client.Client,
		) (*orchestrator.Result, error) {
			resp, err := c.Docker.Exec(ctx, target, containerName,
				client.DockerExecOpts{
					Command: []string{"sh", "-c", "head -2 /etc/os-release"},
				},
			)
			if err != nil {
				return nil, err
			}

			return orchestrator.CollectionResult(resp.Data,
				func(r client.DockerExecResult) orchestrator.HostResult {
					return orchestrator.HostResult{
						Hostname: r.Hostname,
						Changed:  r.Changed,
						Error:    r.Error,
					}
				},
			), nil
		},
	)
	execOS.DependsOn(create)

	// ── Inspect: read-only, reports unchanged ────────────────────

	inspect := plan.TaskFunc("inspect-container",
		func(
			ctx context.Context,
			c *client.Client,
		) (*orchestrator.Result, error) {
			resp, err := c.Docker.Inspect(ctx, target, containerName)
			if err != nil {
				return nil, err
			}

			return orchestrator.CollectionResult(resp.Data,
				func(r client.DockerDetailResult) orchestrator.HostResult {
					return orchestrator.HostResult{
						Hostname: r.Hostname,
						Changed:  r.Changed,
						Error:    r.Error,
					}
				},
			), nil
		},
	)
	inspect.DependsOn(create)

	// ── Deliberately failing task: shows StatusFailed ─────────────

	deliberatelyFails := plan.TaskFunc("deliberately-fails",
		func(
			_ context.Context,
			_ *client.Client,
		) (*orchestrator.Result, error) {
			return nil, fmt.Errorf("this task always fails to demonstrate error reporting")
		},
	)
	deliberatelyFails.DependsOn(create)

	// ── Cleanup ──────────────────────────────────────────────────
	// Depends on all tasks that use the container so it runs last.

	plan.TaskFunc("cleanup",
		func(
			ctx context.Context,
			c *client.Client,
		) (*orchestrator.Result, error) {
			resp, err := c.Docker.Remove(ctx, target, containerName,
				&client.DockerRemoveParams{Force: true},
			)
			if err != nil {
				return nil, err
			}

			return orchestrator.CollectionResult(resp.Data,
				func(r client.DockerActionResult) orchestrator.HostResult {
					return orchestrator.HostResult{
						Hostname: r.Hostname,
						Changed:  r.Changed,
						Error:    r.Error,
					}
				},
			), nil
		},
	).DependsOn(execHostname, execUname, execOS, inspect)

	// ── Run ──────────────────────────────────────────────────────

	fmt.Println("=== Docker Orchestration Example ===")
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
