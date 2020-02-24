package filters

import (
	"github.com/cbsinteractive/bakery/pkg/parsers"
	"strings"
)

// Filter is an interface for HLS and DASH filters
type Filter interface {
	FilterManifest(filters *parsers.MediaFilters) (string, error)
}

// ContentType represents the content in the stream
type ContentType string

const (
	captionContentType ContentType = "text"
	audioContentType   ContentType = "audio"
	videoContentType   ContentType = "video"
)

// CodecFilterID is the formatted codec represented in a given playlist
type CodecFilterID string

const (
	hevcCodec  CodecFilterID = "hvc"
	avcCodec   CodecFilterID = "avc"
	dolbyCodec CodecFilterID = "dvh"
	aacCodec   CodecFilterID = "mp4a"
	ec3Codec   CodecFilterID = "ec-3"
	ac3Codec   CodecFilterID = "ac-3"
	stppCodec  CodecFilterID = "stpp"
	wvttCodec  CodecFilterID = "wvtt"
)

// ValidCodecs returns a map of all formatted values for a given codec filter
func ValidCodecs(codec string, filter CodecFilterID) bool {
	return strings.Contains(codec, string(filter))
}
