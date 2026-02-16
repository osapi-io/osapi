// Copyright (c) 2025 John Dewey

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

package client_test

import (
	"github.com/golang/mock/gomock"
	"github.com/nats-io/nats.go"

	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
)

// publishAndWaitErrorMode controls which step of the publishAndWait flow fails.
type publishAndWaitErrorMode int

const (
	// errorOnKVPut makes kv.Put return the error.
	errorOnKVPut publishAndWaitErrorMode = iota
	// errorOnPublish makes natsClient.Publish return the error.
	errorOnPublish
	// errorOnWatch makes kv.Watch return the error.
	errorOnWatch
	// errorOnTimeout simulates a timeout by providing a channel that never sends.
	errorOnTimeout
)

// publishAndWaitMockOpts configures the mock behavior for publishAndWait tests.
type publishAndWaitMockOpts struct {
	// responseData is the JSON response to return from the watcher entry.
	responseData string
	// mockError is the error to inject (used with errorMode).
	mockError error
	// errorMode controls which step fails when mockError is set.
	errorMode publishAndWaitErrorMode
	// sendNilFirst sends a nil entry before the real entry on the watcher channel.
	sendNilFirst bool
}

// setupPublishAndWaitMocks configures mocks for the publishAndWait flow.
func setupPublishAndWaitMocks(
	ctrl *gomock.Controller,
	mockKV *jobmocks.MockKeyValue,
	mockNATSClient *jobmocks.MockNATSClient,
	subject string,
	responseData string,
	mockError error,
) {
	setupPublishAndWaitMocksWithOpts(ctrl, mockKV, mockNATSClient, subject, &publishAndWaitMockOpts{
		responseData: responseData,
		mockError:    mockError,
		errorMode:    errorOnKVPut,
	})
}

// setupPublishAndWaitMocksWithOpts configures mocks with fine-grained control.
func setupPublishAndWaitMocksWithOpts(
	ctrl *gomock.Controller,
	mockKV *jobmocks.MockKeyValue,
	mockNATSClient *jobmocks.MockNATSClient,
	subject string,
	opts *publishAndWaitMockOpts,
) {
	if opts.mockError != nil && opts.errorMode == errorOnKVPut {
		mockKV.EXPECT().
			Put(gomock.Any(), gomock.Any()).
			Return(uint64(0), opts.mockError)
		return
	}

	// kv.Put succeeds
	mockKV.EXPECT().
		Put(gomock.Any(), gomock.Any()).
		Return(uint64(1), nil)

	if opts.mockError != nil && opts.errorMode == errorOnPublish {
		mockNATSClient.EXPECT().
			Publish(gomock.Any(), subject, gomock.Any()).
			Return(opts.mockError)
		return
	}

	// natsClient.Publish succeeds
	mockNATSClient.EXPECT().
		Publish(gomock.Any(), subject, gomock.Any()).
		Return(nil)

	if opts.mockError != nil && opts.errorMode == errorOnWatch {
		mockKV.EXPECT().
			Watch(gomock.Any()).
			Return(nil, opts.mockError)
		return
	}

	if opts.mockError != nil && opts.errorMode == errorOnTimeout {
		// Return a channel that never sends anything, causing timeout
		ch := make(chan nats.KeyValueEntry)

		mockWatcher := jobmocks.NewMockKeyWatcher(ctrl)
		mockWatcher.EXPECT().Updates().Return(ch).AnyTimes()
		mockWatcher.EXPECT().Stop().Return(nil)

		mockKV.EXPECT().
			Watch(gomock.Any()).
			Return(mockWatcher, nil)
		return
	}

	// Create mock entry with response data
	mockEntry := jobmocks.NewMockKeyValueEntry(ctrl)
	mockEntry.EXPECT().Value().Return([]byte(opts.responseData))

	// Create buffered channel and optionally send nil first
	bufSize := 1
	if opts.sendNilFirst {
		bufSize = 2
	}
	ch := make(chan nats.KeyValueEntry, bufSize)
	if opts.sendNilFirst {
		ch <- nil
	}
	ch <- mockEntry

	// Create mock watcher
	mockWatcher := jobmocks.NewMockKeyWatcher(ctrl)
	mockWatcher.EXPECT().Updates().Return(ch).AnyTimes()
	mockWatcher.EXPECT().Stop().Return(nil)

	// kv.Watch returns the mock watcher
	mockKV.EXPECT().
		Watch(gomock.Any()).
		Return(mockWatcher, nil)
}
