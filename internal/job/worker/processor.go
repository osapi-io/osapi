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

package worker

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/shirou/gopsutil/v4/host"

	"github.com/retr0h/osapi/internal/exec"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/provider/network/dns"
	"github.com/retr0h/osapi/internal/provider/network/ping"
	"github.com/retr0h/osapi/internal/provider/system/disk"
	systemHost "github.com/retr0h/osapi/internal/provider/system/host"
	"github.com/retr0h/osapi/internal/provider/system/load"
	"github.com/retr0h/osapi/internal/provider/system/mem"
)

// handleJobMessage processes incoming job messages from NATS.
func (w *Worker) handleJobMessage(msg *nats.Msg) error {
	// Extract the key (job ID) from the message data
	jobKey := string(msg.Data)

	w.logger.Info(
		"received job notification",
		slog.String("subject", msg.Subject),
		slog.String("job_key", jobKey),
	)

	w.logger.Debug(
		"processing job message",
		slog.String("subject", msg.Subject),
		slog.String("job_key", jobKey),
		slog.String("raw_data", string(msg.Data)),
	)

	// Parse subject to extract category and operation
	prefix, _, category, operation, err := job.ParseSubject(msg.Subject)
	if err != nil {
		return fmt.Errorf("failed to parse subject %s: %w", msg.Subject, err)
	}

	// Try to find the job with different status prefixes using nats-client KV methods
	statuses := []string{"unprocessed", "processing", "completed", "failed"}
	var jobDataBytes []byte
	var currentStatus string
	var fullKey string

	for _, status := range statuses {
		key := status + "." + jobKey
		data, err := w.natsClient.KVGet(w.appConfig.Job.KVBucket, key)
		if err == nil {
			jobDataBytes = data
			currentStatus = status
			fullKey = key
			break
		}
		// Continue to next status if key not found
	}

	if jobDataBytes == nil {
		return fmt.Errorf("job not found: %s", jobKey)
	}

	// Parse the job data
	var jobData map[string]interface{}
	if err := json.Unmarshal(jobDataBytes, &jobData); err != nil {
		return fmt.Errorf("failed to parse job data: %w", err)
	}

	// Extract the RequestID from top-level job data
	requestID, ok := jobData["id"].(string)
	if !ok {
		return fmt.Errorf("invalid job format: missing id")
	}

	// Extract the operation data
	operationData, ok := jobData["operation"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid job format: missing operation")
	}

	// Convert operation data to job.Request and override category/operation from subject
	operationJSON, err := json.Marshal(operationData)
	if err != nil {
		return fmt.Errorf("failed to marshal operation data: %w", err)
	}

	var jobRequest job.Request
	if err := json.Unmarshal(operationJSON, &jobRequest); err != nil {
		return fmt.Errorf("failed to parse job request: %w", err)
	}

	// Set the RequestID and override category/operation from subject (subject is source of truth)
	jobRequest.RequestID = requestID
	jobRequest.Category = category
	jobRequest.Operation = operation

	// Determine Type from subject prefix (jobs.query vs jobs.modify)
	if strings.HasPrefix(prefix, "jobs.query") {
		jobRequest.Type = job.TypeQuery
	} else if strings.HasPrefix(prefix, "jobs.modify") {
		jobRequest.Type = job.TypeModify
	}

	// Process the job
	w.logger.Info(
		"processing job",
		slog.String("job_id", jobKey),
		slog.String("type", string(jobRequest.Type)),
		slog.String("category", jobRequest.Category),
		slog.String("operation", jobRequest.Operation),
	)

	// Create job response
	response := job.Response{
		RequestID: jobRequest.RequestID,
		Status:    job.StatusProcessing,
		Timestamp: time.Now(),
	}

	// Mark job as processing in KV store immediately
	if currentStatus == "unprocessed" {
		jobData["status"] = "processing"
		jobData["updated_at"] = time.Now().Format(time.RFC3339)

		// Add status transition tracking
		if statusHistory, ok := jobData["status_history"].([]interface{}); ok {
			jobData["status_history"] = append(statusHistory, map[string]interface{}{
				"status":    "processing",
				"timestamp": time.Now().Format(time.RFC3339),
			})
		} else {
			// Initialize status history if not present
			jobData["status_history"] = []interface{}{
				map[string]interface{}{
					"status":    "unprocessed",
					"timestamp": jobData["created"], // Original creation time
				},
				map[string]interface{}{
					"status":    "processing",
					"timestamp": time.Now().Format(time.RFC3339),
				},
			}
		}

		processingJSON, err := json.Marshal(jobData)
		if err != nil {
			return fmt.Errorf("failed to marshal processing status: %w", err)
		}

		// Create new key FIRST for safety
		processingKey := "processing." + jobKey
		err = w.natsClient.KVPut(w.appConfig.Job.KVBucket, processingKey, processingJSON)
		if err != nil {
			return fmt.Errorf("failed to create processing key: %w", err)
		}

		// Only delete old key after successful creation
		err = w.natsClient.KVDelete(w.appConfig.Job.KVBucket, fullKey)
		if err != nil {
			// Log error but continue - we have the new key
			w.logger.Error("failed to delete old unprocessed key",
				slog.String("old_key", fullKey),
				slog.String("error", err.Error()),
			)
		}

		// Update tracking variables
		currentStatus = "processing"
		fullKey = processingKey

		w.logger.Debug("marked job as processing",
			slog.String("job_id", jobKey),
			slog.String("created", jobData["created"].(string)), // Show we preserved original
		)
	}

	// Process based on category and operation
	result, err := w.processJobOperation(jobRequest)
	if err != nil {
		w.logger.Error(
			"job processing failed",
			slog.String("job_id", jobKey),
			slog.String("category", jobRequest.Category),
			slog.String("operation", jobRequest.Operation),
			slog.String("error", err.Error()),
		)
		response.Status = job.StatusFailed
		response.Error = err.Error()
	} else {
		response.Status = job.StatusCompleted
		response.Data = result
	}

	response.Timestamp = time.Now()

	// Store response in KV bucket
	responseJSON, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	// NATS KV keys must be valid NATS subject tokens (alphanumeric + underscores only)
	responseKey := sanitizeKeyForNATS(jobRequest.RequestID)
	w.logger.Debug(
		"storing response",
		slog.String("original_request_id", jobRequest.RequestID),
		slog.String("sanitized_key", responseKey),
	)
	err = w.natsClient.KVPut(w.appConfig.Job.KVResponseBucket, responseKey, responseJSON)
	if err != nil {
		return fmt.Errorf("failed to store response with key %s: %w", responseKey, err)
	}

	// Update job status in original KV bucket
	jobData["status"] = string(response.Status)
	if response.Error != "" {
		jobData["error"] = response.Error
	}
	jobData["updated_at"] = time.Now().Format(time.RFC3339)

	// Include result data in the original job entry for retrieval by UUID
	if response.Data != nil {
		jobData["result"] = response.Data
	}

	updatedJobJSON, err := json.Marshal(jobData)
	if err != nil {
		return fmt.Errorf("failed to marshal updated job: %w", err)
	}

	// Add status transition to history
	if statusHistory, ok := jobData["status_history"].([]interface{}); ok {
		jobData["status_history"] = append(statusHistory, map[string]interface{}{
			"status":    string(response.Status),
			"timestamp": time.Now().Format(time.RFC3339),
		})
	} else {
		// This shouldn't happen if processing was tracked, but handle it
		jobData["status_history"] = []interface{}{
			map[string]interface{}{
				"status":    string(response.Status),
				"timestamp": time.Now().Format(time.RFC3339),
			},
		}
	}

	// Re-marshal with status history
	updatedJobJSON, err = json.Marshal(jobData)
	if err != nil {
		return fmt.Errorf("failed to marshal job with history: %w", err)
	}

	// Move job to new status prefix if status changed
	newStatus := string(response.Status)
	if currentStatus != newStatus {
		// Create new key FIRST for safety
		newKey := newStatus + "." + jobKey
		err = w.natsClient.KVPut(w.appConfig.Job.KVBucket, newKey, updatedJobJSON)
		if err != nil {
			return fmt.Errorf("failed to create new status key: %w", err)
		}

		// Only delete old key after successful creation
		err = w.natsClient.KVDelete(w.appConfig.Job.KVBucket, fullKey)
		if err != nil {
			// Log error but continue - we have the new key
			w.logger.Error("failed to delete old status key",
				slog.String("old_key", fullKey),
				slog.String("error", err.Error()),
			)
		}

		w.logger.Debug("moved job to new status prefix",
			slog.String("job_id", jobKey),
			slog.String("old_status", currentStatus),
			slog.String("new_status", newStatus),
			slog.String("original_created", jobData["created"].(string)),
		)
	} else {
		// Status unchanged, just update in place
		err = w.natsClient.KVPut(w.appConfig.Job.KVBucket, fullKey, updatedJobJSON)
		if err != nil {
			return fmt.Errorf("failed to update job: %w", err)
		}
	}

	w.logger.Info(
		"job processing completed",
		slog.String("job_id", jobKey),
		slog.String("status", string(response.Status)),
	)

	// Return error if job failed so message won't be acknowledged and will retry
	if response.Status == job.StatusFailed {
		return fmt.Errorf("job processing failed: %s", response.Error)
	}

	return nil
}

