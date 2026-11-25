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

package bus

import (
	"fmt"
	"sort"

	natsserver "github.com/nats-io/nats-server/v2/server"

	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/pki"
)

// Role is an osapi process, as NATS sees it: what it publishes and what it
// subscribes to.
type Role string

// The roles of osapi on NATS.
const (
	// RoleController is the controller: it sends jobs, reads their answers,
	// keeps the admission records and the audit log.
	RoleController Role = "controller"
	// RoleAgent is an agent: it takes the jobs of its host, and writes their
	// answers, its registration, its facts and its state.
	RoleAgent Role = "agent"
	// RoleSetup makes and updates each stream and bucket: osapi nats setup,
	// and the embedded server, which also writes its own registration.
	RoleSetup Role = "setup"
)

// Roles are each role, in order.
var Roles = []Role{RoleController, RoleAgent, RoleSetup}

// Permissions are the subjects that a role may publish and subscribe to, as
// NATS user permissions take them.
type Permissions struct {
	Publish   []string `json:"publish"`
	Subscribe []string `json:"subscribe"`
}

// RolePermissions are the permissions of role on a NATS server with the
// streams and buckets of cfg, under namespace: the least that the role needs.
// The names come from the same functions that make the streams and buckets,
// so a renamed bucket changes them too.
func RolePermissions(
	role Role,
	namespace string,
	cfg config.NATS,
) (Permissions, error) {
	n := names(namespace, cfg)
	p := perms{pub: map[string]bool{"$JS.API.INFO": true}, sub: map[string]bool{"_INBOX.>": true}}
	switch role {
	case RoleAgent:
		for _, b := range []string{n.jobs, n.registry, n.facts, n.state, n.fileState} {
			p.kv(b)
		}
		p.readObjects(n.objects)
		p.consume(n.stream)
		p.pub[subject(namespace, pki.EnrollRequestSuffix)] = true
		p.sub[subject(namespace, pki.EnrollResponsePrefix)+".>"] = true
	case RoleController:
		for _, b := range []string{n.jobs, n.registry, n.facts, n.state, n.fileState, n.enrollment, n.secrets} {
			p.kv(b)
		}
		p.readObjects(n.objects)
		p.writeObjects(n.objects)
		p.consume(n.stream)
		p.consume(n.dlq)
		p.consume(n.audit)
		p.pub["$JS.API.STREAM.MSG.GET."+n.audit] = true
		for _, s := range []string{n.audit, kvStream(n.enrollment), objStream(n.objects)} {
			p.admin(s)
		}
		p.pub["$JS.API.STREAM.PURGE."+objStream(n.objects)] = true
		p.pub[job.ApplyNamespaceToSubjects(namespace, cfg.Stream.Subjects)] = true
		p.pub[n.auditSubject+".>"] = true
		p.pub[subject(namespace, pki.EnrollResponsePrefix)+".>"] = true
		p.sub[subject(namespace, pki.EnrollRequestSuffix)] = true
	case RoleSetup:
		for _, s := range n.allStreams() {
			p.admin(s)
		}
		p.kv(n.registry)
	default:
		return Permissions{}, fmt.Errorf("%w: %q", ErrRole, role)
	}

	return p.sorted(), nil
}

// ErrRole is a role that osapi does not have.
var ErrRole = fmt.Errorf("no such role")

// infra is the name of each stream and bucket of osapi under a namespace.
type infra struct {
	stream, dlq, audit, auditSubject                                      string
	jobs, registry, facts, state, fileState, enrollment, secrets, objects string
}

// names are the streams and buckets of cfg under namespace.
func names(
	namespace string,
	cfg config.NATS,
) infra {
	stream := job.ApplyNamespaceToInfraName(namespace, cfg.Stream.Name)

	return infra{
		stream:       stream,
		dlq:          stream + "-DLQ",
		audit:        AuditStream(namespace, cfg.Audit).Name,
		auditSubject: job.ApplyNamespaceToSubjects(namespace, cfg.Audit.Subject),
		jobs:         JobKV(namespace, cfg.KV).Bucket,
		registry:     RegistryKV(namespace, cfg.Registry).Bucket,
		facts:        FactsKV(namespace, cfg.Facts).Bucket,
		state:        StateKV(namespace, cfg.State).Bucket,
		fileState:    FileStateKV(namespace, cfg.FileState).Bucket,
		enrollment:   EnrollmentKV(namespace, cfg.Enrollment).Bucket,
		secrets:      SecretsKV(namespace, cfg.Secrets).Bucket,
		objects:      ObjectStore(namespace, cfg.Objects).Bucket,
	}
}

