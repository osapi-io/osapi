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

package config

import (
	"crypto/x509"
	"os"
	"reflect"

	"github.com/go-playground/validator/v10"
)

// SetRegisterValidatorsFn overrides the registerValidatorsFn variable for testing.
func SetRegisterValidatorsFn(
	fn func(*validator.Validate) error,
) {
	registerValidatorsFn = fn
}

// ResetRegisterValidatorsFn restores the registerValidatorsFn variable to its default.
func ResetRegisterValidatorsFn() {
	registerValidatorsFn = registerValidators
}

// KeysOf walks t the way Keys walks Config, for testing fields Config does
// not have.
func KeysOf(
	t reflect.Type,
) []Key {
	return walk(t, "", reflect.Value{})
}

// ExportSharedObjectStoreBucket runs the shared-bucket check on infos.
func ExportSharedObjectStoreBucket(
	infos []ObjectStoreBucketInfo,
) error {
	keys, buckets := make([]string, 0, len(infos)), make([]string, 0, len(infos))
	for _, b := range infos {
		keys, buckets = append(keys, b.Key), append(buckets, b.Bucket)
	}

	return sharedName(keys, buckets)
}

// KeyOf exposes keyOf for testing.
var KeyOf = keyOf

// ValueProblems exposes valueProblems for testing.
var ValueProblems = valueProblems

// SetRandRead replaces how the signing key's random bytes are read.
func SetRandRead(
	fn func([]byte) (int, error),
) {
	randRead = fn
}

// SetSystemCertPoolFn overrides the systemCertPoolFn variable for testing.
func SetSystemCertPoolFn(
	fn func() (*x509.CertPool, error),
) {
	systemCertPoolFn = fn
}

// ResetSystemCertPoolFn restores the systemCertPoolFn variable to its default.
func ResetSystemCertPoolFn() {
	systemCertPoolFn = x509.SystemCertPool
}

// SetSigningKeyFS replaces the os calls that read and write the signing key,
// and returns a func that puts the real ones back.
func SetSigningKeyFS(
	read func(string) ([]byte, error),
	write func(string, []byte, os.FileMode) error,
	ln func(string, string) error,
) func() {
	r, w, l := readFile, writeFile, link
	if read != nil {
		readFile = read
	}
	if write != nil {
		writeFile = write
	}
	if ln != nil {
		link = ln
	}

	return func() { readFile, writeFile, link = r, w, l }
}
