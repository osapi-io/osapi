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

package types

// JobType is the type of a job operation, one of JobTypes.
type JobType string

const (
	// JobTypeQuery represents read operations that query system state.
	JobTypeQuery JobType = "query"
	// JobTypeModify represents write operations that modify system state.
	JobTypeModify JobType = "modify"
)

// JobTypes are each type of a job operation.
var JobTypes = []JobType{JobTypeQuery, JobTypeModify}

// Target constants for job routing.
const (
	// TargetAny routes to any available agent (load-balanced).
	TargetAny = "_any"
	// TargetAll broadcasts to every agent.
	TargetAll = "_all"
)
