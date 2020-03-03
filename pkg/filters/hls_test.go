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
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,CLOSED-CAPTIONS="CC"
http://existing.base/uri/link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CLOSED-CAPTIONS="CC"
http://existing.base/uri/link_2.m3u8
`

	manifestRemovedLowerBW := `#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=CLOSED-CAPTIONS,GROUP-ID="CC",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG"
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
			name:                  "when all audio codecs are supplies, expect audio to be stripped out",
			filters:               &parsers.MediaFilters{Audios: []parsers.AudioType{"mp4a", "ec-3", "ac-3"}},
			manifestContent:       manifestWithAllAudio,
			expectManifestContent: manifestWithoutAudio,
		},
		{
			name:                  "when filtering in ac-3 and mp4a, expect variants with ac-3 and/or mp4a to be stripped out",
			filters:               &parsers.MediaFilters{Audios: []parsers.AudioType{"ac-3", "mp4a"}},
			manifestContent:       manifestWithAllAudio,
			expectManifestContent: manifestFilterInEC3,
		},
		{
			name:                  "when filtering in ac-3, expect variants with ac-3 to be stripped out",
			filters:               &parsers.MediaFilters{Audios: []parsers.AudioType{"ac-3"}},
			manifestContent:       manifestWithAllAudio,
			expectManifestContent: manifestFilterWithoutAC3,
		},
		{
			name:                  "when filtering in mp4a, expect variants with mp4a to be stripped out",
			filters:               &parsers.MediaFilters{Audios: []parsers.AudioType{"mp4a"}},
			manifestContent:       manifestWithAllAudio,
			expectManifestContent: manifestFilterWithoutMP4A,
		},
		{
			name:                  "when filtering in ec-3 and ac-3, expect variants with ec-3 and ac-3 to be stripped out",
			filters:               &parsers.MediaFilters{Audios: []parsers.AudioType{"ec-3", "ac-3"}},
			manifestContent:       manifestWithAllAudio,
			expectManifestContent: manifestFilterInMP4A,
		},
		{
			name:                  "when filtering in ec-3 and mp4a, expect variants with ec-3 and/or mp4a to be stripped out",
			filters:               &parsers.MediaFilters{Audios: []parsers.AudioType{"ec-3", "mp4a"}},
			manifestContent:       manifestWithAllAudio,
			expectManifestContent: manifestFilterInAC3,
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
			name:                  "when all video codecs are supllied, expect variants with avc, hevc, and/or dvh to be stripped out",
			filters:               &parsers.MediaFilters{Videos: []parsers.VideoType{"avc", "hvc", "dvh"}},
			manifestContent:       manifestWithAllVideo,
			expectManifestContent: manifestWithoutVideo,
		},
		{
			name:                  "when filtering in avc, expect variants with avc to be stripped out",
			filters:               &parsers.MediaFilters{Videos: []parsers.VideoType{"avc"}},
			manifestContent:       manifestWithAllVideo,
			expectManifestContent: manifestFilterWithoutAVC,
		},
		{
			name:                  "when filtering in hevc, expect hevc to be stripped out",
			filters:               &parsers.MediaFilters{Videos: []parsers.VideoType{"hvc"}},
			manifestContent:       manifestWithAllVideo,
			expectManifestContent: manifestFilterWithoutHEVC,
		},
		{
			name:                  "when filtering in dvh, expect dvh to be stripped out",
			filters:               &parsers.MediaFilters{Videos: []parsers.VideoType{"dvh"}},
			manifestContent:       manifestWithAllVideo,
			expectManifestContent: manifestFilterWithoutDVH,
		},
		{
			name:                  "when filtering in avc and hevc, expect variants with avc and hevc to be stripped out",
			filters:               &parsers.MediaFilters{Videos: []parsers.VideoType{"avc", "hvc"}},
			manifestContent:       manifestWithAllVideo,
			expectManifestContent: manifestFilterWithoutAVCAndHEVC,
		},
		{
			name:                  "when filtering in avc and dvh, expect variants with avc and dvh to be stripped out",
			filters:               &parsers.MediaFilters{Videos: []parsers.VideoType{"avc", "dvh"}},
			manifestContent:       manifestWithAllVideo,
			expectManifestContent: manifestFilterWithoutAVCAndDVH,
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
			name:                  "when all caption filters are supplied, expect all caption variants with captions to be stripped out",
			filters:               &parsers.MediaFilters{CaptionTypes: []parsers.CaptionType{"stpp", "wvtt"}},
			manifestContent:       manifestWithAllCaptions,
			expectManifestContent: manifestWithNoCaptions,
		},
		{
			name:                  "when filtering in wvtt, expect variants with wvtt to be stripped out",
			filters:               &parsers.MediaFilters{CaptionTypes: []parsers.CaptionType{"wvtt"}},
			manifestContent:       manifestWithAllCaptions,
			expectManifestContent: manifestFilterWithoutWVTT,
		},
		{
			name:                  "when filtering in stpp, expect variants with wvtt to be stripped out",
			filters:               &parsers.MediaFilters{CaptionTypes: []parsers.CaptionType{"stpp"}},
			manifestContent:       manifestWithAllCaptions,
			expectManifestContent: manifestFilterWithoutSTPP,
		},
		{
			name:                  "when no caption filter is given, expect original manifest",
			filters:               &parsers.MediaFilters{},
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
			name:                  "when empty filters are given, expect original manifest",
			filters:               &parsers.MediaFilters{},
			manifestContent:       manifestWithAllCodecs,
			expectManifestContent: manifestWithAllCodecs,
		},
		{
			name:                  "when filtering out audio (ec-3 and mp4a) and video (hevc and dvh), expect variants with ec-3, mp4a, hevc, and/or dvh to be stripped out",
			filters:               &parsers.MediaFilters{Audios: []parsers.AudioType{"ec-3", "mp4a"}, Videos: []parsers.VideoType{"hvc", "dvh"}},
			manifestContent:       manifestWithAllCodecs,
			expectManifestContent: manifestFilterInAC3AndAVC,
		},
		{
			name:                  "when filtering out audio (mp4a) and video (hevc and dvh), expect variants with mp4a, hevc, and/or dvh to be stripped out",
			filters:               &parsers.MediaFilters{Audios: []parsers.AudioType{"mp4a"}, Videos: []parsers.VideoType{"hvc", "dvh"}},
			manifestContent:       manifestWithAllCodecs,
			expectManifestContent: manifestFilterInAC3AndEC3AndAVC,
		},
		{
			name:                  "when filtering out audio (ec-3 and mp4a) and captions (stpp), expect variants with ec-3, mp4a, and/or stpp to be stripped out",
			filters:               &parsers.MediaFilters{Audios: []parsers.AudioType{"ec-3", "mp4a"}, CaptionTypes: []parsers.CaptionType{"stpp"}},
			manifestContent:       manifestWithAllCodecs,
			expectManifestContent: manifestFilterInAC3AndWVTT,
		},
		{
			name:                  "when filtering out audio (ec-3 and mp4a), video (hevc and dvh), and captions (stpp), expect variants with ec-3, mp4a, hevc, dvh, and/or stpp to be stripped out",
			filters:               &parsers.MediaFilters{Audios: []parsers.AudioType{"ec-3", "mp4a"}, Videos: []parsers.VideoType{"hvc", "dvh"}, CaptionTypes: []parsers.CaptionType{"stpp"}},
			manifestContent:       manifestWithAllCodecs,
			expectManifestContent: manifestFilterInAC3AndAVCAndWVTT,
		},
		{
			name:                  "when filtering out all codecs except avc video, expect variants with ac-3, ec-3, mp4a, hevc, and/or dvh to be stripped out",
			filters:               &parsers.MediaFilters{Audios: []parsers.AudioType{"ac-3", "ec-3", "mp4a"}, Videos: []parsers.VideoType{"hvc", "dvh"}, CaptionTypes: []parsers.CaptionType{"wvtt", "stpp"}},
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
			name:                  "when no filters are given, expect original manifest",
			filters:               &parsers.MediaFilters{},
			manifestContent:       manifestWithAllCodecsAndBandwidths,
			expectManifestContent: manifestWithAllCodecsAndBandwidths,
		},
		{
			name:                  "when filtering out audio (ec-3) in bandwidth range 4000-6000, expect variants with ec-3, mp4a, and/or not in range to be stripped out",
			filters:               &parsers.MediaFilters{Audios: []parsers.AudioType{"ec-3"}, MinBitrate: 4000, MaxBitrate: 6000},
			manifestContent:       manifestWithAllCodecsAndBandwidths,
			expectManifestContent: manifestFilter4000To6000BandwidthAndAC3,
		},
		{
			name:                  "when filtering out video (avc and hevc) in bandwidth range 4000-6000, expect variants with avc, hevc, and/or not in range to be stripped out",
			filters:               &parsers.MediaFilters{Videos: []parsers.VideoType{"avc", "hvc"}, MinBitrate: 4000, MaxBitrate: 6000},
			manifestContent:       manifestWithAllCodecsAndBandwidths,
			expectManifestContent: manifestFilter4000To6000BandwidthAndDVH,
		},
		{
			name:                  "when filtering in audio (ac-3, mp4a) and video (hevc and dvh) in bandwidth range 4000-6000, expect variants with ac-3, mp4a, hevc, dvh, and/or not in range to be stripped out",
			filters:               &parsers.MediaFilters{Audios: []parsers.AudioType{"ac-3", "mp4a"}, Videos: []parsers.VideoType{"hvc", "dvh"}, MinBitrate: 4000, MaxBitrate: 6000},
			manifestContent:       manifestWithAllCodecsAndBandwidths,
			expectManifestContent: manifestFilter4000To6000BandwidthAndEC3AndAVC,
		},
		{
			name:                  "when filtering out captions (stpp) in bandwidth range 4000-6000, expect variants with stpp and/or not in range to be stripped out",
			filters:               &parsers.MediaFilters{CaptionTypes: []parsers.CaptionType{"stpp"}, MinBitrate: 4000, MaxBitrate: 6000},
			manifestContent:       manifestWithAllCodecsAndBandwidths,
			expectManifestContent: manifestFilter4000To6000BandwidthAndWVTT,
		},
		{
			name:                  "when filtering out audio and filtering in bandwidth range 4000-6000, expect variants with ac-3, ec-3, mp4a, and/or not in range to be stripped out",
			filters:               &parsers.MediaFilters{Audios: []parsers.AudioType{"ac-3", "ec-3", "mp4a"}, MinBitrate: 4000, MaxBitrate: 6000},
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
			name:                  "when manifest contains only absolute uris, expect same manifest",
			filters:               &parsers.MediaFilters{},
			manifestContent:       manifestWithAbsoluteOnly,
			expectManifestContent: manifestWithAbsoluteOnly,
		},
		{
			name:                  "when manifest contains only relative urls, expect all urls to become absolute",
			filters:               &parsers.MediaFilters{},
			manifestContent:       manifestWithRelativeOnly,
			expectManifestContent: manifestWithAbsoluteOnly,
		},
		{
			name:                  "when manifest contains both absolute and relative urls, expect all urls to be absolute",
			filters:               &parsers.MediaFilters{},
			manifestContent:       manifestWithRelativeAndAbsolute,
			expectManifestContent: manifestWithAbsoluteOnly,
		},
		{
			name:                  "when manifest contains relative urls and absolute urls (with different base url), expect only relative urls to be changes to have base url as base",
			filters:               &parsers.MediaFilters{},
			manifestContent:       manifestWithDifferentAbsolute,
			expectManifestContent: manifestWithDifferentAbsoluteExpected,
		},
		{
			name:            "when manifest contains invalid absolute urls, expect error to be returned",
			filters:         &parsers.MediaFilters{},
			manifestContent: manifestWithIllegalAlternativeURLs,
			expectErr:       true,
		},
		{
			name:            "when manifest contains invalid relative urls, expect error to be returned",
			filters:         &parsers.MediaFilters{},
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
