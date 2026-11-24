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

// Package secret is the endpoint group of the secrets in the vaults of the
// controller: the vaults, the metadata of each key, and a write of a value.
// No method gives a value.
package secret

import (
	"errors"
	"net/http"
	"strings"

	"github.com/osapi-io/osapi/internal/area"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
)

var (
	// ErrNoVault is a vault that controller.secrets.vaults does not name.
	ErrNoVault = errors.New("vault not found")
	// ErrNoKey is a key that is not in its vault.
	ErrNoKey = errors.New("key not found")
)

// noVault404 is the 404 of each method that names a vault.
var noVault404 = endpoint.NotFound(ErrNoVault, "The controller has no vault with that name.")

// Group is the secret group's definition.
var Group = endpoint.Register(endpoint.Define(
	"secret", "Manage the secrets in the vaults of the controller",
	endpoint.Area(area.Secrets),
	endpoint.Title("Secret Management API"),
	endpoint.Tag("secret_operations", "Secret",
		"The vaults of the controller and the metadata of their keys. No operation gives a value."),
	endpoint.CommandLong("A secret is a key in a vault of the controller. A request uses it as a\n"+
		"reference, secret://<vault>/<key>. These commands write a value and read the\n"+
		"metadata of each key. No command shows a value."),
	endpoint.SubDoc("vault", "Read the vaults of the controller"),
	endpoint.Uses[Deps](),
	endpoint.Get[VaultListInput, VaultListResponse]("vaults").
		Handle(VaultList).
		CommandSub("vault").
		Use("list").
		ID("GetSecretVaults").
		Describe("List the vaults").
		Explain("List each vault of controller.secrets.vaults, with its type, its targets "+
			"and whether it answers.").
		Returns("The vaults of the controller.").
		Command("List the vaults of the controller",
			"The list shows each vault of controller.secrets.vaults. A vault that does\n"+
				"not answer is a backend that the controller cannot reach now.").
		Example("").
		Prints(sampleVaults).
		Example("--json"),
	endpoint.Get[ListInput, SecretListResponse]("list").
		Handle(List).
		Key("vault").
		ID("GetSecrets").
		Describe("List the keys of a vault").
		Explain("List the metadata of each key of a vault: its version, when it changed "+
			"and who changed it. The list has no value.").
		Returns("The keys of the vault.").
		Command("List the keys of a vault",
			"The list shows the metadata of each key, and no value. A request uses a\n"+
				"key as secret://<vault>/<key>.").
		Example("--vault default").
		Prints(sampleList).
		Example("--vault default --json").
		Fails("a vault name that is not letters, digits, - or _", ListInput{Vault: "a.b"}).
		Fails("a vault name over 64 characters", ListInput{Vault: long(65)}).
		Errors(noVault404),
	endpoint.Get[GetInput, SecretInfo]("get").
		Handle(Get).
		Key("vault", "key").
		ID("GetSecret").
		Describe("Get the metadata of a key").
		Explain("Get the version of a key, when it changed and who changed it. "+
			"The answer has no value.").
		Returns("The metadata of the key.").
		Command("Get the metadata of a key", "The command shows the metadata of the key, and never its value.").
		Example("--vault default --key registry-token").
		Prints(sampleInfo).
		Example("--vault default --key registry-token --json").
		Fails("a vault name that is not letters, digits, - or _", GetInput{Vault: "a.b", Key: "k"}).
		Fails("a vault name over 64 characters", GetInput{Vault: long(65), Key: "k"}).
		Fails("a key that is not letters, digits, - or _", GetInput{Vault: "default", Key: "a.b"}).
		Fails("a key over 128 characters", GetInput{Vault: "default", Key: long(129)}).
		Errors(noVault404, endpoint.NotFound(ErrNoKey, "The vault has no key with that name.")),
	endpoint.Put[UpdateInput, SecretInfo]("update").
		Handle(Update).
		Key("vault", "key").
		ID("PutSecret").
		Describe("Write the value of a key").
		Explain("Write a value as the next version of a key, or as version 1 of a new key. "+
			"The answer is the metadata of the key. The API never gives the value back.").
		Returns("The metadata of the key after the write.").
		Command("Write the value of a key",
			"The command reads the value from standard input with --value -, so the\n"+
				"value is not in the shell history. A request then uses the key as\n"+
				"secret://<vault>/<key>.").
		Example("--vault default --key registry-token --value -",
			"Read the value from standard input").
		Prints(sampleUpdate).
		Fails("no value", UpdateInput{Vault: "default", Key: "k"}).
		Fails("a value over 64 KiB", UpdateInput{Vault: "default", Key: "k", Body: body(65537)}).
		Fails("a vault name that is not letters, digits, - or _", UpdateInput{Vault: "a.b", Key: "k", Body: body(1)}).
		Fails("a vault name over 64 characters", UpdateInput{Vault: long(65), Key: "k", Body: body(1)}).
		Fails("a key that is not letters, digits, - or _", UpdateInput{Vault: "default", Key: "a.b", Body: body(1)}).
		Fails("a key over 128 characters", UpdateInput{Vault: "default", Key: long(129), Body: body(1)}).
		Errors(noVault404),
	endpoint.Delete[DeleteInput, struct{}]("delete").
		Handle(Delete).
		Key("vault", "key").
		ID("DeleteSecret").
		Describe("Delete a key").
		Explain("Delete a key from its vault. A request that uses its reference after "+
			"that fails.").
		Returns("The key is deleted.").
		Status(http.StatusNoContent).
		Command("Delete a key from a vault", "").
		Example("--vault default --key registry-token").
		Fails("a vault name that is not letters, digits, - or _", DeleteInput{Vault: "a.b", Key: "k"}).
		Fails("a vault name over 64 characters", DeleteInput{Vault: long(65), Key: "k"}).
		Fails("a key that is not letters, digits, - or _", DeleteInput{Vault: "default", Key: "a.b"}).
		Fails("a key over 128 characters", DeleteInput{Vault: "default", Key: long(129)}).
		Errors(noVault404, endpoint.NotFound(ErrNoKey, "The vault has no key with that name.")),
))

// long is a name of n characters, for the inputs that Fails declares.
func long(
	n int,
) string {
	return strings.Repeat("a", n)
}
