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

package job

// redactedPlaceholder replaces a sensitive field's value in an API response.
const redactedPlaceholder = "[redacted]"

// sensitiveOperationFields lists operation-data keys that must never reach
// an API response. "password" covers jobs the user domain stored before
// passwords were hashed at the controller; "password_hash" covers every job
// stored after it. Listing both means a job's age does not matter.
var sensitiveOperationFields = map[string]struct{}{
	"password":      {},
	"password_hash": {},
}

// redactOperation returns a copy of a job's operation data with any
// sensitive field masked, at any nesting depth. It is safe to call with nil.
func redactOperation(
	operation map[string]interface{},
) map[string]interface{} {
	if operation == nil {
		return nil
	}

	redacted, _ := redactValue(operation).(map[string]interface{})

	return redacted
}

// redactValue walks v, masking any map entry whose key is a sensitive
// operation field and recursing into maps and slices.
func redactValue(
	v interface{},
) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		out := make(map[string]interface{}, len(val))
		for k, child := range val {
			if _, sensitive := sensitiveOperationFields[k]; sensitive {
				out[k] = redactedPlaceholder
				continue
			}
			out[k] = redactValue(child)
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(val))
		for i, child := range val {
			out[i] = redactValue(child)
		}
		return out
	default:
		return v
	}
}
