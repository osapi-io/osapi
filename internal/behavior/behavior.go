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

package behavior

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"
	"slices"

	"github.com/osapi-io/osapi/internal/validation"
	"github.com/osapi-io/osapi/pkg/sdk/client"
)

// Placement says where a behavior's work happens.
type Placement string

// Node is the only placement: the work runs on a host, through the job
// system. What the controller answers itself is not a behavior.
const NodePlacement Placement = "node"

// Verb says whether a method reads the host or changes it. It decides the HTTP
// method, the permission, and whether the result carries changed.
type Verb string

const (
	// VerbRead reads the host and changes nothing.
	VerbRead Verb = "read"
	// VerbWrite changes the host. Its output reports whether it did.
	VerbWrite Verb = "write"
	// VerbExecute is the permission verb of an action, such as a restart or
	// a reboot: it makes something happen on the host, but it does not change
	// what the host is configured to be. An action is a write in each other
	// way; only its permission is <name>:execute, which the operator role
	// gets and the automation role does not.
	VerbExecute Verb = "execute"
	// VerbShell is the permission verb of an action declared with .Shell:
	// it runs a shell command line, which reaches anything the host can
	// do. Only the admin role gets a shell permission.
	VerbShell Verb = "shell"
)

// Behavior is one area of a host that osapi manages, such as kernel
// parameters or DNS resolver settings.
type Behavior struct {
	// Name is the behavior's name: the path segment, the CLI command and the
	// permission prefix when it has no group.
	Name string
	// Doc says what the behavior manages, in one line.
	Doc string
	// Placement is where the work runs.
	Placement Placement
	// Group nests the behavior under another path and command, such as
	// network for dns. Empty sits it directly under node.
	Group string
	// Area is where the UI shows the behavior, one of Areas. Area sets it.
	Area AreaName
	// Parent is the behavior whose items this one's items are under, set
	// with Of: a user's SSH keys are under the user. Nil for a behavior of
	// the host.
	Parent *Behavior
	// API serves the behavior over REST: routes, the OpenAPI spec and its
	// API docs. Without it the behavior is reachable only through another
	// behavior or the job system.
	API bool
	// CLI gives the behavior client commands and CLI docs pages. The
	// commands call the API, so CLI needs API.
	CLI bool
	// Requires are the behaviors this one is built on. The agent builds each
	// one's provider first and hands it to this one's builder, and this
	// one's methods call theirs as Go calls on the same host.
	Requires []*Behavior
	// Internal marks a behavior only other behaviors use: it has no route,
	// no command, no spec and no permissions.
	Internal bool
	// Methods are the operations, in the order they were declared.
	Methods []*Method
	// Defaults is the behavior's settings struct as declared with Config,
	// its field values the defaults. Nil when the behavior has none.
	Defaults any

	// settings is the decoded settings struct, set by the config loader.
	settings any
	// Platforms are how the behavior's provider is built on each platform it
	// supports, in the order declared. On any other platform its methods are
	// skipped.
	Platforms []*PlatformProvider
	// Sudo lists the programs the behavior's providers run with sudo, which
	// the agent's preflight checks it may run.
	Sudo []string
	// Runs lists the programs its providers run without sudo, declared with
	// Runs.
	Runs []string
	// PlatformRuns lists the programs its providers run on some platforms
	// only, declared with RunsOnPlatform.
	PlatformRuns []PlatformPrograms
	// ProviderSudo lists the programs that one platform's provider runs with
	// sudo, declared with SudoOnPlatform.
	ProviderSudo []PlatformPrograms
	// RunsIfPresent lists the programs its providers run only when the host
	// has them, declared with RunsIfPresent.
	RunsIfPresent []string
	// Writes lists the paths its providers write, remove or change the mode
	// of, as path.Match patterns, declared with Writes. "*" is any path.
	Writes []string
	// KeepsState is set when its providers keep a record in the file-state
	// bucket, declared with KeepsState.
	KeepsState bool
}

// Option configures a behavior in Define.
type Option interface {
	apply(b *Behavior)
}

type optionFunc func(b *Behavior)

func (f optionFunc) apply(
	b *Behavior,
) {
	f(b)
}

// Node places a behavior on the hosts, through the job system.
var Node Option = optionFunc(func(b *Behavior) { b.Placement = NodePlacement })

