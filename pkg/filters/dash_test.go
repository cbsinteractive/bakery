package filters

import (
	"fmt"
	"testing"

	"github.com/cbsinteractive/bakery/pkg/config"
	"github.com/cbsinteractive/bakery/pkg/parsers"
	"github.com/google/go-cmp/cmp"
)

func TestDASHFilter_FilterManifest_baseURL(t *testing.T) {
	manifestWithoutBaseURL := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT6M16S" minBufferTime="PT1.97S">
</MPD>
`

	manifestWithAbsoluteBaseURL := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT6M16S" minBufferTime="PT1.97S">
  <BaseURL>http://some.absolute/base/url/</BaseURL>
</MPD>
`

	manifestWithBaseURL := func(baseURL string) string {
		return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT6M16S" minBufferTime="PT1.97S">
  <BaseURL>%s</BaseURL>
</MPD>
`, baseURL)
	}

	tests := []struct {
		name                  string
		manifestURL           string
		manifestContent       string
		expectManifestContent string
		expectErr             bool
	}{
		{
			name: "when no baseURL is set, the correct absolute baseURL is added relative to the " +
				"manifest URL",
			manifestURL:           "http://some.url/to/the/manifest.mpd",
			manifestContent:       manifestWithoutBaseURL,
			expectManifestContent: manifestWithBaseURL("http://some.url/to/the/"),
		},
		{
			name:                  "when an absolute baseURL is set, the manifest is unchanged",
			manifestURL:           "http://some.url/to/the/manifest.mpd",
			manifestContent:       manifestWithAbsoluteBaseURL,
			expectManifestContent: manifestWithAbsoluteBaseURL,
		},
		{
			name: "when a relative baseURL is set, the correct absolute baseURL is added relative " +
				"to the manifest URL and the provided relative baseURL",
			manifestURL:           "http://some.url/to/the/manifest.mpd",
			manifestContent:       manifestWithBaseURL("../some/other/path/"),
			expectManifestContent: manifestWithBaseURL("http://some.url/to/some/other/path/"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			filter := NewDASHFilter(tt.manifestURL, tt.manifestContent, config.Config{})

			manifest, err := filter.FilterManifest(&parsers.MediaFilters{})
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
func TestDASHFilter_FilterManifest_audioCodecs(t *testing.T) {
	manifestWithEAC3AndEAC3AudioCodec := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT6M16S" minBufferTime="PT1.97S">
  <BaseURL>http://existing.base/url/</BaseURL>
  <Period>
    <AdaptationSet id="0" lang="en" contentType="video">
      <Representation bandwidth="256" codecs="avc" id="0"></Representation>
    </AdaptationSet>
    <AdaptationSet id="1" lang="en" contentType="audio">
      <Representation bandwidth="256" codecs="ec-3" id="0"></Representation>
      <Representation bandwidth="256" codecs="ac-3" id="1"></Representation>
    </AdaptationSet>
  </Period>
</MPD>
`

	manifestWithEAC3AudioCodec := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT6M16S" minBufferTime="PT1.97S">
  <BaseURL>http://existing.base/url/</BaseURL>
  <Period>
    <AdaptationSet id="0" lang="en" contentType="video">
      <Representation bandwidth="256" codecs="avc" id="0"></Representation>
    </AdaptationSet>
    <AdaptationSet id="1" lang="en" contentType="audio">
      <Representation bandwidth="256" codecs="ec-3" id="0"></Representation>
    </AdaptationSet>
  </Period>
</MPD>
`

	manifestWithAC3AudioCodec := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT6M16S" minBufferTime="PT1.97S">
  <BaseURL>http://existing.base/url/</BaseURL>
  <Period>
    <AdaptationSet id="0" lang="en" contentType="video">
      <Representation bandwidth="256" codecs="avc" id="0"></Representation>
    </AdaptationSet>
    <AdaptationSet id="1" lang="en" contentType="audio">
      <Representation bandwidth="256" codecs="ac-3" id="1"></Representation>
    </AdaptationSet>
  </Period>
</MPD>
`

	manifestWithoutAudio := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT6M16S" minBufferTime="PT1.97S">
  <BaseURL>http://existing.base/url/</BaseURL>
  <Period>
    <AdaptationSet id="0" lang="en" contentType="video">
      <Representation bandwidth="256" codecs="avc" id="0"></Representation>
    </AdaptationSet>
    <AdaptationSet id="1" lang="en" contentType="audio"></AdaptationSet>
  </Period>
</MPD>
`

	tests := []struct {
		name                  string
		filters               *parsers.MediaFilters
		manifestContent       string
		expectManifestContent string
		expectErr             bool
	}{
		{
			name:                  "when an empty audio filter list is supplied, audio is stripped from a manifest",
			filters:               &parsers.MediaFilters{Audios: []parsers.AudioType{}},
			manifestContent:       manifestWithEAC3AndEAC3AudioCodec,
			expectManifestContent: manifestWithoutAudio,
		},
		{
			name:                  "when an audio filter is supplied with Enhanced AC-3 codec, AC-3 is stripped out",
			filters:               &parsers.MediaFilters{Audios: []parsers.AudioType{"ec-3"}},
			manifestContent:       manifestWithEAC3AndEAC3AudioCodec,
			expectManifestContent: manifestWithEAC3AudioCodec,
		},
		{
			name:                  "when an audio filter is supplied with AC-3 codec, Enhanced AC-3 is stripped out",
			filters:               &parsers.MediaFilters{Audios: []parsers.AudioType{"ac-3"}},
			manifestContent:       manifestWithEAC3AndEAC3AudioCodec,
			expectManifestContent: manifestWithAC3AudioCodec,
		},
		{
			name:                  "when no audio filters are supplied, nothing is stripped from manifest",
			filters:               &parsers.MediaFilters{},
			manifestContent:       manifestWithEAC3AndEAC3AudioCodec,
			expectManifestContent: manifestWithEAC3AndEAC3AudioCodec,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			filter := NewDASHFilter("", tt.manifestContent, config.Config{})

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

func TestDASHFilter_FilterManifest_captionTypes(t *testing.T) {
	manifestWithWVTTAndSTPPCaptions := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT6M16S" minBufferTime="PT1.97S">
  <BaseURL>http://existing.base/url/</BaseURL>
  <Period>
    <AdaptationSet id="7357" lang="en" contentType="text">
      <Representation bandwidth="256" codecs="wvtt" id="subtitle_en"></Representation>
      <Representation bandwidth="256" codecs="stpp" id="subtitle_en_ttml"></Representation>
    </AdaptationSet>
  </Period>
</MPD>
`

	manifestWithWVTTCaptions := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT6M16S" minBufferTime="PT1.97S">
  <BaseURL>http://existing.base/url/</BaseURL>
  <Period>
    <AdaptationSet id="7357" lang="en" contentType="text">
      <Representation bandwidth="256" codecs="wvtt" id="subtitle_en"></Representation>
    </AdaptationSet>
  </Period>
</MPD>
`

	manifestWithSTPPCaptions := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT6M16S" minBufferTime="PT1.97S">
  <BaseURL>http://existing.base/url/</BaseURL>
  <Period>
    <AdaptationSet id="7357" lang="en" contentType="text">
      <Representation bandwidth="256" codecs="stpp" id="subtitle_en_ttml"></Representation>
    </AdaptationSet>
  </Period>
</MPD>
`

	manifestWithoutCaptions := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT6M16S" minBufferTime="PT1.97S">
  <BaseURL>http://existing.base/url/</BaseURL>
  <Period>
    <AdaptationSet id="7357" lang="en" contentType="text"></AdaptationSet>
  </Period>
</MPD>
`

	tests := []struct {
		name                  string
		filters               *parsers.MediaFilters
		manifestContent       string
		expectManifestContent string
		expectErr             bool
	}{
		{
			name: "when an empty caption type filter list is supplied, captions are stripped from a " +
				"manifest",
			filters:               &parsers.MediaFilters{CaptionTypes: []parsers.CaptionType{}},
			manifestContent:       manifestWithWVTTAndSTPPCaptions,
			expectManifestContent: manifestWithoutCaptions,
		},
		{
			name: "when a caption type filter is supplied with stpp only, webvtt captions are " +
				"filtered out",
			filters:               &parsers.MediaFilters{CaptionTypes: []parsers.CaptionType{"stpp"}},
			manifestContent:       manifestWithWVTTAndSTPPCaptions,
			expectManifestContent: manifestWithSTPPCaptions,
		},
		{
			name: "when a caption type filter is supplied with wvtt only, stpp captions are " +
				"filtered out",
			filters:               &parsers.MediaFilters{CaptionTypes: []parsers.CaptionType{"wvtt"}},
			manifestContent:       manifestWithWVTTAndSTPPCaptions,
			expectManifestContent: manifestWithWVTTCaptions,
		},
		{
			name:                  "when no filters are supplied, captions are not stripped from a manifest",
			filters:               &parsers.MediaFilters{},
			manifestContent:       manifestWithWVTTAndSTPPCaptions,
			expectManifestContent: manifestWithWVTTAndSTPPCaptions,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			filter := NewDASHFilter("", tt.manifestContent, config.Config{})

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

func TestDASHFilter_FilterManifest_filterStreams(t *testing.T) {
	manifestWithAudioAndVideoStreams := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT6M16S" minBufferTime="PT1.97S">
  <BaseURL>http://existing.base/url/</BaseURL>
  <Period id="0">
    <AdaptationSet id="0" lang="en" contentType="video"></AdaptationSet>
    <AdaptationSet id="1" lang="en" contentType="video"></AdaptationSet>
    <AdaptationSet id="2" lang="en" contentType="video"></AdaptationSet>
    <AdaptationSet id="3" lang="en" contentType="audio"></AdaptationSet>
    <AdaptationSet id="4" lang="en" contentType="audio"></AdaptationSet>
  </Period>
  <Period id="1">
    <AdaptationSet id="0" lang="en" contentType="video"></AdaptationSet>
    <AdaptationSet id="1" lang="en" contentType="video"></AdaptationSet>
    <AdaptationSet id="2" lang="en" contentType="video"></AdaptationSet>
    <AdaptationSet id="3" lang="en" contentType="audio"></AdaptationSet>
    <AdaptationSet id="4" lang="en" contentType="audio"></AdaptationSet>
  </Period>
  <Period id="2">
    <AdaptationSet id="0" lang="en" contentType="video"></AdaptationSet>
    <AdaptationSet id="1" lang="en" contentType="video"></AdaptationSet>
    <AdaptationSet id="2" lang="en" contentType="video"></AdaptationSet>
  </Period>
</MPD>
`

	manifestWithOnlyAudioStreams := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT6M16S" minBufferTime="PT1.97S">
  <BaseURL>http://existing.base/url/</BaseURL>
  <Period id="0">
    <AdaptationSet id="0" lang="en" contentType="audio"></AdaptationSet>
    <AdaptationSet id="1" lang="en" contentType="audio"></AdaptationSet>
  </Period>
  <Period id="1">
    <AdaptationSet id="0" lang="en" contentType="audio"></AdaptationSet>
    <AdaptationSet id="1" lang="en" contentType="audio"></AdaptationSet>
  </Period>
</MPD>
`

	manifestWithOnlyVideoStreams := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT6M16S" minBufferTime="PT1.97S">
  <BaseURL>http://existing.base/url/</BaseURL>
  <Period id="0">
    <AdaptationSet id="0" lang="en" contentType="video"></AdaptationSet>
    <AdaptationSet id="1" lang="en" contentType="video"></AdaptationSet>
    <AdaptationSet id="2" lang="en" contentType="video"></AdaptationSet>
  </Period>
  <Period id="1">
    <AdaptationSet id="0" lang="en" contentType="video"></AdaptationSet>
    <AdaptationSet id="1" lang="en" contentType="video"></AdaptationSet>
    <AdaptationSet id="2" lang="en" contentType="video"></AdaptationSet>
  </Period>
  <Period id="2">
    <AdaptationSet id="0" lang="en" contentType="video"></AdaptationSet>
    <AdaptationSet id="1" lang="en" contentType="video"></AdaptationSet>
    <AdaptationSet id="2" lang="en" contentType="video"></AdaptationSet>
  </Period>
</MPD>
`

	manifestWithoutStreams := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT6M16S" minBufferTime="PT1.97S">
  <BaseURL>http://existing.base/url/</BaseURL>
</MPD>
`

	tests := []struct {
		name                  string
		filters               *parsers.MediaFilters
		manifestContent       string
		expectManifestContent string
	}{
		{
			name:                  "when no streams are configured to be filtered, the manifest is not modified",
			filters:               &parsers.MediaFilters{},
			manifestContent:       manifestWithAudioAndVideoStreams,
			expectManifestContent: manifestWithAudioAndVideoStreams,
		},
		{
			name:                  "when video streams are filtered, the manifest contains no video adaptation sets",
			filters:               &parsers.MediaFilters{FilterStreamTypes: []parsers.StreamType{"video"}},
			manifestContent:       manifestWithAudioAndVideoStreams,
			expectManifestContent: manifestWithOnlyAudioStreams,
		},
		{
			name:                  "when audio streams are filtered, the manifest contains no audio adaptation sets",
			filters:               &parsers.MediaFilters{FilterStreamTypes: []parsers.StreamType{"audio"}},
			manifestContent:       manifestWithAudioAndVideoStreams,
			expectManifestContent: manifestWithOnlyVideoStreams,
		},
		{
			name: "when audio and video streams are filtered, the manifest contains no audio or " +
				"video adaptation sets",
			filters:               &parsers.MediaFilters{FilterStreamTypes: []parsers.StreamType{"video", "audio"}},
			manifestContent:       manifestWithAudioAndVideoStreams,
			expectManifestContent: manifestWithoutStreams,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			filter := NewDASHFilter("", tt.manifestContent, config.Config{})

			manifest, err := filter.FilterManifest(tt.filters)
			if err != nil {
				t.Errorf("FilterManifest() didnt expect an error to be returned, got: %v", err)
				return
			}

			if g, e := manifest, tt.expectManifestContent; g != e {
				t.Errorf("FilterManifest() wrong manifest returned\ngot %v\nexpected: %v\ndiff: %v", g, e,
					cmp.Diff(g, e))
			}
		})
	}
}
