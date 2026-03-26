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

package agent

import (
	"context"
	"encoding/json"
	"io/fs"
	"time"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/retr0h/osapi/internal/config"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/provider/command"
	dockerProv "github.com/retr0h/osapi/internal/provider/container/docker"
	fileProv "github.com/retr0h/osapi/internal/provider/file"
	"github.com/retr0h/osapi/internal/provider/network/dns"
	"github.com/retr0h/osapi/internal/provider/network/netinfo"
	"github.com/retr0h/osapi/internal/provider/network/ping"
	"github.com/retr0h/osapi/internal/provider/node/disk"
	nodeHost "github.com/retr0h/osapi/internal/provider/node/host"
	"github.com/retr0h/osapi/internal/provider/node/load"
	"github.com/retr0h/osapi/internal/provider/node/mem"
	cron "github.com/retr0h/osapi/internal/provider/scheduled/cron"
)

// SetEmbeddedFS overrides the embedded filesystem for testing.
func SetEmbeddedFS(f fs.FS) {
	embeddedFS = f
}

// ResetEmbeddedFS restores the default embedded filesystem.
func ResetEmbeddedFS() {
	embeddedFS = systemTemplates
}

// SetReadEmbeddedFile overrides the read function for testing.
func SetReadEmbeddedFile(fn func(string) ([]byte, error)) {
	readEmbeddedFile = fn
}

// ResetReadEmbeddedFile restores the default read function.
func ResetReadEmbeddedFile() {
	readEmbeddedFile = func(path string) ([]byte, error) {
		return systemTemplates.ReadFile(path)
	}
}

// ExportProcessScheduleOperation exposes the private processScheduleOperation method for testing.
func ExportProcessScheduleOperation(
	a *Agent,
	req job.Request,
) (json.RawMessage, error) {
	return a.processScheduleOperation(req)
}

// ExportProcessCronOperation exposes the private processCronOperation method for testing.
func ExportProcessCronOperation(
	a *Agent,
	req job.Request,
) (json.RawMessage, error) {
	return a.processCronOperation(req)
}

// ExportGetCronProvider exposes the private getCronProvider method for testing.
func ExportGetCronProvider(
	a *Agent,
) cron.Provider {
	return a.getCronProvider()
}

// ExportProcessJobOperation exposes the private processJobOperation method for testing.
func ExportProcessJobOperation(
	a *Agent,
	req job.Request,
) (json.RawMessage, error) {
	return a.processJobOperation(req)
}

// ExportProcessNodeOperation exposes the private processNodeOperation method for testing.
func ExportProcessNodeOperation(
	a *Agent,
	req job.Request,
) (json.RawMessage, error) {
	return a.processNodeOperation(req)
}

// ExportProcessNetworkOperation exposes the private processNetworkOperation method for testing.
func ExportProcessNetworkOperation(
	a *Agent,
	req job.Request,
) (json.RawMessage, error) {
	return a.processNetworkOperation(req)
}

// ExportProcessCommandOperation exposes the private processCommandOperation method for testing.
func ExportProcessCommandOperation(
	a *Agent,
	req job.Request,
) (json.RawMessage, error) {
	return a.processCommandOperation(req)
}

// ExportProcessDockerOperation exposes the private processDockerOperation method for testing.
func ExportProcessDockerOperation(
	a *Agent,
	req job.Request,
) (json.RawMessage, error) {
	return a.processDockerOperation(req)
}

// ExportProcessFileOperation exposes the private processFileOperation method for testing.
func ExportProcessFileOperation(
	a *Agent,
	req job.Request,
) (json.RawMessage, error) {
	return a.processFileOperation(req)
}

// ExportGetHostProvider exposes the private getHostProvider method for testing.
func ExportGetHostProvider(
	a *Agent,
) nodeHost.Provider {
	return a.getHostProvider()
}

// ExportGetDiskProvider exposes the private getDiskProvider method for testing.
func ExportGetDiskProvider(
	a *Agent,
) disk.Provider {
	return a.getDiskProvider()
}

// ExportGetMemProvider exposes the private getMemProvider method for testing.
func ExportGetMemProvider(
	a *Agent,
) mem.Provider {
	return a.getMemProvider()
}

// ExportGetLoadProvider exposes the private getLoadProvider method for testing.
func ExportGetLoadProvider(
	a *Agent,
) load.Provider {
	return a.getLoadProvider()
}

// ExportGetDNSProvider exposes the private getDNSProvider method for testing.
func ExportGetDNSProvider(
	a *Agent,
) dns.Provider {
	return a.getDNSProvider()
}

// ExportGetPingProvider exposes the private getPingProvider method for testing.
func ExportGetPingProvider(
	a *Agent,
) ping.Provider {
	return a.getPingProvider()
}

