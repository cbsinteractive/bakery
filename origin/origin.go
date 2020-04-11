package origin

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"

	"github.com/cbsinteractive/bakery/config"
)

//Origin interface is implemented on Manifest and Propeller struct
type Origin interface {
	GetPlaybackURL() string
	FetchManifest(c config.Config) (string, error)
}

//Manifest struct holds Origin and Path of Manifest
//Variant level manifests will be base64 encoded absolute path
type Manifest struct {
	Origin string
	URL    url.URL
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

		o, err := NewPropeller(c.Propeller, orgID, channelID)
		if err != nil {
			return &Propeller{}, fmt.Errorf("configuring propeller origin: %w", err)
		}

		return o, nil
	}

	//check if rendition URL
	parts := strings.Split(path, "/")
	if len(parts) == 2 { //["", "base64.m3u8"]
		renditionURL, err := decodeRenditionURL(parts[1])
		if err != nil {
			return &Manifest{}, fmt.Errorf("configuring rendition url: %w", err)
		}
		path = renditionURL
	}

	return NewManifest(c, path)
}

//NewManifest returns a new Origin struct
func NewManifest(c config.Config, p string) (*Manifest, error) {
	u, err := url.Parse(p)
	if err != nil {
		return &Manifest{}, nil
	}

	return &Manifest{
		Origin: c.OriginHost,
		URL:    *u,
	}, nil
}

//GetPlaybackURL will retrieve url
func (m *Manifest) GetPlaybackURL() string {
	if m.URL.IsAbs() {
		return m.URL.String()
	}

	return m.Origin + m.URL.String()
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

func decodeRenditionURL(rendition string) (string, error) {
	rendition = strings.TrimSuffix(rendition, ".m3u8")
	url, err := base64.RawURLEncoding.DecodeString(rendition)
	if err != nil {
		return "", fmt.Errorf("decoding rendition: %w", err)
	}

	return string(url), nil
}
