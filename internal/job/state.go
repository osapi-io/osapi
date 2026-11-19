// Copyright (c) 2025 John Dewey

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

// EffectiveState is the scheduling state of an agent as the controller
// reports it, from the state the agent reported and whether its drain flag
// is set. A drain that the agent has not applied yet is Draining: the flag
// asks it to stop taking jobs, but it may still be running them, so it is
// not Cordoned until the agent says so. Otherwise it is the state the
// agent reported, and an agent that reported none is Ready.
func EffectiveState(
	reported string,
	drainFlag bool,
) string {
	if reported == "" {
		reported = AgentStateReady
	}
	if drainFlag && reported == AgentStateReady {
		return AgentStateDraining
	}

	return reported
}