// ExportGetCommandProvider exposes the private getCommandProvider method for testing.
func ExportGetCommandProvider(
	a *Agent,
) command.Provider {
	return a.getCommandProvider()
}

// ExportGetFileProvider exposes the private getFileProvider method for testing.
func ExportGetFileProvider(
	a *Agent,
) fileProv.Provider {
	return a.getFileProvider()
}

// ExportWriteStatusEvent exposes the private writeStatusEvent method for testing.
func ExportWriteStatusEvent(
	ctx context.Context,
	a *Agent,
	jobID string,
	event string,
	data map[string]interface{},
) error {
	return a.writeStatusEvent(ctx, jobID, event, data)
}

// ExportHandleJobMessage exposes the private handleJobMessage method for testing.
func ExportHandleJobMessage(
	a *Agent,
	msg jetstream.Msg,
) error {
	return a.handleJobMessage(msg)
}

// ExportExtractChanged exposes the private extractChanged function for testing.
func ExportExtractChanged(
	data json.RawMessage,
) *bool {
	return extractChanged(data)
}

// ExportConsumeQueryJobs exposes the private consumeQueryJobs method for testing.
func ExportConsumeQueryJobs(
	ctx context.Context,
	a *Agent,
	hostname string,
) error {
	return a.consumeQueryJobs(ctx, hostname)
}

// ExportConsumeModifyJobs exposes the private consumeModifyJobs method for testing.
func ExportConsumeModifyJobs(
	ctx context.Context,
	a *Agent,
	hostname string,
) error {
	return a.consumeModifyJobs(ctx, hostname)
}

// ExportCreateConsumer exposes the private createConsumer method for testing.
func ExportCreateConsumer(
	ctx context.Context,
	a *Agent,
	streamName string,
	consumerName string,
	filterSubject string,
) error {
	return a.createConsumer(ctx, streamName, consumerName, filterSubject)
}

// ExportHandleJobMessageJS exposes the private handleJobMessageJS method for testing.
func ExportHandleJobMessageJS(
	a *Agent,
	msg jetstream.Msg,
) error {
	return a.handleJobMessageJS(msg)
}

// ExportCheckDrainFlag exposes the private checkDrainFlag method for testing.
func ExportCheckDrainFlag(
	ctx context.Context,
	a *Agent,
	hostname string,
) bool {
	return a.checkDrainFlag(ctx, hostname)
}

// ExportHandleDrainDetection exposes the private handleDrainDetection method for testing.
func ExportHandleDrainDetection(
	ctx context.Context,
	a *Agent,
	hostname string,
) {
	a.handleDrainDetection(ctx, hostname)
}

// ExportWriteFacts exposes the private writeFacts method for testing.
func ExportWriteFacts(
	ctx context.Context,
	a *Agent,
	hostname string,
) {
	a.writeFacts(ctx, hostname)
}

// ExportStartFacts exposes the private startFacts method for testing.
func ExportStartFacts(
	ctx context.Context,
	a *Agent,
	hostname string,
) {
	a.startFacts(ctx, hostname)
}

// ExportFactsKey exposes the private factsKey function for testing.
func ExportFactsKey(
	hostname string,
) string {
	return factsKey(hostname)
}

// ExportWriteRegistration exposes the private writeRegistration method for testing.
func ExportWriteRegistration(
	ctx context.Context,
	a *Agent,
	hostname string,
) {
	a.writeRegistration(ctx, hostname)
}

// ExportDeregister exposes the private deregister method for testing.
func ExportDeregister(
	a *Agent,
	hostname string,
) {
	a.deregister(hostname)
}

// ExportStartHeartbeat exposes the private startHeartbeat method for testing.
func ExportStartHeartbeat(
	ctx context.Context,
	a *Agent,
	hostname string,
) {
	a.startHeartbeat(ctx, hostname)
}

// ExportRegistryKey exposes the private registryKey function for testing.
func ExportRegistryKey(
	hostname string,
) string {
	return registryKey(hostname)
}

// ExportFindPrevCondition exposes the private findPrevCondition function for testing.
func ExportFindPrevCondition(
	condType string,
	prev []job.Condition,
) *job.Condition {
	return findPrevCondition(condType, prev)
}

// ExportTransitionTime exposes the private transitionTime function for testing.
func ExportTransitionTime(
	condType string,
	newStatus bool,
	prev []job.Condition,
) time.Time {
	return transitionTime(condType, newStatus, prev)
}

// ExportEvaluateMemoryPressure exposes the private evaluateMemoryPressure function for testing.
func ExportEvaluateMemoryPressure(
	stats *mem.Result,
	threshold int,
	prev []job.Condition,
) job.Condition {
	return evaluateMemoryPressure(stats, threshold, prev)
}

