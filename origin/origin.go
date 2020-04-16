package origin

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
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
		return configurePropeller(c, path)
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
	req, err := http.NewRequest(http.MethodGet, manifestURL, nil)
	if err != nil {
		return "", fmt.Errorf("generating request to fetch manifest: %w", err)
	}

	ctx, cancel := context.WithTimeout(c.Client.Context, c.Client.Timeout)
	defer cancel()

	resp, err := c.Client.New().Do(req.WithContext(ctx))
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
