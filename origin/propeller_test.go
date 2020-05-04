package origin

import (
	"errors"
	"testing"

	test "github.com/cbsinteractive/bakery/tests"
	propeller "github.com/cbsinteractive/propeller-go/client"
	"github.com/google/go-cmp/cmp"
)

// mocks

type mockUrlGetter struct {
	url string
	err error
}

func (mock *mockUrlGetter) GetURL(c propellerClient) (string, error) {
	return mock.url, mock.err
}

type mockPropellerClient struct {
	// mock return value of GetChannel()
	getChannel      propeller.Channel
	getChannelError error

	// mock return value of GetClip()
	getClip      propeller.Clip
	getClipError error

	// record last method call
	getChannelCalled map[string]string
	getClipCalled    map[string]string
}

func (mock *mockPropellerClient) GetChannel(orgID string, channelID string) (propeller.Channel, error) {
	mock.getChannelCalled = map[string]string{"orgID": orgID, "channelID": channelID}
	return mock.getChannel, mock.getChannelError
}
func (mock *mockPropellerClient) GetClip(orgID string, clipID string) (propeller.Clip, error) {
	mock.getClipCalled = map[string]string{"orgID": orgID, "clipID": clipID}
	return mock.getClip, mock.getClipError
}

// tests

func TestPropeller_NewPropeller(t *testing.T) {
	tests := []struct {
		name      string
		getter    urlGetter
		expected  Origin
		expectErr bool
	}{
		{
			name:     "when creating new propeller channel use urlGetter to get url",
			getter:   &mockUrlGetter{url: "playbackurl.com"},
			expected: &Propeller{URL: "playbackurl.com"},
		},
		{
			name:      "when creating new propeller channel return error if urlGetter fails",
			getter:    &mockUrlGetter{err: errors.New("ops")},
			expected:  &Propeller{},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := testConfig(test.FakeClient{})
			got, err := NewPropeller(c, "orgID", "endpointID", tc.getter)
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

func TestPropeller_channelURLGetter(t *testing.T) {
	tests := []struct {
		name         string
		channels     []propeller.Channel
		expectURL    string
		expectErrStr string
	}{
		{
			name: "When ads are set, ad url is returned when channel is running and regardless of other values",
			channels: []propeller.Channel{
				{Ads: true, AdsURL: "ad-url.com", Status: "running"},
				{Ads: true, AdsURL: "ad-url.com", Status: "running", Captions: true, CaptionsURL: "caption-url.com"},
				{Ads: true, AdsURL: "ad-url.com", Status: "running", PlaybackURL: "playback.com"},
			},
			expectURL: "ad-url.com",
		},
		{
			name: "When ads are false and captions are set, captions url is returned regardless of other values",
			channels: []propeller.Channel{
				{Captions: true, CaptionsURL: "caption-url.com"},
				{Captions: true, CaptionsURL: "caption-url.com", AdsURL: "some-ad-url.com", PlaybackURL: "playback.com"},
			},
			expectURL: "caption-url.com",
		},
		{
			name: "When ads and captions are NOT set, playback url is returned",
			channels: []propeller.Channel{
				{PlaybackURL: "playback-url.com"},
			},
			expectURL: "playback-url.com",
		},
		{
			name: "When ads are set but channel isn't running, return playbackURL",
			channels: []propeller.Channel{
				{Ads: true, AdsURL: "ad-url.com", PlaybackURL: "playback-url.com", Status: "stopping"},
				{Ads: true, AdsURL: "ad-url.com", PlaybackURL: "playback-url.com", Status: "ready"},
				{Ads: true, AdsURL: "ad-url.com", PlaybackURL: "playback-url.com", Status: "pending"},
				{Ads: true, AdsURL: "ad-url.com", PlaybackURL: "playback-url.com", Status: "starting"},
			},
			expectURL: "playback-url.com",
		},
		{
			name: "When ads, captions, and playbackURL are NOT set, error is thrown",
			channels: []propeller.Channel{
				{},
			},
			expectErrStr: "parsing channel url: Channel not ready",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for _, channel := range tc.channels {
				client := &mockPropellerClient{getChannel: channel}
				getter := &channelURLGetter{orgID: "org", channelID: "ch123"}
				u, err := getter.GetURL(client)

				if err != nil && tc.expectErrStr == "" {
					t.Errorf("channelURLGetter.GetURL() didn't expect an error, got %v", err)
				} else if err == nil && tc.expectErrStr != "" {
					t.Errorf("channelURLGetter.GetURL() expected error, got nil")
				} else if err != nil && err.Error() != tc.expectErrStr {
					t.Errorf("channelURLGetter.GetURL() got wrong error. expected: %q, got %q", tc.expectErrStr, err.Error())
				}
				if tc.expectURL != u {
					t.Errorf("channelURLGetter.GetURL() got wrong playback url: expect: %q, got %q", tc.expectURL, u)
				}
				if client.getChannelCalled["orgID"] != "org" || client.getChannelCalled["channelID"] != "ch123" {
					t.Errorf("channelURLGetter.GetURL() called client.GetChannel() with wrong arguments: %#v", client.getChannelCalled)
				}
			}
		})
	}
}

func TestPropeller_channelURLGetter_getArchive(t *testing.T) {
	tests := []struct {
		name         string
		getter       urlGetter
		client       *mockPropellerClient
		expectURL    string
		expectErrStr string
	}{
		// channelURLGetter
		{
			name:   "channelURLGetter should get clip archive url if channel not found",
			getter: &channelURLGetter{orgID: "org", channelID: "ch123"},
			client: &mockPropellerClient{
				getChannelError: propeller.StatusError{Code: 404},
				getClip:         propeller.Clip{PlaybackURL: "archive-url.com"},
			},
			expectURL: "archive-url.com",
		},
		{
			name:   "channelURLGetter should return error if fail to get archive when channel not found",
			getter: &channelURLGetter{orgID: "org", channelID: "ch123"},
			client: &mockPropellerClient{
				getChannelError: propeller.StatusError{Code: 404},
				getClipError:    errors.New("fail to get archive"),
			},
			expectErrStr: "fetching clip: fail to get archive",
		},
		// outputURLGetter
		{
			name:   "outputURLGetter should get clip archive url if channel not found",
			getter: &outputURLGetter{orgID: "org", channelID: "ch123"},
			client: &mockPropellerClient{
				getChannelError: propeller.StatusError{Code: 404},
				getClip:         propeller.Clip{PlaybackURL: "archive-url.com"},
			},
			expectURL: "archive-url.com",
		},
		{
			name:   "outputURLGetter should return error if fail to get archive when channel not found",
			getter: &outputURLGetter{orgID: "org", channelID: "ch123"},
			client: &mockPropellerClient{
				getChannelError: propeller.StatusError{Code: 404},
				getClipError:    errors.New("fail to get archive"),
			},
			expectErrStr: "fetching clip: fail to get archive",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u, err := tc.getter.GetURL(tc.client)
			if err != nil && tc.expectErrStr == "" {
				t.Errorf("returned unexpected error: %q", err)
			} else if err == nil && tc.expectErrStr != "" {
				t.Errorf("expected error, got nil")
			} else if err != nil && err.Error() != tc.expectErrStr {
				t.Errorf("got wrong error. expected: %q, got %q", tc.expectErrStr, err.Error())
			}
			if tc.expectURL != u {
				t.Errorf("got wrong playback url: expect %q, got %q", u, tc.expectURL)
			}
			if tc.client.getChannelCalled["orgID"] != "org" || tc.client.getChannelCalled["channelID"] != "ch123" {
				t.Errorf("client.GetChannel() called with wrong arguments: %#v", tc.client.getChannelCalled)
			}
			if tc.client.getClipCalled["orgID"] != "org" || tc.client.getClipCalled["clipID"] != "ch123-archive" {
				t.Errorf("client.GetClip() called with wrong arguments: %#v", tc.client.getClipCalled)
			}
		})
	}
}

func TestPropeller_outputURLGetter(t *testing.T) {
	tests := []struct {
		name         string
		channels     []propeller.Channel
		expectURL    string
		expectErrStr string
	}{
		{
			name: "When ads are set, ad url is returned when channel is running and regardless of other values",
			channels: []propeller.Channel{
				{Ads: true, Status: "running", Outputs: []propeller.ChannelOutput{
					{ID: "out123", AdsURL: "ad-url.com"},
				}},
				{Ads: true, Status: "running", Captions: true, Outputs: []propeller.ChannelOutput{
					{ID: "out123", AdsURL: "ad-url.com", CaptionsURL: "caption-url.com"},
				}},
				{Ads: true, Status: "running", Outputs: []propeller.ChannelOutput{
					{ID: "out123", AdsURL: "ad-url.com", PlaybackURL: "playback.com"},
				}},
			},
			expectURL: "ad-url.com",
		},
		{
			name: "When ads are false and captions are set, captions url is returned regardless of other values",
			channels: []propeller.Channel{
				{Captions: true, Outputs: []propeller.ChannelOutput{
					{ID: "out123", CaptionsURL: "caption-url.com"},
				}},
				{Captions: true, Outputs: []propeller.ChannelOutput{
					{ID: "out123", CaptionsURL: "caption-url.com", AdsURL: "some-ad-url.com", PlaybackURL: "playback.com"},
				}},
			},
			expectURL: "caption-url.com",
		},
		{
			name: "When ads and captions are NOT set, playback url is returned",
			channels: []propeller.Channel{
				{Outputs: []propeller.ChannelOutput{
					{ID: "out123", PlaybackURL: "playback-url.com"},
				}},
			},
			expectURL: "playback-url.com",
		},
		{
			name: "When ads are set but channel isn't running, return playbackURL",
			channels: []propeller.Channel{
				{Ads: true, Status: "stopping", Outputs: []propeller.ChannelOutput{
					{ID: "out123", AdsURL: "ad-url.com", PlaybackURL: "playback-url.com"},
				}},
				{Ads: true, Status: "ready", Outputs: []propeller.ChannelOutput{
					{ID: "out123", AdsURL: "ad-url.com", PlaybackURL: "playback-url.com"},
				}},
				{Ads: true, Status: "pending", Outputs: []propeller.ChannelOutput{
					{ID: "out123", AdsURL: "ad-url.com", PlaybackURL: "playback-url.com"},
				}},
				{Ads: true, Status: "starting", Outputs: []propeller.ChannelOutput{
					{ID: "out123", AdsURL: "ad-url.com", PlaybackURL: "playback-url.com"},
				}},
			},
			expectURL: "playback-url.com",
		},
		{
			name: "When ads, captions, and playbackURL are NOT set, error is thrown",
			channels: []propeller.Channel{
				{ID: "ch123", Outputs: []propeller.ChannelOutput{{}}},
			},
			expectErrStr: "finding channel output: Propeller Channel ch123 has no output with ID out123",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for _, channel := range tc.channels {
				client := &mockPropellerClient{getChannel: channel}
				getter := &outputURLGetter{orgID: "org", channelID: "ch123", outputID: "out123"}
				u, err := getter.GetURL(client)

				if err != nil && tc.expectErrStr == "" {
					t.Errorf("outputURLGetter.GetURL() didn't expect an error, got %q", err)
				} else if err == nil && tc.expectErrStr != "" {
					t.Errorf("outputURLGetter.GetURL() expected error, got nil")
				} else if err != nil && err.Error() != tc.expectErrStr {
					t.Errorf("outputURLGetter.GetURL() got wrong error. expected: %q, got %q", tc.expectErrStr, err.Error())
				}
				if tc.expectURL != u {
					t.Errorf("outputURLGetter.GetURL() got wrong playback url: expect: %q, got %q", tc.expectURL, u)
				}
				if client.getChannelCalled["orgID"] != "org" || client.getChannelCalled["channelID"] != "ch123" {
					t.Errorf("outputURLGetter.GetURL() called client.GetChannel() with wrong arguments: %#v", client.getChannelCalled)
				}
			}
		})
	}
}

