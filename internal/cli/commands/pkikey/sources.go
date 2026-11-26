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

package pkikey

import (
	"path/filepath"

	"github.com/osapi-io/osapi/internal/secret/vault/natsvault"
)

// The types of key, by how osapi reads its file.
const (
	typeSigning    = "signing"
	typeAgentEKey  = "ekey"
	typeJetStream  = "jetstream"
	typeVault      = "vault"
	vaultPrefix    = "vault:"
	jetStreamName  = "jetstream"
	agentName      = "agent"
	controllerName = "controller"
)

// source is a key of the config: its name, kind, component and algorithm,
// where it is, and how osapi reads it.
type source struct {
	name      string
	kind      string
	component string
	algorithm string
	typ       string
	// dir and prefix are the key directory and the prefix of the files of a
	// key of package pki: prefix.key, prefix.pub and prefix.ekey.
	dir    string
	prefix string
	// file is the file of a secret key, and env its environment variable
	// for a vault with key_env.
	file string
	env  string
}

// sources are the keys of osapi that the config names: the controller's
// signing key, the agent's signing and encryption keys, the JetStream
// store's key, and the key of each vault of type nats.
func sources(
	d Deps,
) []source {
	c := d.Config
	out := []source{
		{
			name: controllerName, kind: KindSigning, component: "controller",
			algorithm: "Ed25519", typ: typeSigning,
			dir: c.Controller.PKI.KeyDir, prefix: "controller",
		},
		{
			name: agentName, kind: KindSigning, component: "agent",
			algorithm: "Ed25519", typ: typeSigning,
			dir: c.Agent.PKI.KeyDir, prefix: "agent",
		},
		{
			name: "agent-encryption", kind: KindEncryption, component: "agent",
			algorithm: "ECDH P-256", typ: typeAgentEKey,
			dir: c.Agent.PKI.KeyDir, prefix: "agent",
		},
		{
			name: jetStreamName, kind: KindEncryption, component: "nats",
			algorithm: "AES-256-GCM", typ: typeJetStream,
			file: c.NATS.Server.JetStream.KeyFile,
		},
	}
	for _, v := range c.Controller.Secrets.Vaults {
		if v.Type != natsvault.TypeName {
			continue
		}
		out = append(out, source{
			name: vaultPrefix + v.Name, kind: KindEncryption, component: "vault",
			algorithm: "AES-256-GCM", typ: typeVault, file: v.KeyFile, env: v.KeyEnv,
		})
	}

	return out
}

// files are the files of s: the private and the public file of a signing
// key, the PEM file of a P-256 encryption key, or the file of a secret key.
func (s source) files() []string {
	switch s.typ {
	case typeSigning:
		return []string{
			filepath.Join(s.dir, s.prefix+".key"), filepath.Join(s.dir, s.prefix+".pub"),
		}
	case typeAgentEKey:
		return []string{filepath.Join(s.dir, s.prefix+".ekey")}
	default:
		return []string{s.file}
	}
}

// display is where s is, as list shows it.
func (s source) display() string {
	if s.typ == typeVault && s.file == "" && s.env != "" {
		return "$" + s.env
	}

	return s.files()[0]
}
