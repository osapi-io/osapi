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

package client

import (
	"log/slog"
	"net/http"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/retr0h/osapi/pkg/sdk/client/gen"
)

// ExportCheckError exposes the private checkError function for testing.
func ExportCheckError(
	statusCode int,
	responses ...*gen.ErrorResponse,
) error {
	return checkError(statusCode, responses...)
}

// ExportNewAuthTransport exposes the private authTransport constructor for
// testing.
func ExportNewAuthTransport(
	base http.RoundTripper,
	authHeader string,
	logger *slog.Logger,
) http.RoundTripper {
	return &authTransport{
		base:       base,
		authHeader: authHeader,
		logger:     logger,
	}
}

// ExportLoadAverageFromGen exposes the private loadAverageFromGen for testing.
func ExportLoadAverageFromGen(
	input *gen.LoadAverageResponse,
) *LoadAverage {
	return loadAverageFromGen(input)
}

// ExportMemoryFromGen exposes the private memoryFromGen for testing.
func ExportMemoryFromGen(
	input *gen.MemoryResponse,
) *Memory {
	return memoryFromGen(input)
}

// ExportOSInfoFromGen exposes the private osInfoFromGen for testing.
func ExportOSInfoFromGen(
	input *gen.OSInfoResponse,
) *OSInfo {
	return osInfoFromGen(input)
}

// ExportDisksFromGen exposes the private disksFromGen for testing.
func ExportDisksFromGen(
	input *gen.DisksResponse,
) []Disk {
	return disksFromGen(input)
}

// ExportDerefString exposes the private derefString for testing.
func ExportDerefString(
	s *string,
) string {
	return derefString(s)
}

// ExportDerefInt exposes the private derefInt for testing.
func ExportDerefInt(
	i *int,
) int {
	return derefInt(i)
}

// ExportDerefInt64 exposes the private derefInt64 for testing.
func ExportDerefInt64(
	i *int64,
) int64 {
	return derefInt64(i)
}

// ExportDerefFloat64 exposes the private derefFloat64 for testing.
func ExportDerefFloat64(
	f *float64,
) float64 {
	return derefFloat64(f)
}

// ExportDerefBool exposes the private derefBool for testing.
func ExportDerefBool(
	b *bool,
) bool {
	return derefBool(b)
}

// ExportJobIDFromGen exposes the private jobIDFromGen for testing.
func ExportJobIDFromGen(
	id *openapi_types.UUID,
) string {
	return jobIDFromGen(id)
}

// ExportHostnameCollectionFromGen exposes the private
// hostnameCollectionFromGen for testing.
func ExportHostnameCollectionFromGen(
	input *gen.HostnameCollectionResponse,
) Collection[HostnameResult] {
	return hostnameCollectionFromGen(input)
}

// ExportNodeStatusCollectionFromGen exposes the private
// nodeStatusCollectionFromGen for testing.
func ExportNodeStatusCollectionFromGen(
	input *gen.NodeStatusCollectionResponse,
) Collection[NodeStatus] {
	return nodeStatusCollectionFromGen(input)
}

// ExportDiskCollectionFromGen exposes the private diskCollectionFromGen for
// testing.
func ExportDiskCollectionFromGen(
	input *gen.DiskCollectionResponse,
) Collection[DiskResult] {
	return diskCollectionFromGen(input)
}

// ExportCommandCollectionFromGen exposes the private commandCollectionFromGen
// for testing.
func ExportCommandCollectionFromGen(
	input *gen.CommandResultCollectionResponse,
) Collection[CommandResult] {
	return commandCollectionFromGen(input)
}

// ExportDNSConfigCollectionFromGen exposes the private
// dnsConfigCollectionFromGen for testing.
func ExportDNSConfigCollectionFromGen(
	input *gen.DNSConfigCollectionResponse,
) Collection[DNSConfig] {
	return dnsConfigCollectionFromGen(input)
}

// ExportDNSUpdateCollectionFromGen exposes the private
// dnsUpdateCollectionFromGen for testing.
func ExportDNSUpdateCollectionFromGen(
	input *gen.DNSUpdateCollectionResponse,
) Collection[DNSUpdateResult] {
	return dnsUpdateCollectionFromGen(input)
}

// ExportHostnameUpdateCollectionFromGen exposes the private
// hostnameUpdateCollectionFromGen for testing.
func ExportHostnameUpdateCollectionFromGen(
	input *gen.HostnameUpdateCollectionResponse,
) Collection[HostnameUpdateResult] {
	return hostnameUpdateCollectionFromGen(input)
}

// ExportPingCollectionFromGen exposes the private pingCollectionFromGen for
// testing.
func ExportPingCollectionFromGen(
	input *gen.PingCollectionResponse,
) Collection[PingResult] {
	return pingCollectionFromGen(input)
}

// ExportDockerResultCollectionFromGen exposes the private
// dockerResultCollectionFromGen for testing.
func ExportDockerResultCollectionFromGen(
	input *gen.DockerResultCollectionResponse,
) Collection[DockerResult] {
	return dockerResultCollectionFromGen(input)
}

// ExportDockerListCollectionFromGen exposes the private
// dockerListCollectionFromGen for testing.
func ExportDockerListCollectionFromGen(
	input *gen.DockerListCollectionResponse,
) Collection[DockerListResult] {
	return dockerListCollectionFromGen(input)
}

