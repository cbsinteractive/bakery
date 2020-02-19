package filters

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/cbsinteractive/bakery/pkg/config"
	"github.com/cbsinteractive/bakery/pkg/parsers"
	"github.com/grafov/m3u8"
)

// HLSFilter implements the Filter interface for HLS
// manifests
type HLSFilter struct {
	manifestURL     string
	manifestContent string
	config          config.Config
}

// NewHLSFilter is the HLS filter constructor
func NewHLSFilter(manifestURL, manifestContent string, c config.Config) *HLSFilter {
	return &HLSFilter{
		manifestURL:     manifestURL,
		manifestContent: manifestContent,
		config:          c,
	}
}

// FilterManifest will be responsible for filtering the manifest
// according  to the MediaFilters
func (h *HLSFilter) FilterManifest(filters *parsers.MediaFilters) (string, error) {
	m, manifestType, err := m3u8.DecodeFrom(strings.NewReader(h.manifestContent), true)
	if err != nil {
		return "", err
	}

	if manifestType != m3u8.MASTER {
		return "", errors.New("manifest type is wrong")
	}

	// convert into the master playlist type
	manifest := m.(*m3u8.MasterPlaylist)
	filteredManifest := m3u8.NewMasterPlaylist()

	for _, v := range manifest.Variants {
		absoluteURL, _ := filepath.Split(h.manifestURL)

		normalizedVariant := h.normalizeVariant(v, absoluteURL)
		if h.validateVariants(filters, normalizedVariant) {
			filteredManifest.Append(normalizedVariant.URI, normalizedVariant.Chunklist, normalizedVariant.VariantParams)
		}
	}

	return filteredManifest.String(), nil
}

// Returns true if specified variant passes all filters
func (h *HLSFilter) validateVariants(filters *parsers.MediaFilters, v *m3u8.Variant) bool {
	if filters.DefinesBitrateFilter() {
		if !(h.validateBandwidthVariant(filters.MinBitrate, filters.MaxBitrate, v)) {
			return false
		}
	}

	variantCodecs := strings.Split(v.Codecs, ",")

	if filters.Audios != nil {
		supportedAudioTypes := map[string]struct{}{}
		for _, at := range filters.Audios {
			supportedAudioTypes[string(at)] = struct{}{}
		}
		res := validateVariantCodecs(audioContentType, variantCodecs, supportedAudioTypes)
		if !res {
			return false
		}
	}

	if filters.Videos != nil {
		supportedVideoTypes := map[string]struct{}{}
		for _, vt := range filters.Videos {
			supportedVideoTypes[string(vt)] = struct{}{}
		}
		res := validateVariantCodecs(videoContentType, variantCodecs, supportedVideoTypes)
		if !res {
			return false
		}
	}

	if filters.CaptionTypes != nil {
		supportedCaptionTypes := map[string]struct{}{}
		for _, ct := range filters.CaptionTypes {
			supportedCaptionTypes[string(ct)] = struct{}{}
		}
		res := validateVariantCodecs(captionContentType, variantCodecs, supportedCaptionTypes)
		if !res {
			return false
		}
	}

	return true
}

// Returns true if the given variant (variantCodecs) should be allowed through the filter for supportedCodecs of filterType
func validateVariantCodecs(filterType ContentType, variantCodecs []string, supportedCodecs map[string]struct{}) bool {
	var matchFilterType func(string) bool
	typeInVariant := 0
	matchInVariant := 0

	switch {
	case filterType == audioContentType:
		matchFilterType = isAudioCodec
	case filterType == videoContentType:
		matchFilterType = isVideoCodec
	case filterType == captionContentType:
		matchFilterType = isCaptionCodec
	}

	for _, codec := range variantCodecs {
		if matchFilterType(codec) {
			typeInVariant++
			for sc := range supportedCodecs {
				if ValidCodecs(codec, CodecFilterID(sc)) {
					matchInVariant++
					break
				}
			}
		}
	}
	return typeInVariant == matchInVariant
}

func (h *HLSFilter) validateBandwidthVariant(minBitrate int, maxBitrate int, v *m3u8.Variant) bool {
	bw := int(v.VariantParams.Bandwidth)
	if bw > maxBitrate || bw < minBitrate {
		return false
	}

	return true
}

func (h *HLSFilter) normalizeVariant(v *m3u8.Variant, absoluteURL string) *m3u8.Variant {
	for _, a := range v.VariantParams.Alternatives {
		a.URI = absoluteURL + a.URI
	}

	v.URI = absoluteURL + v.URI

	return v
}

// Returns true if given codec is an audio codec (mp4a, ec-3, or ac-3)
func isAudioCodec(codec string) bool {
	return (ValidCodecs(codec, CodecFilterID("mp4a")) ||
		ValidCodecs(codec, CodecFilterID("ec-3")) ||
		ValidCodecs(codec, CodecFilterID("ac-3")))
}

// Returns true if given codec is a video codec (hvc, avc, or dvh)
func isVideoCodec(codec string) bool {
	return (ValidCodecs(codec, CodecFilterID("hvc")) ||
		ValidCodecs(codec, CodecFilterID("avc")) ||
		ValidCodecs(codec, CodecFilterID("dvh")))
}

// Returns true if goven codec is a caption codec (stpp or wvtt)
func isCaptionCodec(codec string) bool {
	return (ValidCodecs(codec, CodecFilterID("stpp")) ||
		ValidCodecs(codec, CodecFilterID("wvtt")))
}
