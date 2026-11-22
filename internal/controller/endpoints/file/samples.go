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

package file

// The sample answers of the examples, which the docs page of each command
// shows as the CLI prints them.
var (
	sampleSum = "a3f1c2e4b5d6978812f3e4d5c6b7a8f9e0d1c2b3a4f5e6d7c8b9a0f1e2d3c4b5"

	sampleList = FileListResponse{
		TotalItems: 2,
		Items: []FileInfo{
			{Name: "app.conf", Sha256: sampleSum, Size: 1024, ContentType: "raw", Source: "user"},
			{
				Name:        "app.conf.tmpl",
				Sha256:      "7d8e9fa0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8",
				Size:        512,
				ContentType: "template",
				Source:      "user",
			},
		},
	}

	sampleUpload = FileUploadResponse{
		Name:        "app.conf",
		Sha256:      sampleSum,
		Size:        1024,
		Changed:     true,
		ContentType: "raw",
	}

	sampleUploadTemplate = FileUploadResponse{
		Name:        "app.conf.tmpl",
		Sha256:      "7d8e9fa0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8",
		Size:        512,
		Changed:     true,
		ContentType: "template",
	}

	sampleInfo = FileInfoResponse{
		Name:        "app.conf",
		Sha256:      sampleSum,
		Size:        1024,
		ContentType: "raw",
	}

	sampleStale = StaleDeploymentsResponse{
		TotalItems: 1,
		Items: []StaleDeployment{{
			ObjectName:  "app.conf",
			Hostname:    "web-01",
			Provider:    "file",
			DeployedAt:  "2026-01-14T16:20:00Z",
			Path:        "/etc/app/app.conf",
			DeployedSha: "0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c",
			CurrentSha:  sampleSum,
		}},
	}
)
