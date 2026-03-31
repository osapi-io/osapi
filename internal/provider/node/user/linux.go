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

package user

import (
	"context"
	"fmt"

	"github.com/retr0h/osapi/internal/provider"
)

// Linux implements the Provider interface for generic Linux.
// All methods return ErrUnsupported; use the Debian provider for
// Debian-family systems.
type Linux struct{}

// NewLinuxProvider factory to create a new Linux instance.
func NewLinuxProvider() *Linux {
	return &Linux{}
}

// ListUsers returns ErrUnsupported on generic Linux.
func (l *Linux) ListUsers(
	_ context.Context,
) ([]User, error) {
	return nil, fmt.Errorf("user: %w", provider.ErrUnsupported)
}

// GetUser returns ErrUnsupported on generic Linux.
func (l *Linux) GetUser(
	_ context.Context,
	_ string,
) (*User, error) {
	return nil, fmt.Errorf("user: %w", provider.ErrUnsupported)
}

// CreateUser returns ErrUnsupported on generic Linux.
func (l *Linux) CreateUser(
	_ context.Context,
	_ CreateUserOpts,
) (*Result, error) {
	return nil, fmt.Errorf("user: %w", provider.ErrUnsupported)
}

// UpdateUser returns ErrUnsupported on generic Linux.
func (l *Linux) UpdateUser(
	_ context.Context,
	_ string,
	_ UpdateUserOpts,
) (*Result, error) {
	return nil, fmt.Errorf("user: %w", provider.ErrUnsupported)
}

// DeleteUser returns ErrUnsupported on generic Linux.
func (l *Linux) DeleteUser(
	_ context.Context,
	_ string,
) (*Result, error) {
	return nil, fmt.Errorf("user: %w", provider.ErrUnsupported)
}

// ChangePassword returns ErrUnsupported on generic Linux.
func (l *Linux) ChangePassword(
	_ context.Context,
	_ string,
	_ string,
) (*Result, error) {
	return nil, fmt.Errorf("user: %w", provider.ErrUnsupported)
}

// ListGroups returns ErrUnsupported on generic Linux.
func (l *Linux) ListGroups(
	_ context.Context,
) ([]Group, error) {
	return nil, fmt.Errorf("user: %w", provider.ErrUnsupported)
}

// GetGroup returns ErrUnsupported on generic Linux.
func (l *Linux) GetGroup(
	_ context.Context,
	_ string,
) (*Group, error) {
	return nil, fmt.Errorf("user: %w", provider.ErrUnsupported)
}

// CreateGroup returns ErrUnsupported on generic Linux.
func (l *Linux) CreateGroup(
	_ context.Context,
	_ CreateGroupOpts,
) (*GroupResult, error) {
	return nil, fmt.Errorf("user: %w", provider.ErrUnsupported)
}

// UpdateGroup returns ErrUnsupported on generic Linux.
func (l *Linux) UpdateGroup(
	_ context.Context,
	_ string,
	_ UpdateGroupOpts,
) (*GroupResult, error) {
	return nil, fmt.Errorf("user: %w", provider.ErrUnsupported)
}

// DeleteGroup returns ErrUnsupported on generic Linux.
func (l *Linux) DeleteGroup(
	_ context.Context,
	_ string,
) (*GroupResult, error) {
	return nil, fmt.Errorf("user: %w", provider.ErrUnsupported)
}