// processJobOperation handles the actual job processing based on category and operation.
func (w *Worker) processJobOperation(jobRequest job.Request) (json.RawMessage, error) {
	switch jobRequest.Category {
	case "system":
		return w.processSystemOperation(jobRequest)
	case "network":
		return w.processNetworkOperation(jobRequest)
	default:
		return nil, fmt.Errorf("unsupported job category: %s", jobRequest.Category)
	}
}

// processSystemOperation handles system-related operations.
func (w *Worker) processSystemOperation(jobRequest job.Request) (json.RawMessage, error) {
	// Extract base operation from dotted operation (e.g., "hostname.get" -> "hostname")
	baseOperation := strings.Split(jobRequest.Operation, ".")[0]

	switch baseOperation {
	case "hostname":
		return w.getSystemHostname()
	case "status":
		return w.getSystemStatus()
	case "uptime":
		return w.getSystemUptime()
	case "os", "osinfo":
		return w.getSystemOSInfo()
	case "disk":
		return w.getSystemDisk()
	case "memory", "mem":
		return w.getSystemMemory()
	case "load":
		return w.getSystemLoad()
	default:
		return nil, fmt.Errorf("unsupported system operation: %s", jobRequest.Operation)
	}
}

// processNetworkOperation handles network-related operations.
func (w *Worker) processNetworkOperation(jobRequest job.Request) (json.RawMessage, error) {
	// Extract base operation from dotted operation (e.g., "dns.get" -> "dns")
	baseOperation := strings.Split(jobRequest.Operation, ".")[0]

	switch baseOperation {
	case "dns":
		return w.processNetworkDNS(jobRequest)
	case "ping":
		return w.processNetworkPing(jobRequest)
	default:
		return nil, fmt.Errorf("unsupported network operation: %s", jobRequest.Operation)
	}
}

