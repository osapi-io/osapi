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

package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"strings"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/labstack/echo/v5"

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/internal/secret"
	"github.com/osapi-io/osapi/internal/secret/vault"
	"github.com/osapi-io/osapi/internal/validation"
)

// PermSecretUse is the permission that a request needs for each
// reference, secret://<vault>/<key>, that it gives: to use a secret is to
// read it.
const PermSecretUse authtoken.Permission = "secret:use"

// Vaults are the vaults that a reference names, by name: the vaults of
// controller.secrets.vaults. The controller gives its *vault.Set.
type Vaults interface {
	Get(name string) (vault.Vault, bool)
	Spec(name string) (vault.Spec, bool)
}

// Secrets is what the server resolves each reference with. The controller
// gives it to New with the deps of the groups. A server with no Secrets
// has no vault, so it refuses each reference.
type Secrets struct {
	Vaults Vaults
}

// refusal is a request that resolve refuses: its status and its error.
type refusal struct {
	code int
	resp endpoint.ErrorResponse
}

func (r *refusal) Error() string {
	return r.resp.Error
}

// invalid is a 400 with err, which names the reference.
func invalid(
	err error,
) *refusal {
	return &refusal{http.StatusBadRequest, endpoint.ErrorResponse{
		Error: err.Error(), Reason: types.ReasonInvalidRequest,
	}}
}

// forbidden is a 403 for the reference ref.
func forbidden(
	ref string,
	format string,
	args ...any,
) *refusal {
	return &refusal{http.StatusForbidden, endpoint.ErrorResponse{
		Error: ref + ": " + fmt.Sprintf(format, args...), Reason: types.ReasonForbidden,
	}}
}

// resolve replaces each reference in *in with its value from its vault, in
// this order: each reference names a vault that exists (400), the caller
// may use each one (403), each key exists (400), then each value passes
// the validate tags of its field (400). The limits come before the read,
// so a caller that may not use a secret does not learn whether its key
// exists, and the controller reads no value that it does not send. The
// token scope of the host is checked before the handler runs, as for
// each request. Each error gives the reference, never a value.
func resolve[In any](
	c *echo.Context,
	s *Server,
	in *In,
) *refusal {
	ctx := c.Request().Context()
	vaults := s.vaults()
	refs := secret.References(*in)
	specs := make(map[string]vault.Spec, len(refs))
	for _, ref := range refs {
		name, key, err := secret.Parse(ref)
		if err != nil {
			return invalid(err)
		}
		if err := vault.CheckKey(key); err != nil {
			return invalid(fmt.Errorf("%s: %w", ref, err))
		}
		if vaults == nil {
			return invalid(secret.Unresolved(ref))
		}
		spec, ok := vaults.Spec(name)
		if !ok {
			return invalid(secret.Unresolved(ref))
		}
		specs[ref] = spec
	}

	if r := limits(ctx, c.Param("hostname"), refs, specs); r != nil {
		return r
	}

	err := secret.Resolve(in, func(ref string) (string, error) {
		return read(ctx, s, vaults, ref)
	})
	var r *refusal
	if errors.As(err, &r) {
		return r
	}
	if msg, ok := Valid(c, *in); !ok {
		return invalid(errors.New(msg))
	}

	return nil
}

// limits are the limits of use of each reference: the caller has
// secret:use, and the host is in the targets of the vault, when the vault
// names targets. A vault with targets gives no secret to a request that
// names no host, such as one that runs on the controller.
func limits(
	ctx context.Context,
	target string,
	refs []string,
	specs map[string]vault.Spec,
) *refusal {
	if !authtoken.Granted(ctx, PermSecretUse) {
		return forbidden(refs[0], "the token does not have the permission %s", PermSecretUse)
	}
	for _, ref := range refs {
		spec := specs[ref]
		if len(spec.Targets) == 0 {
			continue
		}
		targets := strings.Join(spec.Targets, ", ")
		if target == "" {
			return forbidden(
				ref,
				"the vault %q gives its secrets only to %s, and the request names no host",
				spec.Name,
				targets,
			)
		}
		if !validation.InScope(ctx, target, spec.Targets) {
			return forbidden(ref, "the vault %q gives its secrets only to %s, not to %q",
				spec.Name, targets, target)
		}
	}

	return nil
}

// read is the value of the reference ref, which names a vault that
// exists. A key that is not there is a 400; a vault that does not answer
// is a 500, which is logged.
func read(
	ctx context.Context,
	s *Server,
	vaults Vaults,
	ref string,
) (string, error) {
	name, key, _ := secret.Parse(ref)
	v, _ := vaults.Get(name)
	value, _, err := v.Get(ctx, key)
	switch {
	case errors.Is(err, vault.ErrNotFound):
		return "", invalid(fmt.Errorf("%s: the vault %q has no key %q", ref, name, key))
	case err != nil:
		s.logger.Error("the vault could not be read",
			slog.String("reference", ref),
			slog.String("error", err.Error()))

		return "", &refusal{http.StatusInternalServerError, endpoint.ErrorResponse{
			Error:  ref + ": the vault could not be read",
			Reason: types.ReasonServerError,
		}}
	}

	return string(value), nil
}

// vaults are the vaults of the Secrets that the server was given, or nil
// when it has none.
func (s *Server) vaults() Vaults {
	sec, _ := s.deps[reflect.TypeFor[Secrets]()].(Secrets)

	return sec.Vaults
}
