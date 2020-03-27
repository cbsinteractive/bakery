package parsers

import (
	"fmt"
	"math"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// VideoType is the video codec we need in a given playlist
type VideoType string

// AudioType is the audio codec we need in a given playlist
type AudioType string

// Language is the language we need in a given playlist
type Language string

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
	videoIFrame      VideoType = "i-frame"

	audioAAC                AudioType = "aac"
	audioAC3                AudioType = "ac-3"
	audioEnhacedAC3         AudioType = "ec-3"
	audioNoAudioDescription AudioType = "noAd"

	langEN   Language = "en"
	langPT   Language = "pt"
	langPTBR Language = "pt-BR"
	langES   Language = "es"
	langESMX Language = "es-MX"

	codecHDR10              Codec = "hdr10"
	codecDolbyVision        Codec = "dovi"
	codecHEVC               Codec = "hevc"
	codecH264               Codec = "avc"
	codecAAC                Codec = "aac"
	codecAC3                Codec = "ac-3"
	codecEnhancedAC3        Codec = "ec-3"
	codecNoAudioDescription Codec = "noAd"

	videoContent   ContentType = "video"
	audioContent   ContentType = "audio"
	captionContent ContentType = "caption"

	// ProtocolHLS for manifest in hls
	ProtocolHLS Protocol = "hls"
	// ProtocolDASH for manifests in dash
	ProtocolDASH Protocol = "dash"
)

// Trim is a struct that carries the start and end times to trim playlist
type Trim struct {
	Start int `json:",omitempty"`
	End   int `json:",omitempty"`
}

// Bitrate is a struct that carries Min and Max bitrate values
type Bitrate struct {
	Max int `json:",omitempty"`
	Min int `json:",omitempty"`
}

// MediaFilters is a struct that carry all the information passed via url
type MediaFilters struct {
	Videos       NestedFilters `json:",omitempty"`
	Audios       NestedFilters `json:",omitempty"`
	Captions     NestedFilters `json:",omitempty"`
	ContentTypes []ContentType `json:",omitempty"`
	Plugins      []string      `json:",omitempty"`
	IFrame       bool          `json:",omitempty"`
	Trim         *Trim         `json:",omitempty"`
	Bitrate      *Bitrate      `json:",omitempty"`
	Protocol     Protocol      `json:"protocol"`
}

// NestedFilters is a struct that holds values of filters
// that can be nested within certain Media Filters
type NestedFilters struct {
	Bitrate  *Bitrate   `json:",omitempty"`
	Codecs   []Codec    `json:",omitempty"`
	Language []Language `json:",omitempty"`
}

var urlParseRegexp = regexp.MustCompile(`(.*?)\((.*)\)`)
var nestedFilterRegexp = regexp.MustCompile(`\),`)

func keyError(key string, e error) (string, *MediaFilters, error) {
	return "", &MediaFilters{}, fmt.Errorf("Error parsing filter key: %v. Got error: %w", key, e)
}

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

	for _, part := range parts {
		// FindStringSubmatch should return a slice with
		// the full string, the key and filters (3 elements).
		// If it doesn't match, it means that the path is part
		// of the official manifest path so we concatenate to it.
		subparts := re.FindStringSubmatch(part)
		if len(subparts) != 3 {
			if mf.parsePlugins(part) {
				continue
			}
			masterManifestPath = path.Join(masterManifestPath, part)
			continue
		}

		filters := strings.Split(subparts[2], ",")
		nestedFilterRegexp := regexp.MustCompile(`\),`)
		nestedFilters := splitAfter(subparts[2], nestedFilterRegexp)

		switch key := subparts[1]; key {
		case "v":
			for _, sf := range nestedFilters {
				err := mf.getNestedFilters(sf, videoContent)
				if err != nil {
					return keyError("Video", err)
				}
			}
		case "a":
			for _, sf := range nestedFilters {
				if err := mf.getNestedFilters(sf, audioContent); err != nil {
					return keyError("Audio", err)
				}
			}
		case "c":
			for _, sf := range nestedFilters {
				if err := mf.getNestedFilters(sf, captionContent); err != nil {
					return keyError("Captions", err)
				}
			}
		case "ct":
			for _, contentType := range filters {
				mf.ContentTypes = append(mf.ContentTypes, ContentType(contentType))
			}
		case "l":
			for _, lang := range filters {
				mf.Audios.Language = append(mf.Audios.Language, Language(lang))
				mf.Captions.Language = append(mf.Captions.Language, Language(lang))
			}
		case "b":
			x, y, err := parseAndValidateInts(filters, math.MaxInt32)
			if err != nil {
				return keyError("Bitrate", err)
			}

			mf.Bitrate = &Bitrate{
				Min: x,
				Max: y,
			}
		case "t":
			x, y, err := parseAndValidateInts(filters, int(time.Now().Unix()))
			if err != nil {
				return keyError("Trim", err)
			}

			mf.Trim = &Trim{
				Start: x,
				End:   y,
			}
		}
	}

	mf.normalizeBitrateFilter()

	return masterManifestPath, mf, nil
}