// ExportEvaluateHighLoad exposes the private evaluateHighLoad function for testing.
func ExportEvaluateHighLoad(
	loadAvg *load.Result,
	cpuCount int,
	multiplier float64,
	prev []job.Condition,
) job.Condition {
	return evaluateHighLoad(loadAvg, cpuCount, multiplier, prev)
}

// ExportEvaluateDiskPressure exposes the private evaluateDiskPressure function for testing.
func ExportEvaluateDiskPressure(
	disks []disk.Result,
	threshold int,
	prev []job.Condition,
) job.Condition {
	return evaluateDiskPressure(disks, threshold, prev)
}

// --- Package-level variable accessors for testing ---

// SetMarshalJSON overrides the marshalJSON function for testing.
func SetMarshalJSON(fn func(interface{}) ([]byte, error)) {
	marshalJSON = fn
}

// ResetMarshalJSON restores the default marshalJSON function.
func ResetMarshalJSON() {
	marshalJSON = json.Marshal
}

// SetUnmarshalJSON overrides the unmarshalJSON function for testing.
func SetUnmarshalJSON(fn func([]byte, interface{}) error) {
	unmarshalJSON = fn
}

// ResetUnmarshalJSON restores the default unmarshalJSON function.
func ResetUnmarshalJSON() {
	unmarshalJSON = json.Unmarshal
}

// SetFactsInterval overrides the factsInterval for testing.
func SetFactsInterval(d time.Duration) {
	factsInterval = d
}

// ResetFactsInterval restores the default factsInterval.
func ResetFactsInterval() {
	factsInterval = 60 * time.Second
}

// SetHeartbeatInterval overrides the heartbeatInterval for testing.
func SetHeartbeatInterval(d time.Duration) {
	heartbeatInterval = d
}

// ResetHeartbeatInterval restores the default heartbeatInterval.
func ResetHeartbeatInterval() {
	heartbeatInterval = 10 * time.Second
}

// SetFactoryDockerNewFn overrides the factoryDockerNewFn for testing.
func SetFactoryDockerNewFn(fn func() (*dockerProv.Client, error)) {
	factoryDockerNewFn = fn
}

// ResetFactoryDockerNewFn restores the default factoryDockerNewFn.
func ResetFactoryDockerNewFn() {
	factoryDockerNewFn = dockerProv.New
}

// --- Field accessors for Agent struct ---

// GetAgentAppConfig returns the agent's appConfig field for testing.
func GetAgentAppConfig(
	a *Agent,
) config.Config {
	return a.appConfig
}

// SetAgentAppConfig sets the agent's appConfig field for testing.
func SetAgentAppConfig(
	a *Agent,
	cfg config.Config,
) {
	a.appConfig = cfg
}

// GetAgentHostProvider returns the agent's hostProvider field for testing.
func GetAgentHostProvider(
	a *Agent,
) nodeHost.Provider {
	return a.hostProvider
}

// SetAgentHostProvider sets the agent's hostProvider field for testing.
func SetAgentHostProvider(
	a *Agent,
	p nodeHost.Provider,
) {
	a.hostProvider = p
}

// GetAgentNetinfoProvider returns the agent's netinfoProvider field for testing.
func GetAgentNetinfoProvider(
	a *Agent,
) netinfo.Provider {
	return a.netinfoProvider
}

// SetAgentNetinfoProvider sets the agent's netinfoProvider field for testing.
func SetAgentNetinfoProvider(
	a *Agent,
	p netinfo.Provider,
) {
	a.netinfoProvider = p
}

// GetAgentCachedFacts returns the agent's cachedFacts field for testing.
func GetAgentCachedFacts(
	a *Agent,
) *job.FactsRegistration {
	return a.cachedFacts
}

// SetAgentCachedFacts sets the agent's cachedFacts field for testing.
func SetAgentCachedFacts(
	a *Agent,
	facts *job.FactsRegistration,
) {
	a.cachedFacts = facts
}

// GetAgentState returns the agent's state field for testing.
func GetAgentState(
	a *Agent,
) string {
	return a.state
}

// SetAgentState sets the agent's state field for testing.
func SetAgentState(
	a *Agent,
	state string,
) {
	a.state = state
}

// SetAgentLifecycle sets the agent's ctx/cancel and consumerCtx/consumerCancel for testing.
func SetAgentLifecycle(
	ctx context.Context,
	consumerCtx context.Context,
	a *Agent,
	cancel context.CancelFunc,
	consumerCancel context.CancelFunc,
) {
	a.ctx = ctx
	a.cancel = cancel
	a.consumerCtx = consumerCtx
	a.consumerCancel = consumerCancel
}

// GetAgentWG returns the agent's wg field for testing (as a pointer to allow Wait).
func WaitAgentWG(
	a *Agent,
) {
	a.wg.Wait()
}
