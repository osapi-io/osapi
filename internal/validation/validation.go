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

// Package validation provides a shared validator instance.
package validation

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/osapi-io/osapi/internal/validation/rule"

	"github.com/go-playground/validator/v10"
	cronparser "github.com/robfig/cron/v3"

	"github.com/osapi-io/osapi/internal/facts"
)

var instance = validator.New()

// MinSigningKeyLen is the minimum character length required for a JWT
// signing key, matching the output of `openssl rand -hex 32`. A short key
// makes brute-forcing the HS256 secret and forging tokens practical.
const MinSigningKeyLen = 32

// interfaceNamePattern is a Linux interface name: up to 15 characters, such
// as eth0, br-lan or eth0.100. A name never starts with a dot or a dash, so it
// never leaves a directory in a file name.
var interfaceNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,14}$`)

// TimezonePattern is an IANA timezone name, such as UTC or
// America/Argentina/Buenos_Aires: segments of letters, digits, _, + and -
// joined by /, never starting with - and with no dot. It becomes an
// argument of timedatectl, so the provider checks it again.
// PackageVersionPattern is a Debian package version, such as 8.5.0-2ubuntu10
// or 1:2.3: a digit first, then letters, digits and . + ~ : -. It becomes
// part of an apt-get argument, so the provider checks it again.
var PackageVersionPattern = regexp.MustCompile(`^[0-9][A-Za-z0-9.+~:-]*$`)

// itemIndex is the index of a list item in a field's name, such as [0].
var itemIndex = regexp.MustCompile(`\[\d+\]`)

var TimezonePattern = regexp.MustCompile(`^[A-Za-z0-9_+][A-Za-z0-9_+-]*(/[A-Za-z0-9_+-]+)*$`)

// accountNamePattern matches a POSIX system account name: it must start with
// a lowercase letter or underscore, followed by lowercase letters, digits,
// underscores, or hyphens, with an optional trailing dollar sign (used by
// machine accounts). This is the Debian `adduser` default (NAME_REGEX)
// without --allow-bad-names, and it rejects a leading hyphen, so the name
// cannot be mistaken for an option when passed as a command-line argument
// to useradd, usermod, userdel, groupadd, groupmod, groupdel, or gpasswd.
var accountNamePattern = regexp.MustCompile(`^[a-z_][a-z0-9_-]*\$?$`)

// accountNameMaxLength is the longest account name accepted, matching the
// historical utmp/wtmp field width Debian tools enforce by default.
const accountNameMaxLength = 32

// resourceNamePattern matches a name osapi uses as a file name: letters,
// digits, underscores, and hyphens only. It mirrors the validName regex the
// schedule and certificate providers enforce, so a name the provider will
// reject is rejected on the request path instead of becoming a job. Slashes
// and dots are rejected, so the name cannot escape the directory it is
// written into.
var resourceNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// natsNamePattern matches a name NATS accepts for a KV or Object Store
// bucket: letters, digits, underscores, and hyphens. It is the rule nats.go
// enforces when a bucket is made, so a bad name in the config fails when
// the config loads, not at the first call that uses the bucket. A namespace
// is held to it too, because it becomes the prefix of each bucket name.
var natsNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// containerIDPattern matches a container name or ID as the container runtime
// accepts it. The first character must be a letter or digit so the value
// cannot be mistaken for a command-line option.
var containerIDPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]*$`)

// imageRefPattern matches an image reference as docker takes it, such as
// alpine:3.20 or ghcr.io/org/app@sha256:..., starting with a letter or
// digit so the value cannot be mistaken for an option of the docker CLI.
var imageRefPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._/:@-]*$`)

// dnsNamePattern matches a hostname as RFC 1123 has it: letters, digits,
// dots and dashes, starting with a letter or digit.
var dnsNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9.-]*$`)

// portMappingPattern matches a published port, host:container, such as
// 8080:80.
var portMappingPattern = regexp.MustCompile(`^[0-9]{1,5}:[0-9]{1,5}$`)

