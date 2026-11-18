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

package validation_test

import (
	"context"
	"strings"
	"testing"

	"github.com/osapi-io/osapi/internal/validation/rule"

	"github.com/stretchr/testify/suite"

	"github.com/go-playground/validator/v10"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/validation"
)

// motdIn uses a validate tag a behavior declares, which the validator learns
// when the behavior is defined, after it started.
type motdIn struct {
	Text string `json:"text" validate:"motd_text"`
}

var _ = behavior.Define("motd", "Message of the day",
	behavior.Validator("motd_text",
		func(v string) bool { return !strings.Contains(v, "\x1b") },
		rule.Pattern("^[^\\x1b]*$"),
		"is not a valid message: it has an escape sequence"),
)

type ValidationPublicTestSuite struct {
	suite.Suite
}

type breaksBody struct {
	Target string `validate:"omitempty,max=3"`
}

type breaksInput struct {
	Name string     `validate:"required,max=3"`
	Body breaksBody `validate:"omitempty"`
}

type breaksUnknown struct {
	Name string `validate:"breaks_not_registered"`
}

// TestBreaks names the field and the rule that each input breaks, so a
// method's .Fails inputs can be held to its rules.
func (s *ValidationPublicTestSuite) TestBreaks() {
	tests := []struct {
		name string
		in   any
		want []string
	}{
		{
			name: "when the input is valid, it breaks nothing",
			in:   breaksInput{Name: "abc"},
			want: nil,
		},
		{
			name: "when fields break rules, each is named with its Go path",
			in:   breaksInput{Name: "abcd", Body: breaksBody{Target: "abcd"}},
			want: []string{"Name:max", "Body.Target:max"},
		},
		{
			name: "when the input is not a struct, it breaks nothing",
			in:   "x",
			want: nil,
		},
		{
			name: "when a tag has no validator, it breaks nothing",
			in:   breaksUnknown{Name: "x"},
			want: nil,
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, validation.Breaks(tc.in))
		})
	}
}

func (s *ValidationPublicTestSuite) TestStruct() {
	type testStruct struct {
		Name  string `validate:"required"`
		Email string `validate:"required,email"`
	}

	tests := []struct {
		name         string
		input        any
		validateFunc func(string, bool)
	}{
		{
			name: "when valid struct",
			input: testStruct{
				Name:  "test",
				Email: "test@example.com",
			},
			validateFunc: func(_ string, ok bool) {
				s.Equal(true, ok)
			},
		},
		{
			name: "when missing required field",
			input: testStruct{
				Email: "test@example.com",
			},
			validateFunc: func(errMsg string, ok bool) {
				s.Equal(false, ok)
				s.Contains(errMsg, "Name")
				s.Contains(errMsg, "required")
			},
		},
		{
			name:  "when a value passes a rule a behavior declares",
			input: motdIn{Text: "Back at 5pm."},
			validateFunc: func(_ string, ok bool) {
				s.True(ok)
			},
		},
		{
			name:  "when a value fails it, the message is the behavior's",
			input: motdIn{Text: "\x1b[31mred"},
			validateFunc: func(errMsg string, ok bool) {
				s.False(ok)
				s.Contains(errMsg, "is not a valid message: it has an escape sequence")
			},
		},
		{
			name: "when invalid email",
			input: testStruct{
				Name:  "test",
				Email: "not-an-email",
			},
			validateFunc: func(errMsg string, ok bool) {
				s.Equal(false, ok)
				s.Contains(errMsg, "Email")
				s.Contains(errMsg, "email")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.validateFunc(validation.Struct(context.Background(), tt.input))
		})
	}
}

func (s *ValidationPublicTestSuite) TestVar() {
	tests := []struct {
		name         string
		field        any
		tag          string
		validateFunc func(string, bool)
	}{
		{
			name:  "when valid field",
			field: "hello",
			tag:   "required",
			validateFunc: func(_ string, ok bool) {
				s.Equal(true, ok)
			},
		},
		{
			name:  "when empty required field",
			field: "",
			tag:   "required",
			validateFunc: func(errMsg string, ok bool) {
				s.Equal(false, ok)
				s.Contains(errMsg, "required")
			},
		},
		{
			name:  "when invalid email",
			field: "not-an-email",
			tag:   "email",
			validateFunc: func(errMsg string, ok bool) {
				s.Equal(false, ok)
				s.Contains(errMsg, "email")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.validateFunc(validation.Var(context.Background(), tt.field, tt.tag))
		})
	}
}

