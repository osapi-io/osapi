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

package admission_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	store "github.com/osapi-io/osapi/internal/controller/admission"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/internal/controller/endpoints/admission"
	"github.com/osapi-io/osapi/internal/controller/endpoints/admission/mocks"
	"github.com/osapi-io/osapi/internal/job"
)

type MethodsPublicTestSuite struct {
	suite.Suite

	ctrl  *gomock.Controller
	store *mocks.MockStore
	jobs  *mocks.MockJobs
	at    time.Time
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}

func (s *MethodsPublicTestSuite) SetupSubTest() {
	s.ctrl = gomock.NewController(s.T())
	s.store = mocks.NewMockStore(s.ctrl)
	s.jobs = mocks.NewMockJobs(s.ctrl)
	s.at = time.Date(2026, 11, 19, 1, 0, 0, 0, time.UTC)
}

func (s *MethodsPublicTestSuite) deps() admission.Deps {
	return admission.Deps{Store: s.store, Jobs: s.jobs}
}

// fleet makes m-on online, reporting at s.at, and m-off seen an hour
// before.
func (s *MethodsPublicTestSuite) fleet() {
	s.jobs.EXPECT().ListAgents(gomock.Any()).
		Return([]job.AgentInfo{{MachineID: "m-on", RegisteredAt: s.at}}, nil)
	s.jobs.EXPECT().AgentsLastSeen(gomock.Any()).
		Return(map[string]time.Time{"m-off": s.at.Add(-time.Hour), "m-on": s.at.Add(-time.Minute)}, nil)
}

// is reports whether err is the endpoint error of kind.
func is(err error, kind error) bool {
	return errors.Is(err, kind)
}

// internal reports whether err is a 500, whose cause stays out of the
// answer.
func internal(err error) bool {
	var e *endpoint.InternalError
	return errors.As(err, &e)
}

func (s *MethodsPublicTestSuite) TestList() {
	records := []store.Record{
		{
			MachineID:  "m-on",
			Hostname:   "web-01",
			State:      store.StateAdmitted,
			AdmittedAt: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
		},
		{MachineID: "m-off", Hostname: "web-02", State: store.StateAdmitted},
		{
			MachineID:           "m-new",
			Hostname:            "web-03",
			State:               store.StatePending,
			Previous:            store.StateRevoked,
			PreviousFingerprint: "SHA256:old",
		},
	}

	s.Run("when the store has records returns each with online and last seen", func() {
		s.store.EXPECT().List(gomock.Any()).Return(records, nil)
		s.fleet()

		got, err := admission.List(context.Background(), s.deps(), admission.ListInput{})

		s.Require().NoError(err)
		s.Equal(3, got.TotalItems)
		s.True(got.Items[0].Online)
		s.Equal(s.at, *got.Items[0].LastSeen, "an online node's last report is its registry entry")
		s.NotNil(got.Items[0].AdmittedAt)
		s.False(got.Items[1].Online)
		s.Equal(s.at.Add(-time.Hour), *got.Items[1].LastSeen, "an offline node's is the stored one")
		s.Nil(got.Items[1].AdmittedAt, "an unset time is absent")
		s.Nil(got.Items[2].LastSeen, "a node never seen has none")
		s.Equal(store.StateRevoked, got.Items[2].Previous)
		s.Equal("SHA256:old", got.Items[2].PreviousFingerprint)
	})

	s.Run("when a state is given returns only those records", func() {
		s.store.EXPECT().List(gomock.Any()).Return(records, nil)
		s.fleet()

		got, err := admission.List(
			context.Background(),
			s.deps(),
			admission.ListInput{State: store.StatePending},
		)

		s.Require().NoError(err)
		s.Equal(1, got.TotalItems)
		s.Equal("m-new", got.Items[0].MachineID)
	})

	s.Run(
		"when the registry and the last-seen store cannot be read still returns the records",
		func() {
			s.store.EXPECT().List(gomock.Any()).Return(records, nil)
			s.jobs.EXPECT().ListAgents(gomock.Any()).Return(nil, errors.New("kv down"))
			s.jobs.EXPECT().AgentsLastSeen(gomock.Any()).Return(nil, errors.New("kv down"))

			got, err := admission.List(context.Background(), s.deps(), admission.ListInput{})

			s.Require().NoError(err)
			s.Equal(3, got.TotalItems)
			s.False(got.Items[0].Online)
		},
	)

	s.Run("when there is no job client returns the records without it", func() {
		s.store.EXPECT().List(gomock.Any()).Return(records, nil)

		got, err := admission.List(
			context.Background(),
			admission.Deps{Store: s.store},
			admission.ListInput{},
		)

		s.Require().NoError(err)
		s.Equal(3, got.TotalItems)
	})

	s.Run("when the store fails is a 500", func() {
		s.store.EXPECT().List(gomock.Any()).Return(nil, errors.New("kv down"))

		_, err := admission.List(context.Background(), s.deps(), admission.ListInput{})

		s.True(internal(err))
	})

	s.Run("when PKI is off is not configured", func() {
		_, err := admission.List(context.Background(), admission.Deps{}, admission.ListInput{})

		s.True(is(err, admission.ErrNotConfigured))
	})
}

