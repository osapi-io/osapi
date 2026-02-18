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
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/retr0h/osapi/internal/api/network/gen"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/validation"
)

// dnsPutMultiResponse wraps multiple DNS modify results for _all broadcast.
type dnsPutMultiResponse struct {
	Results []dnsPutResult `json:"results"`
}

type dnsPutResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
}

func (r dnsPutMultiResponse) VisitPutNetworkDNSResponse(
	w http.ResponseWriter,
) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	return json.NewEncoder(w).Encode(r)
}

// PutNetworkDNS put the network dns API endpoint.
func (n Network) PutNetworkDNS(
	ctx context.Context,
	request gen.PutNetworkDNSRequestObject,
) (gen.PutNetworkDNSResponseObject, error) {
	if errMsg, ok := validation.Struct(request.Body); !ok {
		return gen.PutNetworkDNS400JSONResponse{
			Error: &errMsg,
		}, nil
	}

	if request.Params.TargetHostname != nil {
		th := struct {
			TargetHostname string `validate:"min=1"`
		}{TargetHostname: *request.Params.TargetHostname}
		if errMsg, ok := validation.Struct(th); !ok {
			return gen.PutNetworkDNS400JSONResponse{Error: &errMsg}, nil
		}
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

	hostname := job.AnyHost
	if request.Params.TargetHostname != nil {
		hostname = *request.Params.TargetHostname
	}

	if hostname == job.BroadcastHost {
		return n.putNetworkDNSAll(ctx, servers, searchDomains, interfaceName)
	}

	err := n.JobClient.ModifyNetworkDNS(ctx, hostname, servers, searchDomains, interfaceName)
	if err != nil {
		errMsg := err.Error()
		return gen.PutNetworkDNS500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	return gen.PutNetworkDNS202Response{}, nil
}

// putNetworkDNSAll handles _all broadcast for DNS modification.
func (n Network) putNetworkDNSAll(
	ctx context.Context,
	servers []string,
	searchDomains []string,
	interfaceName string,
) (gen.PutNetworkDNSResponseObject, error) {
	results, err := n.JobClient.ModifyNetworkDNSAll(ctx, servers, searchDomains, interfaceName)
	if err != nil {
		errMsg := err.Error()
		return gen.PutNetworkDNS500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	var responses []dnsPutResult
	for host, hostErr := range results {
		r := dnsPutResult{
			Hostname: host,
			Status:   "ok",
		}
		if hostErr != nil {
			r.Status = "failed"
			r.Error = fmt.Sprintf("%v", hostErr)
		}
		responses = append(responses, r)
	}

	return dnsPutMultiResponse{Results: responses}, nil
}
