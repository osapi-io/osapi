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

package command

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

// ShutdownTimeout is how long a process's components have to stop, after
// its context ends.
var ShutdownTimeout = 10 * time.Second

// Component is a part of a process, such as the NATS server or the agent.
type Component interface {
	// Start starts the component, and does not block.
	Start()
	// Stop stops the component, in the time that ctx gives.
	Stop(ctx context.Context)
}

// Running is what a process's Setup made: the components that Serve starts
// and stops, and what it does after they stop, such as flush the traces.
type Running struct {
	Components []Component
	Cleanup    func(ctx context.Context)
}

// Serve runs the process method of g. The context is the command's: one
// context, which the root makes and cancels on SIGINT or SIGTERM. Setup gets
// it, so a signal stops a setup that waits. Serve starts each component,
// waits until the context ends, then stops the components together, and then
// runs the cleanup. The stop and the cleanup get the context without its
// cancel, which keeps its values, with ShutdownTimeout. An error of Setup is
// the command's error. The generated command of each process calls it.
func Serve(
	cmd *cobra.Command,
	g *LocalGroup,
	method string,
) error {
	m := g.Method(method)
	if m == nil {
		return fmt.Errorf("the %s group has no method %s", g.Name, method)
	}
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	d, err := g.deps()
	if err != nil {
		return err
	}

	results := reflect.ValueOf(m.handler).Call([]reflect.Value{reflect.ValueOf(ctx), d})
	if err, _ := results[1].Interface().(error); err != nil {
		return err
	}
	run(ctx, results[0].Interface().(Running))

	return nil
}

// run starts r's components, waits until ctx ends, and stops them.
func run(
	ctx context.Context,
	r Running,
) {
	for _, c := range r.Components {
		c.Start()
	}

	<-ctx.Done()

	stopCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), ShutdownTimeout)
	defer cancel()

	var wg sync.WaitGroup
	for _, c := range r.Components {
		wg.Go(func() { c.Stop(stopCtx) })
	}
	wg.Wait()

	if r.Cleanup != nil {
		r.Cleanup(stopCtx)
	}
}
