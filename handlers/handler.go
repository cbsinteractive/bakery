package handlers

import (
	"fmt"
	"net/http"

	"github.com/cbsinteractive/bakery/config"
	"github.com/cbsinteractive/bakery/filters"
	"github.com/cbsinteractive/bakery/origin"
	"github.com/cbsinteractive/bakery/parsers"
	"github.com/sirupsen/logrus"
)

// LoadHandler loads the handler for all the requests
func LoadHandler(c config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Client.SetContext(r)

		w.Header().Set("Access-Control-Allow-Origin", "*")
		logger := c.GetLogger()
		logger.WithFields(logrus.Fields{
			"method":      r.Method,
			"uri":         r.RequestURI,
			"remote-addr": r.RemoteAddr,
		}).Info("received request")

		if !c.Authenticate(r.Header.Get("x-bakery-origin-token")) {
			httpError(c, w, fmt.Errorf("authentication"), "failed authenticating request", http.StatusForbidden)
			return
		}

		// parse all the filters from the URL
		masterManifestPath, mediaFilters, err := parsers.URLParse(r.URL.Path)
		if err != nil {
			httpError(logger, w, err, "failed parsing filters", http.StatusInternalServerError)
			return
		}

		//configure origin from path
		manifestOrigin, err := origin.Configure(c, masterManifestPath)
		if err != nil {
			httpError(logger, w, err, "failed configuring origin", http.StatusInternalServerError)
			return
		}

		// fetch manifest from origin
		manifestContent, err := manifestOrigin.FetchManifest(c)
		if err != nil {
			httpError(logger, w, err, "failed fetching origin manifest content", http.StatusInternalServerError)
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
		default:
			err := fmt.Errorf("unsupported protocol %q", mediaFilters.Protocol)
			httpError(logger, w, err, "failed to select filter", http.StatusBadRequest)
			return
		}

		// apply the filters to the origin manifest
		filteredManifest, err := f.FilterManifest(mediaFilters)
		if err != nil {
			httpError(logger, w, err, "failed to filter manifest", http.StatusInternalServerError)
			return
		}

		// set cache-control if servering hls media playlist
		if maxAge := f.GetMaxAge(); maxAge != "" && maxAge != "0" {
			w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%v", maxAge))
		}

		// write the filtered manifest to the response
		fmt.Fprint(w, filteredManifest)
	})
}
