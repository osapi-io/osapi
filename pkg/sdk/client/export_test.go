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

	"github.com/osapi-io/osapi/pkg/sdk/client/gen"
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

func ExportHostnameCollectionFromGen(
	input *gen.HostnameCollectionResponse,
) Collection[HostnameResult] {
	return hostnameCollectionFromGen(input)
}

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
