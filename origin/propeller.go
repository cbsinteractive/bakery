package origin

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/sirupsen/logrus"

	"github.com/cbsinteractive/bakery/config"
	propeller "github.com/cbsinteractive/propeller-go/client"
)

// propellerPaths defines the multiple path formats allowed for propeller entities in Bakery
var propellerPaths = []*regexp.Regexp{
	regexp.MustCompile(`/propeller/(?P<orgID>.+)/clip/(?P<clipID>.+).m3u8`),
	regexp.MustCompile(`/propeller/(?P<orgID>.+)/(?P<channelID>.+).m3u8`),
}

// Propeller Origin holds the URL of a propeller entity (Channel, Clip)
type Propeller struct {
	URL string
}

// configurePropeller builds a new Propeller Origin given the Bakery config and the current url path
//
// The path will be matched agains one of propellerPaths patterns to find out the specific entity
// being requested (channel, clip) and a new Propeller Origin object is returned
//
// Return error if 'path' doesn't match with any of propellerPaths
func configurePropeller(c config.Config, path string) (Origin, error) {
	urlValues, err := parsePropellerPath(path)
	if err != nil {
		log := c.GetLogger()
		log.WithFields(logrus.Fields{
			"origin":  "propeller",
			"request": path,
		}).Error(err)
		return &Propeller{}, err
	}

	orgID := urlValues["orgID"]
	channelID := urlValues["channelID"]
	clipID := urlValues["clipID"]

	var getter urlGetter
	if clipID != "" {
		getter = &clipURLGetter{orgID: orgID, clipID: clipID}
	} else {
		getter = &channelURLGetter{orgID: orgID, channelID: channelID}
	}
	return NewPropeller(c, orgID, channelID, getter)
}

// NewPropeller returns a Propeller origin struct
func NewPropeller(c config.Config, orgID string, endpointID string, getter urlGetter) (*Propeller, error) {
	c.Propeller.UpdateContext(c.Client.Context)

	propellerURL, err := getter.GetURL(&c.Propeller.Client)
	if err != nil {
		err := fmt.Errorf("propeller origin: %w", err)
		log := c.GetLogger()
		log.WithFields(logrus.Fields{
			"origin":      "propeller",
			"org-id":      orgID,
			"manifest-id": endpointID,
		}).Error(err)
		return &Propeller{}, err
	}

	return &Propeller{
		URL: propellerURL,
	}, nil
}

// GetPlaybackURL will retrieve url
func (p *Propeller) GetPlaybackURL() string {
	return p.URL
}

// FetchManifest will grab manifest contents of configured origin
func (p *Propeller) FetchManifest(c config.Client) (string, error) {
	return fetch(c, p.URL)
}

// parsePropellerPath matches path against all proellerPaths patterns and return a map
// of values extracted from that url
//
// Return error if path does not match with any url
func parsePropellerPath(path string) (map[string]string, error) {
	values := make(map[string]string)
	for _, pattern := range propellerPaths {
		match := pattern.FindStringSubmatch(path)
		if len(match) == 0 {
			continue
		}
		for i, name := range pattern.SubexpNames() {
			if i != 0 {
				values[name] = match[i]
			}
		}
		return values, nil
	}
	return map[string]string{}, errors.New("propeller origin: request format is not `/propeller/orgID/channelID.m3u8`")
}

// propellerClient interface is the subset of methods from propeller-go client used by this module
type propellerClient interface {
	GetChannel(orgID string, channelID string) (propeller.Channel, error)
	GetClip(orgID string, clipID string) (propeller.Clip, error)
}

// urlGetter defines an interface for types that given a Propeller API Client know how to retrieve
// the playback url of that entity
type urlGetter interface {
	GetURL(client propellerClient) (string, error)
}

// channelURLGetter is a urlGetter for a Propeller channel
//
// Finds the channel playback_url using the Propeller API. If the channel is not found try
// to get the Archive url
type channelURLGetter struct {
	orgID     string
	channelID string
}

func (g *channelURLGetter) GetURL(client propellerClient) (string, error) {
	channel, err := client.GetChannel(g.orgID, g.channelID)
	if err != nil {
		if g.errChannelNotFound(err) {
			return g.getArchiveURL(client)
		}
		return "", fmt.Errorf("fetching channel: %w", err)
	}
	return g.getURL(channel)
}

func (g *channelURLGetter) errChannelNotFound(err error) bool {
	var se propeller.StatusError
	return errors.As(err, &se) && se.NotFound()
}

func (g *channelURLGetter) getArchiveURL(client propellerClient) (string, error) {
	clipGetter := &clipURLGetter{
		orgID:  g.orgID,
		clipID: fmt.Sprintf("%v-archive", g.channelID),
	}
	return clipGetter.GetURL(client)
}

func (g *channelURLGetter) getURL(channel propeller.Channel) (string, error) {
	// If a channel is "stopped", it will have an #EXT-X-ENDLIST tag
	// in its manifest, causing the DAI live playlist to 404.
	if channel.Ads && channel.Status == "running" {
		return channel.AdsURL, nil
	}
	if channel.Captions {
		return channel.CaptionsURL, nil
	}
	playbackURL, err := channel.URL()
	if err != nil {
		return "", fmt.Errorf("parsing channel url: %w", err)
	}
	return playbackURL.String(), nil
}

// clipURLGetter is a urlGetter for a Propeller clip
//
// Finds the Clip playback_url using the Propeller API
type clipURLGetter struct {
	orgID  string
	clipID string
}

func (g *clipURLGetter) GetURL(client propellerClient) (string, error) {
	clip, err := client.GetClip(g.orgID, g.clipID)
	if err != nil {
		return "", fmt.Errorf("fetching clip: %w", err)
	}
	return g.getURL(clip)
}

func (g *clipURLGetter) getURL(clip propeller.Clip) (string, error) {
	playbackURL, err := clip.URL()
	if err != nil {
		return "", fmt.Errorf("parsing clip url: %w", err)
	}
	playback := playbackURL.String()
	if playback == "" {
		return playback, fmt.Errorf("clip status: not ready")
	}
	return playback, nil
}
