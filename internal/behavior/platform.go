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
	"fmt"
	"reflect"
	"slices"

	"github.com/osapi-io/gohai/pkg/gohai/family"
)

// PlatformProvider is how a behavior's provider is built on one platform.
type PlatformProvider struct {
	// Name is the platform, as gohai's family.Detect names it.
	Name string
	// Type is the provider type the builder returns.
	Type  reflect.Type
	build func(h any) (any, error)
}

// Platform declares how the behavior's provider is built on one platform, as
// gohai's family.Detect names it: debian, darwin and so on, or PlatformLinux for
// every Linux host without a provider of its own. build takes the
// host.Host from internal/behavior/host, which is kept out of this package so the
// SDK does not import what only the agent needs. The agent builds the provider
// for its own platform once, at startup, and hands it to every method. On a
// platform the behavior does not declare, or when build returns
// ErrUnsupported, its methods are skipped rather than failed.
//
// The provider P must implement the provider type of every method, which
// osapi behavior validate checks.
func Platform[H any, P any](
	name string,
	build func(h H) (P, error),
) Option {
	return optionFunc(func(b *Behavior) {
		b.Platforms = append(b.Platforms, &PlatformProvider{
			Name: name,
			Type: reflect.TypeFor[P](),
			build: func(h any) (any, error) {
				hh, ok := h.(H)
				if !ok {
					return nil, fmt.Errorf(
						"the builder takes a %s, not a %T",
						reflect.TypeFor[H](),
						h,
					)
				}

				return build(hh)
			},
		})
	})
}

// Sudo declares the programs the behavior's providers run with sudo. The
// agent's preflight checks it may run each one, and the behavior's docs page
// lists them.
func Sudo(
	programs ...string,
) Option {
	return optionFunc(func(b *Behavior) { b.Sudo = append(b.Sudo, programs...) })
}

// Runs declares the programs the behavior's providers run without sudo, such
// as ip or dpkg-query. The agent warns at startup when one is not installed,
// the integration test fails when its image lacks one, and the docs page lists
// them. A program declared with Sudo need not be declared again.
func Runs(
	programs ...string,
) Option {
	return optionFunc(func(b *Behavior) { b.Runs = append(b.Runs, programs...) })
}

// PlatformPrograms are the programs that a behavior runs on one platform
// only, such as vm_stat on darwin.
type PlatformPrograms struct {
	Platform string
	Programs []string
}

// RunsOnPlatform declares the programs the behavior's providers run without
// sudo on platform only. linux is each platform that is not darwin. Thus the
// agent and the integration test of a Linux host do not look for a program
// that only macOS has.
func RunsOnPlatform(
	platform string,
	programs ...string,
) Option {
	return optionFunc(func(b *Behavior) {
		b.PlatformRuns = append(
			b.PlatformRuns,
			PlatformPrograms{Platform: platform, Programs: programs},
		)
	})
}

// RunsIfPresent declares the programs the behavior's providers run only when
// the host has them, such as zfs: a provider that does not find one reads
// without it. The fake host permits them, and the agent and the integration
// test do not require them.
func RunsIfPresent(
	programs ...string,
) Option {
	return optionFunc(func(b *Behavior) { b.RunsIfPresent = append(b.RunsIfPresent, programs...) })
}

// Writes declares the paths the behavior's providers write, remove or change
// the mode of, as path.Match patterns such as "/etc/sysctl.d/osapi-*.conf".
// "*" is any path: a behavior such as managedfile, which writes the path its
// caller names, declares it, and each behavior that requires it declares its
// own paths. The examples fail on a path that is not declared, and the docs
// page and behavior get list them. Each such call is a file span in a trace.
func Writes(
	paths ...string,
) Option {
	return optionFunc(func(b *Behavior) { b.Writes = append(b.Writes, paths...) })
}

// KeepsState declares that the behavior's providers keep a record in the
// file-state bucket. Each read and write of it is a kv span in a trace. The
// examples fail when a behavior uses the bucket without it; a behavior that
// requires one that keeps state gets it from there.
var KeepsState Option = optionFunc(func(b *Behavior) { b.KeepsState = true })

// The platforms a behavior can name besides a distribution.
const (
	// PlatformLinux is any Linux host: the provider for a host whose own
	// platform the behavior does not declare, unless it is macOS. A behavior
	// that works the same on every distribution, such as one reading /proc,
	// declares it once.
	PlatformLinux = "linux"
	// PlatformDarwin is macOS.
	PlatformDarwin = family.Darwin
	// PlatformDebian is the Debian family, as gohai names it: Debian,
	// Ubuntu, Raspbian, Linux Mint and the others of Ohai's table.
	PlatformDebian = family.Debian
)

// Provider builds b's provider for platform with h, the host.Host the agent
// passes: the provider declared for platform itself, or else, on a platform
// other than macOS, the one declared for PlatformLinux. The agent runs on
// Linux and macOS only, so every platform name but darwin is a Linux
// distribution. It returns nil and no error when b declares neither.
func (b *Behavior) Provider(
	platform string,
	h any,
) (any, error) {
	pp := b.resolve(platform)
	if pp == nil {
		return nil, nil
	}

	p, err := pp.build(h)
	if err != nil {
		return nil, fmt.Errorf("%s provider on %s: %w", b.Name, platform, err)
	}

	return p, nil
}

// platform is the provider b declares for name, or nil.
func (b *Behavior) platform(
	name string,
) *PlatformProvider {
	for _, pp := range b.Platforms {
		if pp.Name == name {
			return pp
		}
	}

	return nil
}

// Serves reports whether the provider b declares for platform builds the
// provider m runs with on some platform: every provider without On, else
// the one each platform On names resolves to, as Provider resolves it.
func (b *Behavior) Serves(
	m *Method,
	platform string,
) bool {
	if len(m.Platforms) == 0 {
		return true
	}
	for _, p := range m.Platforms {
		if pp := b.resolve(p); pp != nil && pp.Name == platform {
			return true
		}
	}

	return false
}

// resolve is the provider b declares for platform itself, or else, on a
// platform other than macOS, the one declared for PlatformLinux, or nil.
func (b *Behavior) resolve(
	platform string,
) *PlatformProvider {
	pp := b.platform(platform)
	if pp == nil && platform != PlatformDarwin {
		pp = b.platform(PlatformLinux)
	}

	return pp
}

// Programs are the programs b runs on platform, with sudo and without: the
// behavior's own, and each method's that runs there, once each. An empty
// platform is every platform.
func (b *Behavior) Programs(
	platform string,
) (sudo []string, runs []string) {
	add := func(to []string, programs []string) []string {
		for _, p := range programs {
			if !slices.Contains(to, p) {
				to = append(to, p)
			}
		}

		return to
	}
	sudo = add(sudo, b.Sudo)
	runs = add(runs, b.Runs)
	for _, pp := range b.PlatformRuns {
		if platform == "" || covers(pp.Platform, platform) {
			runs = add(runs, pp.Programs)
		}
	}
	for _, m := range b.Methods {
		if platform == "" || m.RunsOn(platform) {
			sudo = add(sudo, m.SudoPrograms)
			runs = add(runs, m.RunsPrograms)
		}
	}

	return sudo, runs
}