// getSystemHostname retrieves the system hostname.
func (w *Worker) getSystemHostname() (json.RawMessage, error) {
	hostProvider := w.getHostProvider()
	hostname, err := hostProvider.GetHostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}

	result := map[string]interface{}{
		"hostname": hostname,
	}

	return json.Marshal(result)
}

// getSystemStatus retrieves comprehensive system status.
func (w *Worker) getSystemStatus() (json.RawMessage, error) {
	hostProvider := w.getHostProvider()
	diskProvider := w.getDiskProvider()
	memProvider := w.getMemProvider()
	loadProvider := w.getLoadProvider()

	// Get all system information
	hostname, _ := hostProvider.GetHostname()
	osInfo, _ := hostProvider.GetOSInfo()
	uptime, _ := hostProvider.GetUptime()
	diskUsage, _ := diskProvider.GetLocalUsageStats()
	memInfo, _ := memProvider.GetStats()
	loadAvg, _ := loadProvider.GetAverageStats()

	result := map[string]interface{}{
		"hostname": hostname,
		"os":       osInfo,
		"uptime":   uptime,
		"disk":     diskUsage,
		"memory":   memInfo,
		"load":     loadAvg,
	}

	return json.Marshal(result)
}

// getSystemUptime retrieves the system uptime.
func (w *Worker) getSystemUptime() (json.RawMessage, error) {
	hostProvider := w.getHostProvider()
	uptime, err := hostProvider.GetUptime()
	if err != nil {
		return nil, fmt.Errorf("failed to get uptime: %w", err)
	}

	result := map[string]interface{}{
		"uptime_seconds": uptime.Seconds(),
		"uptime":         uptime.String(),
	}

	return json.Marshal(result)
}

// getSystemOSInfo retrieves the operating system information.
func (w *Worker) getSystemOSInfo() (json.RawMessage, error) {
	hostProvider := w.getHostProvider()
	osInfo, err := hostProvider.GetOSInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get OS info: %w", err)
	}

	return json.Marshal(osInfo)
}

// getSystemDisk retrieves disk usage statistics.
func (w *Worker) getSystemDisk() (json.RawMessage, error) {
	diskProvider := w.getDiskProvider()
	diskUsage, err := diskProvider.GetLocalUsageStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get disk usage: %w", err)
	}

	result := map[string]interface{}{
		"disks": diskUsage,
	}

	return json.Marshal(result)
}

// getSystemMemory retrieves memory statistics.
func (w *Worker) getSystemMemory() (json.RawMessage, error) {
	memProvider := w.getMemProvider()
	memInfo, err := memProvider.GetStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory stats: %w", err)
	}

	return json.Marshal(memInfo)
}

// getSystemLoad retrieves load average statistics.
func (w *Worker) getSystemLoad() (json.RawMessage, error) {
	loadProvider := w.getLoadProvider()
	loadAvg, err := loadProvider.GetAverageStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get load averages: %w", err)
	}

	return json.Marshal(loadAvg)
}

