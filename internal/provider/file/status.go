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

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/afero"

	"github.com/retr0h/osapi/internal/job"
)

// Status checks the current state of a deployed file against its expected
// SHA-256 from the file-state KV. Returns "in-sync" if the file matches,
// "drifted" if it differs, or "missing" if the file or state entry is absent.
func (p *FileProvider) Status(
	ctx context.Context,
	req StatusRequest,
) (*StatusResult, error) {
	stateKey := buildStateKey(p.hostname, req.Path)

	entry, err := p.stateKV.Get(ctx, stateKey)
	if err != nil {
		return &StatusResult{
			Path:   req.Path,
			Status: "missing",
		}, nil
	}

	var state job.FileState
	if err := json.Unmarshal(entry.Value(), &state); err != nil {
		return nil, fmt.Errorf("failed to parse file state: %w", err)
	}

	data, err := afero.ReadFile(p.fs, req.Path)
	if err != nil {
		return &StatusResult{
			Path:   req.Path,
			Status: "missing",
		}, nil
	}

	localSHA := computeSHA256(data)
	if localSHA == state.SHA256 {
		return &StatusResult{
			Path:   req.Path,
			Status: "in-sync",
			SHA256: localSHA,
		}, nil
	}

	return &StatusResult{
		Path:   req.Path,
		Status: "drifted",
		SHA256: localSHA,
	}, nil
}
