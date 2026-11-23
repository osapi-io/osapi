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
// Package main runs the methods of the File service that a person wrote, and
// that file.go, which internal/sdk/gen writes, cannot call: Upload and Update
// send a multipart form, and GetChanged compares content with the stored
// file on the client. The program uploads a file, checks it for a change,
// and replaces it.
//
// Run with: OSAPI_TOKEN="<jwt>" go run file_upload.go
package main

import (
	"bytes"
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

	// Upload a raw file to the Object Store. Upload skips a file that the
	// Object Store has with the same content.
	content := []byte("listen_address = 0.0.0.0:8080\nworkers = 4\n")
	upload, err := c.File.Upload(ctx, "app.conf", "raw", bytes.NewReader(content))
	if err != nil {
		log.Fatalf("upload: %v", err)
	}

	fmt.Printf("Uploaded: name=%s sha256=%s size=%d changed=%v\n",
		upload.Data.Name, upload.Data.SHA256, upload.Data.Size, upload.Data.Changed)

	// Check if the file has changed, with no upload.
	chk, err := c.File.GetChanged(ctx, "app.conf", bytes.NewReader(content))
	if err != nil {
		log.Fatalf("changed: %v", err)
	}

	fmt.Printf("Changed: name=%s changed=%v\n", chk.Data.Name, chk.Data.Changed)

	// Update replaces the file, or creates it when the name is new.
	updated, err := c.File.Update(ctx, "app.conf", "raw", bytes.NewReader(content))
	if err != nil {
		log.Fatalf("update: %v", err)
	}

	fmt.Printf("Update: name=%s changed=%v\n", updated.Data.Name, updated.Data.Changed)
}
