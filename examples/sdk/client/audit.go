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

// Package main demonstrates the AuditService: listing audit entries,
// retrieving a specific entry, and exporting all entries.
//
// Run with: OSAPI_TOKEN="<jwt>" go run audit.go
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

	// List recent audit entries.
	list, err := c.Audit.List(ctx, 10, 0)
	if err != nil {
		log.Fatalf("list audit: %v", err)
	}

	fmt.Printf("Audit entries: %d total\n", list.Data.TotalItems)

	for _, e := range list.Data.Items {
		fmt.Printf("  %s  %s %s  code=%d  user=%s\n",
			e.ID, e.Method, e.Path, e.ResponseCode, e.User)
	}

	if len(list.Data.Items) == 0 {
		return
	}

	// Get a specific audit entry.
	id := list.Data.Items[0].ID

	entry, err := c.Audit.Get(ctx, id)
	if err != nil {
		log.Fatalf("get audit %s: %v", id, err)
	}

	fmt.Printf("\nEntry %s:\n", entry.Data.ID)
	fmt.Printf("  Method:    %s\n", entry.Data.Method)
	fmt.Printf("  Path:      %s\n", entry.Data.Path)
	fmt.Printf("  User:      %s\n", entry.Data.User)
	fmt.Printf("  Duration:  %dms\n", entry.Data.DurationMs)

	// Export all audit entries.
	export, err := c.Audit.Export(ctx)
	if err != nil {
		log.Fatalf("export audit: %v", err)
	}

	fmt.Printf("\nExported: %d entries\n", export.Data.TotalItems)

	for _, e := range export.Data.Items {
		fmt.Printf("  %s  %s %s  code=%d\n",
			e.ID, e.Method, e.Path, e.ResponseCode)
	}
}
