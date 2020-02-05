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
