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

package secret

import (
	"time"

	secretvalue "github.com/osapi-io/osapi/internal/secret"
)

// VaultListInput is the input of vault list, which has none.
type VaultListInput struct{}

// ListInput names the vault whose keys to list.
type ListInput struct {
	Vault string `json:"vault" in:"path" validate:"required,resource_name,max=64" doc:"The name of the vault, such as default." example:"default"`
}

// GetInput names one key of a vault.
type GetInput struct {
	Vault string `json:"vault" in:"path" validate:"required,resource_name,max=64"  doc:"The name of the vault, such as default."      example:"default"`
	Key   string `json:"key"   in:"path" validate:"required,resource_name,max=128" doc:"The name of the key, such as registry-token." example:"registry-token"`
}

// RevealInput names the key whose value to reveal.
type RevealInput struct {
	Vault string `json:"vault" in:"path" validate:"required,resource_name,max=64"  doc:"The name of the vault, such as default."      example:"default"`
	Key   string `json:"key"   in:"path" validate:"required,resource_name,max=128" doc:"The name of the key, such as registry-token." example:"registry-token"`
}

// Value is the value of a key: the answer of reveal only. The CLI
// prints it as it is, for a pipe.
type Value struct {
	_ struct{} `schema:"SecretValue"`

	Value string `json:"value" doc:"The value of the key." example:"glpat-xxxxxxxxxxxxxxxxxxxx" render:"body"`
}

// DeleteInput names the key to delete.
type DeleteInput struct {
	Vault string `json:"vault" in:"path" validate:"required,resource_name,max=64"  doc:"The name of the vault, such as default."      example:"default"`
	Key   string `json:"key"   in:"path" validate:"required,resource_name,max=128" doc:"The name of the key, such as registry-token." example:"registry-token"`
}

// UpdateInput is the key to write, and its new value.
type UpdateInput struct {
	Vault string        `json:"vault" in:"path" validate:"required,resource_name,max=64"  doc:"The name of the vault, such as default."      example:"default"`
	Key   string        `json:"key"   in:"path" validate:"required,resource_name,max=128" doc:"The name of the key, such as registry-token." example:"registry-token"`
	Body  UpdateRequest `json:"body"  in:"body" validate:"required"`
}

// UpdateRequest is the value of a key. The API never gives it back.
type UpdateRequest struct {
	_ struct{} `schema:"SecretUpdateRequest"`

	Value secretvalue.Value `json:"value" validate:"required,max=65536" doc:"The value of the key, which no method gives back"`
}

// VaultInfo is one vault of controller.secrets.vaults.
type VaultInfo struct {
	Name    string   `json:"name"    doc:"The name of the vault, as a reference gives it."             example:"default"`
	Type    string   `json:"type"    doc:"The type of the vault, such as nats."                        example:"nats"`
	Targets []string `json:"targets" doc:"The targets that its secrets can go to. Empty is each host."                   empty:"each host"`
	Answers bool     `json:"answers" doc:"Whether the vault answered a list of its keys."              example:"true"`
}

// VaultListResponse is each vault of the controller.
type VaultListResponse struct {
	_          struct{}    `title:"Vaults"`
	Items      []VaultInfo `               json:"items"       doc:"The vaults of the controller." empty:"The controller has no vault: add one to controller.secrets.vaults."`
	TotalItems int         `               json:"total_items" doc:"The number of vaults."                                                                                    example:"1"`
}

// Info is the metadata of one key. It has no value: no method gives
// one.
type Info struct {
	_ struct{} `schema:"SecretInfo"`

	Vault   string    `json:"vault"   doc:"The name of the vault."                                          example:"default"              table:"-"`
	Key     string    `json:"key"     doc:"The name of the key."                                            example:"registry-token"`
	Version int64     `json:"version" doc:"1 for the first value of the key, and one more for each update." example:"3"`
	Changed time.Time `json:"changed" doc:"When the last update occurred."                                  example:"2026-11-23T22:56:50Z"           render:"time"`
	By      string    `json:"by"      doc:"Who did the last update, or empty when it is not known."         example:"ops@example.com"                              none:"-"`
}

// ListResponse is the metadata of each key of a vault.
type ListResponse struct {
	_          struct{} `title:"Keys" schema:"SecretListResponse"`
	Items      []Info   `                                         json:"items"       doc:"The keys of the vault, sorted by key." empty:"The vault has no key. Add one with osapi client secret update."`
	TotalItems int      `                                         json:"total_items" doc:"The number of keys."                                                                                          example:"1"`
}
