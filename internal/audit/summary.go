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

package audit

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	// maxSummaryBytes caps what an audit entry keeps of a request. A file deploy
	// or a template can carry far more than an audit trail needs, and an entry
	// that grows without bound is an audit store that fills up.
	maxSummaryBytes = 2048

	// redacted replaces a value that must not be stored.
	redacted = "[redacted]"
)

// sensitiveFields are the field names whose values never enter the audit log,
// matched case-insensitively against the whole name.
//
// This is a deny list rather than an allow list because the point of the summary
// is to answer "what ran" after an incident, and an allow list would answer it
// only for the fields somebody remembered. New fields are therefore recorded by
// default, which puts the burden on this list: a field carrying a secret is added
// here in the same change that adds the field.
var sensitiveFields = map[string]bool{
	"password":      true,
	"password_hash": true,
	"passwordhash":  true,
	"secret":        true,
	"token":         true,
	"private_key":   true,
	"privatekey":    true,
	"key_data":      true,
	"stdin":         true,
	"content":       true,
	"authorization": true,
}

// Summarize returns what a request asked for, in a form an audit log can keep:
// the JSON body with sensitive values replaced and the whole thing capped.
//
// A body that is not JSON is described rather than stored, because an audit entry
// is not the place to keep an arbitrary upload. An empty body summarizes to
// nothing at all, so a request with no body adds no field.
func Summarize(
	body []byte,
	contentType string,
) string {
	if len(body) == 0 {
		return ""
	}

	if !strings.Contains(contentType, "json") {
		return fmt.Sprintf("[%s, %d bytes]", describeContentType(contentType), len(body))
	}

	var parsed any
	if err := json.Unmarshal(body, &parsed); err != nil {
		return fmt.Sprintf("[unparsable json, %d bytes]", len(body))
	}

	// marshalJSON is the package's existing seam, in stream_store.go. Reaching
	// this failure needs it: a value that came from json.Unmarshal always
	// marshals again.
	encoded, err := marshalJSON(redactValue(parsed))
	if err != nil {
		return fmt.Sprintf("[unsummarizable json, %d bytes]", len(body))
	}

	return truncate(string(encoded))
}

// describeContentType names a content type for the summary, without letting an
// arbitrary header into the log at arbitrary length.
func describeContentType(
	contentType string,
) string {
	if contentType == "" {
		return "no content type"
	}

	if i := strings.Index(contentType, ";"); i >= 0 {
		contentType = contentType[:i]
	}

	return truncateTo(strings.TrimSpace(contentType), 64)
}

// redactValue walks a decoded body and replaces the values of sensitive fields,
// at every depth, because a password nested inside a list of users is still a
// password.
func redactValue(
	v any,
) any {
	switch value := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(value))

		for k, inner := range value {
			if sensitiveFields[strings.ToLower(k)] {
				out[k] = redacted

				continue
			}

			out[k] = redactValue(inner)
		}

		return out
	case []any:
		out := make([]any, 0, len(value))

		for _, inner := range value {
			out = append(out, redactValue(inner))
		}

		return out
	}

	return v
}

// truncate caps a summary and says that it did, so a reader does not mistake a
// cut-off value for the whole one.
func truncate(
	s string,
) string {
	return truncateTo(s, maxSummaryBytes)
}

// truncateTo caps a string at n bytes, on a rune boundary.
func truncateTo(
	s string,
	n int,
) string {
	if len(s) <= n {
		return s
	}

	cut := n
	for cut > 0 && !utf8ValidCut(s, cut) {
		cut--
	}

	return s[:cut] + "…[truncated]"
}

// utf8ValidCut reports whether cutting s at i leaves valid UTF-8, so a truncated
// summary is still text.
func utf8ValidCut(
	s string,
	i int,
) bool {
	return i == 0 || (s[i]&0xC0) != 0x80
}
