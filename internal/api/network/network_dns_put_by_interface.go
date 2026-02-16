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

package network

import (
	"context"

	"github.com/go-playground/validator/v10"

	"github.com/retr0h/osapi/internal/api/network/gen"
)

// PutNetworkDNS put the network dns API endpoint.
func (n Network) PutNetworkDNS(
	ctx context.Context,
	request gen.PutNetworkDNSRequestObject,
) (gen.PutNetworkDNSResponseObject, error) {
	validate := validator.New()
	if err := validate.Struct(request.Body); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		errMsg := validationErrors.Error()
		return gen.PutNetworkDNS400JSONResponse{
			Error: &errMsg,
		}, nil
	}

	var servers []string
	if request.Body.Servers != nil {
		servers = *request.Body.Servers
	}

	var searchDomains []string
	if request.Body.SearchDomains != nil {
		searchDomains = *request.Body.SearchDomains
	}

	interfaceName := request.Body.InterfaceName

	err := n.JobClient.ModifyNetworkDNSAny(ctx, servers, searchDomains, interfaceName)
	if err != nil {
		errMsg := err.Error()
		return gen.PutNetworkDNS500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	return gen.PutNetworkDNS202Response{}, nil
}
