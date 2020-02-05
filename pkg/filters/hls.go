package filters

import (
	"errors"
	"math"
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

func (h *HLSFilter) validateVariants(filters *parsers.MediaFilters, v *m3u8.Variant) bool {

	if bandwidthInRange(filters.MinBitrate, filters.MaxBitrate) {

		if !(h.validateBandwidthVariant(filters.MinBitrate, filters.MaxBitrate, v)) {
			return false
		}
	}

	return true
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

func bandwidthInRange(minBandwidth int, maxBandwidth int) bool {
	if (minBandwidth >= 0 && maxBandwidth <= math.MaxInt32) &&
		(minBandwidth < maxBandwidth) &&
		!(minBandwidth == 0 && maxBandwidth == math.MaxInt32) {
		return true
	}

	return false
}
