// Copyright (c) 2024 John Dewey

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

package api

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"path/filepath"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/osfs"

	"github.com/osapi-io/osapi/internal/config"
)

// hostFS is the filesystem of the host.
var hostFS avfs.VFS = osfs.NewWithNoIdm()

// The files of the self-signed certificate in controller.pki.key_dir.
const (
	SelfSignedCert = "api.crt"
	SelfSignedKey  = "api.key"
)

// TLSConfig is the server's TLS config: the certificate of c, or the
// self-signed certificate in keyDir, which it makes when it does not exist.
// A certificate that c names later replaces the self-signed one with no
// other change. The server reads the files again when they change, at the
// next handshake: a rotated certificate takes effect with no restart.
func TLSConfig(
	c config.ServerTLS,
	keyDir string,
	urls ...string,
) (*tls.Config, error) {
	c.Enabled = true
	c, err := c.WithSelfSigned(keyDir, "api", urls...)
	if err != nil {
		return nil, err
	}
	out, err := c.Config()
	if err != nil {
		return nil, fmt.Errorf("load the API's certificate: %w", err)
	}

	return out, nil
}

// CertFingerprint is the SHA-256 of the certificate that the API serves, as
// SHA256:<hex>, which a join token pins: the certificate of c, or the
// self-signed one in keyDir.
func CertFingerprint(
	c config.ServerTLS,
	keyDir string,
) (string, error) {
	certFile := c.CertFile
	if certFile == "" {
		certFile = filepath.Join(keyDir, SelfSignedCert)
	}
	data, err := avfs.ReadFile(hostFS, certFile)
	if err != nil {
		return "", fmt.Errorf("read the API's certificate: %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return "", fmt.Errorf("%s: no PEM certificate", certFile)
	}
	sum := sha256.Sum256(block.Bytes)

	return "SHA256:" + hex.EncodeToString(sum[:]), nil
}