func (s *ValidationPublicTestSuite) TestAlphanumOrFact() {
	tests := []struct {
		name         string
		field        string
		validateFunc func(bool)
	}{
		{
			name:  "when alphanumeric value",
			field: "eth0",
			validateFunc: func(got bool) {
				s.Equal(true, got)
			},
		},
		{
			name:  "when fact reference",
			field: "@fact.interface.primary",
			validateFunc: func(got bool) {
				s.Equal(true, got)
			},
		},
		{
			name:  "when fact custom reference",
			field: "@fact.custom.mykey",
			validateFunc: func(got bool) {
				s.Equal(true, got)
			},
		},
		{
			name:  "when non-alphanum non-fact value",
			field: "eth-0!",
			validateFunc: func(got bool) {
				s.Equal(false, got)
			},
		},
		{
			name:  "when empty value",
			field: "",
			validateFunc: func(got bool) {
				s.Equal(false, got)
			},
		},
		{
			name:  "when partial fact prefix",
			field: "@fact",
			validateFunc: func(got bool) {
				s.Equal(false, got)
			},
		},
		{
			name:  "when at-sign without fact",
			field: "@notfact.x",
			validateFunc: func(got bool) {
				s.Equal(false, got)
			},
		},
		{
			name:  "when unknown fact key",
			field: "@fact.primary_interface",
			validateFunc: func(got bool) {
				s.Equal(false, got)
			},
		},
		{
			name:  "when fact with bare custom prefix",
			field: "@fact.custom.",
			validateFunc: func(got bool) {
				s.Equal(false, got)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			_, ok := validation.Var(context.Background(), tt.field, "required,alphanum_or_fact")
			tt.validateFunc(ok)
		})
	}
}

func (s *ValidationPublicTestSuite) TestIpOrFact() {
	tests := []struct {
		name         string
		field        string
		validateFunc func(bool)
	}{
		{
			name:  "when valid IPv4",
			field: "1.1.1.1",
			validateFunc: func(got bool) {
				s.Equal(true, got)
			},
		},
		{
			name:  "when valid IPv6",
			field: "::1",
			validateFunc: func(got bool) {
				s.Equal(true, got)
			},
		},
		{
			name:  "when fact reference",
			field: "@fact.custom.gateway",
			validateFunc: func(got bool) {
				s.Equal(true, got)
			},
		},
		{
			name:  "when fact interface primary",
			field: "@fact.interface.primary",
			validateFunc: func(got bool) {
				s.Equal(true, got)
			},
		},
		{
			name:  "when invalid address",
			field: "not-an-ip",
			validateFunc: func(got bool) {
				s.Equal(false, got)
			},
		},
		{
			name:  "when empty value",
			field: "",
			validateFunc: func(got bool) {
				s.Equal(false, got)
			},
		},
		{
			name:  "when partial fact prefix",
			field: "@fact",
			validateFunc: func(got bool) {
				s.Equal(false, got)
			},
		},
		{
			name:  "when at-sign without fact",
			field: "@notfact.x",
			validateFunc: func(got bool) {
				s.Equal(false, got)
			},
		},
		{
			name:  "when unknown fact key",
			field: "@fact.primary_interface",
			validateFunc: func(got bool) {
				s.Equal(false, got)
			},
		},
		{
			name:  "when fact with bare custom prefix",
			field: "@fact.custom.",
			validateFunc: func(got bool) {
				s.Equal(false, got)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			_, ok := validation.Var(context.Background(), tt.field, "required,ip_or_fact")
			tt.validateFunc(ok)
		})
	}
}

