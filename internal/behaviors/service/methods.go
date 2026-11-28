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

package service

import "context"

// List is every service of the host.
func List(
	ctx context.Context,
	p Provider,
	_ ListInput,
) ([]Entry, error) {
	return p.Services(ctx)
}

// Get is one service of the host.
func Get(
	ctx context.Context,
	p Provider,
	in GetInput,
) (Entry, error) {
	return p.Service(ctx, in.Name)
}

// Create deploys a unit file that osapi does not have yet.
func Create(
	ctx context.Context,
	p Provider,
	in CreateInput,
) (Result, error) {
	changed, err := p.Deploy(ctx, Unit(in), true)

	return Result{Name: in.Name, Changed: changed}, err
}

// Update deploys a unit file over the one osapi has.
func Update(
	ctx context.Context,
	p Provider,
	in UpdateInput,
) (Result, error) {
	changed, err := p.Deploy(ctx, Unit(in), false)

	return Result{Name: in.Name, Changed: changed}, err
}

// Delete stops and disables a service that osapi deployed, and drops its
// unit file.
func Delete(
	ctx context.Context,
	p Provider,
	in DeleteInput,
) (Result, error) {
	changed, err := p.Remove(ctx, in.Name)

	return Result{Name: in.Name, Changed: changed}, err
}

// act runs one action on a service that osapi manages.
func act(
	ctx context.Context,
	p Provider,
	action string,
	name string,
) (Result, error) {
	changed, err := p.Act(ctx, action, name)

	return Result{Name: name, Changed: changed}, err
}

// Start starts a service that osapi manages.
func Start(
	ctx context.Context,
	p Provider,
	in StartInput,
) (Result, error) {
	return act(ctx, p, "start", in.Name)
}

// Stop stops a service that osapi manages.
func Stop(
	ctx context.Context,
	p Provider,
	in StopInput,
) (Result, error) {
	return act(ctx, p, "stop", in.Name)
}

// Restart restarts a service that osapi manages.
func Restart(
	ctx context.Context,
	p Provider,
	in RestartInput,
) (Result, error) {
	return act(ctx, p, "restart", in.Name)
}

// Enable makes a service that osapi manages start at boot.
func Enable(
	ctx context.Context,
	p Provider,
	in EnableInput,
) (Result, error) {
	return act(ctx, p, "enable", in.Name)
}

// Disable makes a service that osapi manages not start at boot.
func Disable(
	ctx context.Context,
	p Provider,
	in DisableInput,
) (Result, error) {
	return act(ctx, p, "disable", in.Name)
}
