package filters

import (
	"errors"
	"net/url"
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

var matchFunctions = map[ContentType]func(string) bool{
	audioContentType:   isAudioCodec,
	videoContentType:   isVideoCodec,
	captionContentType: isCaptionCodec,
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
		absolute, aErr := url.Parse(absoluteURL)
		if aErr != nil {
			return h.manifestContent, aErr
		}
		normalizedVariant, err := h.normalizeVariant(v, *absolute)
		if err != nil {
			return "", err
		}
		validatedFilters, err := h.validateVariants(filters, normalizedVariant)
		if err != nil {
			return "", err
		}

		if validatedFilters {
			continue
		}

		filteredManifest.Append(normalizedVariant.URI, normalizedVariant.Chunklist, normalizedVariant.VariantParams)
	}

	return filteredManifest.String(), nil
}

// Returns true if specified variant passes all filters
func (h *HLSFilter) validateVariants(filters *parsers.MediaFilters, v *m3u8.Variant) (bool, error) {
	if filters.DefinesBitrateFilter() {
		if !(h.validateBandwidthVariant(filters.MinBitrate, filters.MaxBitrate, v)) {
			return true, nil
		}
	}

	variantCodecs := strings.Split(v.Codecs, ",")

	if filters.Audios != nil {
		supportedAudioTypes := map[string]struct{}{}
		for _, at := range filters.Audios {
			supportedAudioTypes[string(at)] = struct{}{}
		}
		res, err := validateVariantCodecs(audioContentType, variantCodecs, supportedAudioTypes, matchFunctions)
		if res {
			return true, err
		}
	}

	if filters.Videos != nil {
		supportedVideoTypes := map[string]struct{}{}
		for _, vt := range filters.Videos {
			supportedVideoTypes[string(vt)] = struct{}{}
		}
		res, err := validateVariantCodecs(videoContentType, variantCodecs, supportedVideoTypes, matchFunctions)
		if res {
			return true, err
		}
	}

	if filters.CaptionTypes != nil {
		supportedCaptionTypes := map[string]struct{}{}
		for _, ct := range filters.CaptionTypes {
			supportedCaptionTypes[string(ct)] = struct{}{}
		}
		res, err := validateVariantCodecs(captionContentType, variantCodecs, supportedCaptionTypes, matchFunctions)
		if res {
			return true, err
		}
	}

	return false, nil
}

// Returns true if the given variant (variantCodecs) should be allowed through the filter for supportedCodecs of filterType
func validateVariantCodecs(filterType ContentType, variantCodecs []string, supportedCodecs map[string]struct{}, supportedFilterTypes map[ContentType]func(string) bool) (bool, error) {
	var matchFilterType func(string) bool

	matchFilterType, found := supportedFilterTypes[filterType]

	if !found {
		return false, errors.New("filter type is unsupported")
	}

	variantFound := false
	for _, codec := range variantCodecs {
		if matchFilterType(codec) {
			for sc := range supportedCodecs {
				if ValidCodecs(codec, CodecFilterID(sc)) {
					variantFound = true
					break
				}
			}
		}
	}

	return variantFound, nil
}

func (h *HLSFilter) validateBandwidthVariant(minBitrate int, maxBitrate int, v *m3u8.Variant) bool {
	bw := int(v.VariantParams.Bandwidth)
	if bw > maxBitrate || bw < minBitrate {
		return false
	}

	return true
}

func (h *HLSFilter) normalizeVariant(v *m3u8.Variant, absolute url.URL) (*m3u8.Variant, error) {
	for _, a := range v.VariantParams.Alternatives {
		aURL, aErr := combinedIfRelative(a.URI, absolute)
		if aErr != nil {
			return v, aErr
		}
		a.URI = aURL
	}

	vURL, vErr := combinedIfRelative(v.URI, absolute)
	if vErr != nil {
		return v, vErr
	}
	v.URI = vURL
	return v, nil
}

func combinedIfRelative(uri string, absolute url.URL) (string, error) {
	if len(uri) == 0 {
		return uri, nil
	}
	relative, err := isRelative(uri)
	if err != nil {
		return uri, err
	}
	if relative {
		combined, err := absolute.Parse(uri)
		if err != nil {
			return uri, err
		}
		return combined.String(), err
	}
	return uri, nil
}

func isRelative(urlStr string) (bool, error) {
	u, e := url.Parse(urlStr)
	if e != nil {
		return false, e
	}
	return !u.IsAbs(), nil
}

// Returns true if given codec is an audio codec (mp4a, ec-3, or ac-3)
func isAudioCodec(codec string) bool {
	return (ValidCodecs(codec, aacCodec) ||
		ValidCodecs(codec, ec3Codec) ||
		ValidCodecs(codec, ac3Codec))
}

// Returns true if given codec is a video codec (hvc, avc, or dvh)
func isVideoCodec(codec string) bool {
	return (ValidCodecs(codec, hevcCodec) ||
		ValidCodecs(codec, avcCodec) ||
		ValidCodecs(codec, dolbyCodec))
}

// Returns true if goven codec is a caption codec (stpp or wvtt)
func isCaptionCodec(codec string) bool {
	return (ValidCodecs(codec, stppCodec) ||
		ValidCodecs(codec, wvttCodec))
}
