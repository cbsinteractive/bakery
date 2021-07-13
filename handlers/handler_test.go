package handlers

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cbsinteractive/bakery/config"
	"github.com/cbsinteractive/bakery/filters"
	test "github.com/cbsinteractive/bakery/tests"
	"github.com/cbsinteractive/pkg/tracing"
	"github.com/google/go-cmp/cmp"
)

func testConfig(fc test.FakeClient) config.Config {
	return config.Config{
		Listen:      "8080",
		LogLevel:    "panic",
		OriginHost:  "http://localhost:8080",
		Hostname:    "hostname",
		OriginKey:   "x-bakery-origin-token",
		OriginToken: "authenticate-me",
		AuthEnabled: true,
		Client: config.Client{
			Timeout:    5 * time.Second,
			Tracer:     tracing.NoopTracer{},
			HTTPClient: fc,
		},
		Propeller: config.Propeller{
			Enabled: true,
		},
	}
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

func default200Response(msg string) func(req *http.Request) (*http.Response, error) {
	return func(*http.Request) (*http.Response, error) {
		resp := &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBufferString(msg)),
			Header:     http.Header{},
		}

		lastModified := time.Now().UTC().Format(http.TimeFormat)
		resp.Header.Add("Last-Modified", lastModified)

		return resp, nil
	}
}

func default404Response(msg string) func(req *http.Request) (*http.Response, error) {
	return func(*http.Request) (*http.Response, error) {
		resp := &http.Response{
			StatusCode: 404,
			Body:       ioutil.NopCloser(bytes.NewBufferString(msg)),
			Header:     http.Header{},
		}

		lastModified := time.Now().UTC().Format(http.TimeFormat)
		resp.Header.Add("Last-Modified", lastModified)

		return resp, nil
	}
}

func getManifest() string {
	return `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=2000,CODECS="avc1.77.30,mp4a"
http://existing.base/uri/link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,CODECS="avc1.77.30,mp4a"
http://existing.base/uri/link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=6000,CODECS="avc1.77.30,mp4a"
http://existing.base/uri/link_3.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=8000,CODECS="avc1.77.30,mp4a"
http://existing.base/uri/link_4.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=10000,CODECS="avc1.77.30,mp4a"
http://existing.base/uri/link_5.m3u8
`
}

func TestHandler(t *testing.T) {

	tests := []struct {
		name           string
		url            string
		auth           string
		mockResp       func(req *http.Request) (*http.Response, error)
		expectStatus   int
		expectManifest string
	}{
		{
			name:           "when manifest returns 4xx, expect 500  w/ err msg reflecting origin status code",
			url:            "origin/some/path/to/master.m3u8",
			auth:           "authenticate-me",
			mockResp:       default200Response(getManifest()),
			expectStatus:   200,
			expectManifest: getManifest(),
		},
		{
			name:           "when PreventHTTPStatusError (phe) filter is enabled, should prevent returning 404 for m3u8",
			url:            "phe(true)/aHR0cHM6Ly8wODc2M2JmMGIxZ2IuYWlyc3BhY2UtY2RuLmNic2l2aWRlby5jb20vbXR2LWVtYS11ay1obHMvbWFzdGVyLzQwNC5tM3U4.m3u8",
			auth:           "authenticate-me",
			mockResp:       default404Response("404"),
			expectStatus:   200,
			expectManifest: filters.EmptyHLSManifestContent,
		},
		{
			name:           "when PreventHTTPStatusError (phe) filter is enabled, should prevent returning 404 for vtt",
			url:            "phe(true)/aHR0cHM6Ly8wODc2M2JmMGIxZ2IuYWlyc3BhY2UtY2RuLmNic2l2aWRlby5jb20vbXR2LWVtYS11ay1obHMvbWFzdGVyLzQwNC5tM3U4.vtt",
			auth:           "authenticate-me",
			mockResp:       default404Response("404"),
			expectStatus:   200,
			expectManifest: filters.EmptyVTTContent,
		},
    {
			name:           "when PreventHTTPStatusError filter is not enabled, should passthrough vtt content",
			url:            "phe(false)/aHR0cHM6Ly8wODc2M2JmMGIxZ2IuYWlyc3BhY2UtY2RuLmNic2l2aWRlby5jb20vbXR2LWVtYS11ay1obHMvbWFzdGVyLzQwNC5tM3U4.vtt",
			auth:           "authenticate-me",
			mockResp:       default200Response(filters.EmptyVTTContent),
			expectStatus:   200,
			expectManifest: filters.EmptyVTTContent,
		},
    {
			name:           "when requesting vtt, should passthrough vtt content",
			url:            "/aHR0cHM6Ly8wODc2M2JmMGIxZ2IuYWlyc3BhY2UtY2RuLmNic2l2aWRlby5jb20vbXR2LWVtYS11ay1obHMvbWFzdGVyLzQwNC5tM3U4.vtt",
			auth:           "authenticate-me",
			mockResp:       default200Response(filters.EmptyVTTContent),
			expectStatus:   200,
			expectManifest: filters.EmptyVTTContent,
		},
	}

	for _, tc := range tests {
		c := testConfig(test.MockClient(tc.mockResp))
		handler := LoadHandler(c)
		// set req + response recorder and serve it
		req := getRequest(tc.url, t)
		req.Header.Set("x-bakery-origin-token", tc.auth)
		rec := getResponseRecorder()
		handler.ServeHTTP(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		if res.StatusCode != tc.expectStatus {
			t.Errorf("expected status 500; got %v", res.StatusCode)
		}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}

		got := string(body)
		if !cmp.Equal(got, tc.expectManifest) {
			t.Errorf("Wrong error returned\ngot %v\nexpected: %v\ndiff: %v",
				got, tc.expectManifest, cmp.Diff(got, tc.expectManifest))
		}
	}
}
