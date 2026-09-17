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

package user

import (
	"fmt"
	"regexp"
)

// accountNamePattern matches a POSIX system account name: it must start
// with a lowercase letter or underscore, followed by lowercase letters,
// digits, underscores, or hyphens, with an optional trailing dollar sign
// (used by machine accounts). This is the Debian `adduser` default
// (NAME_REGEX) without --allow-bad-names, and it rejects a leading
// hyphen, so the name cannot be mistaken for an option when it reaches
// useradd, usermod, userdel, groupadd, groupmod, groupdel, or gpasswd.
//
// This mirrors the "account_name" validator in internal/validation. It is
// duplicated rather than imported so this defense holds even if a caller
// reaches the provider without going through API validation, and so a
// change to the API-layer validator does not silently weaken this one.
var accountNamePattern = regexp.MustCompile(`^[a-z_][a-z0-9_-]*\$?$`)

// accountNameMaxLength is the longest account name accepted, matching the
// historical utmp/wtmp field width Debian tools enforce by default.
const accountNameMaxLength = 32

// validateAccountName rejects a user or group name that does not match
// accountNamePattern or exceeds accountNameMaxLength. kind names the
// resource in the error message ("user" or "group").
func validateAccountName(
	kind string,
	name string,
) error {
	if len(name) > accountNameMaxLength || !accountNamePattern.MatchString(name) {
		return fmt.Errorf(
			"invalid %s name %q: must match %s and be at most %d characters",
			kind,
			name,
			accountNamePattern.String(),
			accountNameMaxLength,
		)
	}

	return nil
}
