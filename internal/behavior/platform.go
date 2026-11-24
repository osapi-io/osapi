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
	Name PlatformName
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
// osapi dsl behavior validate checks.
func Platform[H any, P any](
	name PlatformName,
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
// them. A program declared with Sudo need not be declared again. AnyProgram,
// "*", is each program: a behavior such as command, which runs the program
// its caller names, declares it, and the agent and the integration test look
// for none.
func Runs(
	programs ...string,
) Option {
	return optionFunc(func(b *Behavior) { b.Runs = append(b.Runs, programs...) })
}

// PlatformPrograms are the programs that a behavior runs on one platform
// only, such as vm_stat on darwin.
type PlatformPrograms struct {
	Platform PlatformName
	Programs []string
}

// RunsOnPlatform declares the programs the behavior's providers run without
// sudo on platform only. linux is each platform that is not darwin. Thus the
// agent and the integration test of a Linux host do not look for a program
// that only macOS has.
func RunsOnPlatform(
	platform PlatformName,
	programs ...string,
) Option {
	return optionFunc(func(b *Behavior) {
		b.PlatformRuns = append(
			b.PlatformRuns,
			PlatformPrograms{Platform: platform, Programs: programs},
		)
	})
}

// SudoOnPlatform declares the programs that the provider declared for
// platform runs with sudo, such as update-ca-certificates on debian and
// update-ca-trust on linux. They hold on each host that this provider
// serves: linux holds on a Debian host only when the behavior declares no
// provider for debian. Thus the agent and the integration test do not ask a
// host for a program that its provider does not run.
func SudoOnPlatform(
	platform PlatformName,
	programs ...string,
) Option {
	return optionFunc(func(b *Behavior) {
		b.ProviderSudo = append(
			b.ProviderSudo,
			PlatformPrograms{Platform: platform, Programs: programs},
		)
	})
}

// ProviderSudoOn is the programs that SudoOnPlatform declares for the
// provider of a host whose platform is platform. An empty platform is every
// provider.
func (b *Behavior) ProviderSudoOn(
	platform PlatformName,
) []string {
	var out []string
	for _, ps := range b.ProviderSudo {
		if platform == "" {
			out = append(out, ps.Programs...)

			continue
		}
		if pp := b.resolve(platform); pp != nil && pp == b.platform(ps.Platform) {
			out = append(out, ps.Programs...)
		}
	}

	return out
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
// page and dsl behavior get list them. Each such call is a file span in a trace.
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

// AnyProgram is the program name Runs takes for each program.
const AnyProgram = "*"

// PlatformName is a platform that a behavior can name: PlatformLinux, or a
// platform as gohai's family.Detect names it. Use a constant: behavior
// validate fails on a name that is not one of KnownPlatforms, because a host
// never reports it and the declaration would never apply.
type PlatformName string

// KnownPlatforms are the platforms that a behavior can name. Add a constant
// here before a behavior names a new family of gohai.
func KnownPlatforms() []PlatformName {
	return []PlatformName{
		PlatformLinux,
		PlatformDarwin,
		PlatformDebian,
		PlatformSUSE,
	}
}

// The platforms a behavior can name.
const (
	// PlatformLinux is any Linux host: the provider for a host whose own
	// platform the behavior does not declare, unless it is macOS. A behavior
	// that works the same on every distribution, such as one reading /proc,
	// declares it once.
	PlatformLinux PlatformName = "linux"
	// PlatformDarwin is macOS.
	PlatformDarwin PlatformName = family.Darwin
	// PlatformDebian is the Debian family, as gohai names it: Debian,
	// Ubuntu, Raspbian, Linux Mint and the others of Ohai's table.
	PlatformDebian PlatformName = family.Debian
	// PlatformSUSE is the SUSE family, as gohai names it: openSUSE, SLES,
	// SLED and SLE Micro.
	PlatformSUSE PlatformName = family.SUSE
)

// Provider builds b's provider for platform with h, the host.Host the agent
// passes: the provider declared for platform itself, or else, on a platform
// other than macOS, the one declared for PlatformLinux. The agent runs on
// Linux and macOS only, so every platform name but darwin is a Linux
// distribution. It returns nil and no error when b declares neither.
func (b *Behavior) Provider(
	platform PlatformName,
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
	name PlatformName,
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
	platform PlatformName,
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

// HostFamilies are the platforms that gohai's family.Detect names, in the
// order the UI lists them. A Linux host whose platform has no family gets
// its own name, which is none of these.
var HostFamilies = []PlatformName{
	family.Darwin, family.Debian, family.RHEL, family.Amazon, family.SUSE,
	family.Fedora, family.WRLinux, family.Gentoo, family.Arch, family.Exherbo,
	family.Alpine, family.ClearLinux, family.Mandriva, family.Slackware,
}

// RunsOnHost reports whether m runs on a host whose platform is platform,
// as the agent decides: b has a provider there, and m runs there.
func (b *Behavior) RunsOnHost(
	m *Method,
	platform PlatformName,
) bool {
	return b.resolve(platform) != nil && m.RunsOn(platform)
}

// resolve is the provider b declares for platform itself, or else, on a
// platform other than macOS, the one declared for PlatformLinux, or nil.
func (b *Behavior) resolve(
	platform PlatformName,
) *PlatformProvider {
	pp := b.platform(platform)
	if pp == nil && platform != PlatformDarwin {
		pp = b.platform(PlatformLinux)
	}

	return pp
}

// Programs are the programs b runs on platform, with sudo and without: the
// behavior's own, and each method's that runs there, once each. An empty
// platform is every platform. A behavior that declares providers and has
// none for platform runs nothing there, so it needs no program there: an
// agent on Fedora does not ask for the apt-get of a Debian-only behavior.
func (b *Behavior) Programs(
	platform PlatformName,
) (sudo []string, runs []string) {
	if platform != "" && len(b.Platforms) > 0 && b.resolve(platform) == nil {
		return nil, nil
	}
	add := func(to []string, programs []string) []string {
		for _, p := range programs {
			if !slices.Contains(to, p) {
				to = append(to, p)
			}
		}

		return to
	}
	sudo = add(sudo, b.Sudo)
	sudo = add(sudo, b.ProviderSudoOn(platform))
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
