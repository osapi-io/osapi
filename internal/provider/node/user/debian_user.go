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
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
)

const (
	passwdFile     = "/etc/passwd"
	passwdFields   = 7
	passwdLocked   = "L"
	passwdStatusFn = 2
)

// ListUsers returns all user accounts from /etc/passwd.
func (d *Debian) ListUsers(
	ctx context.Context,
) ([]User, error) {
	_ = ctx

	entries, err := d.parsePasswd()
	if err != nil {
		return nil, fmt.Errorf("user: %w", err)
	}

	var users []User
	for _, u := range entries {
		groups, locked, err := d.getUserDetails(u.Name)
		if err != nil {
			d.logger.Warn("failed to get user details",
				slog.String("name", u.Name),
				slog.String("error", err.Error()),
			)
		}

		u.Groups = groups
		u.Locked = locked
		users = append(users, u)
	}

	return users, nil
}

// GetUser returns a single user account by name.
func (d *Debian) GetUser(
	ctx context.Context,
	name string,
) (*User, error) {
	_ = ctx

	entries, err := d.parsePasswd()
	if err != nil {
		return nil, fmt.Errorf("user: %w", err)
	}

	for _, u := range entries {
		if u.Name == name {
			groups, locked, detailErr := d.getUserDetails(u.Name)
			if detailErr != nil {
				d.logger.Warn("failed to get user details",
					slog.String("name", u.Name),
					slog.String("error", detailErr.Error()),
				)
			}

			u.Groups = groups
			u.Locked = locked

			return &u, nil
		}
	}

	return nil, fmt.Errorf("user: %q not found", name)
}

// CreateUser creates a new user account.
func (d *Debian) CreateUser(
	ctx context.Context,
	opts CreateUserOpts,
) (*Result, error) {
	_ = ctx

	args := d.buildUseraddArgs(opts)

	_, err := d.execManager.RunPrivilegedCmd("useradd", args)
	if err != nil {
		return nil, fmt.Errorf("user: useradd failed: %w", err)
	}

	if opts.Password != "" {
		if err := d.setPassword(opts.Name, opts.Password); err != nil {
			return nil, fmt.Errorf("user: set password failed: %w", err)
		}
	}

	d.logger.Info("user created",
		slog.String("name", opts.Name),
	)

	return &Result{
		Name:    opts.Name,
		Changed: true,
	}, nil
}

// UpdateUser modifies an existing user account.
func (d *Debian) UpdateUser(
	ctx context.Context,
	name string,
	opts UpdateUserOpts,
) (*Result, error) {
	_ = ctx

	args := d.buildUsermodArgs(name, opts)
	if len(args) == 0 {
		return &Result{
			Name:    name,
			Changed: false,
		}, nil
	}

	_, err := d.execManager.RunPrivilegedCmd("usermod", args)
	if err != nil {
		return nil, fmt.Errorf("user: usermod failed: %w", err)
	}

	d.logger.Info("user updated",
		slog.String("name", name),
	)

	return &Result{
		Name:    name,
		Changed: true,
	}, nil
}

// DeleteUser removes a user account and its home directory.
func (d *Debian) DeleteUser(
	ctx context.Context,
	name string,
) (*Result, error) {
	_ = ctx

	_, err := d.execManager.RunPrivilegedCmd("userdel", []string{"-r", name})
	if err != nil {
		return nil, fmt.Errorf("user: userdel failed: %w", err)
	}

	d.logger.Info("user deleted",
		slog.String("name", name),
	)

	return &Result{
		Name:    name,
		Changed: true,
	}, nil
}

// ChangePassword changes a user's password.
func (d *Debian) ChangePassword(
	ctx context.Context,
	name string,
	password string,
) (*Result, error) {
	_ = ctx

	if err := d.setPassword(name, password); err != nil {
		return nil, fmt.Errorf("user: %w", err)
	}

	d.logger.Info("password changed",
		slog.String("name", name),
	)

	return &Result{
		Name:    name,
		Changed: true,
	}, nil
}

// parsePasswd reads and parses /etc/passwd into User entries.
func (d *Debian) parsePasswd() ([]User, error) {
	f, err := d.fs.Open(passwdFile)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", passwdFile, err)
	}
	defer func() { _ = f.Close() }()

	var users []User
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Split(line, ":")
		if len(fields) < passwdFields {
			continue
		}

		uid, err := strconv.Atoi(fields[2])
		if err != nil {
			continue
		}

		gid, err := strconv.Atoi(fields[3])
		if err != nil {
			continue
		}

		users = append(users, User{
			Name:  fields[0],
			UID:   uid,
			GID:   gid,
			Home:  fields[5],
			Shell: fields[6],
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", passwdFile, err)
	}

	return users, nil
}

// getUserDetails returns a user's supplementary groups and locked status.
func (d *Debian) getUserDetails(
	name string,
) ([]string, bool, error) {
	groupsOut, err := d.execManager.RunCmd("id", []string{"-Gn", name})
	if err != nil {
		return nil, false, fmt.Errorf("id -Gn %s: %w", name, err)
	}

	groups := strings.Fields(strings.TrimSpace(groupsOut))

	statusOut, err := d.execManager.RunCmd("passwd", []string{"-S", name})
	if err != nil {
		return groups, false, fmt.Errorf("passwd -S %s: %w", name, err)
	}

	locked := false
	statusFields := strings.Fields(strings.TrimSpace(statusOut))
	if len(statusFields) >= passwdStatusFn && statusFields[1] == passwdLocked {
		locked = true
	}

	return groups, locked, nil
}

// buildUseraddArgs constructs command-line arguments for useradd.
func (d *Debian) buildUseraddArgs(
	opts CreateUserOpts,
) []string {
	args := []string{"--create-home"}

	if opts.UID != 0 {
		args = append(args, "-u", strconv.Itoa(opts.UID))
	}

	if opts.GID != 0 {
		args = append(args, "-g", strconv.Itoa(opts.GID))
	}

	if opts.Home != "" {
		args = append(args, "-d", opts.Home)
	}

	if opts.Shell != "" {
		args = append(args, "-s", opts.Shell)
	}

	if len(opts.Groups) > 0 {
		args = append(args, "-G", strings.Join(opts.Groups, ","))
	}

	if opts.System {
		args = append(args, "-r")
	}

	args = append(args, opts.Name)

	return args
}

// buildUsermodArgs constructs command-line arguments for usermod.
func (d *Debian) buildUsermodArgs(
	name string,
	opts UpdateUserOpts,
) []string {
	var args []string

	if opts.Shell != "" {
		args = append(args, "-s", opts.Shell)
	}

	if opts.Home != "" {
		args = append(args, "-d", opts.Home, "-m")
	}

	if len(opts.Groups) > 0 {
		args = append(args, "-G", strings.Join(opts.Groups, ","))
	}

	if opts.Lock != nil {
		if *opts.Lock {
			args = append(args, "-L")
		} else {
			args = append(args, "-U")
		}
	}

	if len(args) == 0 {
		return nil
	}

	args = append(args, name)

	return args
}

// setPassword sets a user's password via chpasswd.
func (d *Debian) setPassword(
	name string,
	password string,
) error {
	input := fmt.Sprintf("%s:%s", name, password)

	_, err := d.execManager.RunPrivilegedCmd(
		"sh",
		[]string{"-c", fmt.Sprintf("echo '%s' | chpasswd", input)},
	)
	if err != nil {
		return fmt.Errorf("chpasswd failed: %w", err)
	}

	return nil
}
