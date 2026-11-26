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

package agent_test

import (
	"github.com/osapi-io/osapi/internal/agent"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/pki"
	"github.com/osapi-io/osapi/internal/secret/seal"
)

// TestOpenJobAfterARotation: a running agent opens a job sealed to its new
// key after a rotation moved the keys on disk, and a job sealed to its old
// key just before.
func (s *AgentPublicTestSuite) TestOpenJobAfterARotation() {
	// sealedTo is a job sealed to the public encryption key pub.
	sealedTo := func(pub string) job.Request {
		sealer, err := seal.New("job-1")
		s.Require().NoError(err)
		key, err := sealer.KeyFor(pub)
		s.Require().NoError(err)

		return job.Request{
			JobID: "job-1", SealedKeys: map[string]string{"m-1": key}, SealedFrom: sealer.From(),
		}
	}
	// running is an agent that loaded the keys on disk, and the keys.
	running := func() (*agent.Agent, *pki.Manager) {
		a := s.buildAgent()
		agent.SetAgentMachineID(a, "m-1")
		m := pki.New(s.appFs, keyDir, "agent", pki.WithEncryptionKey())
		s.Require().NoError(m.LoadOrGenerate())
		agent.SetAgentPKIManager(a, m)

		return a, m
	}
	// rotated moves new keys in place on disk, as osapi pki key rotate
	// does, and returns their public encryption key.
	rotated := func() string {
		disk := pki.New(s.appFs, keyDir, "agent", pki.WithEncryptionKey())
		s.Require().NoError(disk.Load())
		next, err := disk.Next()
		s.Require().NoError(err)
		s.Require().NoError(disk.Promote(next))

		return next.EncryptionPublicKey()
	}

	s.Run("a job sealed to the new key opens: the agent reads the new keys", func() {
		a, m := running()
		before := m.Fingerprint()
		job := sealedTo(rotated())
		_, err := agent.ExportOpenJob(a, job)
		s.Require().NoError(err)
		s.NotEqual(before, m.Fingerprint())
	})
	s.Run("a job sealed to the old key just before the rotation opens", func() {
		a, m := running()
		job := sealedTo(m.EncryptionPublicKey())
		rotated()
		_, err := agent.ExportOpenJob(a, job)
		s.Require().NoError(err)
	})
	s.Run("a job sealed to a key that is neither", func() {
		a, _ := running()
		rotated()
		other := pki.New(s.appFs, "/other", "agent", pki.WithEncryptionKey())
		s.Require().NoError(other.LoadOrGenerate())
		_, err := agent.ExportOpenJob(a, sealedTo(other.EncryptionPublicKey()))
		s.Error(err)
	})
	s.Run("keys that are gone from the disk: the agent keeps the keys it has", func() {
		a, m := running()
		job := sealedTo(m.EncryptionPublicKey())
		s.Require().NoError(s.appFs.Remove(keyDir + "/agent.pub"))
		_, err := agent.ExportOpenJob(a, job)
		s.Require().NoError(err)
	})
}
