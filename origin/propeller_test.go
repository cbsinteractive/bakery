package origin

import (
	"testing"

	propeller "github.com/cbsinteractive/propeller-go/client"
)

func TestPropeller_getChannelURL(t *testing.T) {
	getChannel := func(ads bool, captions bool, play string) propeller.Channel {
		return propeller.Channel{
			Ads:         ads,
			AdsURL:      "some-ad-url.com",
			Captions:    captions,
			CaptionsURL: "some-caption-url.com",
			PlaybackURL: play,
		}
	}

	tests := []struct {
		name        string
		channels    []propeller.Channel
		expectURL   string
		expectError bool
	}{
		{
			name: "When ads are set, ad url is returned regardless of other values",
			channels: []propeller.Channel{
				getChannel(true, false, "who cares"),
				getChannel(true, true, "who cares again"),
			},
			expectURL: "some-ad-url.com",
		},
		{
			name: "When ads are false and captions are set, ad url is returned regardless of other values",
			channels: []propeller.Channel{
				getChannel(false, true, "who cares"),
				getChannel(false, true, "who cares again"),
			},
			expectURL: "some-caption-url.com",
		},
		{
			name: "When ads and captions are NOT set, playback url is returned",
			channels: []propeller.Channel{
				getChannel(false, false, "playback-url.com"),
			},
			expectURL: "playback-url.com",
		},
		{
			name: "When ads, captions, and playbaclURL are NOT set, error is thrown",
			channels: []propeller.Channel{
				getChannel(false, false, ""),
			},
			expectURL:   "",
			expectError: true,
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

				if tc.expectURL != u {
					t.Errorf("Wrong playback url: expect: %q, got %q", tc.expectURL, u)
				}
			}
		})

	}
}

func TestPropeller_getClipURL(t *testing.T) {
	getClip := func(status string, desc string, play string) propeller.Clip {
		return propeller.Clip{
			Status:            status,
			StatusDescription: desc,
			PlaybackURL:       play,
		}
	}

	tests := []struct {
		name        string
		clips       []propeller.Clip
		expectURL   string
		expectError bool
	}{
		{
			name: "When status is created, expect playback url",
			clips: []propeller.Clip{
				getClip("created", "", "playback-url.com"),
			},
			expectURL: "playback-url.com",
		},
		{
			name: "When status is not created, expect error",
			clips: []propeller.Clip{
				getClip("pending", "", "who cares"),
				getClip("error", "some failure description", "who cares again"),
			},
			expectURL:   "",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for _, clip := range tc.clips {
				u, err := getClipURL(clip)

				// pending and error status clips should return error
				if err != nil && !tc.expectError {
					t.Errorf("getClipURL() didn't expect an error, got %v", err)
				} else if err == nil && tc.expectError {
					t.Errorf("getClipURL() expected error, got nil")
				}

				if tc.expectURL != u {
					t.Errorf("Wrong playback url: expect: %q, got %q", tc.expectURL, u)
				}
			}
		})

	}
}
