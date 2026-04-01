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

// Package main demonstrates CA certificate management: upload a PEM file to
// the Object Store, then create, list, update, and delete CA certificates
// in the system trust store.
//
// All responses return Collection[T] with per-host results.
// Use .Data.Results to iterate over the per-host entries.
//
// Run with: OSAPI_TOKEN="<jwt>" go run certificate.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

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
	target := "_any"

	// Upload a CA certificate PEM to the Object Store.
	fmt.Println("=== Uploading CA certificate PEM ===")
	pem := strings.NewReader(
		"-----BEGIN CERTIFICATE-----\nMIIBkTCB...\n-----END CERTIFICATE-----\n",
	)
	uploadResp, err := c.File.Upload(ctx, "internal-ca", "raw", pem)
	if err != nil {
		log.Fatalf("upload failed: %v", err)
	}
	fmt.Printf("Uploaded: %s (changed: %v)\n", uploadResp.Data.Name, uploadResp.Data.Changed)

	// Deploy the certificate to the host.
	fmt.Println("\n=== Creating CA certificate ===")
	createResp, err := c.Certificate.Create(ctx, target, client.CertificateCreateOpts{
		Name:   "internal-ca",
		Object: "internal-ca",
	})
	if err != nil {
		log.Fatalf("create failed: %v", err)
	}
	for _, r := range createResp.Data.Results {
		fmt.Printf("  %s: changed=%v error=%s\n", r.Hostname, r.Changed, r.Error)
	}

	// List all certificates.
	fmt.Println("\n=== Listing CA certificates ===")
	listResp, err := c.Certificate.List(ctx, target)
	if err != nil {
		log.Fatalf("list failed: %v", err)
	}
	for _, r := range listResp.Data.Results {
		if r.Error != "" {
			fmt.Printf("  %s: ERROR %s\n", r.Hostname, r.Error)
		} else {
			for _, cert := range r.Certificates {
				fmt.Printf("  %s: %s (source=%s)\n", r.Hostname, cert.Name, cert.Source)
			}
		}
	}

	// Clean up: delete the certificate and the uploaded object.
	fmt.Println("\n=== Deleting CA certificate ===")
	deleteResp, err := c.Certificate.Delete(ctx, target, "internal-ca")
	if err != nil {
		log.Fatalf("delete failed: %v", err)
	}
	for _, r := range deleteResp.Data.Results {
		fmt.Printf("  %s: changed=%v error=%s\n", r.Hostname, r.Changed, r.Error)
	}

	fmt.Println("\n=== Cleaning up uploaded object ===")
	_, err = c.File.Delete(ctx, "internal-ca")
	if err != nil {
		log.Fatalf("delete object failed: %v", err)
	}
	fmt.Println("Deleted object: internal-ca")
}
