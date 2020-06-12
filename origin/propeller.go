package origin

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/cbsinteractive/bakery/config"
	"github.com/cbsinteractive/bakery/logging"
	propeller "github.com/cbsinteractive/propeller-go/client"
)

const (
	orgIDKey     = "orgID"
	clipIDKey    = "clipID"
	channelIDKey = "channelID"
	outputIDKey  = "outputID"
)

// propellerPaths defines the multiple path formats allowed for propeller entities in Bakery
var propellerPaths = []*regexp.Regexp{
	regexp.MustCompile(`/propeller/(?P<` + orgIDKey + `>.+)/clip/(?P<` + clipIDKey + `>.+).m3u8`),
	regexp.MustCompile(`/propeller/(?P<` + orgIDKey + `>.+)/(?P<` + channelIDKey + `>.+)/(?P<` + outputIDKey + `>.+).(m3u8|mpd)`),
	regexp.MustCompile(`/propeller/(?P<` + orgIDKey + `>.+)/(?P<` + channelIDKey + `>.+).(m3u8|mpd)`),
}

// Propeller Origin holds the URL of a propeller entity (Channel, Clip)
type Propeller struct {
	URL string
}

// configurePropeller builds a new Propeller Origin given the Bakery config and the current url path
//
// The path will be matched against one of propellerPaths patterns to find out the specific entity
// being requested (channel, clip) and a new Propeller Origin object is returned
//
// Return error if 'path' doesn't match with any of propellerPaths
func configurePropeller(ctx context.Context, c config.Config, path string) (Origin, error) {
	logging.UpdateCtx(ctx, logging.Params{"origin": "propeller"})

	urlValues, err := parsePropellerPath(path)
	if err != nil {
		return &Propeller{}, err
	}

	orgID := urlValues[orgIDKey]
	channelID := urlValues[channelIDKey]
	outputID := urlValues[outputIDKey]
	clipID := urlValues[clipIDKey]

	var getter urlGetter
	if clipID != "" {
		logging.UpdateCtx(ctx, logging.Params{orgIDKey: orgID, clipIDKey: clipID})
		getter = &clipURLGetter{orgID: orgID, clipID: clipID}
	} else if outputID != "" {
		logging.UpdateCtx(ctx, logging.Params{orgIDKey: orgID, channelIDKey: channelID, outputIDKey: outputID})
		getter = &outputURLGetter{orgID: orgID, channelID: channelID, outputID: outputID}
	} else {
		logging.UpdateCtx(ctx, logging.Params{orgIDKey: orgID, channelIDKey: channelID})
		getter = &channelURLGetter{orgID: orgID, channelID: channelID}
	}
	return NewPropeller(ctx, c, getter)
}

// NewPropeller returns a Propeller origin struct
func NewPropeller(ctx context.Context, c config.Config, getter urlGetter) (*Propeller, error) {
	c.Propeller.UpdateContext(c.Client.Context)

	propellerURL, err := getter.GetURL(&c.Propeller.Client)
	if err != nil {
		return &Propeller{}, fmt.Errorf("fetching propeller channel: %w", err)
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
	return nil, fmt.Errorf("propeller origin: invalid url format %v", path)
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
		return handleGetUrlChannelNotFound(err, g.orgID, g.channelID, client)
	}
	return g.getURL(channel)
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
	if channel.PlaybackURL == "" {
		if channel.Outputs != nil {
			return "", fmt.Errorf("channel has multiple outputs. Expect request format /propeller/org-id/channel-id/output-id.m3u8")
		}

		return "", fmt.Errorf("parsing channel url: channel not ready")
	}
	return channel.PlaybackURL, nil
}

// outputURLGetter is a urlGetter for a Propeller channel output
//
// Finds the output playback_url using the Propeller API. If the channel is not found try
// to get the Archive url
type outputURLGetter struct {
	orgID     string
	channelID string
	outputID  string
}

func (g *outputURLGetter) GetURL(client propellerClient) (string, error) {
	channel, err := client.GetChannel(g.orgID, g.channelID)
	if err != nil {
		return handleGetUrlChannelNotFound(err, g.orgID, g.channelID, client)
	}
	output, err := channel.FindOutput(g.outputID)
	if err != nil {
		return "", fmt.Errorf("finding channel output: %w", err)
	}
	return g.getURL(&channel, output)
}

func (g *outputURLGetter) getURL(channel *propeller.Channel, output *propeller.ChannelOutput) (string, error) {
	if output.AdsURL != "" && channel.Status == "running" {
		return output.AdsURL, nil
	}
	if output.CaptionsURL != "" {
		return output.CaptionsURL, nil
	}
	if output.PlaybackURL != "" {
		return output.PlaybackURL, nil
	}
	return "", fmt.Errorf("Channel output not ready")
}

// handleGetUrlChannelNotFound is an error handler used when trying to GET a channel
// in Propeller API and it failed
//
// When a channel is not found in Propeller we try to get the Clip archive URL
func handleGetUrlChannelNotFound(err error, orgID string, channelID string, client propellerClient) (string, error) {
	var se propeller.StatusError
	if errors.As(err, &se) && se.NotFound() {
		clipGetter := &clipURLGetter{
			orgID:  orgID,
			clipID: fmt.Sprintf("%v-archive", channelID),
		}
		archive, clipErr := clipGetter.GetURL(client)
		if clipErr != nil {
			return "", fmt.Errorf("Channel %v Not Found", channelID)
		}
		return archive, nil
	}
	return "", err
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
