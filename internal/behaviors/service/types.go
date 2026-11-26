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

package service

// ListInput takes nothing: list reads every service.
type ListInput struct{}

// GetInput names the service to read.
type GetInput struct {
	Name string `json:"name" validate:"required,service_name" doc:"The service, such as nginx, with .service or without it"`
}

// Unit is the unit file of a service that osapi manages: the name that
// /etc/systemd/system/osapi-<name>.service has, and the object of the object
// store that holds its content.
type Unit struct {
	Name   string `json:"name"   validate:"required,service_name"  doc:"The service, such as app for osapi-app.service"                     telemetry:"attr"`
	Object string `json:"object" validate:"required,min=1,max=255" doc:"The object that holds the unit file, from osapi client file upload"`
}

// CreateInput is a unit file for osapi to deploy.
type CreateInput Unit

// UpdateInput is the unit file to deploy over the one osapi has.
type UpdateInput Unit

// Target names a service that osapi manages, for a delete or an action.
type Target struct {
	Name string `json:"name" validate:"required,service_name" doc:"The service, such as app for osapi-app.service" telemetry:"attr"`
}

// DeleteInput names the service whose unit file osapi drops.
type DeleteInput Target

// StartInput names the service that start acts on.
type StartInput Target

// StopInput names the service that stop acts on.
type StopInput Target

// RestartInput names the service that restart acts on.
type RestartInput Target

// EnableInput names the service that enable acts on.
type EnableInput Target

// DisableInput names the service that disable acts on.
type DisableInput Target

// Entry is a service of the host.
type Entry struct {
	Name        string `json:"name"        doc:"The service, without .service"`
	State       string `json:"state"       doc:"What systemd says it does now, such as running, exited or dead"`
	Enabled     bool   `json:"enabled"     doc:"Whether it starts at boot: the state of its unit file"`
	Description string `json:"description" doc:"The Description= of its unit"`
	Managed     bool   `json:"managed"     doc:"Whether osapi deployed it, so its writes and actions act on it"`
}

// Result is a service after a write or an action.
type Result struct {
	Name    string `json:"name"    doc:"The service"`
	Changed bool   `json:"changed" doc:"Whether the host changed"`
}
