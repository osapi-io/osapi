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

import (
	"strings"

	"github.com/osapi-io/osapi/pkg/sdk/client"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behaviors/managedfile"
	"github.com/osapi-io/osapi/internal/behaviors/object"
)

// The objects of the examples: an entry with a cron expression, a new
// version of it, a daily script, and a template.
const (
	backupText   = "0 2 * * * root /usr/local/bin/backup\n"
	backupV2Text = "30 2 * * * root /usr/local/bin/backup --full\n"
	scriptText   = "#!/bin/sh\n/usr/local/bin/backup\n"
	templateText = "0 2 * * * root /usr/local/bin/backup {{ .Vars.target }}\n"
	backupPath   = "/etc/cron.d/osapitest"
	scriptPath   = "/etc/cron.daily/osapitest"
)

var (
	stored   = behavior.Stored("osapitest.cron", backupText, client.ContentRaw)
	storedV2 = behavior.Stored("osapitest-v2.cron", backupV2Text, client.ContentRaw)
	deployed = behavior.After(
		CreateInput{
			Name:     "osapitest",
			Object:   "osapitest.cron",
			Schedule: "0 2 * * *",
			User:     "root",
		},
	)
	// theirs is a cron file at the path that osapi did not deploy.
	theirs = behavior.Exists(backupPath, backupText)
)

// entry is the entry that the examples deploy, as a read gives it.
var entry = Entry{
	Name: "osapitest", Object: "osapitest.cron", Schedule: "0 2 * * *", User: "root",
	Source: "cron.d",
}

