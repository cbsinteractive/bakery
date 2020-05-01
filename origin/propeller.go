package origin

import (
	"errors"
	"fmt"
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

func configurePropeller(c config.Config, path string) (Origin, error) {
	// Propeller channels can be requested in multiple formats.
	// When split, the parts in the path should evaluate to the following:
	// /propeller/orgID/channelID.m3u8
	// /propeller/orgID/clip/clipID.m3u8
	// /propeller/orgID/channelID/outputID.m3u8
	parts := strings.Split(path, "/")

	if len(parts) < 4 {
		err := fmt.Errorf("propeller origin: request format is not `/propeller/orgID/channelID.m3u8`")
		log := c.GetLogger()
		log.WithFields(logrus.Fields{
			"origin":  "propeller",
			"request": path,
		}).Error(err)
		return &Propeller{}, err
	}

	orgID := parts[2]

	if strings.Contains(parts[3], "clip") {
		return NewPropeller(c, orgID, extractID(parts[4]), getPropellerClipURL)
	}

	return NewPropeller(c, orgID, extractID(parts[3]), getPropellerChannelURL)
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
