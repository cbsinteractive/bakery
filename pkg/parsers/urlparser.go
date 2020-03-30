package parsers

import (
	"fmt"
	"math"
	"path"
	"regexp"
	"strconv"
	"strings"
)

// VideoType is the video codec we need in a given playlist
type VideoType string

// AudioType is the audio codec we need in a given playlist
type AudioType string

// AudioLanguage is the audio language we need in a given playlist
type AudioLanguage string

// CaptionLanguage is the audio language we need in a given playlist
type CaptionLanguage string

// CaptionType is an allowed caption format for the stream
type CaptionType string

// ContentType represents one stream type (e.g. video, audio, text)
type ContentType string

// Codec represet the codec of the ContentType
type Codec string

// Protocol describe the valid protocols
type Protocol string

const (
	videoHDR10       VideoType = "hdr10"
	videoDolbyVision VideoType = "dovi"
	videoHEVC        VideoType = "hevc"
	videoH264        VideoType = "avc"

	audioAAC                AudioType = "aac"
	audioAC3                AudioType = "ac-3"
	audioEnhacedAC3         AudioType = "ec-3"
	audioNoAudioDescription AudioType = "noAd"

	audioLangPTBR AudioLanguage = "pt-BR"
	audioLangES   AudioLanguage = "es-MX"
	audioLangEN   AudioLanguage = "en"

	captionPTBR CaptionLanguage = "pt-BR"
	captionES   CaptionLanguage = "es-MX"
	captionEN   CaptionLanguage = "en"

	codecHDR10              Codec = "hdr10"
	codecDolbyVision        Codec = "dovi"
	codecHEVC               Codec = "hevc"
	codecH264               Codec = "avc"
	codecAAC                Codec = "aac"
	codecAC3                Codec = "ac-3"
	codecEnhancedAC3        Codec = "ec-3"
	codecNoAudioDescription Codec = "noAd"

	// ProtocolHLS for manifest in hls
	ProtocolHLS Protocol = "hls"
	// ProtocolDASH for manifests in dash
	ProtocolDASH Protocol = "dash"
)

// Trim is a struct that carries the start and end times to trim playlist
type Trim struct {
	Start int64 `json:",omitempty"`
	End   int64 `json:",omitempty"`
}

// MediaFilters is a struct that carry all the information passed via url
type MediaFilters struct {
	VideoFilters NestedFilters `json:",omitempty"`
	AudioFilters NestedFilters `json:",omitempty"`
	CaptionTypes []CaptionType `json:",omitempty"`
	ContentTypes []ContentType `json:",omitempty"`
	MaxBitrate   int           `json:",omitempty"`
	MinBitrate   int           `json:",omitempty"`
	Plugins      []string      `json:",omitempty"`
	Trim         *Trim         `json:",omitempty"`
	Protocol     Protocol      `json:"protocol"`
}

type NestedFilters struct {
	MinBitrate int     `json:",omitempty"`
	MaxBitrate int     `json:",omitempty"`
	Codecs     []Codec `json:",omitempty"`
}

var urlParseRegexp = regexp.MustCompile(`(.*?)\((.*)\)`)

// URLParse will generate a MediaFilters struct with
// all the filters that needs to be applied to the
// master manifest. It will also return the master manifest
// url without the filters.
func URLParse(urlpath string) (string, *MediaFilters, error) {
	mf := new(MediaFilters)
	parts := strings.Split(urlpath, "/")
	re := urlParseRegexp
	masterManifestPath := "/"

	if strings.Contains(urlpath, ".m3u8") {
		mf.Protocol = ProtocolHLS
	} else if strings.Contains(urlpath, ".mpd") {
		mf.Protocol = ProtocolDASH
	}

	mf.initializeBitrateRange()

	for _, part := range parts {
		// FindStringSubmatch should return a slice with
		// the full string, the key and filters (3 elements).
		// If it doesn't match, it means that the path is part
		// of the official manifest path so we concatenate to it.
		subparts := re.FindStringSubmatch(part)
		if len(subparts) != 3 {
			if mf.filterPlugins(part) {
				continue
			}
			masterManifestPath = path.Join(masterManifestPath, part)
			continue
		}

		filters := strings.Split(subparts[2], ",")

		var err error

		nestedFilterRegexp := regexp.MustCompile(`\),`)
		nestedFilters := SplitAfter(subparts[2], nestedFilterRegexp)

		switch key := subparts[1]; key {
		case "v":
			for _, sf := range nestedFilters {
				result, filter, err := mf.findNestedFilters(sf, ContentType("video"))
				if err != nil {
					return result, filter, err
				}
			}

		case "a":
			for _, sf := range nestedFilters {
				result, filter, err := mf.findNestedFilters(sf, ContentType("audio"))
				if err != nil {
					return result, filter, err
				}
			}
		case "c":
			if mf.CaptionTypes == nil {
				mf.CaptionTypes = []CaptionType{}
			}

			for _, captionType := range filters {
				mf.CaptionTypes = append(mf.CaptionTypes, CaptionType(captionType))
			}
		case "ct":
			for _, contentType := range filters {
				mf.ContentTypes = append(mf.ContentTypes, ContentType(contentType))
			}
		case "b":
			if filters[0] != "" {
				mf.MinBitrate, err = strconv.Atoi(filters[0])
				if err != nil {
					return keyError("MinBitrate", err)
				}
			}

			if filters[1] != "" {
				mf.MaxBitrate, err = strconv.Atoi(filters[1])
				if err != nil {
					return keyError("MaxBitrate", err)
				}
			}

			if isGreater(mf.MinBitrate, mf.MaxBitrate) {
				return keyError("bitrate", fmt.Errorf("MinBitrate is greater than or equal to MaxBitrate"))
			}
		case "t":
			var trim Trim
			if filters[0] != "" {
				trim.Start, err = strconv.ParseInt(filters[0], 10, 64)
				if err != nil {
					return keyError("trim", err)
				}
			}

			if filters[1] != "" {
				trim.End, err = strconv.ParseInt(filters[1], 10, 64)
				if err != nil {
					return keyError("trim", err)
				}
			}

			if isGreater(int(trim.Start), int(trim.End)) {
				return keyError("trim", fmt.Errorf("Start Time is greater than or equal to End Time"))
			}
			mf.Trim = &trim
		}
	}

	return masterManifestPath, mf, nil
}

