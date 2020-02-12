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
#EXT-X-MEDIA:TYPE=CLOSED-CAPTIONS,GROUP-ID="CC",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG",URI="http://existing.base/uri/"
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,CLOSED-CAPTIONS="CC"
http://existing.base/uri/link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CLOSED-CAPTIONS="CC"
http://existing.base/uri/link_2.m3u8
`

	manifestRemovedLowerBW := `#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=CLOSED-CAPTIONS,GROUP-ID="CC",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG",URI="http://existing.base/uri/"
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,CLOSED-CAPTIONS="CC"
http://existing.base/uri/link_1.m3u8
`

	manifestRemovedHigherBW := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CLOSED-CAPTIONS="CC"
http://existing.base/uri/link_2.m3u8
`

	tests := []struct {
		name                  string
		filters               *parsers.MediaFilters
		manifestContent       string
		expectManifestContent string
		expectErr             bool
	}{
		{
			name:                  "when no bitrate filters given, expect unfiltered manifest",
			filters:               &parsers.MediaFilters{MinBitrate: 0, MaxBitrate: math.MaxInt32},
			manifestContent:       baseManifest,
			expectManifestContent: baseManifest,
		},
		{
			name:                  "when negative bitrates entered, expect unfiltered manifest",
			filters:               &parsers.MediaFilters{MinBitrate: -1000, MaxBitrate: -100},
			manifestContent:       baseManifest,
			expectManifestContent: baseManifest,
		},
		{
			name:                  "when both bitrate bounds are exceeded, expect unfiltered manifest",
			filters:               &parsers.MediaFilters{MinBitrate: -100, MaxBitrate: math.MaxInt32 + 1},
			manifestContent:       baseManifest,
			expectManifestContent: baseManifest,
		},
		{
			name:                  "when lower bitrate bound is greater than upper bound, expect unfiltered manifest",
			filters:               &parsers.MediaFilters{MinBitrate: 1000, MaxBitrate: 100},
			manifestContent:       baseManifest,
			expectManifestContent: baseManifest,
		},
		{
			name:                  "when only hitting lower boundary (MinBitrate = 0), expect results to be filtered",
			filters:               &parsers.MediaFilters{MinBitrate: 0, MaxBitrate: 3000},
			manifestContent:       baseManifest,
			expectManifestContent: manifestRemovedLowerBW,
		},
		{
			name:                  "when only hitting upper boundary (MaxBitrate = math.MaxInt32), expect results to be filtered",
			filters:               &parsers.MediaFilters{MinBitrate: 3000, MaxBitrate: math.MaxInt32},
			manifestContent:       baseManifest,
			expectManifestContent: manifestRemovedHigherBW,
		},
		{
			name:                  "when invalid minimum bitrate and valid maximum bitrate, expect unfiltered manifest",
			filters:               &parsers.MediaFilters{MinBitrate: -100, MaxBitrate: 2000},
			manifestContent:       baseManifest,
			expectManifestContent: baseManifest,
		},
		{
			name:                  "when valid minimum bitrate and invlid maximum bitrate, expect unfiltered manifest",
			filters:               &parsers.MediaFilters{MinBitrate: 3000, MaxBitrate: math.MaxInt32 + 1},
			manifestContent:       baseManifest,
			expectManifestContent: baseManifest,
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
`

	manifestFilterInAC3 := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1100,AVERAGE-BANDWIDTH=1100,CODECS="ac-3"
http://existing.base/uri/link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1500,AVERAGE-BANDWIDTH=1500,CODECS="avc1.77.30"
http://existing.base/uri/link_6.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300,CODECS="wvtt"
http://existing.base/uri/link_7.m3u8
`

	manifestFilterInMP4A := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CODECS="mp4a.40.2"
http://existing.base/uri/link_3.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1500,AVERAGE-BANDWIDTH=1500,CODECS="avc1.77.30"
http://existing.base/uri/link_6.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300,CODECS="wvtt"
http://existing.base/uri/link_7.m3u8
`

	manifestFilterInEC3AndAC3 := `#EXTM3U
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
`

	manifestFilterInEC3AndMP4A := `#EXTM3U
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
`

	manifestWithoutAudio := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1500,AVERAGE-BANDWIDTH=1500,CODECS="avc1.77.30"
http://existing.base/uri/link_6.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300,CODECS="wvtt"
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
			name:                  "when given empty audio filter list, expect manifest with no audio",
			filters:               &parsers.MediaFilters{Audios: []parsers.AudioType{}},
			manifestContent:       manifestWithAllAudio,
			expectManifestContent: manifestWithoutAudio,
		},
		{
			name:                  "when filtering in ec-3, expect manifest without ac-3 or mp4a",
			filters:               &parsers.MediaFilters{Audios: []parsers.AudioType{"ec-3"}},
			manifestContent:       manifestWithAllAudio,
			expectManifestContent: manifestFilterInEC3,
		},
		{
			name:                  "when filtering in ac-3, expect manifest without ec-3 or mp4a",
			filters:               &parsers.MediaFilters{Audios: []parsers.AudioType{"ac-3"}},
			manifestContent:       manifestWithAllAudio,
			expectManifestContent: manifestFilterInAC3,
		},
		{
			name:                  "when filtering in mp4a, expect manifest without ec-3 or ac-3",
			filters:               &parsers.MediaFilters{Audios: []parsers.AudioType{"mp4a"}},
			manifestContent:       manifestWithAllAudio,
			expectManifestContent: manifestFilterInMP4A,
		},
		{
			name:                  "when filtering in ec-3 and ac-3, expect manifest without any variants containing mp4a",
			filters:               &parsers.MediaFilters{Audios: []parsers.AudioType{"ec-3", "ac-3"}},
			manifestContent:       manifestWithAllAudio,
			expectManifestContent: manifestFilterInEC3AndAC3,
		},
		{
			name:                  "when filtering in ec-3 and mp4a, expect manifest without any variants containing ac-3",
			filters:               &parsers.MediaFilters{Audios: []parsers.AudioType{"ec-3", "mp4a"}},
			manifestContent:       manifestWithAllAudio,
			expectManifestContent: manifestFilterInEC3AndMP4A,
		},
		{
			name:                  "when no audio filters are given, expect unfiltered manifest",
			filters:               &parsers.MediaFilters{},
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
`

	manifestFilterInAVC := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,CODECS="avc1.640020"
http://existing.base/uri/link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1100,AVERAGE-BANDWIDTH=1100,CODECS="avc1.77.30"
http://existing.base/uri/link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=6000,AVERAGE-BANDWIDTH=6000,CODECS="avc1.77.30,ec-3"
http://existing.base/uri/link_6.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1500,AVERAGE-BANDWIDTH=1500,CODECS="ec-3"
http://existing.base/uri/link_7.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300,CODECS="wvtt"
http://existing.base/uri/link_8.m3u8
`

	manifestFilterInHEVC := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CODECS="hvc1.2.4.L93.90"
http://existing.base/uri/link_3.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1500,AVERAGE-BANDWIDTH=1500,CODECS="ec-3"
http://existing.base/uri/link_7.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300,CODECS="wvtt"
http://existing.base/uri/link_8.m3u8
`

	manifestFilterInDVH := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4500,AVERAGE-BANDWIDTH=4500,CODECS="dvh1.05.01"
http://existing.base/uri/link_4.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1500,AVERAGE-BANDWIDTH=1500,CODECS="ec-3"
http://existing.base/uri/link_7.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300,CODECS="wvtt"
http://existing.base/uri/link_8.m3u8
`

	manifestFilterInAVCAndHEVC := `#EXTM3U
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
`

	manifestFilterInAVCAndDVH := `#EXTM3U
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
`

	manifestWithoutVideo := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1500,AVERAGE-BANDWIDTH=1500,CODECS="ec-3"
http://existing.base/uri/link_7.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1300,AVERAGE-BANDWIDTH=1300,CODECS="wvtt"
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
			name:                  "when given empty video filter list, expect manifest with no video",
			filters:               &parsers.MediaFilters{Videos: []parsers.VideoType{}},
			manifestContent:       manifestWithAllVideo,
			expectManifestContent: manifestWithoutVideo,
		},
		{
			name:                  "when filtering in avc, expect manifest without any variants containing other video codecs",
			filters:               &parsers.MediaFilters{Videos: []parsers.VideoType{"avc"}},
			manifestContent:       manifestWithAllVideo,
			expectManifestContent: manifestFilterInAVC,
		},
		{
			name:                  "when filtering in hevc, expect manifest without any variants containing other video codecs",
			filters:               &parsers.MediaFilters{Videos: []parsers.VideoType{"hvc"}},
			manifestContent:       manifestWithAllVideo,
			expectManifestContent: manifestFilterInHEVC,
		},
		{
			name:                  "when filtering in dvh, expect manifest without any variants containing other video codecs",
			filters:               &parsers.MediaFilters{Videos: []parsers.VideoType{"dvh"}},
			manifestContent:       manifestWithAllVideo,
			expectManifestContent: manifestFilterInDVH,
		},
		{
			name:                  "when filtering in avc and hevc, expect manifest without any variants containing dvh",
			filters:               &parsers.MediaFilters{Videos: []parsers.VideoType{"avc", "hvc"}},
			manifestContent:       manifestWithAllVideo,
			expectManifestContent: manifestFilterInAVCAndHEVC,
		},
		{
			name:                  "when filtering in avc and dvh, expect manifest without any variants containing hvc",
			filters:               &parsers.MediaFilters{Videos: []parsers.VideoType{"avc", "dvh"}},
			manifestContent:       manifestWithAllVideo,
			expectManifestContent: manifestFilterInAVCAndDVH,
		},
		{
			name:                  "when no video filters are given, expect unfiltered manifest",
			filters:               &parsers.MediaFilters{},
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

}
