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

// Package main demonstrates command execution: direct exec and
// shell-interpreted commands.
//
// Run with: OSAPI_TOKEN="<jwt>" go run command.go
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

	// Direct exec — runs a binary with arguments.
	exec, err := c.Command.Exec(ctx, client.ExecRequest{
		Target:  target,
		Command: "uptime",
	})
	if err != nil {
		log.Fatalf("exec: %v", err)
	}

	for _, r := range exec.Data.Results {
		fmt.Printf("Exec (%s):\n", r.Hostname)
		fmt.Printf("  stdout: %s\n", r.Stdout)
		fmt.Printf("  exit:   %d\n", r.ExitCode)
	}

	// Shell — interpreted by /bin/sh, supports pipes and redirection.
	shell, err := c.Command.Shell(ctx, client.ShellRequest{
		Target:  target,
		Command: "uname -a",
	})
	if err != nil {
		log.Fatalf("shell: %v", err)
	}

	for _, r := range shell.Data.Results {
		fmt.Printf("Shell (%s):\n", r.Hostname)
		fmt.Printf("  stdout: %s\n", r.Stdout)
		fmt.Printf("  exit:   %d\n", r.ExitCode)
	}
}
