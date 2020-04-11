package filters

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/cbsinteractive/bakery/config"
	"github.com/cbsinteractive/bakery/parsers"
	"github.com/grafov/m3u8"
)

type execPluginHLS func(variant *m3u8.Variant)

// HLSFilter implements the Filter interface for HLS
// manifests
type HLSFilter struct {
	manifestURL     string
	manifestContent string
	maxSegmentSize  float64
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

// GetMaxAge returns max_age to be overwritten via cache control headers
func (h *HLSFilter) GetMaxAge() string {
	return fmt.Sprintf("%.0f", h.maxSegmentSize/2)
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
	filteredManifest.Twitch = manifest.Twitch

	for _, v := range manifest.Variants {
		if filters.SuppressIFrame() && v.Iframe {
			continue
		}

		absolute, aErr := getAbsoluteURL(h.manifestURL)
		if aErr != nil {
			return h.manifestContent, aErr
		}

		normalizedVariant, err := h.normalizeVariant(v, *absolute)
		if err != nil {
			return "", err
		}

		filteredVariants, err := h.filterVariants(filters, normalizedVariant)
		if err != nil {
			return "", err
		}

		if filteredVariants {
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
func (h *HLSFilter) filterVariants(filters *parsers.MediaFilters, v *m3u8.Variant) (bool, error) {
	variantCodecs := strings.Split(v.Codecs, ",")

	if filters.Videos.Bitrate != nil || filters.Audios.Bitrate != nil {
		if h.filterVariantBandwidth(int(v.VariantParams.Bandwidth), variantCodecs, filters) {
			return true, nil
		}
	}

	if filters.Videos.Codecs != nil {
		supportedVideoTypes := map[string]struct{}{}
		for _, vt := range filters.Videos.Codecs {
			supportedVideoTypes[string(vt)] = struct{}{}
		}
		res, err := filterVariantCodecs(videoContentType, variantCodecs, supportedVideoTypes, matchFunctions)
		if res {
			return true, err
		}
	}

	if filters.Audios.Codecs != nil {
		supportedAudioTypes := map[string]struct{}{}
		for _, at := range filters.Audios.Codecs {
			supportedAudioTypes[string(at)] = struct{}{}
		}
		res, err := filterVariantCodecs(audioContentType, variantCodecs, supportedAudioTypes, matchFunctions)
		if res {
			return true, err
		}
	}

	if filters.Captions.Codecs != nil {
		supportedCaptions := map[string]struct{}{}
		for _, ct := range filters.Captions.Codecs {
			supportedCaptions[string(ct)] = struct{}{}
		}
		res, err := filterVariantCodecs(captionContentType, variantCodecs, supportedCaptions, matchFunctions)
		if res {
			return true, err
		}
	}

	if filters.FrameRate != nil {
		if filterVariantFrameRate(v.FrameRate, filters.FrameRate) {
			return true, nil
		}
	}

	// This filter should run last as it is not removing variants, rather updating the alternatives attached to
	// the variant. This function will only execute if no matches have been found
	if filters.Audios.Language != nil || filters.Captions.Language != nil {
		h.filterVariantLanguage(v, filters)
	}

	return false, nil
}

// Returns true if the provided variant is out of range since filters are removed when true.
func (h *HLSFilter) filterVariantBandwidth(b int, variantCodecs []string, filters *parsers.MediaFilters) bool {
	for _, codec := range variantCodecs {
		var min, max int

		switch {
		case isAudioCodec(codec):
			if filters.Audios.Bitrate == nil {
				continue
			}

			min = filters.Audios.Bitrate.Min
			max = filters.Audios.Bitrate.Max
		case isVideoCodec(codec):
			if filters.Videos.Bitrate == nil {
				continue
			}

			min = filters.Videos.Bitrate.Min
			max = filters.Videos.Bitrate.Max
		default:
			continue
		}

		if !inRange(min, max, b) {
			return true
		}
	}

	return false
}

// Returns true if the given variant (variantCodecs) should be filtered out
func filterVariantCodecs(filterType ContentType, variantCodecs []string, supportedCodecs map[string]struct{}, supportedFilterTypes map[ContentType]func(string) bool) (bool, error) {
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

func filterVariantFrameRate(floatFPS float64, frameRates []parsers.FPS) bool {
	strFPS := fmt.Sprintf("%.3f", floatFPS)

	for _, fr := range frameRates {
		if strFPS == string(fr) {
			return true
		}
	}

	return false
}

// Returns true if a given variant matches the provided language filter
func (h *HLSFilter) filterVariantLanguage(v *m3u8.Variant, filters *parsers.MediaFilters) {
	if v.Alternatives == nil {
		return
	}

	match := func(alt *m3u8.Alternative, langs []parsers.Language) bool {
		if langs == nil {
			return false
		}

		for _, lang := range langs {
			if strings.EqualFold(string(lang), alt.Language) {
				return true
			}
		}

		return false
	}

	var alts []*m3u8.Alternative
	var groupIDs = map[string]struct{}{}
	for _, alt := range v.Alternatives {
		remove := true
		switch alt.Type {
		case "AUDIO":
			remove = match(alt, filters.Audios.Language)
		case "SUBTITLES":
			remove = match(alt, filters.Captions.Language)
		case "CLOSED-CAPTIONS":
			remove = match(alt, filters.Captions.Language)
		}

		if !remove {
			alts = append(alts, alt)
			groupIDs[alt.GroupId] = struct{}{}
		}

	}

	v.Alternatives = alts
	if _, audio := groupIDs[v.Audio]; !audio {
		v.Audio = ""
	}
	if _, subs := groupIDs[v.Subtitles]; !subs {
		v.Subtitles = ""
	}
	if _, captions := groupIDs[v.Captions]; !captions {
		v.Captions = ""
	}
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

	if h.config.IsLocalHost() {
		return fmt.Sprintf("http://%v%v/t(%v,%v)/%v.m3u8", h.config.Hostname, h.config.Listen, start, end, encoded), nil
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

	// Append mode will be set to true when first segment is encountered in range.
	// Once true, we can append segments with tags that don't normally carry PDT
	// EX: #EXT-X-ASSET, #EXT-OATCLS-SCTE35, or any other custom tags advertised in playlist
	var append bool
	var maxSize float64
	for _, segment := range m.Segments {
		if segment == nil {
			continue
		}

		if filters.SuppressAds() && segment.SCTE != nil {
			segment.SCTE = nil
		}

		if segment.ProgramDateTime == (time.Time{}) && append {
			if err := appendSegment(h.manifestURL, segment, filteredPlaylist); err != nil {
				return "", fmt.Errorf("trimming segments: %w", err)
			}
			continue
		}

		append = inRange(filters.Trim.Start, filters.Trim.End, int(segment.ProgramDateTime.Unix()))

		if append {
			if err := appendSegment(h.manifestURL, segment, filteredPlaylist); err != nil {
				return "", fmt.Errorf("trimming segments: %w", err)
			}
		}

		if maxSize < segment.Duration && append {
			maxSize = segment.Duration
		}
	}

	h.maxSegmentSize = maxSize
	filteredPlaylist.Close()

	return isEmpty(filteredPlaylist.Encode().String())
}

func isEmpty(p string) (string, error) {
	emptyPlaylist := fmt.Sprintf("%v\n%v\n%v\n%v\n%v\n",
		"#EXTM3U",
		"#EXT-X-VERSION:3",
		"#EXT-X-MEDIA-SEQUENCE:0",
		"#EXT-X-TARGETDURATION:0",
		"#EXT-X-ENDLIST",
	)

	var err error
	if emptyPlaylist == p {
		err = fmt.Errorf("No segments found in range. Is PDT set?")
	}

	return p, err

}

//appends segment to provided media playlist with absolute urls
func appendSegment(manifest string, s *m3u8.MediaSegment, p *m3u8.MediaPlaylist) error {
	absolute, err := getAbsoluteURL(manifest)
	if err != nil {
		return fmt.Errorf("formatting segment URLs: %w", err)
	}

	s.URI, err = combinedIfRelative(s.URI, *absolute)
	if err != nil {
		return fmt.Errorf("formatting segment URLs: %w", err)
	}

	err = p.AppendSegment(s)
	if err != nil {
		return fmt.Errorf("trimming segments: %w", err)
	}

	return nil
}

//Returns absolute url of given manifest as a string
func getAbsoluteURL(path string) (*url.URL, error) {
	absoluteURL, _ := filepath.Split(path)
	return url.Parse(absoluteURL)
}
