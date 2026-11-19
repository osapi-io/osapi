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

// Package main demonstrates the admission SDK operations: listing the
// machines that ask to join, and admitting one.
//
// The fingerprint comes from the machine, not from the list: run
// `osapi agent admission` on it and give what it prints. The controller
// admits the machine only when that fingerprint is the one of its request.
//
// Run with: OSAPI_TOKEN="<jwt>" go run admission.go [<machine-id> <fingerprint>]
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

	// List the machines that wait for a decision.
	pending, err := c.Admission.List(ctx, "pending")
	if err != nil {
		log.Fatalf("list pending: %v", err)
	}

	fmt.Printf("Pending machines: %d\n", pending.Data.TotalItems)

	for _, a := range pending.Data.Items {
		fmt.Printf("  %s  %s  fingerprint=%s\n", a.MachineID, a.Hostname, a.Fingerprint)
	}

	if len(os.Args) != 3 {
		fmt.Println("\nTo admit one: go run admission.go <machine-id> <fingerprint>")
		return
	}

	// Admit the machine, with the fingerprint read on the machine.
	machineID, fingerprint := os.Args[1], os.Args[2]
	if err := c.Admission.Admit(ctx, machineID, fingerprint); err != nil {
		log.Fatalf("admit %s: %v", machineID, err)
	}

	fmt.Printf("\nAdmitted: %s\n", machineID)
}
