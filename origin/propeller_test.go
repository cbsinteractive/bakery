package origin

import (
	"testing"

	test "github.com/cbsinteractive/bakery/tests"
	propeller "github.com/cbsinteractive/propeller-go/client"
	"github.com/google/go-cmp/cmp"
)

func getChannel(ads bool, captions bool, play string, status string) propeller.Channel {
	return propeller.Channel{
		Ads:         ads,
		AdsURL:      "some-ad-url.com",
		Captions:    captions,
		CaptionsURL: "some-caption-url.com",
		PlaybackURL: play,
		Status:      status,
	}
}

func getClip(status string, desc string, play string) propeller.Clip {
	return propeller.Clip{
		Status:            status,
		StatusDescription: desc,
		PlaybackURL:       play,
	}
}

func mockChannelResp(ads, captions bool, play, status string) func(*propeller.Client, string, string) (string, error) {
	return func(*propeller.Client, string, string) (string, error) {
		return getChannelURL(getChannel(ads, captions, play, status))
	}
}

func mockClipResp(status string, desc string, play string) func(*propeller.Client, string, string) (string, error) {
	return func(*propeller.Client, string, string) (string, error) {
		return getClipURL(getClip(status, desc, play))
	}
}

func TestPropeller_NewPropeller(t *testing.T) {
	tests := []struct {
		name      string
		fetch     fetchURL
		expected  Origin
		expectErr bool
	}{
		{
			name:     "when creating new propeller channel, expect playbackURL in config",
			fetch:    mockChannelResp(false, false, "playbackurl.com", "running"),
			expected: &Propeller{URL: "playbackurl.com"},
		},
		{
			name:     "when creating new propeller channel, expect playbackURL in config",
			fetch:    mockClipResp("created", "", "playbackurl.com"),
			expected: &Propeller{URL: "playbackurl.com"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := testConfig(test.FakeClient{})
			got, err := NewPropeller(c, "orgID", "endpointID", tc.fetch)
			if err != nil && !tc.expectErr {
				t.Errorf("NewPropeller() didnt expect an error to be returned, got: %v", err)
				return
			} else if err == nil && tc.expectErr {
				t.Error("NewPropeller() expected an error, got nil")
				return
			}

			if !cmp.Equal(got, tc.expected) {
				t.Errorf("Wrong Propeller Origin returned\ngot %v\nexpected: %v\ndiff: %v",
					got, tc.expected, cmp.Diff(got, tc.expected))
			}

		})

	}
}

func TestPropeller_getChannelURL(t *testing.T) {
	tests := []struct {
		name        string
		channels    []propeller.Channel
		expectURL   string
		expectError bool
		errStr      string
	}{
		{
			name: "When ads are set, ad url is returned when channel is running and regardless of other values",
			channels: []propeller.Channel{
				getChannel(true, false, "who cares", "running"),
				getChannel(true, true, "who cares again", "running"),
			},
			expectURL: "some-ad-url.com",
		},
		{
			name: "When ads are false and captions are set, ad url is returned regardless of other values",
			channels: []propeller.Channel{
				getChannel(false, true, "who cares", "running"),
				getChannel(false, true, "who cares again", "running"),
			},
			expectURL: "some-caption-url.com",
		},
		{
			name: "When ads and captions are NOT set, playback url is returned",
			channels: []propeller.Channel{
				getChannel(false, false, "playback-url.com", "running"),
			},
			expectURL: "playback-url.com",
		},

		{
			name: "When ads are set but channel isn't running, return playbackURL",
			channels: []propeller.Channel{
				getChannel(true, false, "playback-url.com", "stopping"),
				getChannel(true, false, "playback-url.com", "ready"),
				getChannel(true, false, "playback-url.com", "pending"),
				getChannel(true, false, "playback-url.com", "starting"),
			},
			expectURL: "playback-url.com",
		},
		{
			name: "When ads, captions, and playbackURL are NOT set, error is thrown",
			channels: []propeller.Channel{
				getChannel(false, false, "", "running"),
			},
			expectURL:   "",
			expectError: true,
			errStr:      "parsing channel url: Channel not ready",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for _, channel := range tc.channels {
				u, err := getChannelURL(channel)

				// pending and error status channels should return error
				if err != nil && !tc.expectError {
					t.Errorf("getChannelURL() didn't expect an error, got %v", err)
				} else if err == nil && tc.expectError {
					t.Errorf("getChannelURL() expected error, got nil")
				}

				if tc.expectError && err.Error() != tc.errStr {
					t.Errorf("Wrong error string. expected: %q, got %q", tc.errStr, err.Error())
				}

				if tc.expectURL != u {
					t.Errorf("Wrong playback url: expect: %q, got %q", tc.expectURL, u)
				}
			}
		})

	}
}

func TestPropeller_getClipURL(t *testing.T) {
	tests := []struct {
		name        string
		clip        propeller.Clip
		expectURL   string
		expectError bool
		errStr      string
	}{
		{
			name:      "When status is created, expect playback url",
			clip:      getClip("created", "", "playback-url.com"),
			expectURL: "playback-url.com",
		},
		{
			name:        "When status is created, but no playbackURL available, expect error not ready",
			clip:        getClip("created", "", ""),
			expectURL:   "",
			expectError: true,
			errStr:      "clip status: not ready",
		},
		{
			name:        "When status is error, expect error",
			clip:        getClip("error", "some failure description", "who cares again"),
			expectURL:   "",
			expectError: true,
			errStr:      "parsing clip url: some failure description",
		},
		{
			name:        "When status is pending, expect clip not ready error",
			clip:        getClip("pending", "", "who cares"),
			expectURL:   "",
			expectError: true,
			errStr:      "parsing clip url: Clip not ready",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u, err := getClipURL(tc.clip)

			if err != nil && !tc.expectError {
				t.Errorf("getClipURL() didn't expect an error, got %v", err)
			} else if err == nil && tc.expectError {
				t.Errorf("getClipURL() expected error, got nil")
			}

			if tc.expectError && err.Error() != tc.errStr {
				t.Errorf("Wrong error string. expected: %q, got %q", tc.errStr, err.Error())
			}

			if tc.expectURL != u {
				t.Errorf("Wrong playback url: expect: %q, got %q", tc.expectURL, u)
			}
		})

	}
}

func TestPropeller_extractID(t *testing.T) {
	tests := []struct {
		name       string
		manifest   []string
		expectedID []string
	}{
		{
			name: "When extracting ids from manifest path, return correct id",
			manifest: []string{
				"id.m3u8",
				"id",
			},
			expectedID: []string{
				"id",
				"id",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for i, m := range tc.manifest {
				got := extractID(m)

				if got != tc.expectedID[i] {
					t.Errorf("Wrong ID reurned. expect: %v, got %v", tc.expectedID[i], got)
				}
			}
		})

	}
}
