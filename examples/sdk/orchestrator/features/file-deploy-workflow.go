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

// Package main demonstrates file deployment orchestration: upload a
// template file, deploy to all agents with template rendering, then
// verify status.
//
// DAG:
//
//	upload-template
//	    └── deploy-config
//	            └── verify-status
//
// Run with: OSAPI_TOKEN="<jwt>" go run file-deploy-workflow.go
package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/retr0h/osapi/pkg/sdk/client"
	"github.com/retr0h/osapi/pkg/sdk/orchestrator"
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

	hooks := orchestrator.Hooks{
		BeforeTask: func(task *orchestrator.Task) {
			fmt.Printf("  [start] %s\n", task.Name())
		},
		AfterTask: func(_ *orchestrator.Task, result orchestrator.TaskResult) {
			fmt.Printf("  [%s] %s  changed=%v\n",
				result.Status, result.Name, result.Changed)
		},
	}

	plan := orchestrator.NewPlan(client, orchestrator.WithHooks(hooks))

	// Step 1: Upload the template file to Object Store.
	upload := plan.TaskFunc(
		"upload-template",
		func(ctx context.Context, c *client.Client) (*orchestrator.Result, error) {
			tmpl := []byte(`# Generated for {{ .Hostname }}
listen_address = {{ .Vars.listen_address }}
workers = {{ .Facts.cpu_count }}
`)
			resp, err := c.File.Upload(
				ctx,
				"app.conf.tmpl",
				"template",
				bytes.NewReader(tmpl),
				client.WithForce(),
			)
			if err != nil {
				return nil, fmt.Errorf("upload: %w", err)
			}

			fmt.Printf("    uploaded %s (sha256=%s changed=%v)\n",
				resp.Data.Name, resp.Data.SHA256, resp.Data.Changed)

			return &orchestrator.Result{Changed: resp.Data.Changed}, nil
		},
	)

	// Step 2: Deploy the template to all agents.
	deploy := plan.Task("deploy-config", &orchestrator.Op{
		Operation: "file.deploy.execute",
		Target:    "_all",
		Params: map[string]any{
			"object_name":  "app.conf.tmpl",
			"path":         "/tmp/app.conf",
			"content_type": "template",
			"mode":         "0644",
			"vars": map[string]any{
				"listen_address": "0.0.0.0:8080",
			},
		},
	})
	deploy.DependsOn(upload)

	// Step 3: Verify the deployed file is in-sync.
	verify := plan.Task("verify-status", &orchestrator.Op{
		Operation: "file.status.get",
		Target:    "_all",
		Params: map[string]any{
			"path": "/tmp/app.conf",
		},
	})
	verify.DependsOn(deploy)

	report, err := plan.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n%s in %s\n", report.Summary(), report.Duration)
}
