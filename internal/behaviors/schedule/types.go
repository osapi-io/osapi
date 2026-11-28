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

package schedule

import "github.com/osapi-io/osapi/pkg/sdk/types"

// ListInput takes nothing: list reads every entry that osapi deployed.
type ListInput struct{}

// GetInput names the entry to read.
type GetInput struct {
	Name string `json:"name" validate:"required,max=64,resource_name" doc:"The entry, such as backup"`
}

// CreateInput is an entry for osapi to deploy: an object of the object store,
// written as /etc/cron.d/<name>, or as /etc/cron.<interval>/<name> for an
// interval.
type CreateInput struct {
	Name        string            `json:"name"                   validate:"required,max=64,resource_name"                                            doc:"The entry, and its file name, such as backup"                                                                         telemetry:"attr"`
	Object      string            `json:"object"                 validate:"required,min=1,max=255"                                                   doc:"The object that holds the entry, from osapi client file upload"`
	Schedule    string            `json:"schedule,omitempty"     validate:"required_without=Interval,excluded_with=Interval,omitempty,cron_schedule" doc:"A cron expression, such as */5 * * * *; give it or interval"`
	Interval    types.Interval    `json:"interval,omitempty"     validate:"required_without=Schedule,excluded_with=Schedule,omitempty,enum"          doc:"The interval of a script in /etc/cron.daily and the like; give it or schedule"`
	User        string            `json:"user,omitempty"         validate:"excluded_with=Interval,omitempty,account_name"                            doc:"The user a cron expression's entry runs as; root when empty. A script of an interval runs as root"`
	ContentType types.ContentType `json:"content_type,omitempty" validate:"omitempty,enum"                                                           doc:"The content type: raw, or template to render the object with the agent's facts and vars; the object's own when empty"`
	Vars        map[string]string `json:"vars,omitempty"         validate:"omitempty"                                                                doc:"The vars of a template, as key=value"`
}

// UpdateInput is a new object, or new settings, for an entry that osapi
// deployed. An empty field keeps what the entry has.
type UpdateInput struct {
	Name        string            `json:"name"                   validate:"required,max=64,resource_name" doc:"The entry, such as backup"                                      telemetry:"attr"`
	Object      string            `json:"object,omitempty"       validate:"omitempty,min=1,max=255"       doc:"The object to deploy; the entry's own when empty"`
	Schedule    string            `json:"schedule,omitempty"     validate:"omitempty,cron_schedule"       doc:"A new cron expression; the entry's own when empty"`
	User        string            `json:"user,omitempty"         validate:"omitempty,account_name"        doc:"The user the entry runs as; the entry's own when empty"`
	ContentType types.ContentType `json:"content_type,omitempty" validate:"omitempty,enum"                doc:"The content type, raw or template; the object's own when empty"`
	Vars        map[string]string `json:"vars,omitempty"         validate:"omitempty"                     doc:"The vars of a template, as key=value"`
}

// DeleteInput names the entry that osapi removes.
type DeleteInput struct {
	Name string `json:"name" validate:"required,max=64,resource_name" doc:"The entry, such as backup" telemetry:"attr"`
}

// Entry is an entry that osapi deployed.
type Entry struct {
	Name     string `json:"name"               doc:"The entry"`
	Object   string `json:"object"             doc:"The object it was deployed from"`
	Schedule string `json:"schedule,omitempty" doc:"Its cron expression, for an entry in /etc/cron.d"`
	Interval string `json:"interval,omitempty" doc:"Its interval, for a script in /etc/cron.daily and the like"`
	User     string `json:"user,omitempty"     doc:"The user it runs as"`
	Source   string `json:"source"             doc:"Where its file is: cron.d, or the interval"`
}

// Result is an entry after a write.
type Result struct {
	Name    string `json:"name"    doc:"The entry"`
	Changed bool   `json:"changed" doc:"Whether the host changed"`
}