// API serves a behavior over REST: its routes, its part of the OpenAPI
// spec, and its API reference pages.
var API Option = optionFunc(func(b *Behavior) { b.API = true })

// CLI gives a behavior client commands, generated into cmd/next, and CLI
// docs pages. It needs API, because the commands call it.
var CLI Option = optionFunc(func(b *Behavior) { b.CLI = true })

// GroupName is a group that behaviors nest under. The group is a path
// segment, a CLI command and the prefix of the permissions of each behavior
// in it. Thus a misspelled group makes a new route, command and permission
// with no error, and behavior validate refuses a name outside KnownGroups.
type GroupName string

// The groups that behaviors nest under.
const (
	// GroupNetwork holds the network behaviors: dns, iface and route.
	GroupNetwork GroupName = "network"
	// GroupContainer holds the container behaviors: docker and image.
	GroupContainer GroupName = "container"
)

// GroupLog holds the journal behaviors: entry and source.
const GroupLog GroupName = "log"

// KnownGroups are the groups that a behavior can be Under. Add a constant
// here to make a new group.
func KnownGroups() []GroupName {
	return []GroupName{GroupNetwork, GroupLog, GroupContainer}
}

// Under nests a behavior inside a group: Under(GroupNetwork) puts dns at
// /api/node/{hostname}/network/dns and `osapi client node network dns`.
func Under(
	group GroupName,
) Option {
	return optionFunc(func(b *Behavior) { b.Group = string(group) })
}

// AreaName is an area of the UI, one of Areas.
type AreaName string

// The areas of the UI.
const (
	AreaServices   AreaName = "services"
	AreaSoftware   AreaName = "software"
	AreaConfig     AreaName = "config"
	AreaNode       AreaName = "node"
	AreaNetworking AreaName = "networking"
	AreaSecurity   AreaName = "security"
	AreaCommand    AreaName = "command"
	AreaHardware   AreaName = "hardware"
	AreaContainers AreaName = "containers"
)

// Areas are the areas of the UI that a served behavior can sit in: the
// groups of the Configure page's Add Block panel.
var Areas = []AreaName{
	AreaServices, AreaSoftware, AreaConfig, AreaNode, AreaNetworking, AreaSecurity, AreaCommand,
	AreaHardware, AreaContainers,
}

// Area puts a served behavior in an area of the UI: Area(AreaHardware) puts
// disk's blocks under Hardware in the Configure page's Add Block panel.
// behavior validate refuses a behavior with API and no area, or an area
// that is not one of Areas.
func Area(
	name AreaName,
) Option {
	return optionFunc(func(b *Behavior) { b.Area = name })
}

// Of puts a behavior's items under an item of parent: Of(user.Behavior)
// puts sshkey at /api/node/{hostname}/user/{name}/sshkey and
// `osapi client node user sshkey`. Each input of the behavior has the
// parent's key, such as name, which goes in the path. Its permissions are
// the parent's, so the keys of a user need user:read and user:write.
func Of(
	parent *Behavior,
) Option {
	return optionFunc(func(b *Behavior) { b.Parent = parent })
}

// ParentKey is the key of the parent's items, such as name for a user: the
// key that the parent's keyed methods have. Empty for a behavior with no
// parent, or a parent with no keyed method.
func (b *Behavior) ParentKey() string {
	if b.Parent == nil {
		return ""
	}
	for _, m := range b.Parent.Methods {
		if m.KeyField != "" {
			return m.KeyField
		}
	}

	return ""
}

// base is the route of the behavior's collection: the group and the name,
// or the parent's item and the name.
func (b *Behavior) base() string {
	if b.Parent != nil {
		return b.Parent.base() + "/{" + b.ParentKey() + "}/" + b.Name
	}
	if b.Group != "" {
		return b.Group + "/" + b.Name
	}

	return b.Name
}

// words are the words of the behavior's command after osapi client node.
func (b *Behavior) words() string {
	if b.Parent != nil {
		return b.Parent.words() + " " + b.Name
	}
	if b.Group != "" {
		return b.Group + " " + b.Name
	}

	return b.Name
}

// Define declares a behavior. Methods are options too, so a definition reads
// top to bottom as name, description, placement and methods.
func Define(
	name string,
	doc string,
	opts ...Option,
) *Behavior {
	b := &Behavior{Name: name, Doc: doc}
	for _, o := range opts {
		o.apply(b)
	}
	for _, m := range b.Methods {
		m.owner = b
	}

	return b
}

