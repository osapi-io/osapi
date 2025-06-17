// Copyright (c) 2025 John Dewey

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

// Package jobs provides NATS subject hierarchy for distributed job routing.
//
// Subject Format: jobs.{type}.{hostname}
//
// Routing Patterns:
//   - Direct: jobs.query.server1 (specific host)
//   - Any: jobs.query._any (load-balanced across available workers)
//   - Broadcast: jobs.modify._all (all workers receive)
//
// Workers subscribe to:
//   - Their specific hostname: jobs.*.server1
//   - Load-balanced work: jobs.*._any (with queue group)
//   - Broadcast messages: jobs.*._all
package job

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	// JobsQueryPrefix is the subject hierarchy prefix for query operations.
	JobsQueryPrefix = "jobs.query"
	// JobsModifyPrefix is the subject hierarchy prefix for modify operations.
	JobsModifyPrefix = "jobs.modify"

	// AllHosts is a wildcard for targeting all hosts.
	AllHosts = "*" // Wildcard for targeting all hosts
	// AnyHost is load-balanced across available hosts.
	AnyHost = "_any" // Load-balanced across available hosts
	// LocalHost targets the API server's host.
	LocalHost = "_local" // Target the API server's host
	// BroadcastHost broadcasts to all hosts (no queue group).
	BroadcastHost = "_all" // Broadcast to all hosts (no queue group)
)

// Subject categories for different operations
const (
	SubjectCategorySystem  = "system"
	SubjectCategoryNetwork = "network"
)

// System operation types
const (
	SystemOperationHostname = "hostname"
	SystemOperationStatus   = "status"
)

// Network operation types
const (
	NetworkOperationDNS  = "dns"
	NetworkOperationPing = "ping"
)

// BuildQuerySubject creates a subject for query operations.
// Example: jobs.query.hostname
func BuildQuerySubject(
	hostname string,
) string {
	return fmt.Sprintf("%s.%s", JobsQueryPrefix, hostname)
}

// BuildModifySubject creates a subject for modify operations.
// Example: jobs.modify.hostname
func BuildModifySubject(
	hostname string,
) string {
	return fmt.Sprintf("%s.%s", JobsModifyPrefix, hostname)
}

// BuildQuerySubjectForAllHosts creates a query subject targeting all hosts.
// Example: jobs.query.*
func BuildQuerySubjectForAllHosts() string {
	return BuildQuerySubject(AllHosts)
}

// BuildModifySubjectForAllHosts creates a modify subject targeting all hosts.
// Example: jobs.modify.*
func BuildModifySubjectForAllHosts() string {
	return BuildModifySubject(AllHosts)
}

// ParseSubject extracts components from a job subject.
// Expected format: jobs.{type}.{hostname}
func ParseSubject(
	subject string,
) (prefix, hostname string, err error) {
	parts := strings.Split(subject, ".")
	if len(parts) != 3 {
		return "", "", fmt.Errorf("invalid subject format: %s", subject)
	}

	prefix = fmt.Sprintf("%s.%s", parts[0], parts[1])
	hostname = parts[2]

	return prefix, hostname, nil
}

// BuildWorkerSubscriptionPattern creates subscription patterns for workers.
// Workers typically subscribe to their own hostname and special routing patterns.
func BuildWorkerSubscriptionPattern(
	hostname string,
) []string {
	return []string{
		fmt.Sprintf("jobs.*.%s", hostname),      // Direct messages to this host
		fmt.Sprintf("jobs.*.%s", AnyHost),       // Load-balanced messages
		fmt.Sprintf("jobs.*.%s", BroadcastHost), // Broadcast messages
	}
}

// BuildWorkerQueueGroup returns the queue group name for load-balanced subscriptions.
// This ensures only one worker processes each "_any" message.
func BuildWorkerQueueGroup(
	category string,
) string {
	return fmt.Sprintf("workers.%s", category)
}

// IsSpecialHostname checks if a hostname is a special routing directive.
func IsSpecialHostname(
	hostname string,
) bool {
	return hostname == AllHosts || hostname == AnyHost ||
		hostname == LocalHost || hostname == BroadcastHost
}

// SanitizeHostname converts a hostname to a valid NATS consumer/routing name.
// NATS consumer names and routing must be alphanumeric with underscores only.
func SanitizeHostname(
	hostname string,
) string {
	// Replace any non-alphanumeric characters (except underscores) with underscores
	reg := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	return reg.ReplaceAllString(hostname, "_")
}
