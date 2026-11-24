// Copyright (c) 2024 John Dewey

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
package authtoken

import (
	"context"

	"github.com/osapi-io/osapi/internal/validation"
	"github.com/osapi-io/osapi/internal/validation/rule"
)

// TargetScopeRule is the validate rule of an entry of a token's target
// scope: a hostname, such as web-01, or a label, such as group:web.
const TargetScopeRule = "target_scope"

func init() {
	rule.Declare(rule.Custom{
		Tag:     TargetScopeRule,
		Check:   validation.ScopeEntry,
		Schema:  rule.None,
		Message: "is not a hostname or a label such as group:web; _all, _any and _local name each host",
	})
}

// targetsKey is the context key of a token's target scope.
type targetsKey struct{}

// WithTargets is ctx with the target scope of the token of the request.
func WithTargets(
	ctx context.Context,
	targets []string,
) context.Context {
	return context.WithValue(ctx, targetsKey{}, targets)
}

// TargetsFrom is the target scope that WithTargets put in ctx, or nil when
// the token has none, which covers each host.
func TargetsFrom(
	ctx context.Context,
) []string {
	targets, _ := ctx.Value(targetsKey{}).([]string)

	return targets
}

// InScope reports whether the token of ctx may send work to target.
func InScope(
	ctx context.Context,
	target string,
) bool {
	return validation.InScope(ctx, target, TargetsFrom(ctx))
}

// Sees reports whether the token of ctx may read a job, a trace or an audit
// entry whose target is target. A token with no scope sees each one. A scoped
// token sees one only when its target is in the scope, so a record with no
// target, or with the target _all or _any, is hidden from it.
func Sees(
	ctx context.Context,
	target string,
) bool {
	if len(TargetsFrom(ctx)) == 0 {
		return true
	}

	return target != "" && InScope(ctx, target)
}

// subjectKey is the context key of the subject of a token.
type subjectKey struct{}

// WithSubject is ctx with the subject of the token of the request, such as
// ops@example.com.
func WithSubject(
	ctx context.Context,
	subject string,
) context.Context {
	return context.WithValue(ctx, subjectKey{}, subject)
}

// SubjectFrom is the subject that WithSubject put in ctx, or "" when the
// request has no token.
func SubjectFrom(
	ctx context.Context,
) string {
	subject, _ := ctx.Value(subjectKey{}).(string)

	return subject
}
