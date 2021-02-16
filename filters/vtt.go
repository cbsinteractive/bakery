package filters

import (
	"context"

	"github.com/cbsinteractive/bakery/config"
	"github.com/cbsinteractive/bakery/parsers"
)

const EmptyVTTContent = "WEBVTT\nX-TIMESTAMP-MAP=LOCAL:00:00:00.000,MPEGTS:0\n\nNOTE Failed to fetch origin WebVTT, preventing HTTP status error.\n"


// VTTFilter implements the Filter interface for VTT files
type VTTFilter struct {
	originURL     string
	originContent string
	config        config.Config
}

// NewVTTFilter is the VTT filter constructor
func NewVTTFilter(originURL, originContent string, c config.Config) *VTTFilter {
	return &VTTFilter{
		originURL:     originURL,
		originContent: originContent,
		config:        c,
	}
}

// FilterContent will be responsible for filtering VTT files based on
// mediaFilters
func (v *VTTFilter) FilterContent(ctx context.Context, filters *parsers.MediaFilters) (string, error) {
	return v.originContent, nil
}

// GetMaxAge returns max_age to  be overwritten via cache control
// headers, currently not supported for VTT files
func (v *VTTFilter) GetMaxAge() string {
	return ""
}
