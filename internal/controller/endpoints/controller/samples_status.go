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

import "github.com/osapi-io/osapi/internal/controller/endpoints/health"

// sampleStatus is the sample answer of the status example, which the docs
// page of the command shows as the CLI prints it.
var sampleStatus = health.ComponentList{
	TotalItems: 1,
	Items: []health.ComponentEntry{
		{
			Type:       "controller",
			Hostname:   "osapi-01",
			Status:     "ok",
			Age:        "51h3m",
			CPUPercent: 1.2,
			MemBytes:   52428800,
			Version:    "v1.4.0",
			Address:    "0.0.0.0:8080",
			Components: []health.ComponentEntry{
				{Type: "api", Status: "ok"},
				{Type: "metrics", Status: "ok"},
			},
		},
	},
}
