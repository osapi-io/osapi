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

package schedule

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
	"github.com/osapi-io/osapi/internal/validation"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// cronDir holds the entries with a cron expression.
const cronDir = "/etc/cron.d"

// intervals are the directories of the scripts that run hourly, daily,
// weekly or monthly, in the order that a lookup tries them.
var intervals = types.Interval("").Values()

// check refuses a name that a program would read as an option, such as
// --now: cron's own tools and a person's shell treat a file of that name as
// one.
func check(
	name string,
) error {
	if strings.HasPrefix(name, "-") {
		return fmt.Errorf("%q is not a valid entry name", name)
	}

	return nil
}

// intervalDir is the directory of an interval's scripts.
func intervalDir(
	interval types.Interval,
) string {
	return "/etc/cron." + string(interval)
}

// debian deploys each entry as a file that managedfile keeps a record of:
// the entry with a cron expression in /etc/cron.d, mode 0644, and a script
// of an interval in /etc/cron.<interval>, mode 0755. Only a file with a
// record is osapi's: a list, a get, an update or a delete leaves every
// other file alone.
type debian struct {
	objects  object.Provider
	files    managedfile.Provider
	fs       avfs.VFS
	hostname string
	facts    func() map[string]any
}

// newDebian builds the provider of the Debian family.
func newDebian(
	h host.Host,
) (Provider, error) {
	return &debian{
		objects:  host.Require[object.Provider](h, object.Behavior),
		files:    host.Require[managedfile.Provider](h, managedfile.Behavior),
		fs:       h.FS,
		hostname: h.Hostname,
		facts:    h.Facts,
	}, nil
}

// List reads each file in /etc/cron.d and the interval directories that
// osapi deployed.
func (d *debian) List(
	ctx context.Context,
) ([]Entry, error) {
	out := []Entry{}
	for _, dir := range append([]string{cronDir}, intervalDirs()...) {
		entries, err := d.fs.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			entry, ok, err := d.entry(ctx, e.Name(), dir+"/"+e.Name())
			if err != nil {
				return nil, err
			}
			if ok {
				out = append(out, entry)
			}
		}
	}

	return out, nil
}

// Get reads one entry that osapi deployed. An entry that osapi did not
// deploy is not found.
func (d *debian) Get(
	ctx context.Context,
	name string,
) (Entry, error) {
	path, err := d.find(ctx, name)
	if err != nil {
		return Entry{}, err
	}
	entry, _, err := d.entry(ctx, name, path)

	return entry, err
}

// Create deploys the object of in as the entry's file. An entry of the name
// in a different directory is left as it is, unchanged: moving it is not a
// create. A file at the path that osapi did not deploy is refused.
func (d *debian) Create(
	ctx context.Context,
	in CreateInput,
) (bool, error) {
	if err := check(in.Name); err != nil {
		return false, err
	}
	path, mode := cronDir+"/"+in.Name, "0644"
	if in.Interval != "" {
		path, mode = intervalDir(in.Interval)+"/"+in.Name, "0755"
	}
	if existing, err := d.find(ctx, in.Name); err == nil && existing != path {
		return false, nil
	}
	owned, err := d.owned(ctx, path)
	if err != nil {
		return false, err
	}
	if _, statErr := d.fs.Stat(path); statErr == nil && !owned {
		return false, fmt.Errorf("%s is on the host and osapi did not deploy it", path)
	}

	return d.deploy(
		ctx,
		path,
		mode,
		in.Object,
		types.ContentType(in.ContentType),
		in.Vars,
		map[string]string{
			"schedule": in.Schedule, "interval": string(in.Interval), "user": in.User,
		},
	)
}

// Update deploys an entry that osapi deployed again, with the fields that
// in gives and the entry's own for the rest. An entry that osapi did not
// deploy is not found.
func (d *debian) Update(
	ctx context.Context,
	in UpdateInput,
) (bool, error) {
	path, err := d.find(ctx, in.Name)
	if err != nil {
		return false, err
	}
	rec, err := behavior.Call(ctx, managedfile.Behavior, managedfile.Get, d.files,
		managedfile.GetInput{Path: path})
	if err != nil {
		return false, err
	}
	meta := map[string]string{
		"schedule": rec.Metadata["schedule"],
		"interval": rec.Metadata["interval"],
		"user":     rec.Metadata["user"],
	}
	obj := or(in.Object, rec.Metadata["object"])
	if in.Schedule != "" {
		meta["schedule"] = in.Schedule
	}
	mode := "0644"
	if !strings.HasPrefix(path, cronDir+"/") {
		mode = "0755"
	}
	if in.User != "" {
		if mode != "0644" {
			return false, fmt.Errorf(
				"%s runs as root: a user is for an entry in %s: %w",
				path,
				cronDir,
				behavior.ErrInvalid,
			)
		}
		meta["user"] = in.User
	}

	return d.deploy(ctx, path, mode, obj, types.ContentType(in.ContentType), in.Vars, meta)
}

