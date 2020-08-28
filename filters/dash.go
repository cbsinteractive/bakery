package filters

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/cbsinteractive/bakery/config"
	"github.com/cbsinteractive/bakery/parsers"
	"github.com/zencoder/go-dash/mpd"
)

type execFilter func(filters *parsers.MediaFilters, manifest *mpd.MPD)

// DASHFilter implements the Filter interface for DASH manifests
type DASHFilter struct {
	manifestURL     string
	manifestContent string
	config          config.Config
}

// NewDASHFilter is the DASH filter constructor
func NewDASHFilter(manifestURL, manifestContent string, c config.Config) *DASHFilter {
	return &DASHFilter{
		manifestURL:     manifestURL,
		manifestContent: manifestContent,
		config:          c,
	}
}

// GetMaxAge returns max_age to be overwritten via cache control headers
// currently no support in dash for managing header
func (d *DASHFilter) GetMaxAge() string {
	return ""
}

// FilterManifest will be responsible for filtering the manifest according  to the MediaFilters
func (d *DASHFilter) FilterManifest(ctx context.Context, filters *parsers.MediaFilters) (string, error) {
	manifest, err := mpd.ReadFromString(d.manifestContent)
	if err != nil {
		return "", err
	}

	u, err := url.Parse(d.manifestURL)
	if err != nil {
		return "", fmt.Errorf("parsing manifest url: %w", err)
	}

	baseURLWithPath := func(p string) string {
		var sb strings.Builder
		sb.WriteString(u.Scheme)
		sb.WriteString("://")
		sb.WriteString(u.Host)
		sb.WriteString(p)
		sb.WriteString("/")
		return sb.String()
	}

	if manifest.BaseURL == "" {
		manifest.BaseURL = baseURLWithPath(path.Dir(u.Path))
	} else if !strings.HasPrefix(manifest.BaseURL, "http") {
		manifest.BaseURL = baseURLWithPath(path.Join(path.Dir(u.Path), manifest.BaseURL))
	}

	for _, filter := range d.getFilters(filters) {
		filter(filters, manifest)
	}

	for _, plugin := range filters.Plugins {
		if exec, ok := pluginDASH[plugin]; ok {
			exec(manifest)
		}
	}

	return manifest.WriteToString()
}

func (d *DASHFilter) getFilters(filters *parsers.MediaFilters) []execFilter {
	filterList := []execFilter{}
	if filters.ContentTypes != nil && len(filters.ContentTypes) > 0 {
		filterList = append(filterList, d.filterAdaptationSetContentType)
	}

	if filters.Videos.Bitrate != nil || filters.Audios.Bitrate != nil {
		filterList = append(filterList, d.filterBandwidth)
	}

	if filters.Videos.Codecs != nil {
		filterList = append(filterList, d.filterVideoTypes)
	}

	if filters.Audios.Codecs != nil {
		filterList = append(filterList, d.filterAudioTypes)
	}

	if filters.Captions.Codecs != nil {
		filterList = append(filterList, d.filterCaptionTypes)
	}

	if filters.FrameRate != nil {
		filterList = append(filterList, d.filterFrameRate)
	}

	if filters.Audios.Language != nil || filters.Captions.Language != nil {
		filterList = append(filterList, d.filterAdaptationSetLanguage)
	}

	return filterList
}

func (d *DASHFilter) filterVideoTypes(filters *parsers.MediaFilters, manifest *mpd.MPD) {
	supportedVideoTypes := map[string]struct{}{}
	for _, videoType := range filters.Videos.Codecs {
		supportedVideoTypes[string(videoType)] = struct{}{}
	}

	filterContentType(videoContentType, supportedVideoTypes, manifest)
}

func (d *DASHFilter) filterAudioTypes(filters *parsers.MediaFilters, manifest *mpd.MPD) {
	supportedAudioTypes := map[string]struct{}{}
	for _, audioType := range filters.Audios.Codecs {
		supportedAudioTypes[string(audioType)] = struct{}{}
	}

	filterContentType(audioContentType, supportedAudioTypes, manifest)
}

func (d *DASHFilter) filterCaptionTypes(filters *parsers.MediaFilters, manifest *mpd.MPD) {
	supportedCaptionTypes := map[string]struct{}{}
	for _, captionType := range filters.Captions.Codecs {
		supportedCaptionTypes[string(captionType)] = struct{}{}
	}

	filterContentType(captionContentType, supportedCaptionTypes, manifest)
}

func filterContentType(filter ContentType, supportedContentTypes map[string]struct{}, manifest *mpd.MPD) {
	for _, period := range manifest.Periods {
		var filteredAdaptationSets []*mpd.AdaptationSet
		for _, as := range period.AdaptationSets {
			if as.ContentType != nil && *as.ContentType == string(filter) {
				var filteredReps []*mpd.Representation
				for _, r := range as.Representations {
					if r.Codecs == nil {
						filteredReps = append(filteredReps, r)
						continue
					}

					if matchCodec(*r.Codecs, filter, supportedContentTypes) {
						continue
					}

					filteredReps = append(filteredReps, r)
				}
				as.Representations = filteredReps
			}

			if len(as.Representations) != 0 {
				filteredAdaptationSets = append(filteredAdaptationSets, as)
			}
		}

		for i, as := range filteredAdaptationSets {
			as.ID = strptr(strconv.Itoa(i))
		}
		period.AdaptationSets = filteredAdaptationSets
	}
}

