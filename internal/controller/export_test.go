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

package controller

import (
	"context"
	"encoding/json"

	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/provider/process"
)

// ExportWriteRegistration exposes the private writeRegistration method for testing.
func ExportWriteRegistration(
	ctx context.Context,
	h *ComponentHeartbeat,
) {
	h.writeRegistration(ctx)
}

// ExportDeregister exposes the private deregister method for testing.
func ExportDeregister(
	h *ComponentHeartbeat,
) {
	h.deregister()
}

// ExportRegistryKey exposes the private registryKey method for testing.
func ExportRegistryKey(
	h *ComponentHeartbeat,
) string {
	return h.registryKey()
}

// SetHeartbeatThresholds sets the thresholds field on ComponentHeartbeat for testing.
func SetHeartbeatThresholds(
	h *ComponentHeartbeat,
	thresholds process.ConditionThresholds,
) {
	h.thresholds = thresholds
}

// SetHeartbeatSubComponents sets the subComponents field on ComponentHeartbeat for testing.
func SetHeartbeatSubComponents(
	h *ComponentHeartbeat,
	subComponents map[string]job.SubComponentInfo,
) {
	h.subComponents = subComponents
}

// SetHeartbeatMarshalFn overrides the heartbeatMarshalFn variable for testing.
func SetHeartbeatMarshalFn(
	fn func(any) ([]byte, error),
) {
	heartbeatMarshalFn = fn
}

// ResetHeartbeatMarshalFn restores the heartbeatMarshalFn variable to its default.
func ResetHeartbeatMarshalFn() {
	heartbeatMarshalFn = json.Marshal
}
