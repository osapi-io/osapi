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

package agentservice

// InstallInput is the join string, and the labels that the host adds.
type InstallInput struct {
	Join   string            `json:"join"   validate:"required,max=8192" doc:"The join string that osapi controller join token create printed."`
	Labels map[string]string `json:"labels" validate:"omitempty"         doc:"Labels of the host, as key=value, beside the labels of the join token."`
	Force  bool              `json:"force"  validate:"omitempty"         doc:"Replace a config file that exists."`
}

// InstallOutput is what install wrote and whether the host joined.
type InstallOutput struct {
	_          struct{} `title:"Service"`
	Service    string   `                json:"service"`
	ConfigFile string   `                json:"config_file"    label:"Config"`
	Joined     bool     `                json:"joined"`
	Next       string   `                json:"next,omitempty"`
}

// UninstallInput is whether to remove the config and the keys too.
type UninstallInput struct {
	Purge bool `json:"purge" validate:"omitempty" doc:"Also remove the config file and the keys of agent.pki.key_dir."`
}

// UninstallOutput is each file that uninstall removed.
type UninstallOutput struct {
	_       struct{} `title:"Removed"`
	Removed []string `                json:"removed"`
}