// volumeMappingPattern matches a bind mount, host:container, each an
// absolute path, such as /srv/data:/data.
var volumeMappingPattern = regexp.MustCompile(`^/[^:]*:/[^:]*$`)

// serviceNamePattern matches a service name as the service provider takes
// it, which it checks again before systemctl runs.
var serviceNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_@.-]+$`)

// unitNamePattern matches a systemd unit name: letters, digits and :_.@-,
// with its type, such as nginx.service or user@1000.service.
var unitNamePattern = regexp.MustCompile(`^[a-zA-Z0-9:_.@-]+$`)

// sshFingerprintPattern matches an SSH key fingerprint as the user provider
// computes it: SHA256: and the base64 of the digest.
var sshFingerprintPattern = regexp.MustCompile(`^SHA256:[A-Za-z0-9+/]+=*$`)

// PackageNamePattern matches a package name as apt and dpkg take it, with an
// architecture (:amd64) or a version (=1.2-1). The first character is a
// letter or digit, so the package manager cannot read the name as an option.
var PackageNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9.+_:~=-]*$`)

func init() {
	// valid_target is registered here, so a struct that has it validates
	// before the controller starts, as the validate command needs. With no
	// AgentLister it passes only _any and _all; RegisterTargetValidator gives
	// it the agents.
	_ = instance.RegisterValidationCtx("valid_target", validTarget)

	// alphanum_or_fact accepts alphanumeric values or @fact. prefixed references
	// with a known fact key. Fact references are resolved agent-side.
	_ = instance.RegisterValidation("alphanum_or_fact", func(fl validator.FieldLevel) bool {
		v := fl.Field().String()
		if strings.HasPrefix(v, facts.Prefix) {
			return facts.IsKnownKey(v[len(facts.Prefix):])
		}
		return instance.Var(v, "alphanum") == nil
	})

	// interface_name accepts a Linux interface name, or an @fact. reference
	// with a known fact key, which the agent resolves to one.
	_ = instance.RegisterValidation("interface_name", func(fl validator.FieldLevel) bool {
		v := fl.Field().String()
		if strings.HasPrefix(v, facts.Prefix) {
			return facts.IsKnownKey(v[len(facts.Prefix):])
		}
		return interfaceNamePattern.MatchString(v)
	})

	// ip_or_fact accepts IP addresses (v4/v6) or @fact. prefixed references
	// with a known fact key. Fact references are resolved agent-side.
	_ = instance.RegisterValidation("ip_or_fact", func(fl validator.FieldLevel) bool {
		v := fl.Field().String()
		if strings.HasPrefix(v, facts.Prefix) {
			return facts.IsKnownKey(v[len(facts.Prefix):])
		}
		return instance.Var(v, "ip") == nil
	})

	// go_duration validates a Go time.Duration string (e.g., "30s", "5m", "1h", "720h").
	_ = instance.RegisterValidation("go_duration", func(fl validator.FieldLevel) bool {
		_, err := time.ParseDuration(fl.Field().String())
		return err == nil
	})

	// file_mode validates a file permission mode as three or four octal digits,
	// which is what a provider can actually apply. An unparsable mode used to
	// become 0644 further down, quietly granting more than was asked for.
	_ = instance.RegisterValidation("file_mode", func(fl validator.FieldLevel) bool {
		mode := fl.Field().String()
		if len(mode) < 3 || len(mode) > 4 {
			return false
		}

		_, err := strconv.ParseUint(mode, 8, 32)

		return err == nil
	})

	// cron_schedule validates a standard 5-field cron expression
	// (minute hour day-of-month month day-of-week).
	cronParser := cronparser.NewParser(
		cronparser.Minute | cronparser.Hour | cronparser.Dom | cronparser.Month | cronparser.Dow,
	)
	_ = instance.RegisterValidation("cron_schedule", func(fl validator.FieldLevel) bool {
		_, err := cronParser.Parse(fl.Field().String())
		return err == nil
	})

	// package_version validates a package version; PackageVersionPattern
	// says what it permits.
	_ = instance.RegisterValidation("package_version", func(fl validator.FieldLevel) bool {
		return PackageVersionPattern.MatchString(fl.Field().String())
	})

	// timezone_name validates an IANA timezone name; TimezonePattern says
	// what it permits.
	_ = instance.RegisterValidation("timezone_name", func(fl validator.FieldLevel) bool {
		return TimezonePattern.MatchString(fl.Field().String())
	})

	// resource_name validates a name osapi writes as a file name, rejecting
	// path separators, dots, and anything else that could escape the
	// directory. The same pattern is enforced again in the provider, which is
	// the contract; this one exists so the caller is told their input was
	// wrong rather than told the host failed.
	_ = instance.RegisterValidation("resource_name", func(fl validator.FieldLevel) bool {
		return resourceNamePattern.MatchString(fl.Field().String())
	})

	// nats_name validates a NATS bucket name or the namespace that prefixes
	// one.
	_ = instance.RegisterValidation("nats_name", func(fl validator.FieldLevel) bool {
		return natsNamePattern.MatchString(fl.Field().String())
	})

	// container_id validates a container name or ID.
	_ = instance.RegisterValidation("container_id", func(fl validator.FieldLevel) bool {
		return containerIDPattern.MatchString(fl.Field().String())
	})
	_ = instance.RegisterValidation("image_ref", func(fl validator.FieldLevel) bool {
		return imageRefPattern.MatchString(fl.Field().String())
	})
	_ = instance.RegisterValidation("dns_name", func(fl validator.FieldLevel) bool {
		return dnsNamePattern.MatchString(fl.Field().String())
	})
	_ = instance.RegisterValidation("port_mapping", func(fl validator.FieldLevel) bool {
		return portMappingPattern.MatchString(fl.Field().String())
	})
	_ = instance.RegisterValidation("volume_mapping", func(fl validator.FieldLevel) bool {
		return volumeMappingPattern.MatchString(fl.Field().String())
	})

	// service_name, unit_name and ssh_fingerprint validate what a path
	// names: a service, a systemd unit, an SSH key.
	_ = instance.RegisterValidation("service_name", func(fl validator.FieldLevel) bool {
		return serviceNamePattern.MatchString(fl.Field().String())
	})
	_ = instance.RegisterValidation("unit_name", func(fl validator.FieldLevel) bool {
		return unitNamePattern.MatchString(fl.Field().String())
	})
	_ = instance.RegisterValidation("ssh_fingerprint", func(fl validator.FieldLevel) bool {
		return sshFingerprintPattern.MatchString(fl.Field().String())
	})

	// package_name validates a package name, which reaches the package
	// manager's command line as root.
	_ = instance.RegisterValidation("package_name", func(fl validator.FieldLevel) bool {
		return PackageNamePattern.MatchString(fl.Field().String())
	})

	// no_linebreak rejects values containing a carriage return or line feed,
	// which would let the value inject additional lines into a config file.
	_ = instance.RegisterValidation("no_linebreak", func(fl validator.FieldLevel) bool {
		return !strings.ContainsAny(fl.Field().String(), "\r\n")
	})

	// The JSON Schema each custom validator means, so a behavior's generated
	// spec validates what the server does. A tag without one fails schema
	// generation rather than being dropped.
	// required_without and excluded_with tie one field to another, which the
	// schema of one field cannot say. The server checks them, and an example's
	// .Fails proves each.
	rule.Register("required_without", rule.None)
	rule.Register("excluded_with", rule.None)
	rule.Register(
		"interface_name",
		rule.Pattern(
			`^(@fact\.[a-z0-9_.]+|`+interfaceNamePattern.String()[1:len(interfaceNamePattern.String())-1]+`)$`,
		),
	)
	rule.Register("timezone_name", rule.Pattern(TimezonePattern.String()))
	rule.Register("package_version", rule.Pattern(PackageVersionPattern.String()))
	rule.Register("no_linebreak", rule.Pattern(`^[^\r\n]*$`))
	rule.Register("file_mode", rule.Pattern(`^[0-7]{3,4}$`))
	// base64 is the validator's own rule, standard base64 with padding, which
	// OpenAPI names the format byte.
	rule.Register("base64", rule.Format("byte"))
	rule.Register("resource_name", rule.Pattern(resourceNamePattern.String()))
	rule.Register("nats_name", rule.Pattern(natsNamePattern.String()))
	rule.Register("container_id", rule.Pattern(containerIDPattern.String()))
	rule.Register("image_ref", rule.Pattern(imageRefPattern.String()))
	rule.Register("dns_name", rule.Pattern(dnsNamePattern.String()))
	rule.Register("port_mapping", rule.Pattern(portMappingPattern.String()))
	rule.Register("volume_mapping", rule.Pattern(volumeMappingPattern.String()))
	rule.Register("package_name", rule.Pattern(PackageNamePattern.String()))
	rule.Register("service_name", rule.Pattern(serviceNamePattern.String()))
	rule.Register("unit_name", rule.Pattern(unitNamePattern.String()))
	rule.Register("ssh_fingerprint", rule.Pattern(sshFingerprintPattern.String()))
	rule.Register("account_name", rule.Pattern(accountNamePattern.String()))
	// A pattern cannot say what these check: a target the controller
	// resolves, a known fact key, a duration or a cron expression that a
	// parser reads, or a length that is config. The spec shows the type.
	for _, tag := range []string{
		"valid_target", "alphanum_or_fact", "ip_or_fact", "go_duration",
		"cron_schedule", "signing_key",
	} {
		rule.Register(tag, rule.None)
	}

	// An author can declare a rule with rule.Declare, as behavior.Validator
	// does; each one
	// is registered here, whether its behavior was defined before this ran
	// or after.
	rule.OnDeclare(registerRule)

	// account_name validates a system user or group name: it rejects a
	// leading hyphen, path separators, colons, whitespace, and control
	// characters, so the value cannot be mistaken for a command-line option
	// or used for path traversal when passed to a user/group provider.
	_ = instance.RegisterValidation("account_name", func(fl validator.FieldLevel) bool {
		v := fl.Field().String()
		return len(v) <= accountNameMaxLength && accountNamePattern.MatchString(v)
	})

	// enum accepts a value of a closed set that its type knows: Valid of the
	// type, such as types.ContentType, says. rule.Apply puts the type's
	// Values in the spec.
	_ = instance.RegisterValidation(rule.Enum, func(fl validator.FieldLevel) bool {
		v, ok := fl.Field().Interface().(interface{ Valid() bool })

		return ok && v.Valid()
	})

	// signing_key rejects a JWT signing key shorter than MinSigningKeyLen,
	// so a weak or placeholder value is caught at startup instead of
	// silently signing tokens with a brute-forceable secret.
	_ = instance.RegisterValidation("signing_key", func(fl validator.FieldLevel) bool {
		return len(fl.Field().String()) >= MinSigningKeyLen
	})
}

