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
			"/ct(text)/",
			MediaFilters{
				ContentTypes: []ContentType{"text"},
			},
			"/",
			false,
		},
		{
			"multiple content types",
			"/ct(audio,video)/",
			MediaFilters{
				ContentTypes: []ContentType{"audio", "video"},
			},
			"/",
			false,
		},
		{
			"one video type",
			"/v(hdr10)/",
			MediaFilters{
				Videos: NestedFilters{
					Codecs: []Codec{"hev1.2", "hvc1.2"},
				},
			},
			"/",
			false,
		},
		{
			"two video types",
			"/v(hdr10,hevc)/",
			MediaFilters{
				Videos: NestedFilters{
					Codecs: []Codec{"hev1.2", "hvc1.2", codecHEVC},
				},
			},
			"/",
			false,
		},
		{
			"two video types and two audio types",
			"/v(hdr10,hevc)/a(aac,noAd)/",
			MediaFilters{
				Videos: NestedFilters{
					Codecs: []Codec{"hev1.2", "hvc1.2", codecHEVC},
				},
				Audios: NestedFilters{
					Codecs: []Codec{codecAAC, codecNoAudioDescription},
				},
			},
			"/",
			false,
		},
		{
			"videos, audio, captions and bitrate range",
			"/v(hdr10,hevc)/a(aac,l(pt-BR,en),b(10,20))/b(100,4000)/",
			MediaFilters{
				Videos: NestedFilters{
					Codecs: []Codec{"hev1.2", "hvc1.2", codecHEVC},
					Bitrate: &Bitrate{
						Max: 4000,
						Min: 100,
					},
				},
				Audios: NestedFilters{
					Codecs:   []Codec{codecAAC},
					Language: []Language{langPTBR, langEN},
					Bitrate: &Bitrate{
						Max: 20,
						Min: 10,
					},
				},
			},
			"/",
			false,
		},
		{
			"bitrate range doesn't override nested audio and video filter",
			"/v(hdr10,hevc,b(100,500))/a(aac,l(pt-BR,en),b(10,20))/b(100,4000)/",
			MediaFilters{
				Videos: NestedFilters{
					Codecs: []Codec{"hev1.2", "hvc1.2", codecHEVC},
					Bitrate: &Bitrate{
						Max: 500,
						Min: 100,
					},
				},
				Audios: NestedFilters{
					Codecs:   []Codec{codecAAC},
					Language: []Language{langPTBR, langEN},
					Bitrate: &Bitrate{
						Max: 20,
						Min: 10,
					},
				},
			},
			"/",
			false,
		},
		{
			"bitrate range with minimum bitrate only",
			"/b(100,)/",
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
			},
			"/",
			false,
		},
		{
			"bitrate range with maximum bitrate only",
			"/b(,3000)/",
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
			},
			"/",
			false,
		},
		{
			"bitrate range with minimum greater than maximum throws error",
			"/b(30000,3000)/",
			MediaFilters{},
			"",
			true,
		},
		{
			"bitrate range with minimum equal to maximum throws error",
			"/b(3000,3000)/",
			MediaFilters{},
			"",
			true,
		},
		{
			"audio bitrate range with minimum equal to maximum throws error",
			"/a(b(1000,1000))/",
			MediaFilters{},
			"",
			true,
		},
		{
			"video bitrate range with minimum greater than maximum throws error",
			"/v(b(2000,1000))/",
			MediaFilters{},
			"",
			true,
		},
		{
			"audio bitrate range with inavlid, negative minimum",
			"/a(b(-100,1000))/",
			MediaFilters{},
			"",
			true,
		},
		{
			"video bitrate range with invalid, greater than math.MaxInt32 minimum",
			"/v(b(2147483648))/",
			MediaFilters{},
			"",
			true,
		},
		{
			"audio bitrate range with invalid, greater than math.MaxInt32 maximum",
			"/a(b(10,2147483648))/",
			MediaFilters{},
			"",
			true,
		},
		{
			"video bitrate range with invalid, negative maximum",
			"/v(b(10,-100))/",
			MediaFilters{},
			"",
			true,
		},
		{
			"trim filter",
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
			"detect a signle plugin for execution from url",
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
			"/v(hdr10,hevc)/[plugin1,plugin2,plugin3]/some/path/master.m3u8",
			MediaFilters{
				Videos: NestedFilters{
					Codecs: []Codec{"hev1.2", "hvc1.2", codecHEVC},
				},
				Protocol: ProtocolHLS,
				Plugins:  []string{"plugin1", "plugin2", "plugin3"},
			},
			"/some/path/master.m3u8",
			false,
		},
		{
			"nested audio and video bitrate filters",
			"/a(b(100,))/v(b(,5000))/",
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
			},
			"/",
			false,
		},
		{
			"nested codec and bitrate filters in audio",
			"/a(b(100,200),co(ac-3,aac))/",
			MediaFilters{
				Audios: NestedFilters{
					Bitrate: &Bitrate{
						Min: 100,
						Max: 200,
					},
					Codecs: []Codec{codecAC3, codecAAC},
				},
			},
			"/",
			false,
		},
		{
			"nested codec and bitrate filters in video, plus overall bitrate filters",
			"/v(co(avc,hdr10),b(1000,2000))/",
			MediaFilters{
				Videos: NestedFilters{
					Bitrate: &Bitrate{
						Min: 1000,
						Max: 2000,
					},
					Codecs: []Codec{codecH264, "hev1.2", "hvc1.2"},
				},
			},
			"/",
			false,
		},
		{
			"nested bitrate and old format of codec filter",
			"/a(mp4a,ac-3,b(0,10))/",
			MediaFilters{
				Audios: NestedFilters{
					Bitrate: &Bitrate{
						Max: 10,
					},
					Codecs: []Codec{"mp4a", codecAC3},
				},
			},
			"/",
			false,
		},
		{
			"detect overall lang filter when passed in url, populate nested filters",
			"v(avc)/a(mp4a)/l(es)/path/here/with/master.m3u8",
			MediaFilters{
				Videos: NestedFilters{
					Codecs: []Codec{codecH264},
				},
				Audios: NestedFilters{
					Codecs:   []Codec{"mp4a"},
					Language: []Language{langES},
				},
				Captions: NestedFilters{
					Language: []Language{langES},
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
					Codecs: []Codec{codecH264},
				},
				Audios: NestedFilters{
					Codecs: []Codec{"mp4a"},
				},
				Captions: NestedFilters{
					Language: []Language{langES},
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
					Codecs: []Codec{"wvtt"},
				},
			},
			"/path/here/with/master.m3u8",
			false,
		},
		{
			"detect content type filter when passed in url",
			"ct(text,video)/path/here/with/master.m3u8",
			MediaFilters{
				Protocol:     ProtocolHLS,
				ContentTypes: []ContentType{"text", "video"},
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
					Codecs:   []Codec{codecH264},
					Language: []Language{langEN},
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
					Codecs:   []Codec{codecH264},
					Language: []Language{langEN},
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
				FrameRate: []FPS{"59.94"},
				Protocol:  ProtocolHLS,
			},
			"/path/here/to/master.m3u8",
			false,
		},
		{
			"detect multiple values including fractions when fps filter is passed in url",
			"fps(60,30000:1001)/path/here/to/master.mpd",
			MediaFilters{
				FrameRate: []FPS{"60", "30000/1001"},
				Protocol:  ProtocolDASH,
			},
			"/path/here/to/master.mpd",
			false,
		},
		{
			"detect mutliple filters when fps filter is passed in url",
			"tags(i-frame)/fps(59.94,60)/path/here/to/master.m3u8",
			MediaFilters{
				FrameRate: []FPS{"59.94", "60"},
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
			"detect filters for propeller channels and set path properly",
			"/v(avc)/a(aac)/propeller/orgID/master.m3u8",
			MediaFilters{
				Protocol: ProtocolHLS,
				Videos: NestedFilters{
					Codecs: []Codec{codecH264},
				},
				Audios: NestedFilters{
					Codecs: []Codec{codecAAC},
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
