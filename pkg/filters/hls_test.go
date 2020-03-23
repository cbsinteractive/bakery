package filters

import (
	"math"
	"testing"

	"github.com/cbsinteractive/bakery/pkg/config"
	"github.com/cbsinteractive/bakery/pkg/parsers"
	"github.com/google/go-cmp/cmp"
)

func TestHLSFilter_FilterManifest_BandwidthFilter(t *testing.T) {

	baseManifest := `#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=CLOSED-CAPTIONS,GROUP-ID="CC",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG"
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,CODECS="ac-3,avc",CLOSED-CAPTIONS="CC"
http://existing.base/uri/link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CODECS="ec-3",CLOSED-CAPTIONS="CC"
http://existing.base/uri/link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CODECS="avc",CLOSED-CAPTIONS="CC"
http://existing.base/uri/link_3.m3u8
`

	manifestRemovedHigherBW := `#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=CLOSED-CAPTIONS,GROUP-ID="CC",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG"
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,CODECS="ac-3,avc",CLOSED-CAPTIONS="CC"
http://existing.base/uri/link_1.m3u8
`

	manifestRemovedHigherBWOnlyAudio := `#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=CLOSED-CAPTIONS,GROUP-ID="CC",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG"
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,CODECS="ac-3,avc",CLOSED-CAPTIONS="CC"
http://existing.base/uri/link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CODECS="avc",CLOSED-CAPTIONS="CC"
http://existing.base/uri/link_3.m3u8
`

	manifestRemovedLowerBW := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CODECS="ec-3",CLOSED-CAPTIONS="CC"
http://existing.base/uri/link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CODECS="avc",CLOSED-CAPTIONS="CC"
http://existing.base/uri/link_3.m3u8
`

	tests := []struct {
		name                  string
		filters               *parsers.MediaFilters
		manifestContent       string
		expectManifestContent string
		expectErr             bool
	}{
		{
			name: "when no bitrate filters given, expect unfiltered manifest",
			filters: &parsers.MediaFilters{
				MinBitrate: 0, MaxBitrate: math.MaxInt32,
				VideoFilters: parsers.NestedFilters{MaxBitrate: math.MaxInt32},
				AudioFilters: parsers.NestedFilters{MaxBitrate: math.MaxInt32},
			},
			manifestContent:       baseManifest,
			expectManifestContent: baseManifest,
		},
		{
			name: "when only hitting lower boundary (MinBitrate = 0), expect results to be filtered",
			filters: &parsers.MediaFilters{
				MinBitrate: 0, MaxBitrate: 3000,
				VideoFilters: parsers.NestedFilters{MaxBitrate: math.MaxInt32},
				AudioFilters: parsers.NestedFilters{MaxBitrate: math.MaxInt32},
			},
			manifestContent:       baseManifest,
			expectManifestContent: manifestRemovedHigherBW,
		},
		{
			name: "when only hitting upper boundary (MaxBitrate = math.MaxInt32), expect results to be filtered",
			filters: &parsers.MediaFilters{
				MinBitrate: 3000, MaxBitrate: math.MaxInt32,
				VideoFilters: parsers.NestedFilters{MaxBitrate: math.MaxInt32},
				AudioFilters: parsers.NestedFilters{MaxBitrate: math.MaxInt32},
			},
			manifestContent:       baseManifest,
			expectManifestContent: manifestRemovedLowerBW,
		},
		{
			name: "when filtering valid bitrate range in audio only, expect filtered manifest",
			filters: &parsers.MediaFilters{
				MinBitrate: 0, MaxBitrate: math.MaxInt32,
				VideoFilters: parsers.NestedFilters{MaxBitrate: math.MaxInt32},
				AudioFilters: parsers.NestedFilters{MinBitrate: 10, MaxBitrate: 3000},
			},
			manifestContent:       baseManifest,
			expectManifestContent: manifestRemovedHigherBWOnlyAudio,
		},
		{
			name: "when filtering a valid audio bitrate range touching upper bound (math.MaxInt32), expect audio to be filtered",
			filters: &parsers.MediaFilters{
				MinBitrate: 0, MaxBitrate: math.MaxInt32,
				VideoFilters: parsers.NestedFilters{MaxBitrate: math.MaxInt32},
				AudioFilters: parsers.NestedFilters{MinBitrate: 2000, MaxBitrate: math.MaxInt32},
			},
			manifestContent:       baseManifest,
			expectManifestContent: manifestRemovedLowerBW,
		},
		{
			name: "when given valid overall bitrate range and an valid bitrate range for video/audio overlapping, but not within overall range, expect manifest to be filtered according to overall bitrate range",
			filters: &parsers.MediaFilters{
				MinBitrate: 0, MaxBitrate: 3000,
				VideoFilters: parsers.NestedFilters{MaxBitrate: math.MaxInt32},
				AudioFilters: parsers.NestedFilters{MaxBitrate: math.MaxInt32}},
			manifestContent:       baseManifest,
			expectManifestContent: manifestRemovedHigherBW,
		},
		{
			name: "when given valid overall bitrate range and a valid, non-overlapping bitrate range for audio, expect manifest to be filtered according to overall bitrate range",
			filters: &parsers.MediaFilters{
				MinBitrate: 100, MaxBitrate: 1000,
				VideoFilters: parsers.NestedFilters{MaxBitrate: math.MaxInt32},
				AudioFilters: parsers.NestedFilters{MinBitrate: 3000, MaxBitrate: 4000},
			},
			manifestContent:       baseManifest,
			expectManifestContent: manifestRemovedHigherBW,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			filter := NewHLSFilter("", tt.manifestContent, config.Config{})
			manifest, err := filter.FilterManifest(tt.filters)

			if err != nil && !tt.expectErr {
				t.Errorf("FilterManifest() didnt expect an error to be returned, got: %v", err)
				return
			} else if err == nil && tt.expectErr {
				t.Error("FilterManifest() expected an error, got nil")
				return
			}

			if g, e := manifest, tt.expectManifestContent; g != e {
				t.Errorf("FilterManifest() wrong manifest returned\ngot %v\nexpected: %v\ndiff: %v", g, e,
					cmp.Diff(g, e))
			}

		})
	}
}

func TestHLSFilter_FilterManifest_AudioFilter(t *testing.T) {
	manifestWithAllAudio := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,CODECS="ec-3"
http://existing.base/uri/link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1100,AVERAGE-BANDWIDTH=1100,CODECS="ac-3"
http://existing.base/uri/link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CODECS="mp4a.40.2"
http://existing.base/uri/link_3.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4500,AVERAGE-BANDWIDTH=4500,CODECS="ec-3,ac-3"
http://existing.base/uri/link_4.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=6000,AVERAGE-BANDWIDTH=6000,CODECS="avc1.77.30,ec-3"
http://existing.base/uri/link_5.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1500,AVERAGE-BANDWIDTH=1500,CODECS="avc1.77.30"
http://existing.base/uri/link_6.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300,CODECS="wvtt"
http://existing.base/uri/link_7.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300
http://existing.base/uri/link_8.m3u8
`

	manifestFilterInEC3 := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,CODECS="ec-3"
http://existing.base/uri/link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=6000,AVERAGE-BANDWIDTH=6000,CODECS="avc1.77.30,ec-3"
http://existing.base/uri/link_5.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1500,AVERAGE-BANDWIDTH=1500,CODECS="avc1.77.30"
http://existing.base/uri/link_6.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300,CODECS="wvtt"
http://existing.base/uri/link_7.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300
http://existing.base/uri/link_8.m3u8
`

	manifestFilterInAC3 := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1100,AVERAGE-BANDWIDTH=1100,CODECS="ac-3"
http://existing.base/uri/link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1500,AVERAGE-BANDWIDTH=1500,CODECS="avc1.77.30"
http://existing.base/uri/link_6.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300,CODECS="wvtt"
http://existing.base/uri/link_7.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300
http://existing.base/uri/link_8.m3u8
`

	manifestFilterInMP4A := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CODECS="mp4a.40.2"
http://existing.base/uri/link_3.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1500,AVERAGE-BANDWIDTH=1500,CODECS="avc1.77.30"
http://existing.base/uri/link_6.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300,CODECS="wvtt"
http://existing.base/uri/link_7.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300
http://existing.base/uri/link_8.m3u8
`

	manifestFilterWithoutMP4A := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,CODECS="ec-3"
http://existing.base/uri/link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1100,AVERAGE-BANDWIDTH=1100,CODECS="ac-3"
http://existing.base/uri/link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4500,AVERAGE-BANDWIDTH=4500,CODECS="ec-3,ac-3"
http://existing.base/uri/link_4.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=6000,AVERAGE-BANDWIDTH=6000,CODECS="avc1.77.30,ec-3"
http://existing.base/uri/link_5.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1500,AVERAGE-BANDWIDTH=1500,CODECS="avc1.77.30"
http://existing.base/uri/link_6.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300,CODECS="wvtt"
http://existing.base/uri/link_7.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300
http://existing.base/uri/link_8.m3u8
`

	manifestFilterWithoutAC3 := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,CODECS="ec-3"
http://existing.base/uri/link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CODECS="mp4a.40.2"
http://existing.base/uri/link_3.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=6000,AVERAGE-BANDWIDTH=6000,CODECS="avc1.77.30,ec-3"
http://existing.base/uri/link_5.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1500,AVERAGE-BANDWIDTH=1500,CODECS="avc1.77.30"
http://existing.base/uri/link_6.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300,CODECS="wvtt"
http://existing.base/uri/link_7.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300
http://existing.base/uri/link_8.m3u8
`

	manifestWithoutAudio := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1500,AVERAGE-BANDWIDTH=1500,CODECS="avc1.77.30"
http://existing.base/uri/link_6.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300,CODECS="wvtt"
http://existing.base/uri/link_7.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300
http://existing.base/uri/link_8.m3u8
`

	tests := []struct {
		name                  string
		filters               *parsers.MediaFilters
		manifestContent       string
		expectManifestContent string
		expectErr             bool
	}{
		{
			name: "when all audio codecs are supplied, expect audio to be stripped out",
			filters: &parsers.MediaFilters{
				MaxBitrate: math.MaxInt32,
				VideoFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
				},
				AudioFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
					Codecs:     []parsers.Codec{"mp4a", "ec-3", "ac-3"},
				},
			},
			manifestContent:       manifestWithAllAudio,
			expectManifestContent: manifestWithoutAudio,
		},
		{
			name: "when filter is supplied with ac-3 and mp4a, expect variants with ac-3 and/or mp4a to be stripped out",
			filters: &parsers.MediaFilters{
				MaxBitrate: math.MaxInt32,
				VideoFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
				},
				AudioFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
					Codecs:     []parsers.Codec{"mp4a", "ac-3"},
				},
			},
			manifestContent:       manifestWithAllAudio,
			expectManifestContent: manifestFilterInEC3,
		},
		{
			name: "when filter is supplied with ac-3, expect variants with ac-3 to be stripped out",
			filters: &parsers.MediaFilters{
				MaxBitrate: math.MaxInt32,
				VideoFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
				},
				AudioFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
					Codecs:     []parsers.Codec{"ac-3"},
				},
			},
			manifestContent:       manifestWithAllAudio,
			expectManifestContent: manifestFilterWithoutAC3,
		},
		{
			name: "when filter is supplied with mp4a, expect variants with mp4a to be stripped out",
			filters: &parsers.MediaFilters{
				MaxBitrate: math.MaxInt32,
				VideoFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
				},
				AudioFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
					Codecs:     []parsers.Codec{"mp4a"},
				},
			},
			manifestContent:       manifestWithAllAudio,
			expectManifestContent: manifestFilterWithoutMP4A,
		},
		{
			name: "when filter is supplied with ec-3 and ac-3, expect variants with ec-3 and ac-3 to be stripped out",
			filters: &parsers.MediaFilters{
				MaxBitrate: math.MaxInt32,
				VideoFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
				},
				AudioFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
					Codecs:     []parsers.Codec{"ec-3", "ac-3"},
				},
			},
			manifestContent:       manifestWithAllAudio,
			expectManifestContent: manifestFilterInMP4A,
		},
		{
			name: "when filter is supplied with ec-3 and mp4a, expect variants with ec-3 and/or mp4a to be stripped out",
			filters: &parsers.MediaFilters{
				MaxBitrate: math.MaxInt32,
				VideoFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
				},
				AudioFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
					Codecs:     []parsers.Codec{"mp4a", "ec-3"},
				},
			},
			manifestContent:       manifestWithAllAudio,
			expectManifestContent: manifestFilterInAC3,
		},
		{
			name: "when no audio filters are given, expect unfiltered manifest",
			filters: &parsers.MediaFilters{
				MaxBitrate: math.MaxInt32,
				VideoFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
				},
				AudioFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
				},
			},
			manifestContent:       manifestWithAllAudio,
			expectManifestContent: manifestWithAllAudio,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			filter := NewHLSFilter("", tt.manifestContent, config.Config{})
			manifest, err := filter.FilterManifest(tt.filters)

			if err != nil && !tt.expectErr {
				t.Errorf("FilterManifest() didnt expect an error to be returned, got: %v", err)
				return
			} else if err == nil && tt.expectErr {
				t.Error("FilterManifest() expected an error, got nil")
				return
			} else if err != nil && tt.expectErr {
				return
			}

			if g, e := manifest, tt.expectManifestContent; g != e {
				t.Errorf("FilterManifest() wrong manifest returned)\ngot %v\nexpected: %v\ndiff: %v", g, e,
					cmp.Diff(g, e))
			}

		})
	}
}

