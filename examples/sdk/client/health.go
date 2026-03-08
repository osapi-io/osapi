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

// Package main demonstrates the HealthService: liveness, readiness,
// and detailed system status checks.
//
// Run with: OSAPI_TOKEN="<jwt>" go run health.go
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

	client := client.New(url, token)
	ctx := context.Background()

	// Liveness — is the API process running?
	live, err := client.Health.Liveness(ctx)
	if err != nil {
		log.Fatalf("liveness: %v", err)
	}

	fmt.Printf("Liveness: %s\n", live.Data.Status)

	// Readiness — is the API ready to serve requests?
	ready, err := client.Health.Ready(ctx)
	if err != nil {
		log.Fatalf("readiness: %v", err)
	}

	fmt.Printf("Readiness: %s\n", ready.Data.Status)

	// Status — detailed system info (requires auth).
	status, err := client.Health.Status(ctx)
	if err != nil {
		log.Fatalf("status: %v", err)
	}

	fmt.Printf("Status:  %s\n", status.Data.Status)
	fmt.Printf("Version: %s\n", status.Data.Version)
	fmt.Printf("Uptime:  %s\n", status.Data.Uptime)
}
