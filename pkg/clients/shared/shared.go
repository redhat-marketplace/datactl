// Copyright 2021 IBM Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package shared

import (
	"crypto/tls"
	"log"
	"net/http"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/util/httpstream/spdy"
	"k8s.io/klog/v2/klogr"
)

var (
	logger logr.Logger = klogr.New().V(5).WithName("pkg/clients/shared")
)

type RoundTripperConfig spdy.RoundTripperConfig

type RoundTripperOptions func(http.RoundTripper) http.RoundTripper

func NewHttpClient(
	tlsConfig *tls.Config,
	opts ...RoundTripperOptions,
) (*http.Client, error) {
	client := &http.Client{}

	var roundtripper http.RoundTripper
	roundtripper, err := spdy.NewRoundTripper(tlsConfig)
	if err != nil {
		return nil, err
	}

	for _, apply := range opts {
		roundtripper = apply(roundtripper)
	}

	client.Transport = roundtripper
	return client, nil
}

type withHeader struct {
	http.Header
	rt http.RoundTripper
}

func WithBearerAuth(token string) RoundTripperOptions {
	if token == "" {
		logger.WithCallDepth(1).Info("bearer token is empty")
		return func(rt http.RoundTripper) http.RoundTripper {
			return rt
		}
	}

	rt := WithHeaders("Authorization", "Bearer "+token)
	return rt
}

func WithHeaders(key, value string, values ...string) RoundTripperOptions {
	return func(rt http.RoundTripper) http.RoundTripper {
		if len(values)%2 != 0 {
			log.Fatal("values must be divisible by 2")
		}

		rt2 := withHeader{Header: make(http.Header), rt: rt}

		rt2.addHeader(key, value)

		for i := 0; i < len(values); i = i + 2 {
			rt2.addHeader(values[i], values[2])
		}

		return rt2
	}
}

func (h withHeader) addHeader(key, value string) {
	_, ok := h.Header[key]

	if !ok {
		h.Header[key] = []string{value}
	} else {
		h.Header[key] = append(h.Header[key], value)
	}
}

func (h withHeader) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range h.Header {
		req.Header[k] = v
	}

	return h.rt.RoundTrip(req)
}
