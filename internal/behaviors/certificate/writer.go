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

package certificate

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/avfs/avfs"
	"github.com/osapi-io/gohai/pkg/gohai/collectors/cacertificates"
	"github.com/osapi-io/gohai/pkg/gohai/family"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/behaviors/managedfile"
	"github.com/osapi-io/osapi/internal/behaviors/object"
)

// prefix starts the file name of each CA that osapi deploys. A CA of the
// host has no file with the prefix that managedfile has a record of, so
// osapi writes and removes its own CAs only.
const prefix = "osapi-"

// writer reads through gohai, gets the PEM of a CA from the object store
// with object, writes it with managedfile into the family's local directory,
// the one gohai reads, and rebuilds the trust store.
type writer struct {
	*reader
	objects object.Provider
	files   managedfile.Provider
	exec    host.Exec
	fs      avfs.VFS
	// dir is the local directory of the family, and update the program that
	// rebuilds its trust store from it.
	dir    string
	update string
}

// newWriter is the provider on Linux. The directory is the local one of
// gohai's layout for the family, so a CA that osapi writes is one that gohai
// reads. The Debian family and SUSE rebuild their trust stores with
// update-ca-certificates, each its own, and the others with update-ca-trust,
// as the definition declares. osapi does not support Alpine, whose
// update-ca-certificates the linux provider does not declare, so it reads
// only there. An agent with no file state store has no managedfile, and
// reads only.
func newWriter(
	h host.Host,
) (Provider, error) {
	files := host.Require[managedfile.Provider](h, managedfile.Behavior)
	if files == nil || string(h.Platform) == family.Alpine {
		return newReader(h)
	}
	dir := ""
	for _, d := range cacertificates.LayoutFor(string(h.Platform)).Dirs {
		if d.Source == cacertificates.SourceLocal {
			dir = d.Path
		}
	}
	update := "update-ca-trust"
	switch h.Platform {
	case behavior.PlatformDebian, behavior.PlatformSUSE:
		update = "update-ca-certificates"
	}

	return &writer{
		reader:  &reader{src: h.Sources()},
		objects: host.Require[object.Provider](h, object.Behavior),
		files:   files,
		exec:    h.Exec,
		fs:      h.FS,
		dir:     dir,
		update:  update,
	}, nil
}

// check refuses a name that starts with a dash, which a person reads as an
// option.
func check(
	name string,
) error {
	if strings.HasPrefix(name, "-") {
		return fmt.Errorf("%q is not a valid name for a CA", name)
	}

	return nil
}

// path is the file of the CA name that osapi deploys.
func (w *writer) path(
	name string,
) string {
	return w.dir + "/" + prefix + name + ".crt"
}

// List is each CA of the host. A CA that osapi deployed reads by the name
// that it was deployed with, as its writes take it, not by the file's
// osapi- name, with Managed and the object it came from.
func (w *writer) List(
	ctx context.Context,
) ([]Entry, error) {
	list, err := w.reader.List(ctx)
	for i, e := range list {
		name, ok := strings.CutPrefix(e.Name, prefix)
		if !ok || e.Source != cacertificates.SourceLocal {
			continue
		}
		rec, owned, oerr := w.owned(ctx, w.path(name))
		if oerr != nil {
			return nil, oerr
		}
		if owned {
			list[i].Name, list[i].Managed, list[i].Object = name, true, rec.Metadata["object"]
		}
	}

	return list, err
}

// Get is one CA, by the name that list gives it.
func (w *writer) Get(
	ctx context.Context,
	name string,
) (Entry, error) {
	list, err := w.List(ctx)

	return find(list, name, err)
}

// Deploy writes the PEM of object as the CA name, and rebuilds the trust
// store when the file changed. With create, a CA that osapi deployed is
// written again, and changes nothing when its content is the same; without
// it, a CA that osapi did not deploy is not found, and an empty object keeps
// the one that it was deployed from.
func (w *writer) Deploy(
	ctx context.Context,
	name string,
	obj string,
	create bool,
) (bool, error) {
	if err := check(name); err != nil {
		return false, err
	}
	path := w.path(name)
	rec, owned, err := w.owned(ctx, path)
	if err != nil {
		return false, err
	}
	if !owned && !create {
		return false, fmt.Errorf("certificate %q: %w", name, behavior.ErrNotFound)
	}
	if _, statErr := w.fs.Stat(path); statErr == nil && !owned {
		return false, fmt.Errorf("%s is on the host and osapi did not deploy it", path)
	}
	if obj == "" {
		obj = rec.Metadata["object"]
	}
	got, err := behavior.Call(ctx, object.Behavior, object.Get, w.objects,
		object.GetInput{Name: obj})
	if err != nil {
		return false, err
	}
	result, err := behavior.Call(ctx, managedfile.Behavior, managedfile.Update, w.files,
		managedfile.UpdateInput{
			Path:         path,
			Content:      got.Content,
			Mode:         "0644",
			Metadata:     map[string]string{"object": obj},
			Object:       obj,
			ObjectSHA256: got.SHA256,
		})
	if err != nil || !result.Changed {
		return false, err
	}

	return true, w.rebuild(ctx, nil)
}

// Remove drops the file of a CA that osapi deployed, and rebuilds the trust
// store. A CA that osapi did not deploy is unchanged.
func (w *writer) Remove(
	ctx context.Context,
	name string,
) (bool, error) {
	if err := check(name); err != nil {
		return false, err
	}
	path := w.path(name)
	if _, owned, err := w.owned(ctx, path); err != nil || !owned {
		return false, err
	}
	if _, err := behavior.Call(ctx, managedfile.Behavior, managedfile.Delete, w.files,
		managedfile.DeleteInput{Path: path}); err != nil {
		return false, err
	}
	// update-ca-certificates of Debian and SUSE without --fresh keeps the
	// links in /etc/ssl/certs of a local CA that is gone: Debian cleans them
	// only when a CA was added or a system one turned off.
	var args []string
	if w.update == "update-ca-certificates" {
		args = []string{"--fresh"}
	}

	return true, w.rebuild(ctx, args)
}

// owned is the record of path, and whether osapi deployed it: managedfile
// has a record of it, and has not dropped it.
func (w *writer) owned(
	ctx context.Context,
	path string,
) (managedfile.Record, bool, error) {
	rec, err := behavior.Call(ctx, managedfile.Behavior, managedfile.Get, w.files,
		managedfile.GetInput{Path: path})
	if errors.Is(err, managedfile.ErrNotManaged) {
		return managedfile.Record{}, false, nil
	}

	return rec, err == nil && !rec.Undeployed, err
}

// rebuild runs the family's program that rebuilds the trust store, as root,
// with args.
func (w *writer) rebuild(
	ctx context.Context,
	args []string,
) error {
	if _, err := w.exec.RunPrivilegedCmd(ctx, w.update, args); err != nil {
		return fmt.Errorf("%s: %w", w.update, err)
	}

	return nil
}
