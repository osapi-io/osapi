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

// Package main demonstrates container targeting for running provider
// operations inside Docker containers using the orchestrator DSL.
//
// DAG:
//
//	pull-image
//	    └── create-container
//	            └── exec-inside (scoped via In)
//	                    └── cleanup
//
// Run with: OSAPI_TOKEN="<jwt>" go run container-targeting.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/retr0h/osapi/pkg/sdk/client"
	"github.com/retr0h/osapi/pkg/sdk/client/gen"
	"github.com/retr0h/osapi/pkg/sdk/orchestrator"
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

	client := client.New(url, token)

	hooks := orchestrator.Hooks{
		AfterTask: func(_ *orchestrator.Task, result orchestrator.TaskResult) {
			fmt.Printf("  [%s] %s  changed=%v\n",
				result.Status, result.Name, result.Changed)
		},
	}

	// WithDockerExecFn is required for Plan.Docker() to work.
	// In a real application, this would use the Docker SDK's
	// ContainerExecCreate/ContainerExecAttach APIs.
	plan := orchestrator.NewPlan(client,
		orchestrator.WithHooks(hooks),
	)

	// Pull the image first.
	pull := plan.TaskFunc(
		"pull-image",
		func(
			ctx context.Context,
			c *client.Client,
		) (*orchestrator.Result, error) {
			resp, err := c.Container.Pull(ctx, target, gen.ContainerPullRequest{
				Image: "ubuntu:24.04",
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

	// Create the container.
	autoStart := true

	create := plan.TaskFunc(
		"create-container",
		func(
			ctx context.Context,
			c *client.Client,
		) (*orchestrator.Result, error) {
			resp, err := c.Container.Create(ctx, target, gen.ContainerCreateRequest{
				Image:     "ubuntu:24.04",
				Name:      ptr("example-container"),
				AutoStart: &autoStart,
				Command:   &[]string{"sleep", "300"},
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

	// Exec a command inside the container.
	execInside := plan.TaskFunc(
		"exec-inside",
		func(
			ctx context.Context,
			c *client.Client,
		) (*orchestrator.Result, error) {
			resp, err := c.Container.Exec(
				ctx,
				target,
				"example-container",
				gen.ContainerExecRequest{
					Command: []string{"cat", "/etc/os-release"},
				},
			)
			if err != nil {
				return nil, fmt.Errorf("exec: %w", err)
			}

			r := resp.Data.Results[0]
			fmt.Printf("\n  --- stdout ---\n%s\n", r.Stdout)

			return &orchestrator.Result{
				Changed: false,
				Data:    map[string]any{"exit_code": r.ExitCode},
			}, nil
		},
	)
	execInside.DependsOn(create)

	// Clean up: remove the container.
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
				"example-container",
				&gen.DeleteNodeContainerByIDParams{Force: &force},
			)
			if err != nil {
				return nil, fmt.Errorf("remove: %w", err)
			}

			return &orchestrator.Result{Changed: true}, nil
		},
	)
	cleanup.DependsOn(execInside)

	report, err := plan.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n%s in %s\n", report.Summary(), report.Duration)
}