func TestHLSFilter_FilterManifest_VideoFilter(t *testing.T) {
	manifestWithAllVideo := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,CODECS="avc1.640020"
http://existing.base/uri/link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1100,AVERAGE-BANDWIDTH=1100,CODECS="avc1.77.30"
http://existing.base/uri/link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CODECS="hvc1.2.4.L93.90"
http://existing.base/uri/link_3.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4500,AVERAGE-BANDWIDTH=4500,CODECS="dvh1.05.01"
http://existing.base/uri/link_4.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4500,AVERAGE-BANDWIDTH=4500,CODECS="avc1.640029,hvc1.1.4.L126.B0"
http://existing.base/uri/link_5.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=6000,AVERAGE-BANDWIDTH=6000,CODECS="avc1.77.30,ec-3"
http://existing.base/uri/link_6.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1500,AVERAGE-BANDWIDTH=1500,CODECS="ec-3"
http://existing.base/uri/link_7.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300,CODECS="wvtt"
http://existing.base/uri/link_8.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300
http://existing.base/uri/link_9.m3u8
`

	manifestFilterWithoutAVC := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CODECS="hvc1.2.4.L93.90"
http://existing.base/uri/link_3.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4500,AVERAGE-BANDWIDTH=4500,CODECS="dvh1.05.01"
http://existing.base/uri/link_4.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1500,AVERAGE-BANDWIDTH=1500,CODECS="ec-3"
http://existing.base/uri/link_7.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300,CODECS="wvtt"
http://existing.base/uri/link_8.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300
http://existing.base/uri/link_9.m3u8
`

	manifestFilterWithoutAVCAndDVH := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CODECS="hvc1.2.4.L93.90"
http://existing.base/uri/link_3.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1500,AVERAGE-BANDWIDTH=1500,CODECS="ec-3"
http://existing.base/uri/link_7.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300,CODECS="wvtt"
http://existing.base/uri/link_8.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300
http://existing.base/uri/link_9.m3u8
`

	manifestFilterWithoutAVCAndHEVC := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4500,AVERAGE-BANDWIDTH=4500,CODECS="dvh1.05.01"
http://existing.base/uri/link_4.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1500,AVERAGE-BANDWIDTH=1500,CODECS="ec-3"
http://existing.base/uri/link_7.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300,CODECS="wvtt"
http://existing.base/uri/link_8.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300
http://existing.base/uri/link_9.m3u8
`

	manifestFilterWithoutDVH := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,CODECS="avc1.640020"
http://existing.base/uri/link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1100,AVERAGE-BANDWIDTH=1100,CODECS="avc1.77.30"
http://existing.base/uri/link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CODECS="hvc1.2.4.L93.90"
http://existing.base/uri/link_3.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4500,AVERAGE-BANDWIDTH=4500,CODECS="avc1.640029,hvc1.1.4.L126.B0"
http://existing.base/uri/link_5.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=6000,AVERAGE-BANDWIDTH=6000,CODECS="avc1.77.30,ec-3"
http://existing.base/uri/link_6.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1500,AVERAGE-BANDWIDTH=1500,CODECS="ec-3"
http://existing.base/uri/link_7.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300,CODECS="wvtt"
http://existing.base/uri/link_8.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300
http://existing.base/uri/link_9.m3u8
`

	manifestFilterWithoutHEVC := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,CODECS="avc1.640020"
http://existing.base/uri/link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1100,AVERAGE-BANDWIDTH=1100,CODECS="avc1.77.30"
http://existing.base/uri/link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4500,AVERAGE-BANDWIDTH=4500,CODECS="dvh1.05.01"
http://existing.base/uri/link_4.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=6000,AVERAGE-BANDWIDTH=6000,CODECS="avc1.77.30,ec-3"
http://existing.base/uri/link_6.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1500,AVERAGE-BANDWIDTH=1500,CODECS="ec-3"
http://existing.base/uri/link_7.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300,CODECS="wvtt"
http://existing.base/uri/link_8.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300
http://existing.base/uri/link_9.m3u8
`

	manifestWithoutVideo := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1500,AVERAGE-BANDWIDTH=1500,CODECS="ec-3"
http://existing.base/uri/link_7.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300,CODECS="wvtt"
http://existing.base/uri/link_8.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300
http://existing.base/uri/link_9.m3u8
`

	tests := []struct {
		name                  string
		filters               *parsers.MediaFilters
		manifestContent       string
		expectManifestContent string
		expectErr             bool
	}{
		{
			name: "when all video codecs are supplied, expect variants with avc, hevc, and/or dvh to be stripped out",
			filters: &parsers.MediaFilters{
				MaxBitrate: math.MaxInt32,
				VideoFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
					Codecs:     []parsers.Codec{"avc", "hvc", "dvh"},
				},
				AudioFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
				},
			},
			manifestContent:       manifestWithAllVideo,
			expectManifestContent: manifestWithoutVideo,
		},
		{
			name: "when filter is supplied with avc, expect variants with avc to be stripped out",
			filters: &parsers.MediaFilters{
				MaxBitrate: math.MaxInt32,
				VideoFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
					Codecs:     []parsers.Codec{"avc"},
				},
				AudioFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
				},
			},
			manifestContent:       manifestWithAllVideo,
			expectManifestContent: manifestFilterWithoutAVC,
		},
		{
			name: "when filter is supplied with hevc, expect hevc to be stripped out",
			filters: &parsers.MediaFilters{
				MaxBitrate: math.MaxInt32,
				VideoFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
					Codecs:     []parsers.Codec{"hvc"},
				},
				AudioFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
				},
			},
			manifestContent:       manifestWithAllVideo,
			expectManifestContent: manifestFilterWithoutHEVC,
		},
		{
			name: "when filter is supplied with dvh, expect dvh to be stripped out",
			filters: &parsers.MediaFilters{
				MaxBitrate: math.MaxInt32,
				VideoFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
					Codecs:     []parsers.Codec{"dvh"},
				},
				AudioFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
				},
			},
			manifestContent:       manifestWithAllVideo,
			expectManifestContent: manifestFilterWithoutDVH,
		},
		{
			name: "when filter is supplied with avc and hevc, expect variants with avc and hevc to be stripped out",
			filters: &parsers.MediaFilters{
				MaxBitrate: math.MaxInt32,
				VideoFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
					Codecs:     []parsers.Codec{"avc", "hvc"},
				},
				AudioFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
				},
			},
			manifestContent:       manifestWithAllVideo,
			expectManifestContent: manifestFilterWithoutAVCAndHEVC,
		},
		{
			name: "when filter is supplied with avc and dvh, expect variants with avc and dvh to be stripped out",
			filters: &parsers.MediaFilters{
				MaxBitrate: math.MaxInt32,
				VideoFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
					Codecs:     []parsers.Codec{"avc", "dvh"},
				},
				AudioFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
				},
			},
			manifestContent:       manifestWithAllVideo,
			expectManifestContent: manifestFilterWithoutAVCAndDVH,
		},
		{
			name: "when no video filters are given, expect unfiltered manifest",
			filters: &parsers.MediaFilters{
				MaxBitrate: math.MaxInt32,
				VideoFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
				},
				AudioFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
				},
			},
			manifestContent:       manifestWithAllVideo,
			expectManifestContent: manifestWithAllVideo,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			filter := NewHLSFilter("", tt.manifestContent, config.Config{})
			manifest, err := filter.FilterManifest(tt.filters)

			if err != nil && !tt.expectErr {
				t.Errorf("FilterManifest() didnt expect an error to be returned, got: %v", err)
				return
			} else if err == nil && tt.expectErr {
				t.Error("FilterManifest() expected an error, got nil")
				return
			}

			if g, e := manifest, tt.expectManifestContent; g != e {
				t.Errorf("FilterManifest() wrong manifest returned)\ngot %v\nexpected: %v\ndiff: %v", g, e,
					cmp.Diff(g, e))
			}

		})
	}
}

func TestHLSFilter_FilterManifest_CaptionsFilter(t *testing.T) {
	manifestWithAllCaptions := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,CODECS="wvtt"
http://existing.base/uri/link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1100,AVERAGE-BANDWIDTH=1100,CODECS="stpp"
http://existing.base/uri/link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CODECS="wvtt,stpp"
http://existing.base/uri/link_3.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4500,AVERAGE-BANDWIDTH=4500,CODECS="wvtt,ac-3"
http://existing.base/uri/link_4.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4500,AVERAGE-BANDWIDTH=4500,CODECS="avc1.640029"
http://existing.base/uri/link_5.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=6000,AVERAGE-BANDWIDTH=6000,CODECS="ec-3"
http://existing.base/uri/link_6.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300
http://existing.base/uri/link_7.m3u8
`

	manifestFilterWithoutSTPP := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,CODECS="wvtt"
http://existing.base/uri/link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4500,AVERAGE-BANDWIDTH=4500,CODECS="wvtt,ac-3"
http://existing.base/uri/link_4.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4500,AVERAGE-BANDWIDTH=4500,CODECS="avc1.640029"
http://existing.base/uri/link_5.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=6000,AVERAGE-BANDWIDTH=6000,CODECS="ec-3"
http://existing.base/uri/link_6.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300
http://existing.base/uri/link_7.m3u8
`

	manifestFilterWithoutWVTT := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1100,AVERAGE-BANDWIDTH=1100,CODECS="stpp"
http://existing.base/uri/link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4500,AVERAGE-BANDWIDTH=4500,CODECS="avc1.640029"
http://existing.base/uri/link_5.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=6000,AVERAGE-BANDWIDTH=6000,CODECS="ec-3"
http://existing.base/uri/link_6.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300
http://existing.base/uri/link_7.m3u8
`

	manifestWithNoCaptions := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4500,AVERAGE-BANDWIDTH=4500,CODECS="avc1.640029"
http://existing.base/uri/link_5.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=6000,AVERAGE-BANDWIDTH=6000,CODECS="ec-3"
http://existing.base/uri/link_6.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300
http://existing.base/uri/link_7.m3u8
`

	tests := []struct {
		name                  string
		filters               *parsers.MediaFilters
		manifestContent       string
		expectManifestContent string
		expectErr             bool
	}{
		{
			name: "when all caption filters are supplied, expect all caption variants with captions to be stripped out",
			filters: &parsers.MediaFilters{
				MaxBitrate:   math.MaxInt32,
				CaptionTypes: []parsers.CaptionType{"stpp", "wvtt"},
				VideoFilters: parsers.NestedFilters{MaxBitrate: math.MaxInt32},
				AudioFilters: parsers.NestedFilters{MaxBitrate: math.MaxInt32},
			},
			manifestContent:       manifestWithAllCaptions,
			expectManifestContent: manifestWithNoCaptions,
		},
		{
			name: "when filter is supplied with wvtt, expect variants with wvtt to be stripped out",
			filters: &parsers.MediaFilters{
				MaxBitrate:   math.MaxInt32,
				CaptionTypes: []parsers.CaptionType{"wvtt"},
				VideoFilters: parsers.NestedFilters{MaxBitrate: math.MaxInt32},
				AudioFilters: parsers.NestedFilters{MaxBitrate: math.MaxInt32},
			},
			manifestContent:       manifestWithAllCaptions,
			expectManifestContent: manifestFilterWithoutWVTT,
		},
		{
			name: "when filter is supplied with stpp, expect variants with wvtt to be stripped out",
			filters: &parsers.MediaFilters{
				MaxBitrate:   math.MaxInt32,
				CaptionTypes: []parsers.CaptionType{"stpp"},
				VideoFilters: parsers.NestedFilters{MaxBitrate: math.MaxInt32},
				AudioFilters: parsers.NestedFilters{MaxBitrate: math.MaxInt32},
			},
			manifestContent:       manifestWithAllCaptions,
			expectManifestContent: manifestFilterWithoutSTPP,
		},
		{
			name: "when no caption filter is given, expect original manifest",
			filters: &parsers.MediaFilters{
				MaxBitrate:   math.MaxInt32,
				VideoFilters: parsers.NestedFilters{MaxBitrate: math.MaxInt32},
				AudioFilters: parsers.NestedFilters{MaxBitrate: math.MaxInt32},
			},
			manifestContent:       manifestWithAllCaptions,
			expectManifestContent: manifestWithAllCaptions,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			filter := NewHLSFilter("", tt.manifestContent, config.Config{})
			manifest, err := filter.FilterManifest(tt.filters)

			if err != nil && !tt.expectErr {
				t.Errorf("FilterManifest() didnt expect an error to be returned, got: %v", err)
				return
			} else if err == nil && tt.expectErr {
				t.Error("FilterManifest() expected an error, got nil")
				return
			}

			if g, e := manifest, tt.expectManifestContent; g != e {
				t.Errorf("FilterManifest() wrong manifest returned\ngot %v\nexpected: %v\ndiff: %v", g, e,
					cmp.Diff(g, e))
			}

		})
	}
}

func TestHLSFilter_FilterManifest_MultiCodecFilter(t *testing.T) {
	manifestWithAllCodecs := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,CODECS="ac-3"
http://existing.base/uri/link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1100,AVERAGE-BANDWIDTH=1100,CODECS="avc1.77.30"
http://existing.base/uri/link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1100,AVERAGE-BANDWIDTH=1100,CODECS="ac-3,avc1.77.30,wvtt"
http://existing.base/uri/link_3.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CODECS="ac-3,hvc1.2.4.L93.90"
http://existing.base/uri/link_4.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4500,AVERAGE-BANDWIDTH=4500,CODECS="ac-3,avc1.77.30,dvh1.05.01"
http://existing.base/uri/link_5.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4500,AVERAGE-BANDWIDTH=4500,CODECS="ec-3,avc1.640029"
http://existing.base/uri/link_6.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=6000,AVERAGE-BANDWIDTH=6000,CODECS="ac-3,avc1.77.30,ec-3"
http://existing.base/uri/link_7.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1500,AVERAGE-BANDWIDTH=1500,CODECS="ac-3,hvc1.2.4.L93.90,ec-3"
http://existing.base/uri/link_8.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300,CODECS="ac-3,ec-3,mp4a.40.2,avc1.640029"
http://existing.base/uri/link_9.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300,CODECS="ac-3,ec-3,wvtt"
http://existing.base/uri/link_10.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300,CODECS="ac-3,wvtt"
http://existing.base/uri/link_11.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300,CODECS="ec-3,wvtt"
http://existing.base/uri/link_12.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300,CODECS="ac-3,stpp"
http://existing.base/uri/link_13.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300
http://existing.base/uri/link_14.m3u8
`

	manifestFilterInAC3AndAVC := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,CODECS="ac-3"
http://existing.base/uri/link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1100,AVERAGE-BANDWIDTH=1100,CODECS="avc1.77.30"
http://existing.base/uri/link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1100,AVERAGE-BANDWIDTH=1100,CODECS="ac-3,avc1.77.30,wvtt"
http://existing.base/uri/link_3.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300,CODECS="ac-3,wvtt"
http://existing.base/uri/link_11.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300,CODECS="ac-3,stpp"
http://existing.base/uri/link_13.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300
http://existing.base/uri/link_14.m3u8
`

	manifestFilterInAC3AndEC3AndAVC := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,CODECS="ac-3"
http://existing.base/uri/link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1100,AVERAGE-BANDWIDTH=1100,CODECS="avc1.77.30"
http://existing.base/uri/link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1100,AVERAGE-BANDWIDTH=1100,CODECS="ac-3,avc1.77.30,wvtt"
http://existing.base/uri/link_3.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4500,AVERAGE-BANDWIDTH=4500,CODECS="ec-3,avc1.640029"
http://existing.base/uri/link_6.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=6000,AVERAGE-BANDWIDTH=6000,CODECS="ac-3,avc1.77.30,ec-3"
http://existing.base/uri/link_7.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300,CODECS="ac-3,ec-3,wvtt"
http://existing.base/uri/link_10.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300,CODECS="ac-3,wvtt"
http://existing.base/uri/link_11.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300,CODECS="ec-3,wvtt"
http://existing.base/uri/link_12.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300,CODECS="ac-3,stpp"
http://existing.base/uri/link_13.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300
http://existing.base/uri/link_14.m3u8
`

	manifestFilterInAC3AndWVTT := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,CODECS="ac-3"
http://existing.base/uri/link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1100,AVERAGE-BANDWIDTH=1100,CODECS="avc1.77.30"
http://existing.base/uri/link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1100,AVERAGE-BANDWIDTH=1100,CODECS="ac-3,avc1.77.30,wvtt"
http://existing.base/uri/link_3.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CODECS="ac-3,hvc1.2.4.L93.90"
http://existing.base/uri/link_4.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4500,AVERAGE-BANDWIDTH=4500,CODECS="ac-3,avc1.77.30,dvh1.05.01"
http://existing.base/uri/link_5.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300,CODECS="ac-3,wvtt"
http://existing.base/uri/link_11.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300
http://existing.base/uri/link_14.m3u8
`

	manifestFilterInAC3AndAVCAndWVTT := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,CODECS="ac-3"
http://existing.base/uri/link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1100,AVERAGE-BANDWIDTH=1100,CODECS="avc1.77.30"
http://existing.base/uri/link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1100,AVERAGE-BANDWIDTH=1100,CODECS="ac-3,avc1.77.30,wvtt"
http://existing.base/uri/link_3.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300,CODECS="ac-3,wvtt"
http://existing.base/uri/link_11.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300
http://existing.base/uri/link_14.m3u8
`

	manifestNoAudioAndFilterInAVC := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1100,AVERAGE-BANDWIDTH=1100,CODECS="avc1.77.30"
http://existing.base/uri/link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300
http://existing.base/uri/link_14.m3u8
`

	tests := []struct {
		name                  string
		filters               *parsers.MediaFilters
		manifestContent       string
		expectManifestContent string
		expectErr             bool
	}{
		{
			name: "when empty filters are given, expect original manifest",
			filters: &parsers.MediaFilters{
				MaxBitrate:   math.MaxInt32,
				VideoFilters: parsers.NestedFilters{MaxBitrate: math.MaxInt32},
				AudioFilters: parsers.NestedFilters{MaxBitrate: math.MaxInt32},
			},
			manifestContent:       manifestWithAllCodecs,
			expectManifestContent: manifestWithAllCodecs,
		},
		{
			name: "when filter is supplied with audio (ec-3 and mp4a) and video (hevc and dvh), expect variants with ec-3, mp4a, hevc, and/or dvh to be stripped out",
			filters: &parsers.MediaFilters{
				MaxBitrate: math.MaxInt32,
				VideoFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
					Codecs:     []parsers.Codec{"hvc", "dvh"},
				},
				AudioFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
					Codecs:     []parsers.Codec{"ec-3", "mp4a"},
				},
			},
			manifestContent:       manifestWithAllCodecs,
			expectManifestContent: manifestFilterInAC3AndAVC,
		},
		{
			name: "when filter is supplied with audio (mp4a) and video (hevc and dvh), expect variants with mp4a, hevc, and/or dvh to be stripped out",
			filters: &parsers.MediaFilters{
				MaxBitrate: math.MaxInt32,
				VideoFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
					Codecs:     []parsers.Codec{"hvc", "dvh"},
				},
				AudioFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
					Codecs:     []parsers.Codec{"mp4a"},
				},
			},
			manifestContent:       manifestWithAllCodecs,
			expectManifestContent: manifestFilterInAC3AndEC3AndAVC,
		},
		{
			name: "when filter is supplied with audio (ec-3 and mp4a) and captions (stpp), expect variants with ec-3, mp4a, and/or stpp to be stripped out",
			filters: &parsers.MediaFilters{
				MaxBitrate:   math.MaxInt32,
				CaptionTypes: []parsers.CaptionType{"stpp"},
				VideoFilters: parsers.NestedFilters{MaxBitrate: math.MaxInt32},
				AudioFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
					Codecs:     []parsers.Codec{"ec-3", "mp4a"},
				},
			},
			manifestContent:       manifestWithAllCodecs,
			expectManifestContent: manifestFilterInAC3AndWVTT,
		},
		{
			name: "when filter is supplied with audio (ec-3 and mp4a), video (hevc and dvh), and captions (stpp), expect variants with ec-3, mp4a, hevc, dvh, and/or stpp to be stripped out",
			filters: &parsers.MediaFilters{
				MaxBitrate:   math.MaxInt32,
				CaptionTypes: []parsers.CaptionType{"stpp"},
				VideoFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
					Codecs:     []parsers.Codec{"hvc", "dvh"},
				},
				AudioFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
					Codecs:     []parsers.Codec{"ec-3", "mp4a"},
				},
			},
			manifestContent:       manifestWithAllCodecs,
			expectManifestContent: manifestFilterInAC3AndAVCAndWVTT,
		},
		{
			name: "when filtering out all codecs except avc video, expect variants with ac-3, ec-3, mp4a, hevc, and/or dvh to be stripped out",
			filters: &parsers.MediaFilters{
				MaxBitrate:   math.MaxInt32,
				CaptionTypes: []parsers.CaptionType{"wvtt", "stpp"},
				VideoFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
					Codecs:     []parsers.Codec{"hvc", "dvh"},
				},
				AudioFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
					Codecs:     []parsers.Codec{"ac-3", "ec-3", "mp4a"},
				},
			},
			manifestContent:       manifestWithAllCodecs,
			expectManifestContent: manifestNoAudioAndFilterInAVC,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			filter := NewHLSFilter("", tt.manifestContent, config.Config{})
			manifest, err := filter.FilterManifest(tt.filters)

			if err != nil && !tt.expectErr {
				t.Errorf("FilterManifest() didnt expect an error to be returned, got: %v", err)
				return
			} else if err == nil && tt.expectErr {
				t.Error("FilterManifest() expected an error, got nil")
				return
			}

			if g, e := manifest, tt.expectManifestContent; g != e {
				t.Errorf("FilterManifest() wrong manifest returned)\ngot %v\nexpected: %v\ndiff: %v", g, e,
					cmp.Diff(g, e))
			}

		})
	}
}

func TestHLSFilter_FilterManifest_MultiFilter(t *testing.T) {

	manifestWithAllCodecsAndBandwidths := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,CODECS="ac-3"
http://existing.base/uri/link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4200,AVERAGE-BANDWIDTH=4200,CODECS="avc1.77.30"
http://existing.base/uri/link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CODECS="ac-3,hvc1.2.4.L93.90"
http://existing.base/uri/link_4.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4100,AVERAGE-BANDWIDTH=4100,CODECS="ac-3,avc1.77.30,dvh1.05.01"
http://existing.base/uri/link_5.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4500,AVERAGE-BANDWIDTH=4500,CODECS="ec-3,avc1.640029"
http://existing.base/uri/link_6.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=6000,AVERAGE-BANDWIDTH=6000,CODECS="ac-3,avc1.77.30,ec-3"
http://existing.base/uri/link_7a.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=5900,AVERAGE-BANDWIDTH=5900,CODECS="ac-3,ec-3"
http://existing.base/uri/link_7b.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=500,AVERAGE-BANDWIDTH=500,CODECS="wvtt"
http://existing.base/uri/link_14.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=5300,AVERAGE-BANDWIDTH=5300
http://existing.base/uri/link_13.m3u8
`

	manifestFilter4000To6000BandwidthAndAC3 := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4200,AVERAGE-BANDWIDTH=4200,CODECS="avc1.77.30"
http://existing.base/uri/link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CODECS="ac-3,hvc1.2.4.L93.90"
http://existing.base/uri/link_4.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4100,AVERAGE-BANDWIDTH=4100,CODECS="ac-3,avc1.77.30,dvh1.05.01"
http://existing.base/uri/link_5.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=5300,AVERAGE-BANDWIDTH=5300
http://existing.base/uri/link_13.m3u8
`

	manifestFilter4000To6000BandwidthAndDVH := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=5900,AVERAGE-BANDWIDTH=5900,CODECS="ac-3,ec-3"
http://existing.base/uri/link_7b.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=5300,AVERAGE-BANDWIDTH=5300
http://existing.base/uri/link_13.m3u8
`

	manifestFilter4000To6000BandwidthAndEC3AndAVC := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4200,AVERAGE-BANDWIDTH=4200,CODECS="avc1.77.30"
http://existing.base/uri/link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4500,AVERAGE-BANDWIDTH=4500,CODECS="ec-3,avc1.640029"
http://existing.base/uri/link_6.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=5300,AVERAGE-BANDWIDTH=5300
http://existing.base/uri/link_13.m3u8
`

	manifestFilter4000To6000BandwidthAndWVTT := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4200,AVERAGE-BANDWIDTH=4200,CODECS="avc1.77.30"
http://existing.base/uri/link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CODECS="ac-3,hvc1.2.4.L93.90"
http://existing.base/uri/link_4.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4100,AVERAGE-BANDWIDTH=4100,CODECS="ac-3,avc1.77.30,dvh1.05.01"
http://existing.base/uri/link_5.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4500,AVERAGE-BANDWIDTH=4500,CODECS="ec-3,avc1.640029"
http://existing.base/uri/link_6.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=6000,AVERAGE-BANDWIDTH=6000,CODECS="ac-3,avc1.77.30,ec-3"
http://existing.base/uri/link_7a.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=5900,AVERAGE-BANDWIDTH=5900,CODECS="ac-3,ec-3"
http://existing.base/uri/link_7b.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=5300,AVERAGE-BANDWIDTH=5300
http://existing.base/uri/link_13.m3u8
`

	manifestFilter4000To6000BandwidthAndNoAudio := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4200,AVERAGE-BANDWIDTH=4200,CODECS="avc1.77.30"
http://existing.base/uri/link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=5300,AVERAGE-BANDWIDTH=5300
http://existing.base/uri/link_13.m3u8
`

	tests := []struct {
		name                  string
		filters               *parsers.MediaFilters
		manifestContent       string
		expectManifestContent string
		expectErr             bool
	}{
		{
			name: "when no filters are given, expect original manifest",
			filters: &parsers.MediaFilters{
				MaxBitrate:   math.MaxInt32,
				VideoFilters: parsers.NestedFilters{MaxBitrate: math.MaxInt32},
				AudioFilters: parsers.NestedFilters{MaxBitrate: math.MaxInt32},
			},
			manifestContent:       manifestWithAllCodecsAndBandwidths,
			expectManifestContent: manifestWithAllCodecsAndBandwidths,
		},
		{
			name: "when filtering out audio (ec-3) and filtering in bandwidth range 4000-6000, expect variants with ec-3, mp4a, and/or not in range to be stripped out",
			filters: &parsers.MediaFilters{
				MinBitrate: 4000,
				MaxBitrate: 6000,
				VideoFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
				},
				AudioFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
					Codecs:     []parsers.Codec{"ec-3"},
				},
			},
			manifestContent:       manifestWithAllCodecsAndBandwidths,
			expectManifestContent: manifestFilter4000To6000BandwidthAndAC3,
		},
		{
			name: "when filtering out video (avc and hevc) and filtering in bandwidth range 4000-6000, expect variants with avc, hevc, and/or not in range to be stripped out",
			filters: &parsers.MediaFilters{
				MinBitrate: 4000,
				MaxBitrate: 6000,
				VideoFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
					Codecs:     []parsers.Codec{"avc", "hvc"},
				},
				AudioFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
				},
			},
			manifestContent:       manifestWithAllCodecsAndBandwidths,
			expectManifestContent: manifestFilter4000To6000BandwidthAndDVH,
		},
		{
			name: "when filtering out audio (ac-3, mp4a) and video (hevc and dvh) and filtering in bandwidth range 4000-6000, expect variants with ac-3, mp4a, hevc, dvh, and/or not in range to be stripped out",
			filters: &parsers.MediaFilters{
				MinBitrate: 4000,
				MaxBitrate: 6000,
				VideoFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
					Codecs:     []parsers.Codec{"hvc", "dvh"},
				},
				AudioFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
					Codecs:     []parsers.Codec{"ac-3", "mp4a"},
				},
			},
			manifestContent:       manifestWithAllCodecsAndBandwidths,
			expectManifestContent: manifestFilter4000To6000BandwidthAndEC3AndAVC,
		},
		{
			name: "when filtering out captions (stpp) and filtering in bandwidth range 4000-6000, expect variants with stpp and/or not in range to be stripped out",
			filters: &parsers.MediaFilters{
				CaptionTypes: []parsers.CaptionType{"stpp"},
				MinBitrate:   4000,
				MaxBitrate:   6000,
				VideoFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
				},
				AudioFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
				},
			},
			manifestContent:       manifestWithAllCodecsAndBandwidths,
			expectManifestContent: manifestFilter4000To6000BandwidthAndWVTT,
		},
		{
			name: "when filtering out audio and filtering in bandwidth range 4000-6000, expect variants with ac-3, ec-3, mp4a, and/or not in range to be stripped out",
			filters: &parsers.MediaFilters{
				MinBitrate: 4000,
				MaxBitrate: 6000,
				VideoFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
				},
				AudioFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
					Codecs:     []parsers.Codec{"ac-3", "ec-3", "mp4a"},
				},
			},
			manifestContent:       manifestWithAllCodecsAndBandwidths,
			expectManifestContent: manifestFilter4000To6000BandwidthAndNoAudio,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			filter := NewHLSFilter("", tt.manifestContent, config.Config{})
			manifest, err := filter.FilterManifest(tt.filters)

			if err != nil && !tt.expectErr {
				t.Errorf("FilterManifest() didnt expect an error to be returned, got: %v", err)
				return
			} else if err == nil && tt.expectErr {
				t.Error("FilterManifest() expected an error, got nil")
				return
			}

			if g, e := manifest, tt.expectManifestContent; g != e {
				t.Errorf("FilterManifest() wrong manifest returned)\ngot %v\nexpected: %v\ndiff: %v", g, e,
					cmp.Diff(g, e))
			}

		})
	}
}

func TestHLSFilter_FilterManifest_NormalizeVariant(t *testing.T) {

	manifestWithRelativeOnly := `#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=CLOSED-CAPTIONS,GROUP-ID="CC",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG"
#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="AU",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG",URI="audio.mp3u8"
#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="AU2",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG",URI="../../audio_nested.mp3u8"
#EXT-X-MEDIA:TYPE=VIDEO,GROUP-ID="VID",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG",URI="video.mp3u8"
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,AUDIO="AU",VIDEO="VID",CLOSED-CAPTIONS="CC"
link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CLOSED-CAPTIONS="CC"
link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,AUDIO="AU2",CLOSED-CAPTIONS="CC"
../../link_3.m3u8
`

	manifestWithAbsoluteOnly := `#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=CLOSED-CAPTIONS,GROUP-ID="CC",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG"
#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="AU",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG",URI="http://existing.base/uri/nested/folders/audio.mp3u8"
#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="AU2",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG",URI="http://existing.base/uri/audio_nested.mp3u8"
#EXT-X-MEDIA:TYPE=VIDEO,GROUP-ID="VID",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG",URI="http://existing.base/uri/nested/folders/video.mp3u8"
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,AUDIO="AU",VIDEO="VID",CLOSED-CAPTIONS="CC"
http://existing.base/uri/nested/folders/link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CLOSED-CAPTIONS="CC"
http://existing.base/uri/nested/folders/link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,AUDIO="AU2",CLOSED-CAPTIONS="CC"
http://existing.base/uri/link_3.m3u8
`

	manifestWithRelativeAndAbsolute := `#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=CLOSED-CAPTIONS,GROUP-ID="CC",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG"
#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="AU",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG",URI="audio.mp3u8"
#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="AU2",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG",URI="../../audio_nested.mp3u8"
#EXT-X-MEDIA:TYPE=VIDEO,GROUP-ID="VID",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG",URI="http://existing.base/uri/nested/folders/video.mp3u8"
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,AUDIO="AU",VIDEO="VID",CLOSED-CAPTIONS="CC"
link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CLOSED-CAPTIONS="CC"
http://existing.base/uri/nested/folders/link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,AUDIO="AU2",CLOSED-CAPTIONS="CC"
../../link_3.m3u8
`

	manifestWithDifferentAbsolute := `#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=CLOSED-CAPTIONS,GROUP-ID="CC",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG"
#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="AU",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG",URI="audio.mp3u8"
#EXT-X-MEDIA:TYPE=VIDEO,GROUP-ID="VID",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG",URI="http://different.base/uri/video.mp3u8"
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,AUDIO="AU",VIDEO="VID",CLOSED-CAPTIONS="CC"
link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CLOSED-CAPTIONS="CC"
http://different.base/uri/link_2.m3u8
`

	manifestWithDifferentAbsoluteExpected := `#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=CLOSED-CAPTIONS,GROUP-ID="CC",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG"
#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="AU",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG",URI="http://existing.base/uri/nested/folders/audio.mp3u8"
#EXT-X-MEDIA:TYPE=VIDEO,GROUP-ID="VID",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG",URI="http://different.base/uri/video.mp3u8"
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,AUDIO="AU",VIDEO="VID",CLOSED-CAPTIONS="CC"
http://existing.base/uri/nested/folders/link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CLOSED-CAPTIONS="CC"
http://different.base/uri/link_2.m3u8
`

	manifestWithIllegalAlternativeURLs := `#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=CLOSED-CAPTIONS,GROUP-ID="CC",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG"
#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="AU",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG",URI="http://exist\ing.base/uri/illegal.mp3u8"
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,AUDIO="AU",VIDEO="VID",CLOSED-CAPTIONS="CC"
http://existing.base/uri/nested/folders/link_1.m3u8
`

	manifestWithIllegalVariantURLs := `#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=CLOSED-CAPTIONS,GROUP-ID="CC",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG"
#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="AU",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG",URI="\nillegal.mp3u8"
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,AUDIO="AU",VIDEO="VID",CLOSED-CAPTIONS="CC"
http://existi\ng.base/uri/link_1.m3u8
`

	tests := []struct {
		name                  string
		filters               *parsers.MediaFilters
		manifestContent       string
		expectManifestContent string
		expectErr             bool
	}{
		{
			name: "when manifest contains only absolute uris, expect same manifest",
			filters: &parsers.MediaFilters{
				MaxBitrate: math.MaxInt32,
			},
			manifestContent:       manifestWithAbsoluteOnly,
			expectManifestContent: manifestWithAbsoluteOnly,
		},
		{
			name: "when manifest contains only relative urls, expect all urls to become absolute",
			filters: &parsers.MediaFilters{
				MaxBitrate: math.MaxInt32,
			},
			manifestContent:       manifestWithRelativeOnly,
			expectManifestContent: manifestWithAbsoluteOnly,
		},
		{
			name: "when manifest contains both absolute and relative urls, expect all urls to be absolute",
			filters: &parsers.MediaFilters{
				MaxBitrate: math.MaxInt32,
			},
			manifestContent:       manifestWithRelativeAndAbsolute,
			expectManifestContent: manifestWithAbsoluteOnly,
		},
		{
			name: "when manifest contains relative urls and absolute urls (with different base url), expect only relative urls to be changes to have base url as base",
			filters: &parsers.MediaFilters{
				MaxBitrate: math.MaxInt32,
			},
			manifestContent:       manifestWithDifferentAbsolute,
			expectManifestContent: manifestWithDifferentAbsoluteExpected,
		},
		{
			name: "when manifest contains invalid absolute urls, expect error to be returned",
			filters: &parsers.MediaFilters{
				MaxBitrate: math.MaxInt32,
			},
			manifestContent: manifestWithIllegalAlternativeURLs,
			expectErr:       true,
		},
		{
			name: "when manifest contains invalid relative urls, expect error to be returned",
			filters: &parsers.MediaFilters{
				MaxBitrate: math.MaxInt32,
			},
			manifestContent: manifestWithIllegalVariantURLs,
			expectErr:       true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			filter := NewHLSFilter("http://existing.base/uri/nested/folders/manifest_link.m3u8", tt.manifestContent, config.Config{})
			manifest, err := filter.FilterManifest(tt.filters)

			if err != nil && !tt.expectErr {
				t.Errorf("FilterManifest() didnt expect an error to be returned, got: %v", err)
				return
			} else if err == nil && tt.expectErr {
				t.Error("FilterManifest() expected an error, got nil")
				return
			} else if err != nil && tt.expectErr {
				return
			}

			if g, e := manifest, tt.expectManifestContent; g != e {
				t.Errorf("FilterManifest() wrong manifest returned\ngot %v\nexpected: %v\ndiff: %v", g, e,
					cmp.Diff(g, e))
			}

		})
	}

	badBaseManifestTest := []struct {
		name                  string
		filters               *parsers.MediaFilters
		manifestContent       string
		expectManifestContent string
		expectErr             bool
	}{
		{
			name:            "when link to manifest is invalid, expect error",
			filters:         &parsers.MediaFilters{},
			manifestContent: manifestWithRelativeOnly,
			expectErr:       true,
		},
	}

	for _, tt := range badBaseManifestTest {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			filter := NewHLSFilter("existi\ng.base/uri/manifest_link.m3u8", tt.manifestContent, config.Config{})
			manifest, err := filter.FilterManifest(tt.filters)
			if err != nil && !tt.expectErr {
				t.Errorf("FilterManifest() didnt expect an error to be returned, got: %v", err)
				return
			} else if err == nil && tt.expectErr {
				t.Error("FilterManifest() expected an error, got nil")
				return
			} else if err != nil && tt.expectErr {
				return
			}

			if g, e := manifest, tt.expectManifestContent; g != e {
				t.Errorf("FilterManifest() wrong manifest returned\ngot %v\nexpected: %v\ndiff: %v", g, e,
					cmp.Diff(g, e))
			}
		})
	}
}

func TestHLSFilter_FilterManifest_TrimFilter_MasterManifest(t *testing.T) {

	masterManifestWithAbsoluteURLs := `#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=CLOSED-CAPTIONS,GROUP-ID="CC",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG"
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,CODECS="avc1.64001f,mp4a.40.2"
https://existing.base/path/link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4200,AVERAGE-BANDWIDTH=4200,CODECS="avc1.64001f,mp4a.40.2"
https://existing.base/path/link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CODECS="avc1.64001f,mp4a.40.2"
https://existing.base/path/link_4.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4100,AVERAGE-BANDWIDTH=4100,CODECS="avc1.64001f,mp4a.40.2"
https://existing.base/path/link_5.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4500,AVERAGE-BANDWIDTH=4500,CODECS="avc1.64001f,mp4a.40.2"
https://existing.base/path/link_6.m3u8
`

	masterManifestWithRelativeURLs := `#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=CLOSED-CAPTIONS,GROUP-ID="CC",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG"
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,CODECS="avc1.64001f,mp4a.40.2"
link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4200,AVERAGE-BANDWIDTH=4200,CODECS="avc1.64001f,mp4a.40.2"
link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CODECS="avc1.64001f,mp4a.40.2"
link_4.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4100,AVERAGE-BANDWIDTH=4100,CODECS="avc1.64001f,mp4a.40.2"
link_5.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4500,AVERAGE-BANDWIDTH=4500,CODECS="avc1.64001f,mp4a.40.2"
link_6.m3u8
`

	manifestWithBase64EncodedVariantURLS := `#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=CLOSED-CAPTIONS,GROUP-ID="CC",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG"
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,CODECS="avc1.64001f,mp4a.40.2"
https://bakery.cbsi.video/t(10000,100000)/aHR0cHM6Ly9leGlzdGluZy5iYXNlL3BhdGgvbGlua18xLm0zdTg.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4200,AVERAGE-BANDWIDTH=4200,CODECS="avc1.64001f,mp4a.40.2"
https://bakery.cbsi.video/t(10000,100000)/aHR0cHM6Ly9leGlzdGluZy5iYXNlL3BhdGgvbGlua18yLm0zdTg.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CODECS="avc1.64001f,mp4a.40.2"
https://bakery.cbsi.video/t(10000,100000)/aHR0cHM6Ly9leGlzdGluZy5iYXNlL3BhdGgvbGlua180Lm0zdTg.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4100,AVERAGE-BANDWIDTH=4100,CODECS="avc1.64001f,mp4a.40.2"
https://bakery.cbsi.video/t(10000,100000)/aHR0cHM6Ly9leGlzdGluZy5iYXNlL3BhdGgvbGlua181Lm0zdTg.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4500,AVERAGE-BANDWIDTH=4500,CODECS="avc1.64001f,mp4a.40.2"
https://bakery.cbsi.video/t(10000,100000)/aHR0cHM6Ly9leGlzdGluZy5iYXNlL3BhdGgvbGlua182Lm0zdTg.m3u8
`

	manifestWithFilteredBitrateAndBase64EncodedVariantURLS := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4200,AVERAGE-BANDWIDTH=4200,CODECS="avc1.64001f,mp4a.40.2"
https://bakery.cbsi.video/t(10000,100000)/aHR0cHM6Ly9leGlzdGluZy5iYXNlL3BhdGgvbGlua18yLm0zdTg.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CODECS="avc1.64001f,mp4a.40.2"
https://bakery.cbsi.video/t(10000,100000)/aHR0cHM6Ly9leGlzdGluZy5iYXNlL3BhdGgvbGlua180Lm0zdTg.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4100,AVERAGE-BANDWIDTH=4100,CODECS="avc1.64001f,mp4a.40.2"
https://bakery.cbsi.video/t(10000,100000)/aHR0cHM6Ly9leGlzdGluZy5iYXNlL3BhdGgvbGlua181Lm0zdTg.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4500,AVERAGE-BANDWIDTH=4500,CODECS="avc1.64001f,mp4a.40.2"
https://bakery.cbsi.video/t(10000,100000)/aHR0cHM6Ly9leGlzdGluZy5iYXNlL3BhdGgvbGlua182Lm0zdTg.m3u8
`

	trim := &parsers.Trim{
		Start: 10000,
		End:   100000,
	}

	tests := []struct {
		name                  string
		filters               *parsers.MediaFilters
		manifestContent       string
		expectManifestContent string
		expectErr             bool
	}{
		{
			name: "when trim filter is given and master has absolute urls, variant level manifest will point to" +
				"bakery with trim filter and base64 encoding string in the manifest",
			filters: &parsers.MediaFilters{
				MaxBitrate: math.MaxInt32,
				VideoFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
				},
				AudioFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
				},
				Trim: trim,
			},
			manifestContent:       masterManifestWithAbsoluteURLs,
			expectManifestContent: manifestWithBase64EncodedVariantURLS,
		},
		{
			name: "when trim filter is given and master has relative urls, variant level manifest will point to" +
				"bakery with trim filter and base64 encoding string in the manifest",
			filters: &parsers.MediaFilters{
				MaxBitrate: math.MaxInt32,
				VideoFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
				},
				AudioFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
				},
				Trim: trim,
			},
			manifestContent:       masterManifestWithRelativeURLs,
			expectManifestContent: manifestWithBase64EncodedVariantURLS,
		},
		{
			name: "when bitrate and trim filter are given, variant level manifest will point to" +
				"bakery with only included bitrates, the trim filter, and base64 encoding string in the manifest",
			filters: &parsers.MediaFilters{
				MinBitrate: 4000,
				MaxBitrate: 6000,
				Trim:       trim,
				VideoFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
				},
				AudioFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
				},
			},
			manifestContent:       masterManifestWithRelativeURLs,
			expectManifestContent: manifestWithFilteredBitrateAndBase64EncodedVariantURLS,
		},
		{
			name: "when no filter is given, variant level manifest will hold absolute urls only",
			filters: &parsers.MediaFilters{
				MaxBitrate: math.MaxInt32,
				VideoFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
				},
				AudioFilters: parsers.NestedFilters{
					MaxBitrate: math.MaxInt32,
				},
			},
			manifestContent:       masterManifestWithRelativeURLs,
			expectManifestContent: masterManifestWithAbsoluteURLs,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			filter := NewHLSFilter("https://existing.base/path/master.m3u8", tt.manifestContent, config.Config{Hostname: "bakery.cbsi.video"})
			manifest, err := filter.FilterManifest(tt.filters)

			if err != nil && !tt.expectErr {
				t.Errorf("FilterManifest() didnt expect an error to be returned, got: %v", err)
				return
			} else if err == nil && tt.expectErr {
				t.Error("FilterManifest() expected an error, got nil")
				return
			}

			if g, e := manifest, tt.expectManifestContent; g != e {
				t.Errorf("FilterManifest() wrong manifest returned)\ngot %v\nexpected: %v\ndiff: %v", g, e,
					cmp.Diff(g, e))
			}

		})
	}
}