func TestPropeller_clipURLGetter(t *testing.T) {
	tests := []struct {
		name         string
		clip         propeller.Clip
		expectURL    string
		expectErrStr string
	}{
		{
			name:      "When status is created, expect playback url",
			clip:      propeller.Clip{Status: "created", PlaybackURL: "playback-url.com"},
			expectURL: "playback-url.com",
		},
		{
			name:         "When status is created, but no playbackURL available, expect error not ready",
			clip:         propeller.Clip{Status: "created"},
			expectErrStr: "clip status: not ready",
		},
		{
			name:         "When status is error, expect error",
			clip:         propeller.Clip{Status: "error", StatusDescription: "some failure description", PlaybackURL: "who-cares.com"},
			expectErrStr: "parsing clip url: some failure description",
		},
		{
			name:         "When status is pending, expect clip not ready error",
			clip:         propeller.Clip{Status: "pending", PlaybackURL: "who cares"},
			expectErrStr: "parsing clip url: Clip not ready",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := &mockPropellerClient{getClip: tc.clip}
			getter := &clipURLGetter{orgID: "org", clipID: "cl123"}
			u, err := getter.GetURL(client)

			if err != nil && tc.expectErrStr == "" {
				t.Errorf("clipURLGetter.GetURL() didn't expect an error, got %v", err)
			} else if err == nil && tc.expectErrStr != "" {
				t.Errorf("clipURLGetter.GetURL() expected error, got nil")
			} else if err != nil && err.Error() != tc.expectErrStr {
				t.Errorf("clipURLGetter.GetURL() got wrong error. expected: %q, got %q", tc.expectErrStr, err.Error())
			}
			if tc.expectURL != u {
				t.Errorf("clipURLGetter.GetURL() got Wrong playback url: expect: %q, got %q", tc.expectURL, u)
			}
			if client.getClipCalled["orgID"] != "org" || client.getClipCalled["clipID"] != "cl123" {
				t.Errorf("clipURLGetter.GetURL() called client.GetClip() with wrong arguments: %#v", client.getClipCalled)
			}
		})
	}
}