func (d *DASHFilter) filterAdaptationSetLanguage(filters *parsers.MediaFilters, manifest *mpd.MPD) {
	for _, period := range manifest.Periods {
		var filteredAdaptationSets []*mpd.AdaptationSet
		for _, as := range period.AdaptationSets {
			if as.ContentType == nil {
				filteredAdaptationSets = append(filteredAdaptationSets, as)
				continue
			}

			var langs []string
			switch ContentType(*as.ContentType) {
			case audioContentType:
				langs = filters.Audios.Language
			case captionContentType:
				langs = filters.Captions.Language
			default:
				filteredAdaptationSets = append(filteredAdaptationSets, as)
				continue
			}

			if !(matchLang(*as.Lang, langs)) {
				filteredAdaptationSets = append(filteredAdaptationSets, as)
			}
		}

		for i, as := range filteredAdaptationSets {
			as.ID = strptr(strconv.Itoa(i))
		}
		period.AdaptationSets = filteredAdaptationSets
	}
}

func (d *DASHFilter) filterAdaptationSetContentType(filters *parsers.MediaFilters, manifest *mpd.MPD) {
	filteredAdaptationSetTypes := map[string]struct{}{}
	for _, streamType := range filters.ContentTypes {
		filteredAdaptationSetTypes[streamType] = struct{}{}
	}

	periodIndex := 0
	var filteredPeriods []*mpd.Period
	for _, period := range manifest.Periods {
		var filteredAdaptationSets []*mpd.AdaptationSet
		asIndex := 0
		for _, as := range period.AdaptationSets {
			if as.ContentType != nil {
				if _, filtered := filteredAdaptationSetTypes[*as.ContentType]; filtered {
					continue
				}
			}

			as.ID = strptr(strconv.Itoa(asIndex))
			asIndex++

			filteredAdaptationSets = append(filteredAdaptationSets, as)
		}

		if len(filteredAdaptationSets) == 0 {
			continue
		}

		period.AdaptationSets = filteredAdaptationSets
		period.ID = strconv.Itoa(periodIndex)
		periodIndex++

		filteredPeriods = append(filteredPeriods, period)
	}

	manifest.Periods = filteredPeriods
}

func (d *DASHFilter) filterFrameRate(filters *parsers.MediaFilters, manifest *mpd.MPD) {
	for _, period := range manifest.Periods {
		var filteredAdaptationSets []*mpd.AdaptationSet
		for _, as := range period.AdaptationSets {
			if as.FrameRate != nil {
				if matchFPS(*as.FrameRate, filters.FrameRate) {
					continue
				}
			}

			var filteredReps []*mpd.Representation
			for _, r := range as.Representations {
				if r.FrameRate == nil {
					filteredReps = append(filteredReps, r)
					continue
				}

				if matchFPS(*r.FrameRate, filters.FrameRate) {
					continue
				}

				filteredReps = append(filteredReps, r)
			}
			as.Representations = filteredReps

			if len(as.Representations) != 0 {
				filteredAdaptationSets = append(filteredAdaptationSets, as)
			}

		}

		for i, as := range filteredAdaptationSets {
			as.ID = strptr(strconv.Itoa(i))
		}
		period.AdaptationSets = filteredAdaptationSets
	}
}

func (d *DASHFilter) filterBandwidth(filters *parsers.MediaFilters, manifest *mpd.MPD) {
	for _, period := range manifest.Periods {
		var filteredAdaptationSets []*mpd.AdaptationSet
		for _, as := range period.AdaptationSets {
			if as.ContentType == nil {
				continue
			}

			//evaluate bitrate filter for codec type
			var bitrate *parsers.Bitrate
			switch ContentType(*as.ContentType) {
			case videoContentType:
				bitrate = filters.Videos.Bitrate
			case audioContentType:
				bitrate = filters.Audios.Bitrate
			}

			// if bitrate is nil, then no filtering needs to be applied
			// for this content type and we should append the representation
			if bitrate == nil {
				filteredAdaptationSets = append(filteredAdaptationSets, as)
				continue
			}

			var filteredRepresentations []*mpd.Representation
			for _, r := range as.Representations {
				if r.Bandwidth == nil {
					continue
				}

				if inRange(bitrate.Min, bitrate.Max, int(*r.Bandwidth)) {
					filteredRepresentations = append(filteredRepresentations, r)
				}
			}

			as.Representations = filteredRepresentations
			if len(as.Representations) != 0 {
				filteredAdaptationSets = append(filteredAdaptationSets, as)
			}
		}

		period.AdaptationSets = filteredAdaptationSets

		// Recalculate AdaptationSet id numbers
		for index, as := range period.AdaptationSets {
			as.ID = strptr(strconv.Itoa(index))
		}
	}
}

func matchLang(l string, langs []string) bool {
	for _, lang := range langs {
		if string(lang) == l {
			return true
		}
	}

	return false
}

func matchCodec(codec string, ct ContentType, supportedCodecs map[string]struct{}) bool {
	//the key in supportedCodecs for captionContentType is equivalent to codec
	//advertised in manifest. we can avoid iterating through each key
	if ct == captionContentType {
		_, found := supportedCodecs[codec]
		return found
	}

	for key := range supportedCodecs {
		if ValidCodecs(codec, CodecFilterID(key)) {
			return true
		}
	}

	return false
}

func matchFPS(fps string, framerates []string) bool {
	for _, fr := range framerates {
		if string(fr) == fps {
			return true
		}
	}

	return false
}

func strptr(s string) *string {
	return &s
}
