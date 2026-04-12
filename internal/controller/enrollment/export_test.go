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

package enrollment

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"
)

// SetMarshalFn overrides the marshalFn variable for testing.
func SetMarshalFn(
	fn func(any) ([]byte, error),
) {
	marshalFn = fn
}

// ResetMarshalFn restores the marshalFn variable to its default.
func ResetMarshalFn() {
	marshalFn = json.Marshal
}

// SetUnmarshalFn overrides the unmarshalFn variable for testing.
func SetUnmarshalFn(
	fn func([]byte, any) error,
) {
	unmarshalFn = fn
}

// ResetUnmarshalFn restores the unmarshalFn variable to its default.
func ResetUnmarshalFn() {
	unmarshalFn = json.Unmarshal
}

// SetNowFn overrides the nowFn variable for testing.
func SetNowFn(
	fn func() time.Time,
) {
	nowFn = fn
}

// ResetNowFn restores the nowFn variable to its default.
func ResetNowFn() {
	nowFn = time.Now
}

// ExportHandleEnrollmentRequest exposes handleEnrollmentRequest for testing.
func ExportHandleEnrollmentRequest(
	ctx context.Context,
	w *Watcher,
	msg *nats.Msg,
) {
	w.handleEnrollmentRequest(ctx, msg)
}

// KVPrefix returns the kvPrefix constant for testing.
func KVPrefix() string {
	return kvPrefix
}

// EnrollSubject exposes enrollSubject for testing.
func EnrollSubject(
	namespace string,
	suffix string,
) string {
	return enrollSubject(namespace, suffix)
}
