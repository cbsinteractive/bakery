package filters

import (
	"context"
	"testing"

	"github.com/cbsinteractive/bakery/config"
	"github.com/cbsinteractive/bakery/parsers"
)

func TestVTTFilter_FilterContent(t *testing.T) {
	tests := []struct {
		name               string
		filters            *parsers.MediaFilters
		vttContent         string
		expectedVTTContent string
	}{
		{
			name:               "passthrough vtt content",
			filters:            &parsers.MediaFilters{},
			vttContent:         EmptyVTTContent,
			expectedVTTContent: EmptyVTTContent,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			filter := NewVTTFilter("", tt.vttContent, config.Config{})
			r, err := filter.FilterContent(context.Background(), tt.filters)
			if err != nil {
				t.Errorf("FilterContent() didnt expect an error to be returned, got: %v", err)
				return
			}

			if r != tt.expectedVTTContent {
				t.Errorf("wrong content  returned\ngot %v\nexpected: %v", r, tt.expectedVTTContent)
			}
		})
	}
}
