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
// a container, list, get, exec, stop, and remove.
//
// Run with: OSAPI_TOKEN="<jwt>" go run docker.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/osapi-io/osapi/pkg/sdk/client"
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
	pull, err := c.ContainerImage.Pull(
		ctx,
		target,
		client.ContainerImagePullInput{Image: "alpine:3.20"},
	)
	if err != nil {
		log.Fatalf("pull: %v", err)
	}

	for _, r := range pull.Data.Results {
		fmt.Printf("Pull (%s): id=%s size=%d changed=%v\n",
			r.Hostname, r.ImageID, r.Size, r.Changed)
	}

	// Create a container, and start it.
	name := "osapi-example"
	create, err := c.ContainerDocker.Create(ctx, target, client.ContainerDockerCreateInput{
		Name:      name,
		Image:     "alpine:3.20",
		Command:   []string{"sleep", "300"},
		AutoStart: true,
	})
	if err != nil {
		log.Fatalf("create: %v", err)
	}

	for _, r := range create.Data.Results {
		fmt.Printf("Create (%s): id=%s changed=%v\n", r.Hostname, r.ID, r.Changed)
	}

	// List the running containers.
	list, err := c.ContainerDocker.List(
		ctx,
		target,
		client.ContainerDockerListParams{State: "running"},
	)
	if err != nil {
		log.Fatalf("list: %v", err)
	}

	for _, r := range list.Data.Results {
		fmt.Printf("Container (%s): %s  %s  %s\n", r.Hostname, r.Name, r.Image, r.State)
	}

	// Get the container.
	get, err := c.ContainerDocker.Get(ctx, target, name)
	if err != nil {
		log.Fatalf("get: %v", err)
	}

	for _, r := range get.Data.Results {
		fmt.Printf("Get (%s): id=%s ports=%v env=%v\n", r.Hostname, r.ID, r.Ports, r.EnvNames)
	}

	// Run a command inside the container.
	exec, err := c.ContainerDocker.Exec(ctx, target, client.ContainerDockerExecInput{
		Name:    name,
		Command: []string{"cat", "/etc/hostname"},
	})
	if err != nil {
		log.Fatalf("exec: %v", err)
	}

	for _, r := range exec.Data.Results {
		fmt.Printf("Exec (%s): stdout=%s exit=%d\n", r.Hostname, r.Stdout, r.ExitCode)
	}

	// Stop the container.
	stop, err := c.ContainerDocker.Stop(
		ctx,
		target,
		client.ContainerDockerStopInput{Name: name, Timeout: 5},
	)
	if err != nil {
		log.Fatalf("stop: %v", err)
	}

	for _, r := range stop.Data.Results {
		fmt.Printf("Stop (%s): changed=%v\n", r.Hostname, r.Changed)
	}

	// Remove the container.
	remove, err := c.ContainerDocker.Delete(ctx, target, name)
	if err != nil {
		log.Fatalf("remove: %v", err)
	}

	for _, r := range remove.Data.Results {
		fmt.Printf("Remove (%s): changed=%v\n", r.Hostname, r.Changed)
	}

	// Remove the image.
	imgRemove, err := c.ContainerImage.Delete(
		ctx,
		target,
		client.ContainerImageDeleteInput{Image: "alpine:3.20"},
	)
	if err != nil {
		log.Fatalf("image remove: %v", err)
	}

	for _, r := range imgRemove.Data.Results {
		fmt.Printf("ImageRemove (%s): changed=%v\n", r.Hostname, r.Changed)
	}
}