// allStreams is each stream: the job, dead letter and audit streams, and the
// stream of each bucket and object store.
func (n infra) allStreams() []string {
	out := []string{n.stream, n.dlq, n.audit, objStream(n.objects)}
	for _, b := range []string{n.jobs, n.registry, n.facts, n.state, n.fileState, n.enrollment, n.secrets} {
		out = append(out, kvStream(b))
	}

	return out
}

func kvStream(bucket string) string  { return "KV_" + bucket }
func objStream(bucket string) string { return "OBJ_" + bucket }

// subject is suffix under namespace, as osapi names its subjects.
func subject(
	namespace string,
	suffix string,
) string {
	if namespace == "" {
		return suffix
	}

	return namespace + "." + suffix
}

// perms collects subjects, each once.
type perms struct {
	pub, sub map[string]bool
}

// consume lets the role read stream: its info, its consumers, their
// messages and acks, and a direct get.
func (p perms) consume(stream string) {
	for _, s := range []string{
		"$JS.API.STREAM.INFO." + stream,
		"$JS.API.CONSUMER.*." + stream,
		"$JS.API.CONSUMER.*." + stream + ".>",
		"$JS.API.CONSUMER.DURABLE.CREATE." + stream + ".>",
		"$JS.FC." + stream + ".>",
		"$JS.API.CONSUMER.MSG.NEXT." + stream + ".>",
		"$JS.API.DIRECT.GET." + stream + ".>",
		"$JS.ACK." + stream + ".>",
	} {
		p.pub[s] = true
	}
}

// kv lets the role read and write bucket.
func (p perms) kv(bucket string) {
	p.consume(kvStream(bucket))
	p.pub["$KV."+bucket+".>"] = true
}

// readObjects lets the role read the object store bucket.
func (p perms) readObjects(bucket string) {
	p.consume(objStream(bucket))
}

// writeObjects lets the role write the object store bucket.
func (p perms) writeObjects(bucket string) {
	p.pub["$O."+bucket+".>"] = true
}

// admin lets the role make and update stream.
func (p perms) admin(stream string) {
	for _, verb := range []string{"CREATE", "UPDATE", "INFO"} {
		p.pub["$JS.API.STREAM."+verb+"."+stream] = true
	}
}

// sorted is p as sorted lists.
func (p perms) sorted() Permissions {
	keys := func(m map[string]bool) []string {
		out := make([]string, 0, len(m))
		for k := range m {
			out = append(out, k)
		}
		sort.Strings(out)

		return out
	}

	return Permissions{Publish: keys(p.pub), Subscribe: keys(p.sub)}
}

// ApplyRoles limits each user of opts that has a role in serverAuth to the
// permissions of the role, on the streams and buckets of cfg under
// namespace. A user with no role keeps each subject.
func ApplyRoles(
	opts *natsserver.Options,
	serverAuth config.NATSServerAuth,
	namespace string,
	cfg config.NATS,
) error {
	roles := map[string]string{}
	for _, u := range serverAuth.Users {
		if u.Role != "" {
			roles[u.Username] = u.Role
		}
	}
	for _, u := range opts.Users {
		role, ok := roles[u.Username]
		if !ok {
			continue
		}
		p, err := RolePermissions(Role(role), namespace, cfg)
		if err != nil {
			return fmt.Errorf("user %s: %w", u.Username, err)
		}
		u.Permissions = &natsserver.Permissions{
			Publish:   &natsserver.SubjectPermission{Allow: p.Publish},
			Subscribe: &natsserver.SubjectPermission{Allow: p.Subscribe},
		}
	}

	return nil
}
