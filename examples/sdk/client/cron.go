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

// Package main demonstrates cron drop-in management: create, list, get,
// update, and delete cron entries in /etc/cron.d/.
//
// Run with: OSAPI_TOKEN="<jwt>" go run schedule.go
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
	target := "_any"

	// Create a cron entry.
	fmt.Println("=== Creating cron entry ===")
	createResp, err := c.Schedule.CronCreate(ctx, target, client.CronCreateOpts{
		Name:     "backup-daily",
		Schedule: "0 2 * * *",
		Command:  "/usr/local/bin/backup.sh --full",
		User:     "root",
	})
	if err != nil {
		log.Fatalf("create failed: %v", err)
	}
	fmt.Printf("Created: %s (changed: %v)\n", createResp.Data.Name, createResp.Data.Changed)

	// List all cron entries.
	fmt.Println("\n=== Listing cron entries ===")
	listResp, err := c.Schedule.CronList(ctx, target)
	if err != nil {
		log.Fatalf("list failed: %v", err)
	}
	for _, entry := range listResp.Data.Results {
		fmt.Printf("  %s: %s %s %s\n", entry.Name, entry.Schedule, entry.User, entry.Command)
	}

	// Get a specific entry.
	fmt.Println("\n=== Getting cron entry ===")
	getResp, err := c.Schedule.CronGet(ctx, target, "backup-daily")
	if err != nil {
		log.Fatalf("get failed: %v", err)
	}
	fmt.Printf("Name: %s\nSchedule: %s\nUser: %s\nCommand: %s\n",
		getResp.Data.Name, getResp.Data.Schedule,
		getResp.Data.User, getResp.Data.Command)

	// Update the schedule.
	fmt.Println("\n=== Updating cron entry ===")
	newSchedule := "0 3 * * *"
	updateResp, err := c.Schedule.CronUpdate(ctx, target, "backup-daily", client.CronUpdateOpts{
		Schedule: newSchedule,
	})
	if err != nil {
		log.Fatalf("update failed: %v", err)
	}
	fmt.Printf("Updated: %s (changed: %v)\n", updateResp.Data.Name, updateResp.Data.Changed)

	// Delete the entry.
	fmt.Println("\n=== Deleting cron entry ===")
	deleteResp, err := c.Schedule.CronDelete(ctx, target, "backup-daily")
	if err != nil {
		log.Fatalf("delete failed: %v", err)
	}
	fmt.Printf("Deleted: %s (changed: %v)\n", deleteResp.Data.Name, deleteResp.Data.Changed)
}
