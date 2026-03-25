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

// Package main demonstrates cron drop-in management: upload a script to the
// Object Store, then create, list, get, update, and delete cron entries in
// /etc/cron.d/ using the object-based workflow.
//
// All mutation and query responses return Collection[T] with per-host results.
// Use .Data.Results to iterate over the per-host entries.
//
// Run with: OSAPI_TOKEN="<jwt>" go run cron.go
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
	target := "_all"

	// Upload the backup script to the Object Store first.
	// The cron entry references the stored object by name.
	fmt.Println("=== Uploading backup script ===")
	backupScript := strings.NewReader("#!/bin/sh\n/usr/local/bin/backup.sh --full\n")
	uploadResp, err := c.File.Upload(
		ctx,
		"backup-script",
		"raw",
		backupScript,
			)
	if err != nil {
		log.Fatalf("upload failed: %v", err)
	}
	fmt.Printf("Uploaded: %s (changed: %v)\n", uploadResp.Data.Name, uploadResp.Data.Changed)

	// Upload the logrotate script for the periodic entry.
	fmt.Println("\n=== Uploading logrotate script ===")
	logrotateScript := strings.NewReader("#!/bin/sh\n/usr/sbin/logrotate /etc/logrotate.conf\n")
	_, err = c.File.Upload(ctx, "logrotate-script", "raw", logrotateScript)
	if err != nil {
		log.Fatalf("upload logrotate script failed: %v", err)
	}
	fmt.Println("Uploaded: logrotate-script")

	// Create a cron entry referencing the uploaded object.
	// Returns Collection[CronMutationResult] with per-host results.
	fmt.Println("\n=== Creating cron entry ===")
	createResp, err := c.Schedule.CronCreate(ctx, target, client.CronCreateOpts{
		Name:     "backup-daily",
		Schedule: "0 2 * * *",
		Object:   "backup-script",
		User:     "root",
	})
	if err != nil {
		log.Fatalf("create failed: %v", err)
	}
	for _, r := range createResp.Data.Results {
		fmt.Printf("  %s: changed=%v error=%s\n", r.Hostname, r.Changed, r.Error)
	}

	// Create a periodic entry (interval-based) referencing an uploaded object.
	fmt.Println("\n=== Creating periodic cron entry ===")
	periodicResp, err := c.Schedule.CronCreate(ctx, target, client.CronCreateOpts{
		Name:     "logrotate",
		Interval: "daily",
		Object:   "logrotate-script",
	})
	if err != nil {
		log.Fatalf("create periodic failed: %v", err)
	}
	for _, r := range periodicResp.Data.Results {
		fmt.Printf("  %s: changed=%v error=%s\n", r.Hostname, r.Changed, r.Error)
	}

	// List all cron entries.
	// Returns Collection[CronEntryResult] with per-host entries.
	fmt.Println("\n=== Listing cron entries ===")
	listResp, err := c.Schedule.CronList(ctx, target)
	if err != nil {
		log.Fatalf("list failed: %v", err)
	}
	for _, entry := range listResp.Data.Results {
		if entry.Error != "" {
			fmt.Printf("  %s: ERROR %s\n", entry.Hostname, entry.Error)
		} else {
			fmt.Printf("  %s: %s %s %s %s\n",
				entry.Hostname, entry.Name, entry.Schedule, entry.User, entry.Object)
		}
	}

	// Get a specific entry.
	// Returns Collection[CronEntryResult] — one result per host.
	fmt.Println("\n=== Getting cron entry ===")
	getResp, err := c.Schedule.CronGet(ctx, target, "backup-daily")
	if err != nil {
		log.Fatalf("get failed: %v", err)
	}
	for _, r := range getResp.Data.Results {
		if r.Error != "" {
			fmt.Printf("  %s: ERROR %s\n", r.Hostname, r.Error)
		} else {
			fmt.Printf("  %s: name=%s schedule=%s user=%s object=%s\n",
				r.Hostname, r.Name, r.Schedule, r.User, r.Object)
		}
	}

	// Update: upload a new version of the script and redeploy.
	fmt.Println("\n=== Uploading new backup script version ===")
	newBackupScript := strings.NewReader("#!/bin/sh\n/usr/local/bin/backup.sh --full --compress\n")
	_, err = c.File.Upload(ctx, "backup-script-v2", "raw", newBackupScript)
	if err != nil {
		log.Fatalf("upload new version failed: %v", err)
	}
	fmt.Println("Uploaded: backup-script-v2")

	fmt.Println("\n=== Updating cron entry ===")
	updateResp, err := c.Schedule.CronUpdate(ctx, target, "backup-daily", client.CronUpdateOpts{
		Schedule: "0 3 * * *",
		Object:   "backup-script-v2",
	})
	if err != nil {
		log.Fatalf("update failed: %v", err)
	}
	for _, r := range updateResp.Data.Results {
		fmt.Printf("  %s: changed=%v error=%s\n", r.Hostname, r.Changed, r.Error)
	}

	// Delete the entry. The cron file is removed from disk; file-state KV
	// tracking is preserved so the undeploy is recorded.
	fmt.Println("\n=== Deleting cron entry ===")
	deleteResp, err := c.Schedule.CronDelete(ctx, target, "backup-daily")
	if err != nil {
		log.Fatalf("delete failed: %v", err)
	}
	for _, r := range deleteResp.Data.Results {
		fmt.Printf("  %s: changed=%v error=%s\n", r.Hostname, r.Changed, r.Error)
	}

	// Clean up the periodic entry.
	fmt.Println("\n=== Cleaning up periodic entry ===")
	cleanupResp, err := c.Schedule.CronDelete(ctx, target, "logrotate")
	if err != nil {
		log.Fatalf("cleanup failed: %v", err)
	}
	for _, r := range cleanupResp.Data.Results {
		fmt.Printf("  %s: changed=%v error=%s\n", r.Hostname, r.Changed, r.Error)
	}

	// Clean up uploaded objects from the Object Store.
	fmt.Println("\n=== Cleaning up uploaded objects ===")
	for _, name := range []string{"backup-script", "backup-script-v2", "logrotate-script"} {
		_, err = c.File.Delete(ctx, name)
		if err != nil {
			log.Fatalf("delete object %s failed: %v", name, err)
		}
		fmt.Printf("Deleted object: %s\n", name)
	}
}