// processNetworkDNS handles DNS configuration operations.
func (w *Worker) processNetworkDNS(jobRequest job.Request) (json.RawMessage, error) {
	var dnsData map[string]interface{}
	if err := json.Unmarshal(jobRequest.Data, &dnsData); err != nil {
		return nil, fmt.Errorf("failed to parse DNS data: %w", err)
	}

	if jobRequest.Type == job.TypeQuery {
		// Get DNS configuration
		interfaceName, _ := dnsData["interface"].(string)
		if interfaceName == "" {
			interfaceName = "eth0" // Default interface
		}

		dnsProvider := w.getDNSProvider()
		config, err := dnsProvider.GetResolvConfByInterface(interfaceName)
		if err != nil {
			return nil, fmt.Errorf("failed to get DNS config: %w", err)
		}

		return json.Marshal(config)
	} else {
		// Set DNS configuration
		servers, _ := dnsData["servers"].([]interface{})
		searchDomains, _ := dnsData["search_domains"].([]interface{})
		interfaceName, _ := dnsData["interface"].(string)

		var serverStrings []string
		for _, s := range servers {
			if str, ok := s.(string); ok {
				serverStrings = append(serverStrings, str)
			}
		}

		var searchStrings []string
		for _, s := range searchDomains {
			if str, ok := s.(string); ok {
				searchStrings = append(searchStrings, str)
			}
		}

		dnsProvider := w.getDNSProvider()
		err := dnsProvider.UpdateResolvConfByInterface(serverStrings, searchStrings, interfaceName)
		if err != nil {
			return nil, fmt.Errorf("failed to set DNS config: %w", err)
		}

		result := map[string]interface{}{
			"success": true,
			"message": "DNS configuration updated successfully",
		}

		return json.Marshal(result)
	}
}

// processNetworkPing handles ping operations.
func (w *Worker) processNetworkPing(jobRequest job.Request) (json.RawMessage, error) {
	var pingData map[string]interface{}
	if err := json.Unmarshal(jobRequest.Data, &pingData); err != nil {
		return nil, fmt.Errorf("failed to parse ping data: %w", err)
	}

	address, ok := pingData["address"].(string)
	if !ok {
		return nil, fmt.Errorf("missing ping address")
	}

	pingProvider := w.getPingProvider()
	result, err := pingProvider.Do(address)
	if err != nil {
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	return json.Marshal(result)
}

// Provider factory methods following the existing pattern
func (w *Worker) getHostProvider() systemHost.Provider {
	var hostProvider systemHost.Provider

	info, _ := host.Info()

	switch strings.ToLower(info.Platform) {
	case "ubuntu":
		hostProvider = systemHost.NewUbuntuProvider()
	default:
		hostProvider = systemHost.NewLinuxProvider()
	}

	return hostProvider
}

func (w *Worker) getDiskProvider() disk.Provider {
	var diskProvider disk.Provider

	info, _ := host.Info()

	switch strings.ToLower(info.Platform) {
	case "ubuntu":
		diskProvider = disk.NewUbuntuProvider(w.logger)
	default:
		diskProvider = disk.NewLinuxProvider()
	}

	return diskProvider
}

func (w *Worker) getMemProvider() mem.Provider {
	var memProvider mem.Provider

	info, _ := host.Info()

	switch strings.ToLower(info.Platform) {
	case "ubuntu":
		memProvider = mem.NewUbuntuProvider()
	default:
		memProvider = mem.NewLinuxProvider()
	}

	return memProvider
}

func (w *Worker) getLoadProvider() load.Provider {
	var loadProvider load.Provider

	info, _ := host.Info()

	switch strings.ToLower(info.Platform) {
	case "ubuntu":
		loadProvider = load.NewUbuntuProvider()
	default:
		loadProvider = load.NewLinuxProvider()
	}

	return loadProvider
}

func (w *Worker) getDNSProvider() dns.Provider {
	var dnsProvider dns.Provider
	var execManager exec.Manager

	info, _ := host.Info()
	execManager = exec.New(w.logger)

	switch strings.ToLower(info.Platform) {
	case "ubuntu":
		dnsProvider = dns.NewUbuntuProvider(w.logger, execManager)
	default:
		dnsProvider = dns.NewLinuxProvider()
	}

	return dnsProvider
}

func (w *Worker) getPingProvider() ping.Provider {
	var pingProvider ping.Provider

	info, _ := host.Info()

	switch strings.ToLower(info.Platform) {
	case "ubuntu":
		pingProvider = ping.NewUbuntuProvider()
	default:
		pingProvider = ping.NewLinuxProvider()
	}

	return pingProvider
}

// sanitizeKeyForNATS converts a string to a valid NATS KV key.
// NATS KV keys must be valid NATS subject tokens (alphanumeric and underscores only).
func sanitizeKeyForNATS(key string) string {
	// Replace any non-alphanumeric characters (except underscores) with underscores
	reg := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	return reg.ReplaceAllString(key, "_")
}