// customHints maps validator tags to a hint appended to the default error.
// hintsMu guards customHints, which a behavior's rule adds to.
var hintsMu sync.RWMutex

var customHints = map[string]func(ctx context.Context, fe validator.FieldError) string{
	rule.Enum: func(_ context.Context, fe validator.FieldError) string {
		values, _ := rule.Values(fe.Type())

		return fmt.Sprintf("%q is not one of %s", fe.Value(), strings.Join(values, ", "))
	},
	"valid_target": func(ctx context.Context, fe validator.FieldError) string {
		target, _ := fe.Value().(string)
		if t, ok := PendingTarget(ctx, target); ok {
			if t == types.TargetAll || t == types.TargetAny {
				return "all agents are pending PKI enrollment; list them with: osapi client admission list --state pending"
			}

			return fmt.Sprintf(
				"agent %q is pending PKI enrollment; admit it with: osapi client admission admit",
				t,
			)
		}

		if t, ok := UnverifiedTarget(ctx, target); ok {
			return fmt.Sprintf(
				"agent %q is registered, but its registration is not signed by the key that "+
					"the controller admitted, so no work goes to it; its condition KeyMismatch "+
					"says why: osapi client agent get --hostname %s",
				t, t,
			)
		}

		return fmt.Sprintf("target agent %q not found", fe.Value())
	},
	"go_duration": func(_ context.Context, fe validator.FieldError) string {
		return fmt.Sprintf(
			"%q is not a valid Go duration (e.g., 30s, 5m, 1h)",
			fe.Value(),
		)
	},
	"file_mode": func(_ context.Context, fe validator.FieldError) string {
		return fmt.Sprintf(
			"%q is not a valid file mode (expected three or four octal digits, e.g., 0644)",
			fe.Value(),
		)
	},
	"cron_schedule": func(_ context.Context, fe validator.FieldError) string {
		return fmt.Sprintf(
			"%q is not a valid cron expression (expected: minute hour day-of-month month day-of-week)",
			fe.Value(),
		)
	},
	"package_version": func(_ context.Context, fe validator.FieldError) string {
		return fmt.Sprintf("%q is not a valid package version, such as 8.5.0-2ubuntu10", fe.Value())
	},
	"timezone_name": func(_ context.Context, fe validator.FieldError) string {
		return fmt.Sprintf(
			"%q is not a valid timezone name, such as UTC or America/New_York",
			fe.Value(),
		)
	},
	"resource_name": func(_ context.Context, fe validator.FieldError) string {
		return fmt.Sprintf(
			"%q is not a valid name (expected: %s)",
			fe.Value(),
			resourceNamePattern.String(),
		)
	},
	"nats_name": func(_ context.Context, fe validator.FieldError) string {
		return fmt.Sprintf(
			"%q is not a valid NATS bucket name (expected: %s)",
			fe.Value(),
			natsNamePattern.String(),
		)
	},
	"service_name": func(_ context.Context, fe validator.FieldError) string {
		return fmt.Sprintf("%q is not a valid service name (expected: %s)",
			fe.Value(), serviceNamePattern.String())
	},
	"unit_name": func(_ context.Context, fe validator.FieldError) string {
		return fmt.Sprintf("%q is not a valid systemd unit name (expected: %s)",
			fe.Value(), unitNamePattern.String())
	},
	"ssh_fingerprint": func(_ context.Context, fe validator.FieldError) string {
		return fmt.Sprintf("%q is not a valid SSH key fingerprint (expected: %s)",
			fe.Value(), sshFingerprintPattern.String())
	},
	"package_name": func(_ context.Context, fe validator.FieldError) string {
		return fmt.Sprintf(
			"%q is not a valid package name (expected: %s)",
			fe.Value(),
			PackageNamePattern.String(),
		)
	},
	"image_ref": func(_ context.Context, fe validator.FieldError) string {
		return fmt.Sprintf("%q is not a valid image reference (expected: %s)",
			fe.Value(), imageRefPattern.String())
	},
	"dns_name": func(_ context.Context, fe validator.FieldError) string {
		return fmt.Sprintf("%q is not a valid hostname (expected: %s)",
			fe.Value(), dnsNamePattern.String())
	},
	"port_mapping": func(_ context.Context, fe validator.FieldError) string {
		return fmt.Sprintf("%q is not a port mapping host:container (expected: %s)",
			fe.Value(), portMappingPattern.String())
	},
	"volume_mapping": func(_ context.Context, fe validator.FieldError) string {
		return fmt.Sprintf("%q is not a bind mount host:container of absolute paths (expected: %s)",
			fe.Value(), volumeMappingPattern.String())
	},
	"container_id": func(_ context.Context, fe validator.FieldError) string {
		return fmt.Sprintf(
			"%q is not a valid container name or ID (expected: %s)",
			fe.Value(),
			containerIDPattern.String(),
		)
	},
	"no_linebreak": func(_ context.Context, _ validator.FieldError) string {
		return "value must not contain line breaks"
	},
	"account_name": func(_ context.Context, fe validator.FieldError) string {
		return fmt.Sprintf(
			"%q is not a valid account name (expected: %s, max %d characters)",
			fe.Value(),
			accountNamePattern.String(),
			accountNameMaxLength,
		)
	},
	"signing_key": func(_ context.Context, _ validator.FieldError) string {
		return fmt.Sprintf(
			"signing key must be at least %d characters — generate one with: openssl rand -hex 32",
			MinSigningKeyLen,
		)
	},
}

