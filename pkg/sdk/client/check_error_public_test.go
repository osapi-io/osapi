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

package client_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/pkg/sdk/client"
)

type CheckErrorPublicTestSuite struct {
	suite.Suite
}

func (suite *CheckErrorPublicTestSuite) TestCheckError() {
	tests := []struct {
		name         string
		statusCode   int
		validateFunc func(error)
	}{
		{
			name:       "when status is 200",
			statusCode: 200,
			validateFunc: func(err error) {
				suite.NoError(err)
			},
		},
		{
			name:       "when status is 201",
			statusCode: 201,
			validateFunc: func(err error) {
				suite.NoError(err)
			},
		},
		{
			name:       "when status is 202",
			statusCode: 202,
			validateFunc: func(err error) {
				suite.NoError(err)
			},
		},
		{
			name:       "when status is 204",
			statusCode: 204,
			validateFunc: func(err error) {
				suite.NoError(err)
			},
		},
		{
			name:       "when status is 400",
			statusCode: 400,
			validateFunc: func(err error) {
				suite.Error(err)
				var target *client.ValidationError
				suite.Require().True(errors.As(err, &target), "err is %T: %v", err, err)
				suite.Equal(400, target.StatusCode)
			},
		},
		{
			name:       "when status is 401",
			statusCode: 401,
			validateFunc: func(err error) {
				suite.Error(err)
				var target *client.AuthError
				suite.Require().True(errors.As(err, &target), "err is %T: %v", err, err)
				suite.Equal(401, target.StatusCode)
			},
		},
		{
			name:       "when status is 403",
			statusCode: 403,
			validateFunc: func(err error) {
				suite.Error(err)
				var target *client.AuthError
				suite.Require().True(errors.As(err, &target), "err is %T: %v", err, err)
				suite.Equal(403, target.StatusCode)
			},
		},
		{
			name:       "when status is 404",
			statusCode: 404,
			validateFunc: func(err error) {
				suite.Error(err)
				var target *client.NotFoundError
				suite.Require().True(errors.As(err, &target), "err is %T: %v", err, err)
				suite.Equal(404, target.StatusCode)
			},
		},
		{
			name:       "when status is 409",
			statusCode: 409,
			validateFunc: func(err error) {
				suite.Error(err)
				var target *client.ConflictError
				suite.Require().True(errors.As(err, &target), "err is %T: %v", err, err)
				suite.Equal(409, target.StatusCode)
			},
		},
		{
			name:       "when status is 500",
			statusCode: 500,
			validateFunc: func(err error) {
				suite.Error(err)
				var target *client.ServerError
				suite.Require().True(errors.As(err, &target), "err is %T: %v", err, err)
				suite.Equal(500, target.StatusCode)
			},
		},
		{
			name:       "when status is 503",
			statusCode: 503,
			validateFunc: func(err error) {
				suite.Error(err)
				var target *client.UnexpectedStatusError
				suite.Require().True(errors.As(err, &target), "err is %T: %v", err, err)
				suite.Equal(503, target.StatusCode)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			err := client.ExportCheckError(tc.statusCode, nil)
			tc.validateFunc(err)
		})
	}
}

func (suite *CheckErrorPublicTestSuite) TestCheckErrorMessages() {
	tests := []struct {
		name         string
		statusCode   int
		body         *client.ErrorResponse
		validateFunc func(error)
	}{
		{
			name:       "when the error body has a message and a reason",
			statusCode: 400,
			body: &client.ErrorResponse{
				Error:  "field 'name' is required",
				Reason: "invalid_request",
			},
			validateFunc: func(err error) {
				suite.Contains(err.Error(), "field 'name' is required")
				var apiErr *client.ValidationError
				suite.Require().ErrorAs(err, &apiErr)
				suite.Equal("invalid_request", apiErr.Reason)
			},
		},
		{
			name:       "when there is no error body",
			statusCode: 500,
			validateFunc: func(err error) {
				suite.Contains(err.Error(), "unexpected status 500")
			},
		},
		{
			name:       "when the error body has no message",
			statusCode: 404,
			body:       &client.ErrorResponse{},
			validateFunc: func(err error) {
				suite.Contains(err.Error(), "unexpected status 404")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			err := client.ExportCheckError(tc.statusCode, tc.body)
			tc.validateFunc(err)
		})
	}
}

func TestCheckErrorPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(CheckErrorPublicTestSuite))
}
