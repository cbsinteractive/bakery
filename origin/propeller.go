package origin

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cbsinteractive/bakery/config"
	propeller "github.com/cbsinteractive/propeller-client-go/pkg/client"
)

//Propeller struct holds basic config of a Propeller Channel
//ClipID and ChannelID will never be set at the same time
type Propeller struct {
	URL string
}

type fetchURL func(config.Propeller, string, string) (string, error)

func configurePropeller(c config.Config, path string) (Origin, error) {
	// Propeller channels can be requested in multiple formats.
	// When split, the parts in the path should evaluate to the following:
	// /propeller/orgID/channelID.m3u8
	// /propeller/orgID/clip/clipID.m3u8
	// /propeller/orgID/channelID/outputID.m3u8
	parts := strings.Split(path, "/")

	if len(parts) < 4 {
		return &Propeller{}, fmt.Errorf("url path does not follow `/propeller/orgID/channelID.m3u8`")
	}

	orgID := parts[2]

	if strings.Contains(parts[3], "clip") {
		return NewPropeller(c.Propeller, orgID, extractID(parts[4]), getPropellerClipURL)
	}

	return NewPropeller(c.Propeller, orgID, extractID(parts[3]), getPropellerChannelURL)
}

//GetPlaybackURL will retrieve url
func (p *Propeller) GetPlaybackURL() string {
	return p.URL
}

//FetchManifest will grab manifest contents of configured origin
func (p *Propeller) FetchManifest(c config.Config) (string, error) {
	return fetch(c, p.URL)
}

//NewPropeller returns a propeller struct
func NewPropeller(p config.Propeller, orgID string, endpointID string, get fetchURL) (*Propeller, error) {
	propellerURL, err := get(p, orgID, endpointID)
	if err != nil {
		return &Propeller{}, fmt.Errorf("fetching propeller channel: %w", err)
	}

	return &Propeller{
		URL: propellerURL,
	}, nil
}

func getPropellerChannelURL(p config.Propeller, orgID string, channelID string) (string, error) {
	channel, err := p.Client.GetChannel(orgID, channelID)
	if err != nil {
		var se propeller.StatusError
		if errors.As(err, &se) && se.NotFound() {
			fmt.Println("fetching archive url")
			return getPropellerClipURL(p, orgID, fmt.Sprintf("%v-archive", channelID))
		}

		fmt.Println(channel)
		return "", fmt.Errorf("fetching channel from propeller: %w", err)
	}

	return getChannelURL(channel)
}

func getChannelURL(channel propeller.Channel) (string, error) {
	if channel.Ads {
		return channel.AdsURL, nil
	}

	if channel.Captions {
		return channel.CaptionsURL, nil
	}

	playbackURL, err := channel.URL()
	if err != nil {
		return "", fmt.Errorf("reading url from propeller channel: %w", err)
	}

	playback := playbackURL.String()
	if playback != "" {
		return playback, fmt.Errorf("channel not ready")
	}

	return playback, nil
}

func getPropellerClipURL(p config.Propeller, orgID string, clipID string) (string, error) {
	clip, err := p.Client.GetClip(orgID, clipID)
	if err != nil {
		return "", fmt.Errorf("fetching clip from propeller: %w", err)
	}

	return getClipURL(clip)
}

func getClipURL(clip propeller.Clip) (string, error) {
	playbackURL, err := clip.URL()
	if err != nil {
		return "", fmt.Errorf("reading url from propeller clip: %w", err)
	}

	playback := playbackURL.String()
	if playback != "" {
		return playback, fmt.Errorf("clip not ready")
	}

	return playback, nil

}

// extracID will extract the id from manifest name (id.m3u8)
func extractID(s string) string {
	return strings.Split(s, ".")[0]
}
