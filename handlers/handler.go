package handlers

import (
	"fmt"
	"net/http"

	"github.com/cbsinteractive/bakery/config"
	"github.com/cbsinteractive/bakery/filters"
	"github.com/cbsinteractive/bakery/logging"
	"github.com/cbsinteractive/bakery/origin"
	"github.com/cbsinteractive/bakery/parsers"
)

// LoadHandler loads the handler for all the requests
func LoadHandler(c config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// parse all the filters from the URL
		masterManifestPath, mediaFilters, err := parsers.URLParse(r.URL.Path)
		if err != nil {
			e := NewErrorResponse("failed parsing filters", err)
			e.HandleError(r.Context(), w, http.StatusBadRequest)
			return
		}

		//configure origin from path
		o, err := origin.Configure(r.Context(), c, masterManifestPath)
		if err != nil {
			e := NewErrorResponse("failed configuring origin", err)
			e.HandleError(r.Context(), w, http.StatusInternalServerError)
			return
		}

		logging.UpdateCtx(r.Context(), logging.Params{"playbackURL": o.GetPlaybackURL()})

		// fetch manifest from origin
		contentInfo, err := o.FetchOriginContent(r.Context(), c.Client)
		if err != nil {
			e := NewErrorResponse("failed fetching manifest", err)
			e.HandleError(r.Context(), w, http.StatusInternalServerError)
			return
		}

		//throw status error if not 2xx
		if contentInfo.Status/100 > 3 {
			if mediaFilters.PreventHTTPStatusError {
				switch mediaFilters.Protocol {
				case parsers.ProtocolHLS:
					w.Header().Set("Content-Type", "application/x-mpegURL")
					fmt.Fprint(w, filters.EmptyHLSManifestContent)
				case parsers.ProtocolVTT:
					w.Header().Set("Content-Type", "text/vtt")
					fmt.Fprint(w, filters.EmptyVTTContent)
				}
				return
			}
			err := fmt.Errorf("fetching manifest: returning http status of %v", contentInfo.Status)
			e := NewErrorResponse("manifest origin error", err)
			e.HandleError(r.Context(), w, contentInfo.Status)
			return
		}

		// create filter associated to the protocol and set
		// response headers accordingly
		var f filters.Filter
		switch mediaFilters.Protocol {
		case parsers.ProtocolHLS:
			f = filters.NewHLSFilter(o.GetPlaybackURL(), contentInfo.Payload, c)
			w.Header().Set("Content-Type", "application/x-mpegURL")
		case parsers.ProtocolDASH:
			f = filters.NewDASHFilter(o.GetPlaybackURL(), contentInfo.Payload, c)
			w.Header().Set("Content-Type", "application/dash+xml")
		case parsers.ProtocolVTT:
			f = filters.NewVTTFilter(o.GetPlaybackURL(), contentInfo.Payload, c)
			w.Header().Set("Content-Type", "text/vtt")
		}

		// apply the filters to the origin manifest
		filteredManifest, err := f.FilterContent(r.Context(), mediaFilters)
		if err != nil {
			e := NewErrorResponse("failed to filter manifest", err)
			e.HandleError(r.Context(), w, http.StatusInternalServerError)
			return
		}

		// set cache-control if serving hls media playlist
		if maxAge := f.GetMaxAge(); maxAge != "" && maxAge != "0" {
			w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%v", maxAge))
		}

		// write the filtered manifest to the response
		fmt.Fprint(w, filteredManifest)
	})
}
