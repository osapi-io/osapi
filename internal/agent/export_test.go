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
	"log/slog"
	"os"
	osexec "os/exec"
	"sync"
	"time"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/osapi-io/gohai/pkg/gohai/collectors/cpu"
	"github.com/osapi-io/gohai/pkg/gohai/collectors/hostname"
	"github.com/osapi-io/gohai/pkg/gohai/collectors/initd"
	"github.com/osapi-io/gohai/pkg/gohai/collectors/kernel"
	"github.com/osapi-io/gohai/pkg/gohai/collectors/network"
	"github.com/osapi-io/gohai/pkg/gohai/collectors/packagemgr"
	"github.com/osapi-io/gohai/pkg/gohai/collectors/virtualization"

	"github.com/osapi-io/gohai/pkg/gohai"
	"github.com/osapi-io/gohai/pkg/gohai/source"

	"github.com/avfs/avfs"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/avfs/avfs/vfs/memfs"
	"github.com/osapi-io/osapi/internal/agent/identity"
	behaviorHost "github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/exec"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/pki"
	"github.com/osapi-io/osapi/internal/provider"
	diskProv "github.com/osapi-io/osapi/internal/provider/node/disk"
	nodeHost "github.com/osapi-io/osapi/internal/provider/node/host"
	"github.com/osapi-io/osapi/internal/provider/node/load"
	memProv "github.com/osapi-io/osapi/internal/provider/node/memory"
)

// SetEmbeddedFS overrides the embedded filesystem for testing.
func SetEmbeddedFS(
	f fs.FS,
) {
	embeddedFS = f
}

// ResetEmbeddedFS restores the default embedded filesystem.
func ResetEmbeddedFS() {
	embeddedFS = systemTemplates
}

// SetReadEmbeddedFile overrides the read function for testing.
func SetReadEmbeddedFile(
	fn func(string) ([]byte, error),
) {
	readEmbeddedFile = fn
}

// ResetReadEmbeddedFile restores the default read function.
func ResetReadEmbeddedFile() {
	readEmbeddedFile = defaultReadEmbeddedFile
}

// ExportDefaultReadEmbeddedFile exposes defaultReadEmbeddedFile for testing.
func ExportDefaultReadEmbeddedFile(
	path string,
) ([]byte, error) {
	return defaultReadEmbeddedFile(path)
}

// ExportProcessJobOperation exposes the private processJobOperation method for testing.
func ExportProcessJobOperation(
	a *Agent,
	req job.Request,
) (json.RawMessage, error) {
	return a.processJobOperation(context.Background(), req)
}

// ExportProcessNodeOperation invokes the "node" processor from the agent's
// registry directly, for testing node-specific dispatch.
func ExportProcessNodeOperation(
	a *Agent,
	req job.Request,
) (json.RawMessage, error) {
	req.Category = "node"
	return a.registry.Dispatch(context.Background(), req)
}

// ExportProcessNetworkOperation invokes the "network" processor from the
// agent's registry directly, for testing network-specific dispatch.
func ExportProcessNetworkOperation(
	a *Agent,
	req job.Request,
) (json.RawMessage, error) {
	req.Category = "network"
	return a.registry.Dispatch(context.Background(), req)
}

// ExportWriteStatusEvent exposes the private writeStatusEvent method for testing.
func ExportWriteStatusEvent(
	ctx context.Context,
	a *Agent,
	jobID string,
	event types.AgentJobStatus,
	data map[string]interface{},
) error {
	return a.writeStatusEvent(ctx, jobID, event, data)
}

