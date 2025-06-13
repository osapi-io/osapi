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
// Subject Format: job.{type}.{hostname}.{category}.{operation}
//
// Routing Patterns:
//   - Direct: job.query.server1.system.status (specific host)
//   - Any: job.query._any.system.status (load-balanced across available workers)
//   - Broadcast: job.query._all.system.status (all workers receive)
//   - Wildcard: job.query.*.system.status (matches any hostname)
//
// Workers subscribe to:
//   - Their specific hostname: job.*.server1.>
//   - Load-balanced work: job.*._any.> (with queue group)
//   - Broadcast messages: job.*._all.>
package job

import (
	"fmt"
	"os"
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
// Example: job.query.hostname.system.status
func BuildQuerySubject(
	hostname string,
	category string,
	operation string,
) string {
	return fmt.Sprintf("%s.%s.%s.%s", JobsQueryPrefix, hostname, category, operation)
}

// BuildModifySubject creates a subject for modify operations.
// Example: job.modify.hostname.network.dns
func BuildModifySubject(
	hostname string,
	category string,
	operation string,
) string {
	return fmt.Sprintf("%s.%s.%s.%s", JobsModifyPrefix, hostname, category, operation)
}

// BuildQuerySubjectForAllHosts creates a query subject targeting all hosts.
// Example: job.query.*.system.status
func BuildQuerySubjectForAllHosts(
	category string,
	operation string,
) string {
	return BuildQuerySubject(AllHosts, category, operation)
}

// BuildModifySubjectForAllHosts creates a modify subject targeting all hosts.
// Example: job.modify.*.network.dns
func BuildModifySubjectForAllHosts(
	category string,
	operation string,
) string {
	return BuildModifySubject(AllHosts, category, operation)
}

// GetLocalHostname returns the current system hostname.
func GetLocalHostname() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}
	return hostname, nil
}

// ParseSubject extracts components from a job subject.
// Supports both legacy format (5 parts) and new dotted operations (6+ parts).
func ParseSubject(
	subject string,
) (prefix, hostname, category, operation string, err error) {
	parts := strings.Split(subject, ".")
	if len(parts) < 5 {
		return "", "", "", "", fmt.Errorf("invalid subject format: %s", subject)
	}

	prefix = fmt.Sprintf("%s.%s", parts[0], parts[1])
	hostname = parts[2]
	category = parts[3]

	// Operation may be dotted (e.g., "hostname.get"), so join remaining parts
	operation = strings.Join(parts[4:], ".")

	return prefix, hostname, category, operation, nil
}

// BuildWorkerSubscriptionPattern creates subscription patterns for workers.
// Workers typically subscribe to their own hostname and special routing patterns.
func BuildWorkerSubscriptionPattern(
	hostname string,
) []string {
	return []string{
		fmt.Sprintf("job.*.%s.>", hostname),      // Direct messages to this host
		fmt.Sprintf("job.*.%s.>", AnyHost),       // Load-balanced messages
		fmt.Sprintf("job.*.%s.>", BroadcastHost), // Broadcast messages
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
func SanitizeHostname(hostname string) string {
	// Replace any non-alphanumeric characters (except underscores) with underscores
	reg := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	return reg.ReplaceAllString(hostname, "_")
}
