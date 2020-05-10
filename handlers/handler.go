package handlers

import (
	"fmt"
	"net/http"

	"github.com/cbsinteractive/bakery/config"
	"github.com/cbsinteractive/bakery/filters"
	"github.com/cbsinteractive/bakery/origin"
	"github.com/cbsinteractive/bakery/parsers"
)

// LoadHandler loads the handler for all the requests
func LoadHandler(c config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//set context on client which is copied over
		//when fetching manifest and configuring origins
		c.Client.SetContext(r)

		w.Header().Set("Access-Control-Allow-Origin", "*")
		ctxLogger := c.Logger.With().Str("uri", r.RequestURI).Logger()
		ctxLogger.Info().
			Str("method", r.Method).
			Str("raddr", r.RemoteAddr).
			Str("ref", r.Referer()).
			Str("ua", r.UserAgent()).
			Msg("received request")

		// parse all the filters from the URL
		masterManifestPath, mediaFilters, err := parsers.URLParse(r.URL.Path)
		if err != nil {
			e := NewErrorResponse("failed parsing filters", err)
			e.HandleError(ctxLogger, w, http.StatusBadRequest)
			return
		}

		//configure origin from path
		manifestOrigin, err := origin.Configure(c, masterManifestPath)
		if err != nil {
			e := NewErrorResponse("failed configuring origin", err)
			e.HandleError(ctxLogger, w, http.StatusInternalServerError)
			return
		}

		// fetch manifest from origin
		manifestContent, err := manifestOrigin.FetchManifest(c.Client)
		if err != nil {
			ctxl := ctxLogger.With().Str("playbackURL", manifestOrigin.GetPlaybackURL()).Logger()
			e := NewErrorResponse("failed fetching manifest", err)
			e.HandleError(ctxl, w, http.StatusInternalServerError)
			return
		}

		// create filter associated to the protocol and set
		// response headers accordingly
		var f filters.Filter
		switch mediaFilters.Protocol {
		case parsers.ProtocolHLS:
			f = filters.NewHLSFilter(manifestOrigin.GetPlaybackURL(), manifestContent, c)
			w.Header().Set("Content-Type", "application/x-mpegURL")
		case parsers.ProtocolDASH:
			f = filters.NewDASHFilter(manifestOrigin.GetPlaybackURL(), manifestContent, c)
			w.Header().Set("Content-Type", "application/dash+xml")
		}

		// apply the filters to the origin manifest
		filteredManifest, err := f.FilterManifest(mediaFilters)
		if err != nil {
			e := NewErrorResponse("failed to filter manifest", err)
			e.HandleError(ctxLogger, w, http.StatusInternalServerError)
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
