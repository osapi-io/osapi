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
	groupFile   = "/etc/group"
	groupFields = 4
)

// ListGroups returns all system groups.
func (d *Debian) ListGroups(
	ctx context.Context,
) ([]Group, error) {
	_ = ctx

	groups, err := d.parseGroup()
	if err != nil {
		return nil, fmt.Errorf("group: %w", err)
	}

	return groups, nil
}

// GetGroup returns a single group by name.
func (d *Debian) GetGroup(
	ctx context.Context,
	name string,
) (*Group, error) {
	_ = ctx

	groups, err := d.parseGroup()
	if err != nil {
		return nil, fmt.Errorf("group: %w", err)
	}

	for _, g := range groups {
		if g.Name == name {
			return &g, nil
		}
	}

	return nil, fmt.Errorf("group: %q not found", name)
}

// CreateGroup creates a new system group.
func (d *Debian) CreateGroup(
	ctx context.Context,
	opts CreateGroupOpts,
) (*GroupResult, error) {
	_ = ctx

	args := d.buildGroupaddArgs(opts)

	_, err := d.execManager.RunCmd("groupadd", args)
	if err != nil {
		return nil, fmt.Errorf("group: groupadd failed: %w", err)
	}

	d.logger.Info("group created",
		slog.String("name", opts.Name),
	)

	return &GroupResult{
		Name:    opts.Name,
		Changed: true,
	}, nil
}

// UpdateGroup updates group membership.
func (d *Debian) UpdateGroup(
	ctx context.Context,
	name string,
	opts UpdateGroupOpts,
) (*GroupResult, error) {
	_ = ctx

	members := strings.Join(opts.Members, ",")

	_, err := d.execManager.RunCmd("gpasswd", []string{"-M", members, name})
	if err != nil {
		return nil, fmt.Errorf("group: gpasswd failed: %w", err)
	}

	d.logger.Info("group updated",
		slog.String("name", name),
	)

	return &GroupResult{
		Name:    name,
		Changed: true,
	}, nil
}

// DeleteGroup removes a system group.
func (d *Debian) DeleteGroup(
	ctx context.Context,
	name string,
) (*GroupResult, error) {
	_ = ctx

	_, err := d.execManager.RunCmd("groupdel", []string{name})
	if err != nil {
		return nil, fmt.Errorf("group: groupdel failed: %w", err)
	}

	d.logger.Info("group deleted",
		slog.String("name", name),
	)

	return &GroupResult{
		Name:    name,
		Changed: true,
	}, nil
}

// parseGroup reads and parses /etc/group into Group entries.
func (d *Debian) parseGroup() ([]Group, error) {
	f, err := d.fs.Open(groupFile)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", groupFile, err)
	}
	defer func() { _ = f.Close() }()

	var groups []Group
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Split(line, ":")
		if len(fields) < groupFields {
			continue
		}

		gid, err := strconv.Atoi(fields[2])
		if err != nil {
			continue
		}

		var members []string
		if fields[3] != "" {
			members = strings.Split(fields[3], ",")
		}

		groups = append(groups, Group{
			Name:    fields[0],
			GID:     gid,
			Members: members,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", groupFile, err)
	}

	return groups, nil
}

// buildGroupaddArgs constructs command-line arguments for groupadd.
func (d *Debian) buildGroupaddArgs(
	opts CreateGroupOpts,
) []string {
	var args []string

	if opts.GID != 0 {
		args = append(args, "-g", strconv.Itoa(opts.GID))
	}

	if opts.System {
		args = append(args, "-r")
	}

	args = append(args, opts.Name)

	return args
}