func (mf *MediaFilters) parsePlugins(path string) bool {
	re := regexp.MustCompile(`\[(.*)\]`)
	subparts := re.FindStringSubmatch(path)

	if len(subparts) == 2 {
		for _, plugin := range strings.Split(subparts[1], ",") {
			mf.Plugins = append(mf.Plugins, plugin)
		}
		return true
	}

	return false
}

func (mf *MediaFilters) getNestedFilters(nestedFilter string, streamType ContentType) error {
	// assumes nested filters are properly formatted
	splitNestedFilter := urlParseRegexp.FindStringSubmatch(nestedFilter)
	var key string
	var param []string

	if len(splitNestedFilter) == 0 { //default behavior is codec values
		key = "co"
		param = strings.Split(nestedFilter, ",")
	} else {
		key = splitNestedFilter[1]
		param = strings.Split(splitNestedFilter[2], ",")
	}

	// split key by ',' to account for situations like filter(codec,codec,b(low,high))
	// where key = codec,codec,b
	splitKey := strings.Split(key, ",")
	if len(splitKey) == 1 {
		return mf.parseNestedFilterKeys(streamType, key, param)
	}

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

	for i := range keys {
		err := mf.parseNestedFilterKeys(streamType, keys[i], params[i])
		if err != nil {
			return err
		}
	}

	return nil
}

// ParseNestedFilter takes a NestedFilter and sets Audios' or Videos' values accordingly.
func (mf *MediaFilters) parseNestedFilterKeys(streamType ContentType, key string, values []string) error {
	var nf *NestedFilters

	switch streamType {
	case audioContent:
		nf = &mf.Audios
	case videoContent:
		nf = &mf.Videos
	case captionContent:
		nf = &mf.Captions
	}

	switch key {
	case "co":
		for _, v := range values {
			switch v {
			case "hdr10":
				nf.Codecs = append(nf.Codecs, Codec("hev1.2"), Codec("hvc1.2"))
			case "i-frame":
				mf.IFrame = true
			default:
				nf.Codecs = append(nf.Codecs, Codec(v))
			}
		}
	case "l":
		for _, v := range values {
			nf.Language = append(nf.Language, Language(v))
		}
	case "b":
		x, y, err := parseAndValidateInts(values, math.MaxInt32)
		if err != nil {
			return err
		}
		nf.Bitrate = &Bitrate{
			Min: x,
			Max: y,
		}
	}

	return nil
}

// normalizeBitrateFilter will finalize the nested bitrate filter by comparing it to
// overall bitrate filter and overriding any necessary values
func (mf *MediaFilters) normalizeBitrateFilter() {
	if mf.Bitrate == nil {
		return
	}

	if mf.Audios.Bitrate == nil {
		mf.Audios.Bitrate = mf.Bitrate
	}

	if mf.Videos.Bitrate == nil {
		mf.Videos.Bitrate = mf.Bitrate
	}

	mf.Bitrate = nil
}

// parseAndValidateInts will parse a range of two ints and validate their range
func parseAndValidateInts(values []string, max int) (int, int, error) {
	var x, y int
	var err error

	if values[0] != "" {
		x, err = strconv.Atoi(values[0])
		if err != nil {
			return x, y, err
		}
	} else { // if lower bound is not set, default to 0
		x = 0
	}

	if len(values) > 1 && values[1] != "" {
		y, err = strconv.Atoi(values[1])
		if err != nil {
			return x, y, err
		}
	} else { // if higher bound is not set, set it to max value
		y = max
	}

	if !validatePositiveRange(x, y, max) {
		return x, y, fmt.Errorf("invalid range for provided values: ( %v, %v )", x, y)
	}

	return x, y, nil
}
