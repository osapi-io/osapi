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

package client_test

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/job/client"
	jobmocks "github.com/osapi-io/osapi/internal/job/mocks"
)

// The keys of the tests: the controller's, which signs each job record,
// and the agents', which signs each response.
var (
	testControllerPub, testControllerPriv, _ = ed25519.GenerateKey(rand.Reader)
	testAgentPub, testAgentPriv, _           = ed25519.GenerateKey(rand.Reader)
)

// testSigner signs with the controller's key, and verifies a record with
// it, as the controller's job client does.
type testSigner struct{}

func (testSigner) Sign(data []byte) []byte { return ed25519.Sign(testControllerPriv, data) }

func (testSigner) Fingerprint() string { return "SHA256:test-controller" }

func (testSigner) ControllerPublicKey() ed25519.PublicKey { return testControllerPub }

func (testSigner) StoredRecordKeys() []ed25519.PublicKey {
	return []ed25519.PublicKey{testControllerPub}
}

// testAgentKeys is the admission store of the tests: each machine ID
// m-<hostname> is an agent admitted as <hostname>, with testAgentPub.
type testAgentKeys struct{}

func (testAgentKeys) RenameAgentKey(
	context.Context,
	string,
	string,
) error {
	return nil
}

func (testAgentKeys) LookupAgentKey(
	_ context.Context,
	machineID string,
) (*client.AgentKey, error) {
	return &client.AgentKey{
		MachineID: machineID,
		Hostname:  strings.TrimPrefix(machineID, "m-"),
		PublicKey: testAgentPub,
	}, nil
}

// trust gives c the keys of the controller, as the controller sets them at
// startup: the admission store, and a signer when opts names none.
func trust(
	c *client.Client,
	opts *client.Options,
) {
	if opts == nil || opts.PKISigner == nil {
		c.SetPKISigner(testSigner{})
	}
	c.SetAgentKeyStore(testAgentKeys{})
}

// envelope is payload signed with priv, from machineID.
func envelope(
	priv ed25519.PrivateKey,
	machineID string,
	payload []byte,
) []byte {
	data, _ := json.Marshal(job.SignedEnvelope{
		Payload:     payload,
		Signature:   ed25519.Sign(priv, payload),
		Fingerprint: "SHA256:test",
		MachineID:   machineID,
	})

	return data
}

// agentSigned is the response resp as the agent that it names writes it:
// signed, from m-<hostname>.
func agentSigned(
	resp []byte,
) []byte {
	var r struct {
		Hostname string `json:"hostname"`
	}
	_ = json.Unmarshal(resp, &r)

	return envelope(testAgentPriv, "m-"+r.Hostname, resp)
}

// controllerSigned is the job record as the controller writes it.
func controllerSigned(
	record []byte,
) []byte {
	return envelope(testControllerPriv, "", record)
}

// signingNATS is the NATS of the tests, with each value signed as the
// controller and the agents write it: a job record (jobs.<id>) by the
// controller, a response (responses.*) by its agent. The tests give the
// mock plain JSON.
type signingNATS struct {
	*jobmocks.MockNATSClient
}

// signed is the entry of key with its value signed as its writer signs it.
type signedEntry struct {
	jetstream.KeyValueEntry
	sign func([]byte) []byte
}

// Value is the value of the entry, signed. A value that a test gives as a
// signed envelope already is the value as the test wrote it.
func (e signedEntry) Value() []byte {
	value := e.KeyValueEntry.Value()
	var env job.SignedEnvelope
	if json.Unmarshal(value, &env) == nil && len(env.Signature) > 0 {
		return value
	}

	return e.sign(value)
}

func signerOf(
	key string,
) func([]byte) []byte {
	switch {
	case strings.HasPrefix(key, "jobs."):
		return controllerSigned
	case strings.HasPrefix(key, "responses."):
		return agentSigned
	default:
		return nil
	}
}

func (n signingNATS) KVGet(
	ctx context.Context,
	bucket string,
	key string,
) (jetstream.KeyValueEntry, error) {
	entry, err := n.MockNATSClient.KVGet(ctx, bucket, key)
	if sign := signerOf(key); entry != nil && sign != nil {
		return signedEntry{KeyValueEntry: entry, sign: sign}, err
	}

	return entry, err
}

func (n signingNATS) KVWatch(
	ctx context.Context,
	bucket string,
	pattern string,
	opts ...jetstream.WatchOpt,
) (<-chan jetstream.KeyValueEntry, error) {
	in, err := n.MockNATSClient.KVWatch(ctx, bucket, pattern, opts...)
	sign := signerOf(pattern)
	if in == nil || sign == nil {
		return in, err
	}
	out := make(chan jetstream.KeyValueEntry, cap(in))
	go func() {
		defer close(out)
		for e := range in {
			if e == nil {
				out <- nil

				continue
			}
			out <- signedEntry{KeyValueEntry: e, sign: sign}
		}
	}()

	return out, err
}

// newTrusted is a job client of the controller on nc: it signs, verifies,
// and reads the values of nc as its writers signed them.
func newTrusted(
	nc *jobmocks.MockNATSClient,
	opts *client.Options,
) (*client.Client, error) {
	c, err := client.New(slog.Default(), signingNATS{nc}, opts)
	if err == nil {
		trust(c, opts)
	}

	return c, err
}