// Struct validates a struct and returns the error message and false if
// invalid. ctx is the request's, so a rule that reads NATS, such as
// valid_target, does it inside the request's trace and stops with it.
func Struct(
	ctx context.Context,
	v any,
) (string, bool) {
	if err := instance.StructCtx(ctx, v); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		return formatErrors(ctx, validationErrors), false
	}

	return "", true
}

// Var validates a single variable against a tag string and returns the error
// message and false if invalid. This is useful for path parameters that
// oapi-codegen does not generate validate tags for in strict-server mode.
func Var(
	ctx context.Context,
	field any,
	tag string,
) (string, bool) {
	if err := instance.VarCtx(ctx, field, tag); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		return formatErrors(ctx, validationErrors), false
	}

	return "", true
}

// formatErrors builds the error string, appending a custom hint for known
// tags while keeping the standard validator prefix.
func formatErrors(
	ctx context.Context,
	errs validator.ValidationErrors,
) string {
	msgs := make([]string, 0, len(errs))
	for _, fe := range errs {
		msg := fe.Error()
		if hint := Hint(ctx, fe); hint != "" {
			msg = fmt.Sprintf("%s: %s", msg, hint)
		}
		msgs = append(msgs, msg)
	}

	return strings.Join(msgs, "; ")
}

