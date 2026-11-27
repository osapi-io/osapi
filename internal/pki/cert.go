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

package pki

import (
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
	"time"

	"github.com/avfs/avfs"
)

// CertSource is a certificate that a server of osapi serves: its name, the
// component that serves it, its files, and whether osapi made it.
type CertSource struct {
	Name       string
	Component  string
	CertFile   string
	KeyFile    string
	SelfSigned bool
}

// Cert is what a certificate of a host is: osapi pki cert reads it on the
// host, and the controller and each agent report theirs through the API.
type Cert struct {
	Name        string    `json:"name"`
	State       string    `json:"state"`
	Component   string    `json:"component"`
	Source      string    `json:"source"`
	File        string    `json:"file"`
	Names       []string  `json:"names,omitempty"`
	Subject     string    `json:"subject,omitempty"`
	Issuer      string    `json:"issuer,omitempty"`
	NotBefore   time.Time `json:"not_before,omitzero"`
	Expires     time.Time `json:"expires,omitzero"`
	DaysLeft    *int      `json:"days_left,omitempty"`
	Fingerprint string    `json:"fingerprint,omitempty"`
	Key         string    `json:"key,omitempty"`
	Signature   string    `json:"signature,omitempty"`
	FIPS        *bool     `json:"fips_approved,omitempty"`
}

// The states of a certificate.
const (
	// CertOK is a certificate that osapi read.
	CertOK = "ok"
	// CertNotMade is a certificate whose file is not there, as on a host
	// where its component has not started.
	CertNotMade = "not made"
)

// ReadCert reads the certificate of src from fsys, with the days that are
// left of it at now.
func ReadCert(
	fsys avfs.VFS,
	src CertSource,
	now time.Time,
) (Cert, error) {
	data, err := avfs.ReadFile(fsys, src.CertFile)
	if err != nil {
		return Cert{}, fmt.Errorf("read %s: %w", src.CertFile, err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return Cert{}, fmt.Errorf("%s: no PEM certificate", src.CertFile)
	}
	c, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return Cert{}, fmt.Errorf("%s: %w", src.CertFile, err)
	}
	names := append([]string{}, c.DNSNames...)
	for _, ip := range c.IPAddresses {
		names = append(names, ip.String())
	}
	sum := sha256.Sum256(block.Bytes)
	key, approved := keyAlgorithm(c)
	days := int(c.NotAfter.Sub(now).Hours() / 24)
	fips := approved && fipsSignature[c.SignatureAlgorithm]

	return Cert{
		Name: src.Name, State: CertOK, Component: src.Component, Source: certSource(src),
		Key: key, Signature: c.SignatureAlgorithm.String(),
		FIPS: &fips, File: src.CertFile, Names: names,
		Subject: c.Subject.String(), Issuer: c.Issuer.String(),
		NotBefore: c.NotBefore, Expires: c.NotAfter, DaysLeft: &days,
		Fingerprint: "SHA256:" + hex.EncodeToString(sum[:]),
	}, nil
}

// ReadCerts reads each certificate of srcs. One that cannot be read is in
// the list too, with its state: not made, or unreadable with the reason.
func ReadCerts(
	fsys avfs.VFS,
	srcs []CertSource,
	now time.Time,
) []Cert {
	out := make([]Cert, 0, len(srcs))
	for _, src := range srcs {
		c, err := ReadCert(fsys, src, now)
		if err != nil {
			c = Cert{
				Name: src.Name, Component: src.Component, Source: certSource(src),
				File: src.CertFile, State: CertState(err),
			}
		}
		out = append(out, c)
	}

	return out
}

// CertState is why a certificate could not be read: not made when the file
// is not there, or unreadable with the reason, such as a permission that
// only root has.
func CertState(
	err error,
) string {
	if errors.Is(err, fs.ErrNotExist) {
		return CertNotMade
	}

	return "unreadable: " + err.Error()
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

// certSource is where src comes from: self-signed, which osapi made, or
// the cert_file of the operator.
func certSource(
	src CertSource,
) string {
	if src.SelfSigned {
		return "self-signed"
	}

	return "cert_file"
}
