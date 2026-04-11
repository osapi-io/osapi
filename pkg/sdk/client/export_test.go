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

// ExportStaleDeploymentFromGen exposes the private staleDeploymentFromGen for
// testing.
func ExportStaleDeploymentFromGen(
	input gen.StaleDeployment,
) StaleDeployment {
	return staleDeploymentFromGen(input)
}

// ExportStaleListFromGen exposes the private staleListFromGen for testing.
func ExportStaleListFromGen(
	input *gen.StaleDeploymentsResponse,
) StaleList {
	return staleListFromGen(input)
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

// ExportPendingAgentListFromGen exposes the private pendingAgentListFromGen
// for testing.
func ExportPendingAgentListFromGen(
	input *gen.ListPendingAgentsResponse,
) PendingAgentList {
	return pendingAgentListFromGen(input)
}

// SysctlEntryCollectionFromGen exposes the private
// sysctlEntryCollectionFromGen for testing.
func SysctlEntryCollectionFromGen(
	input *gen.SysctlCollectionResponse,
) Collection[SysctlEntryResult] {
	return sysctlEntryCollectionFromGen(input)
}

// SysctlEntryCollectionFromGet exposes the private
// sysctlEntryCollectionFromGet for testing.
func SysctlEntryCollectionFromGet(
	input *gen.SysctlGetResponse,
) Collection[SysctlEntryResult] {
	return sysctlEntryCollectionFromGet(input)
}

// SysctlMutationCollectionFromCreate exposes the private
// sysctlMutationCollectionFromCreate for testing.
func SysctlMutationCollectionFromCreate(
	input *gen.SysctlCreateResponse,
) Collection[SysctlMutationResult] {
	return sysctlMutationCollectionFromCreate(input)
}

// SysctlMutationCollectionFromUpdate exposes the private
// sysctlMutationCollectionFromUpdate for testing.
func SysctlMutationCollectionFromUpdate(
	input *gen.SysctlUpdateResponse,
) Collection[SysctlMutationResult] {
	return sysctlMutationCollectionFromUpdate(input)
}

// SysctlMutationCollectionFromDelete exposes the private
// sysctlMutationCollectionFromDelete for testing.
func SysctlMutationCollectionFromDelete(
	input *gen.SysctlDeleteResponse,
) Collection[SysctlMutationResult] {
	return sysctlMutationCollectionFromDelete(input)
}

// NtpStatusCollectionFromGen exposes the private
// ntpStatusCollectionFromGen for testing.
func NtpStatusCollectionFromGen(
	input *gen.NtpCollectionResponse,
) Collection[NtpStatusResult] {
	return ntpStatusCollectionFromGen(input)
}

// NtpMutationCollectionFromCreate exposes the private
// ntpMutationCollectionFromCreate for testing.
func NtpMutationCollectionFromCreate(
	input *gen.NtpCreateResponse,
) Collection[NtpMutationResult] {
	return ntpMutationCollectionFromCreate(input)
}

// NtpMutationCollectionFromUpdate exposes the private
// ntpMutationCollectionFromUpdate for testing.
func NtpMutationCollectionFromUpdate(
	input *gen.NtpUpdateResponse,
) Collection[NtpMutationResult] {
	return ntpMutationCollectionFromUpdate(input)
}

// NtpMutationCollectionFromDelete exposes the private
// ntpMutationCollectionFromDelete for testing.
func NtpMutationCollectionFromDelete(
	input *gen.NtpDeleteResponse,
) Collection[NtpMutationResult] {
	return ntpMutationCollectionFromDelete(input)
}

// ExportTimezoneCollectionFromGen exposes the private
// timezoneCollectionFromGen for testing.
func ExportTimezoneCollectionFromGen(
	input *gen.TimezoneCollectionResponse,
) Collection[TimezoneResult] {
	return timezoneCollectionFromGen(input)
}

// ExportTimezoneMutationCollectionFromUpdate exposes the private
// timezoneMutationCollectionFromUpdate for testing.
func ExportTimezoneMutationCollectionFromUpdate(
	input *gen.TimezoneUpdateResponse,
) Collection[TimezoneMutationResult] {
	return timezoneMutationCollectionFromUpdate(input)
}

// PowerCollectionFromReboot exposes the private
// powerCollectionFromReboot for testing.
func PowerCollectionFromReboot(
	input *gen.PowerRebootResponse,
) Collection[PowerResult] {
	return powerCollectionFromReboot(input)
}

// PowerCollectionFromShutdown exposes the private
// powerCollectionFromShutdown for testing.
func PowerCollectionFromShutdown(
	input *gen.PowerShutdownResponse,
) Collection[PowerResult] {
	return powerCollectionFromShutdown(input)
}

// ExportDerefFloat32 exposes the private derefFloat32 for testing.
func ExportDerefFloat32(
	f *float32,
) float32 {
	return derefFloat32(f)
}

// ProcessInfoCollectionFromList exposes the private
// processInfoCollectionFromList for testing.
func ProcessInfoCollectionFromList(
	input *gen.ProcessCollectionResponse,
) Collection[ProcessInfoResult] {
	return processInfoCollectionFromList(input)
}

// ProcessInfoCollectionFromGet exposes the private
// processInfoCollectionFromGet for testing.
func ProcessInfoCollectionFromGet(
	input *gen.ProcessGetResponse,
) Collection[ProcessInfoResult] {
	return processInfoCollectionFromGet(input)
}

// ProcessSignalCollectionFromGen exposes the private
// processSignalCollectionFromGen for testing.
func ProcessSignalCollectionFromGen(
	input *gen.ProcessSignalResponse,
) Collection[ProcessSignalResult] {
	return processSignalCollectionFromGen(input)
}

// ExportDerefStringSlice exposes the private derefStringSlice for testing.
func ExportDerefStringSlice(
	s *[]string,
) []string {
	return derefStringSlice(s)
}

// UserInfoCollectionFromList exposes the private
// userInfoCollectionFromList for testing.
func UserInfoCollectionFromList(
	input *gen.UserCollectionResponse,
) Collection[UserInfoResult] {
	return userInfoCollectionFromList(input)
}

// UserInfoCollectionFromGet exposes the private
// userInfoCollectionFromGet for testing.
func UserInfoCollectionFromGet(
	input *gen.UserCollectionResponse,
) Collection[UserInfoResult] {
	return userInfoCollectionFromGet(input)
}

// UserMutationCollectionFromCreate exposes the private
// userMutationCollectionFromCreate for testing.
func UserMutationCollectionFromCreate(
	input *gen.UserMutationResponse,
) Collection[UserMutationResult] {
	return userMutationCollectionFromCreate(input)
}

// UserMutationCollectionFromUpdate exposes the private
// userMutationCollectionFromUpdate for testing.
func UserMutationCollectionFromUpdate(
	input *gen.UserMutationResponse,
) Collection[UserMutationResult] {
	return userMutationCollectionFromUpdate(input)
}

// UserMutationCollectionFromDelete exposes the private
// userMutationCollectionFromDelete for testing.
func UserMutationCollectionFromDelete(
	input *gen.UserMutationResponse,
) Collection[UserMutationResult] {
	return userMutationCollectionFromDelete(input)
}

// UserMutationCollectionFromPassword exposes the private
// userMutationCollectionFromPassword for testing.
func UserMutationCollectionFromPassword(
	input *gen.UserMutationResponse,
) Collection[UserMutationResult] {
	return userMutationCollectionFromPassword(input)
}

// SSHKeyCollectionFromGen exposes the private
// sshKeyCollectionFromGen for testing.
func SSHKeyCollectionFromGen(
	input *gen.SSHKeyCollectionResponse,
) Collection[SSHKeyInfoResult] {
	return sshKeyCollectionFromGen(input)
}

// SSHKeyInfoResultFromGen exposes the private
// sshKeyInfoResultFromGen for testing.
func SSHKeyInfoResultFromGen(
	input gen.SSHKeyEntry,
) SSHKeyInfoResult {
	return sshKeyInfoResultFromGen(input)
}

// SSHKeyInfoFromGen exposes the private
// sshKeyInfoFromGen for testing.
func SSHKeyInfoFromGen(
	input gen.SSHKeyInfo,
) SSHKeyInfo {
	return sshKeyInfoFromGen(input)
}

// SSHKeyMutationCollectionFromGen exposes the private
// sshKeyMutationCollectionFromGen for testing.
func SSHKeyMutationCollectionFromGen(
	input *gen.SSHKeyMutationResponse,
) Collection[SSHKeyMutationResult] {
	return sshKeyMutationCollectionFromGen(input)
}

// SSHKeyMutationResultFromGen exposes the private
// sshKeyMutationResultFromGen for testing.
func SSHKeyMutationResultFromGen(
	input gen.SSHKeyMutationEntry,
) SSHKeyMutationResult {
	return sshKeyMutationResultFromGen(input)
}

// GroupInfoCollectionFromList exposes the private
// groupInfoCollectionFromList for testing.
func GroupInfoCollectionFromList(
	input *gen.GroupCollectionResponse,
) Collection[GroupInfoResult] {
	return groupInfoCollectionFromList(input)
}

// GroupInfoCollectionFromGet exposes the private
// groupInfoCollectionFromGet for testing.
func GroupInfoCollectionFromGet(
	input *gen.GroupCollectionResponse,
) Collection[GroupInfoResult] {
	return groupInfoCollectionFromGet(input)
}

// GroupMutationCollectionFromCreate exposes the private
// groupMutationCollectionFromCreate for testing.
func GroupMutationCollectionFromCreate(
	input *gen.GroupMutationResponse,
) Collection[GroupMutationResult] {
	return groupMutationCollectionFromCreate(input)
}

// GroupMutationCollectionFromUpdate exposes the private
// groupMutationCollectionFromUpdate for testing.
func GroupMutationCollectionFromUpdate(
	input *gen.GroupMutationResponse,
) Collection[GroupMutationResult] {
	return groupMutationCollectionFromUpdate(input)
}

// GroupMutationCollectionFromDelete exposes the private
// groupMutationCollectionFromDelete for testing.
func GroupMutationCollectionFromDelete(
	input *gen.GroupMutationResponse,
) Collection[GroupMutationResult] {
	return groupMutationCollectionFromDelete(input)
}

// PackageInfoCollectionFromList exposes the private
// packageInfoCollectionFromList for testing.
func PackageInfoCollectionFromList(
	input *gen.PackageCollectionResponse,
) Collection[PackageInfoResult] {
	return packageInfoCollectionFromList(input)
}

// PackageInfoCollectionFromGet exposes the private
// packageInfoCollectionFromGet for testing.
func PackageInfoCollectionFromGet(
	input *gen.PackageCollectionResponse,
) Collection[PackageInfoResult] {
	return packageInfoCollectionFromGet(input)
}

// PackageMutationCollectionFromInstall exposes the private
// packageMutationCollectionFromInstall for testing.
func PackageMutationCollectionFromInstall(
	input *gen.PackageMutationResponse,
) Collection[PackageMutationResult] {
	return packageMutationCollectionFromInstall(input)
}

// PackageMutationCollectionFromRemove exposes the private
// packageMutationCollectionFromRemove for testing.
func PackageMutationCollectionFromRemove(
	input *gen.PackageMutationResponse,
) Collection[PackageMutationResult] {
	return packageMutationCollectionFromRemove(input)
}

// PackageMutationCollectionFromUpdate exposes the private
// packageMutationCollectionFromUpdate for testing.
func PackageMutationCollectionFromUpdate(
	input *gen.PackageMutationResponse,
) Collection[PackageMutationResult] {
	return packageMutationCollectionFromUpdate(input)
}

// PackageUpdateCollectionFromGen exposes the private
// packageUpdateCollectionFromGen for testing.
func PackageUpdateCollectionFromGen(
	input *gen.UpdateCollectionResponse,
) Collection[PackageUpdateResult] {
	return packageUpdateCollectionFromGen(input)
}

// ExportPackageInfosFromGen exposes the private packageInfosFromGen
// for testing.
func ExportPackageInfosFromGen(
	input *[]gen.PackageInfo,
) []PackageInfo {
	return packageInfosFromGen(input)
}

// ExportUpdateInfosFromGen exposes the private updateInfosFromGen
// for testing.
func ExportUpdateInfosFromGen(
	input *[]gen.UpdateInfo,
) []UpdateInfo {
	return updateInfosFromGen(input)
}

// CertificateCACollectionFromGen exposes the private
// certificateCACollectionFromGen for testing.
func CertificateCACollectionFromGen(
	input *gen.CertificateCACollectionResponse,
) Collection[CertificateCAResult] {
	return certificateCACollectionFromGen(input)
}

// CertificateCAInfoFromGen exposes the private
// certificateCAInfoFromGen for testing.
func CertificateCAInfoFromGen(
	input gen.CertificateCAInfo,
) CertificateCA {
	return certificateCAInfoFromGen(input)
}

// CertificateCAMutationCollectionFromGen exposes the private
// certificateCAMutationCollectionFromGen for testing.
func CertificateCAMutationCollectionFromGen(
	input *gen.CertificateCAMutationResponse,
) Collection[CertificateCAMutationResult] {
	return certificateCAMutationCollectionFromGen(input)
}

// ServiceListCollectionFromGen exposes the private
// serviceListCollectionFromGen for testing.
func ServiceListCollectionFromGen(
	input *gen.ServiceListResponse,
) Collection[ServiceInfoResult] {
	return serviceListCollectionFromGen(input)
}

// ServiceInfoFromGen exposes the private serviceInfoFromGen for testing.
func ServiceInfoFromGen(
	input gen.ServiceInfo,
) ServiceInfo {
	return serviceInfoFromGen(input)
}

// ServiceGetCollectionFromGen exposes the private
// serviceGetCollectionFromGen for testing.
func ServiceGetCollectionFromGen(
	input *gen.ServiceGetResponse,
) Collection[ServiceGetResult] {
	return serviceGetCollectionFromGen(input)
}

// ServiceMutationCollectionFromGen exposes the private
// serviceMutationCollectionFromGen for testing.
func ServiceMutationCollectionFromGen(
	input *gen.ServiceMutationResponse,
) Collection[ServiceMutationResult] {
	return serviceMutationCollectionFromGen(input)
}

// LogCollectionFromGen exposes the private logCollectionFromGen for testing.
func LogCollectionFromGen(
	input *gen.LogCollectionResponse,
) Collection[LogEntryResult] {
	return logCollectionFromGen(input)
}

// LogEntryInfoFromGen exposes the private logEntryInfoFromGen for testing.
func LogEntryInfoFromGen(
	input gen.LogEntryInfo,
) LogEntry {
	return logEntryInfoFromGen(input)
}

// ExportDNSDeleteCollectionFromGen exposes the private
// dnsDeleteCollectionFromGen for testing.
func ExportDNSDeleteCollectionFromGen(
	input *gen.DNSDeleteCollectionResponse,
) Collection[DNSDeleteResult] {
	return dnsDeleteCollectionFromGen(input)
}

// InterfaceInfoFromGen exposes the private interfaceInfoFromGen for testing.
func InterfaceInfoFromGen(
	input gen.InterfaceInfo,
) InterfaceInfo {
	return interfaceInfoFromGen(input)
}

// InterfaceListCollectionFromGen exposes the private
// interfaceListCollectionFromGen for testing.
func InterfaceListCollectionFromGen(
	input *gen.InterfaceListResponse,
) Collection[InterfaceListResult] {
	return interfaceListCollectionFromGen(input)
}

// InterfaceGetCollectionFromGen exposes the private
// interfaceGetCollectionFromGen for testing.
func InterfaceGetCollectionFromGen(
	input *gen.InterfaceGetResponse,
) Collection[InterfaceGetResult] {
	return interfaceGetCollectionFromGen(input)
}

// InterfaceMutationCollectionFromCreate exposes the private
// interfaceMutationCollectionFromCreate for testing.
func InterfaceMutationCollectionFromCreate(
	input *gen.InterfaceMutationResponse,
) Collection[InterfaceMutationResult] {
	return interfaceMutationCollectionFromCreate(input)
}

// InterfaceMutationCollectionFromUpdate exposes the private
// interfaceMutationCollectionFromUpdate for testing.
func InterfaceMutationCollectionFromUpdate(
	input *gen.InterfaceMutationResponse,
) Collection[InterfaceMutationResult] {
	return interfaceMutationCollectionFromUpdate(input)
}

// InterfaceMutationCollectionFromDelete exposes the private
// interfaceMutationCollectionFromDelete for testing.
func InterfaceMutationCollectionFromDelete(
	input *gen.InterfaceMutationResponse,
) Collection[InterfaceMutationResult] {
	return interfaceMutationCollectionFromDelete(input)
}

// RouteInfoFromGen exposes the private routeInfoFromGen for testing.
func RouteInfoFromGen(
	input gen.RouteInfo,
) RouteInfo {
	return routeInfoFromGen(input)
}

// RouteListCollectionFromGen exposes the private
// routeListCollectionFromGen for testing.
func RouteListCollectionFromGen(
	input *gen.RouteListResponse,
) Collection[RouteListResult] {
	return routeListCollectionFromGen(input)
}

// RouteGetCollectionFromGen exposes the private
// routeGetCollectionFromGen for testing.
func RouteGetCollectionFromGen(
	input *gen.RouteGetResponse,
) Collection[RouteGetResult] {
	return routeGetCollectionFromGen(input)
}

// RouteMutationCollectionFromCreate exposes the private
// routeMutationCollectionFromCreate for testing.
func RouteMutationCollectionFromCreate(
	input *gen.RouteMutationResponse,
) Collection[RouteMutationResult] {
	return routeMutationCollectionFromCreate(input)
}

// RouteMutationCollectionFromUpdate exposes the private
// routeMutationCollectionFromUpdate for testing.
func RouteMutationCollectionFromUpdate(
	input *gen.RouteMutationResponse,
) Collection[RouteMutationResult] {
	return routeMutationCollectionFromUpdate(input)
}

// RouteMutationCollectionFromDelete exposes the private
// routeMutationCollectionFromDelete for testing.
func RouteMutationCollectionFromDelete(
	input *gen.RouteMutationResponse,
) Collection[RouteMutationResult] {
	return routeMutationCollectionFromDelete(input)
}
