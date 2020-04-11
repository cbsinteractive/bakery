package origin

import (
	"fmt"

	"github.com/cbsinteractive/bakery/config"
	propeller "github.com/cbsinteractive/propeller-client-go/pkg/client"
)

//Propeller struct holds basic config of a Propeller Channel
type Propeller struct {
	URL       string
	OrgID     string
	ChannelID string
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
func NewPropeller(p config.Propeller, orgID string, channelID string) (*Propeller, error) {
	propellerURL, err := getPropellerChannelURL(p, orgID, channelID)
	if err != nil {
		return &Propeller{}, fmt.Errorf("fetching propeller channel: %w", err)
	}

	return &Propeller{
		URL:       propellerURL,
		OrgID:     orgID,
		ChannelID: channelID,
	}, nil
}

func getPropellerChannelURL(p config.Propeller, orgID string, channelID string) (string, error) {
	channel, err := p.Client.GetChannel(orgID, channelID)
	if err != nil {
		return "", fmt.Errorf("fetching channel from propeller: %w", err)
	}

	return getURL(channel)
}

func getURL(channel propeller.Channel) (string, error) {
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

	return playbackURL.String(), nil
}
