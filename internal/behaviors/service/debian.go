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

package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/avfs/avfs"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/behaviors/managedfile"
	"github.com/osapi-io/osapi/internal/behaviors/object"
)

// unitDir is where systemd reads the units of the administrator, and prefix
// starts the name of each unit that osapi deploys. A service that osapi did
// not deploy has no unit file with the prefix, so osapi writes and acts on
// its own services only.
const (
	unitDir = "/etc/systemd/system"
	prefix  = "osapi-"
)

// check refuses a name that systemctl would read as an option.
func check(
	name string,
) error {
	if strings.HasPrefix(name, "-") {
		return fmt.Errorf("%q is not a valid service name", name)
	}

	return nil
}

// unitPath is the unit file of the service name.
func unitPath(
	name string,
) string {
	return unitDir + "/" + prefix + name + ".service"
}

// debian reads through gohai, gets each unit file from the object store with
// object, writes it with managedfile, and runs systemctl.
type debian struct {
	*reader
	objects object.Provider
	files   managedfile.Provider
	exec    host.Exec
	fs      avfs.VFS
}

// newDebian is the writer, and on an agent in a container, where osapi does
// not manage the services of the host, the reader.
func newDebian(
	h host.Host,
) (Provider, error) {
	if h.Container {
		return newReader(h)
	}

	return &debian{
		reader:  &reader{src: h.Sources()},
		objects: host.Require[object.Provider](h, object.Behavior),
		files:   host.Require[managedfile.Provider](h, managedfile.Behavior),
		exec:    h.Exec,
		fs:      h.FS,
	}, nil
}

// Services is each service of the host. A unit that osapi deployed reads by
// the name it was deployed with, as its writes and actions take it, not by
// the unit's name: app, not osapi-app. Its Managed is true.
func (d *debian) Services(
	ctx context.Context,
) ([]Entry, error) {
	list, err := d.reader.Services(ctx)
	for i, s := range list {
		name, ok := strings.CutPrefix(s.Name, prefix)
		if !ok {
			continue
		}
		owned, oerr := d.owned(ctx, unitPath(name))
		if oerr != nil {
			return nil, oerr
		}
		if owned {
			list[i].Name, list[i].Managed = name, true
		}
	}

	return list, err
}

// Service is one service, by the name that its writes take. A service that
// osapi deployed is found by that name, or by the unit's name with the
// prefix, and it wins over a unit of the host with the same name, as the
// writes act on it. Any other name is the unit's own.
func (d *debian) Service(
	ctx context.Context,
	name string,
) (Entry, error) {
	name = strings.TrimSuffix(name, ".service")
	bare := strings.TrimPrefix(name, prefix)
	owned, err := d.owned(ctx, unitPath(bare))
	if err != nil {
		return Entry{}, err
	}
	if !owned {
		return d.reader.Service(ctx, name)
	}
	s, err := d.reader.Service(ctx, prefix+bare)
	if err != nil {
		return Entry{}, err
	}
	s.Name, s.Managed = bare, true

	return s, nil
}

// Deploy writes the object of u as the unit file, and reloads systemd when
// the file changed. An update of a service that osapi did not deploy is not
// found.
func (d *debian) Deploy(
	ctx context.Context,
	u Unit,
	create bool,
) (bool, error) {
	if err := check(u.Name); err != nil {
		return false, err
	}
	path := unitPath(u.Name)
	owned, err := d.owned(ctx, path)
	if err != nil {
		return false, err
	}
	if !owned && !create {
		return false, fmt.Errorf("service %q: %w", u.Name, behavior.ErrNotFound)
	}
	if _, statErr := d.fs.Stat(path); statErr == nil && !owned {
		return false, fmt.Errorf("%s is on the host and osapi did not deploy it", path)
	}
	obj, err := behavior.Call(ctx, object.Behavior, object.Get, d.objects,
		object.GetInput{Name: u.Object})
	if err != nil {
		return false, err
	}
	result, err := behavior.Call(ctx, managedfile.Behavior, managedfile.Update, d.files,
		managedfile.UpdateInput{
			Path:     path,
			Content:  obj.Content,
			Mode:     "0644",
			Metadata: map[string]string{"object": u.Object},
			Object:   u.Object,
		})
	if err != nil || !result.Changed {
		return false, err
	}

	return true, d.systemctl(ctx, "daemon-reload")
}

// Remove stops and disables the service, drops its unit file and reloads
// systemd. A service with no unit file is unchanged.
func (d *debian) Remove(
	ctx context.Context,
	name string,
) (bool, error) {
	if err := check(name); err != nil {
		return false, err
	}
	path := unitPath(name)
	if owned, err := d.owned(ctx, path); err != nil || !owned {
		return false, err
	}
	unit := prefix + name + ".service"
	if err := d.systemctl(ctx, "disable", "--now", unit); err != nil {
		return false, err
	}
	if _, err := behavior.Call(ctx, managedfile.Behavior, managedfile.Delete, d.files,
		managedfile.DeleteInput{Path: path}); err != nil {
		return false, err
	}

	return true, d.systemctl(ctx, "daemon-reload")
}

// Act runs action on the service, unless the service is already where the
// action puts it. It reports changed when systemd says the service moved: an
// enable of a unit with no [Install] section, which systemd calls static,
// runs and changes nothing. restart always changes.
func (d *debian) Act(
	ctx context.Context,
	action string,
	name string,
) (bool, error) {
	if err := check(name); err != nil {
		return false, err
	}
	if owned, err := d.owned(ctx, unitPath(name)); err != nil || !owned {
		if err == nil {
			err = fmt.Errorf("service %q: %w", name, behavior.ErrNotFound)
		}

		return false, err
	}
	unit := prefix + name
	if done(action, d.state(ctx, unit)) {
		return false, nil
	}
	if err := d.systemctl(ctx, action, unit+".service"); err != nil {
		return false, err
	}

	return action == "restart" || done(action, d.state(ctx, unit)), nil
}

// done reports whether the service s is where action puts it. restart is
// never done.
func done(
	action string,
	s Entry,
) bool {
	active := s.State != "dead" && s.State != "failed"

	return map[string]bool{
		"start":   active,
		"stop":    !active,
		"enable":  s.Enabled,
		"disable": !s.Enabled,
	}[action]
}

// owned reports whether osapi deployed the unit file path: managedfile has a
// record of it, and has not dropped it. A unit file with the prefix that
// osapi did not write, such as the agent's own osapi-agent.service, is not
// osapi's to change.
func (d *debian) owned(
	ctx context.Context,
	path string,
) (bool, error) {
	rec, err := behavior.Call(ctx, managedfile.Behavior, managedfile.Get, d.files,
		managedfile.GetInput{Path: path})
	if errors.Is(err, managedfile.ErrNotManaged) {
		return false, nil
	}

	return err == nil && !rec.Undeployed, err
}

// systemctl runs systemctl with args, as root.
func (d *debian) systemctl(
	ctx context.Context,
	args ...string,
) error {
	if _, err := d.exec.RunPrivilegedCmd(ctx, "systemctl", args); err != nil {
		return fmt.Errorf("systemctl %s: %w", args[0], err)
	}

	return nil
}
