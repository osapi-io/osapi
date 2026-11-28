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
	"io/fs"
	"strings"

	"github.com/avfs/avfs"
	gohaisvc "github.com/osapi-io/gohai/pkg/gohai/collectors/services"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/behaviors/managedfile"
	"github.com/osapi-io/osapi/internal/behaviors/object"
	"github.com/osapi-io/osapi/internal/initsys"
)

// labelPrefix starts the launchd label of each daemon that osapi deploys:
// io.osapi.app for app. The agent's own io.osapi.agent is not osapi's to
// change: managedfile has no record of its plist.
const labelPrefix = "io.osapi."

// darwin reads the daemons of launchd's system domain through gohai, gets each
// plist from the object store with object, writes it with managedfile, and
// runs launchctl. A daemon is in launchctl list only while it is loaded:
// bootstrap loads and starts it, and bootout stops and unloads it. enable and
// disable are launchd's lasting override, apart from loading, as systemctl's
// are apart from starting.
type darwin struct {
	*reader
	objects object.Provider
	files   managedfile.Provider
	exec    host.Exec
	fs      avfs.VFS
}

func newDarwin(
	h host.Host,
) (Provider, error) {
	return &darwin{
		reader:  &reader{src: h.Sources()},
		objects: host.Require[object.Provider](h, object.Behavior),
		files:   host.Require[managedfile.Provider](h, managedfile.Behavior),
		exec:    h.Exec,
		fs:      h.FS,
	}, nil
}

// plistPath is the plist of the daemon name.
func plistPath(
	name string,
) string {
	return initsys.LaunchdPath(labelPrefix + name)
}

// Services is each daemon of the system domain that launchd has loaded or
// that has an override, as gohai reads them. A daemon that osapi deployed
// reads by the name it was deployed with, app, not its label, io.osapi.app,
// and is Managed.
func (d *darwin) Services(
	ctx context.Context,
) ([]Entry, error) {
	list, err := d.reader.Services(ctx)
	for i, s := range list {
		name, ok := strings.CutPrefix(s.Name, labelPrefix)
		if !ok {
			continue
		}
		owned, oerr := d.owned(ctx, plistPath(name))
		if oerr != nil {
			return nil, oerr
		}
		if owned {
			list[i].Name, list[i].Managed = name, true
		}
	}

	return list, err
}

// Service is one daemon, by the name that its writes take, or by its label.
// A daemon that osapi deployed and launchd has not loaded is dead.
func (d *darwin) Service(
	ctx context.Context,
	name string,
) (Entry, error) {
	bare := strings.TrimPrefix(name, labelPrefix)
	owned, err := d.owned(ctx, plistPath(bare))
	if err != nil {
		return Entry{}, err
	}
	if !owned {
		return d.reader.Service(ctx, name)
	}
	s, err := d.reader.Service(ctx, labelPrefix+bare)
	if err != nil {
		// gohai has each daemon that launchd loaded or that has an
		// override, and reads launchctl that fails as no daemons. One
		// with neither is not loaded, and launchd enables a daemon unless
		// an override disables it.
		s = Entry{State: "dead", Enabled: true}
	}
	s.Name, s.Managed = bare, true

	return s, nil
}

// Deploy writes the object of u as the plist of the daemon. It does not load
// it: start and enable are their own actions. An update of a daemon that osapi
// did not deploy is not found.
func (d *darwin) Deploy(
	ctx context.Context,
	u Unit,
	create bool,
) (bool, error) {
	if err := check(u.Name); err != nil {
		return false, err
	}
	path := plistPath(u.Name)
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
			Path:         path,
			Content:      obj.Content,
			Mode:         "0644",
			Metadata:     map[string]string{"object": u.Object},
			Object:       u.Object,
			ObjectSHA256: obj.SHA256,
		})
	if err != nil {
		return false, err
	}

	return result.Changed, nil
}

