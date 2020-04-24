package handlers

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cbsinteractive/bakery/config"
	test "github.com/cbsinteractive/bakery/tests"
	"github.com/cbsinteractive/pkg/tracing"
)

func testConfig(fc test.FakeClient) (config.Config, error) {
	timeout, err := time.ParseDuration("5s")
	if err != nil {
		return config.Config{}, err
	}

	return config.Config{
		Listen:     "8080",
		LogLevel:   "panic",
		OriginHost: "http://localhost:8080",
		Hostname:   "hostname",
		Client: config.Client{
			Timeout:    timeout,
			Tracer:     tracing.NoopTracer{},
			HTTPClient: fc,
		},
	}, nil
}

func getRequest(url string, t *testing.T) *http.Request {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("could not create request to endpoint: %v, got error: %v", url, err)
	}

	return req
}

func getResponseRecorder() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}

func default200Response() func(req *http.Request) (*http.Response, error) {
	return func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBufferString("OK")),
		}, nil
	}
}
