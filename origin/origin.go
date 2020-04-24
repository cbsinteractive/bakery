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
	"github.com/sirupsen/logrus"
)

//Origin interface is implemented on Manifest and Propeller struct
type Origin interface {
	GetPlaybackURL() string
	FetchManifest(c config.Client) (string, error)
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
		variantURL, err := decodeVariantURL(parts[1])
		if err != nil {
			err := fmt.Errorf("decoding variant manifest url: %w", err)
			log := c.GetLogger()
			log.WithFields(logrus.Fields{
				"origin":  "variant manifest",
				"request": path,
			}).Error(err)
			return &Manifest{}, err
		}
		path = variantURL
	}

	return NewManifest(c.OriginHost, path)
}

//NewManifest returns a new Origin struct
func NewManifest(origin string, p string) (*Manifest, error) {
	u, err := url.Parse(p)
	if err != nil {
		return &Manifest{}, nil
	}

	return &Manifest{
		Origin: origin,
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
func (m *Manifest) FetchManifest(c config.Client) (string, error) {
	return fetch(c, m.GetPlaybackURL())
}

func fetch(client config.Client, manifestURL string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, manifestURL, nil)
	if err != nil {
		return "", fmt.Errorf("generating request to fetch manifest: %w", err)
	}

	ctx, cancel := context.WithTimeout(client.Context, client.Timeout)
	defer cancel()

	resp, err := client.Do(req.WithContext(ctx))
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

func decodeVariantURL(variant string) (string, error) {
	variant = strings.TrimSuffix(variant, ".m3u8")
	url, err := base64.RawURLEncoding.DecodeString(variant)
	if err != nil {
		return "", err
	}

	return string(url), nil
}
