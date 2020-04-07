package filters

import (
	"fmt"
	"math"
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

func TestDASHFilter_FilterManifest_videoCodecs(t *testing.T) {
	manifestWithMultiVideoCodec := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT6M16S" minBufferTime="PT1.97S">
  <BaseURL>http://existing.base/url/</BaseURL>
  <Period>
    <AdaptationSet id="0" lang="en" contentType="video">
      <Representation bandwidth="256" codecs="hvc1.2.4.L93.90" id="0"></Representation>
      <Representation bandwidth="256" codecs="hvc1.2.4.L90.90" id="1"></Representation>
      <Representation bandwidth="256" codecs="hvc1.2.4.L120.90" id="2"></Representation>
      <Representation bandwidth="256" codecs="hvc1.2.4.L63.90" id="3"></Representation>
      <Representation bandwidth="256" codecs="hvc1.1.4.L120.90" id="4"></Representation>
      <Representation bandwidth="256" codecs="hvc1.1.4.L63.90" id="5"></Representation>
      <Representation bandwidth="256" codecs="hev1.1.4.L120.90" id="6"></Representation>
      <Representation bandwidth="256" codecs="hev1.1.4.L63.90" id="7"></Representation>
      <Representation bandwidth="256" codecs="hev1.2.4.L120.90" id="8"></Representation>
      <Representation bandwidth="256" codecs="hev1.3.4.L63.90" id="9"></Representation>
    </AdaptationSet>
    <AdaptationSet id="1" lang="en" contentType="video">
      <Representation bandwidth="256" codecs="dvh1.05.01" id="0"></Representation>
      <Representation bandwidth="256" codecs="dvh1.05.03" id="1"></Representation>
    </AdaptationSet>
    <AdaptationSet id="2" lang="en" contentType="video">
      <Representation bandwidth="256" codecs="avc1.640028" id="0"></Representation>
    </AdaptationSet>
    <AdaptationSet id="3" lang="en" contentType="audio">
      <Representation bandwidth="256" codecs="mp4a.40.2" id="0"></Representation>
    </AdaptationSet>
    <AdaptationSet id="4" lang="en" contentType="text">
      <Representation bandwidth="256" codecs="wvtt" id="0"></Representation>
    </AdaptationSet>
  </Period>
</MPD>
`

	manifestWithoutDolbyVisionCodec := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT6M16S" minBufferTime="PT1.97S">
  <BaseURL>http://existing.base/url/</BaseURL>
  <Period>
    <AdaptationSet id="0" lang="en" contentType="video">
      <Representation bandwidth="256" codecs="hvc1.2.4.L93.90" id="0"></Representation>
      <Representation bandwidth="256" codecs="hvc1.2.4.L90.90" id="1"></Representation>
      <Representation bandwidth="256" codecs="hvc1.2.4.L120.90" id="2"></Representation>
      <Representation bandwidth="256" codecs="hvc1.2.4.L63.90" id="3"></Representation>
      <Representation bandwidth="256" codecs="hvc1.1.4.L120.90" id="4"></Representation>
      <Representation bandwidth="256" codecs="hvc1.1.4.L63.90" id="5"></Representation>
      <Representation bandwidth="256" codecs="hev1.1.4.L120.90" id="6"></Representation>
      <Representation bandwidth="256" codecs="hev1.1.4.L63.90" id="7"></Representation>
      <Representation bandwidth="256" codecs="hev1.2.4.L120.90" id="8"></Representation>
      <Representation bandwidth="256" codecs="hev1.3.4.L63.90" id="9"></Representation>
    </AdaptationSet>
    <AdaptationSet id="1" lang="en" contentType="video">
      <Representation bandwidth="256" codecs="avc1.640028" id="0"></Representation>
    </AdaptationSet>
    <AdaptationSet id="2" lang="en" contentType="audio">
      <Representation bandwidth="256" codecs="mp4a.40.2" id="0"></Representation>
    </AdaptationSet>
    <AdaptationSet id="3" lang="en" contentType="text">
      <Representation bandwidth="256" codecs="wvtt" id="0"></Representation>
    </AdaptationSet>
  </Period>
</MPD>
`

	manifestWithoutHEVCAndAVCVideoCodec := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT6M16S" minBufferTime="PT1.97S">
  <BaseURL>http://existing.base/url/</BaseURL>
  <Period>
    <AdaptationSet id="0" lang="en" contentType="video">
      <Representation bandwidth="256" codecs="dvh1.05.01" id="0"></Representation>
      <Representation bandwidth="256" codecs="dvh1.05.03" id="1"></Representation>
    </AdaptationSet>
    <AdaptationSet id="1" lang="en" contentType="audio">
      <Representation bandwidth="256" codecs="mp4a.40.2" id="0"></Representation>
    </AdaptationSet>
    <AdaptationSet id="2" lang="en" contentType="text">
      <Representation bandwidth="256" codecs="wvtt" id="0"></Representation>
    </AdaptationSet>
  </Period>
</MPD>
`

	manifestWithoutAVCVideoCodec := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT6M16S" minBufferTime="PT1.97S">
  <BaseURL>http://existing.base/url/</BaseURL>
  <Period>
    <AdaptationSet id="0" lang="en" contentType="video">
      <Representation bandwidth="256" codecs="hvc1.2.4.L93.90" id="0"></Representation>
      <Representation bandwidth="256" codecs="hvc1.2.4.L90.90" id="1"></Representation>
      <Representation bandwidth="256" codecs="hvc1.2.4.L120.90" id="2"></Representation>
      <Representation bandwidth="256" codecs="hvc1.2.4.L63.90" id="3"></Representation>
      <Representation bandwidth="256" codecs="hvc1.1.4.L120.90" id="4"></Representation>
      <Representation bandwidth="256" codecs="hvc1.1.4.L63.90" id="5"></Representation>
      <Representation bandwidth="256" codecs="hev1.1.4.L120.90" id="6"></Representation>
      <Representation bandwidth="256" codecs="hev1.1.4.L63.90" id="7"></Representation>
      <Representation bandwidth="256" codecs="hev1.2.4.L120.90" id="8"></Representation>
      <Representation bandwidth="256" codecs="hev1.3.4.L63.90" id="9"></Representation>
    </AdaptationSet>
    <AdaptationSet id="1" lang="en" contentType="video">
      <Representation bandwidth="256" codecs="dvh1.05.01" id="0"></Representation>
      <Representation bandwidth="256" codecs="dvh1.05.03" id="1"></Representation>
    </AdaptationSet>
    <AdaptationSet id="2" lang="en" contentType="audio">
      <Representation bandwidth="256" codecs="mp4a.40.2" id="0"></Representation>
    </AdaptationSet>
    <AdaptationSet id="3" lang="en" contentType="text">
      <Representation bandwidth="256" codecs="wvtt" id="0"></Representation>
    </AdaptationSet>
  </Period>
</MPD>
`

	manifestWithoutHEVCVideoCodec := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT6M16S" minBufferTime="PT1.97S">
  <BaseURL>http://existing.base/url/</BaseURL>
  <Period>
    <AdaptationSet id="0" lang="en" contentType="video">
      <Representation bandwidth="256" codecs="dvh1.05.01" id="0"></Representation>
      <Representation bandwidth="256" codecs="dvh1.05.03" id="1"></Representation>
    </AdaptationSet>
    <AdaptationSet id="1" lang="en" contentType="video">
      <Representation bandwidth="256" codecs="avc1.640028" id="0"></Representation>
    </AdaptationSet>
    <AdaptationSet id="2" lang="en" contentType="audio">
      <Representation bandwidth="256" codecs="mp4a.40.2" id="0"></Representation>
    </AdaptationSet>
    <AdaptationSet id="3" lang="en" contentType="text">
      <Representation bandwidth="256" codecs="wvtt" id="0"></Representation>
    </AdaptationSet>
  </Period>
</MPD>
`

	manifestWithoutHDR10 := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT6M16S" minBufferTime="PT1.97S">
  <BaseURL>http://existing.base/url/</BaseURL>
  <Period>
    <AdaptationSet id="0" lang="en" contentType="video">
      <Representation bandwidth="256" codecs="hvc1.1.4.L120.90" id="4"></Representation>
      <Representation bandwidth="256" codecs="hvc1.1.4.L63.90" id="5"></Representation>
      <Representation bandwidth="256" codecs="hev1.1.4.L120.90" id="6"></Representation>
      <Representation bandwidth="256" codecs="hev1.1.4.L63.90" id="7"></Representation>
      <Representation bandwidth="256" codecs="hev1.3.4.L63.90" id="9"></Representation>
    </AdaptationSet>
    <AdaptationSet id="1" lang="en" contentType="video">
      <Representation bandwidth="256" codecs="dvh1.05.01" id="0"></Representation>
      <Representation bandwidth="256" codecs="dvh1.05.03" id="1"></Representation>
    </AdaptationSet>
    <AdaptationSet id="2" lang="en" contentType="video">
      <Representation bandwidth="256" codecs="avc1.640028" id="0"></Representation>
    </AdaptationSet>
    <AdaptationSet id="3" lang="en" contentType="audio">
      <Representation bandwidth="256" codecs="mp4a.40.2" id="0"></Representation>
    </AdaptationSet>
    <AdaptationSet id="4" lang="en" contentType="text">
      <Representation bandwidth="256" codecs="wvtt" id="0"></Representation>
    </AdaptationSet>
  </Period>
</MPD>
`

	manifestWithoutVideo := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT6M16S" minBufferTime="PT1.97S">
  <BaseURL>http://existing.base/url/</BaseURL>
  <Period>
    <AdaptationSet id="0" lang="en" contentType="audio">
      <Representation bandwidth="256" codecs="mp4a.40.2" id="0"></Representation>
    </AdaptationSet>
    <AdaptationSet id="1" lang="en" contentType="text">
      <Representation bandwidth="256" codecs="wvtt" id="0"></Representation>
    </AdaptationSet>
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
			name: "when all video codecs are supplied, all video is stripped from a manifest",
			filters: &parsers.MediaFilters{
				Videos: parsers.NestedFilters{
					Codecs: []parsers.Codec{"hvc", "hev", "avc", "dvh"},
				},
			},
			manifestContent:       manifestWithMultiVideoCodec,
			expectManifestContent: manifestWithoutVideo,
		},
		{
			name: "when a video filter is supplied with HEVC and AVC, HEVC and AVC is stripped from manifest",
			filters: &parsers.MediaFilters{
				Videos: parsers.NestedFilters{
					Codecs: []parsers.Codec{"hvc", "hev", "avc"},
				},
			},
			manifestContent:       manifestWithMultiVideoCodec,
			expectManifestContent: manifestWithoutHEVCAndAVCVideoCodec,
		},
		{
			name: "when a video filter is suplied with Dolby Vision ID, dolby vision is stripped from manifest",
			filters: &parsers.MediaFilters{
				Videos: parsers.NestedFilters{
					Codecs: []parsers.Codec{"dvh"},
				},
			},
			manifestContent:       manifestWithMultiVideoCodec,
			expectManifestContent: manifestWithoutDolbyVisionCodec,
		},
		{
			name: "when a video filter is suplied with HEVC ID, HEVC is stripped from manifest",
			filters: &parsers.MediaFilters{
				Videos: parsers.NestedFilters{
					Codecs: []parsers.Codec{"hvc", "hev"},
				},
			},
			manifestContent:       manifestWithMultiVideoCodec,
			expectManifestContent: manifestWithoutHEVCVideoCodec,
		},
		{
			name: "when a video filter is suplied with AVC, AVC is stripped from manifest",
			filters: &parsers.MediaFilters{
				Videos: parsers.NestedFilters{
					Codecs: []parsers.Codec{"avc"},
				},
			},
			manifestContent:       manifestWithMultiVideoCodec,
			expectManifestContent: manifestWithoutAVCVideoCodec,
		},
		{
			name: "when a video filter is suplied with HDR10, all hevc main10 profiles are stripped from manifest",
			filters: &parsers.MediaFilters{
				Videos: parsers.NestedFilters{
					Codecs: []parsers.Codec{"hvc1.2", "hev1.2"},
				},
			},
			manifestContent:       manifestWithMultiVideoCodec,
			expectManifestContent: manifestWithoutHDR10,
		},
		{
			name:                  "when no video filters are supplied, nothing is stripped from manifest",
			filters:               &parsers.MediaFilters{},
			manifestContent:       manifestWithMultiVideoCodec,
			expectManifestContent: manifestWithMultiVideoCodec,
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

func TestDASHFilter_FilterManifest_audioCodecs(t *testing.T) {
	manifestWithEAC3AndAC3AudioCodec := `<?xml version="1.0" encoding="UTF-8"?>
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

	manifestWithoutAC3AudioCodec := `<?xml version="1.0" encoding="UTF-8"?>
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

	manifestWithoutEAC3AudioCodec := `<?xml version="1.0" encoding="UTF-8"?>
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
			name: "when all codecs are applied, audio is stripped from a manifest",
			filters: &parsers.MediaFilters{
				Audios: parsers.NestedFilters{
					Codecs: []parsers.Codec{"ac-3", "ec-3"},
				},
			},
			manifestContent:       manifestWithEAC3AndAC3AudioCodec,
			expectManifestContent: manifestWithoutAudio,
		},
		{
			name: "when an audio filter is supplied with Enhanced AC-3 codec, Enhanced AC-3 is stripped out",
			filters: &parsers.MediaFilters{
				Audios: parsers.NestedFilters{
					Codecs: []parsers.Codec{"ec-3"},
				},
			},
			manifestContent:       manifestWithEAC3AndAC3AudioCodec,
			expectManifestContent: manifestWithoutEAC3AudioCodec,
		},
		{
			name: "when an audio filter is supplied with AC-3 codec, AC-3 is stripped out",
			filters: &parsers.MediaFilters{
				Audios: parsers.NestedFilters{
					Codecs: []parsers.Codec{"ac-3"},
				},
			},
			manifestContent:       manifestWithEAC3AndAC3AudioCodec,
			expectManifestContent: manifestWithoutAC3AudioCodec,
		},
		{
			name:                  "when no audio filters are supplied, nothing is stripped from manifest",
			filters:               &parsers.MediaFilters{},
			manifestContent:       manifestWithEAC3AndAC3AudioCodec,
			expectManifestContent: manifestWithEAC3AndAC3AudioCodec,
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

	manifestWithoutSTPPCaptions := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT6M16S" minBufferTime="PT1.97S">
  <BaseURL>http://existing.base/url/</BaseURL>
  <Period>
    <AdaptationSet id="0" lang="en" contentType="text">
      <Representation bandwidth="256" codecs="wvtt" id="subtitle_en"></Representation>
    </AdaptationSet>
  </Period>
</MPD>
`

	manifestWithoutWVTTCaptions := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT6M16S" minBufferTime="PT1.97S">
  <BaseURL>http://existing.base/url/</BaseURL>
  <Period>
    <AdaptationSet id="0" lang="en" contentType="text">
      <Representation bandwidth="256" codecs="stpp" id="subtitle_en_ttml"></Representation>
    </AdaptationSet>
  </Period>
</MPD>
`

	manifestWithoutCaptions := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT6M16S" minBufferTime="PT1.97S">
  <BaseURL>http://existing.base/url/</BaseURL>
  <Period></Period>
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
			name: "when all caption types are supplied, captions are stripped from a " +
				"manifest",
			filters: &parsers.MediaFilters{
				Captions: parsers.NestedFilters{
					Codecs: []parsers.Codec{"stpp", "wvtt"},
				},
			},
			manifestContent:       manifestWithWVTTAndSTPPCaptions,
			expectManifestContent: manifestWithoutCaptions,
		},
		{
			name: "when a caption type filter is supplied with stpp only, webvtt captions are " +
				"filtered out",
			filters: &parsers.MediaFilters{
				Captions: parsers.NestedFilters{
					Codecs: []parsers.Codec{"stpp"},
				},
			},
			manifestContent:       manifestWithWVTTAndSTPPCaptions,
			expectManifestContent: manifestWithoutSTPPCaptions,
		},
		{
			name: "when a caption type filter is supplied with wvtt only, stpp captions are " +
				"filtered out",
			filters: &parsers.MediaFilters{
				Captions: parsers.NestedFilters{
					Codecs: []parsers.Codec{"wvtt"},
				},
			},
			manifestContent:       manifestWithWVTTAndSTPPCaptions,
			expectManifestContent: manifestWithoutWVTTCaptions,
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
			name: "when video streams are filtered, the manifest contains no video adaptation sets",
			filters: &parsers.MediaFilters{
				ContentTypes: []parsers.ContentType{"video"},
			},
			manifestContent:       manifestWithAudioAndVideoStreams,
			expectManifestContent: manifestWithOnlyAudioStreams,
		},
		{
			name: "when audio streams are filtered, the manifest contains no audio adaptation sets",
			filters: &parsers.MediaFilters{
				ContentTypes: []parsers.ContentType{"audio"},
			},
			manifestContent:       manifestWithAudioAndVideoStreams,
			expectManifestContent: manifestWithOnlyVideoStreams,
		},
		{
			name: "when audio and video streams are filtered, the manifest contains no audio or " +
				"video adaptation sets",
			filters: &parsers.MediaFilters{
				ContentTypes: []parsers.ContentType{"video", "audio"},
			},
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

func TestDASHFilter_FilterManifest_bitrate(t *testing.T) {
	baseManifest := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT6M16S" minBufferTime="PT1.97S">
  <BaseURL>http://existing.base/url/</BaseURL>
  <Period>
    <AdaptationSet id="0" lang="en" contentType="video">
      <Representation bandwidth="2048" codecs="avc" id="0"></Representation>
      <Representation bandwidth="4096" codecs="avc" id="1"></Representation>
    </AdaptationSet>
    <AdaptationSet id="1" lang="en" contentType="audio">
      <Representation bandwidth="256" codecs="ac-3" id="0"></Representation>
      <Representation bandwidth="100" codecs="ec-3" id="1"></Representation>
    </AdaptationSet>
  </Period>
</MPD>
`

	manifestFiltering256And2048Representations := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT6M16S" minBufferTime="PT1.97S">
  <BaseURL>http://existing.base/url/</BaseURL>
  <Period>
    <AdaptationSet id="0" lang="en" contentType="video">
      <Representation bandwidth="4096" codecs="avc" id="1"></Representation>
    </AdaptationSet>
  </Period>
</MPD>
`

	manifestFiltering2048Representation := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT6M16S" minBufferTime="PT1.97S">
  <BaseURL>http://existing.base/url/</BaseURL>
  <Period>
    <AdaptationSet id="0" lang="en" contentType="video">
      <Representation bandwidth="4096" codecs="avc" id="1"></Representation>
    </AdaptationSet>
    <AdaptationSet id="1" lang="en" contentType="audio">
      <Representation bandwidth="256" codecs="ac-3" id="0"></Representation>
      <Representation bandwidth="100" codecs="ec-3" id="1"></Representation>
    </AdaptationSet>
  </Period>
</MPD>
`

	manifestFiltering4096Representation := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT6M16S" minBufferTime="PT1.97S">
  <BaseURL>http://existing.base/url/</BaseURL>
  <Period>
    <AdaptationSet id="0" lang="en" contentType="video">
      <Representation bandwidth="2048" codecs="avc" id="0"></Representation>
    </AdaptationSet>
    <AdaptationSet id="1" lang="en" contentType="audio">
      <Representation bandwidth="256" codecs="ac-3" id="0"></Representation>
      <Representation bandwidth="100" codecs="ec-3" id="1"></Representation>
    </AdaptationSet>
  </Period>
</MPD>
`

	manifestFiltering2048And4096Representations := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT6M16S" minBufferTime="PT1.97S">
  <BaseURL>http://existing.base/url/</BaseURL>
  <Period>
    <AdaptationSet id="0" lang="en" contentType="audio">
      <Representation bandwidth="256" codecs="ac-3" id="0"></Representation>
      <Representation bandwidth="100" codecs="ec-3" id="1"></Representation>
    </AdaptationSet>
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
			name:                  "when no filters are given, nothing is stripped from manifest",
			filters:               &parsers.MediaFilters{},
			manifestContent:       baseManifest,
			expectManifestContent: baseManifest,
		},
		{
			name: "when hitting lower boundary (minBitrate = 0), expect results to be filtered",
			filters: &parsers.MediaFilters{
				Videos: parsers.NestedFilters{
					Bitrate: &parsers.Bitrate{
						Min: 0,
						Max: 4000,
					},
				},
				Audios: parsers.NestedFilters{
					Bitrate: &parsers.Bitrate{
						Min: 0,
						Max: 4000,
					},
				},
			},
			manifestContent:       baseManifest,
			expectManifestContent: manifestFiltering4096Representation,
		},
		{
			name: "when hitting upper bounary (maxBitrate = math.MaxInt32), expect results to be filtered",
			filters: &parsers.MediaFilters{
				Videos: parsers.NestedFilters{
					Bitrate: &parsers.Bitrate{
						Min: 4000,
						Max: math.MaxInt32,
					},
				},
				Audios: parsers.NestedFilters{
					Bitrate: &parsers.Bitrate{
						Min: 4000,
						Max: math.MaxInt32,
					},
				},
			},
			manifestContent:       baseManifest,
			expectManifestContent: manifestFiltering256And2048Representations,
		},
		{
			name: "when valid input, expect filtered results with no adaptation sets removed",
			filters: &parsers.MediaFilters{
				Videos: parsers.NestedFilters{
					Bitrate: &parsers.Bitrate{
						Min: 10,
						Max: 4000,
					},
				},
				Audios: parsers.NestedFilters{
					Bitrate: &parsers.Bitrate{
						Min: 10,
						Max: 4000,
					},
				},
			},
			manifestContent:       baseManifest,
			expectManifestContent: manifestFiltering4096Representation,
		},
		{
			name: "when valid input, expect filtered results with one adaptation set removed",
			filters: &parsers.MediaFilters{
				Videos: parsers.NestedFilters{
					Bitrate: &parsers.Bitrate{
						Min: 100,
						Max: 1000,
					},
				},
				Audios: parsers.NestedFilters{
					Bitrate: &parsers.Bitrate{
						Min: 100,
						Max: 1000,
					},
				},
			},
			manifestContent:       baseManifest,
			expectManifestContent: manifestFiltering2048And4096Representations,
		},
		{
			name: "when filtering a valid bitrate range in video only, expect filtered results",
			filters: &parsers.MediaFilters{
				Videos: parsers.NestedFilters{
					Bitrate: &parsers.Bitrate{
						Min: 10,
						Max: 3000,
					},
				},
			},
			manifestContent:       baseManifest,
			expectManifestContent: manifestFiltering4096Representation,
		},
		{
			name: "when filtering a valid video bitrate range touching upper bound (maxBitrate = math.MaxInt32), expect results to be filtered",
			filters: &parsers.MediaFilters{
				Videos: parsers.NestedFilters{
					Bitrate: &parsers.Bitrate{
						Min: 3000,
						Max: math.MaxInt32,
					},
				},
			},
			manifestContent:       baseManifest,
			expectManifestContent: manifestFiltering2048Representation,
		},
		{
			name: "when both audio and video filters are given, expect both to be filtered accordingly",
			filters: &parsers.MediaFilters{
				Videos: parsers.NestedFilters{
					Bitrate: &parsers.Bitrate{
						Min: 0,
						Max: 4000,
					},
				},
				Audios: parsers.NestedFilters{
					Bitrate: &parsers.Bitrate{
						Min: 0,
						Max: 3000,
					},
				},
			},
			manifestContent:       baseManifest,
			expectManifestContent: manifestFiltering4096Representation,
		},
		{
			name: "when both audio and video bitrates hit both bounds, expect no filtering ",
			filters: &parsers.MediaFilters{
				Videos: parsers.NestedFilters{
					Bitrate: &parsers.Bitrate{
						Min: 0,
						Max: math.MaxInt32,
					},
				},
				Audios: parsers.NestedFilters{
					Bitrate: &parsers.Bitrate{
						Min: 0,
						Max: math.MaxInt32,
					},
				},
			},
			manifestContent:       baseManifest,
			expectManifestContent: baseManifest,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			filter := NewDASHFilter("", tt.manifestContent, config.Config{})

			manifest, err := filter.FilterManifest(tt.filters)
			if err != nil && !tt.expectErr {
				t.Errorf("FilterManifest() didn't expect error to be returned, got: %v", err)
				return
			} else if err == nil && tt.expectErr {
				t.Error("FilterManifest() expected an error, got nil")
				return
			}

			if g, e := manifest, tt.expectManifestContent; g != e {
				t.Fatalf("FilterManifest() returned wrong manifest\ngot %v\nexpected %v\ndiff: %v", g, e, cmp.Diff(g, e))
			}
		})
	}
}

func TestDASHFilter_FilterManifest_LanguageFilter(t *testing.T) {
	manifestWithMultiLanguages := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT6M16S" minBufferTime="PT1.97S">
  <BaseURL>http://existing.base/url/</BaseURL>
  <Period>
    <AdaptationSet id="0" lang="en" contentType="video">
      <Representation bandwidth="256" codecs="hvc1.2.4.L93.90" id="0"></Representation>
      <Representation bandwidth="256" codecs="hvc1.2.4.L90.90" id="1"></Representation>
    </AdaptationSet>
    <AdaptationSet id="1" lang="en" contentType="audio">
      <Representation bandwidth="256" codecs="mp4a.40.2" id="0"></Representation>
    </AdaptationSet>
    <AdaptationSet id="2" lang="en" contentType="text">
      <Representation bandwidth="256" codecs="wvtt" id="0"></Representation>
    </AdaptationSet>
    <AdaptationSet id="3" lang="es" contentType="audio">
      <Representation bandwidth="256" codecs="mp4a.40.2" id="0"></Representation>
    </AdaptationSet>
    <AdaptationSet id="4" lang="es" contentType="text">
      <Representation bandwidth="256" codecs="wvtt" id="0"></Representation>
    </AdaptationSet>
    <AdaptationSet id="5" lang="pt" contentType="audio">
      <Representation bandwidth="256" codecs="mp4a.40.2" id="0"></Representation>
    </AdaptationSet>
    <AdaptationSet id="6" lang="pt" contentType="text">
      <Representation bandwidth="256" codecs="wvtt" id="0"></Representation>
    </AdaptationSet>
  </Period>
</MPD>
`

	manifestWithNoSpanish := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT6M16S" minBufferTime="PT1.97S">
  <BaseURL>http://existing.base/url/</BaseURL>
  <Period>
    <AdaptationSet id="0" lang="en" contentType="video">
      <Representation bandwidth="256" codecs="hvc1.2.4.L93.90" id="0"></Representation>
      <Representation bandwidth="256" codecs="hvc1.2.4.L90.90" id="1"></Representation>
    </AdaptationSet>
    <AdaptationSet id="1" lang="en" contentType="audio">
      <Representation bandwidth="256" codecs="mp4a.40.2" id="0"></Representation>
    </AdaptationSet>
    <AdaptationSet id="2" lang="en" contentType="text">
      <Representation bandwidth="256" codecs="wvtt" id="0"></Representation>
    </AdaptationSet>
    <AdaptationSet id="3" lang="pt" contentType="audio">
      <Representation bandwidth="256" codecs="mp4a.40.2" id="0"></Representation>
    </AdaptationSet>
    <AdaptationSet id="4" lang="pt" contentType="text">
      <Representation bandwidth="256" codecs="wvtt" id="0"></Representation>
    </AdaptationSet>
  </Period>
</MPD>
`

	manifestWithNoSpanishAndPortugese := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT6M16S" minBufferTime="PT1.97S">
  <BaseURL>http://existing.base/url/</BaseURL>
  <Period>
    <AdaptationSet id="0" lang="en" contentType="video">
      <Representation bandwidth="256" codecs="hvc1.2.4.L93.90" id="0"></Representation>
      <Representation bandwidth="256" codecs="hvc1.2.4.L90.90" id="1"></Representation>
    </AdaptationSet>
    <AdaptationSet id="1" lang="en" contentType="audio">
      <Representation bandwidth="256" codecs="mp4a.40.2" id="0"></Representation>
    </AdaptationSet>
    <AdaptationSet id="2" lang="en" contentType="text">
      <Representation bandwidth="256" codecs="wvtt" id="0"></Representation>
    </AdaptationSet>
  </Period>
</MPD>
`

	manifestWithNoCaptions := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT6M16S" minBufferTime="PT1.97S">
  <BaseURL>http://existing.base/url/</BaseURL>
  <Period>
    <AdaptationSet id="0" lang="en" contentType="video">
      <Representation bandwidth="256" codecs="hvc1.2.4.L93.90" id="0"></Representation>
      <Representation bandwidth="256" codecs="hvc1.2.4.L90.90" id="1"></Representation>
    </AdaptationSet>
    <AdaptationSet id="1" lang="en" contentType="audio">
      <Representation bandwidth="256" codecs="mp4a.40.2" id="0"></Representation>
    </AdaptationSet>
    <AdaptationSet id="2" lang="es" contentType="audio">
      <Representation bandwidth="256" codecs="mp4a.40.2" id="0"></Representation>
    </AdaptationSet>
    <AdaptationSet id="3" lang="pt" contentType="audio">
      <Representation bandwidth="256" codecs="mp4a.40.2" id="0"></Representation>
    </AdaptationSet>
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
			name:                  "when no filters are set, nothing is stripped from manifest",
			filters:               &parsers.MediaFilters{},
			manifestContent:       manifestWithMultiLanguages,
			expectManifestContent: manifestWithMultiLanguages,
		},
		{
			name: "when es lang is set, adaptation sets with es are stripped from manifest",
			filters: &parsers.MediaFilters{
				Audios: parsers.NestedFilters{
					Language: []parsers.Language{"es"},
				},
				Captions: parsers.NestedFilters{
					Language: []parsers.Language{"es"},
				},
			},
			manifestContent:       manifestWithMultiLanguages,
			expectManifestContent: manifestWithNoSpanish,
		},
		{
			name: "when es and pt lang is set, adaptation sets with es and pt are stripped from manifest",
			filters: &parsers.MediaFilters{
				Audios: parsers.NestedFilters{
					Language: []parsers.Language{"es", "pt"},
				},
				Captions: parsers.NestedFilters{
					Language: []parsers.Language{"es", "pt"},
				},
			},
			manifestContent:       manifestWithMultiLanguages,
			expectManifestContent: manifestWithNoSpanishAndPortugese,
		},
		{
			name: "when es, pt, and en caption filters are set, expect those captions to be removed",
			filters: &parsers.MediaFilters{
				Captions: parsers.NestedFilters{
					Language: []parsers.Language{"es", "pt", "en"},
				},
			},
			manifestContent:       manifestWithMultiLanguages,
			expectManifestContent: manifestWithNoCaptions,
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

func TestDASHFilter_FilterRole_OverwriteValue(t *testing.T) {
	manifestWithAccessibilityElement := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT6M16S" minBufferTime="PT1.97S">
  <BaseURL>http://existing.base/url/</BaseURL>
  <Period>
    <AdaptationSet id="7357" lang="en" contentType="audio">
      <Role schemeIdUri="urn:mpeg:dash:role:2011" value="alternate"></Role>
      <Representation bandwidth="256" codecs="ac-3" id="1"></Representation>
      <Accessibility schemeIdUri="urn:tva:metadata:cs:AudioPurposeCS:2007" value="1"></Accessibility>
    </AdaptationSet>
  </Period>
</MPD>
`

	manifestWithoutAccessibilityElement := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT6M16S" minBufferTime="PT1.97S">
  <BaseURL>http://existing.base/url/</BaseURL>
  <Period>
    <AdaptationSet id="7357" lang="en" contentType="audio">
      <Role schemeIdUri="urn:mpeg:dash:role:2011" value="alternate"></Role>
      <Representation bandwidth="256" codecs="ac-3" id="1"></Representation>
    </AdaptationSet>
  </Period>
</MPD>
`

	manifestWithOverwrittenRoleValue := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT6M16S" minBufferTime="PT1.97S">
  <BaseURL>http://existing.base/url/</BaseURL>
  <Period>
    <AdaptationSet id="7357" lang="en" contentType="audio">
      <Role schemeIdUri="urn:mpeg:dash:role:2011" value="description"></Role>
      <Representation bandwidth="256" codecs="ac-3" id="1"></Representation>
      <Accessibility schemeIdUri="urn:tva:metadata:cs:AudioPurposeCS:2007" value="1"></Accessibility>
    </AdaptationSet>
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
			name: "when proper value is set and manifest has accessibility element, role value is overwritten.",
			filters: &parsers.MediaFilters{
				Plugins: []string{"dvsRoleOverride"},
			},
			manifestContent:       manifestWithAccessibilityElement,
			expectManifestContent: manifestWithOverwrittenRoleValue,
		},
		{
			name: "when proper value is set but no accessibility element is found, role value is not overwritten.",
			filters: &parsers.MediaFilters{
				Plugins: []string{"dvsRoleOverride"},
			},
			manifestContent:       manifestWithoutAccessibilityElement,
			expectManifestContent: manifestWithoutAccessibilityElement,
		},
		{
			name: "when proper value is not set and manifest has accessibility element, role value is not overwritten.",
			filters: &parsers.MediaFilters{
				Plugins: []string{},
			},
			manifestContent:       manifestWithAccessibilityElement,
			expectManifestContent: manifestWithAccessibilityElement,
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
