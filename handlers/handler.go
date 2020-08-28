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
		manifestOrigin, err := origin.Configure(r.Context(), c, masterManifestPath)
		if err != nil {
			e := NewErrorResponse("failed configuring origin", err)
			e.HandleError(r.Context(), w, http.StatusInternalServerError)
			return
		}

		logging.UpdateCtx(r.Context(), logging.Params{"playbackURL": manifestOrigin.GetPlaybackURL()})

		// fetch manifest from origin
		manifestInfo, err := manifestOrigin.FetchManifest(r.Context(), c.Client)
		if err != nil {
			e := NewErrorResponse("failed fetching manifest", err)
			e.HandleError(r.Context(), w, http.StatusInternalServerError)
			return
		}

		//throw status error if not 2xx
		if manifestInfo.Status/100 > 3 {
			err := fmt.Errorf("fetching manifest: returning http status of %v", manifestInfo.Status)
			e := NewErrorResponse("manifest origin error", err)
			e.HandleError(r.Context(), w, manifestInfo.Status)
			return
		}

		// create filter associated to the protocol and set
		// response headers accordingly
		var f filters.Filter
		switch mediaFilters.Protocol {
		case parsers.ProtocolHLS:
			f = filters.NewHLSFilter(manifestOrigin.GetPlaybackURL(), manifestInfo.Manifest, c)
			w.Header().Set("Content-Type", "application/x-mpegURL")
		case parsers.ProtocolDASH:
			f = filters.NewDASHFilter(manifestOrigin.GetPlaybackURL(), manifestInfo.Manifest, c)
			w.Header().Set("Content-Type", "application/dash+xml")
		}

		// apply the filters to the origin manifest
		filteredManifest, err := f.FilterManifest(r.Context(), mediaFilters)
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
