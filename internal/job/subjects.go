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

// Package job provides NATS subject hierarchy for distributed job routing.
//
// Subject Format: [namespace.]jobs.{type}.{routing_type}.{value...}
//
// Routing Patterns:
//   - Direct: jobs.query.host.server1 (specific host)
//   - Any: jobs.query._any (load-balanced across available workers)
//   - Broadcast: jobs.modify._all (all workers receive)
//   - Label: jobs.query.label.group.web (broadcast to label group)
//   - Hierarchical: jobs.query.label.group.web.dev.us-east (prefix matching)
//
// Workers subscribe to:
//   - Their specific hostname: jobs.*.host.server1
//   - Load-balanced work: jobs.*._any (with queue group)
//   - Broadcast messages: jobs.*._all
//   - Label prefixes: jobs.*.label.group.web, jobs.*.label.group.web.dev, etc.
//
// When a namespace is configured via Init(), all subjects are prefixed:
//
//	Init("osapi") -> osapi.jobs.query._any, osapi.jobs.*.host.server1, etc.
package job

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// JobsQueryPrefix is the subject hierarchy prefix for query operations.
	JobsQueryPrefix = "jobs.query"
	// JobsModifyPrefix is the subject hierarchy prefix for modify operations.
	JobsModifyPrefix = "jobs.modify"
	// jobsBase is the base subject token (e.g., "jobs" or "osapi.jobs").
	jobsBase = "jobs"
)

// Init configures the subject namespace. An empty namespace keeps the default
// "jobs.*" hierarchy. A non-empty namespace prepends it:
//
//	Init("")      -> jobs.query, jobs.modify
//	Init("osapi") -> osapi.jobs.query, osapi.jobs.modify
func Init(
	namespace string,
) {
	if namespace == "" {
		jobsBase = "jobs"
	} else {
		jobsBase = namespace + ".jobs"
	}
	JobsQueryPrefix = jobsBase + ".query"
	JobsModifyPrefix = jobsBase + ".modify"
}

const (
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
	SubjectCategoryNode    = "node"
	SubjectCategoryNetwork = "network"
)

