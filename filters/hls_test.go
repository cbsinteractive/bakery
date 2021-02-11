package filters

import (
	"bytes"
	"context"
	"io/ioutil"
	"math"
	"net/http"
	"testing"
	"time"

	"github.com/cbsinteractive/bakery/config"
	"github.com/cbsinteractive/bakery/parsers"
	test "github.com/cbsinteractive/bakery/tests"
	"github.com/cbsinteractive/pkg/tracing"
	"github.com/google/go-cmp/cmp"
)

func TestHLSFilter_FilterContent_BandwidthFilter(t *testing.T) {

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
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=CLOSED-CAPTIONS,GROUP-ID="CC",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG"
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
			name:                  "when no bitrate filters given, expect unfiltered manifest",
			filters:               &parsers.MediaFilters{},
			manifestContent:       baseManifest,
			expectManifestContent: baseManifest,
		},
		{
			name: "when setting lower boundary (Min = 0) overall, expect video and audio results to be filtered",
			filters: &parsers.MediaFilters{
				Videos: parsers.NestedFilters{
					Bitrate: &parsers.Bitrate{
						Min: 0, Max: 3000,
					},
				},
				Audios: parsers.NestedFilters{
					Bitrate: &parsers.Bitrate{
						Min: 0, Max: 3000,
					},
				},
			},
			manifestContent:       baseManifest,
			expectManifestContent: manifestRemovedHigherBW,
		},
		{
			name: "when setting upper boundary (Max = math.MaxInt32) overall, expect video and audio results to be filtered",
			filters: &parsers.MediaFilters{
				Videos: parsers.NestedFilters{
					Bitrate: &parsers.Bitrate{
						Min: 3000, Max: math.MaxInt32,
					},
				},
				Audios: parsers.NestedFilters{
					Bitrate: &parsers.Bitrate{
						Min: 3000, Max: math.MaxInt32,
					},
				},
			},
			manifestContent:       baseManifest,
			expectManifestContent: manifestRemovedLowerBW,
		},
		{
			name: "when setting audio only bitrate, expect STREAM with video + audio to be filtered",
			filters: &parsers.MediaFilters{
				Audios: parsers.NestedFilters{
					Bitrate: &parsers.Bitrate{
						Min: 3000, Max: 4000,
					},
				},
			},
			manifestContent:       baseManifest,
			expectManifestContent: manifestRemovedLowerBW,
		},
		{
			name: "when setting audio bitrate, expect audio to be filtered",
			filters: &parsers.MediaFilters{
				Audios: parsers.NestedFilters{
					Bitrate: &parsers.Bitrate{
						Min: 1000, Max: 2000,
					},
				},
			},
			manifestContent:       baseManifest,
			expectManifestContent: manifestRemovedHigherBWOnlyAudio,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			filter := NewHLSFilter("", tt.manifestContent, config.Config{})
			manifest, err := filter.FilterContent(context.Background(), tt.filters)

			if err != nil && !tt.expectErr {
				t.Errorf("FilterContent(context.Background(), ) didnt expect an error to be returned, got: %v", err)
				return
			} else if err == nil && tt.expectErr {
				t.Error("FilterContent(context.Background(), ) expected an error, got nil")
				return
			}

			if g, e := manifest, tt.expectManifestContent; g != e {
				t.Errorf("FilterContent(context.Background(), ) wrong manifest returned\ngot %v\nexpected: %v\ndiff: %v", g, e,
					cmp.Diff(g, e))
			}

		})
	}
}

func TestHLSFilter_FilterContent_AudioFilter(t *testing.T) {
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
				Audios: parsers.NestedFilters{
					Codecs: []string{"mp4a", "ec-3", "ac-3"},
				},
			},
			manifestContent:       manifestWithAllAudio,
			expectManifestContent: manifestWithoutAudio,
		},
		{
			name: "when filter is supplied with ac-3 and mp4a, expect variants with ac-3 and/or mp4a to be stripped out",
			filters: &parsers.MediaFilters{
				Audios: parsers.NestedFilters{
					Codecs: []string{"mp4a", "ac-3"},
				},
			},
			manifestContent:       manifestWithAllAudio,
			expectManifestContent: manifestFilterInEC3,
		},
		{
			name: "when filter is supplied with ac-3, expect variants with ac-3 to be stripped out",
			filters: &parsers.MediaFilters{
				Audios: parsers.NestedFilters{
					Codecs: []string{"ac-3"},
				},
			},
			manifestContent:       manifestWithAllAudio,
			expectManifestContent: manifestFilterWithoutAC3,
		},
		{
			name: "when filter is supplied with mp4a, expect variants with mp4a to be stripped out",
			filters: &parsers.MediaFilters{
				Audios: parsers.NestedFilters{
					Codecs: []string{"mp4a"},
				},
			},
			manifestContent:       manifestWithAllAudio,
			expectManifestContent: manifestFilterWithoutMP4A,
		},
		{
			name: "when filter is supplied with ec-3 and ac-3, expect variants with ec-3 and ac-3 to be stripped out",
			filters: &parsers.MediaFilters{
				Audios: parsers.NestedFilters{
					Codecs: []string{"ec-3", "ac-3"},
				},
			},
			manifestContent:       manifestWithAllAudio,
			expectManifestContent: manifestFilterInMP4A,
		},
		{
			name: "when filter is supplied with ec-3 and mp4a, expect variants with ec-3 and/or mp4a to be stripped out",
			filters: &parsers.MediaFilters{
				Audios: parsers.NestedFilters{
					Codecs: []string{"mp4a", "ec-3"},
				},
			},
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
			manifest, err := filter.FilterContent(context.Background(), tt.filters)

			if err != nil && !tt.expectErr {
				t.Errorf("FilterContent(context.Background(), ) didnt expect an error to be returned, got: %v", err)
				return
			} else if err == nil && tt.expectErr {
				t.Error("FilterContent(context.Background(), ) expected an error, got nil")
				return
			} else if err != nil && tt.expectErr {
				return
			}

			if g, e := manifest, tt.expectManifestContent; g != e {
				t.Errorf("FilterContent(context.Background(), ) wrong manifest returned)\ngot %v\nexpected: %v\ndiff: %v", g, e,
					cmp.Diff(g, e))
			}

		})
	}
}

func TestHLSFilter_FilterContent_VideoFilter(t *testing.T) {
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
				Videos: parsers.NestedFilters{
					Codecs: []string{"avc", "hvc", "dvh"},
				},
			},
			manifestContent:       manifestWithAllVideo,
			expectManifestContent: manifestWithoutVideo,
		},
		{
			name: "when filter is supplied with avc, expect variants with avc to be stripped out",
			filters: &parsers.MediaFilters{
				Videos: parsers.NestedFilters{
					Codecs: []string{"avc"},
				},
			},
			manifestContent:       manifestWithAllVideo,
			expectManifestContent: manifestFilterWithoutAVC,
		},
		{
			name: "when filter is supplied with hevc, expect hevc to be stripped out",
			filters: &parsers.MediaFilters{
				Videos: parsers.NestedFilters{
					Codecs: []string{"hvc"},
				},
			},
			manifestContent:       manifestWithAllVideo,
			expectManifestContent: manifestFilterWithoutHEVC,
		},
		{
			name: "when filter is supplied with dvh, expect dvh to be stripped out",
			filters: &parsers.MediaFilters{
				Videos: parsers.NestedFilters{
					Codecs: []string{"dvh"},
				},
			},
			manifestContent:       manifestWithAllVideo,
			expectManifestContent: manifestFilterWithoutDVH,
		},
		{
			name: "when filter is supplied with avc and hevc, expect variants with avc and hevc to be stripped out",
			filters: &parsers.MediaFilters{
				Videos: parsers.NestedFilters{
					Codecs: []string{"avc", "hvc"},
				},
			},
			manifestContent:       manifestWithAllVideo,
			expectManifestContent: manifestFilterWithoutAVCAndHEVC,
		},
		{
			name: "when filter is supplied with avc and dvh, expect variants with avc and dvh to be stripped out",
			filters: &parsers.MediaFilters{
				Videos: parsers.NestedFilters{
					Codecs: []string{"avc", "dvh"},
				},
			},
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
			manifest, err := filter.FilterContent(context.Background(), tt.filters)

			if err != nil && !tt.expectErr {
				t.Errorf("FilterContent(context.Background(), ) didnt expect an error to be returned, got: %v", err)
				return
			} else if err == nil && tt.expectErr {
				t.Error("FilterContent(context.Background(), ) expected an error, got nil")
				return
			}

			if g, e := manifest, tt.expectManifestContent; g != e {
				t.Errorf("FilterContent(context.Background(), ) wrong manifest returned)\ngot %v\nexpected: %v\ndiff: %v", g, e,
					cmp.Diff(g, e))
			}

		})
	}
}

func TestHLSFilter_FilterContent_CaptionsFilter(t *testing.T) {
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
				Captions: parsers.NestedFilters{
					Codecs: []string{"stpp", "wvtt"},
				},
			},
			manifestContent:       manifestWithAllCaptions,
			expectManifestContent: manifestWithNoCaptions,
		},
		{
			name: "when filter is supplied with wvtt, expect variants with wvtt to be stripped out",
			filters: &parsers.MediaFilters{
				Captions: parsers.NestedFilters{
					Codecs: []string{"wvtt"},
				},
			},
			manifestContent:       manifestWithAllCaptions,
			expectManifestContent: manifestFilterWithoutWVTT,
		},
		{
			name: "when filter is supplied with stpp, expect variants with wvtt to be stripped out",
			filters: &parsers.MediaFilters{
				Captions: parsers.NestedFilters{
					Codecs: []string{"stpp"},
				},
			},
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
			manifest, err := filter.FilterContent(context.Background(), tt.filters)

			if err != nil && !tt.expectErr {
				t.Errorf("FilterContent(context.Background(), ) didnt expect an error to be returned, got: %v", err)
				return
			} else if err == nil && tt.expectErr {
				t.Error("FilterContent(context.Background(), ) expected an error, got nil")
				return
			}

			if g, e := manifest, tt.expectManifestContent; g != e {
				t.Errorf("FilterContent(context.Background(), ) wrong manifest returned\ngot %v\nexpected: %v\ndiff: %v", g, e,
					cmp.Diff(g, e))
			}

		})
	}
}

