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

// Package main demonstrates running provider operations inside Docker
// containers using the orchestrator SDK's typed ContainerProvider.
//
// The example exercises every read-only provider, command execution,
// error handling for failing commands, and a deliberately failing task
// to show how the orchestrator reports each status.
//
// Expected output statuses:
//
//	changed   — setup tasks (ensure-clean, pull, create, deploy) and command exec
//	unchanged — all read-only providers (host, mem, load, disk)
//	failed    — deliberately-fails task (returns an error)
//	skipped   — none (OnError(Continue) lets independent tasks proceed)
//
// Prerequisites:
//   - A running OSAPI stack (API server + agent + NATS)
//   - Docker available on the agent host
//   - Go toolchain (the example builds osapi for linux automatically)
//
// Run with: OSAPI_TOKEN="<jwt>" go run container-targeting.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	osexec "os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/retr0h/osapi/pkg/sdk/client"
	"github.com/retr0h/osapi/pkg/sdk/client/gen"
	"github.com/retr0h/osapi/pkg/sdk/orchestrator"
)

const containerName = "example-provider-container"
const containerImage = "ubuntu:24.04"

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

	// WithDockerExecFn bridges the orchestrator's Docker DSL to the
	// OSAPI Container.Exec REST API. Every ExecProvider call flows
	// through: SDK → REST API → NATS → agent → docker exec.
	dockerExecFn := func(
		ctx context.Context,
		containerID string,
		command []string,
	) (stdout, stderr string, exitCode int, err error) {
		resp, execErr := apiClient.Container.Exec(
			ctx,
			target,
			containerID,
			gen.ContainerExecRequest{Command: command},
		)
		if execErr != nil {
			return "", "", -1, execErr
		}

		r := resp.Data.Results[0]

		return r.Stdout, r.Stderr, r.ExitCode, nil
	}

	plan := orchestrator.NewPlan(apiClient,
		orchestrator.WithHooks(hooks),
		orchestrator.WithDockerExecFn(dockerExecFn),
		orchestrator.OnError(orchestrator.Continue),
	)

	// ── Setup: pull + create + deploy osapi ──────────────────────

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
				// Container already exists — carry on.
				return &orchestrator.Result{}, nil
			}

			r := resp.Data.Results[0]

			return &orchestrator.Result{
				Changed: true,
				Data:    map[string]any{"id": r.ID, "name": r.Name},
			}, nil
		},
	)
	create.DependsOn(pull)

	// Build osapi for linux and docker cp it into the container.
	// In production you would bake the binary into the image.
	deploy := plan.TaskFunc(
		"deploy-osapi",
		func(
			_ context.Context,
			_ *client.Client,
		) (*orchestrator.Result, error) {
			tmpBin := "/tmp/osapi-container"

			build := osexec.Command(
				"go", "build", "-o", tmpBin, "github.com/retr0h/osapi",
			)
			build.Dir = findProjectRoot()
			build.Env = append(os.Environ(), "GOOS=linux", "GOARCH="+runtime.GOARCH)
			if out, err := build.CombinedOutput(); err != nil {
				return nil, fmt.Errorf("build osapi: %s: %w", string(out), err)
			}

			cp := osexec.Command("docker", "cp", tmpBin, containerName+":/osapi")
			if out, err := cp.CombinedOutput(); err != nil {
				return nil, fmt.Errorf("docker cp: %s: %w", string(out), err)
			}

			_ = os.Remove(tmpBin)

			return &orchestrator.Result{Changed: true}, nil
		},
	)
	deploy.DependsOn(create)

	// ── Provider: typed ContainerProvider bound to the target ──────

	dockerTarget := plan.Docker(containerName, containerImage)
	provider := orchestrator.NewContainerProvider(dockerTarget)
	scoped := plan.In(dockerTarget)

	// ── Host provider: 9 read-only operations (unchanged) ─────────

	getHostname := scoped.TaskFunc(
		"host/get-hostname",
		func(ctx context.Context, _ *client.Client) (*orchestrator.Result, error) {
			v, err := provider.GetHostname(ctx)
			if err != nil {
				return nil, err
			}
			fmt.Printf("    hostname      = %s\n", v)

			return &orchestrator.Result{Data: map[string]any{"hostname": v}}, nil
		},
	)
	getHostname.DependsOn(deploy)

	getOSInfo := scoped.TaskFunc(
		"host/get-os-info",
		func(ctx context.Context, _ *client.Client) (*orchestrator.Result, error) {
			info, err := provider.GetOSInfo(ctx)
			if err != nil {
				return nil, err
			}
			fmt.Printf("    os            = %s %s\n", info.Distribution, info.Version)

			return &orchestrator.Result{
				Changed: info.Changed,
				Data: map[string]any{
					"distribution": info.Distribution,
					"version":      info.Version,
				},
			}, nil
		},
	)
	getOSInfo.DependsOn(deploy)

	getArch := scoped.TaskFunc(
		"host/get-architecture",
		func(ctx context.Context, _ *client.Client) (*orchestrator.Result, error) {
			v, err := provider.GetArchitecture(ctx)
			if err != nil {
				return nil, err
			}
			fmt.Printf("    architecture  = %s\n", v)

			return &orchestrator.Result{Data: map[string]any{"architecture": v}}, nil
		},
	)
	getArch.DependsOn(deploy)

	getKernel := scoped.TaskFunc(
		"host/get-kernel-version",
		func(ctx context.Context, _ *client.Client) (*orchestrator.Result, error) {
			v, err := provider.GetKernelVersion(ctx)
			if err != nil {
				return nil, err
			}
			fmt.Printf("    kernel        = %s\n", v)

			return &orchestrator.Result{Data: map[string]any{"kernel": v}}, nil
		},
	)
	getKernel.DependsOn(deploy)

	getUptime := scoped.TaskFunc(
		"host/get-uptime",
		func(ctx context.Context, _ *client.Client) (*orchestrator.Result, error) {
			v, err := provider.GetUptime(ctx)
			if err != nil {
				return nil, err
			}
			fmt.Printf("    uptime        = %s\n", v)

			return &orchestrator.Result{Data: map[string]any{"uptime": v.String()}}, nil
		},
	)
	getUptime.DependsOn(deploy)

	getFQDN := scoped.TaskFunc(
		"host/get-fqdn",
		func(ctx context.Context, _ *client.Client) (*orchestrator.Result, error) {
			v, err := provider.GetFQDN(ctx)
			if err != nil {
				return nil, err
			}
			fmt.Printf("    fqdn          = %s\n", v)

			return &orchestrator.Result{Data: map[string]any{"fqdn": v}}, nil
		},
	)
	getFQDN.DependsOn(deploy)

	getCPUCount := scoped.TaskFunc(
		"host/get-cpu-count",
		func(ctx context.Context, _ *client.Client) (*orchestrator.Result, error) {
			v, err := provider.GetCPUCount(ctx)
			if err != nil {
				return nil, err
			}
			fmt.Printf("    cpu_count     = %d\n", v)

			return &orchestrator.Result{Data: map[string]any{"cpu_count": v}}, nil
		},
	)
	getCPUCount.DependsOn(deploy)

	getSvcMgr := scoped.TaskFunc(
		"host/get-service-manager",
		func(ctx context.Context, _ *client.Client) (*orchestrator.Result, error) {
			v, err := provider.GetServiceManager(ctx)
			if err != nil {
				return nil, err
			}
			fmt.Printf("    service_mgr   = %s\n", v)

			return &orchestrator.Result{Data: map[string]any{"service_manager": v}}, nil
		},
	)
	getSvcMgr.DependsOn(deploy)

	getPkgMgr := scoped.TaskFunc(
		"host/get-package-manager",
		func(ctx context.Context, _ *client.Client) (*orchestrator.Result, error) {
			v, err := provider.GetPackageManager(ctx)
			if err != nil {
				return nil, err
			}
			fmt.Printf("    package_mgr   = %s\n", v)

			return &orchestrator.Result{Data: map[string]any{"package_manager": v}}, nil
		},
	)
	getPkgMgr.DependsOn(deploy)

	// ── Memory, load, disk providers (unchanged) ──────────────────

	getMemStats := scoped.TaskFunc(
		"mem/get-stats",
		func(ctx context.Context, _ *client.Client) (*orchestrator.Result, error) {
			stats, err := provider.GetMemStats(ctx)
			if err != nil {
				return nil, err
			}
			fmt.Printf("    mem_total     = %d MB\n", stats.Total/1024/1024)
			fmt.Printf("    mem_available = %d MB\n", stats.Available/1024/1024)

			return &orchestrator.Result{
				Changed: stats.Changed,
				Data: map[string]any{
					"total":     stats.Total,
					"available": stats.Available,
				},
			}, nil
		},
	)
	getMemStats.DependsOn(deploy)

	getLoadStats := scoped.TaskFunc(
		"load/get-average-stats",
		func(ctx context.Context, _ *client.Client) (*orchestrator.Result, error) {
			stats, err := provider.GetLoadStats(ctx)
			if err != nil {
				return nil, err
			}
			fmt.Printf("    load1         = %.2f\n", stats.Load1)
			fmt.Printf("    load5         = %.2f\n", stats.Load5)
			fmt.Printf("    load15        = %.2f\n", stats.Load15)

			return &orchestrator.Result{
				Changed: stats.Changed,
				Data: map[string]any{
					"load1": stats.Load1, "load5": stats.Load5, "load15": stats.Load15,
				},
			}, nil
		},
	)
	getLoadStats.DependsOn(deploy)

	getDiskUsage := scoped.TaskFunc(
		"disk/get-usage",
		func(ctx context.Context, _ *client.Client) (*orchestrator.Result, error) {
			disks, err := provider.GetDiskUsage(ctx)
			if err != nil {
				return nil, err
			}
			for _, d := range disks {
				fmt.Printf("    disk %-8s  total=%d MB  used=%d MB\n",
					d.Name, d.Total/1024/1024, d.Used/1024/1024)
			}

			return &orchestrator.Result{Data: map[string]any{"mounts": len(disks)}}, nil
		},
	)
	getDiskUsage.DependsOn(deploy)

	// ── Command provider: exec + shell (changed) ──────────────────

	execUname := scoped.TaskFunc(
		"command/exec-uname",
		func(ctx context.Context, _ *client.Client) (*orchestrator.Result, error) {
			r, err := provider.Exec(ctx, orchestrator.ExecParams{
				Command: "uname",
				Args:    []string{"-a"},
			})
			if err != nil {
				return nil, err
			}
			fmt.Printf("    uname -a      = %s", r.Stdout)

			return &orchestrator.Result{
				Changed: r.Changed,
				Data:    map[string]any{"exit_code": r.ExitCode},
			}, nil
		},
	)
	execUname.DependsOn(deploy)

	// Reads /etc/os-release inside the container to prove we are
	// running inside Ubuntu 24.04, not the host OS.
	shellOSRelease := scoped.TaskFunc(
		"command/shell-os-release",
		func(ctx context.Context, _ *client.Client) (*orchestrator.Result, error) {
			r, err := provider.Shell(ctx, orchestrator.ShellParams{
				Command: "head -2 /etc/os-release && echo container-hostname=$(hostname)",
			})
			if err != nil {
				return nil, err
			}
			fmt.Printf("    os-release    =\n%s", r.Stdout)

			return &orchestrator.Result{
				Changed: r.Changed,
				Data:    map[string]any{"exit_code": r.ExitCode},
			}, nil
		},
	)
	shellOSRelease.DependsOn(deploy)

	// ── Command that exits non-zero: handled gracefully ───────────
	//
	// The command provider returns the exit code in the result.
	// The task inspects it and reports unchanged (no system mutation)
	// rather than failing the task.

	execFails := scoped.TaskFunc(
		"command/exec-nonzero",
		func(ctx context.Context, _ *client.Client) (*orchestrator.Result, error) {
			r, err := provider.Exec(ctx, orchestrator.ExecParams{
				Command: "ls",
				Args:    []string{"/does-not-exist"},
			})
			if err != nil {
				return nil, err
			}

			fmt.Printf("    exit_code     = %d\n", r.ExitCode)
			fmt.Printf("    stderr        = %s", r.Stderr)

			return &orchestrator.Result{
				Changed: r.Changed,
				Data: map[string]any{
					"exit_code": r.ExitCode,
					"stderr":    r.Stderr,
				},
			}, nil
		},
	)
	execFails.DependsOn(deploy)

	// ── Deliberately failing task: shows StatusFailed ──────────────
	//
	// Returning an error from the task function marks it as failed.
	// With OnError(Continue), independent tasks keep running but
	// any task that DependsOn this one would be skipped.

	deliberatelyFails := scoped.TaskFunc(
		"deliberately-fails",
		func(
			_ context.Context,
			_ *client.Client,
		) (*orchestrator.Result, error) {
			return nil, fmt.Errorf("this task always fails to demonstrate error reporting")
		},
	)
	deliberatelyFails.DependsOn(deploy)

	// ── Cleanup ───────────────────────────────────────────────────
	//
	// Depends on all provider tasks EXCEPT deliberately-fails so
	// that cleanup is not skipped when OnError(Continue) is active.

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
	cleanup.DependsOn(
		getHostname, getOSInfo, getArch, getKernel, getUptime,
		getFQDN, getCPUCount, getSvcMgr, getPkgMgr,
		getMemStats, getLoadStats, getDiskUsage,
		execUname, shellOSRelease, execFails,
	)

	// Suppress unused variable warning — deliberately-fails has no
	// dependents by design.
	_ = deliberatelyFails

	// ── Run ───────────────────────────────────────────────────────

	fmt.Println("=== Container Provider Example ===")
	fmt.Println()

	report, err := plan.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n=== Summary: %s in %s ===\n", report.Summary(), report.Duration.Truncate(time.Millisecond))
}

// findProjectRoot walks up from the current directory to find go.mod.
func findProjectRoot() string {
	dir, _ := os.Getwd()

	for {
		if _, err := os.Stat(dir + "/go.mod"); err == nil {
			return dir
		}

		idx := strings.LastIndex(dir, "/")
		if idx <= 0 {
			return "."
		}

		dir = dir[:idx]
	}
}
