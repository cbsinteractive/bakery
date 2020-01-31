package filters

import (
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/cbsinteractive/bakery/pkg/config"
	"github.com/cbsinteractive/bakery/pkg/parsers"
	"github.com/zencoder/go-dash/mpd"
)

const captionContentType = "text"
const audioContentType = "audio"
const videoContentType = "video"

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

// FilterManifest will be responsible for filtering the manifest according  to the MediaFilters
func (d *DASHFilter) FilterManifest(filters *parsers.MediaFilters) (string, error) {
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

	if filters.FilterStreamTypes != nil && len(filters.FilterStreamTypes) > 0 {
		d.filterAdaptationSetType(filters, manifest)
	}

	if filters.Videos != nil {
		d.filterAudioTypes(filters, manifest)
	}

	if filters.Audios != nil {
		d.filterAudioTypes(filters, manifest)
	}

	if filters.CaptionTypes != nil {
		d.filterCaptionTypes(filters, manifest)
	}

	return manifest.WriteToString()
}

func (d *DASHFilter) filterVideoTypes(filters *parsers.MediaFilters, manifest *mpd.MPD) {
	supportedVideoTypes := map[string]struct{}{}
	for _, videoType := range filters.Videos {
		supportedVideoTypes[string(videoType)] = struct{}{}
	}

	filterContentType(captionContentType, supportedVideoTypes, manifest)
}

func (d *DASHFilter) filterAudioTypes(filters *parsers.MediaFilters, manifest *mpd.MPD) {
	supportedAudioTypes := map[string]struct{}{}
	for _, audioType := range filters.Audios {
		supportedAudioTypes[string(audioType)] = struct{}{}
	}

	filterContentType(audioContentType, supportedAudioTypes, manifest)
}

func (d *DASHFilter) filterCaptionTypes(filters *parsers.MediaFilters, manifest *mpd.MPD) {
	supportedCaptionTypes := map[string]struct{}{}
	for _, captionType := range filters.CaptionTypes {
		supportedCaptionTypes[string(captionType)] = struct{}{}
	}

	filterContentType(captionContentType, supportedCaptionTypes, manifest)
}

func filterContentType(filter string, supportedContentTypes map[string]struct{}, manifest *mpd.MPD) {
	for _, period := range manifest.Periods {
		for _, as := range period.AdaptationSets {
			if as.ContentType != nil && *as.ContentType == filter {
				var filteredReps []*mpd.Representation
				for _, r := range as.Representations {
					if r.Codecs == nil {
						filteredReps = append(filteredReps, r)
						continue
					}

					if _, supported := supportedContentTypes[*r.Codecs]; supported {
						filteredReps = append(filteredReps, r)
					}
				}
				as.Representations = filteredReps
			}
		}
	}
}

func (d *DASHFilter) filterAdaptationSetType(filters *parsers.MediaFilters, manifest *mpd.MPD) {
	filteredAdaptationSetTypes := map[parsers.StreamType]struct{}{}
	for _, streamType := range filters.FilterStreamTypes {
		filteredAdaptationSetTypes[streamType] = struct{}{}
	}

	periodIndex := 0
	var filteredPeriods []*mpd.Period
	for _, period := range manifest.Periods {
		var filteredAdaptationSets []*mpd.AdaptationSet
		asIndex := 0
		for _, as := range period.AdaptationSets {
			if as.ContentType != nil {
				if _, filtered := filteredAdaptationSetTypes[parsers.StreamType(*as.ContentType)]; filtered {
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

func strptr(s string) *string {
	return &s
}