// Delete removes an entry that osapi deployed. An entry with no file that
// osapi deployed is unchanged.
func (d *debian) Delete(
	ctx context.Context,
	name string,
) (bool, error) {
	if err := check(name); err != nil {
		return false, err
	}
	path, err := d.find(ctx, name)
	if errors.Is(err, behavior.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	result, err := behavior.Call(ctx, managedfile.Behavior, managedfile.Delete, d.files,
		managedfile.DeleteInput{Path: path})

	return result.Changed, err
}

// deploy writes the object, rendered when it is a template, at path, with
// meta and the object in the record.
func (d *debian) deploy(
	ctx context.Context,
	path string,
	mode string,
	objectName string,
	contentType types.ContentType,
	vars map[string]string,
	meta map[string]string,
) (bool, error) {
	obj, err := behavior.Call(ctx, object.Behavior, object.Get, d.objects,
		object.GetInput{Name: objectName})
	if err != nil {
		return false, err
	}
	content := obj.Content
	kind := or(or(contentType, obj.ContentType), types.ContentRaw)
	if kind == types.ContentTemplate {
		data := behavior.TemplateData{Vars: vars, Hostname: d.hostname}
		if d.facts != nil {
			data.Facts = d.facts()
		}
		if content, err = behavior.Render(content, data); err != nil {
			return false, err
		}
	}
	if strings.HasPrefix(path, cronDir+"/") {
		// The object is the command, and the line is what cron runs: the
		// cron expression and the user of the record, so a read gives what
		// cron runs. An entry with no user runs as root.
		meta["user"] = or(meta["user"], "root")
		if content, err = cronLine(ctx, meta["schedule"], meta["user"], content); err != nil {
			return false, err
		}
	}
	record := map[string]string{"object": objectName}
	for k, v := range meta {
		if v != "" {
			record[k] = v
		}
	}
	result, err := behavior.Call(ctx, managedfile.Behavior, managedfile.Update, d.files,
		managedfile.UpdateInput{
			Path:         path,
			Content:      content,
			Mode:         mode,
			Metadata:     record,
			Object:       objectName,
			ContentType:  kind,
			ObjectSHA256: obj.SHA256,
		})

	return result.Changed, err
}

// cronLine is the line of an entry in /etc/cron.d that runs command, which
// is one line, at schedule as user. The schedule and the user become fields
// of the line that cron runs, and an update can take them from the record,
// so they follow the rules of the input again: cron_schedule and
// account_name.
func cronLine(
	ctx context.Context,
	schedule string,
	user string,
	command string,
) (string, error) {
	if msg, ok := validation.Var(ctx, schedule, "required,cron_schedule"); !ok {
		return "", fmt.Errorf("cron expression %q: %s: %w", schedule, msg, behavior.ErrInvalid)
	}
	if msg, ok := validation.Var(ctx, user, "required,account_name"); !ok {
		return "", fmt.Errorf("user %q: %s: %w", user, msg, behavior.ErrInvalid)
	}
	command = strings.TrimSpace(command)
	if command == "" || strings.Contains(command, "\n") {
		return "", fmt.Errorf(
			"the object of an entry in %s is one command line: %w",
			cronDir,
			behavior.ErrInvalid,
		)
	}

	return schedule + " " + user + " " + command + "\n", nil
}

// find is the path of the entry that osapi deployed under name: in
// /etc/cron.d, then in each interval directory. An entry that osapi did not
// deploy is not found.
func (d *debian) find(
	ctx context.Context,
	name string,
) (string, error) {
	if err := check(name); err != nil {
		return "", err
	}
	for _, dir := range append([]string{cronDir}, intervalDirs()...) {
		path := dir + "/" + name
		owned, err := d.owned(ctx, path)
		if err != nil {
			return "", err
		}
		if owned {
			return path, nil
		}
	}

	return "", fmt.Errorf("schedule %q: %w", name, behavior.ErrNotFound)
}

// entry is the entry of the file at path, and whether osapi deployed it.
func (d *debian) entry(
	ctx context.Context,
	name string,
	path string,
) (Entry, bool, error) {
	rec, err := behavior.Call(ctx, managedfile.Behavior, managedfile.Get, d.files,
		managedfile.GetInput{Path: path})
	if errors.Is(err, managedfile.ErrNotManaged) {
		return Entry{}, false, nil
	}
	if err != nil {
		return Entry{}, false, err
	}
	if rec.Undeployed {
		return Entry{}, false, nil
	}
	source := "cron.d"
	if !strings.HasPrefix(path, cronDir+"/") {
		source = rec.Metadata["interval"]
	}

	return Entry{
		Name:     name,
		Object:   rec.Metadata["object"],
		Schedule: rec.Metadata["schedule"],
		Interval: rec.Metadata["interval"],
		User:     rec.Metadata["user"],
		Source:   source,
	}, true, nil
}

// owned reports whether osapi deployed path: managedfile has a record of
// it, and has not dropped it.
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

// intervalDirs are the directories of the intervals.
func intervalDirs() []string {
	out := make([]string, 0, len(intervals))
	for _, i := range intervals {
		out = append(out, intervalDir(i))
	}

	return out
}

// or is a, or b when a is empty.
func or[T ~string](
	a T,
	b T,
) T {
	if a != "" {
		return a
	}

	return b
}