// validate ranges like Trim and Bitrate
func isGreater(x int, y int) bool {
	return x >= y
}

func keyError(key string, e error) (string, *MediaFilters, error) {
	return "", &MediaFilters{}, fmt.Errorf("Error parsing filter key: %v. Got error: %w", key, e)
}

func (f *MediaFilters) filterPlugins(path string) bool {
	re := regexp.MustCompile(`\[(.*)\]`)
	subparts := re.FindStringSubmatch(path)

	if len(subparts) == 2 {
		for _, plugin := range strings.Split(subparts[1], ",") {
			f.Plugins = append(f.Plugins, plugin)
		}
		return true
	}

	return false
}

func (mf *MediaFilters) findNestedFilters(nestedFilter string, streamType ContentType) (string, *MediaFilters, error) {
	// assumes nested filters are properly formatted
	splitNestedFilter := urlParseRegexp.FindStringSubmatch(nestedFilter)
	var key string
	var param []string
	if len(splitNestedFilter) == 0 {
		key = "co"
		param = strings.Split(nestedFilter, ",")
	} else {
		key = splitNestedFilter[1]
		param = strings.Split(splitNestedFilter[2], ",")
	}

	// split key by ',' to account for situations like filter(codec,codec,b(low,high))
	// as in such a situation, key = codec,codec,b
	splitKey := strings.Split(key, ",")
	if len(splitKey) == 1 {
		result, filter, err := mf.normalizeNestedFilter(streamType, key, param)
		if err != nil {
			return result, filter, err
		}
	} else {
		var keys []string
		var params [][]string
		for i, part := range splitKey {
			if i == len(splitKey)-1 {
				keys = append(keys, part)
				params = append(params, param)
			} else {
				keys = append(keys, "co")
				params = append(params, []string{part})
			}
		}

		for i, _ := range keys {
			result, filter, err := mf.normalizeNestedFilter(streamType, keys[i], params[i])
			if err != nil {
				return result, filter, err
			}
		}
	}

	return "", mf, nil
}

// Initialize bitrate range for overall, audio, and video bitrate filters to 0, math.MaxInt32
func (f *MediaFilters) initializeBitrateRange() {
	f.MinBitrate = 0
	f.MaxBitrate = math.MaxInt32
	f.AudioFilters.MinBitrate = 0
	f.AudioFilters.MaxBitrate = math.MaxInt32
	f.VideoFilters.MinBitrate = 0
	f.VideoFilters.MaxBitrate = math.MaxInt32
}

// SplitAfter splits a string after the matchs of the specified regexp
func SplitAfter(s string, re *regexp.Regexp) []string {
	var splitResults []string
	var position int
	indices := re.FindAllStringIndex(s, -1)
	if indices == nil {
		return append(splitResults, s)
	}
	for _, idx := range indices {
		section := s[position:idx[1]]
		splitResults = append(splitResults, section)
		position = idx[1]
	}
	return append(splitResults, s[position:])
}

// normalizeNestedFilter takes a NestedFilter and sets AudioFilters' or VideoFilters' values accordingly.
func (mf *MediaFilters) normalizeNestedFilter(streamType ContentType, key string, values []string) (string, *MediaFilters, error) {
	var streamNestedFilters *NestedFilters
	var err error
	switch streamType {
	case "audio":
		streamNestedFilters = &mf.AudioFilters
	case "video":
		streamNestedFilters = &mf.VideoFilters
	}

	switch key {
	case "co":
		for _, v := range values {
			if v == "hdr10" {
				streamNestedFilters.Codecs = append(streamNestedFilters.Codecs, Codec("hev1.2"), Codec("hvc1.2"))
			} else {
				streamNestedFilters.Codecs = append(streamNestedFilters.Codecs, Codec(v))
			}
		}
	case "b":
		if values[0] != "" {
			streamNestedFilters.MinBitrate, err = strconv.Atoi(values[0])
			if err != nil {
				return keyError("MinBitrate", err)
			}
			if streamNestedFilters.MinBitrate < 0 || streamNestedFilters.MinBitrate > math.MaxInt32 {
				return keyError("MaxBitrate", fmt.Errorf("MinBitrate is negative or exceeds math.MaxInt32"))
			}
		}

		if values[1] != "" {
			streamNestedFilters.MaxBitrate, err = strconv.Atoi(values[1])
			if err != nil {
				return keyError("MaxBitrate", err)
			}
			if streamNestedFilters.MaxBitrate < 0 || streamNestedFilters.MaxBitrate > math.MaxInt32 {
				return keyError("MaxBitrate", fmt.Errorf("MaxBitrate is negative or exceeds math.MaxInt32"))
			}
		}

		if isGreater(streamNestedFilters.MinBitrate, streamNestedFilters.MaxBitrate) {
			return keyError((string(streamType) + "bitrate"), fmt.Errorf("MinBitrate is greater than or equal to MaxBitrate"))
		}
	}

	return "", mf, nil
}
