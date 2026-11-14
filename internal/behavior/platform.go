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
)

// PlatformProvider is how a behavior's provider is built on one platform.
type PlatformProvider struct {
	// Name is the platform, as platform.Detect names it.
	Name string
	// Type is the provider type the builder returns.
	Type  reflect.Type
	build func(h any) (any, error)
}

// Platform declares how the behavior's provider is built on one platform, as
// platform.Detect names it: debian, darwin and so on, or PlatformLinux for
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

// The platforms a behavior can name besides a distribution.
const (
	// PlatformLinux is any Linux host: the provider for a host whose own
	// platform the behavior does not declare, unless it is macOS. A behavior
	// that works the same on every distribution, such as one reading /proc,
	// declares it once.
	PlatformLinux = "linux"
	// PlatformDarwin is macOS.
	PlatformDarwin = "darwin"
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
	pp := b.platform(platform)
	if pp == nil && platform != PlatformDarwin {
		pp = b.platform(PlatformLinux)
	}
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