// Behavior is the cron entries that osapi deploys. Each is an object that a
// person uploads once with osapi client file upload, written as a file that
// managedfile keeps a record of: /etc/cron.d/<name> for an entry with a
// cron expression, /etc/cron.<interval>/<name> for a script that runs
// hourly, daily, weekly or monthly. A cron file that osapi did not deploy is
// not osapi's to read or change. A cron entry runs a command as its user,
// root included, so each write is admin only.
var Behavior = behavior.Register(behavior.Define(
	"schedule", "The cron entries that osapi deploys",
	behavior.Node,
	behavior.API,
	behavior.Area(behavior.AreaConfig),
	behavior.CLI,
	behavior.Requires(object.Behavior),
	behavior.Requires(managedfile.Behavior),
	behavior.Platform(behavior.PlatformDebian, newDebian),
	behavior.Writes(
		"/etc/cron.d/*", "/etc/cron.hourly/*", "/etc/cron.daily/*",
		"/etc/cron.weekly/*", "/etc/cron.monthly/*",
	),
	behavior.Read("list", List).
		Describe("List the cron entries that osapi deployed").
		Explain("Lists each file in /etc/cron.d and in /etc/cron.hourly, daily, weekly and "+
			"monthly that osapi deployed, with its object, its cron expression or interval, "+
			"and its user. A cron file that osapi did not deploy is not listed.").
		Example("An entry that osapi deployed", ListInput{}, []Entry{entry}, stored, deployed).
		Example("A daily script that osapi deployed", ListInput{},
			[]Entry{
				{Name: "osapitest", Object: "osapitest.sh", Interval: "daily", Source: "daily"},
			},
			behavior.Stored("osapitest.sh", scriptText, client.ContentRaw),
			behavior.After(CreateInput{Name: "osapitest", Object: "osapitest.sh", Interval: "daily"}),
		).
		Example("A host where osapi deployed none", ListInput{}, []Entry{}, theirs),
	behavior.Read("get", Get).Key("name").
		Describe("Get one cron entry that osapi deployed").
		Explain("Reads the entry from /etc/cron.d, then from each interval directory. "+
			"An entry that osapi did not deploy is not found.").
		Example("An entry that osapi deployed", GetInput{Name: "osapitest"}, entry, stored, deployed).
		Fails("An entry that osapi did not deploy", GetInput{Name: "osapitest"}, "not found", theirs).
		Fails("A name with a space", GetInput{Name: "a b"}, "Name").
		Fails("A name that is an option", GetInput{Name: "--now"}, "not a valid entry name").
		Fails("A name over 64 characters", GetInput{Name: strings.Repeat("a", 65)}, "Name"),
	behavior.Write("create", Create).
		AdminOnly().
		Describe("Deploy a cron entry").
		Explain("Gets the object from the object store and writes it as /etc/cron.d/backup for "+
			"backup, mode 0644, for a cron expression, or as /etc/cron.daily/backup, mode 0755, "+
			"for a daily interval. An object of content type template, or content_type template, "+
			"is rendered with the agent's facts and the vars. A file that has the content "+
			"does not change. An entry of the name in a different directory does not change: "+
			"move it with delete and create. A file at the path that osapi did not deploy is "+
			"refused. An entry runs as its user, root included, so only admin can deploy one.").
		Example("An entry with a cron expression",
			CreateInput{
				Name:     "osapitest",
				Object:   "osapitest.cron",
				Schedule: "0 2 * * *",
				User:     "root",
			},
			Result{Name: "osapitest", Changed: true},
			stored,
			behavior.Wrote(backupPath, backupText),
		).
		Example("A script that runs daily",
			CreateInput{Name: "osapitest", Object: "osapitest.sh", Interval: "daily"},
			Result{Name: "osapitest", Changed: true},
			behavior.Stored("osapitest.sh", scriptText, client.ContentRaw),
			behavior.Wrote(scriptPath, scriptText),
		).
		Example("A template, rendered with its vars",
			CreateInput{
				Name:     "osapitest",
				Object:   "osapitest.tmpl",
				Schedule: "0 2 * * *",
				ContentType: string(
					client.ContentTemplate,
				),
				Vars: map[string]string{"target": "/srv"},
			},
			Result{Name: "osapitest", Changed: true},
			behavior.Stored("osapitest.tmpl", templateText, client.ContentRaw),
			behavior.Wrote(backupPath, "0 2 * * * root /usr/local/bin/backup /srv\n"),
		).
		Example("An entry that has the content",
			CreateInput{
				Name:     "osapitest",
				Object:   "osapitest.cron",
				Schedule: "0 2 * * *",
				User:     "root",
			},
			Result{Name: "osapitest"},
			stored, deployed,
		).
		Example("An entry of the name in a different directory",
			CreateInput{Name: "osapitest", Object: "osapitest.sh", Interval: "daily"},
			Result{Name: "osapitest"},
			stored, deployed,
			behavior.Stored("osapitest.sh", scriptText, client.ContentRaw),
		).
		Fails("An object the store does not have",
			CreateInput{Name: "osapitest", Object: "nosuch", Schedule: "0 2 * * *"}, "not found").
		Fails("A template that reads a var the request does not give",
			CreateInput{
				Name: "osapitest", Object: "osapitest.tmpl", Schedule: "0 2 * * *",
				ContentType: string(client.ContentTemplate),
			},
			"render template",
			behavior.Stored("osapitest.tmpl", templateText, client.ContentRaw)).
		Fails("A file that osapi did not deploy",
			CreateInput{Name: "osapitest", Object: "osapitest.cron", Schedule: "0 2 * * *"},
			"osapi did not deploy it", stored, theirs).
		Fails("No object", CreateInput{Name: "osapitest", Schedule: "0 2 * * *"}, "Object").
		Fails("Neither a cron expression nor an interval",
			CreateInput{Name: "osapitest", Object: "osapitest.cron"}, "Schedule").
		Fails("A cron expression and an interval",
			CreateInput{
				Name:     "osapitest",
				Object:   "osapitest.cron",
				Schedule: "0 2 * * *",
				Interval: "daily",
			},
			"Schedule").
		Fails("A cron expression that does not parse",
			CreateInput{
				Name:     "osapitest",
				Object:   "osapitest.cron",
				Schedule: "every day",
			}, "Schedule").
		Fails("An interval that is not one",
			CreateInput{
				Name:     "osapitest",
				Object:   "osapitest.cron",
				Interval: "yearly",
			}, "Interval").
		Fails("A name with a slash",
			CreateInput{Name: "a/b", Object: "osapitest.cron", Schedule: "0 2 * * *"}, "Name").
		Fails("A name that is an option",
			CreateInput{
				Name:     "--now",
				Object:   "osapitest.cron",
				Schedule: "0 2 * * *",
			}, "not a valid entry name").
		Fails("A name over 64 characters",
			CreateInput{
				Name:     strings.Repeat("a", 65),
				Object:   "osapitest.cron",
				Schedule: "0 2 * * *",
			}, "Name").
		Fails("An object name over 255 characters",
			CreateInput{
				Name:     "osapitest",
				Object:   strings.Repeat("a", 256),
				Schedule: "0 2 * * *",
			}, "Object").
		Fails("A user that is not an account name",
			CreateInput{
				Name:     "osapitest",
				Object:   "osapitest.cron",
				Schedule: "0 2 * * *",
				User:     "a b",
			},
			"User").
		Fails("A content type that is not one",
			CreateInput{
				Name:        "osapitest",
				Object:      "osapitest.cron",
				Schedule:    "0 2 * * *",
				ContentType: "yaml",
			},
			"ContentType"),
	behavior.Write("update", Update).Key("name").
		AdminOnly().
		Describe("Deploy a cron entry that osapi deployed again").
		Explain("Writes a new object, or the entry's own, at the entry's path, with the "+
			"fields that the request gives and the entry's own for the rest. A file that has "+
			"the content does not change. An entry that osapi did not deploy is not found. "+
			"The entry stays in its directory: an interval does not change with update.").
		Example("A new object",
			UpdateInput{Name: "osapitest", Object: "osapitest-v2.cron"},
			Result{Name: "osapitest", Changed: true},
			stored, deployed, storedV2,
			behavior.Wrote(backupPath, backupV2Text),
		).
		Example("A new object, cron expression and user",
			UpdateInput{
				Name:     "osapitest",
				Object:   "osapitest-v2.cron",
				Schedule: "30 2 * * *",
				User:     "nobody",
			},
			Result{Name: "osapitest", Changed: true},
			stored, deployed, storedV2,
			behavior.Wrote(backupPath, backupV2Text),
		).
		Example("A daily script with a new object",
			UpdateInput{Name: "osapitest", Object: "osapitest-v2.sh"},
			Result{Name: "osapitest", Changed: true},
			behavior.Stored("osapitest.sh", scriptText, client.ContentRaw),
			behavior.After(CreateInput{Name: "osapitest", Object: "osapitest.sh", Interval: "daily"}),
			behavior.Stored("osapitest-v2.sh", scriptText+"/usr/local/bin/prune\n", client.ContentRaw),
			behavior.Wrote(scriptPath, scriptText+"/usr/local/bin/prune\n"),
		).
		Fails("An entry that osapi did not deploy",
			UpdateInput{Name: "osapitest", Object: "osapitest.cron"}, "not found").
		Fails("A cron expression that does not parse",
			UpdateInput{Name: "osapitest", Schedule: "every day"}, "Schedule").
		Fails("A name with a space", UpdateInput{Name: "a b"}, "Name").
		Fails("A name that is an option", UpdateInput{Name: "--now"}, "not a valid entry name").
		Fails("A name over 64 characters", UpdateInput{Name: strings.Repeat("a", 65)}, "Name").
		Fails("An object name over 255 characters",
			UpdateInput{Name: "osapitest", Object: strings.Repeat("a", 256)}, "Object").
		Fails("A user that is not an account name", UpdateInput{Name: "osapitest", User: "a b"}, "User").
		Fails("A content type that is not one", UpdateInput{Name: "osapitest", ContentType: "yaml"},
			"ContentType"),
	behavior.Write("delete", Delete).Key("name").
		AdminOnly().
		Describe("Remove a cron entry that osapi deployed").
		Explain("Removes the entry's file and its record. An entry that osapi did not "+
			"deploy does not change, and a cron file that osapi did not deploy stays.").
		Example("Remove an entry", DeleteInput{Name: "osapitest"},
			Result{Name: "osapitest", Changed: true},
			stored, deployed,
			behavior.Removed(backupPath),
		).
		Example("An entry that osapi did not deploy", DeleteInput{Name: "osapitest"},
			Result{Name: "osapitest"},
		).
		Fails("A name with a space", DeleteInput{Name: "a b"}, "Name").
		Fails("A name that is an option", DeleteInput{Name: "--now"}, "not a valid entry name").
		Fails("A name over 64 characters", DeleteInput{Name: strings.Repeat("a", 65)}, "Name"),
))
