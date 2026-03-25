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

// Package main demonstrates container management: pull an image, create
// a container, list, inspect, exec, stop, and remove.
//
// Run with: OSAPI_TOKEN="<jwt>" go run container.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/retr0h/osapi/pkg/sdk/client"
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

	c := client.New(url, token)
	ctx := context.Background()
	target := "_all"

	// Pull an image.
	pull, err := c.Docker.Pull(ctx, target, client.DockerPullOpts{
		Image: "nginx:alpine",
	})
	if err != nil {
		log.Fatalf("pull: %v", err)
	}

	for _, r := range pull.Data.Results {
		fmt.Printf("Pull (%s): image=%s tag=%s size=%d\n",
			r.Hostname, r.ImageID, r.Tag, r.Size)
	}

	// Create a container.
	autoStart := true
	create, err := c.Docker.Create(ctx, target, client.DockerCreateOpts{
		Image:     "nginx:alpine",
		Name:      "osapi-example",
		AutoStart: &autoStart,
	})
	if err != nil {
		log.Fatalf("create: %v", err)
	}

	var containerID string
	for _, r := range create.Data.Results {
		containerID = r.ID
		fmt.Printf("Create (%s): id=%s name=%s state=%s\n",
			r.Hostname, r.ID, r.Name, r.State)
	}

	// List running containers.
	list, err := c.Docker.List(ctx, target, &client.DockerListParams{
		State: "running",
	})
	if err != nil {
		log.Fatalf("list: %v", err)
	}

	for _, r := range list.Data.Results {
		fmt.Printf("\nContainers (%s):\n", r.Hostname)
		for _, ct := range r.Containers {
			fmt.Printf("  %s  %s  %s\n", ct.ID[:12], ct.Image, ct.State)
		}
	}

	// Inspect the container.
	inspect, err := c.Docker.Inspect(ctx, target, containerID)
	if err != nil {
		log.Fatalf("inspect: %v", err)
	}

	for _, r := range inspect.Data.Results {
		fmt.Printf("\nInspect (%s): id=%s image=%s state=%s\n",
			r.Hostname, r.ID, r.Image, r.State)
	}

	// Exec a command inside the container.
	exec, err := c.Docker.Exec(ctx, target, containerID, client.DockerExecOpts{
		Command: []string{"cat", "/etc/hostname"},
	})
	if err != nil {
		log.Fatalf("exec: %v", err)
	}

	for _, r := range exec.Data.Results {
		fmt.Printf("Exec (%s): stdout=%s exit=%d\n",
			r.Hostname, r.Stdout, r.ExitCode)
	}

	// Stop the container.
	stop, err := c.Docker.Stop(ctx, target, containerID, client.DockerStopOpts{
		Timeout: 5,
	})
	if err != nil {
		log.Fatalf("stop: %v", err)
	}

	for _, r := range stop.Data.Results {
		fmt.Printf("Stop (%s): id=%s message=%s\n",
			r.Hostname, r.ID, r.Message)
	}

	// Remove the container.
	remove, err := c.Docker.Remove(
		ctx,
		target,
		containerID,
		&client.DockerRemoveParams{Force: true},
	)
	if err != nil {
		log.Fatalf("remove: %v", err)
	}

	for _, r := range remove.Data.Results {
		fmt.Printf("Remove (%s): id=%s message=%s\n",
			r.Hostname, r.ID, r.Message)
	}

	// Remove the image.
	imgRemove, err := c.Docker.ImageRemove(
		ctx,
		target,
		"nginx:alpine",
		&client.DockerImageRemoveParams{Force: true},
	)
	if err != nil {
		log.Fatalf("image remove: %v", err)
	}

	for _, r := range imgRemove.Data.Results {
		fmt.Printf("ImageRemove (%s): id=%s message=%s\n",
			r.Hostname, r.ID, r.Message)
	}
}