func TestHLSFilter_FilterContent_MultiCodecFilter(t *testing.T) {
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
			name: "when filter is supplied with audio (ec-3 and mp4a) and video (hevc and dvh), expect variants with ec-3, mp4a, hevc, and/or dvh to be stripped out",
			filters: &parsers.MediaFilters{
				Videos: parsers.NestedFilters{
					Codecs: []string{"hvc", "dvh"},
				},
				Audios: parsers.NestedFilters{
					Codecs: []string{"ec-3", "mp4a"},
				},
			},
			manifestContent:       manifestWithAllCodecs,
			expectManifestContent: manifestFilterInAC3AndAVC,
		},
		{
			name: "when filter is supplied with audio (mp4a) and video (hevc and dvh), expect variants with mp4a, hevc, and/or dvh to be stripped out",
			filters: &parsers.MediaFilters{
				Videos: parsers.NestedFilters{
					Codecs: []string{"hvc", "dvh"},
				},
				Audios: parsers.NestedFilters{
					Codecs: []string{"mp4a"},
				},
			},
			manifestContent:       manifestWithAllCodecs,
			expectManifestContent: manifestFilterInAC3AndEC3AndAVC,
		},
		{
			name: "when filter is supplied with audio (ec-3 and mp4a) and captions (stpp), expect variants with ec-3, mp4a, and/or stpp to be stripped out",
			filters: &parsers.MediaFilters{
				Captions: parsers.NestedFilters{
					Codecs: []string{"stpp"},
				},
				Audios: parsers.NestedFilters{
					Codecs: []string{"ec-3", "mp4a"},
				},
			},
			manifestContent:       manifestWithAllCodecs,
			expectManifestContent: manifestFilterInAC3AndWVTT,
		},
		{
			name: "when filter is supplied with audio (ec-3 and mp4a), video (hevc and dvh), and captions (stpp), expect variants with ec-3, mp4a, hevc, dvh, and/or stpp to be stripped out",
			filters: &parsers.MediaFilters{
				Captions: parsers.NestedFilters{
					Codecs: []string{"stpp"},
				},
				Videos: parsers.NestedFilters{
					Codecs: []string{"hvc", "dvh"},
				},
				Audios: parsers.NestedFilters{
					Codecs: []string{"ec-3", "mp4a"},
				},
			},
			manifestContent:       manifestWithAllCodecs,
			expectManifestContent: manifestFilterInAC3AndAVCAndWVTT,
		},
		{
			name: "when filtering out all codecs except avc video, expect variants with ac-3, ec-3, mp4a, hevc, and/or dvh to be stripped out",
			filters: &parsers.MediaFilters{
				Captions: parsers.NestedFilters{
					Codecs: []string{"stpp", "wvtt"},
				},
				Videos: parsers.NestedFilters{
					Codecs: []string{"hvc", "dvh"},
				},
				Audios: parsers.NestedFilters{
					Codecs: []string{"ac-3", "ec-3", "mp4a"},
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
			manifest, err := filter.FilterContent(context.Background(), tt.filters)

			if err != nil && !tt.expectErr {
				t.Errorf("FilterContent(context.Background(), ) didnt expect an error to be returned, got: %v", err)
				return
			} else if err == nil && tt.expectErr {
				t.Error("FilterContent(context.Background(), ) expected an error, got nil")
				return
			}

			if g, e := manifest, tt.expectManifestContent; g != e {
				t.Errorf("FilterContent(context.Background(), ) wrong manifest returned)\ngot %v\nexpected: %v\ndiff: %v", g, e,
					cmp.Diff(g, e))
			}

		})
	}
}

func TestHLSFilter_FilterContent_MultiFilter(t *testing.T) {

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
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=500,AVERAGE-BANDWIDTH=500,CODECS="wvtt"
http://existing.base/uri/link_14.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=5300,AVERAGE-BANDWIDTH=5300
http://existing.base/uri/link_13.m3u8
`

	manifestFilter4000To6000BandwidthAndDVH := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=5900,AVERAGE-BANDWIDTH=5900,CODECS="ac-3,ec-3"
http://existing.base/uri/link_7b.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=500,AVERAGE-BANDWIDTH=500,CODECS="wvtt"
http://existing.base/uri/link_14.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=5300,AVERAGE-BANDWIDTH=5300
http://existing.base/uri/link_13.m3u8
`

	manifestFilter4000To6000BandwidthAndEC3AndAVC := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4200,AVERAGE-BANDWIDTH=4200,CODECS="avc1.77.30"
http://existing.base/uri/link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4500,AVERAGE-BANDWIDTH=4500,CODECS="ec-3,avc1.640029"
http://existing.base/uri/link_6.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=500,AVERAGE-BANDWIDTH=500,CODECS="wvtt"
http://existing.base/uri/link_14.m3u8
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
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=500,AVERAGE-BANDWIDTH=500,CODECS="wvtt"
http://existing.base/uri/link_14.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=5300,AVERAGE-BANDWIDTH=5300
http://existing.base/uri/link_13.m3u8
`

	manifestFilter4000To6000BandwidthAndNoAudio := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4200,AVERAGE-BANDWIDTH=4200,CODECS="avc1.77.30"
http://existing.base/uri/link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=500,AVERAGE-BANDWIDTH=500,CODECS="wvtt"
http://existing.base/uri/link_14.m3u8
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
			name: "when filtering out audio (ec-3) and setting bandwidth range 4000-6000, expect variants with ec-3, mp4a, and/or not in range to be stripped out",
			filters: &parsers.MediaFilters{
				Videos: parsers.NestedFilters{
					Bitrate: &parsers.Bitrate{
						Min: 4000,
						Max: 6000,
					},
				},
				Audios: parsers.NestedFilters{
					Codecs: []string{"ec-3"},
					Bitrate: &parsers.Bitrate{
						Min: 4000,
						Max: 6000,
					},
				},
			},
			manifestContent:       manifestWithAllCodecsAndBandwidths,
			expectManifestContent: manifestFilter4000To6000BandwidthAndAC3,
		},
		{
			name: "when filtering out video (avc and hevc) and filtering in bandwidth range 4000-6000, expect variants with avc, hevc, and/or not in range to be stripped out",
			filters: &parsers.MediaFilters{
				Videos: parsers.NestedFilters{
					Codecs: []string{"avc", "hvc"},
					Bitrate: &parsers.Bitrate{
						Min: 4000,
						Max: 6000,
					},
				},
				Audios: parsers.NestedFilters{
					Bitrate: &parsers.Bitrate{
						Min: 4000,
						Max: 6000,
					},
				},
			},
			manifestContent:       manifestWithAllCodecsAndBandwidths,
			expectManifestContent: manifestFilter4000To6000BandwidthAndDVH,
		},
		{
			name: "when filtering out audio (ac-3, mp4a) and video (hevc and dvh) and filtering in bandwidth range 4000-6000, expect variants with ac-3, mp4a, hevc, dvh, and/or not in range to be stripped out",
			filters: &parsers.MediaFilters{
				Videos: parsers.NestedFilters{
					Codecs: []string{"hvc", "dvh"},
					Bitrate: &parsers.Bitrate{
						Min: 4000,
						Max: 6000,
					},
				},
				Audios: parsers.NestedFilters{
					Codecs: []string{"ac-3", "mp4a"},
					Bitrate: &parsers.Bitrate{
						Min: 4000,
						Max: 6000,
					},
				},
			},
			manifestContent:       manifestWithAllCodecsAndBandwidths,
			expectManifestContent: manifestFilter4000To6000BandwidthAndEC3AndAVC,
		},
		{
			name: "when filtering out captions (stpp) and filtering in bandwidth range 4000-6000, expect variants with stpp and/or not in range to be stripped out",
			filters: &parsers.MediaFilters{
				Captions: parsers.NestedFilters{
					Codecs: []string{"stpp"},
				},
				Videos: parsers.NestedFilters{
					Bitrate: &parsers.Bitrate{
						Min: 4000,
						Max: 6000,
					},
				},
				Audios: parsers.NestedFilters{
					Bitrate: &parsers.Bitrate{
						Min: 4000,
						Max: 6000,
					},
				},
			},
			manifestContent:       manifestWithAllCodecsAndBandwidths,
			expectManifestContent: manifestFilter4000To6000BandwidthAndWVTT,
		},
		{
			name: "when filtering out audio and filtering in bandwidth range 4000-6000, expect variants with ac-3, ec-3, mp4a, and/or not in range to be stripped out",
			filters: &parsers.MediaFilters{
				Videos: parsers.NestedFilters{
					Bitrate: &parsers.Bitrate{
						Min: 4000,
						Max: 6000,
					},
				},
				Audios: parsers.NestedFilters{
					Codecs: []string{"ac-3", "ec-3", "mp4a"},
					Bitrate: &parsers.Bitrate{
						Min: 4000,
						Max: 6000,
					},
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
			manifest, err := filter.FilterContent(context.Background(), tt.filters)

			if err != nil && !tt.expectErr {
				t.Errorf("FilterContent(context.Background(), ) didnt expect an error to be returned, got: %v", err)
				return
			} else if err == nil && tt.expectErr {
				t.Error("FilterContent(context.Background(), ) expected an error, got nil")
				return
			}

			if g, e := manifest, tt.expectManifestContent; g != e {
				t.Errorf("FilterContent(context.Background(), ) wrong manifest returned)\ngot %v\nexpected: %v\ndiff: %v", g, e,
					cmp.Diff(g, e))
			}

		})
	}
}

func TestHLSFilter_FilterContent_NormalizeVariant(t *testing.T) {

	manifestWithRelativeOnly := `#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="AU",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG",URI="audio.m3u8"
#EXT-X-MEDIA:TYPE=VIDEO,GROUP-ID="VID",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG",URI="video.m3u8"
#EXT-X-MEDIA:TYPE=CLOSED-CAPTIONS,GROUP-ID="CC",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG"
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,AUDIO="AU",VIDEO="VID",CLOSED-CAPTIONS="CC"
link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CLOSED-CAPTIONS="CC"
link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,AUDIO="AU2",CLOSED-CAPTIONS="CC"
../../link_3.m3u8
`

	manifestWithAbsoluteOnly := `#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="AU",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG",URI="http://existing.base/uri/nested/folders/audio.m3u8"
#EXT-X-MEDIA:TYPE=VIDEO,GROUP-ID="VID",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG",URI="http://existing.base/uri/nested/folders/video.m3u8"
#EXT-X-MEDIA:TYPE=CLOSED-CAPTIONS,GROUP-ID="CC",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG"
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,AUDIO="AU",VIDEO="VID",CLOSED-CAPTIONS="CC"
http://existing.base/uri/nested/folders/link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CLOSED-CAPTIONS="CC"
http://existing.base/uri/nested/folders/link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,AUDIO="AU2",CLOSED-CAPTIONS="CC"
http://existing.base/uri/link_3.m3u8
`

	manifestWithRelativeAndAbsolute := `#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="AU",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG",URI="http://existing.base/uri/nested/folders/audio.m3u8"
#EXT-X-MEDIA:TYPE=VIDEO,GROUP-ID="VID",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG",URI="http://existing.base/uri/nested/folders/video.m3u8"
#EXT-X-MEDIA:TYPE=CLOSED-CAPTIONS,GROUP-ID="CC",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG"
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,AUDIO="AU",VIDEO="VID",CLOSED-CAPTIONS="CC"
link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CLOSED-CAPTIONS="CC"
http://existing.base/uri/nested/folders/link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,AUDIO="AU2",CLOSED-CAPTIONS="CC"
../../link_3.m3u8
`

	manifestWithDifferentAbsolute := `#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="AU",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG",URI="audio.m3u8"
#EXT-X-MEDIA:TYPE=VIDEO,GROUP-ID="VID",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG",URI="http://different.base/uri/video.m3u8"
#EXT-X-MEDIA:TYPE=CLOSED-CAPTIONS,GROUP-ID="CC",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG"
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,AUDIO="AU",VIDEO="VID",CLOSED-CAPTIONS="CC"
link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CLOSED-CAPTIONS="CC"
http://different.base/uri/link_2.m3u8
`

	manifestWithDifferentAbsoluteExpected := `#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="AU",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG",URI="http://existing.base/uri/nested/folders/audio.m3u8"
#EXT-X-MEDIA:TYPE=VIDEO,GROUP-ID="VID",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG",URI="http://different.base/uri/video.m3u8"
#EXT-X-MEDIA:TYPE=CLOSED-CAPTIONS,GROUP-ID="CC",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG"
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,AUDIO="AU",VIDEO="VID",CLOSED-CAPTIONS="CC"
http://existing.base/uri/nested/folders/link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CLOSED-CAPTIONS="CC"
http://different.base/uri/link_2.m3u8
`

	manifestWithIllegalAlternativeURLs := `#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=CLOSED-CAPTIONS,GROUP-ID="CC",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG"
#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="AU",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG",URI="http://exist\ing.base/uri/illegal.m3u8"
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,AUDIO="AU",VIDEO="VID",CLOSED-CAPTIONS="CC"
http://existing.base/uri/nested/folders/link_1.m3u8
`

	manifestWithIllegalVariantURLs := `#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=CLOSED-CAPTIONS,GROUP-ID="CC",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG"
#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="AU",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG",URI="\nillegal.m3u8"
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
			manifest, err := filter.FilterContent(context.Background(), tt.filters)

			if err != nil && !tt.expectErr {
				t.Errorf("FilterContent(context.Background(), ) didnt expect an error to be returned, got: %v", err)
				return
			} else if err == nil && tt.expectErr {
				t.Error("FilterContent(context.Background(), ) expected an error, got nil")
				return
			} else if err != nil && tt.expectErr {
				return
			}

			if g, e := manifest, tt.expectManifestContent; g != e {
				t.Errorf("FilterContent(context.Background(), ) wrong manifest returned\ngot %v\nexpected: %v\ndiff: %v", g, e,
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
			manifest, err := filter.FilterContent(context.Background(), tt.filters)
			if err != nil && !tt.expectErr {
				t.Errorf("FilterContent(context.Background(), ) didnt expect an error to be returned, got: %v", err)
				return
			} else if err == nil && tt.expectErr {
				t.Error("FilterContent(context.Background(), ) expected an error, got nil")
				return
			} else if err != nil && tt.expectErr {
				return
			}

			if g, e := manifest, tt.expectManifestContent; g != e {
				t.Errorf("FilterContent(context.Background(), ) wrong manifest returned\ngot %v\nexpected: %v\ndiff: %v", g, e,
					cmp.Diff(g, e))
			}
		})
	}
}

