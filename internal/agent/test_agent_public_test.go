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

package agent_test

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"encoding/pem"
	"log/slog"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/memfs"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/agent"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/exec"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/job/client"
	"github.com/osapi-io/osapi/internal/job/mocks"
	"github.com/osapi-io/osapi/internal/pki"
	"github.com/osapi-io/osapi/internal/provider/node/disk"
	nodeHost "github.com/osapi-io/osapi/internal/provider/node/host"
	"github.com/osapi-io/osapi/internal/provider/node/load"
	"github.com/osapi-io/osapi/internal/provider/node/memory"
	"github.com/osapi-io/osapi/internal/telemetry/process"
)

// newTestAgentParams holds all provider dependencies for building a test agent.
// This mirrors the old agent.New() parameter list, making test migration mechanical.
type newTestAgentParams struct {
	appFs           avfs.VFS
	appConfig       config.Config
	logger          *slog.Logger
	jobClient       client.JobClient
	streamName      string
	hostProvider    nodeHost.Provider
	diskProvider    disk.Provider
	memoryProvider  memory.Provider
	loadProvider    load.Provider
	processProvider process.Provider
	execManager     exec.Manager
	registryBucket  string
	factsBucket     string
	natsClient      agent.NATSClient
}

// newTestAgent builds a ProviderRegistry from the supplied providers and
// constructs an Agent — replacing the old 19-argument agent.New() call used
// across all test suites.
func newTestAgent(
	p newTestAgentParams,
) *agent.Agent {
	logger := p.logger
	if logger == nil {
		logger = slog.Default()
	}

	streamName := p.streamName
	if streamName == "" {
		streamName = "test-stream"
	}

	registry := agent.NewProviderRegistry()

	registry.Register("node", agent.NewNodeProcessor(logger))

	// The handler tests send network.ping.do as a sample job. The network
	// domain is a behavior now, so a test processor answers it.
	registry.Register(
		"network",
		func(context.Context, job.Request) (json.RawMessage, error) {
			return json.RawMessage(`{"packets_sent":3,"packets_received":3,"packet_loss":0}`), nil
		},
	)

	a := agent.New(
		p.appFs,
		p.appConfig,
		logger,
		p.jobClient,
		streamName,
		p.hostProvider,
		p.diskProvider,
		p.memoryProvider,
		p.loadProvider,
		p.processProvider,
		registry,
		p.registryBucket,
		p.factsBucket,
		p.execManager,
		p.natsClient,
	)
	// The agent is enrolled: it trusts testControllerKey, and verifies
	// each job that signedJob signs with it.
	agent.SetAgentPKIManager(a, trustedPKI())

	return a
}

// testControllerKey signs the jobs of the tests, as the controller does.
var testControllerPub, testControllerPriv, _ = ed25519.GenerateKey(rand.Reader)

// trustedPKI is the keys of an enrolled agent, which trusts
// testControllerKey.
func trustedPKI() *pki.Manager {
	m := pki.New(memfs.New(), keyDir, "agent", pki.WithEncryptionKey())
	_ = m.LoadOrGenerate()
	m.SetControllerPublicKey(testControllerPub)

	return m
}

// signedJob is the job data payload in the envelope that the controller
// writes, signed with testControllerKey.
func signedJob(
	payload string,
) []byte {
	data, _ := json.Marshal(job.SignedEnvelope{
		Payload:     []byte(payload),
		Signature:   ed25519.Sign(testControllerPriv, []byte(payload)),
		Fingerprint: "SHA256:test-controller",
	})

	return data
}

// keyDir is the PKI key directory of the test agents.
const keyDir = "/keys"

// enroll makes the agent of fs enrolled: keyDir has the controller's public
// key, so Start takes jobs at once and does not wait for admission. The
// enrollment gives the job client the agent's signer and machine ID.
func enroll(
	fs avfs.VFS,
	jobClient *mocks.MockJobClient,
) {
	ctrlPub, _, _ := ed25519.GenerateKey(rand.Reader)
	_ = fs.MkdirAll(keyDir, 0o700)
	_ = fs.WriteFile(keyDir+"/controller.pub", pem.EncodeToMemory(&pem.Block{
		Type:  "ED25519 PUBLIC KEY",
		Bytes: ctrlPub,
	}), 0o644)
	jobClient.EXPECT().SetPKISigner(gomock.Any()).AnyTimes()
	jobClient.EXPECT().SetMachineID(gomock.Any()).AnyTimes()
}