func (s *ValidationPublicTestSuite) TestSchedule() {
	tests := []struct {
		name         string
		field        string
		contains     []string
		validateFunc func(string, bool)
	}{
		// Valid expressions
		{
			name:  "when every minute",
			field: "* * * * *",
			validateFunc: func(_ string, ok bool) {
				s.True(ok)
			},
		},
		{
			name:  "when daily at 2am",
			field: "0 2 * * *",
			validateFunc: func(_ string, ok bool) {
				s.True(ok)
			},
		},
		{
			name:  "when every 5 minutes",
			field: "*/5 * * * *",
			validateFunc: func(_ string, ok bool) {
				s.True(ok)
			},
		},
		{
			name:  "when weekdays at 9am",
			field: "0 9 * * 1-5",
			validateFunc: func(_ string, ok bool) {
				s.True(ok)
			},
		},
		{
			name:  "when first of month at midnight",
			field: "0 0 1 * *",
			validateFunc: func(_ string, ok bool) {
				s.True(ok)
			},
		},
		{
			name:  "when multiple hours",
			field: "0 2,14 * * *",
			validateFunc: func(_ string, ok bool) {
				s.True(ok)
			},
		},
		{
			name:  "when range with step",
			field: "0-30/5 * * * *",
			validateFunc: func(_ string, ok bool) {
				s.True(ok)
			},
		},
		{
			name:  "when month and day names",
			field: "0 0 * jan-mar mon",
			validateFunc: func(_ string, ok bool) {
				s.True(ok)
			},
		},
		// Invalid expressions
		{
			name:  "when empty string",
			field: "",
			validateFunc: func(_ string, ok bool) {
				s.False(ok)
			},
		},
		{
			name:  "when random text",
			field: "not a cron expression",
			validateFunc: func(_ string, ok bool) {
				s.False(ok)
			},
		},
		{
			name:  "when too few fields",
			field: "* * *",
			validateFunc: func(_ string, ok bool) {
				s.False(ok)
			},
		},
		{
			name:  "when too many fields (6 fields)",
			field: "* * * * * *",
			validateFunc: func(_ string, ok bool) {
				s.False(ok)
			},
		},
		{
			name:  "when minute out of range",
			field: "60 * * * *",
			validateFunc: func(_ string, ok bool) {
				s.False(ok)
			},
		},
		{
			name:  "when hour out of range",
			field: "0 25 * * *",
			validateFunc: func(_ string, ok bool) {
				s.False(ok)
			},
		},
		{
			name:  "when day of month out of range",
			field: "0 0 32 * *",
			validateFunc: func(_ string, ok bool) {
				s.False(ok)
			},
		},
		{
			name:  "when month out of range",
			field: "0 0 * 13 *",
			validateFunc: func(_ string, ok bool) {
				s.False(ok)
			},
		},
		{
			name:  "when day of week out of range",
			field: "0 0 * * 8",
			validateFunc: func(_ string, ok bool) {
				s.False(ok)
			},
		},
		{
			name:  "when invalid character",
			field: "0 0 * * abc",
			validateFunc: func(_ string, ok bool) {
				s.False(ok)
			},
		},
		{
			name:  "when invalid expression shows hint in struct validation",
			field: "bad",
			contains: []string{
				"cron_schedule",
				"is not a valid cron expression",
			},
			validateFunc: func(errMsg string, ok bool) {
				s.False(ok)
				s.Contains(errMsg, "cron_schedule")
				s.Contains(errMsg, "is not a valid cron expression")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if len(tt.contains) > 0 {
				// Test through Struct() to verify hint formatting.
				type cronReq struct {
					Schedule string `validate:"required,cron_schedule"`
				}

				tt.validateFunc(
					validation.Struct(context.Background(), cronReq{Schedule: tt.field}),
				)
			} else {
				tt.validateFunc(validation.Var(context.Background(), tt.field, "cron_schedule"))
			}
		})
	}
}

