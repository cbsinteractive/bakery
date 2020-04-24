package handlers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestHandler_ErrorResponse(t *testing.T) {

	tests := []struct {
		name         string
		url          string
		mockResp     func(req *http.Request) (*http.Response, error)
		expectErr    ErrorResponse
		expectStatus int
	}{
		{
			name: "when manifest returns 4xx, expect 500  w/ err msg reflecting origin status code",
			url:  "origin/some/path/to/master.m3u8",
			mockResp: func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 403,
					Body:       ioutil.NopCloser(bytes.NewBufferString("")),
				}, nil
			},
			expectStatus: 500,
			expectErr: ErrorResponse{
				Message: "failed fetching manifest",
				Errors: map[string][]string{
					"fetching manifest": []string{"returning http status of 403"},
				},
			},
		},
		{
			name:         "when request is made with bad filters, expect error from parser",
			url:          "/b(10000,10)/origin/some/path/to/master.mpd",
			mockResp:     default200Response(),
			expectStatus: 500,
			expectErr: ErrorResponse{
				Message: "failed parsing filters",
				Errors: map[string][]string{
					"Bitrate": []string{"invalid range for provided values", "( 10000, 10 )"},
				},
			},
		},
		{
			name:         "when propeller channel is passed with bad path, expect 500 status code w/ err msg reflecting origin configuration",
			url:          "propeller/master.m3u8",
			mockResp:     default200Response(),
			expectStatus: 500,
			expectErr: ErrorResponse{
				Message: "failed configuring origin",
				Errors: map[string][]string{
					"propeller origin": []string{"request format is not `/propeller/orgID/channelID.m3u8`"},
				},
			},
		},
		{
			name:         "when request is made without protocol, proper error response is thrown",
			url:          "/some/random/request",
			mockResp:     default200Response(),
			expectStatus: 400,
			expectErr: ErrorResponse{
				Message: "failed to select filter",
				Errors: map[string][]string{
					`unsupported protocol ""`: []string{},
				},
			},
		},
		{
			name:         "when request is made and bad HLS manifest is returned, expect error",
			url:          "origin/some/path/to/master.m3u8",
			mockResp:     default200Response(),
			expectStatus: 500,
			expectErr: ErrorResponse{
				Message: "failed to filter manifest",
				Errors: map[string][]string{
					"#EXTM3U absent": []string{},
				},
			},
		},
		{
			name:         "when request is made and bad MPD manifest is returned, expect error",
			url:          "origin/some/path/to/master.mpd",
			mockResp:     default200Response(),
			expectStatus: 500,
			expectErr: ErrorResponse{
				Message: "failed to filter manifest",
				Errors: map[string][]string{
					"EOF": []string{},
				},
			},
		},
		{
			name:         "when request is made and bad MPD manifest is returned, expect error",
			url:          "origin/some/path/to/master.mpd",
			mockResp:     default200Response(),
			expectStatus: 500,
			expectErr: ErrorResponse{
				Message: "failed to filter manifest",
				Errors: map[string][]string{
					"EOF": []string{},
				},
			},
		},
	}

	// set handler
	c, err := testConfig()
	if err != nil {
		t.Fatalf("error parsing test config")
	}
	handler := LoadHandler(c)

	for _, tc := range tests {
		//mock client response
		SetDoFuncReturns = tc.mockResp

		// set req + response recorder and serve it
		req := getRequest(tc.url, t)
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

		var got ErrorResponse
		json.Unmarshal(body, &got)
		if !cmp.Equal(got, tc.expectErr) {
			t.Errorf("Wrong error returned\ngot %v\nexpected: %v\ndiff: %v",
				got, tc.expectErr, cmp.Diff(got, tc.expectErr))
		}
	}

}
