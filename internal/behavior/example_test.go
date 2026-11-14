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

package behavior_test

import (
	"context"
	"fmt"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
)

// MotdProvider is what each platform implements for the example behavior.
type MotdProvider interface {
	Read(ctx context.Context) (string, error)
	Write(ctx context.Context, text string) (bool, error)
}

// MotdGetInput is get's input; it has no fields.
type MotdGetInput struct{}

// MotdUpdateInput is update's input.
type MotdUpdateInput struct {
	Text string `json:"text" validate:"required,max=1024" doc:"The message to show at login"`
}

// Motd is get's output.
type Motd struct {
	Text string `json:"text" doc:"The message shown at login"`
}

// MotdResult is update's output. A write reports whether it changed the host.
type MotdResult struct {
	Changed bool `json:"changed" doc:"Whether the message changed"`
}

// MotdUserInput addresses one user's message.
type MotdUserInput struct {
	User string `json:"user" validate:"required" doc:"The user whose message to read"`
}

// MotdSettings are the behavior's standing agent config.
type MotdSettings struct {
	Path string `mapstructure:"path" doc:"File the message is written to" validate:"required,startswith=/"`
}

func getMotd(
	ctx context.Context,
	p MotdProvider,
	_ MotdGetInput,
) (Motd, error) {
	text, err := p.Read(ctx)

	return Motd{Text: text}, err
}

func getUserMotd(
	ctx context.Context,
	p MotdProvider,
	_ MotdUserInput,
) (Motd, error) {
	text, err := p.Read(ctx)

	return Motd{Text: text}, err
}

func updateMotd(
	ctx context.Context,
	p MotdProvider,
	in MotdUpdateInput,
) (MotdResult, error) {
	changed, err := p.Write(ctx, in.Text)

	return MotdResult{Changed: changed}, err
}

// A behavior is a definition. Everything served for it is derived from the
// definition: the route, the command and the permission of each method.
func Example() {
	motd := behavior.Define("motd", "The message shown at login",
		behavior.Node,
		behavior.API,
		behavior.CLI,
		behavior.Read("get", getMotd).
			Describe("Get the message of the day"),
		behavior.Write("update", updateMotd).
			Describe("Set the message of the day").
			Explain("Writes the file and reports changed: false when it already holds the text."),
	)

	for _, m := range motd.Methods {
		fmt.Println(m.HTTPMethod(), motd.Path(m), motd.Permission(m.Verb))
		fmt.Println("  ", motd.Command(m))
	}

	// Output:
	// GET /api/node/{hostname}/motd motd:read
	//    osapi client node motd get
	// PUT /api/node/{hostname}/motd motd:write
	//    osapi client node motd update
}

// Under nests a behavior in a group, which also names its permission.
func ExampleUnder() {
	dns := behavior.Define("dns", "Resolver settings",
		behavior.Node,
		behavior.Under("network"),
		behavior.Read("get", getMotd).Describe("Get the resolvers"),
	)

	m := dns.Methods[0]
	fmt.Println(dns.Path(m), dns.Permission(m.Verb))

	// Output:
	// /api/node/{hostname}/network/dns network:read
}

// Key puts an input field in the path, for a method that addresses one item.
func ExampleMethod_Key() {
	b := behavior.Define("motd", "The message shown at login",
		behavior.Node,
		behavior.Read("get", getUserMotd).Key("user").Describe("Get one user's message"),
	)

	m := b.Methods[0]
	fmt.Println(m.HTTPMethod(), b.Path(m))

	// Output:
	// GET /api/node/{hostname}/motd/{user}
}

// Config declares settings. Their keys sit under behaviors.<name>, each with
// an agent start flag, and a method reads the decoded struct with Settings.
func ExampleConfig() {
	b := behavior.Define("motd", "The message shown at login",
		behavior.Node,
		behavior.Config(MotdSettings{Path: "/etc/motd"}),
		behavior.Read("get", getMotd).Describe("Get the message of the day"),
	)

	fmt.Println(b.SettingsKey())
	fmt.Println("--" + b.SettingFlag("path"))

	// At startup the agent decodes the config and hands it over; a method
	// then reads it.
	b.SetSettings(MotdSettings{Path: "/etc/motd"})
	fmt.Println(behavior.Settings[MotdSettings](b).Path)

	// Output:
	// behaviors.motd
	// --motd-path
	// /etc/motd
}

// motdFile is the example's Linux provider: it keeps the message in a file.
type motdFile struct {
	path string
	text string
}

func (f *motdFile) Read(context.Context) (string, error) { return f.text, nil }

func (f *motdFile) Write(_ context.Context, text string) (bool, error) {
	changed := f.text != text
	f.text = text

	return changed, nil
}

// Platform declares how the provider is built on each platform. The agent
// builds the one for its own platform at startup, with its settings; on any
// other platform the behavior's methods are skipped.
func ExamplePlatform() {
	motd := behavior.Define("motd", "The message shown at login",
		behavior.Node,
		behavior.Config(MotdSettings{Path: "/etc/motd"}),
		behavior.Platform("debian", func(h host.Host) (MotdProvider, error) {
			return &motdFile{path: host.Settings[MotdSettings](h).Path}, nil
		}),
		behavior.Sudo("tee"),
		behavior.Read("get", getMotd).Describe("Get the message of the day"),
	)
	motd.SetSettings(MotdSettings{Path: "/etc/motd"})

	// The agent puts the behavior's settings in the host it passes.
	h := host.Host{Settings: motd.CurrentSettings()}

	p, _ := motd.Provider("debian", h)
	fmt.Println(p.(*motdFile).path)

	p, _ = motd.Provider("darwin", h)
	fmt.Println(p == nil)

	// Output:
	// /etc/motd
	// true
}

// ValidateBehavior reports what a definition gets wrong, as osapi behavior
// validate does for every registered one.
func ExampleValidateBehavior() {
	b := behavior.Define("motd", "The message shown at login",
		behavior.Node,
		behavior.Read("get", getMotd),
	)

	for _, p := range behavior.ValidateBehavior(b) {
		fmt.Println(p.Rule)
	}

	// Output:
	// platform
	// doc
}