// ExportHandleJobMessage exposes the private handleJobMessage method for testing.
func ExportHandleJobMessage(
	ctx context.Context,
	a *Agent,
	msg jetstream.Msg,
) error {
	return a.handleJobMessage(ctx, msg)
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

// ExportConsumerNamePrefix exposes the private consumerNamePrefix method for testing.
func ExportConsumerNamePrefix(
	a *Agent,
) string {
	return a.consumerNamePrefix()
}

// ExportHandleJobMessageJS exposes the private handleJobMessageJS method for testing.
func ExportHandleJobMessageJS(
	a *Agent,
	msg jetstream.Msg,
) error {
	return a.handleJobMessageJS(context.Background(), msg)
}

// SetDrainStuckAfter sets how long a drain waits before it reports the
// DrainStuck condition, and returns a func that restores it.
func SetDrainStuckAfter(
	d time.Duration,
) func() {
	prev := drainStuckAfter
	drainStuckAfter = d

	return func() { drainStuckAfter = prev }
}

// ExportDrainCondition exposes drainCondition for testing.
func ExportDrainCondition(
	a *Agent,
) job.Condition {
	return a.drainCondition()
}

// SetConsumerRetryInterval sets how long the agent waits before it tries
// again to create or bind a failed job consumer, and returns a func that
// restores it.
func SetConsumerRetryInterval(
	d time.Duration,
) func() {
	prev := consumerRetryInterval
	consumerRetryInterval = d

	return func() { consumerRetryInterval = prev }
}

// ExportConsumersCondition exposes consumersCondition for testing.
func ExportConsumersCondition(
	a *Agent,
) job.Condition {
	return a.consumersCondition()
}

// ExportWaitConsumers waits for each consumer goroutine of the agent.
func ExportWaitConsumers(
	a *Agent,
) {
	a.consumerWg.Wait()
}

// ExportWriteSeen exposes writeSeen for testing.
func ExportWriteSeen(
	ctx context.Context,
	a *Agent,
	machineID string,
) {
	a.writeSeen(ctx, machineID)
}

// ExportStartConsumers exposes startConsumers, the start of the job
// consumers at startup, for testing.
func ExportStartConsumers(
	a *Agent,
) {
	a.startConsumers()
}

// ExportCheckDrainFlag exposes the private checkDrainFlag method for testing.
func ExportCheckDrainFlag(
	ctx context.Context,
	a *Agent,
	machineID string,
) bool {
	return a.checkDrainFlag(ctx, machineID)
}

// ExportHandleDrainDetection exposes the private handleDrainDetection method for testing.
func ExportHandleDrainDetection(
	ctx context.Context,
	a *Agent,
	machineID string,
	hostname string,
) {
	a.handleDrainDetection(ctx, machineID, hostname)
}

// ExportWriteFacts exposes the private writeFacts method for testing.
func ExportWriteFacts(
	ctx context.Context,
	a *Agent,
	machineID string,
	hostname string,
) {
	a.writeFacts(ctx, machineID, hostname)
}

// ExportStartFacts exposes the private startFacts method for testing.
func ExportStartFacts(
	ctx context.Context,
	a *Agent,
	machineID string,
	hostname string,
) {
	a.startFacts(ctx, machineID, hostname)
}

// ExportFactsKey exposes the private factsKey function for testing.
func ExportFactsKey(
	machineID string,
) string {
	return factsKey(machineID)
}

// ExportWriteRegistration exposes the private writeRegistration method for testing.
func ExportWriteRegistration(
	ctx context.Context,
	a *Agent,
	machineID string,
	hostname string,
) {
	a.writeRegistration(ctx, machineID, hostname)
}

// ExportDeregister exposes the private deregister method for testing.
func ExportDeregister(
	a *Agent,
	machineID string,
) {
	a.deregister(machineID)
}

// ExportStartHeartbeat exposes the private startHeartbeat method for testing.
func ExportStartHeartbeat(
	ctx context.Context,
	a *Agent,
	machineID string,
	hostname string,
) {
	a.startHeartbeat(ctx, machineID, hostname)
}

// ExportRegistryKey exposes the private registryKey function for testing.
func ExportRegistryKey(
	machineID string,
) string {
	return registryKey(machineID)
}

// SetAgentMachineID sets the agent's machineID field for testing.
func SetAgentMachineID(
	a *Agent,
	machineID string,
) {
	a.machineID = machineID
}

// GetAgentMachineID returns the agent's machineID field for testing.
func GetAgentMachineID(
	a *Agent,
) string {
	return a.machineID
}

// ExportFindPrevCondition exposes the private findPrevCondition function for testing.
func ExportFindPrevCondition(
	condType types.ConditionType,
	prev []job.Condition,
) *job.Condition {
	return findPrevCondition(condType, prev)
}

// ExportTransitionTime exposes the private transitionTime function for testing.
func ExportTransitionTime(
	condType types.ConditionType,
	newStatus bool,
	prev []job.Condition,
) time.Time {
	return transitionTime(condType, newStatus, prev)
}

// ExportEvaluateMemoryPressure exposes the private evaluateMemoryPressure function for testing.
func ExportEvaluateMemoryPressure(
	stats *memProv.Result,
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
	disks []diskProv.Result,
	threshold int,
	prev []job.Condition,
) job.Condition {
	return evaluateDiskPressure(disks, threshold, prev)
}

// --- Package-level variable accessors for testing ---

// SetMarshalJSON overrides the marshalJSON function for testing.
func SetMarshalJSON(
	fn func(interface{}) ([]byte, error),
) {
	marshalJSON = fn
}

// ResetMarshalJSON restores the default marshalJSON function.
func ResetMarshalJSON() {
	marshalJSON = json.Marshal
}

// SetUnmarshalJSON overrides the unmarshalJSON function for testing.
func SetUnmarshalJSON(
	fn func([]byte, interface{}) error,
) {
	unmarshalJSON = fn
}

// ResetUnmarshalJSON restores the default unmarshalJSON function.
func ResetUnmarshalJSON() {
	unmarshalJSON = json.Unmarshal
}

// SetDefaultFactsInterval overrides the defaultFactsInterval for testing.
func SetDefaultFactsInterval(
	d time.Duration,
) {
	defaultFactsInterval = d
}

// ResetDefaultFactsInterval restores the default defaultFactsInterval.
func ResetDefaultFactsInterval() {
	defaultFactsInterval = 60 * time.Second
}

// SetInProgressInterval overrides the InProgress keepalive interval for
// testing. Zero restores deriving it from the consumer's AckWait.
func SetInProgressInterval(
	d time.Duration,
) {
	inProgressInterval = d
}

// ResetInProgressInterval restores deriving the InProgress keepalive
// interval from the consumer's AckWait.
func ResetInProgressInterval() {
	inProgressInterval = 0
}

// ExportStartInProgressKeepAlive exposes the private startInProgressKeepAlive
// method for testing.
func ExportStartInProgressKeepAlive(
	ctx context.Context,
	a *Agent,
	msg jetstream.Msg,
) func() {
	return a.startInProgressKeepAlive(ctx, msg)
}

// SetHeartbeatInterval overrides the heartbeatInterval for testing.
func SetHeartbeatInterval(
	d time.Duration,
) {
	heartbeatInterval = d
}

// ResetHeartbeatInterval restores the default heartbeatInterval.
func ResetHeartbeatInterval() {
	heartbeatInterval = 10 * time.Second
}

// SetGetAgentHostnameFn overrides the getAgentHostnameFn for testing.
func SetGetAgentHostnameFn(
	fn func(string) (string, error),
) {
	getAgentHostnameFn = fn
}

// ResetGetAgentHostnameFn restores the default getAgentHostnameFn.
func ResetGetAgentHostnameFn() {
	getAgentHostnameFn = job.GetAgentHostname
}

// GetAgentHostname returns the agent's hostname field for testing.
func GetAgentHostname(
	a *Agent,
) string {
	return a.hostname
}

// SetAgentHostname sets the agent's hostname field for testing.
func SetAgentHostname(
	a *Agent,
	hostname string,
) {
	a.hostname = hostname
}

// SetProcStatusPath overrides the procStatusPath for testing.
func SetProcStatusPath(
	p string,
) {
	procStatusPath = p
}

// ResetProcStatusPath restores the default procStatusPath.
func ResetProcStatusPath() {
	procStatusPath = "/proc/self/status"
}

func ExportCheckCapabilities(
	logger *slog.Logger,
) []PreflightResult {
	return checkCapabilities(logger)
}

// --- Field accessors for Agent struct ---

// GetAgentAppConfig returns the agent's appConfig field for testing.
func GetAgentAppConfig(
	a *Agent,
) config.Config {
	return a.appConfig
}

// SetAgentAppConfig sets the agent's appConfig field for testing.
// Also re-registers the node processor so that label changes are picked up
// by subsequent ExportProcessNodeOperation calls.
func SetAgentAppConfig(
	a *Agent,
	cfg config.Config,
) {
	a.appConfig = cfg
	// Re-register the node processor with the updated config so that
	// label/config-dependent dispatch picks up the change.
	if a.registry != nil {
		a.registry.processors["node"] = NewNodeProcessor(a.logger)
	}
}

// GetAgentHostProvider returns the agent's hostProvider field for testing.
func GetAgentHostProvider(
	a *Agent,
) nodeHost.Provider {
	return a.hostProvider
}

// SetAgentHostProvider sets the agent's hostProvider field for testing.
// This also updates the corresponding processor in the registry so that
// job processing uses the new provider.
func SetAgentHostProvider(
	a *Agent,
	p nodeHost.Provider,
) {
	a.hostProvider = p
	// Re-register the node processor with the new host provider so job
	// dispatch picks up the change.
	if a.registry != nil {
		a.registry.processors["node"] = NewNodeProcessor(a.logger)
	}
}

// SetAgentProcessorOnce makes the agent's next job of category return err,
// and puts back the category's processor after it.
func SetAgentProcessorOnce(
	a *Agent,
	category string,
	err error,
) {
	orig := a.registry.processors[category]
	a.registry.processors[category] = func(context.Context, job.Request) (json.RawMessage, error) {
		a.registry.processors[category] = orig

		return nil, err
	}
}

// ExportGetHostProvider returns the host provider from the registry for testing.
// Satisfies the TestProviderFactoryMethods suite.
func ExportGetHostProvider(
	a *Agent,
) nodeHost.Provider {
	return a.hostProvider
}

// ExportGetDiskProvider returns the disk provider from the registry for testing.
func ExportGetDiskProvider(
	a *Agent,
) diskProv.Provider {
	return a.diskProvider
}

// ExportGetMemProvider returns the mem provider from the registry for testing.
func ExportGetMemProvider(
	a *Agent,
) memProv.Provider {
	return a.memoryProvider
}

// ExportGetLoadProvider returns the load provider from the registry for testing.
func ExportGetLoadProvider(
	a *Agent,
) load.Provider {
	return a.loadProvider
}

// GetAgentCachedFacts returns the agent's cachedFacts field for testing.
func GetAgentCachedFacts(
	a *Agent,
) *job.FactsRegistration {
	return a.cachedFacts.Load()
}

// SetAgentCachedFacts sets the agent's cachedFacts field for testing.
func SetAgentCachedFacts(
	a *Agent,
	facts *job.FactsRegistration,
) {
	a.cachedFacts.Store(facts)
}

// ExportAgentFacts returns the facts that a behavior's host reads, for the
// agent a. A nil a is an agent that is not made yet.
func ExportAgentFacts(
	a *Agent,
) func() map[string]any {
	return (&agentFacts{agent: a}).get
}

// GetAgentState returns the agent's state field for testing.
func GetAgentState(
	a *Agent,
) types.AgentState {
	return a.State()
}

// SetAgentState sets the agent's state field for testing.
func SetAgentState(
	a *Agent,
	state types.AgentState,
) {
	a.setState(state)
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

// SetGetIdentityFn overrides the getIdentityFn variable for testing.
func SetGetIdentityFn(
	fn func(avfs.VFS, string) (*identity.Identity, error),
) {
	getIdentityFn = fn
}

// ResetGetIdentityFn restores the default getIdentityFn variable.
func ResetGetIdentityFn() {
	getIdentityFn = identity.GetIdentity
}

// ExportHandlePKIEnrollment exposes the private handlePKIEnrollment method for testing.
func ExportHandlePKIEnrollment(
	ctx context.Context,
	a *Agent,
) error {
	return a.handlePKIEnrollment(ctx)
}

// GetAgentPKIManager returns the agent's pkiManager field for testing.
func GetAgentPKIManager(
	a *Agent,
) *pki.Manager {
	return a.pkiManager
}

// SetAgentPKIManager sets the agent's pkiManager field for testing.
func SetAgentPKIManager(
	a *Agent,
	m *pki.Manager,
) {
	a.pkiManager = m
}

// ExportPublishEnrollmentRequest exposes the private publishEnrollmentRequest method for testing.
func ExportPublishEnrollmentRequest(
	a *Agent,
) error {
	return a.publishEnrollmentRequest()
}

// ExportRetryEnrollmentIfPending exposes the republish a pending agent performs
// on each heartbeat.
func ExportRetryEnrollmentIfPending(
	a *Agent,
) {
	a.retryEnrollmentIfPending()
}

// SetMarshalJSONEnrollment overrides the marshalJSONEnrollment function for testing.
func SetMarshalJSONEnrollment(
	fn func(interface{}) ([]byte, error),
) {
	marshalJSONEnrollment = fn
}

// ResetMarshalJSONEnrollment restores the default marshalJSONEnrollment function.
func ResetMarshalJSONEnrollment() {
	marshalJSONEnrollment = json.Marshal
}

// SetAgentNATSClient sets the agent's natsClient field for testing.
func SetAgentNATSClient(
	a *Agent,
	c NATSClient,
) {
	a.natsClient = c
}

// ExportUnwrapJobEnvelope exposes the private unwrapJobEnvelope method for testing.
func ExportUnwrapJobEnvelope(
	a *Agent,
	data []byte,
) ([]byte, error) {
	return a.unwrapJobEnvelope(data)
}

// ExportStartEnrollmentListener exposes the private startEnrollmentListener method for testing.
func ExportStartEnrollmentListener(
	ctx context.Context,
	a *Agent,
) {
	a.startEnrollmentListener(ctx)
}

// ExportHandleEnrollmentResponse exposes the private handleEnrollmentResponse method for testing.
func ExportHandleEnrollmentResponse(
	a *Agent,
	msg *nats.Msg,
) {
	a.handleEnrollmentResponse(msg)
}

// ExportClassifyProviderError exposes the classification that turns a provider's
// error into the code a response carries.
func ExportClassifyProviderError(
	err error,
) job.ErrorCode {
	return classifyProviderError(err)
}

// SetLookPath overrides how MissingPrograms finds a program.
func SetLookPath(
	fn func(string) (string, error),
) {
	lookPath = fn
}

// ResetLookPath restores the PATH lookup.
func ResetLookPath() {
	lookPath = osexec.LookPath
}

// ExportProviders makes each provider of a host with the given platform, and
// returns them by area.
func ExportProviders(
	log *slog.Logger,
	plat string,
	container bool,
	fileState provider.FileState,
) map[string]any {
	e := hostEnv{
		log:             log,
		fs:              memfs.New(),
		exec:            exec.New(log, false),
		platform:        plat,
		container:       container,
		hostname:        "test-host",
		fileState:       fileState,
		fileStateBucket: "file-state",
	}

	return map[string]any{
		"host":   e.host(),
		"disk":   e.disk(),
		"memory": e.memory(),
		"load":   e.load(),
	}
}

// SetPlatformFns replaces the functions that tell the platform of the host.
func SetPlatformFns(
	detect func() string,
	container func() bool,
) {
	detectFn = detect
	containerFn = container
}

// ExportInContainer tells whether the host of src is a container.
func ExportInContainer(
	src source.Sources,
) bool {
	return inContainer(src)
}

// ExportGohaiFacts collects the facts that opts select.
func ExportGohaiFacts(
	opts ...gohai.Option,
) (*gohai.Facts, error) {
	return gohaiFacts(opts...)
}

// SetCollectFactsFn replaces how the agent collects facts.
func SetCollectFactsFn(
	fn func() (*gohai.Facts, error),
) {
	collectFactsFn = fn
}

// ResetCollectFactsFn puts back the fake facts that agent tests collect.
func ResetCollectFactsFn() {
	collectFactsFn = FakeFacts
}

// ExportCollectFacts collects the facts of this host with gohai.
func ExportCollectFacts() (*gohai.Facts, error) {
	return collectFacts()
}

// FakeFacts are the facts an agent test collects: a Linux host with eth0
// up, eth1 down and lo, a default route, and two IPv6 routes. No agent
// test reads the host it runs on.
func FakeFacts() (*gohai.Facts, error) {
	return &gohai.Facts{
		Kernel:     &kernel.Info{Machine: "x86_64", Release: "5.15.0-91-generic"},
		Hostname:   &hostname.Info{FQDN: "default-hostname.local"},
		CPU:        &cpu.Info{Count: 4},
		Init:       &initd.Info{Name: "systemd"},
		PackageMgr: &packagemgr.Info{Name: "apt"},
		Virtualization: &virtualization.Info{
			Systems: map[string]string{"docker": "guest"},
		},
		Network: &network.Info{
			DefaultInterface: "eth0",
			Interfaces: []network.Interface{
				{Name: "lo", State: "up", Flags: []string{"up", "loopback"}},
				{
					Name: "eth0", State: "up", MAC: "02:42:ac:11:00:02", Flags: []string{"up"},
					Addresses: []network.Address{
						{Addr: "10.0.0.5", Family: "inet"},
						{Addr: "10.0.0.6", Family: "inet"},
						{Addr: "fe80::1", Family: "inet6"},
					},
				},
				{Name: "eth1", State: "down"},
				{
					Name: "wg0", State: "up",
					Addresses: []network.Address{{Addr: "fd00::2", Family: "inet6"}},
				},
				{
					Name: "eth2", State: "up",
					Addresses: []network.Address{{Addr: "192.168.1.2", Family: "inet"}},
				},
			},
			Routes: []network.Route{
				{
					Destination: "default",
					Family:      "inet",
					Gateway:     "10.0.0.1",
					Interface:   "eth0",
					Metric:      100,
				},
				{Destination: "10.0.0.0/24", Family: "inet", Interface: "eth0"},
				{Destination: "127", Family: "inet", Gateway: "127.0.0.1", Interface: "lo"},
				{Destination: "default", Family: "inet6", Gateway: "fe80::1", Interface: "eth0"},
			},
		},
	}, nil
}

func init() {
	collectFactsFn = FakeFacts
}

// defaultDetectFn and defaultContainerFn are the functions that setup.go
// gives detectFn and containerFn, kept before a test replaces them.
var (
	defaultDetectFn    = detectFn
	defaultContainerFn = containerFn
)

// ResetPlatformFns puts back the functions that tell the platform of the host.
func ResetPlatformFns() {
	detectFn = defaultDetectFn
	containerFn = defaultContainerFn
}

// ExportDefaultContainerFn calls the containerFn that setup.go gives.
func ExportDefaultContainerFn() bool {
	return defaultContainerFn()
}

// ExportIdentityAttrs gives a the hostname, machine ID and facts that it
// would resolve at startup, and returns the span attributes for its jobs.
func ExportIdentityAttrs(
	a *Agent,
	hostname string,
	machineID string,
	facts *job.FactsRegistration,
) []attribute.KeyValue {
	a.hostname = hostname
	a.machineID = machineID
	a.cachedFacts.Store(facts)

	return a.identityAttrs()
}

// ExportRecordOutcome exposes recordOutcome for testing.
func ExportRecordOutcome(
	span trace.Span,
	response job.Response,
	requestType types.JobType,
) {
	recordOutcome(span, response, requestType)
}

// SetPingRangePath overrides the ping_group_range path for testing.
func SetPingRangePath(
	p string,
) {
	pingRangePath = p
}

// ResetPingRangePath restores the default ping_group_range path.
func ResetPingRangePath() {
	pingRangePath = "/proc/sys/net/ipv4/ping_group_range"
}

// SetProcessGroups overrides the groups of the agent for testing.
func SetProcessGroups(
	fn func() ([]int, error),
) {
	processGroups = fn
}

// ResetProcessGroups restores the groups of the agent.
func ResetProcessGroups() {
	processGroups = defaultProcessGroups
}

// ExportDefaultProcessGroups exposes defaultProcessGroups for testing.
var ExportDefaultProcessGroups = defaultProcessGroups

// SetGetgroups overrides os.Getgroups for testing.
func SetGetgroups(
	fn func() ([]int, error),
) {
	getgroups = fn
}

// ResetGetgroups restores os.Getgroups.
func ResetGetgroups() {
	getgroups = os.Getgroups
}

// ExportStoreObjects is the object store that behaviors read, over store.
func ExportStoreObjects(
	store ObjectReader,
	bucket string,
) behaviorHost.Objects {
	return storeObjects{store: store, bucket: bucket}
}

// SignalRenamed tells the heartbeat of a that a hostname update completed.
func SignalRenamed(
	a *Agent,
) {
	a.renamed <- struct{}{}
}

// ExportNoteRename exposes the private noteRename for testing, and reports
// whether the heartbeat has a rename to read.
func ExportNoteRename(
	a *Agent,
	req job.Request,
) bool {
	a.noteRename(req)

	return len(a.renamed) == 1
}

// HoldConsumers gives a one running consumer, which a drain waits for
// until release is called.
func HoldConsumers(
	a *Agent,
) (release func()) {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	a.lifecycleMu.Lock()
	a.consumerCancel = func() {}
	a.consumerSetWg = wg
	a.lifecycleMu.Unlock()

	var once sync.Once

	return func() { once.Do(wg.Done) }
}
