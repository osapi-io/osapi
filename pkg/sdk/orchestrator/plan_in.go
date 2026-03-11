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

package orchestrator

// ScopedPlan routes provider operations through a RuntimeTarget.
type ScopedPlan struct {
	plan   *Plan
	target RuntimeTarget
}

// In creates a scoped plan context for the given runtime target.
func (p *Plan) In(
	target RuntimeTarget,
) *ScopedPlan {
	return &ScopedPlan{
		plan:   p,
		target: target,
	}
}

// Docker creates a DockerTarget bound to this plan.
// Panics if no ExecFn was provided via WithDockerExecFn option.
func (p *Plan) Docker(
	name string,
	image string,
) *DockerTarget {
	if p.dockerExecFn == nil {
		panic("orchestrator: Plan.Docker() called without WithDockerExecFn option")
	}

	return NewDockerTarget(name, image, p.dockerExecFn)
}

// Target returns the runtime target for this scoped plan.
func (sp *ScopedPlan) Target() RuntimeTarget {
	return sp.target
}

// TaskFunc creates a task on the parent plan within the target context.
func (sp *ScopedPlan) TaskFunc(
	name string,
	fn TaskFn,
) *Task {
	return sp.plan.TaskFunc(name, fn)
}

// TaskFuncWithResults creates a task with results within the target context.
func (sp *ScopedPlan) TaskFuncWithResults(
	name string,
	fn TaskFnWithResults,
) *Task {
	return sp.plan.TaskFuncWithResults(name, fn)
}
