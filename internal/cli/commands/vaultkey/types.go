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

package vaultkey

// Input names the vault.
type Input struct {
	Vault string `json:"vault" validate:"required" doc:"The vault, as controller.secrets.vaults names it, such as default"`
}

// CreateOutput is the key that create made.
type CreateOutput struct {
	_       struct{} `title:"Vault key"`
	Name    string   `                  json:"name"`
	KeyFile string   `                  json:"key_file" label:"Key file"`
	KeyID   string   `                  json:"key_id"   label:"Key ID"`
}

// RotateOutput is the key that rotate made, and how many values it
// encrypted again.
type RotateOutput struct {
	_       struct{} `title:"Vault key"`
	Name    string   `                  json:"name"`
	KeyFile string   `                  json:"key_file" label:"Key file"`
	KeyID   string   `                  json:"key_id"   label:"Key ID"`
	Values  int      `                  json:"values"`
}
