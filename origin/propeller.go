package origin

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/cbsinteractive/bakery/config"
	propeller "github.com/cbsinteractive/propeller-go/client"
)

//Propeller struct holds basic config of a Propeller Channel
type Propeller struct {
	URL string
}

type fetchURL func(*propeller.Client, string, string) (string, error)

// propellerPaths defines the multiple path formats allowed for propeller
// channels and clips
var propellerPaths = []*regexp.Regexp{
	regexp.MustCompile(`/propeller/(?P<orgID>.+)/clip/(?P<clipID>.+).m3u8`),
	regexp.MustCompile(`/propeller/(?P<orgID>.+)/(?P<channelID>.+).m3u8`),
}

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

	if clipID != "" {
		return NewPropeller(c, orgID, clipID, getPropellerClipURL)
	}
	return NewPropeller(c, orgID, channelID, getPropellerChannelURL)
}

//GetPlaybackURL will retrieve url
func (p *Propeller) GetPlaybackURL() string {
	return p.URL
}

//FetchManifest will grab manifest contents of configured origin
func (p *Propeller) FetchManifest(c config.Client) (string, error) {
	return fetch(c, p.URL)
}

//NewPropeller returns a propeller struct
func NewPropeller(c config.Config, orgID string, endpointID string, get fetchURL) (*Propeller, error) {
	c.Propeller.UpdateContext(c.Client.Context)

	propellerURL, err := get(&c.Propeller.Client, orgID, endpointID)
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

func getPropellerChannelURL(client *propeller.Client, orgID string, channelID string) (string, error) {
	channel, err := client.GetChannel(orgID, channelID)
	if err != nil {
		var se propeller.StatusError
		if errors.As(err, &se) && se.NotFound() {
			return getPropellerClipURL(client, orgID, fmt.Sprintf("%v-archive", channelID))
		}

		return "", fmt.Errorf("fetching channel: %w", err)
	}

	return getChannelURL(channel)
}

func getChannelURL(channel propeller.Channel) (string, error) {
	//If a channel is "stopped", it will have an #EXT-X-ENDLIST tag
	//in its manifest, causing the DAI live playlist to 404.
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

func getPropellerClipURL(client *propeller.Client, orgID string, clipID string) (string, error) {
	clip, err := client.GetClip(orgID, clipID)
	if err != nil {
		return "", fmt.Errorf("fetching clip: %w", err)
	}

	return getClipURL(clip)
}

func getClipURL(clip propeller.Clip) (string, error) {
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

// extracID will extract the id from manifest name (id.m3u8)
func extractID(s string) string {
	return strings.Split(s, ".")[0]
}