func (s *MethodsPublicTestSuite) TestGet() {
	s.Run("when the machine has a record returns it", func() {
		s.store.EXPECT().
			Get(gomock.Any(), "m-on").
			Return(store.Record{MachineID: "m-on", State: store.StateAdmitted}, nil)
		s.fleet()

		got, err := admission.Get(
			context.Background(),
			s.deps(),
			admission.MachineInput{MachineID: "m-on"},
		)

		s.Require().NoError(err)
		s.True(got.Online)
	})

	s.Run("when the machine has no record is not found", func() {
		s.store.EXPECT().Get(gomock.Any(), "m-x").Return(store.Record{}, store.ErrNotFound)

		_, err := admission.Get(
			context.Background(),
			s.deps(),
			admission.MachineInput{MachineID: "m-x"},
		)

		s.True(is(err, admission.ErrNotFound))
	})

	s.Run("when the store fails is a 500", func() {
		s.store.EXPECT().Get(gomock.Any(), "m-x").Return(store.Record{}, errors.New("kv down"))

		_, err := admission.Get(
			context.Background(),
			s.deps(),
			admission.MachineInput{MachineID: "m-x"},
		)

		s.True(internal(err))
	})

	s.Run("when PKI is off is not configured", func() {
		_, err := admission.Get(
			context.Background(),
			admission.Deps{},
			admission.MachineInput{MachineID: "m-x"},
		)

		s.True(is(err, admission.ErrNotConfigured))
	})
}

func (s *MethodsPublicTestSuite) TestDecisions() {
	admit := func(d admission.Deps) error {
		_, err := admission.Admit(
			context.Background(),
			d,
			admission.AdmitInput{
				MachineID: "m-1",
				Body:      admission.AdmitRequest{Fingerprint: "SHA256:ab"},
			},
		)
		return err
	}
	reject := func(d admission.Deps) error {
		_, err := admission.Reject(
			context.Background(),
			d,
			admission.RejectInput{
				MachineID: "m-1",
				Body:      admission.RejectRequest{Reason: "unknown"},
			},
		)
		return err
	}
	revoke := func(d admission.Deps) error {
		_, err := admission.Revoke(
			context.Background(),
			d,
			admission.MachineInput{MachineID: "m-1"},
		)
		return err
	}
	del := func(d admission.Deps) error {
		_, err := admission.Delete(
			context.Background(),
			d,
			admission.MachineInput{MachineID: "m-1"},
		)
		return err
	}

	tests := []struct {
		name  string
		setup func()
		call  func(admission.Deps) error
		kind  error
		code  int
	}{
		{
			name:  "when admit matches admits",
			setup: func() { s.store.EXPECT().Admit(gomock.Any(), "m-1", "SHA256:ab").Return(nil) },
			call:  admit,
		},
		{
			name:  "when admit finds no request is not found",
			setup: func() { s.store.EXPECT().Admit(gomock.Any(), "m-1", "SHA256:ab").Return(store.ErrNoPendingAgent) },
			call:  admit,
			kind:  admission.ErrNotFound,
		},
		{
			name:  "when admit's fingerprint differs is a mismatch",
			setup: func() { s.store.EXPECT().Admit(gomock.Any(), "m-1", "SHA256:ab").Return(store.ErrFingerprintMismatch) },
			call:  admit,
			kind:  admission.ErrMismatch,
		},
		{
			name:  "when admit fails is a 500",
			setup: func() { s.store.EXPECT().Admit(gomock.Any(), "m-1", "SHA256:ab").Return(errors.New("kv down")) },
			call:  admit,
			code:  500,
		},
		{
			name:  "when reject finds the request rejects with the reason",
			setup: func() { s.store.EXPECT().Reject(gomock.Any(), "m-1", "unknown").Return(nil) },
			call:  reject,
		},
		{
			name:  "when reject finds no request is not found",
			setup: func() { s.store.EXPECT().Reject(gomock.Any(), "m-1", "unknown").Return(store.ErrNoPendingAgent) },
			call:  reject,
			kind:  admission.ErrNotFound,
		},
		{
			name:  "when reject fails is a 500",
			setup: func() { s.store.EXPECT().Reject(gomock.Any(), "m-1", "unknown").Return(errors.New("kv down")) },
			call:  reject,
			code:  500,
		},
		{
			name:  "when revoke finds the machine revokes",
			setup: func() { s.store.EXPECT().Revoke(gomock.Any(), "m-1").Return(nil) },
			call:  revoke,
		},
		{
			name:  "when revoke finds no admitted machine is not found",
			setup: func() { s.store.EXPECT().Revoke(gomock.Any(), "m-1").Return(store.ErrNoAcceptedAgent) },
			call:  revoke,
			kind:  admission.ErrNotFound,
		},
		{
			name:  "when revoke fails is a 500",
			setup: func() { s.store.EXPECT().Revoke(gomock.Any(), "m-1").Return(errors.New("kv down")) },
			call:  revoke,
			code:  500,
		},
		{
			name:  "when delete finds the machine deletes it",
			setup: func() { s.store.EXPECT().Delete(gomock.Any(), "m-1").Return(nil) },
			call:  del,
		},
		{
			name:  "when delete finds no record is not found",
			setup: func() { s.store.EXPECT().Delete(gomock.Any(), "m-1").Return(store.ErrNotFound) },
			call:  del,
			kind:  admission.ErrNotFound,
		},
		{
			name:  "when delete fails is a 500",
			setup: func() { s.store.EXPECT().Delete(gomock.Any(), "m-1").Return(errors.New("kv down")) },
			call:  del,
			code:  500,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()

			err := tc.call(s.deps())

			switch {
			case tc.kind != nil:
				s.True(is(err, tc.kind), "%v", err)
			case tc.code != 0:
				s.True(internal(err), "%v", err)
			default:
				s.NoError(err)
			}
		})
	}

	s.Run("when PKI is off each decision is not configured", func() {
		for _, call := range []func(admission.Deps) error{admit, reject, revoke, del} {
			s.True(is(call(admission.Deps{}), admission.ErrNotConfigured))
		}
	})
}
