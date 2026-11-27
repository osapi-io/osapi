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

package pkicert

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/osapi-io/osapi/internal/audit"
	"github.com/osapi-io/osapi/internal/bus"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller/certrotation"
	"github.com/osapi-io/osapi/internal/pki"
)

// Errors of get and rotate.
var (
	ErrNoCert       = errors.New("no certificate with this name: list shows each")
	ErrNotMade      = errors.New("not made")
	ErrOperatorCert = certrotation.ErrOperatorCert
)

// List is each certificate of osapi on this host, with its expiry.
func List(
	_ context.Context,
	d Deps,
	_ ListInput,
) (ListOutput, error) {
	out := ListOutput{Items: []CertInfo{}}
	for _, c := range pki.ReadCerts(d.FS, d.Certs(), d.Now()) {
		out.Items = append(out.Items, CertInfo(c))
	}

	return out, nil
}

// Get is one certificate in full.
func Get(
	_ context.Context,
	d Deps,
	in GetInput,
) (CertInfo, error) {
	src, err := find(d, in.Name)
	if err != nil {
		return CertInfo{}, err
	}
	info, err := read(d, src)
	if errors.Is(err, fs.ErrNotExist) {
		return CertInfo{}, fmt.Errorf("%s is not on this host: the %s makes it at its first "+
			"start; run this on the host of the %s, or read it through the API with "+
			"osapi client pki cert get --name %s: %w",
			src.Name, src.Component, src.Component, src.Name, ErrNotMade)
	}

	return info, err
}

// Rotate makes a new self-signed certificate in place of src's, which its
// server serves at its next handshake, with no restart. The NATS server's
// certificate, which each joined agent pins, rotates with an overlap: the
// controller gives the agents the next certificate, and switches to it when
// they trust it.
func Rotate(
	ctx context.Context,
	d Deps,
	in RotateInput,
) (RotateOutput, error) {
	src, err := find(d, in.Name)
	if err != nil {
		return RotateOutput{}, err
	}
	if !src.SelfSigned {
		return RotateOutput{}, ErrOperatorCert
	}
	if src.Name == "nats" {
		return rotateNATS(ctx, d, src)
	}
	for _, f := range []string{src.CertFile, src.KeyFile} {
		if err := d.FS.Remove(f); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return RotateOutput{}, fmt.Errorf("remove %s: %w", f, err)
		}
	}
	dir, name := filepath.Dir(src.CertFile), strings.TrimSuffix(filepath.Base(src.CertFile), ".crt")
	if _, err := (config.ServerTLS{Enabled: true}).WithSelfSigned(dir, name); err != nil {
		return RotateOutput{}, err
	}
	info, err := read(d, src)
	if err != nil {
		return RotateOutput{}, err
	}
	audit.RecordLocal(ctx, d.Audit, "cert.rotate", src.Name, "fingerprint="+info.Fingerprint)
	next := "the server serves it at its next handshake"
	if src.Name == "api" {
		next += "; join tokens made from now carry its fingerprint, and a client on " +
			"another host that trusts the old one needs the new one as its ca_file"
	}

	return RotateOutput{Cert: info, Next: next}, nil
}

// rotateNATS starts the rotation of the NATS server's certificate, on the
// controller's host: it makes the next certificate beside the current one,
// and writes the record that gives both to the agents, over NATS as the
// controller connects. The running controller switches to it when each
// agent trusts it.
func rotateNATS(
	ctx context.Context,
	d Deps,
	src pki.CertSource,
) (RotateOutput, error) {
	keys := pki.New(d.FS, d.Config.Controller.PKI.KeyDir, "controller")
	if err := keys.Load(); err != nil {
		return RotateOutput{}, fmt.Errorf("the controller's key signs the rotation of the NATS "+
			"server's certificate: run this on the controller's host: %w", err)
	}
	info, err := read(d, src)
	if err != nil {
		return RotateOutput{}, err
	}
	kv, closeKV, err := d.Dial(ctx)
	if err != nil {
		return RotateOutput{}, err
	}
	defer closeKV()
	bucket := bus.StateKV(d.Config.Controller.NATS.Namespace, d.Config.NATS.State).Bucket
	status, err := certrotation.New(certrotation.Config{
		KV: kv, Bucket: bucket, Keys: keys, Source: src,
	}).Start(ctx)
	if err != nil {
		return RotateOutput{}, err
	}
	audit.RecordLocal(ctx, d.Audit, "cert.rotate", src.Name,
		"fingerprint="+info.Fingerprint+" next="+status.Next)

	return RotateOutput{
		Cert:     info,
		NextCert: status.Next,
		Next:     "each agent trusts the next certificate at its next heartbeat, and the controller switches to it when each agent does, within controller.pki.rotation_grace_period; pki cert get shows the agents that it waits for",
	}, nil
}

// find is the certificate named name.
func find(
	d Deps,
	name string,
) (pki.CertSource, error) {
	for _, src := range d.Certs() {
		if src.Name == name {
			return src, nil
		}
	}

	return pki.CertSource{}, fmt.Errorf("%s: %w", name, ErrNoCert)
}

// read is the certificate of src as list shows it.
func read(
	d Deps,
	src pki.CertSource,
) (CertInfo, error) {
	c, err := pki.ReadCert(d.FS, src, d.Now())
	if err != nil {
		return CertInfo{}, err
	}

	return CertInfo(c), nil
}
