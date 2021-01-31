package parsers

import (
	"encoding/json"
	"math"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestURLParseUrl(t *testing.T) {
	tests := []struct {
		name                 string
		input                string
		expectedFilters      MediaFilters
		expectedManifestPath string
		expectedErr          bool
	}{
		{
			"one content type",
			"/ct(text)/test.m3u8",
			MediaFilters{
				ContentTypes: []string{"text"},
				Protocol:     ProtocolHLS,
			},
			"/test.m3u8",
			false,
		},
		{
			"bad content type",
			"/ct(content)/test.m3u8",
			MediaFilters{},
			"",
			true,
		},
		{
			"multiple content types",
			"/ct(audio,video)/test.m3u8",
			MediaFilters{
				ContentTypes: []string{"audio", "video"},
				Protocol:     ProtocolHLS,
			},
			"/test.m3u8",
			false,
		},
		{
			"one video type",
			"/v(hdr10)/test.m3u8",
			MediaFilters{
				Videos: NestedFilters{
					Codecs: []string{"hev1.2", "hvc1.2"},
				},
				Protocol: ProtocolHLS,
			},
			"/test.m3u8",
			false,
		},
		{
			"bad video type",
			"/v(codec)/test.m3u8",
			MediaFilters{},
			"",
			true,
		},
		{
			"two video types",
			"/v(hdr10,hvc)/test.m3u8",
			MediaFilters{
				Videos: NestedFilters{
					Codecs: []string{"hev1.2", "hvc1.2", "hvc"},
				},
				Protocol: ProtocolHLS,
			},
			"/test.m3u8",
			false,
		},
		{
			"two video types and two audio types",
			"/v(hdr10,hvc)/a(mp4a)/test.m3u8",
			MediaFilters{
				Videos: NestedFilters{
					Codecs: []string{"hev1.2", "hvc1.2", "hvc"},
				},
				Audios: NestedFilters{
					Codecs: []string{"mp4a"},
				},
				Protocol: ProtocolHLS,
			},
			"/test.m3u8",
			false,
		},
		{
			"videos, audio, captions and bitrate range",
			"/v(hdr10,hvc)/a(mp4a,l(pt-BR,en),b(10,20))/b(100,4000)/test.m3u8",
			MediaFilters{
				Videos: NestedFilters{
					Codecs: []string{"hev1.2", "hvc1.2", "hvc"},
					Bitrate: &Bitrate{
						Max: 4000,
						Min: 100,
					},
				},
				Audios: NestedFilters{
					Codecs:   []string{"mp4a"},
					Language: []string{"pt-BR", "en"},
					Bitrate: &Bitrate{
						Max: 20,
						Min: 10,
					},
				},
				Protocol: ProtocolHLS,
			},
			"/test.m3u8",
			false,
		},
		{
			"bitrate range doesn't override nested audio and video filter",
			"/v(hdr10,hvc,b(100,500))/a(mp4a,l(pt-BR,en),b(10,20))/b(100,4000)/test.m3u8",
			MediaFilters{
				Videos: NestedFilters{
					Codecs: []string{"hev1.2", "hvc1.2", "hvc"},
					Bitrate: &Bitrate{
						Max: 500,
						Min: 100,
					},
				},
				Audios: NestedFilters{
					Codecs:   []string{"mp4a"},
					Language: []string{"pt-BR", "en"},
					Bitrate: &Bitrate{
						Max: 20,
						Min: 10,
					},
				},
				Protocol: ProtocolHLS,
			},
			"/test.m3u8",
			false,
		},
		{
			"bitrate range with minimum bitrate only",
			"/b(100,)/test.m3u8",
			MediaFilters{
				Videos: NestedFilters{
					Bitrate: &Bitrate{
						Max: math.MaxInt32,
						Min: 100,
					},
				},
				Audios: NestedFilters{
					Bitrate: &Bitrate{
						Max: math.MaxInt32,
						Min: 100,
					},
				},
				Protocol: ProtocolHLS,
			},
			"/test.m3u8",
			false,
		},
		{
			"bitrate range with maximum bitrate only",
			"/b(,3000)/test.m3u8",
			MediaFilters{
				Videos: NestedFilters{
					Bitrate: &Bitrate{
						Max: 3000,
						Min: 0,
					},
				},
				Audios: NestedFilters{
					Bitrate: &Bitrate{
						Max: 3000,
						Min: 0,
					},
				},
				Protocol: ProtocolHLS,
			},
			"/test.m3u8",
			false,
		},
		{
			"bitrate range with minimum greater than maximum throws error",
			"/b(30000,3000)/test.m3u8",
			MediaFilters{},
			"",
			true,
		},
		{
			"bitrate range with minimum equal to maximum throws error",
			"/b(3000,3000)/test.m3u8",
			MediaFilters{},
			"",
			true,
		},
		{
			"audio bitrate range with minimum equal to maximum throws error",
			"/a(b(1000,1000))/test.m3u8",
			MediaFilters{},
			"",
			true,
		},
		{
			"video bitrate range with minimum greater than maximum throws error",
			"/v(b(2000,1000))/test.m3u8",
			MediaFilters{},
			"",
			true,
		},
		{
			"audio bitrate range with inavlid, negative minimum",
			"/a(b(-100,1000))/test.m3u8",
			MediaFilters{},
			"",
			true,
		},
		{
			"video bitrate range with invalid, greater than math.MaxInt32 minimum",
			"/v(b(2147483648))/test.m3u8",
			MediaFilters{},
			"",
			true,
		},
		{
			"audio bitrate range with invalid, greater than math.MaxInt32 maximum",
			"/a(b(10,2147483648))/test.m3u8",
			MediaFilters{},
			"",
			true,
		},
		{
			"video bitrate range with invalid, negative maximum",
			"/v(b(10,-100))/test.m3u8",
			MediaFilters{},
			"",
			true,
		},
		{
			"bitrate range with invalid int for minimum throws error",
			"/b(one,-100)/test.m3u8",
			MediaFilters{},
			"",
			true,
		},
		{
			"bitrate range with invalid int for maximum throws error",
			"/b(100,million)/test.m3u8",
			MediaFilters{},
			"",
			true,
		},
		{
			"trim filter converts to milliseconds",
			"/t(100,1000)/path/to/test.m3u8",
			MediaFilters{
				Protocol: ProtocolHLS,
				Trim: &Trim{
					Start: 100,
					End:   1000,
				},
			},
			"/path/to/test.m3u8",
			false,
		},
		{
			"trim filter where start time is greater than end time throws error",
			"/t(10000,1000)/path/to/test.m3u8",
			MediaFilters{},
			"",
			true,
		},
		{
			"trim filter where start time and end time are equal throws error",
			"/t(10000,1000)/path/to/test.m3u8",
			MediaFilters{},
			"",
			true,
		},
		{
			"detect a single plugin for execution from url",
			"[plugin1]/some/path/master.m3u8",
			MediaFilters{
				Protocol: ProtocolHLS,
				Plugins:  []string{"plugin1"},
			},
			"/some/path/master.m3u8",
			false,
		},
		{
			"detect plugins for execution from url",
			"/v(hdr10,hvc)/[plugin1,plugin2,plugin3]/some/path/master.m3u8",
			MediaFilters{
				Videos: NestedFilters{
					Codecs: []string{"hev1.2", "hvc1.2", "hvc"},
				},
				Protocol: ProtocolHLS,
				Plugins:  []string{"plugin1", "plugin2", "plugin3"},
			},
			"/some/path/master.m3u8",
			false,
		},
		{
			"nested audio and video bitrate filters are properly detected",
			"/a(b(100,))/v(b(,5000))/test.m3u8",
			MediaFilters{
				Videos: NestedFilters{
					Bitrate: &Bitrate{
						Max: 5000,
					},
				},
				Audios: NestedFilters{
					Bitrate: &Bitrate{
						Min: 100,
						Max: math.MaxInt32,
					},
				},
				Protocol: ProtocolHLS,
			},
			"/test.m3u8",
			false,
		},
		{
			"nested codec and bitrate filters in audio",
			"/a(b(100,200),co(ac-3,mp4a))/test.m3u8",
			MediaFilters{
				Audios: NestedFilters{
					Bitrate: &Bitrate{
						Min: 100,
						Max: 200,
					},
					Codecs: []string{"ac-3", "mp4a"},
				},
				Protocol: ProtocolHLS,
			},
			"/test.m3u8",
			false,
		},
		{
			"nested codec and bitrate filters in video, plus overall bitrate filters",
			"/v(co(avc,hdr10),b(1000,2000))/test.m3u8",
			MediaFilters{
				Videos: NestedFilters{
					Bitrate: &Bitrate{
						Min: 1000,
						Max: 2000,
					},
					Codecs: []string{"avc", "hev1.2", "hvc1.2"},
				},
				Protocol: ProtocolHLS,
			},
			"/test.m3u8",
			false,
		},
		{
			"nested bitrate and old format of codec filter",
			"/a(mp4a,ac-3,b(0,10))/test.m3u8",
			MediaFilters{
				Audios: NestedFilters{
					Bitrate: &Bitrate{
						Max: 10,
					},
					Codecs: []string{"mp4a", "ac-3"},
				},
				Protocol: ProtocolHLS,
			},
			"/test.m3u8",
			false,
		},
		{
			"detect overall lang filter when passed in url, populate nested filters",
			"v(avc)/a(mp4a)/l(es)/path/here/with/master.m3u8",
			MediaFilters{
				Videos: NestedFilters{
					Codecs: []string{"avc"},
				},
				Audios: NestedFilters{
					Codecs:   []string{"mp4a"},
					Language: []string{"es"},
				},
				Captions: NestedFilters{
					Language: []string{"es"},
				},
				Protocol: ProtocolHLS,
			},
			"/path/here/with/master.m3u8",
			false,
		},
		{
			"detect nested lang filter when passed in via caption type",
			"v(avc)/a(mp4a)/c(l(es))/path/here/with/master.m3u8",
			MediaFilters{
				Videos: NestedFilters{
					Codecs: []string{"avc"},
				},
				Audios: NestedFilters{
					Codecs: []string{"mp4a"},
				},
				Captions: NestedFilters{
					Language: []string{"es"},
				},
				Protocol: ProtocolHLS,
			},
			"/path/here/with/master.m3u8",
			false,
		},
		{
			"detect caption type filter when passed in url",
			"c(wvtt)/path/here/with/master.m3u8",
			MediaFilters{
				Protocol: ProtocolHLS,
				Captions: NestedFilters{
					Codecs: []string{"wvtt"},
				},
			},
			"/path/here/with/master.m3u8",
			false,
		},
		{
			"bad nested codec filter throws error",
			"v(avc)/a(mp4a)/c(co(codec))/path/here/with/master.m3u8",
			MediaFilters{},
			"",
			true,
		},
		{
			"bad caption type filter throws error",
			"c(codec)/master.m3u8",
			MediaFilters{},
			"",
			true,
		},
		{
			"detect content type filter when passed in url",
			"ct(text,video)/path/here/with/master.m3u8",
			MediaFilters{
				Protocol:     ProtocolHLS,
				ContentTypes: []string{"text", "video"},
			},
			"/path/here/with/master.m3u8",
			false,
		},
		{
			"detect ads filter when passed in",
			"tags(ads)/path/here/with/master.m3u8",
			MediaFilters{
				Protocol: ProtocolHLS,
				Tags: &Tags{
					Ads: true,
				},
			},
			"/path/here/with/master.m3u8",
			false,
		},
		{
			"detect ads and iframe filter when passed in url with other nested filters",
			"v(avc,l(en))/tags(iframe,ads)/path/here/with/master.m3u8",
			MediaFilters{
				Videos: NestedFilters{
					Codecs:   []string{"avc"},
					Language: []string{"en"},
				},
				Protocol: ProtocolHLS,
				Tags: &Tags{
					Ads:    true,
					IFrame: true,
				},
			},
			"/path/here/with/master.m3u8",
			false,
		},

		{
			"detect iframe filter when passed in url",
			"tags(i-frame)/path/here/with/master.m3u8",
			MediaFilters{
				Protocol: ProtocolHLS,
				Tags: &Tags{
					IFrame: true,
				},
			},
			"/path/here/with/master.m3u8",
			false,
		},
		{
			"detect iframe filter when passed in url with other nested filters",
			"v(avc,l(en))/tags(iframe)/path/here/with/master.m3u8",
			MediaFilters{
				Videos: NestedFilters{
					Codecs:   []string{"avc"},
					Language: []string{"en"},
				},
				Protocol: ProtocolHLS,
				Tags: &Tags{
					IFrame: true,
				},
			},
			"/path/here/with/master.m3u8",
			false,
		},
		{
			"detect fps filter when passed in url",
			"fps(59.94)/path/here/to/master.m3u8",
			MediaFilters{
				FrameRate: []string{"59.94"},
				Protocol:  ProtocolHLS,
			},
			"/path/here/to/master.m3u8",
			false,
		},
		{
			"detect multiple values including fractions when fps filter is passed in url",
			"fps(60,30000:1001)/path/here/to/master.mpd",
			MediaFilters{
				FrameRate: []string{"60", "30000/1001"},
				Protocol:  ProtocolDASH,
			},
			"/path/here/to/master.mpd",
			false,
		},
		{
			"detect mutliple filters when fps filter is passed in url",
			"tags(i-frame)/fps(59.94,60)/path/here/to/master.m3u8",
			MediaFilters{
				FrameRate: []string{"59.94", "60"},
				Protocol:  ProtocolHLS,
				Tags: &Tags{
					IFrame: true,
				},
			},
			"/path/here/to/master.m3u8",
			false,
		},
		{
			"detect protocol hls for urls with .m3u8 extension",
			"/path/here/with/master.m3u8",
			MediaFilters{
				Protocol: ProtocolHLS,
			},
			"/path/here/with/master.m3u8",
			false,
		},
		{
			"detect protocol dash for urls with .mpd extension",
			"/path/here/with/manifest.mpd",
			MediaFilters{
				Protocol: ProtocolDASH,
			},
			"/path/here/with/manifest.mpd",
			false,
		},
		{
			"wrong protocol extension throws error",
			"/path/here/with/manifest.html",
			MediaFilters{},
			"",
			true,
		},
		{
			"no protocol extension throws error",
			"/path/here/with/manifest",
			MediaFilters{},
			"",
			true,
		},
		{
			"detect filters for propeller channels and set path properly",
			"/v(avc)/a(mp4a)/propeller/orgID/master.m3u8",
			MediaFilters{
				Protocol: ProtocolHLS,
				Videos: NestedFilters{
					Codecs: []string{"avc"},
				},
				Audios: NestedFilters{
					Codecs: []string{"mp4a"},
				},
			},
			"/propeller/orgID/master.m3u8",
			false,
		},
		{
			"set path properly for propeller channel with no filters",
			"/propeller/orgID/master.m3u8",
			MediaFilters{
				Protocol: ProtocolHLS,
			},
			"/propeller/orgID/master.m3u8",
			false,
		},
		{
			"ensure DeWeaved filter is set to true",
			"dw(true)/some/path/to/manifest.m3u8",
			MediaFilters{
				Protocol: ProtocolHLS,
				DeWeave:  true,
			},
			"/some/path/to/manifest.m3u8",
			false,
		},
		{
			"ensure DeWeaved filter is set to false",
			"dw(false)/some/path/to/manifest.m3u8",
			MediaFilters{
				Protocol: ProtocolHLS,
				DeWeave:  false,
			},
			"/some/path/to/manifest.m3u8",
			false,
		},
		{
			"ensure DeWeaved filter throws error if mutliple values are requested",
			"dw(true,false)/some/path/to/manifest.m3u8",
			MediaFilters{},
			"",
			true,
		},
		{
			"ensure DeWeaved filter throws error if value is not true or false",
			"dw(flase)/some/path/to/manifest.m3u8",
			MediaFilters{},
			"",
			true,
		},
		{
			"parse the http status error fallback filter and HLS protocol for m3u8",
			"fhe(true)/base64string.m3u8",
			MediaFilters{
				Protocol:                ProtocolHLS,
				FallbackHTTPStatusError: true,
			},
			"/base64string.m3u8",
			false,
		},
		{
			"parse the http status error fallback filter and VTT protocol for WebVTT",
			"fhe(true)/base64string.vtt",
			MediaFilters{
				Protocol:                ProtocolVTT,
				FallbackHTTPStatusError: true,
			},
			"/base64string.vtt",
			false,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			masterManifestPath, output, err := URLParse(test.input)
			if !test.expectedErr && err != nil {
				t.Errorf("Did not expect an error returned, got: %v", err)
				return
			} else if test.expectedErr && err == nil {
				t.Errorf("Expected an error returned, got nil")
				return
			}

			jsonOutput, err := json.Marshal(output)
			if err != nil {
				t.Fatal(err)
			}

			jsonExpected, err := json.Marshal(test.expectedFilters)
			if err != nil {
				t.Fatal(err)
			}

			if test.expectedManifestPath != masterManifestPath {
				t.Errorf("wrong master manifest generated.\nwant %v\n\ngot %v", test.expectedManifestPath, masterManifestPath)
			}

			if !reflect.DeepEqual(jsonOutput, jsonExpected) {
				t.Errorf("wrong struct generated.\nwant %v\ngot %v\n diff: %v", string(jsonExpected), string(jsonOutput), cmp.Diff(jsonExpected, jsonOutput))
			}
		})
	}
}

