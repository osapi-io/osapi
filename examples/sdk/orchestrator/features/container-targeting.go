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

	pull := plan.DockerPull("pull-image", target, containerImage)

	// ── Create container ─────────────────────────────────────────

	autoStart := true
	create := plan.DockerCreate("create-container", target,
		gen.DockerCreateRequest{
			Image:     containerImage,
			Name:      ptr(containerName),
			AutoStart: &autoStart,
			Command:   &[]string{"sleep", "600"},
		},
	)
	create.DependsOn(pull)

	// ── Exec: run commands inside the container ──────────────────

	plan.DockerExec("exec-hostname", target, containerName,
		gen.DockerExecRequest{Command: []string{"hostname"}},
	).DependsOn(create)

	plan.DockerExec("exec-uname", target, containerName,
		gen.DockerExecRequest{Command: []string{"uname", "-a"}},
	).DependsOn(create)

	plan.DockerExec("exec-os-release", target, containerName,
		gen.DockerExecRequest{
			Command: []string{"sh", "-c", "head -2 /etc/os-release"},
		},
	).DependsOn(create)

	// ── Inspect: read-only, reports unchanged ────────────────────

	plan.DockerInspect("inspect-container", target, containerName).
		DependsOn(create)

	// ── Deliberately failing task: shows StatusFailed ─────────────

	plan.TaskFunc("deliberately-fails",
		func(
			_ context.Context,
			_ *client.Client,
		) (*orchestrator.Result, error) {
			return nil, fmt.Errorf("this task always fails to demonstrate error reporting")
		},
	).DependsOn(create)

	// ── Cleanup ──────────────────────────────────────────────────

	force := true
	plan.DockerRemove("cleanup", target, containerName,
		&gen.DeleteNodeContainerDockerByIDParams{Force: &force},
	).DependsOn(create)

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