func TestHLSFilter_FilterContent_TrimFilter_MasterManifest(t *testing.T) {

	masterManifestWithAbsoluteURLs := `#EXTM3U
#EXT-X-VERSION:3
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
#EXT-X-VERSION:3
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
#EXT-X-VERSION:3
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

	manifestWithFilteredAdsAndBase64EncodedVariantURLS := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,CODECS="avc1.64001f,mp4a.40.2"
https://bakery.cbsi.video/t(10000,100000)/tags(ads)/aHR0cHM6Ly9leGlzdGluZy5iYXNlL3BhdGgvbGlua18xLm0zdTg.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4200,AVERAGE-BANDWIDTH=4200,CODECS="avc1.64001f,mp4a.40.2"
https://bakery.cbsi.video/t(10000,100000)/tags(ads)/aHR0cHM6Ly9leGlzdGluZy5iYXNlL3BhdGgvbGlua18yLm0zdTg.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CODECS="avc1.64001f,mp4a.40.2"
https://bakery.cbsi.video/t(10000,100000)/tags(ads)/aHR0cHM6Ly9leGlzdGluZy5iYXNlL3BhdGgvbGlua180Lm0zdTg.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4100,AVERAGE-BANDWIDTH=4100,CODECS="avc1.64001f,mp4a.40.2"
https://bakery.cbsi.video/t(10000,100000)/tags(ads)/aHR0cHM6Ly9leGlzdGluZy5iYXNlL3BhdGgvbGlua181Lm0zdTg.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4500,AVERAGE-BANDWIDTH=4500,CODECS="avc1.64001f,mp4a.40.2"
https://bakery.cbsi.video/t(10000,100000)/tags(ads)/aHR0cHM6Ly9leGlzdGluZy5iYXNlL3BhdGgvbGlua182Lm0zdTg.m3u8
`

	manifestWithBase64EncodedVariantURLSAndLocalHost := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,CODECS="avc1.64001f,mp4a.40.2"
http://localhost:8080/t(10000,100000)/aHR0cHM6Ly9leGlzdGluZy5iYXNlL3BhdGgvbGlua18xLm0zdTg.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4200,AVERAGE-BANDWIDTH=4200,CODECS="avc1.64001f,mp4a.40.2"
http://localhost:8080/t(10000,100000)/aHR0cHM6Ly9leGlzdGluZy5iYXNlL3BhdGgvbGlua18yLm0zdTg.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CODECS="avc1.64001f,mp4a.40.2"
http://localhost:8080/t(10000,100000)/aHR0cHM6Ly9leGlzdGluZy5iYXNlL3BhdGgvbGlua180Lm0zdTg.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4100,AVERAGE-BANDWIDTH=4100,CODECS="avc1.64001f,mp4a.40.2"
http://localhost:8080/t(10000,100000)/aHR0cHM6Ly9leGlzdGluZy5iYXNlL3BhdGgvbGlua181Lm0zdTg.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4500,AVERAGE-BANDWIDTH=4500,CODECS="avc1.64001f,mp4a.40.2"
http://localhost:8080/t(10000,100000)/aHR0cHM6Ly9leGlzdGluZy5iYXNlL3BhdGgvbGlua182Lm0zdTg.m3u8
`

	manifestWithFilteredAdsAndBase64EncodedVariantURLSAndLocalHost := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,CODECS="avc1.64001f,mp4a.40.2"
http://localhost:8080/t(10000,100000)/tags(ads)/aHR0cHM6Ly9leGlzdGluZy5iYXNlL3BhdGgvbGlua18xLm0zdTg.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4200,AVERAGE-BANDWIDTH=4200,CODECS="avc1.64001f,mp4a.40.2"
http://localhost:8080/t(10000,100000)/tags(ads)/aHR0cHM6Ly9leGlzdGluZy5iYXNlL3BhdGgvbGlua18yLm0zdTg.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CODECS="avc1.64001f,mp4a.40.2"
http://localhost:8080/t(10000,100000)/tags(ads)/aHR0cHM6Ly9leGlzdGluZy5iYXNlL3BhdGgvbGlua180Lm0zdTg.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4100,AVERAGE-BANDWIDTH=4100,CODECS="avc1.64001f,mp4a.40.2"
http://localhost:8080/t(10000,100000)/tags(ads)/aHR0cHM6Ly9leGlzdGluZy5iYXNlL3BhdGgvbGlua181Lm0zdTg.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4500,AVERAGE-BANDWIDTH=4500,CODECS="avc1.64001f,mp4a.40.2"
http://localhost:8080/t(10000,100000)/tags(ads)/aHR0cHM6Ly9leGlzdGluZy5iYXNlL3BhdGgvbGlua182Lm0zdTg.m3u8
`

	masterManifestWithMediaTag := `#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:URI="https://devcarrier-hls.amlg.io/captions/734adfe1-0223-487a-960b-863f6f8273dc/en.m3u8",TYPE=SUBTITLES,GROUP-ID="subs",LANGUAGE="eng",NAME="English",DEFAULT=YES,AUTOSELECT=YES,FORCED=NO
#EXT-X-STREAM-INF:BANDWIDTH=391534,AVERAGE-BANDWIDTH=380600,RESOLUTION=400x224,FRAME-RATE=29.97,CODECS="avc1.64000d,mp4a.40.2",SUBTITLES="subs"
https://cbss64eb-cbss64eb-ms-dev.global.ssl.fastly.net/cbssc0a7/master/cbssc0a7_1.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=677468,AVERAGE-BANDWIDTH=655600,RESOLUTION=512x288,FRAME-RATE=29.97,CODECS="avc1.640015,mp4a.40.2",SUBTITLES="subs"
https://cbss64eb-cbss64eb-ms-dev.global.ssl.fastly.net/cbssc0a7/master/cbssc0a7_2.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=1055788,AVERAGE-BANDWIDTH=1020800,RESOLUTION=640x360,FRAME-RATE=29.97,CODECS="avc1.64001e,mp4a.40.2",SUBTITLES="subs"
https://cbss64eb-cbss64eb-ms-dev.global.ssl.fastly.net/cbssc0a7/master/cbssc0a7_3.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=1513283,AVERAGE-BANDWIDTH=1460800,RESOLUTION=768x432,FRAME-RATE=29.97,CODECS="avc1.64001e,mp4a.40.2",SUBTITLES="subs"
https://cbss64eb-cbss64eb-ms-dev.global.ssl.fastly.net/cbssc0a7/master/cbssc0a7_4.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=2199524,AVERAGE-BANDWIDTH=2120800,RESOLUTION=960x540,FRAME-RATE=29.97,CODECS="avc1.64001f,mp4a.40.2",SUBTITLES="subs"
https://cbss64eb-cbss64eb-ms-dev.global.ssl.fastly.net/cbssc0a7/master/cbssc0a7_5.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=4143876,AVERAGE-BANDWIDTH=3990800,RESOLUTION=1280x720,FRAME-RATE=29.97,CODECS="avc1.64001f,mp4a.40.2",SUBTITLES="subs"
https://cbss64eb-cbss64eb-ms-dev.global.ssl.fastly.net/cbssc0a7/master/cbssc0a7_6.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=5859480,AVERAGE-BANDWIDTH=5640800,RESOLUTION=1920x1080,FRAME-RATE=29.97,CODECS="avc1.640028,mp4a.40.2",SUBTITLES="subs"
https://cbss64eb-cbss64eb-ms-dev.global.ssl.fastly.net/cbssc0a7/master/cbssc0a7_7.m3u8
`

	masterManifestFilteredWithMediaTagAndBase64EncodedMediaURI := `#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=SUBTITLES,GROUP-ID="subs",NAME="English",DEFAULT=YES,AUTOSELECT=YES,LANGUAGE="eng",FORCED="NO",URI="https://bakery.cbsi.video/t(10000,100000)/aHR0cHM6Ly9kZXZjYXJyaWVyLWhscy5hbWxnLmlvL2NhcHRpb25zLzczNGFkZmUxLTAyMjMtNDg3YS05NjBiLTg2M2Y2ZjgyNzNkYy9lbi5tM3U4.m3u8"
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=391534,AVERAGE-BANDWIDTH=380600,CODECS="avc1.64000d,mp4a.40.2",RESOLUTION=400x224,SUBTITLES="subs",FRAME-RATE=29.970
https://bakery.cbsi.video/t(10000,100000)/aHR0cHM6Ly9jYnNzNjRlYi1jYnNzNjRlYi1tcy1kZXYuZ2xvYmFsLnNzbC5mYXN0bHkubmV0L2Nic3NjMGE3L21hc3Rlci9jYnNzYzBhN18xLm0zdTg.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=677468,AVERAGE-BANDWIDTH=655600,CODECS="avc1.640015,mp4a.40.2",RESOLUTION=512x288,SUBTITLES="subs",FRAME-RATE=29.970
https://bakery.cbsi.video/t(10000,100000)/aHR0cHM6Ly9jYnNzNjRlYi1jYnNzNjRlYi1tcy1kZXYuZ2xvYmFsLnNzbC5mYXN0bHkubmV0L2Nic3NjMGE3L21hc3Rlci9jYnNzYzBhN18yLm0zdTg.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1055788,AVERAGE-BANDWIDTH=1020800,CODECS="avc1.64001e,mp4a.40.2",RESOLUTION=640x360,SUBTITLES="subs",FRAME-RATE=29.970
https://bakery.cbsi.video/t(10000,100000)/aHR0cHM6Ly9jYnNzNjRlYi1jYnNzNjRlYi1tcy1kZXYuZ2xvYmFsLnNzbC5mYXN0bHkubmV0L2Nic3NjMGE3L21hc3Rlci9jYnNzYzBhN18zLm0zdTg.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1513283,AVERAGE-BANDWIDTH=1460800,CODECS="avc1.64001e,mp4a.40.2",RESOLUTION=768x432,SUBTITLES="subs",FRAME-RATE=29.970
https://bakery.cbsi.video/t(10000,100000)/aHR0cHM6Ly9jYnNzNjRlYi1jYnNzNjRlYi1tcy1kZXYuZ2xvYmFsLnNzbC5mYXN0bHkubmV0L2Nic3NjMGE3L21hc3Rlci9jYnNzYzBhN180Lm0zdTg.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=2199524,AVERAGE-BANDWIDTH=2120800,CODECS="avc1.64001f,mp4a.40.2",RESOLUTION=960x540,SUBTITLES="subs",FRAME-RATE=29.970
https://bakery.cbsi.video/t(10000,100000)/aHR0cHM6Ly9jYnNzNjRlYi1jYnNzNjRlYi1tcy1kZXYuZ2xvYmFsLnNzbC5mYXN0bHkubmV0L2Nic3NjMGE3L21hc3Rlci9jYnNzYzBhN181Lm0zdTg.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4143876,AVERAGE-BANDWIDTH=3990800,CODECS="avc1.64001f,mp4a.40.2",RESOLUTION=1280x720,SUBTITLES="subs",FRAME-RATE=29.970
https://bakery.cbsi.video/t(10000,100000)/aHR0cHM6Ly9jYnNzNjRlYi1jYnNzNjRlYi1tcy1kZXYuZ2xvYmFsLnNzbC5mYXN0bHkubmV0L2Nic3NjMGE3L21hc3Rlci9jYnNzYzBhN182Lm0zdTg.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=5859480,AVERAGE-BANDWIDTH=5640800,CODECS="avc1.640028,mp4a.40.2",RESOLUTION=1920x1080,SUBTITLES="subs",FRAME-RATE=29.970
https://bakery.cbsi.video/t(10000,100000)/aHR0cHM6Ly9jYnNzNjRlYi1jYnNzNjRlYi1tcy1kZXYuZ2xvYmFsLnNzbC5mYXN0bHkubmV0L2Nic3NjMGE3L21hc3Rlci9jYnNzYzBhN183Lm0zdTg.m3u8
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
		config                config.Config
		expectErr             bool
	}{
		{
			name: "when trim filter is given and master has absolute urls, variant level manifest will point to " +
				"bakery with trim filter and base64 encoding string in the manifest",
			filters: &parsers.MediaFilters{
				Trim: trim,
			},
			manifestContent:       masterManifestWithAbsoluteURLs,
			expectManifestContent: manifestWithBase64EncodedVariantURLS,
			config:                config.Config{Hostname: "bakery.cbsi.video"},
		},
		{
			name: "when trim filter is given and master has relative urls, variant level manifest will point to " +
				"bakery with trim filter and base64 encoding string in the manifest",
			filters: &parsers.MediaFilters{
				Trim: trim,
			},
			manifestContent:       masterManifestWithRelativeURLs,
			expectManifestContent: manifestWithBase64EncodedVariantURLS,
			config:                config.Config{Hostname: "bakery.cbsi.video"},
		},
		{
			name: "when trim and ads filter is given and master has absolute urls, variant level manifest will point to " +
				"bakery with trim and ads filter and base64 encoding string in the manifest",
			filters: &parsers.MediaFilters{
				Trim: trim,
				Tags: &parsers.Tags{
					Ads: true,
				},
			},
			manifestContent:       masterManifestWithAbsoluteURLs,
			expectManifestContent: manifestWithFilteredAdsAndBase64EncodedVariantURLS,
			config:                config.Config{Hostname: "bakery.cbsi.video"},
		},
		{
			name: "when trim and ads filter is given and master has absolute urls, variant level manifest will point to " +
				"bakery with trim and ads filter and base64 encoding string in the manifest",
			filters: &parsers.MediaFilters{
				Trim: trim,
				Tags: &parsers.Tags{
					Ads: true,
				},
			},
			manifestContent:       masterManifestWithAbsoluteURLs,
			expectManifestContent: manifestWithFilteredAdsAndBase64EncodedVariantURLS,
			config:                config.Config{Hostname: "bakery.cbsi.video"},
		},
		{
			name: "when bitrate and trim filter are given, variant level manifest will point to" +
				"bakery with only included bitrates, the trim filter, and base64 encoding string in the manifest ",
			filters: &parsers.MediaFilters{
				Videos: parsers.NestedFilters{
					Bitrate: &parsers.Bitrate{
						Min: 4000,
						Max: 6000,
					},
				},
				Audios: parsers.NestedFilters{
					Bitrate: &parsers.Bitrate{
						Min: 4000,
						Max: 6000,
					},
				},
				Trim: trim,
			},
			manifestContent:       masterManifestWithRelativeURLs,
			expectManifestContent: manifestWithFilteredBitrateAndBase64EncodedVariantURLS,
			config:                config.Config{Hostname: "bakery.cbsi.video"},
		},
		{
			name: "when trim and ads filter, variant level manifest will point to localhost and port" +
				"with trim and ads filter and base64 encoding string in the manifest",
			filters: &parsers.MediaFilters{
				Trim: trim,
				Tags: &parsers.Tags{
					Ads: true,
				},
			},
			manifestContent:       masterManifestWithAbsoluteURLs,
			expectManifestContent: manifestWithFilteredAdsAndBase64EncodedVariantURLSAndLocalHost,
			config:                config.Config{Hostname: "localhost", Listen: ":8080"},
		},
		{
			name: "when trim is set, variant level manifest will point to localhost and port" +
				"with trim filter and base64 encoding string in the manifest",
			filters: &parsers.MediaFilters{
				Trim: trim,
			},
			manifestContent:       masterManifestWithAbsoluteURLs,
			expectManifestContent: manifestWithBase64EncodedVariantURLSAndLocalHost,
			config:                config.Config{Hostname: "localhost", Listen: ":8080"},
		},
		{
			name:                  "when no filter is given, variant level manifest will hold absolute urls only",
			filters:               &parsers.MediaFilters{},
			manifestContent:       masterManifestWithRelativeURLs,
			expectManifestContent: masterManifestWithAbsoluteURLs,
			config:                config.Config{Hostname: "bakery.cbsi.video"},
		},
		{
			name: "when media tag is present, filter and replace media url with base64 encoded string",
			filters: &parsers.MediaFilters{
				Trim: trim,
			},
			manifestContent:       masterManifestWithMediaTag,
			expectManifestContent: masterManifestFilteredWithMediaTagAndBase64EncodedMediaURI,
			config:                config.Config{Hostname: "bakery.cbsi.video"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			filter := NewHLSFilter("https://existing.base/path/master.m3u8", tt.manifestContent, tt.config)
			manifest, err := filter.FilterContent(context.Background(), tt.filters)

			if err != nil && !tt.expectErr {
				t.Errorf("FilterContent(context.Background(), ) didnt expect an error to be returned, got: %v", err)
				return
			} else if err == nil && tt.expectErr {
				t.Error("FilterContent(context.Background(), ) expected an error, got nil")
				return
			}

			if g, e := manifest, tt.expectManifestContent; g != e {
				t.Errorf("FilterContent(context.Background(), ) wrong manifest returned)\ngot %v\nexpected: %v\ndiff: %v", g, e,
					cmp.Diff(g, e))
			}

		})
	}
}

func TestHLSFilter_FilterContent_TrimFilter_VariantManifest(t *testing.T) {

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

	variantManifestInMillisecondsWithAbsoluteURLs := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-MEDIA-SEQUENCE:10
#EXT-X-TARGETDURATION:6
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:51:48.256Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202743_1_00019.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:51:54.256Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202748_1_00020.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:00.256Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202754_1_00021.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:06.256Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202801_1_00022.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:12.256Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202806_1_00023.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:18.256Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202813_1_00024.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:24.256Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202818_1_00025.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:30.256Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202824_1_00026.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:36.256Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202818_1_00027.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:42.256Z
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

	variantManifestTrimmedInMilliseconds := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-TARGETDURATION:6
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:51:54.256Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202748_1_00020.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:00.256Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202754_1_00021.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:06.256Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202801_1_00022.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:12.256Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202806_1_00023.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:18.256Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202813_1_00024.ts
#EXT-X-ENDLIST
`

	variantManifestTrimmedUnalignedTrimFilters := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-TARGETDURATION:6
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
#EXT-X-ENDLIST
`

	emptyVariantManifest := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-TARGETDURATION:0
#EXT-X-ENDLIST
`

	trim := &parsers.Trim{
		Start: 1583887920, //2020-03-11T00:52:00
		End:   1583887944, //2020-03-11T00:52:24
	}

	unalignedTrim := &parsers.Trim{
		Start: 1583887919, //2020-03-11T00:51:59
		End:   1583887944, //2020-03-11T00:52:24
	}

	tests := []struct {
		name                  string
		filters               *parsers.MediaFilters
		manifestContent       string
		expectManifestContent string
		expectAge             string
		expectErr             bool
	}{
		{
			name: "when trim filter is given, variant level manifest have absolute url with base64" +
				"encoding string in the manifest",
			filters:               &parsers.MediaFilters{Trim: trim},
			manifestContent:       variantManifestWithRelativeURLs,
			expectManifestContent: variantManifestTrimmed,
			expectAge:             "3",
		},
		{
			name: "when trim filter is given, variant level manifest is in milliseconds have absolute url with base64" +
				"encoding string in the manifest",
			filters:               &parsers.MediaFilters{Trim: trim},
			manifestContent:       variantManifestInMillisecondsWithAbsoluteURLs,
			expectManifestContent: variantManifestTrimmedInMilliseconds,
			expectAge:             "3",
		},
		{
			name:                  "when trim filter start range is not aligned to a PDT, all segments with any content in the range are included",
			filters:               &parsers.MediaFilters{Trim: unalignedTrim},
			manifestContent:       variantManifestWithAbsoluteURLs,
			expectManifestContent: variantManifestTrimmedUnalignedTrimFilters,
			expectAge:             "3",
		},
		{
			name:                  "when no filter is given, variant level manifest will hold absolute urls only",
			filters:               &parsers.MediaFilters{Trim: trim},
			manifestContent:       variantManifestWithAbsoluteURLs,
			expectManifestContent: variantManifestTrimmed,
			expectAge:             "3",
		},
		{
			name:                  "when no pdt present for segment, empty manifest is returned",
			filters:               &parsers.MediaFilters{Trim: trim},
			manifestContent:       variantManifestWithNoPDT,
			expectManifestContent: emptyVariantManifest,
			expectAge:             "0",
			expectErr:             true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			filter := NewHLSFilter("https://existing.base/path/master.m3u8", tt.manifestContent, config.Config{Hostname: "bakery.cbsi.video"})
			manifest, err := filter.FilterContent(context.Background(), tt.filters)

			if err != nil && !tt.expectErr {
				t.Errorf("FilterContent(context.Background(), ) didnt expect an error to be returned, got: %v", err)
				return
			} else if err == nil && tt.expectErr {
				t.Error("FilterContent(context.Background(), ) expected an error, got nil")
				return
			}

			if g, e := manifest, tt.expectManifestContent; g != e {
				t.Errorf("FilterContent(context.Background(), ) wrong manifest returned)\ngot %v\nexpected: %v\ndiff: %v", g, e,
					cmp.Diff(g, e))
			}

			if g := filter.GetMaxAge(); g != tt.expectAge {
				t.Errorf("Wrong max age returned\ngot %v\nexpected: %v\ndiff: %v", g, tt.expectAge,
					cmp.Diff(g, tt.expectAge))
			}

		})
	}
}

