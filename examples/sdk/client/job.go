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

// Package main demonstrates the JobService: listing jobs, retrieving
// a specific job's details, and viewing timeline events.
//
// Jobs are created implicitly by domain operations (e.g. Node.Hostname,
// Docker.Create). This example shows how to inspect them after the fact.
//
// Run with: OSAPI_TOKEN="<jwt>" go run job.go
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

	// Trigger a job via a domain operation so we have something to inspect.
	hn, err := c.Hostname.Get(ctx, "_all")
	if err != nil {
		log.Fatalf("hostname: %v", err)
	}

	jobID := hn.Data.JobID
	fmt.Printf("Created job via Node.Hostname: %s\n", jobID)

	// Get the job's full details.
	job, err := c.Job.Get(ctx, jobID)
	if err != nil {
		log.Fatalf("get job: %v", err)
	}

	fmt.Printf("\nJob %s:\n", job.Data.ID)
	fmt.Printf("  Status:    %s\n", job.Data.Status)
	fmt.Printf("  Hostname:  %s\n", job.Data.Hostname)
	fmt.Printf("  Operation: %v\n", job.Data.Operation)
	fmt.Printf("  Created:   %s\n", job.Data.Created)

	if job.Data.Error != "" {
		fmt.Printf("  Error:     %s\n", job.Data.Error)
	}

	// Timeline events show the job's lifecycle.
	if len(job.Data.Timeline) > 0 {
		fmt.Printf("\n  Timeline:\n")
		for _, e := range job.Data.Timeline {
			fmt.Printf("    %s  %-12s  host=%s\n",
				e.Timestamp, e.Event, e.Hostname)
		}
	}

	// List recent jobs with status counts.
	list, err := c.Job.List(ctx, client.ListParams{Limit: 10})
	if err != nil {
		log.Fatalf("list jobs: %v", err)
	}

	fmt.Printf("\nRecent jobs: %d total\n", list.Data.TotalItems)

	if len(list.Data.StatusCounts) > 0 {
		fmt.Printf("  Status counts: %v\n", list.Data.StatusCounts)
	}

	for _, j := range list.Data.Items {
		fmt.Printf("  %s  status=%-10s  op=%v\n",
			j.ID, j.Status, j.Operation)
	}
}