func (s *ValidationPublicTestSuite) TestGoDuration() {
	tests := []struct {
		name         string
		field        string
		contains     []string
		validateFunc func(string, bool)
	}{
		{
			name:  "when valid duration 30s passes",
			field: "30s",
			validateFunc: func(_ string, ok bool) {
				s.True(ok)
			},
		},
		{
			name:  "when valid duration 5m passes",
			field: "5m",
			validateFunc: func(_ string, ok bool) {
				s.True(ok)
			},
		},
		{
			name:  "when valid duration 1h passes",
			field: "1h",
			validateFunc: func(_ string, ok bool) {
				s.True(ok)
			},
		},
		{
			name:  "when valid duration 720h passes",
			field: "720h",
			validateFunc: func(_ string, ok bool) {
				s.True(ok)
			},
		},
		{
			name:  "when invalid duration 7d fails",
			field: "7d",
			validateFunc: func(_ string, ok bool) {
				s.False(ok)
			},
		},
		{
			name:  "when invalid duration abc fails",
			field: "abc",
			validateFunc: func(_ string, ok bool) {
				s.False(ok)
			},
		},
		{
			name:  "when empty string passes with omitempty",
			field: "",
			validateFunc: func(_ string, ok bool) {
				s.True(ok)
			},
		},
		{
			name:  "when invalid duration shows hint in struct validation",
			field: "7d",
			contains: []string{
				"go_duration",
				"is not a valid Go duration",
			},
			validateFunc: func(errMsg string, ok bool) {
				s.False(ok)
				s.Contains(errMsg, "go_duration")
				s.Contains(errMsg, "is not a valid Go duration")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if len(tt.contains) > 0 {
				// Test through Struct() to verify hint formatting.
				type durationReq struct {
					MaxAge string `validate:"required,go_duration"`
				}

				tt.validateFunc(
					validation.Struct(context.Background(), durationReq{MaxAge: tt.field}),
				)
			} else if tt.field == "" {
				// Test empty string with omitempty — validation should pass.
				tt.validateFunc(validation.Var(context.Background(), tt.field, "omitempty,go_duration"))
			} else {
				tt.validateFunc(validation.Var(context.Background(), tt.field, "go_duration"))
			}
		})
	}
}

func (s *ValidationPublicTestSuite) TestSysctlKey() {
	tests := []struct {
		name         string
		field        string
		validateFunc func(bool)
	}{
		{
			name:  "when dotted key",
			field: "net.ipv4.ip_forward",
			validateFunc: func(got bool) {
				s.True(got)
			},
		},
		{
			name:  "when key with hyphen and underscore",
			field: "some-key_name.value",
			validateFunc: func(got bool) {
				s.True(got)
			},
		},
		{
			name:  "when path traversal key",
			field: "../../etc/cron.d/pwn",
			validateFunc: func(got bool) {
				s.False(got)
			},
		},
		{
			name:  "when key contains slash",
			field: "a/b",
			validateFunc: func(got bool) {
				s.False(got)
			},
		},
		{
			name:  "when key contains space",
			field: "net.ipv4 ip_forward",
			validateFunc: func(got bool) {
				s.False(got)
			},
		},
		{
			name:  "when key contains equals sign",
			field: "net.ipv4.ip_forward=1",
			validateFunc: func(got bool) {
				s.False(got)
			},
		},
		{
			name:  "when key starts with a hyphen",
			field: "-pfoo",
			validateFunc: func(got bool) {
				s.False(got)
			},
		},
		{
			name:  "when key starts with a dot",
			field: ".hidden",
			validateFunc: func(got bool) {
				s.False(got)
			},
		},
		{
			name:  "when empty string",
			field: "",
			validateFunc: func(got bool) {
				s.False(got)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			_, ok := validation.Var(context.Background(), tt.field, "sysctl_key")
			tt.validateFunc(ok)
		})
	}

	s.Run("invalid key shows hint through struct validation", func() {
		type keyReq struct {
			Key string `validate:"required,sysctl_key"`
		}

		errMsg, ok := validation.Struct(context.Background(), keyReq{Key: "a/b"})
		s.False(ok)
		s.Contains(errMsg, "sysctl_key")
		s.Contains(errMsg, "not a valid sysctl key")
	})
}