func TestHLSFilter_FilterContent_TrimFilter_VariantManifest_AdSuppression(t *testing.T) {

	variantManifestWithAds := `#EXTM3U
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
chan_1/chan_1_20200311T202801_1_00021.ts
#EXT-OATCLS-SCTE35:/DAuAAAAAAAAAP/wBQb/Ldjb7wAYAhZDVUVJCiuBsH/DAADN/lIMAgEANAAADbYGAw==
#EXT-X-ASSET:CAID=0x0100
#EXT-X-CUE-OUT:30
#EXTINF:6.000,
chan_1/chan_1_20200311T202806_1_00022.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:12Z
#EXT-X-CUE-OUT-CONT:CAID=0x0100,ElapsedTime=6.00,Duration=150,SCTE35=/DAuAAAAAAAAAP/wBQb/Ldjb7wAYAhZDVUVJCiuBsH/DAADN/lIMAgEANAAADbYGAw==
#EXTINF:6.000,
chan_1/chan_1_20200311T202806_1_00023.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:18Z
#EXT-X-CUE-OUT-CONT:CAID=0x0100,ElapsedTime=12.00,Duration=150,SCTE35=/DAuAAAAAAAAAP/wBQb/Ldjb7wAYAhZDVUVJCiuBsH/DAADN/lIMAgEANAAADbYGAw==
#EXTINF:6.000,
chan_1/chan_1_20200311T202813_1_00024.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:24Z
#EXT-X-CUE-OUT-CONT:CAID=0x0100,ElapsedTime=18.00,Duration=150,SCTE35=/DAuAAAAAAAAAP/wBQb/Ldjb7wAYAhZDVUVJCiuBsH/DAADN/lIMAgEANAAADbYGAw==
#EXTINF:6.000,
chan_1/chan_1_20200311T202818_1_00025.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:30Z
#EXT-X-CUE-OUT-CONT:CAID=0x0100,ElapsedTime=24.00,Duration=150,SCTE35=/DAuAAAAAAAAAP/wBQb/Ldjb7wAYAhZDVUVJCiuBsH/DAADN/lIMAgEANAAADbYGAw==
#EXTINF:6.000,
chan_1/chan_1_20200311T202824_1_00026.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:36Z
#EXT-X-CUE-OUT-CONT:CAID=0x0100,ElapsedTime=30.00,Duration=150,SCTE35=/DAuAAAAAAAAAP/wBQb/Ldjb7wAYAhZDVUVJCiuBsH/DAADN/lIMAgEANAAADbYGAw==
#EXTINF:6.000,
chan_1/chan_1_20200311T202818_1_00027.ts
#EXT-X-CUE-IN
#EXTINF:6.000,
chan_1/chan_1_20200311T202824_1_00028.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:48Z
#EXTINF:6.000,
chan_1/chan_1_20200311T202818_1_00029.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:56Z
#EXTINF:6.000,
chan_1/chan_1_20200311T202824_1_00030.ts
`
	variantManifestInMillisecondsWithAds := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-MEDIA-SEQUENCE:10
#EXT-X-TARGETDURATION:6
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:51:48.256Z
#EXTINF:6.000,
chan_1/chan_1_20200311T202743_1_00019.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:51:54.256Z
#EXTINF:6.000,
chan_1/chan_1_20200311T202748_1_00020.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:00.256Z
#EXTINF:6.000,
chan_1/chan_1_20200311T202801_1_00021.ts
#EXT-OATCLS-SCTE35:/DAuAAAAAAAAAP/wBQb/Ldjb7wAYAhZDVUVJCiuBsH/DAADN/lIMAgEANAAADbYGAw==
#EXT-X-ASSET:CAID=0x0100
#EXT-X-CUE-OUT:30
#EXTINF:6.000,
chan_1/chan_1_20200311T202806_1_00022.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:12.256Z
#EXT-X-CUE-OUT-CONT:CAID=0x0100,ElapsedTime=6.00,Duration=150,SCTE35=/DAuAAAAAAAAAP/wBQb/Ldjb7wAYAhZDVUVJCiuBsH/DAADN/lIMAgEANAAADbYGAw==
#EXTINF:6.000,
chan_1/chan_1_20200311T202806_1_00023.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:18.256Z
#EXT-X-CUE-OUT-CONT:CAID=0x0100,ElapsedTime=12.00,Duration=150,SCTE35=/DAuAAAAAAAAAP/wBQb/Ldjb7wAYAhZDVUVJCiuBsH/DAADN/lIMAgEANAAADbYGAw==
#EXTINF:6.000,
chan_1/chan_1_20200311T202813_1_00024.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:24.256Z
#EXT-X-CUE-OUT-CONT:CAID=0x0100,ElapsedTime=18.00,Duration=150,SCTE35=/DAuAAAAAAAAAP/wBQb/Ldjb7wAYAhZDVUVJCiuBsH/DAADN/lIMAgEANAAADbYGAw==
#EXTINF:6.000,
chan_1/chan_1_20200311T202818_1_00025.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:30.256Z
#EXT-X-CUE-OUT-CONT:CAID=0x0100,ElapsedTime=24.00,Duration=150,SCTE35=/DAuAAAAAAAAAP/wBQb/Ldjb7wAYAhZDVUVJCiuBsH/DAADN/lIMAgEANAAADbYGAw==
#EXTINF:6.000,
chan_1/chan_1_20200311T202824_1_00026.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:36.256Z
#EXT-X-CUE-OUT-CONT:CAID=0x0100,ElapsedTime=30.00,Duration=150,SCTE35=/DAuAAAAAAAAAP/wBQb/Ldjb7wAYAhZDVUVJCiuBsH/DAADN/lIMAgEANAAADbYGAw==
#EXTINF:6.000,
chan_1/chan_1_20200311T202818_1_00027.ts
#EXT-X-CUE-IN
#EXTINF:6.000,
chan_1/chan_1_20200311T202824_1_00028.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:48.256Z
#EXTINF:6.000,
chan_1/chan_1_20200311T202818_1_00029.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:56.256Z
#EXTINF:6.000,
chan_1/chan_1_20200311T202824_1_00030.ts
`
	variantManifestInMillisecondsWithNoAds := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-TARGETDURATION:6
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:51:54.256Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202748_1_00020.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:00.256Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202801_1_00021.ts
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202806_1_00022.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:12.256Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202806_1_00023.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:18.256Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202813_1_00024.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:24.256Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202818_1_00025.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:30.256Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202824_1_00026.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:36.256Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202818_1_00027.ts
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202824_1_00028.ts
#EXT-X-ENDLIST
`

	variantManifestTrimWithAds := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-TARGETDURATION:6
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:00Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202801_1_00021.ts
#EXT-OATCLS-SCTE35:/DAuAAAAAAAAAP/wBQb/Ldjb7wAYAhZDVUVJCiuBsH/DAADN/lIMAgEANAAADbYGAw==
#EXT-X-ASSET:CAID=0x0100
#EXT-X-CUE-OUT:30
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202806_1_00022.ts
#EXT-X-CUE-OUT-CONT:CAID=0x0100,ElapsedTime=6,Duration=150,SCTE35=/DAuAAAAAAAAAP/wBQb/Ldjb7wAYAhZDVUVJCiuBsH/DAADN/lIMAgEANAAADbYGAw==
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:12Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202806_1_00023.ts
#EXT-X-CUE-OUT-CONT:CAID=0x0100,ElapsedTime=12,Duration=150,SCTE35=/DAuAAAAAAAAAP/wBQb/Ldjb7wAYAhZDVUVJCiuBsH/DAADN/lIMAgEANAAADbYGAw==
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:18Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202813_1_00024.ts
#EXT-X-CUE-OUT-CONT:CAID=0x0100,ElapsedTime=18,Duration=150,SCTE35=/DAuAAAAAAAAAP/wBQb/Ldjb7wAYAhZDVUVJCiuBsH/DAADN/lIMAgEANAAADbYGAw==
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:24Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202818_1_00025.ts
#EXT-X-CUE-OUT-CONT:CAID=0x0100,ElapsedTime=24,Duration=150,SCTE35=/DAuAAAAAAAAAP/wBQb/Ldjb7wAYAhZDVUVJCiuBsH/DAADN/lIMAgEANAAADbYGAw==
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:30Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202824_1_00026.ts
#EXT-X-CUE-OUT-CONT:CAID=0x0100,ElapsedTime=30,Duration=150,SCTE35=/DAuAAAAAAAAAP/wBQb/Ldjb7wAYAhZDVUVJCiuBsH/DAADN/lIMAgEANAAADbYGAw==
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:36Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202818_1_00027.ts
#EXT-X-CUE-IN
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202824_1_00028.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:48Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202818_1_00029.ts
#EXT-X-ENDLIST
`

	variantManifestTrimWithNoAds := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-TARGETDURATION:6
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:00Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202801_1_00021.ts
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202806_1_00022.ts
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
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202824_1_00028.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:48Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202818_1_00029.ts
#EXT-X-ENDLIST
`

	variantManifestTrimWithNoAdsWithUnalignedPDTsBeforeAdStart := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-TARGETDURATION:6
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:00Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202801_1_00021.ts
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202806_1_00022.ts
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
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202824_1_00028.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:48Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202818_1_00029.ts
#EXT-X-ENDLIST
`

	variantManifestTrimWithNoAdsWithUnalignedPDTsAfterAdStart := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-TARGETDURATION:6
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
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202824_1_00028.ts
#EXT-X-PROGRAM-DATE-TIME:2020-03-11T00:52:48Z
#EXTINF:6.000,
https://existing.base/path/chan_1/chan_1_20200311T202818_1_00029.ts
#EXT-X-ENDLIST
`

	trim := &parsers.Trim{
		Start: 1583887920, //2020-03-11T00:52:00
		End:   1583887968, //2020-03-11T00:52:48
	}

	tests := []struct {
		name                  string
		filters               *parsers.MediaFilters
		manifestContent       string
		expectManifestContent string
		expectAge             string
		expectErr             bool
	}{
		{
			name: "when trim filter is given and ads tag are not enabled, ads are decorated in the mainfest",
			filters: &parsers.MediaFilters{
				Trim: trim,
				Tags: &parsers.Tags{
					Ads: false,
				},
			},
			manifestContent:       variantManifestWithAds,
			expectManifestContent: variantManifestTrimWithAds,
			expectAge:             "3",
		},
		{
			name: "when trim filter is given and ads tag is enabled, ads are suppressed from mainfest",
			filters: &parsers.MediaFilters{
				Trim: trim,
				Tags: &parsers.Tags{
					Ads: true,
				},
			},
			manifestContent:       variantManifestWithAds,
			expectManifestContent: variantManifestTrimWithNoAds,
			expectAge:             "3",
		},
		{
			name: "when PDT is in millisecond, trim filter is given and ads tag is enabled, ads are suppressed from mainfest",
			filters: &parsers.MediaFilters{
				Trim: trim,
				Tags: &parsers.Tags{
					Ads: true,
				},
			},
			manifestContent:       variantManifestInMillisecondsWithAds,
			expectManifestContent: variantManifestInMillisecondsWithNoAds,
			expectAge:             "3",
		},
		{
			name: "when trim filter is given, ads tag is enabled and PDTs are unaligned before Ad start, all content that is in range is included",
			filters: &parsers.MediaFilters{
				Trim: &parsers.Trim{
					Start: 1583887921, //2020-03-11T00:52:01
					End:   1583887968, //2020-03-11T00:52:48
				},
				Tags: &parsers.Tags{
					Ads: true,
				},
			},
			manifestContent:       variantManifestWithAds,
			expectManifestContent: variantManifestTrimWithNoAdsWithUnalignedPDTsBeforeAdStart,
			expectAge:             "3",
		},
		{
			name: "when trim filter is given, ads tag is enabled and PDTs are unaligned after Ad start, all content that is in range is included",
			filters: &parsers.MediaFilters{
				Trim: &parsers.Trim{
					Start: 1583887927, //2020-03-11T00:52:07
					End:   1583887968, //2020-03-11T00:52:48
				},
				Tags: &parsers.Tags{
					Ads: true,
				},
			},
			manifestContent:       variantManifestWithAds,
			expectManifestContent: variantManifestTrimWithNoAdsWithUnalignedPDTsAfterAdStart,
			expectAge:             "3",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			filter := NewHLSFilter("https://existing.base/path/master.m3u8", tt.manifestContent, config.Config{Hostname: "bakery.cbsi.video"})
			manifest, err := filter.FilterContent(context.Background(), tt.filters)

			if err != nil && !tt.expectErr {
				t.Errorf("FilterContent(context.Background(), ) didnt expect an error to be returned, got: %v", err)
				return
			} else if err == nil && tt.expectErr {
				t.Error("FilterContent(context.Background(), ) expected an error, got nil")
				return
			}

			if g, e := manifest, tt.expectManifestContent; g != e {
				t.Errorf("FilterContent(context.Background(), ) wrong manifest returned)\ngot %v\nexpected: %v\ndiff: %v", g, e,
					cmp.Diff(g, e))
			}

			if g := filter.GetMaxAge(); g != tt.expectAge {
				t.Errorf("Wrong max age returned\ngot %v\nexpected: %v\ndiff: %v", g, tt.expectAge,
					cmp.Diff(g, tt.expectAge))
			}

		})
	}
}

