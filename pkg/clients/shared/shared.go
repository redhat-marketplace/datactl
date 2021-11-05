package shared

import (
	"crypto/tls"
	"log"
	"net/http"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/httpstream/spdy"
)

type RoundTripperConfig spdy.RoundTripperConfig

type RoundTripperOptions func(http.RoundTripper) http.RoundTripper

func NewHttpClient(
	tlsConfig *tls.Config,
	opts ...RoundTripperOptions,
) *http.Client {
	client := &http.Client{}

	var roundtripper http.RoundTripper = spdy.NewRoundTripper(tlsConfig, true, true)

	for _, apply := range opts {
		roundtripper = apply(roundtripper)
	}

	client.Transport = roundtripper
	return client
}

type withHeader struct {
	http.Header
	rt http.RoundTripper
}

func WithBearerAuth(token string) RoundTripperOptions {
	if token == "" {
		logrus.Warn("bearer token is empty")
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
