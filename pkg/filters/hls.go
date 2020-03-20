package filters

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/cbsinteractive/bakery/pkg/config"
	"github.com/cbsinteractive/bakery/pkg/parsers"
	"github.com/grafov/m3u8"
)

type execPluginHLS func(variant *m3u8.Variant)

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
		return h.filterRenditionManifest(filters, m.(*m3u8.MediaPlaylist))
	}

	// convert into the master playlist type
	manifest := m.(*m3u8.MasterPlaylist)
	filteredManifest := m3u8.NewMasterPlaylist()

	for _, v := range manifest.Variants {
		absolute, aErr := getAbsoluteURL(h.manifestURL)
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

		uri := normalizedVariant.URI
		if filters.Trim != nil {
			uri, err = h.normalizeTrimmedVariant(filters, uri)
			if err != nil {
				return "", err
			}
		}

		filteredManifest.Append(uri, normalizedVariant.Chunklist, normalizedVariant.VariantParams)
	}

	return filteredManifest.String(), nil
}

// Returns true if specified variant should be removed from filter
func (h *HLSFilter) validateVariants(filters *parsers.MediaFilters, v *m3u8.Variant) (bool, error) {
	variantCodecs := strings.Split(v.Codecs, ",")

	if DefinesBitrateFilter(filters) {
		if !(h.validateBandwidthVariant(int(v.VariantParams.Bandwidth), variantCodecs, filters)) {
			return true, nil
		}
	}

	if filters.AudioFilters.Codecs != nil {
		supportedAudioTypes := map[string]struct{}{}
		for _, at := range filters.AudioFilters.Codecs {
			supportedAudioTypes[string(at)] = struct{}{}
		}
		res, err := validateVariantCodecs(audioContentType, variantCodecs, supportedAudioTypes, matchFunctions)
		if res {
			return true, err
		}
	}

	if filters.VideoFilters.Codecs != nil {
		supportedVideoTypes := map[string]struct{}{}
		for _, vt := range filters.VideoFilters.Codecs {
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

// Returns true if the given variant (variantCodecs) should be allowed filtered out for supportedCodecs of filterType
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

func (h *HLSFilter) validateBandwidthVariant(bw int, variantCodecs []string, filters *parsers.MediaFilters) bool {
	var lowerBitrate int
	var higherBitrate int
	for _, codec := range variantCodecs {
		audio := isAudioCodec(codec)
		video := isVideoCodec(codec)
		switch {
		case audio:
			lowerBitrate = filters.AudioFilters.MinBitrate
			higherBitrate = filters.AudioFilters.MaxBitrate
		case video:
			lowerBitrate = filters.VideoFilters.MinBitrate
			higherBitrate = filters.VideoFilters.MaxBitrate
		default:
			lowerBitrate = filters.MinBitrate
			higherBitrate = filters.MaxBitrate
		}
		if bw > higherBitrate || bw < lowerBitrate {
			return false
		}
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

func (h *HLSFilter) normalizeTrimmedVariant(filters *parsers.MediaFilters, uri string) (string, error) {
	encoded := base64.RawURLEncoding.EncodeToString([]byte(uri))
	start := filters.Trim.Start
	end := filters.Trim.End
	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%v://%v/t(%v,%v)/%v.m3u8", u.Scheme, h.config.Hostname, start, end, encoded), nil
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

// FilterRenditionManifest will be responsible for filtering the manifest
// according  to the MediaFilters
func (h *HLSFilter) filterRenditionManifest(filters *parsers.MediaFilters, m *m3u8.MediaPlaylist) (string, error) {
	filteredPlaylist, err := m3u8.NewMediaPlaylist(m.Count(), m.Count())
	if err != nil {
		return "", fmt.Errorf("filtering Rendition Manifest: %w", err)
	}

	for _, segment := range m.Segments {
		if segment == nil {
			continue
		}

		if segment.ProgramDateTime == (time.Time{}) {
			return "", fmt.Errorf("Program Date Time not set on segments")
		}

		if inRange(filters.Trim.Start, filters.Trim.End, segment.ProgramDateTime.Unix()) {
			absolute, err := getAbsoluteURL(h.manifestURL)
			if err != nil {
				return "", fmt.Errorf("formatting segment URLs: %w", err)
			}

			segment.URI, err = combinedIfRelative(segment.URI, *absolute)
			if err != nil {
				return "", fmt.Errorf("formatting segment URLs: %w", err)
			}

			err = filteredPlaylist.AppendSegment(segment)
			if err != nil {
				return "", fmt.Errorf("trimming segments: %w", err)
			}
		}
	}

	filteredPlaylist.Close()

	return filteredPlaylist.Encode().String(), nil
}

func inRange(start int64, end int64, value int64) bool {
	return (start <= value) && (value <= end)
}

//Returns absolute url of given manifest as a string
func getAbsoluteURL(path string) (*url.URL, error) {
	absoluteURL, _ := filepath.Split(path)
	return url.Parse(absoluteURL)
}