// ExportDockerDetailCollectionFromGen exposes the private
// dockerDetailCollectionFromGen for testing.
func ExportDockerDetailCollectionFromGen(
	input *gen.DockerDetailCollectionResponse,
) Collection[DockerDetailResult] {
	return dockerDetailCollectionFromGen(input)
}

// ExportDockerActionCollectionFromGen exposes the private
// dockerActionCollectionFromGen for testing.
func ExportDockerActionCollectionFromGen(
	input *gen.DockerActionCollectionResponse,
) Collection[DockerActionResult] {
	return dockerActionCollectionFromGen(input)
}

// ExportDockerExecCollectionFromGen exposes the private
// dockerExecCollectionFromGen for testing.
func ExportDockerExecCollectionFromGen(
	input *gen.DockerExecCollectionResponse,
) Collection[DockerExecResult] {
	return dockerExecCollectionFromGen(input)
}

// ExportDockerPullCollectionFromGen exposes the private
// dockerPullCollectionFromGen for testing.
func ExportDockerPullCollectionFromGen(
	input *gen.DockerPullCollectionResponse,
) Collection[DockerPullResult] {
	return dockerPullCollectionFromGen(input)
}

// ExportAuditEntryFromGen exposes the private auditEntryFromGen for testing.
func ExportAuditEntryFromGen(
	input gen.AuditEntry,
) AuditEntry {
	return auditEntryFromGen(input)
}

// ExportAuditListFromGen exposes the private auditListFromGen for testing.
func ExportAuditListFromGen(
	input *gen.ListAuditResponse,
) AuditList {
	return auditListFromGen(input)
}

// ExportJobCreatedFromGen exposes the private jobCreatedFromGen for testing.
func ExportJobCreatedFromGen(
	input *gen.CreateJobResponse,
) JobCreated {
	return jobCreatedFromGen(input)
}

// ExportJobDetailFromGen exposes the private jobDetailFromGen for testing.
func ExportJobDetailFromGen(
	input *gen.JobDetailResponse,
) JobDetail {
	return jobDetailFromGen(input)
}

// ExportJobListFromGen exposes the private jobListFromGen for testing.
func ExportJobListFromGen(
	input *gen.ListJobsResponse,
) JobList {
	return jobListFromGen(input)
}

// ExportFileUploadFromGen exposes the private fileUploadFromGen for testing.
func ExportFileUploadFromGen(
	input *gen.FileUploadResponse,
) FileUpload {
	return fileUploadFromGen(input)
}

// ExportFileListFromGen exposes the private fileListFromGen for testing.
func ExportFileListFromGen(
	input *gen.FileListResponse,
) FileList {
	return fileListFromGen(input)
}

// ExportFileMetadataFromGen exposes the private fileMetadataFromGen for
// testing.
func ExportFileMetadataFromGen(
	input *gen.FileInfoResponse,
) FileMetadata {
	return fileMetadataFromGen(input)
}

// ExportFileDeleteFromGen exposes the private fileDeleteFromGen for testing.
func ExportFileDeleteFromGen(
	input *gen.FileDeleteResponse,
) FileDelete {
	return fileDeleteFromGen(input)
}

// ExportFileDeployCollectionFromGen exposes the private
// fileDeployCollectionFromGen for testing.
func ExportFileDeployCollectionFromGen(
	input *gen.FileDeployCollectionResponse,
) Collection[FileDeployResult] {
	return fileDeployCollectionFromGen(input)
}

// ExportFileUndeployCollectionFromGen exposes the private
// fileUndeployCollectionFromGen for testing.
func ExportFileUndeployCollectionFromGen(
	input *gen.FileUndeployCollectionResponse,
) Collection[FileUndeployResult] {
	return fileUndeployCollectionFromGen(input)
}

// ExportFileStatusCollectionFromGen exposes the private
// fileStatusCollectionFromGen for testing.
func ExportFileStatusCollectionFromGen(
	input *gen.FileStatusCollectionResponse,
) Collection[FileStatusResult] {
	return fileStatusCollectionFromGen(input)
}

// ExportHealthStatusFromGen exposes the private healthStatusFromGen for
// testing.
func ExportHealthStatusFromGen(
	input *gen.HealthResponse,
) HealthStatus {
	return healthStatusFromGen(input)
}

// ExportReadyStatusFromGen exposes the private readyStatusFromGen for testing.
func ExportReadyStatusFromGen(
	input *gen.ReadyResponse,
	serviceUnavailable bool,
) ReadyStatus {
	return readyStatusFromGen(input, serviceUnavailable)
}

// ExportSystemStatusFromGen exposes the private systemStatusFromGen for
// testing.
func ExportSystemStatusFromGen(
	input *gen.StatusResponse,
	serviceUnavailable bool,
) SystemStatus {
	return systemStatusFromGen(input, serviceUnavailable)
}

// ExportAgentFromGen exposes the private agentFromGen for testing.
func ExportAgentFromGen(
	input *gen.AgentInfo,
) Agent {
	return agentFromGen(input)
}

// ExportAgentListFromGen exposes the private agentListFromGen for testing.
func ExportAgentListFromGen(
	input *gen.ListAgentsResponse,
) AgentList {
	return agentListFromGen(input)
}
