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

	"github.com/retr0h/osapi/internal/provider"
)

// Darwin implements the Provider interface for Darwin (macOS).
// All methods return ErrUnsupported as user/group management is not
// available on macOS.
type Darwin struct{}

// NewDarwinProvider factory to create a new Darwin instance.
func NewDarwinProvider() *Darwin {
	return &Darwin{}
}

// ListUsers returns ErrUnsupported on Darwin.
func (d *Darwin) ListUsers(
	_ context.Context,
) ([]User, error) {
	return nil, provider.ErrUnsupported
}

// GetUser returns ErrUnsupported on Darwin.
func (d *Darwin) GetUser(
	_ context.Context,
	_ string,
) (*User, error) {
	return nil, provider.ErrUnsupported
}

// CreateUser returns ErrUnsupported on Darwin.
func (d *Darwin) CreateUser(
	_ context.Context,
	_ CreateUserOpts,
) (*Result, error) {
	return nil, provider.ErrUnsupported
}

// UpdateUser returns ErrUnsupported on Darwin.
func (d *Darwin) UpdateUser(
	_ context.Context,
	_ string,
	_ UpdateUserOpts,
) (*Result, error) {
	return nil, provider.ErrUnsupported
}

// DeleteUser returns ErrUnsupported on Darwin.
func (d *Darwin) DeleteUser(
	_ context.Context,
	_ string,
) (*Result, error) {
	return nil, provider.ErrUnsupported
}

// ChangePassword returns ErrUnsupported on Darwin.
func (d *Darwin) ChangePassword(
	_ context.Context,
	_ string,
	_ string,
) (*Result, error) {
	return nil, provider.ErrUnsupported
}

// ListGroups returns ErrUnsupported on Darwin.
func (d *Darwin) ListGroups(
	_ context.Context,
) ([]Group, error) {
	return nil, provider.ErrUnsupported
}

// GetGroup returns ErrUnsupported on Darwin.
func (d *Darwin) GetGroup(
	_ context.Context,
	_ string,
) (*Group, error) {
	return nil, provider.ErrUnsupported
}

// CreateGroup returns ErrUnsupported on Darwin.
func (d *Darwin) CreateGroup(
	_ context.Context,
	_ CreateGroupOpts,
) (*GroupResult, error) {
	return nil, provider.ErrUnsupported
}

// UpdateGroup returns ErrUnsupported on Darwin.
func (d *Darwin) UpdateGroup(
	_ context.Context,
	_ string,
	_ UpdateGroupOpts,
) (*GroupResult, error) {
	return nil, provider.ErrUnsupported
}

// DeleteGroup returns ErrUnsupported on Darwin.
func (d *Darwin) DeleteGroup(
	_ context.Context,
	_ string,
) (*GroupResult, error) {
	return nil, provider.ErrUnsupported
}

// ListKeys returns ErrUnsupported on Darwin.
func (d *Darwin) ListKeys(
	_ context.Context,
	_ string,
) ([]SSHKey, error) {
	return nil, provider.ErrUnsupported
}

// AddKey returns ErrUnsupported on Darwin.
func (d *Darwin) AddKey(
	_ context.Context,
	_ string,
	_ SSHKey,
) (*SSHKeyResult, error) {
	return nil, provider.ErrUnsupported
}

// RemoveKey returns ErrUnsupported on Darwin.
func (d *Darwin) RemoveKey(
	_ context.Context,
	_ string,
	_ string,
) (*SSHKeyResult, error) {
	return nil, provider.ErrUnsupported
}