// Permission is the permission a method needs: the top-level group, or the
// behavior's name without one, and the verb. sysctl get needs sysctl:read and
// dns under network needs network:read.
func (b *Behavior) Permission(
	v Verb,
) client.Permission {
	if b.Parent != nil {
		return b.Parent.Permission(v)
	}
	prefix := b.Name
	if b.Group != "" {
		prefix = b.Group
	}

	return client.Permission(prefix + ":" + string(v))
}

// Method returns the method with the given name, or nil.
func (b *Behavior) Method(
	name string,
) *Method {
	for _, m := range b.Methods {
		if m.Name == name {
			return m
		}
	}

	return nil
}

// Operation is the job operation name the agent dispatches on, such as
// sysctl.get.
func (b *Behavior) Operation(
	m *Method,
) string {
	return b.Name + "." + m.Name
}

// Method is one operation of a behavior.
type Method struct {
	// Name is the method's name: the CLI subcommand and the operation suffix.
	Name string
	// Doc says what the method does, in one line. It is the command's short
	// help, the operation's summary, and the docs page's first line.
	Doc string
	// Long is what a reader needs beyond Doc: what the method touches on the
	// host and how it behaves when run twice. It is the command's long help
	// and the body of its docs page. Optional.
	Long string
	// Verb is read or write. An action is a write.
	Verb Verb
	// Admin is true for a method declared with .AdminOnly: only the admin
	// role gets its permission, not write or read.
	Admin bool
	// Action is true for a write declared with Action: it is named for what
	// it does, such as restart, and its route ends in its name.
	Action bool
	// Always is true for an action declared with .Always: it changes the
	// host each time it runs, as restart does, so no run is unchanged.
	AlwaysChanges bool
	// Disruptive is true for a method declared with .Disrupts: it takes the
	// host down, as a reboot does, so its examples never run on a live host.
	Disruptive bool
	// Disruption is why, as .Disrupts gives it.
	Disruption string
	// UsesShell is true for an action declared with .Shell: its permission
	// is <name>:shell, which only the admin role gets.
	UsesShell bool
	// NonFact is why a read runs a command outside gohai, set with
	// .NotAFact: what it reads is not a fact of the host.
	NonFact string
	// KeyField is the input field that goes in the path, by its JSON name.
	// Empty means the method addresses the whole collection.
	KeyField string
	// In and Out are the method's input and output types. An output that is a
	// slice becomes one result row per element, each with the host's name.
	In  reflect.Type
	Out reflect.Type
	// Examples are sample calls for the docs pages.
	Examples []Example
	// Refusals are calls the method fails, for the docs pages and the
	// provider's tests.
	Refusals []Refusal
	// Platforms limits the method to these of its behavior's platforms. Empty
	// means every platform the behavior declares. On any other platform the
	// method reports skipped, and its examples do not run there.
	Platforms []PlatformName
	// SudoPrograms and RunsPrograms are the programs only this method runs,
	// with and without sudo. A host needs them only where the method runs.
	SudoPrograms []string
	RunsPrograms []string

	provider reflect.Type
	owner    *Behavior
	// invoke decodes the input, calls the method, and returns the decoded
	// input beside the output so telemetry can read its tagged fields.
	invoke func(ctx context.Context, provider any, input []byte) (out any, in any, err error)
	// fn names the Go function of the method, such as
	// github.com/osapi-io/osapi/internal/behaviors/managedfile.Update, so a
	// test can find a direct call of it that skips Call.
	fn string
	// before is the controller step that .Before sets, or the zero Value.
	before reflect.Value
}

func (m *Method) apply(
	b *Behavior,
) {
	b.Methods = append(b.Methods, m)
}

// Key names the input field, by its JSON name, that is addressed in the path.
func (m *Method) Key(
	field string,
) *Method {
	m.KeyField = field

	return m
}

// PathKeys are the input fields that go in the method's path, by their JSON
// names: the parent's key for a behavior with Of, then the method's own key.
func (m *Method) PathKeys() []string {
	var keys []string
	if m.owner != nil {
		if k := m.owner.ParentKey(); k != "" {
			keys = append(keys, k)
		}
	}
	if m.KeyField != "" {
		keys = append(keys, m.KeyField)
	}

	return keys
}

