package origin

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cbsinteractive/bakery/config"
)

//Origin interface is implemented by DefaultOrigin and Propeller struct
type Origin interface {
	GetPlaybackURL() string
	FetchManifest(c config.Client) (ManifestInfo, error)
}

//DefaultOrigin struct holds Origin and Path of DefaultOrigin
//Variant level DefaultOrigins will be base64 encoded absolute Urls
type DefaultOrigin struct {
	Host string
	URL  url.URL
}

//ManifestInfo holds http response info from manifest request
type ManifestInfo struct {
	Manifest     string
	LastModified time.Time
	Status       int
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
			c.Logger.Err(err).
				Str("origin", "variant").
				Msgf("can't decode url %v", path)
			return &DefaultOrigin{}, err
		}
		path = variantURL
	}

	return NewDefaultOrigin("", path)
}

//NewDefaultOrigin returns a new Origin struct
//host is not required if path is absolute
func NewDefaultOrigin(host string, p string) (*DefaultOrigin, error) {
	u, err := url.Parse(p)
	if err != nil {
		return &DefaultOrigin{}, err
	}

	return &DefaultOrigin{
		Host: host,
		URL:  *u,
	}, nil
}

//GetPlaybackURL will retrieve url
func (d *DefaultOrigin) GetPlaybackURL() string {
	if d.URL.IsAbs() {
		return d.URL.String()
	}

	return d.Host + d.URL.String()
}

//FetchManifest will grab DefaultOrigin contents of configured origin
func (d *DefaultOrigin) FetchManifest(c config.Client) (ManifestInfo, error) {
	return fetch(c, d.GetPlaybackURL())
}

func fetch(client config.Client, manifestURL string) (ManifestInfo, error) {
	req, err := http.NewRequest(http.MethodGet, manifestURL, nil)
	if err != nil {
		return ManifestInfo{}, fmt.Errorf("generating request to fetch manifest: %w", err)
	}

	ctx, cancel := context.WithTimeout(client.Context, client.Timeout)
	defer cancel()

	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return ManifestInfo{}, fmt.Errorf("fetching manifest: %w", err)
	}
	defer resp.Body.Close()

	manifest, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ManifestInfo{}, fmt.Errorf("reading manifest response body: %w", err)
	}

	lastModified, err := http.ParseTime(resp.Header.Get("Last-Modified"))
	if err != nil {
		return ManifestInfo{}, err
	}

	return ManifestInfo{
		Manifest:     string(manifest),
		LastModified: lastModified,
		Status:       resp.StatusCode,
	}, nil
}

func decodeVariantURL(variant string) (string, error) {
	variant = strings.TrimSuffix(variant, ".m3u8")
	url, err := base64.RawURLEncoding.DecodeString(variant)
	if err != nil {
		return "", err
	}

	return string(url), nil
}