func TestHLSFilter_FilterManifest_TrimFilter_VariantManifest(t *testing.T) {

	variantManifestWithRelativeURLs := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-MEDIA-SEQUENCE:10
#EXT-X-TARGETDURATION:6
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:51:48Z
#EXTINF:6.000,
chan_1/chan_1_20200311T202743_1_00019.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:51:54Z
#EXTINF:6.000,
chan_1/chan_1_20200311T202748_1_00020.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:00Z
#EXTINF:6.000,
chan_1/chan_1_20200311T202754_1_00021.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:06Z
#EXTINF:6.000,
chan_1/chan_1_20200311T202801_1_00022.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:12Z
#EXTINF:6.000,
chan_1/chan_1_20200311T202806_1_00023.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:18Z
#EXTINF:6.000,
chan_1/chan_1_20200311T202813_1_00024.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:24Z
#EXTINF:6.000,
chan_1/chan_1_20200311T202818_1_00025.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:30Z
#EXTINF:6.000,
chan_1/chan_1_20200311T202824_1_00026.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:36Z
#EXTINF:6.000,
chan_1/chan_1_20200311T202818_1_00027.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:42Z
#EXTINF:6.000,
chan_1/chan_1_20200311T202824_1_00028.ts
`

	variantManifestWithAbsoluteURLs := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-MEDIA-SEQUENCE:10
#EXT-X-TARGETDURATION:6
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:51:48Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202743_1_00019.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:51:54Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202748_1_00020.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:00Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202754_1_00021.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:06Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202801_1_00022.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:12Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202806_1_00023.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:18Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202813_1_00024.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:24Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202818_1_00025.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:30Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202824_1_00026.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:36Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202818_1_00027.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:42Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202824_1_00028.ts
`

	variantManifestWithNoPDT := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-MEDIA-SEQUENCE:10
#EXT-X-TARGETDURATION:6
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202743_1_00019.ts
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202748_1_00020.ts
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202754_1_00021.ts
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202801_1_00022.ts
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202806_1_00023.ts
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202813_1_00024.ts
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202818_1_00025.ts
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202824_1_00026.ts
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202818_1_00027.ts
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202824_1_00028.ts
`

	variantManifestTrimmed := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-TARGETDURATION:6
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:00Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202754_1_00021.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:06Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202801_1_00022.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:12Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202806_1_00023.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:18Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202813_1_00024.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:24Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202818_1_00025.ts
#EXT-X-ENDLIST
`

	trim := &parsers.Trim{
		Start: 1583887920, //2020-03-11T00:52:00
		End:   1583887944, //2020-03-11T00:52:24
	}

	tests := []struct {
		name                  string
		filters               *parsers.MediaFilters
		manifestContent       string
		expectManifestContent string
		expectErr             bool
	}{
		{
			name: "when trim filter is given, variant level manifest have absolute url with base64" +
				"encoding string in the manifest",
			filters:               &parsers.MediaFilters{Trim: trim},
			manifestContent:       variantManifestWithRelativeURLs,
			expectManifestContent: variantManifestTrimmed,
		},
		{
			name:                  "when no filter is given, variant level manifest will hold absolute urls only",
			filters:               &parsers.MediaFilters{Trim: trim},
			manifestContent:       variantManifestWithAbsoluteURLs,
			expectManifestContent: variantManifestTrimmed,
		},
		{
			name:                  "when no pdt present for segment, error is thrown",
			filters:               &parsers.MediaFilters{Trim: trim},
			manifestContent:       variantManifestWithNoPDT,
			expectManifestContent: "",
			expectErr:             true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			filter := NewHLSFilter("https://existing.base/path/master.m3u8", tt.manifestContent, config.Config{Hostname: "bakery.cbsi.video"})
			manifest, err := filter.FilterManifest(tt.filters)

			if err != nil && !tt.expectErr {
				t.Errorf("FilterManifest() didnt expect an error to be returned, got: %v", err)
				return
			} else if err == nil && tt.expectErr {
				t.Error("FilterManifest() expected an error, got nil")
				return
			}

			if g, e := manifest, tt.expectManifestContent; g != e {
				t.Errorf("FilterManifest() wrong manifest returned)\ngot %v\nexpected: %v\ndiff: %v", g, e,
					cmp.Diff(g, e))
			}

		})
	}
}
