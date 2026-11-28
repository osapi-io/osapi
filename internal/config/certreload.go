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

package config

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/osfs"
)

// hostFS is the filesystem of the host.
var hostFS avfs.VFS = osfs.NewWithNoIdm()

// certReloader serves the certificate of a server from its files, and reads
// them again, at most once each CertCheckInterval, at a handshake: a rotation
// renames a new certificate into place, and the server serves it at a
// handshake after it, with no restart. It compares the content of the
// files, not their times, so a copy that keeps the time, such as rsync -t
// or cp -p, is read too. A pair that does not load, such as a key that is
// written before its certificate, leaves the last one in place.
type certReloader struct {
	certFile string
	keyFile  string

	mu      sync.Mutex
	cert    *tls.Certificate
	certPEM []byte
	keyPEM  []byte
	checked time.Time
}

// CertCheckInterval is the most often that a handshake reads the files of
// its certificate again. A test sets it to 0.
var CertCheckInterval = time.Second

// newCertReloader loads the pair, which must load at the start.
func newCertReloader(
	certFile string,
	keyFile string,
) (*certReloader, error) {
	r := &certReloader{certFile: certFile, keyFile: keyFile}
	certPEM, keyPEM := r.read()
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("load the server certificate: %w", err)
	}
	r.cert, r.certPEM, r.keyPEM, r.checked = &cert, certPEM, keyPEM, time.Now()

	return r, nil
}

// read is the content of the two files, or nil for a file that does not
// read.
func (r *certReloader) read() ([]byte, []byte) {
	certPEM, _ := avfs.ReadFile(hostFS, r.certFile)
	keyPEM, _ := avfs.ReadFile(hostFS, r.keyFile)

	return certPEM, keyPEM
}

// certificate is the certificate to serve: the pair as it is now, when it
// changed and loads, else the last one.
func (r *certReloader) certificate(
	*tls.ClientHelloInfo,
) (*tls.Certificate, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	if now.Sub(r.checked) < CertCheckInterval {
		return r.cert, nil
	}
	r.checked = now
	certPEM, keyPEM := r.read()
	if bytes.Equal(certPEM, r.certPEM) && bytes.Equal(keyPEM, r.keyPEM) {
		return r.cert, nil
	}
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return r.cert, nil
	}
	r.cert, r.certPEM, r.keyPEM = &cert, certPEM, keyPEM

	return r.cert, nil
}
