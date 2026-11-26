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

// The kinds of key.
const (
	// KindSigning is a key that signs, such as the controller's.
	KindSigning = "signing"
	// KindEncryption is a key that encrypts, such as the JetStream store's.
	KindEncryption = "encryption"
)

// The states of a key.
const (
	// StateOK is a key that osapi read.
	StateOK = "ok"
	// StateNotMade is a key whose file is not there, as on a host where its
	// component has not started.
	StateNotMade = "not made"
	// StateOff is the JetStream key when the config sets no key_file: the
	// store is not encrypted.
	StateOff = "off"
)

// ListInput is the input of list, which has none.
type ListInput struct{}

// GetInput names a key.
type GetInput struct {
	Name string `json:"name" validate:"required,max=64" doc:"The name of the key, as list shows it, such as controller, agent, jetstream or vault:default."`
}

// CreateInput names the key to make.
type CreateInput struct {
	Name string `json:"name" validate:"required,max=64" doc:"The name of the key, as list shows it, such as jetstream or vault:default."`
}

// RotateInput names the key to rotate.
type RotateInput struct {
	Name string `json:"name" validate:"required,max=64" doc:"The name of the key, as list shows it: agent, or a vault key such as vault:default."`
}

// ListOutput is each key.
type ListOutput struct {
	Items []Key `json:"items"`
}

// RotateOutput is the key after a rotation, and, for a vault, how many
// values the new key encrypted again.
type RotateOutput struct {
	_      struct{} `title:"Rotated"`
	Key    Key      `                json:"key"`
	Values int      `                json:"values,omitempty"`
}

// Key is a key of osapi on this host as list and get show it. It never
// holds a secret: a public key shows its fingerprint, and a secret key its
// ID, the first 16 hex digits of its SHA-256.
type Key struct {
	Name        string `json:"name"`
	Kind        string `json:"kind"                  doc:"signing or encryption."`
	Component   string `json:"component"             doc:"The component that uses the key: controller, agent, nats or vault."`
	State       string `json:"state"                 doc:"ok; not made, when the component has not made it on this host; off, for a JetStream store with no key; or unreadable, with the reason."`
	File        string `json:"file"                  doc:"The file of the key, or $NAME for a vault key in an environment variable."                                                              table:"-"`
	Algorithm   string `json:"algorithm"             doc:"Ed25519, ECDH P-256 or AES-256-GCM."`
	Fingerprint string `json:"fingerprint,omitempty" doc:"The SHA-256 fingerprint of a public key. The agent's covers its signing and its encryption key, as its admission does."                 table:"-"`
	PublicKey   string `json:"public_key,omitempty"  doc:"The public key of the P-256 encryption key, in base64url."                                                                              table:"-" label:"Public key"`
	KeyID       string `json:"key_id,omitempty"      doc:"The ID of a secret key: the first 16 hex digits of its SHA-256."                                                                                  label:"Key ID"`
	Keys        int    `json:"keys,omitempty"        doc:"The number of keys in a vault's key file. More than one is a rotation that did not finish: rotate again."                               table:"-"`
	Admit       string `json:"admit,omitempty"       doc:"The command that admits this machine, for the agent's key."                                                                             table:"-" label:"To admit"`
}
