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

func ExportHostnameCollectionFromGen(
	input *gen.HostnameGetResponse,
) Collection[HostnameResult] {
	return hostnameCollectionFromGen(input)
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
) SystemStatus {
	return systemStatusFromGen(input)
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
	input *gen.NtpGetResponse,
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

// ProcessInfoCollectionFromGen exposes the private
// processInfoCollectionFromGen for testing.
var ProcessInfoCollectionFromGen = processInfoCollectionFromGen

// ProcessSignalCollectionFromGen exposes the private
// processSignalCollectionFromGen for testing.
var ProcessSignalCollectionFromGen = processSignalCollectionFromGen

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

// GroupInfoCollectionFromGen exposes the private
// groupInfoCollectionFromGen for testing.
var GroupInfoCollectionFromGen = groupInfoCollectionFromGen

// GroupMutationCollectionFromGen exposes the private
// groupMutationCollectionFromGen for testing.
var GroupMutationCollectionFromGen = groupMutationCollectionFromGen

// CertificateInfoCollectionFromGen exposes the private
// certificateInfoCollectionFromGen for testing.
var CertificateInfoCollectionFromGen = certificateInfoCollectionFromGen

// CertificateMutationCollectionFromGen exposes the private
// certificateMutationCollectionFromGen for testing.
var CertificateMutationCollectionFromGen = certificateMutationCollectionFromGen

// ExportNewResponse exposes newResponse for testing.
func ExportNewResponse[T any](
	data T,
	rawJSON []byte,
	httpResp *http.Response,
) *Response[T] {
	return newResponse(data, rawJSON, httpResp)
}

// ExportDiskListFromGen exposes the private diskListFromGen for testing.
func ExportDiskListFromGen(
	input *[]gen.Disk,
) []Disk {
	return diskListFromGen(input)
}

// ExportLoadAveragesFromGen exposes the private loadAveragesFromGen for testing.
func ExportLoadAveragesFromGen(
	input *gen.Averages,
) *LoadAverage {
	return loadAveragesFromGen(input)
}

// ExportMemoryUsageFromGen exposes the private memoryUsageFromGen for testing.
func ExportMemoryUsageFromGen(
	input *gen.Usage,
) *Memory {
	return memoryUsageFromGen(input)
}

// ExportOSReleaseFromGen exposes the private osReleaseFromGen for testing.
func ExportOSReleaseFromGen(
	input *gen.Release,
) *OSInfo {
	return osReleaseFromGen(input)
}