// Remove stops and unloads the daemon, and drops its plist. A daemon with no
// plist is unchanged, and a record of a plist that is not on the host is
// dropped.
func (d *darwin) Remove(
	ctx context.Context,
	name string,
) (bool, error) {
	if err := check(name); err != nil {
		return false, err
	}
	path := plistPath(name)
	if owned, err := d.owned(ctx, path); err != nil || !owned {
		return false, err
	}
	if _, err := d.fs.Stat(path); errors.Is(err, fs.ErrNotExist) {
		_, err := behavior.Call(ctx, managedfile.Behavior, managedfile.Delete, d.files,
			managedfile.DeleteInput{Path: path})

		return false, err
	}
	label := labelPrefix + name
	if d.loaded(ctx, label) {
		if err := d.run(ctx, initsys.LaunchdUnload(label)[0]); err != nil {
			return false, err
		}
	}
	if _, err := behavior.Call(ctx, managedfile.Behavior, managedfile.Delete, d.files,
		managedfile.DeleteInput{Path: path}); err != nil {
		return false, err
	}

	return true, nil
}

// Act runs action on the daemon, unless it is already where the action puts
// it: start bootstraps an unloaded daemon and kickstarts it, stop
// boots it out, restart kickstarts it with -k, and enable and disable set
// launchd's override. restart always changes. launchd does not bootstrap a
// disabled daemon, so a start of one that is not loaded is refused: enable
// it first. disable of a daemon that runs does not stop it, as on systemd.
func (d *darwin) Act(
	ctx context.Context,
	action string,
	name string,
) (bool, error) {
	if err := check(name); err != nil {
		return false, err
	}
	if owned, err := d.owned(ctx, plistPath(name)); err != nil || !owned {
		if err == nil {
			err = fmt.Errorf("service %q: %w", name, behavior.ErrNotFound)
		}

		return false, err
	}
	s, err := d.Service(ctx, name)
	if err != nil {
		return false, err
	}
	if done(action, s) {
		return false, nil
	}
	label := labelPrefix + name
	loaded := d.loaded(ctx, label)
	if !loaded && !s.Enabled && (action == "start" || action == "restart") {
		return false, fmt.Errorf(
			"service %q is disabled, and launchd does not load a disabled daemon: enable it first",
			name)
	}
	// bootstrap loads a daemon, and starts it only when its plist has
	// RunAtLoad: kickstart starts it.
	kickstart := []string{"launchctl", "kickstart", "system/" + label}
	var argvs [][]string
	switch action {
	case "start", "restart":
		if !loaded {
			argvs = append(initsys.LaunchdLoad(label), kickstart)
		} else if action == "start" {
			argvs = [][]string{kickstart}
		} else {
			argvs = [][]string{{"launchctl", "kickstart", "-k", "system/" + label}}
		}
	case "stop":
		argvs = initsys.LaunchdUnload(label)
	default:
		argvs = [][]string{{"launchctl", action, "system/" + label}}
	}
	for _, argv := range argvs {
		if err := d.run(ctx, argv); err != nil {
			return false, err
		}
	}
	after, _ := d.Service(ctx, name)

	return action == "restart" || done(action, after), nil
}

// loaded is whether launchd has the daemon label loaded, as gohai reads it:
// a label of the services block of launchctl print system is loaded, and a
// label only of the disabled block is not. A state of dead cannot say, because
// a daemon that is loaded and does not run is dead too. gohai reads launchctl
// that fails as no daemons, so a read that fails is not loaded.
func (d *darwin) loaded(
	ctx context.Context,
	label string,
) bool {
	out, _ := gohaisvc.NewWith(d.reader.src).Collect(ctx, nil)
	if info, ok := out.(*gohaisvc.Info); ok {
		for _, s := range info.Services {
			if s.Name == label {
				return s.Loaded
			}
		}
	}

	return false
}

// owned reports whether osapi deployed the plist path: managedfile has a
// record of it, and has not dropped it.
func (d *darwin) owned(
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

// run runs argv, a launchctl command, with sudo.
func (d *darwin) run(
	ctx context.Context,
	argv []string,
) error {
	if _, err := d.exec.RunPrivilegedCmd(ctx, argv[0], argv[1:]); err != nil {
		return fmt.Errorf("%s %s: %w", argv[0], argv[1], err)
	}

	return nil
}
