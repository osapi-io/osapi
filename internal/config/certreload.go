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
	"crypto/tls"
	"fmt"
	"os"
	"sync"
	"time"
)

// certReloader serves the certificate of a server from its files, and reads
// them again at a handshake after either file changed: a rotation renames
// a new certificate into place, and the server serves it at its next
// handshake, with no restart. A pair that does not load, such as a key that
// is written before its certificate, leaves the last one in place.
type certReloader struct {
	certFile string
	keyFile  string

	mu      sync.Mutex
	cert    *tls.Certificate
	certMod time.Time
	keyMod  time.Time
}

// newCertReloader loads the pair, which must load at the start.
func newCertReloader(
	certFile string,
	keyFile string,
) (*certReloader, error) {
	r := &certReloader{certFile: certFile, keyFile: keyFile}
	certMod, keyMod := r.mods()
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("load the server certificate: %w", err)
	}
	r.cert, r.certMod, r.keyMod = &cert, certMod, keyMod

	return r, nil
}

// mods are the modification times of the two files, or the zero time of a
// file that is not there.
func (r *certReloader) mods() (time.Time, time.Time) {
	var certMod, keyMod time.Time
	if fi, err := os.Stat(r.certFile); err == nil {
		certMod = fi.ModTime()
	}
	if fi, err := os.Stat(r.keyFile); err == nil {
		keyMod = fi.ModTime()
	}

	return certMod, keyMod
}

// certificate is the certificate to serve: the pair as it is now, when it
// changed and loads, else the last one.
func (r *certReloader) certificate(
	*tls.ClientHelloInfo,
) (*tls.Certificate, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	certMod, keyMod := r.mods()
	if certMod.Equal(r.certMod) && keyMod.Equal(r.keyMod) {
		return r.cert, nil
	}
	cert, err := tls.LoadX509KeyPair(r.certFile, r.keyFile)
	if err != nil {
		return r.cert, nil
	}
	r.cert, r.certMod, r.keyMod = &cert, certMod, keyMod

	return r.cert, nil
}