// Hint is what a failure of a custom validate tag means, such as what a
// signing key must look like, or "" for a tag without one. ctx is the
// request's: the valid_target hint looks up whether the agent is pending.
func Hint(
	ctx context.Context,
	fe validator.FieldError,
) string {
	hintsMu.RLock()
	fn, ok := customHints[fe.Tag()]
	hintsMu.RUnlock()
	if !ok {
		return ""
	}

	return fn(ctx, fe)
}

// AtLeastOneField checks that at least one exported pointer, slice, or map
// field in the struct is non-nil, or that at least one non-pointer field is
// non-zero. Returns an error message and false if all fields are nil/zero
// (i.e., the update body is empty).
func AtLeastOneField(
	v any,
) (string, bool) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return "expected struct", false
	}

	rt := rv.Type()
	for i := range rt.NumField() {
		field := rv.Field(i)
		if !rt.Field(i).IsExported() {
			continue
		}

		switch field.Kind() {
		case reflect.Pointer, reflect.Slice, reflect.Map:
			if !field.IsNil() {
				return "", true
			}
		default:
			if !field.IsZero() {
				return "", true
			}
		}
	}

	return "at least one field must be provided", false
}

// Instance returns the shared validator for registering custom validators.
func Instance() *validator.Validate {
	return instance
}

