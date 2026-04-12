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

import "encoding/json"

// ExportSanitizeKeyForNATS exposes the private sanitizeKeyForNATS for testing.
func ExportSanitizeKeyForNATS(
	s string,
) string {
	return sanitizeKeyForNATS(s)
}

// ExportWrapInSignedEnvelope exposes the private wrapInSignedEnvelope for testing.
func ExportWrapInSignedEnvelope(
	signer PKISigner,
	payload []byte,
) ([]byte, error) {
	return wrapInSignedEnvelope(signer, payload)
}

// ExportUnwrapSignedEnvelope exposes the private unwrapSignedEnvelope for testing.
func ExportUnwrapSignedEnvelope(
	data []byte,
	pubKey []byte,
) ([]byte, bool, error) {
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
