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

// MediaFilters is a struct that carry all the information passed via url
type MediaFilters struct {
	Videos       NestedFilters `json:",omitempty"`
	Audios       NestedFilters `json:",omitempty"`
	Captions     NestedFilters `json:",omitempty"`
	ContentTypes []string      `json:",omitempty"`
	Plugins      []string      `json:",omitempty"`
	Tags         *Tags         `json:",omitempty"`
	Trim         *Trim         `json:",omitempty"`
	Bitrate      *Bitrate      `json:",omitempty"`
	FrameRate    []string      `json:",omitempty"`
	DeWeave      bool          `json:",omitempty"`
	Protocol     Protocol      `json:"protocol"`
}

// NestedFilters is a struct that holds values of filters
// that can be nested within certain Media Filters
type NestedFilters struct {
	Bitrate  *Bitrate `json:",omitempty"`
	Codecs   []string `json:",omitempty"`
	Language []string `json:",omitempty"`
}

// Protocol describe the valid protocols
type Protocol string

const (
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

// Tags holds values of HLS tags that are to be suppressed
// from the manifest
type Tags struct {
	Ads    bool `json:",omitempty"`
	IFrame bool `json:",omitempty"`
}

var urlParseRegexp = regexp.MustCompile(`(.*?)\((.*)\)`)
var nestedFilterRegexp = regexp.MustCompile(`\),`)

var codecSupported = map[string]struct{}{
	"hdr10": struct{}{}, //H265 main profile 2
	"dvh":   struct{}{}, //Dolby Vision
	"hevc":  struct{}{}, //H265
	"hvc":   struct{}{}, //H265
	"avc":   struct{}{}, //h264
	"av1":   struct{}{}, //AV1
	"mp4a":  struct{}{}, //AAC audio
	"ac-3":  struct{}{}, //AC3 audio
	"ec-3":  struct{}{}, //Enhanved AC3
	"stpp":  struct{}{}, //Subtitles
	"wvtt":  struct{}{}, //WebVTT
}

var contentSupported = map[string]struct{}{
	"image": struct{}{},
	"text":  struct{}{},
	"audio": struct{}{},
	"video": struct{}{},
}

func keyError(key string, e error) (string, *MediaFilters, error) {
	return "", &MediaFilters{}, fmt.Errorf("%v: %w", key, e)
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
	} else {
		return keyError("Protocol", fmt.Errorf("unsupported protocol"))
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
			for _, nf := range nestedFilters {
				if err := mf.Videos.parse(nf); err != nil {
					return keyError("Video", err)
				}
			}
		case "a":
			for _, nf := range nestedFilters {
				if err := mf.Audios.parse(nf); err != nil {
					return keyError("Audio", err)
				}
			}
		case "c":
			for _, nf := range nestedFilters {
				if err := mf.Captions.parse(nf); err != nil {
					return keyError("Captions", err)
				}
			}
		case "ct":
			for _, contentType := range filters {
				if _, valid := contentSupported[contentType]; !valid {
					err := fmt.Errorf("Content Type %v is not supported", contentType)
					return keyError("Content Type", err)
				}
				mf.ContentTypes = append(mf.ContentTypes, contentType)
			}
		case "l":
			for _, lang := range filters {
				mf.Audios.Language = append(mf.Audios.Language, lang)
				mf.Captions.Language = append(mf.Captions.Language, lang)
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

			milli := 1000
			mf.Trim = &Trim{
				Start: x*milli,
				End:   y*milli,
			}
		case "tags": //only applied when trimming/serving hls media playlists
			mf.Tags = &Tags{}
			mf.Tags.parse(filters)
		case "fps": //fps types in hls=float64, dash=string
			for _, framerate := range filters {
				fr := strings.ReplaceAll(framerate, ":", "/")
				mf.FrameRate = append(mf.FrameRate, fr)
			}
		case "dw":
			if len(filters) > 1 {
				return keyError("DeWeave", fmt.Errorf("Only accepts one boolean value"))
			}

			w, err := parseAndValidateBooleanString(filters[0])
			if err != nil {
				return keyError("DeWeave", err)
			}

			mf.DeWeave = w
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

func (nf *NestedFilters) parse(nestedFilter string) error {
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
		return nf.parseKeys(key, param)
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
		err := nf.parseKeys(keys[i], params[i])
		if err != nil {
			return err
		}
	}

	return nil
}

// ParseNestedFilter takes a NestedFilter and sets Audios' or Videos' values accordingly.
func (nf *NestedFilters) parseKeys(key string, values []string) error {
	switch key {
	case "co":
		for _, v := range values {
			switch v {
			case "hdr10":
				nf.Codecs = append(nf.Codecs, "hev1.2", "hvc1.2")
			default:
				if _, valid := codecSupported[v]; !valid {
					return fmt.Errorf("Codec %v is not supported", v)
				}
				nf.Codecs = append(nf.Codecs, v)
			}
		}
	case "l":
		for _, v := range values {
			nf.Language = append(nf.Language, v)
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

func (t *Tags) parse(values []string) {
	for _, tag := range values {
		switch tag {
		case "ads":
			t.Ads = true
		case "i-frame":
			t.IFrame = true
		case "iframe":
			t.IFrame = true
		}
	}
}

// SuppressAds will evaluate whether the ad tag was set
func (mf *MediaFilters) SuppressAds() bool {
	if mf.Tags == nil {
		return false
	}

	return mf.Tags.Ads
}

// SuppressIFrame will evaluate whether the i-frame tag was set
func (mf *MediaFilters) SuppressIFrame() bool {
	if mf.Tags == nil {
		return false
	}

	return mf.Tags.IFrame
}

func parseAndValidateBooleanString(v string) (bool, error) {
	switch v {
	case "false":
		return false, nil
	case "true":
		return true, nil
	default:
		return false, fmt.Errorf("Can't recognize value of %v", v)
	}
}
