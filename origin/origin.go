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

//Origin interface is implemented by DefaultOrigin and Propeller struct
type Origin interface {
	GetPlaybackURL() string
	FetchManifest(c config.Client) (string, error)
}

//DefaultOrigin struct holds Origin and Path of DefaultOrigin
//Variant level DefaultOrigins will be base64 encoded absolute path
type DefaultOrigin struct {
	Origin string
	URL    url.URL
}

//Configure will return proper Origin interface
func Configure(ctx context.Context, c config.Config, path string) (Origin, error) {
	if strings.Contains(path, "propeller") {
		return configurePropeller(ctx, c, path)
	}

	//check if rendition URL
	parts := strings.Split(path, "/")
	if len(parts) == 2 { //["", "base64.m3u8"]
		variantURL, err := decodeVariantURL(parts[1])
		if err != nil {
			return &DefaultOrigin{}, fmt.Errorf("decoding variant manifest url %q: %w", path, err)
		}
		path = variantURL
	}

	return NewDefaultOrigin(c.OriginHost, path)
}

//NewDefaultOrigin returns a new Origin struct
func NewDefaultOrigin(origin string, p string) (*DefaultOrigin, error) {
	u, err := url.Parse(p)
	if err != nil {
		return &DefaultOrigin{}, err
	}

	return &DefaultOrigin{
		Origin: origin,
		URL:    *u,
	}, nil
}

//GetPlaybackURL will retrieve url
func (d *DefaultOrigin) GetPlaybackURL() string {
	if d.URL.IsAbs() {
		return d.URL.String()
	}

	return d.Origin + d.URL.String()
}

//FetchManifest will grab DefaultOrigin contents of configured origin
func (d *DefaultOrigin) FetchManifest(c config.Client) (string, error) {
	return fetch(c, d.GetPlaybackURL())
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
