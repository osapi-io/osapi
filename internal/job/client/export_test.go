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
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/secret/seal"
)

// ExportSanitizeKeyForNATS exposes the private sanitizeKeyForNATS for testing.
func ExportSanitizeKeyForNATS(
	s string,
) string {
	return sanitizeKeyForNATS(s)
}

// ExportWrapInSignedEnvelope exposes the private wrapInSignedEnvelope for testing.
func ExportWrapInSignedEnvelope(
	signer PKISigner,
	machineID string,
	payload []byte,
) ([]byte, error) {
	return wrapInSignedEnvelope(signer, machineID, payload)
}

// ExportVerifyAgentResponse exposes the private verifyAgentResponse for
// testing.
func ExportVerifyAgentResponse(
	ctx context.Context,
	store AgentKeyStore,
	data []byte,
) (*job.Response, error) {
	return verifyAgentResponse(ctx, store, data)
}

// ExportUnwrapSignedEnvelope exposes the private unwrapSignedEnvelope for testing.
func ExportUnwrapSignedEnvelope(
	data []byte,
	pubKey []byte,
) ([]byte, error) {
	return unwrapSignedEnvelope(data, pubKey)
}

// SetSigningMarshalFn overrides the signingMarshalFn variable for testing.
func SetSigningMarshalFn(
	fn func(any) ([]byte, error),
) {
	signingMarshalFn = fn
}

// ResetSigningMarshalFn restores the signingMarshalFn variable to its default.
func ResetSigningMarshalFn() {
	signingMarshalFn = json.Marshal
}

// ExportComputeStatusFromKeyNames exposes the private computeStatusFromKeyNames
// for testing, returning the ordered IDs and a simplified map[jobID]status.
func ExportComputeStatusFromKeyNames(
	keys []string,
) ([]string, map[string]string) {
	orderedIDs, jobStatuses := computeStatusFromKeyNames(keys)
	statuses := make(map[string]string, len(jobStatuses))
	for id, info := range jobStatuses {
		statuses[id] = info.Status
	}
	return orderedIDs, statuses
}

// SetNowFn overrides nowFn for testing a rotation grace window.
func SetNowFn(
	fn func() time.Time,
) {
	nowFn = fn
}

// ResetNowFn restores nowFn to its default.
func ResetNowFn() {
	nowFn = time.Now
}

// SetNewFn replaces the function that Connect makes the job client with.
func SetNewFn(
	fn func(*slog.Logger, NATSClient, *Options) (*Client, error),
) {
	newFn = fn
}

// ResetNewFn puts back the function that Connect makes the job client with.
func ResetNewFn() {
	newFn = New
}

// ExportEndCollectSpan exposes endCollectSpan for testing.
var ExportEndCollectSpan = endCollectSpan

// ExportOperation exposes operation for testing.
var ExportOperation = operation

// ExportRequest calls request.
func (c *Client) ExportRequest(
	ctx context.Context,
	typ job.Type,
	target string,
	resolved string,
	category string,
	op job.OperationType,
	data any,
) (*job.Request, error) {
	return c.request(ctx, typ, target, resolved, category, op, data)
}

// SetNewSealer replaces the function that makes the Sealer of a job.
func SetNewSealer(
	fn func(jobID string) (*seal.Sealer, error),
) {
	newSealer = fn
}

// ResetNewSealer restores the function that makes the Sealer of a job.
func ResetNewSealer() {
	newSealer = seal.New
}
