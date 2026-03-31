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

// Package main demonstrates the FileDeployService: deploying files from
// the Object Store to agent hosts and checking deploy status.
//
// Run with: OSAPI_TOKEN="<jwt>" go run file_deploy.go
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

	// Deploy a file to all agents.
	deploy, err := c.FileDeploy.Deploy(ctx, client.FileDeployOpts{
		ObjectName:  "app.conf",
		Path:        "/tmp/app.conf",
		ContentType: "raw",
		Mode:        "0644",
		Target:      "_all",
	})
	if err != nil {
		log.Fatalf("deploy: %v", err)
	}

	fmt.Printf("Deploy: job=%s\n", deploy.Data.JobID)
	for _, r := range deploy.Data.Results {
		fmt.Printf("  %s: changed=%v error=%s\n", r.Hostname, r.Changed, r.Error)
	}

	// Check file status on the agents.
	status, err := c.FileDeploy.Status(ctx, "_all", "/tmp/app.conf")
	if err != nil {
		log.Fatalf("status: %v", err)
	}

	fmt.Printf("\nStatus: job=%s\n", status.Data.JobID)
	for _, r := range status.Data.Results {
		fmt.Printf("  %s: path=%s status=%s\n", r.Hostname, r.Path, r.Status)
	}
}
