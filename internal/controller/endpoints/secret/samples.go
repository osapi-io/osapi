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

package secret

import (
	"strings"
	"time"

	secretvalue "github.com/osapi-io/osapi/internal/secret"
)

// The sample answers of the examples, which the docs page of each command
// shows as the CLI prints them. The docs generator fixes the time at
// 2026-01-15 10:00 UTC.
var (
	sampleChanged = time.Date(2026, 1, 15, 9, 12, 0, 0, time.UTC)

	sampleVaults = VaultListResponse{
		TotalItems: 2,
		Items: []VaultInfo{
			{Name: "default", Type: "nats", Targets: []string{}, Answers: true},
			{Name: "prod", Type: "nats", Targets: []string{"group:prod"}, Answers: true},
		},
	}

	sampleInfo = Info{
		Vault:   "default",
		Key:     "registry-token",
		Version: 3,
		Changed: sampleChanged,
		By:      "ops@example.com",
	}

	sampleList = ListResponse{
		TotalItems: 2,
		Items: []Info{
			sampleInfo,
			{
				Vault:   "default",
				Key:     "db-password",
				Version: 1,
				Changed: time.Date(2026, 1, 12, 16, 40, 0, 0, time.UTC),
				By:      "ci@example.com",
			},
		},
	}

	sampleValue = Value{Value: "glpat-xxxxxxxxxxxxxxxxxxxx"}

	sampleCreate = Info{
		Vault:   "default",
		Key:     "registry-token",
		Version: 1,
		Changed: time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC),
		By:      "ops@example.com",
	}

	sampleUpdate = Info{
		Vault:   "default",
		Key:     "registry-token",
		Version: 4,
		Changed: time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC),
		By:      "ops@example.com",
	}
)

// body is a value of n bytes, for the inputs that Fails declares.
func body(
	n int,
) UpdateRequest {
	return UpdateRequest{Value: secretvalue.Value(strings.Repeat("v", n))}
}

// create is a create of key with a value of n bytes, for the inputs that
// Fails declares.
func create(
	key string,
	n int,
) CreateRequest {
	return CreateRequest{Key: key, Value: secretvalue.Value(strings.Repeat("v", n))}
}
