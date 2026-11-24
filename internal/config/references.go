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

package config

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/osapi-io/osapi/internal/secret/vault"
	"github.com/osapi-io/osapi/internal/validation"
)

// secretField is a string field of the config with mask:"password": a
// value that can be a reference, secret://<vault>/<key>.
type secretField struct {
	// key is the config key, such as
	// controller.api.security.signing_keys[0].key.
	key   string
	value reflect.Value
	// validate is the rule of the field, which its resolved value passes.
	validate string
	// refused, when it is set, tells why the field cannot be a reference:
	// a value that osapi needs before it can read a vault.
	refused string
}

// secretFields are the secret fields of c, each with its key.
func secretFields(
	c *Config,
) []secretField {
	return walkSecrets(reflect.ValueOf(c).Elem(), "")
}

func walkSecrets(
	v reflect.Value,
	prefix string,
) []secretField {
	var fields []secretField
	t := v.Type()
	for i := range t.NumField() {
		f := t.Field(i)
		name, _, _ := strings.Cut(f.Tag.Get("mapstructure"), ",")
		if prefix != "" {
			name = prefix + "." + name
		}
		fv := v.Field(i)
		switch {
		case fv.Kind() == reflect.Struct:
			fields = append(fields, walkSecrets(fv, name)...)
		case fv.Kind() == reflect.Slice && fv.Type().Elem().Kind() == reflect.Struct:
			for j := range fv.Len() {
				fields = append(fields, walkSecrets(fv.Index(j), fmt.Sprintf("%s[%d]", name, j))...)
			}
		case fv.Kind() == reflect.String && f.Tag.Get("mask") == "password":
			fields = append(fields, secretField{
				key:      name,
				value:    fv,
				validate: f.Tag.Get("validate"),
				refused:  f.Tag.Get("noref"),
			})
		}
	}

	return fields
}

// referenceProblems are the problems of each reference in c: one in a field
// that osapi needs before it can read a vault, one that does not parse, and
// one that names a vault that controller.secrets.vaults does not have.
func referenceProblems(
	c *Config,
) []Problem {
	vaults := map[string]bool{}
	for _, spec := range c.Controller.Secrets.Vaults {
		vaults[spec.Name] = true
	}
	var problems []Problem
	add := func(key string, msg string) {
		problems = append(problems, Problem{Rule: RuleReference, Key: key, Message: msg})
	}
	for _, f := range secretFields(c) {
		s := f.value.String()
		if !vault.IsReference(s) {
			continue
		}
		if f.refused != "" {
			add(f.key, "is a reference: "+f.refused)

			continue
		}
		ref, err := vault.ParseReference(s)
		if err != nil {
			add(f.key, err.Error())

			continue
		}
		if !vaults[ref.Vault] {
			add(f.key, fmt.Sprintf(
				"is %s: the vault %q is not in %s", ref, ref.Vault, vaultsKey,
			))
		}
	}

	return problems
}

// referenceKeys are the keys of c whose values are references. The rule of
// such a key applies to its value when it is resolved, not to the reference.
func referenceKeys(
	c *Config,
) map[string]bool {
	keys := map[string]bool{}
	for _, f := range secretFields(c) {
		if vault.IsReference(f.value.String()) {
			keys[f.key] = true
		}
	}

	return keys
}

// References are the keys of c that start with prefix, such as
// controller.api.security.signing_keys, whose values are references.
func References(
	c *Config,
	prefix string,
) []string {
	var keys []string
	for _, f := range secretFields(c) {
		if strings.HasPrefix(f.key, prefix) && vault.IsReference(f.value.String()) {
			keys = append(keys, f.key)
		}
	}

	return keys
}

// Resolver gives the value of the key that a reference names.
type Resolver func(ctx context.Context, ref vault.Reference) ([]byte, error)

// ResolveReferences puts the value of each reference of c, under the keys
// that start with prefix, in place of the reference. Each value passes the
// rule of its key, such as signing_key. An error names the key and the
// reference, and never a value.
func ResolveReferences(
	ctx context.Context,
	c *Config,
	prefix string,
	resolve Resolver,
) error {
	v := validation.Instance()
	if err := registerValidatorsFn(v); err != nil {
		return err
	}
	for _, f := range secretFields(c) {
		s := f.value.String()
		if !strings.HasPrefix(f.key, prefix) || !vault.IsReference(s) {
			continue
		}
		if f.refused != "" {
			return fmt.Errorf("%s is a reference: %s", f.key, f.refused)
		}
		ref, err := vault.ParseReference(s)
		if err != nil {
			return fmt.Errorf("%s: %w", f.key, err)
		}
		value, err := resolve(ctx, ref)
		if err != nil {
			return fmt.Errorf("%s: %w", f.key, err)
		}
		if f.validate != "" {
			if err := v.Var(string(value), f.validate); err != nil {
				return fmt.Errorf("%s: the value of %s %s", f.key, ref, broken(err))
			}
		}
		f.value.SetString(string(value))
	}

	return nil
}

// broken tells which rule err says that a value breaks, with no value.
// Var gives ValidationErrors for a string, which is never an invalid value.
func broken(
	err error,
) string {
	var errs validator.ValidationErrors
	_ = errors.As(err, &errs)
	msg := fmt.Sprintf("breaks its rule %q", errs[0].Tag())
	if hint := validation.Hint(context.Background(), errs[0]); hint != "" {
		msg += ": " + hint
	}

	return msg
}