func TestParsers_SuppressTags(t *testing.T) {
	tests := []struct {
		name         string
		mf           MediaFilters
		expectAds    bool
		expectIFrame bool
	}{
		{
			name:         "When tags are not set, return false to suppress ads & iframe",
			mf:           MediaFilters{},
			expectAds:    false,
			expectIFrame: false,
		},
		{
			name: "When only Ads is set, return true to suppress ads & false to suppress iframe",
			mf: MediaFilters{
				Tags: &Tags{
					Ads: true,
				},
			},
			expectAds:    true,
			expectIFrame: false,
		},
		{
			name: "When only Iframe is set, return true to suppress iframe & false to suppress ads",
			mf: MediaFilters{
				Tags: &Tags{
					IFrame: true,
				},
			},
			expectAds:    false,
			expectIFrame: true,
		},
		{
			name: "When both tags are set, return true to suppress iframe & ads",
			mf: MediaFilters{
				Tags: &Tags{
					Ads:    true,
					IFrame: true,
				},
			},
			expectAds:    true,
			expectIFrame: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if gotIFrame := tc.mf.SuppressIFrame(); gotIFrame != tc.expectIFrame {
				t.Errorf("Wrong SuppressIFrame() response\ngot %v\nexpected: %v", gotIFrame, tc.expectIFrame)
			}

			if gotAds := tc.mf.SuppressAds(); gotAds != tc.expectAds {
				t.Errorf("Wrong SuppressAds() response\ngot %v\nexpected: %v", gotAds, tc.expectAds)
			}
		})
	}
}