// InPath reports whether the input field name, by its JSON name, goes in
// the method's path.
func (m *Method) InPath(
	name string,
) bool {
	return slices.Contains(m.PathKeys(), name)
}

// On limits the method to platforms, as Platform names them: a hostname
// read can run on every platform its behavior declares while the update runs
// only on debian. Elsewhere the method reports skipped.
func (m *Method) On(
	platforms ...PlatformName,
) *Method {
	m.Platforms = append(m.Platforms, platforms...)

	return m
}

// Sudo declares programs only this method runs with sudo, such as
// hostnamectl for a hostname update limited with On. The behavior's Sudo is
// for programs every method can run.
func (m *Method) Sudo(
	programs ...string,
) *Method {
	m.SudoPrograms = append(m.SudoPrograms, programs...)

	return m
}

// Runs declares programs only this method runs without sudo.
func (m *Method) Runs(
	programs ...string,
) *Method {
	m.RunsPrograms = append(m.RunsPrograms, programs...)

	return m
}

// RunsOn reports whether the method runs on a host whose platform is
// platform: every platform without On, else one On names, or any platform
// but darwin when On names PlatformLinux.
func (m *Method) RunsOn(
	platform PlatformName,
) bool {
	if len(m.Platforms) == 0 {
		return true
	}
	for _, p := range m.Platforms {
		if covers(p, platform) {
			return true
		}
	}

	return false
}

// covers reports whether a declaration for platform p holds on a host whose
// platform is platform: the same platform, or linux for each platform that
// is not darwin.
func covers(
	p PlatformName,
	platform PlatformName,
) bool {
	return p == platform || (p == PlatformLinux && platform != PlatformDarwin)
}

// Always marks an action that changes the host each time it runs, such as
// restart: it reports changed: true on each run, and the generated test does
// not run it a second time to see changed: false. behavior validate refuses it
// on a method that is not an action.
func (m *Method) Always() *Method {
	m.AlwaysChanges = true

	return m
}

// Disrupts marks a method that takes the host down, such as a reboot, and
// says why. Its examples run on the fake host as each example does, but the
// integration test never runs them on a live host, because the host would
// go down under the test. behavior validate refuses it without a reason.
func (m *Method) Disrupts(
	why string,
) *Method {
	m.Disruptive = true
	m.Disruption = why

	return m
}

// NotAFact marks a read that runs a command outside gohai, and says why:
// what it reads is not a fact of the host, such as the updates that the
// package index has, which are what the repositories have. A read of a fact
// goes through gohai, with h.Sources(); a fact that gohai does not read is a
// change to gohai first.
func (m *Method) NotAFact(
	why string,
) *Method {
	m.NonFact = why

	return m
}

// Explain sets the method's longer description, for its command's help and
// its docs page.
func (m *Method) Explain(
	long string,
) *Method {
	m.Long = long

	return m
}

// Shell makes an action's permission <name>:shell, in place of
// <name>:execute: the action runs a shell command line. Only the admin role
// gets a shell permission. behavior validate refuses it on a method that is
// not an action.
func (m *Method) Shell() *Method {
	m.UsesShell = true

	return m
}

// AdminOnly makes the method's permission one that only the admin role
// grants. Each method of the behavior with that verb must agree.
func (m *Method) AdminOnly() *Method {
	m.Admin = true

	return m
}

// AdminRole reports whether only the admin role gets the method's
// permission: it is declared with .AdminOnly, or with .Shell, whose verb
// only the admin role gets.
func (m *Method) AdminRole() bool {
	return m.Admin || m.UsesShell
}

// Describe sets the method's one-line description.
func (m *Method) Describe(
	doc string,
) *Method {
	m.Doc = doc

	return m
}

// ProviderType is the type the method's function takes as its provider.
func (m *Method) ProviderType() reflect.Type {
	return m.provider
}

// Check decodes input as the method's input type and checks it against its
// validate tags, as the API does before it queues a job. The agent calls it
// before Invoke, so a job that did not come through the API, such as one
// published to NATS by hand, is held to the same rules.
func (m *Method) Check(
	ctx context.Context,
	input []byte,
) error {
	in := reflect.New(m.In)
	if len(input) > 0 {
		if err := json.Unmarshal(input, in.Interface()); err != nil {
			return fmt.Errorf("method %s: decode input: %w", m.Name, err)
		}
	}

	return check(ctx, m, in.Elem().Interface())
}

