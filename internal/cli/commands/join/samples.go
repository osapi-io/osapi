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

package join

import "time"

// The outputs that the docs pages show.
var (
	sampleCreate = CreateOutput{
		ID:        "1a2b3c4d",
		Token:     "osapi_jt_1a2b3c4d.vC3m9...",
		Join:      "tls://nats.example.com:4222?fp=SHA256%3A9f86d08...&ns=osapi&pass=...&token=osapi_jt_1a2b3c4d.vC3m9...&user=join",
		ExpiresAt: time.Date(2026, 11, 25, 13, 0, 0, 0, time.UTC),
		Uses:      1,
	}

	sampleList = ListOutput{Items: []Row{{
		ID:        "1a2b3c4d",
		CreatedAt: time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC),
		ExpiresAt: time.Date(2026, 11, 25, 13, 0, 0, 0, time.UTC),
		UsesLeft:  1,
		Labels:    map[string]string{"group": "web"},
	}}}

	sampleRevoke = RevokeOutput{ID: "1a2b3c4d"}
)