func (s *ValidationPublicTestSuite) TestResourceName() {
	tests := []struct {
		name  string
		field string
		valid bool
	}{
		{name: "when alphanumeric", field: "backup", valid: true},
		{name: "when it has hyphens and underscores", field: "nightly_db-backup", valid: true},
		{name: "when it has digits", field: "backup2", valid: true},
		{name: "when it is a traversal", field: "../../etc/passwd", valid: false},
		{name: "when it has a slash", field: "a/b", valid: false},
		{name: "when it has a dot", field: "backup.sh", valid: false},
		{name: "when it has a space", field: "my backup", valid: false},
		{name: "when empty", field: "", valid: false},
		// A leading hyphen is accepted, matching the providers' own validName.
		// The name only ever becomes a path segment, never a command argument,
		// so there is no option to be mistaken for.
		{name: "when it starts with a hyphen", field: "-rf", valid: true},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			_, ok := validation.Var(context.Background(), tt.field, "resource_name")
			s.Equal(tt.valid, ok)
		})
	}

	s.Run("invalid name shows hint through struct validation", func() {
		type nameReq struct {
			Name string `validate:"required,resource_name"`
		}

		errMsg, ok := validation.Struct(context.Background(), nameReq{Name: "../etc/passwd"})
		s.False(ok)
		s.Contains(errMsg, "resource_name")
		s.Contains(errMsg, "not a valid name")
	})
}

func (s *ValidationPublicTestSuite) TestNATSName() {
	tests := []struct {
		name  string
		field string
		valid bool
	}{
		{
			name:  "when it has letters, digits, hyphens and underscores",
			field: "agent-registry_2",
			valid: true,
		},
		{name: "when it has a dot", field: "agent.registry", valid: false},
		{name: "when it has a space", field: "agent registry", valid: false},
		{name: "when it has a wildcard", field: "agent-*", valid: false},
		{name: "when it has a slash", field: "agent/registry", valid: false},
		{name: "when empty", field: "", valid: false},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			_, ok := validation.Var(context.Background(), tt.field, "nats_name")
			s.Equal(tt.valid, ok)
		})
	}

	s.Run("an invalid name shows the hint", func() {
		errMsg, ok := validation.Var(context.Background(), "bad name", "nats_name")
		s.False(ok)
		s.Contains(errMsg, "not a valid NATS bucket name")
	})
}

// TestPathNames holds the rules of what a hand-written route's path names:
// a service, a systemd unit, an SSH key.
func (s *ValidationPublicTestSuite) TestPathNames() {
	tests := []struct {
		name  string
		tag   string
		field string
		valid bool
		hint  string
	}{
		{name: "when a service name is plain", tag: "service_name", field: "nginx", valid: true},
		{
			name:  "when a service name has an instance",
			tag:   "service_name",
			field: "getty@tty1",
			valid: true,
		},
		{
			name:  "when a service name has a space",
			tag:   "service_name",
			field: "nginx reboot",
			hint:  "not a valid service name",
		},
		{
			name:  "when a service name has a slash",
			tag:   "service_name",
			field: "../nginx",
			hint:  "not a valid service name",
		},
		{
			name:  "when a unit name has its type",
			tag:   "unit_name",
			field: "nginx.service",
			valid: true,
		},
		{
			name:  "when a unit name is a user's",
			tag:   "unit_name",
			field: "user@1000.service",
			valid: true,
		},
		{
			name:  "when a unit name has a semicolon",
			tag:   "unit_name",
			field: "nginx;reboot",
			hint:  "not a valid systemd unit name",
		},
		{
			name:  "when a fingerprint is SHA256",
			tag:   "ssh_fingerprint",
			field: "SHA256:fs7cRe+Lieb9g9TQ7a4HbYTDyVWnO8tXg6D9H2cAWIY",
			valid: true,
		},
		{
			name:  "when a fingerprint is MD5",
			tag:   "ssh_fingerprint",
			field: "MD5:ab:cd",
			hint:  "not a valid SSH key fingerprint",
		},
		{
			name:  "when a fingerprint has no digest",
			tag:   "ssh_fingerprint",
			field: "SHA256:",
			hint:  "not a valid SSH key fingerprint",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			msg, ok := validation.Var(context.Background(), tt.field, tt.tag)
			s.Equal(tt.valid, ok)
			if !tt.valid {
				s.Contains(msg, tt.hint)
			}
		})
	}
}