// registerRule adds a rule a behavior declared to the validator, and its
// message to the hints a failure reads.
func registerRule(
	r rule.Custom,
) {
	_ = instance.RegisterValidation(r.Tag, func(fl validator.FieldLevel) bool {
		return r.Check(fl.Field().String())
	})
	hintsMu.Lock()
	defer hintsMu.Unlock()
	customHints[r.Tag] = func(_ context.Context, fe validator.FieldError) string {
		return fmt.Sprintf("%q %s", fe.Value(), r.Message)
	}
}

// Breaks is each rule that in breaks, as "<field>:<tag>", with the field as
// its Go path below in, such as "Name:max" or "Body.Target:valid_target". A
// method's declared bad inputs are checked against it, so each rule of an
// input has an input that breaks it.
func Breaks(
	in any,
) (out []string) {
	// A tag the validator has no function for panics. The spec check of
	// each DSL reports that tag, so here it breaks nothing.
	defer func() {
		if recover() != nil {
			out = nil
		}
	}()

	err := instance.Struct(in)
	if err == nil {
		return nil
	}
	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		return nil
	}

	for _, fe := range errs {
		_, field, _ := strings.Cut(fe.StructNamespace(), ".")
		// A rule after dive breaks an item, Servers[0]: it is the rule of
		// the field, which a refusal names.
		out = append(out, itemIndex.ReplaceAllString(field, "")+":"+fe.Tag())
	}

	return out
}
