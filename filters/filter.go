package filters

import (
	"context"
	"strings"

	"github.com/cbsinteractive/bakery/parsers"
)

// Filter is an interface for HLS and DASH filters
type Filter interface {
	FilterManifest(ctx context.Context, filters *parsers.MediaFilters) (string, error)
	GetMaxAge() string
}

// ContentType represents the content in the stream
type ContentType string

const (
	imageContentType   ContentType = "image"
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

// ValidCodecs returns a boolean value for a given codec filter
func ValidCodecs(codec string, filter CodecFilterID) bool {
	return strings.Contains(codec, string(filter))
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

func inRange(start int, end int, value int) bool {
	return (start <= value) && (value <= end)
}
