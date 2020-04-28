package origin

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/cbsinteractive/bakery/config"
	test "github.com/cbsinteractive/bakery/tests"
	propeller "github.com/cbsinteractive/propeller-go/client"
	"github.com/google/go-cmp/cmp"
)

func testConfig(fc test.FakeClient) config.Config {
	timeout := 5 * time.Second

	return config.Config{
		Listen:      "8080",
		LogLevel:    "panic",
		OriginHost:  "http://localhost:8080",
		Hostname:    "hostname",
		OriginToken: "authenticate-me",
		Client: config.Client{
			Context:    context.Background(),
			Timeout:    timeout,
			Tracer:     nil,
			HTTPClient: fc,
		},
		Propeller: config.Propeller{
			Client: propeller.Client{
				Timeout:    timeout,
				HTTPClient: fc,
			},
		},
	}
}

func getMockResp(code int, msg string) func(*http.Request) (*http.Response, error) {
	return func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: code,
			Body:       ioutil.NopCloser(bytes.NewBufferString(msg)),
		}, nil
	}
}

func TestOrigin_FetchManifest(t *testing.T) {
	relativeURL, err := url.Parse("/path/to/manifest/master.m3u8")
	absoluteURL, err := url.Parse("https://origin.com/path/to/manifest/master.m3u8")
	if err != nil {
		t.Errorf("Unable to make test urls")
	}

	tests := []struct {
		name      string
		origin    Origin
		mockResp  func(*http.Request) (*http.Response, error)
		expectStr string
		expectErr bool
	}{
		{
			name:      "when fetching propeller channel, return response message if code < 300",
			origin:    &Propeller{URL: "https://propeller-playback-url.m3u8"},
			mockResp:  getMockResp(200, "OK"),
			expectStr: "OK",
		},
		{
			name:      "when fetching origin manifest, return response message if code < 300",
			origin:    &DefaultOrigin{Origin: "https://origin.com", URL: *relativeURL},
			mockResp:  getMockResp(200, "OK"),
			expectStr: "OK",
		},
		{
			name:      "when fetching origin manifest, expect if code > 300",
			origin:    &DefaultOrigin{Origin: "https://origin.com", URL: *absoluteURL},
			mockResp:  getMockResp(404, "NotFound"),
			expectErr: true,
		},
		{
			name:      "when fetching propeller channel, expect if code > 300",
			origin:    &Propeller{URL: "https://propeller-playback-url.m3u8"},
			mockResp:  getMockResp(404, "NotFound"),
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := testConfig(test.MockClient(tc.mockResp))

			got, err := tc.origin.FetchManifest(c.Client)

			if err != nil && !tc.expectErr {
				t.Errorf("Configure() didnt expect an error to be returned, got: %v", err)
				return
			} else if err == nil && tc.expectErr {
				t.Error("Configure() expected an error, got nil")
				return
			}

			if got != tc.expectStr {
				t.Errorf("Wrong response: expect: %q, got %q", tc.expectStr, got)
			}
		})

	}
}

func TestOrigin_GetPlaybackURL(t *testing.T) {
	relativeURL, err := url.Parse("/path/to/manifest/master.m3u8")
	absoluteURL, err := url.Parse("https://origin.com/path/to/manifest/master.m3u8")
	if err != nil {
		t.Errorf("Unable to make test urls")
	}

	tests := []struct {
		name        string
		origin      Origin
		expectedURL string
	}{
		{
			name:        "when origin is of type propeller, return propeller playback URL",
			origin:      &Propeller{URL: "https://propeller-playback-url.m3u8"},
			expectedURL: "https://propeller-playback-url.m3u8",
		},
		{
			name:        "when origin is of type default with relative url, return full playback URL",
			origin:      &DefaultOrigin{Origin: "https://origin.com", URL: *relativeURL},
			expectedURL: "https://origin.com/path/to/manifest/master.m3u8",
		},
		{
			name:        "when origin is of type default with absolute url, return full playback URL",
			origin:      &DefaultOrigin{Origin: "https://origin.com", URL: *absoluteURL},
			expectedURL: "https://origin.com/path/to/manifest/master.m3u8",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.origin.GetPlaybackURL()

			if got != tc.expectedURL {
				t.Errorf("Wrong playback url: expect: %q, got %q", tc.expectedURL, got)
			}
		})

	}
}

func TestOrigin_Configure(t *testing.T) {
	absTestURL, err := url.Parse("https://stream/some/path/request/to/master.m3u8")
	relTestURL, err := url.Parse("/some/path/request/to/master.m3u8")
	if err != nil {
		t.Error("Unable to make test urls")
	}

	tests := []struct {
		name      string
		path      string
		c         config.Config
		expected  Origin
		expectErr bool
	}{
		{
			name:      "when origin is of type propeller in wrong format, return error",
			path:      "/propeller/chanID.m3u8",
			c:         config.Config{LogLevel: "panic"},
			expected:  &Propeller{},
			expectErr: true,
		},
		{
			name:      "when origin is of type propeller in wrong format, return error",
			path:      "/propeller/chanID.m3u8",
			c:         config.Config{LogLevel: "panic"},
			expected:  &Propeller{},
			expectErr: true,
		},
		{
			name:     "when origin path is at root but not base64 encoded, return default origin type",
			path:     fmt.Sprintf("/%v.m3u8", base64.RawURLEncoding.EncodeToString([]byte(absTestURL.String()))),
			c:        config.Config{LogLevel: "panic", OriginHost: "host"},
			expected: &DefaultOrigin{Origin: "host", URL: *absTestURL},
		},

		{
			name:     "when origin path is at root but not base64 encoded, return default origin type",
			path:     relTestURL.String(),
			c:        config.Config{LogLevel: "panic", OriginHost: "host"},
			expected: &DefaultOrigin{Origin: "host", URL: *relTestURL},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Configure(tc.c, tc.path)

			if err != nil && !tc.expectErr {
				t.Errorf("Configure() didnt expect an error to be returned, got: %v", err)
				return
			} else if err == nil && tc.expectErr {
				t.Error("Configure() expected an error, got nil")
				return
			}

			if !cmp.Equal(got, tc.expected) {
				t.Errorf("Wrong Origin returned\ngot %v\nexpected: %v\ndiff: %v",
					got, tc.expected, cmp.Diff(got, tc.expected))
			}

		})

	}
}