func (s *ValidationPublicTestSuite) TestPackageName() {
	tests := []struct {
		name  string
		field string
		valid bool
	}{
		{name: "when a plain name", field: "nginx", valid: true},
		{name: "when a name with a plus and a dot", field: "libstdc++6.0", valid: true},
		{name: "when a name with an architecture", field: "nginx:amd64", valid: true},
		{name: "when a name with a version", field: "nginx=1.24.0-1~deb12u1", valid: true},
		{
			name:  "when it starts with a hyphen, which reads as an option",
			field: "-oDpkg::Pre-Invoke::=x",
			valid: false,
		},
		{name: "when it starts with a dot", field: ".hidden", valid: false},
		{name: "when it has a space", field: "nginx curl", valid: false},
		{name: "when it has a slash", field: "../nginx", valid: false},
		{name: "when empty", field: "", valid: false},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			_, ok := validation.Var(context.Background(), tt.field, "package_name")
			s.Equal(tt.valid, ok)
		})
	}

	s.Run("an invalid name shows its hint through struct validation", func() {
		type pkgReq struct {
			Name string `validate:"required,package_name"`
		}

		errMsg, ok := validation.Struct(context.Background(), pkgReq{Name: "-y"})
		s.False(ok)
		s.Contains(errMsg, "not a valid package name")
	})
}

func (s *ValidationPublicTestSuite) TestContainerID() {
	tests := []struct {
		name  string
		field string
		valid bool
	}{
		{name: "when a short id", field: "a1b2c3d4e5f6", valid: true},
		{name: "when a name with a dot", field: "web.1", valid: true},
		{name: "when a name with a hyphen", field: "web-01", valid: true},
		{name: "when it starts with a hyphen", field: "-f", valid: false},
		{name: "when it starts with a dot", field: ".hidden", valid: false},
		{name: "when it has a slash", field: "a/b", valid: false},
		{name: "when empty", field: "", valid: false},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			_, ok := validation.Var(context.Background(), tt.field, "container_id")
			s.Equal(tt.valid, ok)
		})
	}

	s.Run("invalid id shows hint through struct validation", func() {
		type idReq struct {
			ID string `validate:"required,container_id"`
		}

		errMsg, ok := validation.Struct(context.Background(), idReq{ID: "-f"})
		s.False(ok)
		s.Contains(errMsg, "container_id")
		s.Contains(errMsg, "not a valid container name or ID")
	})
}

func (s *ValidationPublicTestSuite) TestNoLinebreak() {
	tests := []struct {
		name         string
		field        string
		validateFunc func(bool)
	}{
		{
			name:  "when simple value",
			field: "1",
			validateFunc: func(got bool) {
				s.True(got)
			},
		},
		{
			name:  "when value contains newline",
			field: "1\nkernel.modprobe = /tmp/evil",
			validateFunc: func(got bool) {
				s.False(got)
			},
		},
		{
			name:  "when value contains carriage return",
			field: "1\r\nkernel.modprobe = /tmp/evil",
			validateFunc: func(got bool) {
				s.False(got)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			_, ok := validation.Var(context.Background(), tt.field, "no_linebreak")
			tt.validateFunc(ok)
		})
	}

	s.Run("invalid value shows hint through struct validation", func() {
		type valueReq struct {
			Value string `validate:"required,no_linebreak"`
		}

		errMsg, ok := validation.Struct(
			context.Background(),
			valueReq{Value: "1\nkernel.modprobe = /tmp/evil"},
		)
		s.False(ok)
		s.Contains(errMsg, "no_linebreak")
		s.Contains(errMsg, "must not contain line breaks")
	})
}

func (s *ValidationPublicTestSuite) TestSigningKey() {
	tests := []struct {
		name         string
		field        string
		validateFunc func(bool)
	}{
		{
			name:  "when exactly the minimum length",
			field: strings.Repeat("a", validation.MinSigningKeyLen),
			validateFunc: func(got bool) {
				s.True(got)
			},
		},
		{
			name:  "when one character under the minimum",
			field: strings.Repeat("a", validation.MinSigningKeyLen-1),
			validateFunc: func(got bool) {
				s.False(got)
			},
		},
		{
			name:  "when empty",
			field: "",
			validateFunc: func(got bool) {
				s.False(got)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			_, ok := validation.Var(context.Background(), tt.field, "signing_key")
			tt.validateFunc(ok)
		})
	}

	s.Run("short key shows hint through struct validation", func() {
		type keyReq struct {
			SigningKey string `validate:"required,signing_key"`
		}

		errMsg, ok := validation.Struct(context.Background(), keyReq{SigningKey: "too-short"})
		s.False(ok)
		s.Contains(errMsg, "signing_key")
		s.Contains(errMsg, "openssl rand -hex 32")
	})
}

