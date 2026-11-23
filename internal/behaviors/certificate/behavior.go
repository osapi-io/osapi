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

// Package certificate is the certificate behavior: the certificate
// authorities that the host trusts. It reads them through gohai on every
// platform, and on Linux adds and removes the ones that osapi deploys, from
// the Object Store, into the family's local directory, then rebuilds the
// trust store.
package certificate

import (
	"regexp"
	"strings"

	"github.com/osapi-io/osapi/internal/validation/rule"
	"github.com/osapi-io/osapi/pkg/sdk/client"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behaviors/managedfile"
	"github.com/osapi-io/osapi/internal/behaviors/object"
)

// pem is the certificate authority of the examples, and pem2 a second one
// that an update writes in its place.
const (
	pem = "-----BEGIN CERTIFICATE-----\n" +
		"MIIBvjCCAWOgAwIBAgIUUZjQew/w2XTBbe8VtXbYJ09SIa8wCgYIKoZIzj0EAwIw\n" +
		"KzEZMBcGA1UEAwwQb3NhcGkgZXhhbXBsZSBDQTEOMAwGA1UECgwFb3NhcGkwIBcN\n" +
		"MjYxMTIyMjMzODAxWhgPMjEyNjEwMjkyMzM4MDFaMCsxGTAXBgNVBAMMEG9zYXBp\n" +
		"IGV4YW1wbGUgQ0ExDjAMBgNVBAoMBW9zYXBpMFkwEwYHKoZIzj0CAQYIKoZIzj0D\n" +
		"AQcDQgAEYw7VLFKb+bykrwpFhkfkV1zCWS1PpnaxA7eBrMBNHd8fWKjO6ZRTmufJ\n" +
		"Zw5Cer8Qa/k/CWWaX4rPZR/N3trzSaNjMGEwHQYDVR0OBBYEFJMZvsnO2uSZngzL\n" +
		"s/2/sLvsxF9KMB8GA1UdIwQYMBaAFJMZvsnO2uSZngzLs/2/sLvsxF9KMA8GA1Ud\n" +
		"EwEB/wQFMAMBAf8wDgYDVR0PAQH/BAQDAgIEMAoGCCqGSM49BAMCA0kAMEYCIQCM\n" +
		"KPI9T2lEfDJdlv4h2AY9AhLdLk3PW9vqATbr4gi7PgIhAOqEAMpruGrDNfCo9n27\n" +
		"y9NIOpsPtN8EAH9a6bt+vnPx\n" +
		"-----END CERTIFICATE-----\n"
	pem2 = pem + "\n"

	subject     = "CN=osapi example CA,O=osapi"
	fingerprint = "6513d3f115f766cdffa3922e5ee47c4cccc98475cbe6471c5d32ddd4104c2312"
	notAfter    = "2126-10-29T23:38:01Z"

	// The system CA of the examples on each family, and the file that osapi
	// writes for internal-ca.
	debSystem = "/usr/share/ca-certificates/mozilla/Example_Root.crt"
	rhelSys   = "/usr/share/pki/ca-trust-source/anchors/Example_Root.pem"
	debPath   = "/usr/local/share/ca-certificates/osapi-internal-ca.crt"
	rhelPath  = "/etc/pki/ca-trust/source/anchors/osapi-internal-ca.crt"

	systemKeychain = "/System/Library/Keychains/SystemRootCertificates.keychain"
	localKeychain  = "/Library/Keychains/System.keychain"
)

// caNamePattern is a name of a CA that a path segment can hold: printable,
// with no slash. The router does not decode an encoded slash, %2F, so a %
// is refused too.
const caNamePattern = `^[^/%\x00-\x1f\x7f]+$`

var caNameRE = regexp.MustCompile(caNamePattern)

// The examples' CAs as list reads them.
var (
	debRoot = Entry{
		Name: "mozilla/Example_Root", Path: debSystem, Source: "system", Enabled: true,
		Subject: subject, Fingerprint: fingerprint, NotAfter: notAfter,
	}
	rhelRoot = Entry{
		Name: "anchors/Example_Root", Path: rhelSys, Source: "system", Enabled: true,
		Subject: subject, Fingerprint: fingerprint, NotAfter: notAfter,
	}
	debManaged = Entry{
		Name: "internal-ca", Path: debPath, Source: "local", Enabled: true,
		Subject: subject, Fingerprint: fingerprint, NotAfter: notAfter,
		Managed: true, Object: "internal-ca.pem",
	}
	rhelManaged = Entry{
		Name: "internal-ca", Path: rhelPath, Source: "local", Enabled: true,
		Subject: subject, Fingerprint: fingerprint, NotAfter: notAfter,
		Managed: true, Object: "internal-ca.pem",
	}
	macRoot = Entry{
		Name: "osapi example CA", Path: systemKeychain, Source: "system", Enabled: true,
		Subject: subject, Fingerprint: fingerprint, NotAfter: notAfter,
	}
)

