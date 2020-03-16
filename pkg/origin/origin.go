package origin

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/cbsinteractive/bakery/pkg/config"
)

//Origin interface is implemented on Manifest and Propeller struct
type Origin interface {
	GetPlaybackURL() string
	FetchManifest(c config.Config) (string, error)
}

//Manifest struct holds Origin and Path of Manifest
type Manifest struct {
	Origin string
	Path   string
}

//Configure will return proper Origin interface
func Configure(c config.Config, path string) (Origin, error) {
	if strings.Contains(path, "propeller") {
		parts := strings.Split(path, "/") //["", "propeller", "orgID", "channelID.m3u8"]
		if len(parts) != 4 {
			return &Propeller{}, fmt.Errorf("url path does not follow `/propeller/orgID/channelID.m3u8`")
		}

		orgID := parts[2]
		channelID := strings.Split(parts[3], ".")[0] // split off .m3u8

		o, err := NewPropeller(c, orgID, channelID)
		if err != nil {
			return &Propeller{}, fmt.Errorf("configuring propeller origin: %w", err)
		}

		return o, nil
	}

	return NewManifest(c, path), nil
}

//NewManifest returns a new Origin struct
func NewManifest(c config.Config, path string) *Manifest {
	return &Manifest{
		Origin: c.OriginHost,
		Path:   path,
	}
}

//GetPlaybackURL will retrieve url
func (m *Manifest) GetPlaybackURL() string {
	return m.Origin + m.Path
}

//FetchManifest will grab manifest contents of configured origin
func (m *Manifest) FetchManifest(c config.Config) (string, error) {
	return fetch(c, m.GetPlaybackURL())
}

func fetch(c config.Config, manifestURL string) (string, error) {
	resp, err := c.Client.New().Get(manifestURL)
	if err != nil {
		return "", fmt.Errorf("fetching manifest: %w", err)
	}
	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading manifest response body: %w", err)
	}

	if sc := resp.StatusCode; sc/100 > 3 {
		return "", fmt.Errorf("fetching manifest: returning http status of %v", sc)
	}

	return string(contents), nil
}