func (s *ValidationPublicTestSuite) TestAccountName() {
	tests := []struct {
		name         string
		field        string
		validateFunc func(bool)
	}{
		{
			name:  "when simple lowercase name",
			field: "john",
			validateFunc: func(got bool) {
				s.True(got)
			},
		},
		{
			name:  "when name with digits hyphen and underscore",
			field: "j0hn-doe_2",
			validateFunc: func(got bool) {
				s.True(got)
			},
		},
		{
			name:  "when machine account with trailing dollar sign",
			field: "host$",
			validateFunc: func(got bool) {
				s.True(got)
			},
		},
		{
			name:  "when exactly the maximum length",
			field: strings.Repeat("a", 32),
			validateFunc: func(got bool) {
				s.True(got)
			},
		},
		{
			name:  "when one character over the maximum length",
			field: strings.Repeat("a", 33),
			validateFunc: func(got bool) {
				s.False(got)
			},
		},
		{
			name:  "when name contains an uppercase letter",
			field: "John",
			validateFunc: func(got bool) {
				s.False(got)
			},
		},
		{
			name:  "when name starts with a digit",
			field: "1john",
			validateFunc: func(got bool) {
				s.False(got)
			},
		},
		{
			name:  "when name starts with a hyphen",
			field: "-john",
			validateFunc: func(got bool) {
				s.False(got)
			},
		},
		{
			name:  "when name contains a slash",
			field: "jo/hn",
			validateFunc: func(got bool) {
				s.False(got)
			},
		},
		{
			name:  "when name contains a space",
			field: "jo hn",
			validateFunc: func(got bool) {
				s.False(got)
			},
		},
		{
			name:  "when empty string",
			field: "",
			validateFunc: func(got bool) {
				s.False(got)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			_, ok := validation.Var(context.Background(), tt.field, "account_name")
			tt.validateFunc(ok)
		})
	}

	s.Run("invalid name shows hint through struct validation", func() {
		type nameReq struct {
			Name string `validate:"required,account_name"`
		}

		errMsg, ok := validation.Struct(context.Background(), nameReq{Name: "-bad"})
		s.False(ok)
		s.Contains(errMsg, "account_name")
		s.Contains(errMsg, "not a valid account name")
		s.Contains(errMsg, "max 32 characters")
	})
}

func (s *ValidationPublicTestSuite) TestAtLeastOneField() {
	type allPointers struct {
		Shell  *string
		Home   *string
		Groups *[]string
	}

	type withNonPointer struct {
		Name    string
		Shell   *string
		Enabled bool
	}

	type withSlice struct {
		Items []string
		Shell *string
	}

	type withMap struct {
		Labels map[string]string
		Shell  *string
	}

	type unexportedOnly struct {
		hidden *string //nolint:unused
	}

	str := "bash"
	groups := []string{"admin"}

	tests := []struct {
		name         string
		input        any
		validateFunc func(string, bool)
	}{
		{
			name:  "when one pointer field is non-nil",
			input: allPointers{Shell: &str},
			validateFunc: func(_ string, ok bool) {
				s.True(ok)
			},
		},
		{
			name:  "when slice pointer field is non-nil",
			input: allPointers{Groups: &groups},
			validateFunc: func(_ string, ok bool) {
				s.True(ok)
			},
		},
		{
			name:  "when all pointer fields are nil",
			input: allPointers{},
			validateFunc: func(errMsg string, ok bool) {
				s.False(ok)
				s.Equal("at least one field must be provided", errMsg)
			},
		},
		{
			name:  "when non-pointer field is non-zero",
			input: withNonPointer{Name: "test"},
			validateFunc: func(_ string, ok bool) {
				s.True(ok)
			},
		},
		{
			name:  "when bool field is true",
			input: withNonPointer{Enabled: true},
			validateFunc: func(_ string, ok bool) {
				s.True(ok)
			},
		},
		{
			name:  "when all fields are zero",
			input: withNonPointer{},
			validateFunc: func(errMsg string, ok bool) {
				s.False(ok)
				s.Equal("at least one field must be provided", errMsg)
			},
		},
		{
			name:  "when slice field is non-nil",
			input: withSlice{Items: []string{"a"}},
			validateFunc: func(_ string, ok bool) {
				s.True(ok)
			},
		},
		{
			name:  "when slice field is nil",
			input: withSlice{},
			validateFunc: func(errMsg string, ok bool) {
				s.False(ok)
				s.Equal("at least one field must be provided", errMsg)
			},
		},
		{
			name:  "when map field is non-nil",
			input: withMap{Labels: map[string]string{"env": "dev"}},
			validateFunc: func(_ string, ok bool) {
				s.True(ok)
			},
		},
		{
			name:  "when map field is nil",
			input: withMap{},
			validateFunc: func(errMsg string, ok bool) {
				s.False(ok)
				s.Equal("at least one field must be provided", errMsg)
			},
		},
		{
			name:  "when only unexported fields",
			input: unexportedOnly{},
			validateFunc: func(errMsg string, ok bool) {
				s.False(ok)
				s.Equal("at least one field must be provided", errMsg)
			},
		},
		{
			name:  "when pointer to struct is passed",
			input: &allPointers{Shell: &str},
			validateFunc: func(_ string, ok bool) {
				s.True(ok)
			},
		},
		{
			name:  "when non-struct is passed",
			input: "not a struct",
			validateFunc: func(errMsg string, ok bool) {
				s.False(ok)
				s.Equal("expected struct", errMsg)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.validateFunc(validation.AtLeastOneField(tt.input))
		})
	}
}

