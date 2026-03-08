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

// Package main demonstrates file management: upload, check for changes,
// force upload, list, get metadata, deploy to an agent, check status,
// and delete.
//
// Run with: OSAPI_TOKEN="<jwt>" go run file.go
package main

import (
	"bytes"
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

	// Upload a raw file to the Object Store.
	content := []byte("listen_address = 0.0.0.0:8080\nworkers = 4\n")
	upload, err := client.File.Upload(
		ctx,
		"app.conf",
		"raw",
		bytes.NewReader(content),
	)
	if err != nil {
		log.Fatalf("upload: %v", err)
	}

	fmt.Printf("Uploaded: name=%s sha256=%s size=%d changed=%v\n",
		upload.Data.Name, upload.Data.SHA256, upload.Data.Size, upload.Data.Changed)

	// Check if the file has changed without uploading.
	chk, err := client.File.Changed(ctx, "app.conf", bytes.NewReader(content))
	if err != nil {
		log.Fatalf("changed: %v", err)
	}

	fmt.Printf("Changed: name=%s changed=%v\n", chk.Data.Name, chk.Data.Changed)

	// Force upload bypasses both SDK-side and server-side checks.
	force, err := client.File.Upload(
		ctx,
		"app.conf",
		"raw",
		bytes.NewReader(content),
		client.WithForce(),
	)
	if err != nil {
		log.Fatalf("force upload: %v", err)
	}

	fmt.Printf("Force upload: name=%s changed=%v\n",
		force.Data.Name, force.Data.Changed)

	// List all stored files.
	list, err := client.File.List(ctx)
	if err != nil {
		log.Fatalf("list: %v", err)
	}

	fmt.Printf("\nStored files (%d):\n", list.Data.Total)
	for _, f := range list.Data.Files {
		fmt.Printf("  %s  size=%d\n", f.Name, f.Size)
	}

	// Get metadata for a specific file.
	meta, err := client.File.Get(ctx, "app.conf")
	if err != nil {
		log.Fatalf("get: %v", err)
	}

	fmt.Printf("\nMetadata: name=%s sha256=%s size=%d\n",
		meta.Data.Name, meta.Data.SHA256, meta.Data.Size)

	// Deploy the file to an agent.
	deploy, err := client.Node.FileDeploy(ctx, client.FileDeployOpts{
		ObjectName:  "app.conf",
		Path:        "/tmp/app.conf",
		ContentType: "raw",
		Mode:        "0644",
		Target:      "_any",
	})
	if err != nil {
		log.Fatalf("deploy: %v", err)
	}

	fmt.Printf("\nDeploy: job=%s host=%s changed=%v\n",
		deploy.Data.JobID, deploy.Data.Hostname, deploy.Data.Changed)

	// Check file status on the agent.
	status, err := client.Node.FileStatus(ctx, "_any", "/tmp/app.conf")
	if err != nil {
		log.Fatalf("status: %v", err)
	}

	fmt.Printf("Status: path=%s status=%s\n",
		status.Data.Path, status.Data.Status)

	// Clean up — delete the file from the Object Store.
	del, err := client.File.Delete(ctx, "app.conf")
	if err != nil {
		log.Fatalf("delete: %v", err)
	}

	fmt.Printf("\nDeleted: name=%s deleted=%v\n",
		del.Data.Name, del.Data.Deleted)
}