func TestHLSFilter_FilterContent_LanguageFilter(t *testing.T) {

	masterManifestWithMultipleLangs := `#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="audio0",NAME="English",DEFAULT=YES,AUTOSELECT=YES,LANGUAGE="en",URI="https://existing.base/path/index-f8-a1.m3u8"
#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="audio0",NAME="Dansk",DEFAULT=NO,AUTOSELECT=NO,LANGUAGE="da",URI="https://existing.base/path/index-f10-a1.m3u8"
#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="audio0",NAME="Deutsch",DEFAULT=NO,AUTOSELECT=NO,LANGUAGE="de",URI="https://existing.base/path/index-f11-a1.m3u8"
#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="audio0",NAME="Spanish (Latin America)",DEFAULT=NO,AUTOSELECT=NO,LANGUAGE="es",URI="https://existing.base/path/index-f12-a1.m3u8"
#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="audio0",NAME="Brazilian Portuguese",DEFAULT=NO,AUTOSELECT=NO,LANGUAGE="pt",URI="https://existing.base/path/index-f16-a1.m3u8"
#EXT-X-MEDIA:TYPE=SUBTITLES,GROUP-ID="subs0",NAME="English",DEFAULT=YES,AUTOSELECT=YES,LANGUAGE="en",URI="https://existing.base/path/index-f19.m3u8"
#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=277965,CODECS="avc1.4d401e,mp4a.40.2",RESOLUTION=384x216,AUDIO="audio0",SUBTITLES="subs0",FRAME-RATE=25.000,VIDEO-RANGE=SDR
https://existing.base/path/index-f1-v1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=479857,CODECS="avc1.4d401e,mp4a.40.2",RESOLUTION=512x288,AUDIO="audio0",SUBTITLES="subs0",FRAME-RATE=25.000,VIDEO-RANGE=SDR
https://existing.base/path/index-f2-v1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=786893,CODECS="avc1.4d401e,mp4a.40.2",RESOLUTION=640x360,AUDIO="audio0",SUBTITLES="subs0",FRAME-RATE=25.000,VIDEO-RANGE=SDR
https://existing.base/path/index-f3-v1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=1015563,CODECS="avc1.4d401e,mp4a.40.2",RESOLUTION=768x432,AUDIO="audio0",SUBTITLES="subs0",FRAME-RATE=25.000,VIDEO-RANGE=SDR
https://existing.base/path/index-f4-v1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=1421115,CODECS="avc1.4d401e,mp4a.40.2",RESOLUTION=960x540,AUDIO="audio0",SUBTITLES="subs0",FRAME-RATE=25.000,VIDEO-RANGE=SDR
https://existing.base/path/index-f5-v1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=1956642,CODECS="avc1.640028,mp4a.40.2",RESOLUTION=1280x720,AUDIO="audio0",SUBTITLES="subs0",FRAME-RATE=25.000,VIDEO-RANGE=SDR
https://existing.base/path/index-f6-v1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=3712583,CODECS="avc1.640028,mp4a.40.2",RESOLUTION=1920x1080,AUDIO="audio0",SUBTITLES="subs0",FRAME-RATE=25.000,VIDEO-RANGE=SDR
https://existing.base/path/index-f7-v1.m3u8
`

	masterManifestWithEnglishOnly := `#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="audio0",NAME="English",DEFAULT=YES,AUTOSELECT=YES,LANGUAGE="en",URI="https://existing.base/path/index-f8-a1.m3u8"
#EXT-X-MEDIA:TYPE=SUBTITLES,GROUP-ID="subs0",NAME="English",DEFAULT=YES,AUTOSELECT=YES,LANGUAGE="en",URI="https://existing.base/path/index-f19.m3u8"
#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=277965,CODECS="avc1.4d401e,mp4a.40.2",RESOLUTION=384x216,AUDIO="audio0",SUBTITLES="subs0",FRAME-RATE=25.000,VIDEO-RANGE=SDR
https://existing.base/path/index-f1-v1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=479857,CODECS="avc1.4d401e,mp4a.40.2",RESOLUTION=512x288,AUDIO="audio0",SUBTITLES="subs0",FRAME-RATE=25.000,VIDEO-RANGE=SDR
https://existing.base/path/index-f2-v1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=786893,CODECS="avc1.4d401e,mp4a.40.2",RESOLUTION=640x360,AUDIO="audio0",SUBTITLES="subs0",FRAME-RATE=25.000,VIDEO-RANGE=SDR
https://existing.base/path/index-f3-v1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=1015563,CODECS="avc1.4d401e,mp4a.40.2",RESOLUTION=768x432,AUDIO="audio0",SUBTITLES="subs0",FRAME-RATE=25.000,VIDEO-RANGE=SDR
https://existing.base/path/index-f4-v1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=1421115,CODECS="avc1.4d401e,mp4a.40.2",RESOLUTION=960x540,AUDIO="audio0",SUBTITLES="subs0",FRAME-RATE=25.000,VIDEO-RANGE=SDR
https://existing.base/path/index-f5-v1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=1956642,CODECS="avc1.640028,mp4a.40.2",RESOLUTION=1280x720,AUDIO="audio0",SUBTITLES="subs0",FRAME-RATE=25.000,VIDEO-RANGE=SDR
https://existing.base/path/index-f6-v1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=3712583,CODECS="avc1.640028,mp4a.40.2",RESOLUTION=1920x1080,AUDIO="audio0",SUBTITLES="subs0",FRAME-RATE=25.000,VIDEO-RANGE=SDR
https://existing.base/path/index-f7-v1.m3u8
`

	masterManifestWithSpanishAndPortugeseAndNoSubtitles := `#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="audio0",NAME="Spanish (Latin America)",DEFAULT=NO,AUTOSELECT=NO,LANGUAGE="es",URI="https://existing.base/path/index-f12-a1.m3u8"
#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="audio0",NAME="Brazilian Portuguese",DEFAULT=NO,AUTOSELECT=NO,LANGUAGE="pt",URI="https://existing.base/path/index-f16-a1.m3u8"
#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=277965,CODECS="avc1.4d401e,mp4a.40.2",RESOLUTION=384x216,AUDIO="audio0",FRAME-RATE=25.000,VIDEO-RANGE=SDR
https://existing.base/path/index-f1-v1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=479857,CODECS="avc1.4d401e,mp4a.40.2",RESOLUTION=512x288,AUDIO="audio0",FRAME-RATE=25.000,VIDEO-RANGE=SDR
https://existing.base/path/index-f2-v1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=786893,CODECS="avc1.4d401e,mp4a.40.2",RESOLUTION=640x360,AUDIO="audio0",FRAME-RATE=25.000,VIDEO-RANGE=SDR
https://existing.base/path/index-f3-v1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=1015563,CODECS="avc1.4d401e,mp4a.40.2",RESOLUTION=768x432,AUDIO="audio0",FRAME-RATE=25.000,VIDEO-RANGE=SDR
https://existing.base/path/index-f4-v1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=1421115,CODECS="avc1.4d401e,mp4a.40.2",RESOLUTION=960x540,AUDIO="audio0",FRAME-RATE=25.000,VIDEO-RANGE=SDR
https://existing.base/path/index-f5-v1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=1956642,CODECS="avc1.640028,mp4a.40.2",RESOLUTION=1280x720,AUDIO="audio0",FRAME-RATE=25.000,VIDEO-RANGE=SDR
https://existing.base/path/index-f6-v1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=3712583,CODECS="avc1.640028,mp4a.40.2",RESOLUTION=1920x1080,AUDIO="audio0",FRAME-RATE=25.000,VIDEO-RANGE=SDR
https://existing.base/path/index-f7-v1.m3u8
`

	masterManifestWithEnglishSubsAndCaptions := `#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=SUBTITLES,GROUP-ID="subs",NAME="English",DEFAULT=YES,AUTOSELECT=YES,LANGUAGE="eng",FORCED="NO",URI="https://devcarrier-hls.amlg.io/captions/737e9315-fec6-4809-bd5e-3f2e427ec3a1/en.m3u8"
#EXT-X-MEDIA:TYPE=CLOSED-CAPTIONS,GROUP-ID="CC",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG"
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=356400,AVERAGE-BANDWIDTH=345400,CODECS="avc1.640015,mp4a.40.2",RESOLUTION=400x244,CLOSED-CAPTIONS="CC",SUBTITLES="subs",FRAME-RATE=29.000
https://cbss64eb-cbss64eb-ms-dev.global.ssl.fastly.net/cbssc0a7/master/cbssc0a7_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=642400,AVERAGE-BANDWIDTH=620400,CODECS="avc1.640015,mp4a.40.2",RESOLUTION=512x288,CLOSED-CAPTIONS="CC",SUBTITLES="subs",FRAME-RATE=29.000
https://cbss64eb-cbss64eb-ms-dev.global.ssl.fastly.net/cbssc0a7/master/cbssc0a7_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=985600,AVERAGE-BANDWIDTH=950400,CODECS="avc1.64001e,mp4a.40.2",RESOLUTION=640x360,CLOSED-CAPTIONS="CC",SUBTITLES="subs",FRAME-RATE=29.000
https://cbss64eb-cbss64eb-ms-dev.global.ssl.fastly.net/cbssc0a7/master/cbssc0a7_3.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1443200,AVERAGE-BANDWIDTH=1390400,CODECS="avc1.64001e,mp4a.40.2",RESOLUTION=768x432,CLOSED-CAPTIONS="CC",SUBTITLES="subs",FRAME-RATE=29.000
https://cbss64eb-cbss64eb-ms-dev.global.ssl.fastly.net/cbssc0a7/master/cbssc0a7_4.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=2129600,AVERAGE-BANDWIDTH=2050400,CODECS="avc1.64001f,mp4a.40.2",RESOLUTION=960x540,CLOSED-CAPTIONS="CC",SUBTITLES="subs",FRAME-RATE=29.000
https://cbss64eb-cbss64eb-ms-dev.global.ssl.fastly.net/cbssc0a7/master/cbssc0a7_5.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4074400,AVERAGE-BANDWIDTH=3920400,CODECS="avc1.64001f,mp4a.40.2",RESOLUTION=1280x720,CLOSED-CAPTIONS="CC",SUBTITLES="subs",FRAME-RATE=29.000
https://cbss64eb-cbss64eb-ms-dev.global.ssl.fastly.net/cbssc0a7/master/cbssc0a7_6.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=5790400,AVERAGE-BANDWIDTH=5570400,CODECS="avc1.640028,mp4a.40.2",RESOLUTION=1920x1080,CLOSED-CAPTIONS="CC",SUBTITLES="subs",FRAME-RATE=29.000
https://cbss64eb-cbss64eb-ms-dev.global.ssl.fastly.net/cbssc0a7/master/cbssc0a7_7.m3u8
`

	masterManifestWithNoEnglishSubsAndCaptions := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=356400,AVERAGE-BANDWIDTH=345400,CODECS="avc1.640015,mp4a.40.2",RESOLUTION=400x244,FRAME-RATE=29.000
https://cbss64eb-cbss64eb-ms-dev.global.ssl.fastly.net/cbssc0a7/master/cbssc0a7_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=642400,AVERAGE-BANDWIDTH=620400,CODECS="avc1.640015,mp4a.40.2",RESOLUTION=512x288,FRAME-RATE=29.000
https://cbss64eb-cbss64eb-ms-dev.global.ssl.fastly.net/cbssc0a7/master/cbssc0a7_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=985600,AVERAGE-BANDWIDTH=950400,CODECS="avc1.64001e,mp4a.40.2",RESOLUTION=640x360,FRAME-RATE=29.000
https://cbss64eb-cbss64eb-ms-dev.global.ssl.fastly.net/cbssc0a7/master/cbssc0a7_3.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1443200,AVERAGE-BANDWIDTH=1390400,CODECS="avc1.64001e,mp4a.40.2",RESOLUTION=768x432,FRAME-RATE=29.000
https://cbss64eb-cbss64eb-ms-dev.global.ssl.fastly.net/cbssc0a7/master/cbssc0a7_4.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=2129600,AVERAGE-BANDWIDTH=2050400,CODECS="avc1.64001f,mp4a.40.2",RESOLUTION=960x540,FRAME-RATE=29.000
https://cbss64eb-cbss64eb-ms-dev.global.ssl.fastly.net/cbssc0a7/master/cbssc0a7_5.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4074400,AVERAGE-BANDWIDTH=3920400,CODECS="avc1.64001f,mp4a.40.2",RESOLUTION=1280x720,FRAME-RATE=29.000
https://cbss64eb-cbss64eb-ms-dev.global.ssl.fastly.net/cbssc0a7/master/cbssc0a7_6.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=5790400,AVERAGE-BANDWIDTH=5570400,CODECS="avc1.640028,mp4a.40.2",RESOLUTION=1920x1080,FRAME-RATE=29.000
https://cbss64eb-cbss64eb-ms-dev.global.ssl.fastly.net/cbssc0a7/master/cbssc0a7_7.m3u8
`

	tests := []struct {
		name                  string
		filters               *parsers.MediaFilters
		manifestContent       string
		expectManifestContent string
		expectErr             bool
	}{
		{
			name:                  "when no filters are passed, nothing is removed",
			filters:               &parsers.MediaFilters{},
			manifestContent:       masterManifestWithMultipleLangs,
			expectManifestContent: masterManifestWithMultipleLangs,
		},
		{
			name: "when multiple audio languages are passed in, remove all those languages of type audio",
			filters: &parsers.MediaFilters{
				Audios: parsers.NestedFilters{
					Language: []string{"da", "de", "es", "pt"},
				},
			},
			manifestContent:       masterManifestWithMultipleLangs,
			expectManifestContent: masterManifestWithEnglishOnly,
		},
		{
			name: "when multiple audio and caption languages are passed in, remove provided audio and caption languages ",
			filters: &parsers.MediaFilters{
				Audios: parsers.NestedFilters{
					Language: []string{"da", "de", "en"},
				},
				Captions: parsers.NestedFilters{
					Language: []string{"en"},
				},
			},
			manifestContent:       masterManifestWithMultipleLangs,
			expectManifestContent: masterManifestWithSpanishAndPortugeseAndNoSubtitles,
		},
		{
			name: "when caption languages are passed in, remove provided subtitles and caption languages ",
			filters: &parsers.MediaFilters{
				Captions: parsers.NestedFilters{
					Language: []string{"eng"},
				},
			},
			manifestContent:       masterManifestWithEnglishSubsAndCaptions,
			expectManifestContent: masterManifestWithNoEnglishSubsAndCaptions,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			filter := NewHLSFilter("https://existing.base/path/master.m3u8", tt.manifestContent, config.Config{Hostname: "bakery.cbsi.video"})
			manifest, err := filter.FilterContent(context.Background(), tt.filters)

			if err != nil && !tt.expectErr {
				t.Errorf("FilterContent(context.Background(), ) didnt expect an error to be returned, got: %v", err)
				return
			} else if err == nil && tt.expectErr {
				t.Error("FilterContent(context.Background(), ) expected an error, got nil")
				return
			}

			if g, e := manifest, tt.expectManifestContent; g != e {
				t.Errorf("FilterContent(context.Background(), ) wrong manifest returned)\ngot %v\nexpected: %v\ndiff: %v", g, e,
					cmp.Diff(g, e))
			}

		})
	}
}

func TestHLSFilter_FilterContent_IFrameFilter(t *testing.T) {
	masterManifestWithSingleIFrame := `#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=CLOSED-CAPTIONS,GROUP-ID="CC",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG"
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,CODECS="avc1.64001f,mp4a.40.2",CLOSED-CAPTIONS="CC"
https://existing.base/path/link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4200,AVERAGE-BANDWIDTH=4200,CODECS="avc1.64001f,mp4a.40.2",CLOSED-CAPTIONS="CC"
https://existing.base/path/link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CODECS="avc1.64001f,mp4a.40.2",CLOSED-CAPTIONS="CC"
https://existing.base/path/link_4.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4100,AVERAGE-BANDWIDTH=4100,CODECS="avc1.64001f,mp4a.40.2",CLOSED-CAPTIONS="CC"
https://existing.base/path/link_5.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4500,AVERAGE-BANDWIDTH=4500,CODECS="avc1.64001f,mp4a.40.2",CLOSED-CAPTIONS="CC"
https://existing.base/path/link_6.m3u8
#EXT-X-I-FRAME-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=250,CODECS="avc1.4d401e",RESOLUTION=384x216,URI="https://existing.base/path/link_1.m3u8"
`

	masterManifestWithMultipleIFrame := `#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=CLOSED-CAPTIONS,GROUP-ID="CC",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG"
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,CODECS="avc1.64001f,mp4a.40.2",CLOSED-CAPTIONS="CC"
https://existing.base/path/link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4200,AVERAGE-BANDWIDTH=4200,CODECS="avc1.64001f,mp4a.40.2",CLOSED-CAPTIONS="CC"
https://existing.base/path/link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CODECS="avc1.64001f,mp4a.40.2",CLOSED-CAPTIONS="CC"
https://existing.base/path/link_4.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4100,AVERAGE-BANDWIDTH=4100,CODECS="avc1.64001f,mp4a.40.2",CLOSED-CAPTIONS="CC"
https://existing.base/path/link_5.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4500,AVERAGE-BANDWIDTH=4500,CODECS="avc1.64001f,mp4a.40.2",CLOSED-CAPTIONS="CC"
https://existing.base/path/link_6.m3u8
#EXT-X-I-FRAME-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=250,CODECS="avc1.4d401e",RESOLUTION=384x216,URI="link_1.m3u8",
#EXT-X-I-FRAME-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=550,CODECS="avc1.4d401e",RESOLUTION=512x288,URI="link_2.m3u8",
`

	masterManifestWithNoIframe := `#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=CLOSED-CAPTIONS,GROUP-ID="CC",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG"
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,CODECS="avc1.64001f,mp4a.40.2",CLOSED-CAPTIONS="CC"
https://existing.base/path/link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4200,AVERAGE-BANDWIDTH=4200,CODECS="avc1.64001f,mp4a.40.2",CLOSED-CAPTIONS="CC"
https://existing.base/path/link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CODECS="avc1.64001f,mp4a.40.2",CLOSED-CAPTIONS="CC"
https://existing.base/path/link_4.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4100,AVERAGE-BANDWIDTH=4100,CODECS="avc1.64001f,mp4a.40.2",CLOSED-CAPTIONS="CC"
https://existing.base/path/link_5.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4500,AVERAGE-BANDWIDTH=4500,CODECS="avc1.64001f,mp4a.40.2",CLOSED-CAPTIONS="CC"
https://existing.base/path/link_6.m3u8
`
	tests := []struct {
		name                  string
		filters               *parsers.MediaFilters
		manifestContent       string
		expectManifestContent string
		expectErr             bool
	}{
		{
			name:                  "when empty filter is given, expect no filtering to be done",
			filters:               &parsers.MediaFilters{},
			manifestContent:       masterManifestWithSingleIFrame,
			expectManifestContent: masterManifestWithSingleIFrame,
		},
		{
			name: "when iframe filter has been set, iframe is removed given a playlist with a single iframe",
			filters: &parsers.MediaFilters{
				Tags: &parsers.Tags{
					IFrame: true,
				},
			},
			manifestContent:       masterManifestWithSingleIFrame,
			expectManifestContent: masterManifestWithNoIframe,
		},
		{
			name: "when iframe filter has been set, iframe is removed given a playlist with mutliple iframe",
			filters: &parsers.MediaFilters{
				Tags: &parsers.Tags{
					IFrame: true,
				},
			},
			manifestContent:       masterManifestWithMultipleIFrame,
			expectManifestContent: masterManifestWithNoIframe,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			filter := NewHLSFilter("https://existing.base/path/master.m3u8", tt.manifestContent, config.Config{Hostname: "bakery.cbsi.video"})
			manifest, err := filter.FilterContent(context.Background(), tt.filters)

			if err != nil && !tt.expectErr {
				t.Errorf("FilterContent(context.Background(), ) didnt expect an error to be returned, got: %v", err)
				return
			} else if err == nil && tt.expectErr {
				t.Error("FilterContent(context.Background(), ) expected an error, got nil")
				return
			}

			if g, e := manifest, tt.expectManifestContent; g != e {
				t.Errorf("FilterContent(context.Background(), ) wrong manifest returned)\ngot %v\nexpected: %v\ndiff: %v", g, e,
					cmp.Diff(g, e))
			}

		})
	}
}

func TestHLSFilter_FilterContent_FrameRate(t *testing.T) {
	masterManifestWithMultipleFrameRates := `#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=CLOSED-CAPTIONS,GROUP-ID="CC",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG"
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,CODECS="avc1.64001f,mp4a.40.2",CLOSED-CAPTIONS="CC",FRAME-RATE=30.000
https://existing.base/path/link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4200,AVERAGE-BANDWIDTH=4200,CODECS="avc1.64001f,mp4a.40.2",CLOSED-CAPTIONS="CC",FRAME-RATE=30.000
https://existing.base/path/link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CODECS="avc1.64001f,mp4a.40.2",CLOSED-CAPTIONS="CC",FRAME-RATE=59.940
https://existing.base/path/link_4.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4100,AVERAGE-BANDWIDTH=4100,CODECS="avc1.64001f,mp4a.40.2",CLOSED-CAPTIONS="CC",FRAME-RATE=59.940
https://existing.base/path/link_5.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4500,AVERAGE-BANDWIDTH=4500,CODECS="avc1.64001f,mp4a.40.2",CLOSED-CAPTIONS="CC",FRAME-RATE=60.000
https://existing.base/path/link_6.m3u8
`

	masterManifestWithout5994FrameRate := `#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=CLOSED-CAPTIONS,GROUP-ID="CC",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG"
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1000,AVERAGE-BANDWIDTH=1000,CODECS="avc1.64001f,mp4a.40.2",CLOSED-CAPTIONS="CC",FRAME-RATE=30.000
https://existing.base/path/link_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4200,AVERAGE-BANDWIDTH=4200,CODECS="avc1.64001f,mp4a.40.2",CLOSED-CAPTIONS="CC",FRAME-RATE=30.000
https://existing.base/path/link_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4500,AVERAGE-BANDWIDTH=4500,CODECS="avc1.64001f,mp4a.40.2",CLOSED-CAPTIONS="CC",FRAME-RATE=60.000
https://existing.base/path/link_6.m3u8
`

	masterManifestWithOnly5994FrameRate := `#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=CLOSED-CAPTIONS,GROUP-ID="CC",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG"
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4000,AVERAGE-BANDWIDTH=4000,CODECS="avc1.64001f,mp4a.40.2",CLOSED-CAPTIONS="CC",FRAME-RATE=59.940
https://existing.base/path/link_4.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=4100,AVERAGE-BANDWIDTH=4100,CODECS="avc1.64001f,mp4a.40.2",CLOSED-CAPTIONS="CC",FRAME-RATE=59.940
https://existing.base/path/link_5.m3u8
`

	masterManifestWithNoVariants := `#EXTM3U
#EXT-X-VERSION:3
`

	tests := []struct {
		name                  string
		filters               *parsers.MediaFilters
		manifestContent       string
		expectManifestContent string
		expectErr             bool
	}{
		{
			name:                  "when empty filter is given, expect no filtering to be done",
			filters:               &parsers.MediaFilters{},
			manifestContent:       masterManifestWithMultipleFrameRates,
			expectManifestContent: masterManifestWithMultipleFrameRates,
		},
		{
			name: "when framerate is set to 60, variant of 60fps is removed",
			filters: &parsers.MediaFilters{
				FrameRate: []string{"60.000", "30.000"},
			},
			manifestContent:       masterManifestWithMultipleFrameRates,
			expectManifestContent: masterManifestWithOnly5994FrameRate,
		},
		{
			name: "when framerate is set to 59.94, both variants of 59.94fps is removed",
			filters: &parsers.MediaFilters{
				FrameRate: []string{"59.940"},
			},
			manifestContent:       masterManifestWithMultipleFrameRates,
			expectManifestContent: masterManifestWithout5994FrameRate,
		},
		{
			name: "when all framerates are set, empty playlist is returned",
			filters: &parsers.MediaFilters{
				FrameRate: []string{"59.940", "60.000", "30.000"},
			},
			manifestContent:       masterManifestWithMultipleFrameRates,
			expectManifestContent: masterManifestWithNoVariants,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewHLSFilter("https://existing.base/path/master.m3u8", tt.manifestContent, config.Config{Hostname: "bakery.cbsi.video"})
			manifest, err := filter.FilterContent(context.Background(), tt.filters)

			if err != nil && !tt.expectErr {
				t.Errorf("FilterContent(context.Background(), ) didnt expect an error to be returned, got: %v", err)
				return
			} else if err == nil && tt.expectErr {
				t.Error("FilterContent(context.Background(), ) expected an error, got nil")
				return
			}

			if g, e := manifest, tt.expectManifestContent; g != e {
				t.Errorf("FilterContent(context.Background(), ) wrong manifest returned)\ngot %v\nexpected: %v\ndiff: %v", g, e,
					cmp.Diff(g, e))
			}

		})
	}
}

func TestHLSFilter_FilterContent_RedundantManifests(t *testing.T) {
	redundant := `#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=CLOSED-CAPTIONS,GROUP-ID="CC",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG",INSTREAM-ID="CC1"
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=356400,AVERAGE-BANDWIDTH=345400,CODECS="avc1.64000c,mp4a.40.2",RESOLUTION=400x224,CLOSED-CAPTIONS="CC",FRAME-RATE=15.000
https://cbsi679d-cbsi679d-ms-dev.global.ssl.fastly.net/testa5fe/master/testa5fe_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=356400,AVERAGE-BANDWIDTH=345400,CODECS="avc1.64000c,mp4a.40.2",RESOLUTION=400x224,CLOSED-CAPTIONS="CC",FRAME-RATE=15.000
https://cbsi679d-cbsi679d-ms-dev.global.ssl.fastly.net/testa5fe/master/backup_testa5fe_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=528000,AVERAGE-BANDWIDTH=510400,CODECS="avc1.640015,mp4a.40.2",RESOLUTION=512x288,CLOSED-CAPTIONS="CC",FRAME-RATE=15.000
https://cbsi679d-cbsi679d-ms-dev.global.ssl.fastly.net/testa5fe/master/testa5fe_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=528000,AVERAGE-BANDWIDTH=510400,CODECS="avc1.640015,mp4a.40.2",RESOLUTION=512x288,CLOSED-CAPTIONS="CC",FRAME-RATE=15.000
https://cbsi679d-cbsi679d-ms-dev.global.ssl.fastly.net/testa5fe/master/backup_testa5fe_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1100000,AVERAGE-BANDWIDTH=1060400,CODECS="avc1.64001e,mp4a.40.2",RESOLUTION=640x360,CLOSED-CAPTIONS="CC",FRAME-RATE=30.000
https://cbsi679d-cbsi679d-ms-dev.global.ssl.fastly.net/testa5fe/master/testa5fe_3.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1100000,AVERAGE-BANDWIDTH=1060400,CODECS="avc1.64001e,mp4a.40.2",RESOLUTION=640x360,CLOSED-CAPTIONS="CC",FRAME-RATE=30.000
https://cbsi679d-cbsi679d-ms-dev.global.ssl.fastly.net/testa5fe/master/backup_testa5fe_3.m3u8
`
	primary := `#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=CLOSED-CAPTIONS,GROUP-ID="CC",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG",INSTREAM-ID="CC1"
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=356400,AVERAGE-BANDWIDTH=345400,CODECS="avc1.64000c,mp4a.40.2",RESOLUTION=400x224,CLOSED-CAPTIONS="CC",FRAME-RATE=15.000
https://cbsi679d-cbsi679d-ms-dev.global.ssl.fastly.net/testa5fe/master/testa5fe_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=528000,AVERAGE-BANDWIDTH=510400,CODECS="avc1.640015,mp4a.40.2",RESOLUTION=512x288,CLOSED-CAPTIONS="CC",FRAME-RATE=15.000
https://cbsi679d-cbsi679d-ms-dev.global.ssl.fastly.net/testa5fe/master/testa5fe_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1100000,AVERAGE-BANDWIDTH=1060400,CODECS="avc1.64001e,mp4a.40.2",RESOLUTION=640x360,CLOSED-CAPTIONS="CC",FRAME-RATE=30.000
https://cbsi679d-cbsi679d-ms-dev.global.ssl.fastly.net/testa5fe/master/testa5fe_3.m3u8
`
	backup := `#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=CLOSED-CAPTIONS,GROUP-ID="CC",NAME="ENGLISH",DEFAULT=NO,LANGUAGE="ENG",INSTREAM-ID="CC1"
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=356400,AVERAGE-BANDWIDTH=345400,CODECS="avc1.64000c,mp4a.40.2",RESOLUTION=400x224,CLOSED-CAPTIONS="CC",FRAME-RATE=15.000
https://cbsi679d-cbsi679d-ms-dev.global.ssl.fastly.net/testa5fe/master/backup_testa5fe_1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=528000,AVERAGE-BANDWIDTH=510400,CODECS="avc1.640015,mp4a.40.2",RESOLUTION=512x288,CLOSED-CAPTIONS="CC",FRAME-RATE=15.000
https://cbsi679d-cbsi679d-ms-dev.global.ssl.fastly.net/testa5fe/master/backup_testa5fe_2.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=0,BANDWIDTH=1100000,AVERAGE-BANDWIDTH=1060400,CODECS="avc1.64001e,mp4a.40.2",RESOLUTION=640x360,CLOSED-CAPTIONS="CC",FRAME-RATE=30.000
https://cbsi679d-cbsi679d-ms-dev.global.ssl.fastly.net/testa5fe/master/backup_testa5fe_3.m3u8
`

	variant := `#EXTM3U
	#EXT-X-VERSION:3
	#EXT-X-TARGETDURATION:8
`
	tests := []struct {
		name           string
		filters        *parsers.MediaFilters
		mockResp       func(req *http.Request) (*http.Response, error)
		expectManifest string
		expectErr      bool
	}{
		{
			name: "when redundant manifest returns 4xx for primary manifest, return backup manifest only",
			filters: &parsers.MediaFilters{
				DeWeave: true,
			},
			mockResp: func(*http.Request) (*http.Response, error) {
				resp := &http.Response{
					StatusCode: 404,
					Body:       ioutil.NopCloser(bytes.NewBufferString("NotFound")),
					Header:     http.Header{},
				}

				lastModified := time.Now().UTC().Add(-16 * time.Second).Format(http.TimeFormat)
				resp.Header.Add("Last-Modified", lastModified)

				return resp, nil
			},
			expectManifest: backup,
		},
		{
			name: "when redundant manifest returns 2xx but LastModified time is 2x segment length, return backup manifest only",
			filters: &parsers.MediaFilters{
				DeWeave: true,
			},
			mockResp: func(*http.Request) (*http.Response, error) {
				resp := &http.Response{
					StatusCode: 200,
					Body:       ioutil.NopCloser(bytes.NewBufferString(variant)),
					Header:     http.Header{},
				}

				lastModified := time.Now().UTC().Add(-16 * time.Second).Format(http.TimeFormat)
				resp.Header.Add("Last-Modified", lastModified)

				return resp, nil
			},
			expectManifest: backup,
		},
		{
			name: "when redundant manifest returns 2xx for primary manifest and is not stale, return primary manifest only",
			filters: &parsers.MediaFilters{
				DeWeave: true,
			},
			mockResp: func(*http.Request) (*http.Response, error) {
				resp := &http.Response{
					StatusCode: 200,
					Body:       ioutil.NopCloser(bytes.NewBufferString(variant)),
					Header:     http.Header{},
				}

				lastModified := time.Now().UTC().Format(http.TimeFormat)
				resp.Header.Add("Last-Modified", lastModified)

				return resp, nil
			},
			expectManifest: primary,
		},
		{
			name: "when last modified not set, default to backup",
			filters: &parsers.MediaFilters{
				DeWeave: true,
			},
			mockResp: func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Body:       ioutil.NopCloser(bytes.NewBufferString(primary)),
				}, nil
			},
			expectManifest: backup,
		},
		{
			name: "when manifest request throws http status error, expect error",
			filters: &parsers.MediaFilters{
				DeWeave: true,
			},
			mockResp: func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 500,
					Body:       ioutil.NopCloser(bytes.NewBufferString("")),
				}, nil
			},
			expectManifest: "",
			expectErr:      true,
		},
		{
			name: "when Last-Modified header is not in proper format, expect error",
			filters: &parsers.MediaFilters{
				DeWeave: true,
			},
			mockResp: func(*http.Request) (*http.Response, error) {
				resp := &http.Response{
					StatusCode: 200,
					Body:       ioutil.NopCloser(bytes.NewBufferString(variant)),
					Header:     http.Header{},
				}

				resp.Header.Add("Last-Modified", "hello")

				return resp, nil
			},
			expectManifest: "",
			expectErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{
				Hostname: "bakery.cbsi.video",
				Client: config.Client{
					Timeout:    5 * time.Second,
					Tracer:     tracing.NoopTracer{},
					HTTPClient: test.MockClient(tt.mockResp),
				},
			}
			filter := NewHLSFilter("https://existing.base/path/master.m3u8", redundant, cfg)
			manifest, err := filter.FilterContent(context.Background(), tt.filters)

			if err != nil && !tt.expectErr {
				t.Errorf("FilterContent(context.Background(), ) didnt expect an error to be returned, got: %v", err)
				return
			} else if err == nil && tt.expectErr {
				t.Error("FilterContent(context.Background(), ) expected an error, got nil")
				return
			}

			if g, e := manifest, tt.expectManifest; g != e {
				t.Errorf("FilterContent(context.Background(), ) wrong manifest returned)\ngot %v\nexpected: %v\ndiff: %v", g, e,
					cmp.Diff(g, e))
			}
		})
	}
}