// check holds in, the method's input, to its validate tags.
func check(
	ctx context.Context,
	m *Method,
	in any,
) error {
	if reflect.ValueOf(in).Kind() != reflect.Struct {
		return nil
	}
	if msg, ok := validation.Struct(ctx, in); !ok {
		return fmt.Errorf("method %s: %s", m.Name, msg)
	}
	if msg, ok := CheckTemplates(in); !ok {
		return fmt.Errorf("method %s: %s", m.Name, msg)
	}

	return nil
}

// Invoke decodes input as the method's input type and calls it with
// provider, inside the span and metrics every behavior gets (see
// telemetry.go). The agent calls this after Check. Invoke itself does not
// validate: behaviortest sends each .Fails input through it to the provider,
// to prove the provider refuses it too.
func (m *Method) Invoke(
	ctx context.Context,
	provider any,
	input []byte,
) (any, error) {
	return m.observe(ctx, provider, input)
}

// Read declares a method that reads the host.
func Read[P any, In any, Out any](
	name string,
	fn func(ctx context.Context, p P, in In) (Out, error),
) *Method {
	return newMethod(name, VerbRead, fn)
}

// Write declares a method that changes the host. Its output type reports
// whether anything changed, through a Changed bool field tagged
// json:"changed"; behavior validate fails a write without one.
func Write[P any, In any, Out any](
	name string,
	fn func(ctx context.Context, p P, in In) (Out, error),
) *Method {
	return newMethod(name, VerbWrite, fn)
}

// Action declares a write named for what it does to the host, such as
// restart or reboot, which no create, update or delete says. Its permission
// is <group>:execute. It is a write in each other way: a modify job, and an
// output that reports Changed. Its route is POST on the method's path with its name after it,
// /service/{name}/restart, so two actions on one item have two routes. The
// name is one of the CLI's actions, osapi cli command spec get lists them.
func Action[P any, In any, Out any](
	name string,
	fn func(ctx context.Context, p P, in In) (Out, error),
) *Method {
	m := newMethod(name, VerbWrite, fn)
	m.Action = true

	return m
}

func newMethod[P any, In any, Out any](
	name string,
	verb Verb,
	fn func(ctx context.Context, p P, in In) (Out, error),
) *Method {
	return &Method{
		Name:     name,
		Verb:     verb,
		In:       reflect.TypeFor[In](),
		Out:      reflect.TypeFor[Out](),
		provider: reflect.TypeFor[P](),
		fn:       runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name(),
		invoke: func(ctx context.Context, provider any, input []byte) (any, any, error) {
			p, ok := provider.(P)
			if !ok {
				return nil, nil, fmt.Errorf(
					"method %s: provider is %T, want %s",
					name, provider, reflect.TypeFor[P](),
				)
			}

			var in In
			if len(input) > 0 {
				if err := json.Unmarshal(input, &in); err != nil {
					return nil, nil, fmt.Errorf("method %s: decode input: %w", name, err)
				}
			}

			out, err := fn(ctx, p, in)

			return out, in, err
		},
	}
}

// PermissionVerb is the verb of the method's permission: execute for an
// action, else its verb.
func (m *Method) PermissionVerb() Verb {
	if m.UsesShell {
		return VerbShell
	}
	if m.Action {
		return VerbExecute
	}

	return m.Verb
}

// HTTPMethod is the method's HTTP verb: GET for a read, POST for an action,
// and for a write POST, PUT or DELETE by the method's name.
func (m *Method) HTTPMethod() string {
	if m.Verb == VerbRead {
		return "GET"
	}
	if m.Action {
		return "POST"
	}

	switch m.Name {
	case "update":
		return "PUT"
	case "delete":
		return "DELETE"
	default:
		return "POST"
	}
}

// Path is the method's route: /api/node/{hostname}, the group and name, the
// key when the method addresses one item, and an action's name.
func (b *Behavior) Path(
	m *Method,
) string {
	p := "/api/node/{hostname}/" + b.base()
	if m.KeyField != "" {
		p += "/{" + m.KeyField + "}"
	}
	if m.Action {
		p += "/" + m.Name
	}

	return p
}

// Command is the method's CLI command, such as
// osapi client node network dns get.
func (b *Behavior) Command(
	m *Method,
) string {
	return "osapi client node " + b.words() + " " + m.Name
}