// Node operation types
const (
	NodeOperationHostname = "hostname"
	NodeOperationStatus   = "status"
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

// ParseSubject extracts the prefix and routing target from a job subject.
// Supported formats (with optional namespace prefix):
//   - [ns.]jobs.{type}._any
//   - [ns.]jobs.{type}._all
//   - [ns.]jobs.{type}.host.{hostname}
//   - [ns.]jobs.{type}.label.{key}.{value...}
func ParseSubject(
	subject string,
) (prefix, hostname string, err error) {
	parts := strings.Split(subject, ".")

	// Find the "jobs" token in the parts to determine namespace offset.
	jobsIdx := -1
	for i, p := range parts {
		if p == "jobs" {
			jobsIdx = i
			break
		}
	}
	if jobsIdx < 0 || jobsIdx+2 >= len(parts) {
		return "", "", fmt.Errorf("invalid subject format: %s", subject)
	}

	// prefix = everything up to and including the type token (e.g., "osapi.jobs.query")
	prefix = strings.Join(parts[:jobsIdx+2], ".")

	// remaining parts after the prefix
	remaining := parts[jobsIdx+2:]

	switch {
	case len(remaining) == 1:
		// _any, _all, or legacy hostname
		hostname = remaining[0]
	case len(remaining) == 2 && remaining[0] == "host":
		// [ns.]jobs.{type}.host.{hostname}
		hostname = remaining[1]
	case len(remaining) >= 3 && remaining[0] == "label":
		// [ns.]jobs.{type}.label.{key}.{value...}
		key := remaining[1]
		value := strings.Join(remaining[2:], ".")
		hostname = fmt.Sprintf("%s:%s", key, value)
	default:
		return "", "", fmt.Errorf("invalid subject format: %s", subject)
	}

	return prefix, hostname, nil
}

// BuildWorkerSubscriptionPattern creates subscription patterns for workers.
// Workers typically subscribe to their own hostname and special routing patterns.
// If labels are provided, hierarchical prefix subscriptions are included for
// each label. For example, a label "group: web.dev.us-east" generates subscriptions
// at every prefix level (group:web, group:web.dev, group:web.dev.us-east).
func BuildWorkerSubscriptionPattern(
	hostname string,
	labels map[string]string,
) []string {
	labelCount := 0
	for _, value := range labels {
		labelCount += len(strings.Split(value, "."))
	}

	patterns := make([]string, 0, 3+labelCount)
	patterns = append(
		patterns,
		fmt.Sprintf(
			"%s.*.host.%s",
			jobsBase,
			SanitizeHostname(hostname),
		), // Direct messages to this host
		fmt.Sprintf("%s.*.%s", jobsBase, AnyHost),       // Load-balanced messages
		fmt.Sprintf("%s.*.%s", jobsBase, BroadcastHost), // Broadcast messages
	)

	for key, value := range labels {
		patterns = append(patterns, BuildLabelSubjects(key, value)...)
	}

	return patterns
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

// labelSegmentRegex validates that each segment of a label key or value is NATS subject-safe.
var labelSegmentRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// ValidateLabel checks that a label key and value are valid for use in NATS subjects.
// Keys must be a single segment matching [a-zA-Z0-9_-]+.
// Values may be hierarchical (dot-separated), where each segment matches [a-zA-Z0-9_-]+.
func ValidateLabel(
	key, value string,
) error {
	if !labelSegmentRegex.MatchString(key) {
		return fmt.Errorf("invalid label key %q: must match [a-zA-Z0-9_-]+", key)
	}
	for _, segment := range strings.Split(value, ".") {
		if !labelSegmentRegex.MatchString(segment) {
			return fmt.Errorf(
				"invalid label value segment %q in %q: each segment must match [a-zA-Z0-9_-]+",
				segment,
				value,
			)
		}
	}
	return nil
}

// ParseTarget parses a --target value into routing components.
// Returns routingType ("host", "label", AnyHost, or BroadcastHost), key, and value.
// Label values may contain dots for hierarchical targeting (e.g., "group:web.dev.us-east").
func ParseTarget(
	target string,
) (routingType, key, value string) {
	switch {
	case target == AnyHost || target == BroadcastHost:
		return target, "", ""
	case strings.Contains(target, ":"):
		parts := strings.SplitN(target, ":", 2)
		return "label", parts[0], parts[1]
	default:
		return "host", target, ""
	}
}

// BuildSubjectFromTarget builds the full NATS subject for any target value.
// For label targets with hierarchical values (e.g., "group:web.dev"), each dot-separated
// segment becomes a subject token: jobs.query.label.group.web.dev
func BuildSubjectFromTarget(
	prefix, target string,
) string {
	rt, key, value := ParseTarget(target)
	switch rt {
	case AnyHost, BroadcastHost:
		return fmt.Sprintf("%s.%s", prefix, rt)
	case "label":
		return fmt.Sprintf("%s.label.%s.%s", prefix, key, value)
	default: // "host"
		return fmt.Sprintf("%s.host.%s", prefix, SanitizeHostname(key))
	}
}

// IsBroadcastTarget returns true if the target requires publishAndCollect
// (broadcast) semantics: _all or any key:value label target.
func IsBroadcastTarget(
	target string,
) bool {
	if target == BroadcastHost {
		return true
	}
	return strings.Contains(target, ":")
}

// BuildLabelSubjects builds subscription subjects for a label with hierarchical
// prefix matching. For a label "group: web.dev.us-east", it returns subjects
// for every prefix level:
//
//	jobs.*.label.group.web
//	jobs.*.label.group.web.dev
//	jobs.*.label.group.web.dev.us-east
//
// This enables targeting at any level of the hierarchy: --target group:web
// matches all workers whose group label starts with "web".
func BuildLabelSubjects(
	key, value string,
) []string {
	segments := strings.Split(value, ".")
	subjects := make([]string, 0, len(segments))
	for i := range segments {
		prefix := strings.Join(segments[:i+1], ".")
		subjects = append(subjects, fmt.Sprintf("%s.*.label.%s.%s", jobsBase, key, prefix))
	}
	return subjects
}

// ApplyNamespaceToInfraName prefixes an infrastructure name (stream, KV bucket)
// with the namespace. Returns the name unchanged if namespace is empty.
//
//	ApplyNamespaceToInfraName("", "JOBS")      -> "JOBS"
//	ApplyNamespaceToInfraName("osapi", "JOBS")  -> "osapi-JOBS"
func ApplyNamespaceToInfraName(
	namespace, name string,
) string {
	if namespace == "" {
		return name
	}
	return namespace + "-" + name
}

// ApplyNamespaceToSubjects prefixes a subject filter with the namespace.
// Returns the subject unchanged if namespace is empty.
//
//	ApplyNamespaceToSubjects("", "jobs.>")      -> "jobs.>"
//	ApplyNamespaceToSubjects("osapi", "jobs.>")  -> "osapi.jobs.>"
func ApplyNamespaceToSubjects(
	namespace, subjects string,
) string {
	if namespace == "" {
		return subjects
	}
	return namespace + "." + subjects
}