// stored is the PEM that a person uploads once with osapi client file
// upload, and deployed the CA that create deployed from it.
var (
	stored   = behavior.Stored("internal-ca.pem", pem, client.ContentRaw)
	stored2  = behavior.Stored("internal-ca-2.pem", pem2, client.ContentRaw)
	deployed = behavior.After(CreateInput{Name: "internal-ca", Object: "internal-ca.pem"})
)

// Behavior is the certificate authorities that the host trusts. gohai reads
// each one. osapi writes and removes the CAs it deploys only: each has a file
// osapi-<name>.crt in the family's local directory that managedfile has a
// record of, whose content is an object that a person uploads once with
// osapi client file upload.
var Behavior = behavior.Register(behavior.Define(
	"certificate", "The certificate authorities that the host trusts",
	behavior.Node,
	behavior.API,
	behavior.Area(behavior.AreaSecurity),
	behavior.CLI,
	behavior.Requires(object.Behavior),
	behavior.Requires(managedfile.Behavior),
	behavior.Platform(behavior.PlatformDebian, newWriter),
	behavior.Platform(behavior.PlatformLinux, newWriter),
	behavior.Platform(behavior.PlatformDarwin, newReader),
	behavior.RunsOnPlatform(behavior.PlatformDarwin, "security"),
	// Each family rebuilds its trust store with its own program. On linux,
	// the RHEL family, Fedora and Arch run update-ca-trust.
	behavior.SudoOnPlatform(behavior.PlatformDebian, "update-ca-certificates"),
	behavior.SudoOnPlatform(behavior.PlatformLinux, "update-ca-trust"),
	behavior.Writes(
		"/usr/local/share/ca-certificates/osapi-*.crt",
		"/etc/pki/ca-trust/source/anchors/osapi-*.crt",
		"/etc/ca-certificates/trust-source/anchors/osapi-*.crt",
	),
	behavior.Validator("ca_name", caNameRE.MatchString,
		func(t rule.Target) { t.Pattern(caNamePattern) },
		"is not a name of a CA that a path can hold: it has a slash, a % or a control character"),
	behavior.Read("list", List).
		Describe("List the certificate authorities that the host trusts").
		Explain("Lists each CA of the trust store through gohai: the distribution's and "+
			"the local ones on Linux, from the family's directories, and the system "+
			"keychains on macOS, each with its subject, fingerprint and expiry. A CA "+
			"that osapi deployed reads by the name that it was deployed with, not its "+
			"file's osapi- name, and is managed.").
		Example("A Debian host with a system CA", ListInput{}, []Entry{debRoot},
			behavior.OnPlatform(behavior.PlatformDebian),
			behavior.Exists(debSystem, pem),
		).
		Example("A CA that osapi deployed reads by its name", ListInput{},
			[]Entry{debRoot, debManaged},
			behavior.OnPlatform(behavior.PlatformDebian),
			behavior.Exists(debSystem, pem),
			stored, deployed,
		).
		Example("A host of the RHEL family", ListInput{}, []Entry{rhelRoot},
			behavior.OnPlatform(behavior.PlatformLinux),
			behavior.Exists(rhelSys, pem),
		).
		Example("A Mac", ListInput{}, []Entry{macRoot},
			behavior.OnPlatform(behavior.PlatformDarwin),
			behavior.Prints("security find-certificate -a -p "+systemKeychain, pem),
			behavior.Prints("security find-certificate -a -p "+localKeychain, ""),
		),
	behavior.Read("get", Get).Key("name").
		Describe("Get one certificate authority").
		Explain("Reads one CA from the same list, by the name that list gives it. A CA "+
			"that the host does not trust is not found. A name with a slash, such as "+
			"a system CA of the Debian family, is in list only.").
		Example("A CA that osapi deployed, by the name it was deployed with",
			GetInput{Name: "internal-ca"}, debManaged,
			behavior.OnPlatform(behavior.PlatformDebian),
			stored, deployed,
		).
		Example("A CA that osapi deployed on the RHEL family",
			GetInput{Name: "internal-ca"}, rhelManaged,
			behavior.OnPlatform(behavior.PlatformLinux),
			stored, deployed,
		).
		Example("A CA of a Mac", GetInput{Name: "osapi example CA"}, macRoot,
			behavior.OnPlatform(behavior.PlatformDarwin),
			behavior.Prints("security find-certificate -a -p "+systemKeychain, pem),
			behavior.Prints("security find-certificate -a -p "+localKeychain, ""),
		).
		Fails("A CA that the host does not trust", GetInput{Name: "nothing"}, "not found").
		Fails("A name with a %, which an encoded slash has", GetInput{Name: "mozilla%2FExample_Root"}, "ca_name").
		Fails("A name of 256 characters", GetInput{Name: strings.Repeat("a", 256)}, "max"),
	behavior.Write("create", Create).
		On(behavior.PlatformDebian, behavior.PlatformLinux).
		Describe("Add a certificate authority to the trust store").
		Explain("Writes the PEM of the object as a file with the osapi- prefix, such as "+
			"osapi-internal-ca.crt for internal-ca, in the family's local "+
			"directory: /usr/local/share/ca-certificates on the Debian family, "+
			"/etc/pki/ca-trust/source/anchors on RHEL and Fedora. Then it rebuilds the "+
			"trust store with update-ca-certificates or update-ca-trust. SUSE and Alpine "+
			"report skipped. A CA that osapi "+
			"deployed already with the same content does not change and reports changed: "+
			"false. A file of that name that osapi did not write is refused. macOS "+
			"reports skipped.").
		Example("Add a CA", CreateInput{Name: "internal-ca", Object: "internal-ca.pem"},
			Result{Name: "internal-ca", Changed: true},
			behavior.OnPlatform(behavior.PlatformDebian),
			stored,
			behavior.Wrote(debPath, pem),
			behavior.Ran("update-ca-certificates"),
		).
		Example("Add a CA on the RHEL family", CreateInput{Name: "internal-ca", Object: "internal-ca.pem"},
			Result{Name: "internal-ca", Changed: true},
			behavior.OnPlatform(behavior.PlatformLinux),
			stored,
			behavior.Wrote(rhelPath, pem),
			behavior.Ran("update-ca-trust"),
		).
		Example("A CA that osapi deployed with the same content",
			CreateInput{Name: "internal-ca", Object: "internal-ca.pem"},
			Result{Name: "internal-ca"},
			stored, deployed,
		).
		Fails("A file that osapi did not write", CreateInput{Name: "internal-ca", Object: "internal-ca.pem"},
			"osapi did not deploy it",
			stored, behavior.Exists(debPath, pem), behavior.Exists(rhelPath, pem),
		).
		Fails("A name that is an option", CreateInput{Name: "-r", Object: "internal-ca.pem"},
			"not a valid name").
		Fails("No object", CreateInput{Name: "internal-ca"}, "required").
		Fails("A name with a space", CreateInput{Name: "a b", Object: "internal-ca.pem"}, "resource_name"),
	behavior.Write("update", Update).Key("name").
		On(behavior.PlatformDebian, behavior.PlatformLinux).
		Describe("Write a certificate authority that osapi deployed again").
		Explain("Writes the CA from the object, or from the one that it was deployed from "+
			"when the object is empty, and rebuilds the trust store when the file "+
			"changed. A CA that osapi did not deploy is not found. macOS reports skipped.").
		Example("Write a CA from a new object", UpdateInput{Name: "internal-ca", Object: "internal-ca-2.pem"},
			Result{Name: "internal-ca", Changed: true},
			behavior.OnPlatform(behavior.PlatformDebian),
			stored, stored2, deployed,
			behavior.Wrote(debPath, pem2),
			behavior.Ran("update-ca-certificates"),
		).
		Example("Write a CA from a new object on the RHEL family",
			UpdateInput{Name: "internal-ca", Object: "internal-ca-2.pem"},
			Result{Name: "internal-ca", Changed: true},
			behavior.OnPlatform(behavior.PlatformLinux),
			stored, stored2, deployed,
			behavior.Wrote(rhelPath, pem2),
			behavior.Ran("update-ca-trust"),
		).
		Example("The object that it was deployed from", UpdateInput{Name: "internal-ca"},
			Result{Name: "internal-ca"},
			stored, deployed,
		).
		Fails("A CA that osapi did not deploy", UpdateInput{Name: "nothing"}, "not found").
		Fails("A name that is an option", UpdateInput{Name: "-r"}, "not a valid name").
		Fails("A name with a space", UpdateInput{Name: "a b"}, "resource_name"),
	behavior.Write("delete", Delete).Key("name").
		On(behavior.PlatformDebian, behavior.PlatformLinux).
		Describe("Remove a certificate authority that osapi deployed").
		Explain("Removes the file of the CA and rebuilds the trust store: with "+
			"update-ca-certificates --fresh on the Debian family, which removes the links "+
			"of the CA from /etc/ssl/certs, and update-ca-trust on the RHEL family. A CA that osapi "+
			"did not deploy does not change and reports changed: false: a system CA is "+
			"not osapi's to remove. macOS reports skipped.").
		Example("Remove a CA", DeleteInput{Name: "internal-ca"},
			Result{Name: "internal-ca", Changed: true},
			behavior.OnPlatform(behavior.PlatformDebian),
			stored, deployed,
			behavior.Removed(debPath),
			behavior.Ran("update-ca-certificates --fresh"),
		).
		Example("Remove a CA on the RHEL family", DeleteInput{Name: "internal-ca"},
			Result{Name: "internal-ca", Changed: true},
			behavior.OnPlatform(behavior.PlatformLinux),
			stored, deployed,
			behavior.Removed(rhelPath),
			behavior.Ran("update-ca-trust"),
		).
		Example("A CA that osapi did not deploy", DeleteInput{Name: "internal-ca"},
			Result{Name: "internal-ca"},
		).
		Fails("A name that is an option", DeleteInput{Name: "-r"}, "not a valid name").
		Fails("A name with a space", DeleteInput{Name: "a b"}, "resource_name"),
))
