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
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/avfs/avfs"

	"github.com/osapi-io/osapi/internal/audit"
	"github.com/osapi-io/osapi/internal/config"
)

// Errors of get and rotate.
var (
	ErrNoCert       = errors.New("no certificate with this name: list shows each")
	ErrOperatorCert = errors.New("the certificate is the operator's cert_file: replace that file")
	ErrPinned       = errors.New("each join string pins the NATS server's certificate, and each " +
		"joined agent trusts it: --force rotates it, then make new join tokens and join each agent again")
)

// List is each certificate of osapi on this host, with its expiry.
func List(
	_ context.Context,
	d Deps,
	_ ListInput,
) (ListOutput, error) {
	out := ListOutput{Items: []CertInfo{}}
	for _, src := range d.Certs() {
		info, err := read(d, src)
		if err != nil {
			info = CertInfo{
				Name: src.Name, Component: src.Component, Source: source(src),
				File: src.CertFile, State: state(err),
			}
		}
		out.Items = append(out.Items, info)
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

	return read(d, src)
}

// Rotate makes a new self-signed certificate in place of src's.
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
	if src.Name == "nats" && !in.Force {
		return RotateOutput{}, ErrPinned
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
	next := "restart the server to serve it"
	if src.Name == "nats" {
		next += "; make new join tokens, and join each agent again"
	}

	return RotateOutput{Cert: info, Next: next}, nil
}

// find is the certificate named name.
func find(
	d Deps,
	name string,
) (Source, error) {
	for _, src := range d.Certs() {
		if src.Name == name {
			return src, nil
		}
	}

	return Source{}, fmt.Errorf("%s: %w", name, ErrNoCert)
}

// read is the certificate of src as list shows it.
func read(
	d Deps,
	src Source,
) (CertInfo, error) {
	data, err := avfs.ReadFile(d.FS, src.CertFile)
	if err != nil {
		return CertInfo{}, fmt.Errorf("read %s: %w", src.CertFile, err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return CertInfo{}, fmt.Errorf("%s: no PEM certificate", src.CertFile)
	}
	c, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return CertInfo{}, fmt.Errorf("%s: %w", src.CertFile, err)
	}
	names := append([]string{}, c.DNSNames...)
	for _, ip := range c.IPAddresses {
		names = append(names, ip.String())
	}
	sum := sha256.Sum256(block.Bytes)

	key, approved := keyAlgorithm(c)
	days := int(c.NotAfter.Sub(d.Now()).Hours() / 24)
	fips := approved && fipsSignature[c.SignatureAlgorithm]

	return CertInfo{
		Name: src.Name, State: "ok", Component: src.Component, Source: source(src),
		Key: key, Signature: c.SignatureAlgorithm.String(),
		FIPS: &fips, File: src.CertFile, Names: names,
		Expires: c.NotAfter, DaysLeft: &days,
		Fingerprint: "SHA256:" + hex.EncodeToString(sum[:]),
	}, nil
}

// fipsSignature are the signature algorithms that FIPS 186-5 approves: no
// MD5 and no SHA-1.
var fipsSignature = map[x509.SignatureAlgorithm]bool{
	x509.SHA256WithRSA: true, x509.SHA384WithRSA: true, x509.SHA512WithRSA: true,
	x509.SHA256WithRSAPSS: true, x509.SHA384WithRSAPSS: true, x509.SHA512WithRSAPSS: true,
	x509.ECDSAWithSHA256: true, x509.ECDSAWithSHA384: true, x509.ECDSAWithSHA512: true,
	x509.PureEd25519: true,
}

// keyAlgorithm is the algorithm and size of c's public key, and whether
// FIPS approves it: RSA of 2048 bits or more, ECDSA on P-256, P-384 or
// P-521, and Ed25519.
func keyAlgorithm(
	c *x509.Certificate,
) (string, bool) {
	switch k := c.PublicKey.(type) {
	case *rsa.PublicKey:
		return fmt.Sprintf("RSA %d", k.N.BitLen()), k.N.BitLen() >= 2048
	case *ecdsa.PublicKey:
		name := k.Curve.Params().Name
		return "ECDSA " + name, name == "P-256" || name == "P-384" || name == "P-521"
	case ed25519.PublicKey:
		return "Ed25519", true
	default:
		return c.PublicKeyAlgorithm.String(), false
	}
}

// source is where src comes from: self-signed, which osapi made, or the
// cert_file of the operator.
func source(
	src Source,
) string {
	if src.SelfSigned {
		return "self-signed"
	}

	return "cert_file"
}

// state is why list could not read a certificate: not made when the file is
// not there, as on a host where that component has not started, or
// unreadable with the reason, such as a permission that only root has.
func state(
	err error,
) string {
	if errors.Is(err, fs.ErrNotExist) {
		return "not made"
	}

	return "unreadable: " + err.Error()
}
