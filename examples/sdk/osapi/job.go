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

//go:build ignore

// Package main demonstrates the JobService: creating a job, polling
// for its result, listing jobs, and checking queue statistics.
//
// Run with: OSAPI_TOKEN="<jwt>" go run job.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/retr0h/osapi/pkg/sdk/osapi"
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

	client := osapi.New(url, token)
	ctx := context.Background()

	// Create a job.
	created, err := client.Job.Create(ctx, map[string]any{
		"type": "node.hostname.get",
	}, "_any")
	if err != nil {
		log.Fatalf("create job: %v", err)
	}

	fmt.Printf("Created job: %s  status=%s\n",
		created.Data.JobID, created.Data.Status)

	// Poll until the job completes.
	time.Sleep(2 * time.Second)

	job, err := client.Job.Get(ctx, created.Data.JobID)
	if err != nil {
		log.Fatalf("get job: %v", err)
	}

	fmt.Printf("Job %s: status=%s\n", job.Data.ID, job.Data.Status)

	// List recent jobs.
	list, err := client.Job.List(ctx, osapi.ListParams{Limit: 5})
	if err != nil {
		log.Fatalf("list jobs: %v", err)
	}

	fmt.Printf("\nRecent jobs: %d total\n", list.Data.TotalItems)

	for _, j := range list.Data.Items {
		fmt.Printf("  %s  status=%s  op=%v\n",
			j.ID, j.Status, j.Operation)
	}
}