func (s *ValidationPublicTestSuite) TestInstance() {
	tests := []struct {
		name         string
		validateFunc func(*validator.Validate)
	}{
		{
			name: "when returns shared validator instance",
			validateFunc: func(v *validator.Validate) {
				s.NotNil(v)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.validateFunc(validation.Instance())
		})
	}
}

func TestValidationPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ValidationPublicTestSuite))
}

// TestFileMode covers the mode a provider can actually apply. An unparsable one
// used to become 0644 further down, granting more than was asked for.
func (s *ValidationPublicTestSuite) TestFileMode() {
	tests := []struct {
		name         string
		field        string
		contains     []string
		validateFunc func(string, bool)
	}{
		{
			name:  "four octal digits",
			field: "0644",
			validateFunc: func(_ string, ok bool) {
				s.True(ok)
			},
		},
		{
			name:  "three octal digits",
			field: "755",
			validateFunc: func(_ string, ok bool) {
				s.True(ok)
			},
		},
		{
			name:  "a digit that is not octal",
			field: "0999",
			validateFunc: func(_ string, ok bool) {
				s.False(ok)
			},
		},
		{
			name:  "not a number at all",
			field: "rwxr-xr-x",
			validateFunc: func(_ string, ok bool) {
				s.False(ok)
			},
		},
		{
			name:  "too short to be a mode",
			field: "64",
			validateFunc: func(_ string, ok bool) {
				s.False(ok)
			},
		},
		{
			name:  "too long to be a mode",
			field: "06440",
			validateFunc: func(_ string, ok bool) {
				s.False(ok)
			},
		},
		{
			name:  "empty passes with omitempty",
			field: "",
			validateFunc: func(_ string, ok bool) {
				s.True(ok)
			},
		},
		{
			name:  "an invalid mode explains itself",
			field: "rwx",
			contains: []string{
				"file_mode",
				"three or four octal digits",
			},
			validateFunc: func(errMsg string, ok bool) {
				s.False(ok)
				s.Contains(errMsg, "file_mode")
				s.Contains(errMsg, "three or four octal digits")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			switch {
			case len(tt.contains) > 0:
				type modeReq struct {
					Mode string `validate:"required,file_mode"`
				}

				tt.validateFunc(validation.Struct(context.Background(), modeReq{Mode: tt.field}))
			case tt.field == "":
				tt.validateFunc(
					validation.Var(context.Background(), tt.field, "omitempty,file_mode"),
				)
			default:
				tt.validateFunc(validation.Var(context.Background(), tt.field, "file_mode"))
			}
		})
	}
}
